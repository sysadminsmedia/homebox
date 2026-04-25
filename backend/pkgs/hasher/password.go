package hasher

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

var enabled = true

type params struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

var p = &params{
	memory:      64 * 1024,
	iterations:  3,
	parallelism: 2,
	saltLength:  16,
	keyLength:   32,
}

func init() { // nolint: gochecknoinits
	disableHas := os.Getenv("UNSAFE_DISABLE_PASSWORD_PROJECTION") == "yes_i_am_sure"

	if disableHas {
		// Print a big ol warning in red
		fmt.Println("\\[\\033[0;31m\\]", "=======================================================================")
		fmt.Println("\\[\\033[0;31m\\]", "WARNING: Password protection is disabled. This is unsafe in production.")
		fmt.Println("\\[\\033[0;31m\\]", "You should never, ever use this in production. It is only for development and testing purposes.")
		fmt.Println("\\[\\033[0;31m\\]", "DO NOT USE THIS IN PRODUCTION!")
		fmt.Println("\\[\\033[0;31m\\]", "Remove UNSAFE_DISABLE_PASSWORD_PROJECTION to disable this warning.")
		fmt.Println("\\[\\033[0;31m\\]", "=======================================================================")
		enabled = false
	}
}

func hasherTracer() trace.Tracer {
	return otel.Tracer("hasher")
}

func recordSpanError(span trace.Span, err error) {
	if err == nil {
		return
	}
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

func GenerateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// HashPassword hashes the password with argon2id and returns the encoded hash. It is
// equivalent to HashPasswordCtx with a Background context and exists for callers that
// have no context to thread through.
func HashPassword(password string) (string, error) {
	return HashPasswordCtx(context.Background(), password)
}

// HashPasswordCtx hashes the password with argon2id, emitting a span describing the work.
// Splitting the salt and KDF into sub-spans makes it possible to correlate slow logins
// with the argon2 derivation specifically.
func HashPasswordCtx(ctx context.Context, password string) (string, error) {
	ctx, span := hasherTracer().Start(ctx, "hasher.HashPassword",
		trace.WithAttributes(
			attribute.Bool("password.protection.enabled", enabled),
			attribute.Int("password.length", len(password)),
		))
	defer span.End()

	if !enabled {
		span.SetAttributes(attribute.String("password.algorithm", "plaintext-disabled"))
		return password, nil
	}

	span.SetAttributes(
		attribute.String("password.algorithm", "argon2id"),
		attribute.Int64("argon2.memory_kib", int64(p.memory)),
		attribute.Int64("argon2.iterations", int64(p.iterations)),
		attribute.Int("argon2.parallelism", int(p.parallelism)),
		attribute.Int64("argon2.salt_length", int64(p.saltLength)),
		attribute.Int64("argon2.key_length", int64(p.keyLength)),
	)

	saltCtx, saltSpan := hasherTracer().Start(ctx, "hasher.HashPassword.salt")
	salt, err := GenerateRandomBytes(p.saltLength)
	if err != nil {
		recordSpanError(saltSpan, err)
		saltSpan.End()
		recordSpanError(span, err)
		return "", err
	}
	saltSpan.End()
	_ = saltCtx

	_, kdfSpan := hasherTracer().Start(ctx, "hasher.HashPassword.argon2.derive")
	hash := argon2.IDKey([]byte(password), salt, p.iterations, p.memory, p.parallelism, p.keyLength)
	kdfSpan.End()

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, 64*1024, 3, 2, b64Salt, b64Hash)
	span.SetAttributes(attribute.Int("password.encoded_hash.length", len(encodedHash)))
	return encodedHash, nil
}

// CheckPasswordHash checks if the provided password matches the hash. It is equivalent
// to CheckPasswordHashCtx with a Background context.
func CheckPasswordHash(password, hash string) (bool, bool) {
	return CheckPasswordHashCtx(context.Background(), password, hash)
}

// CheckPasswordHashCtx checks the password and emits spans for each branch (argon2id
// success, argon2id mismatch, bcrypt fallback, both-failed). Errors from the argon2
// decode path are recorded on the span instead of being silently swallowed — that
// silent swallowing was a major obstacle to debugging intermittent password rejections
// from valid hashes.
//
// Returns (match, needsRehash). needsRehash is true when the stored hash was a
// legacy bcrypt hash and the password matched; the caller is expected to rehash with
// argon2id and persist.
func CheckPasswordHashCtx(ctx context.Context, password, hash string) (bool, bool) {
	ctx, span := hasherTracer().Start(ctx, "hasher.CheckPasswordHash",
		trace.WithAttributes(
			attribute.Bool("password.protection.enabled", enabled),
			attribute.Int("password.length", len(password)),
			attribute.Int("password.hash.length", len(hash)),
			attribute.String("password.hash.prefix", hashPrefix(hash)),
		))
	defer span.End()

	if !enabled {
		matched := password == hash
		span.SetAttributes(
			attribute.String("password.algorithm", "plaintext-disabled"),
			attribute.Bool("password.match", matched),
			attribute.String("password.outcome", outcomeForMatch(matched, "plaintext-disabled")),
		)
		return matched, false
	}

	match, decodeErr, compareErr := comparePasswordAndHash(ctx, password, hash)
	span.SetAttributes(
		attribute.Bool("password.argon2.decode_ok", decodeErr == nil),
		attribute.Bool("password.argon2.compare_ok", compareErr == nil),
		attribute.Bool("password.argon2.match", match),
	)

	switch {
	case decodeErr != nil:
		span.SetAttributes(attribute.String("password.argon2.decode_error", decodeErr.Error()))
		recordSpanError(span, fmt.Errorf("argon2 decode failed: %w", decodeErr))
	case compareErr != nil:
		span.SetAttributes(attribute.String("password.argon2.compare_error", compareErr.Error()))
		recordSpanError(span, fmt.Errorf("argon2 compare failed: %w", compareErr))
	}

	if decodeErr == nil && compareErr == nil && match {
		span.SetAttributes(
			attribute.String("password.algorithm", "argon2id"),
			attribute.Bool("password.match", true),
			attribute.Bool("password.rehash_needed", false),
			attribute.String("password.outcome", "argon2id_match"),
		)
		return true, false
	}

	// Argon2 failed or didn't match — try bcrypt as a legacy fallback.
	_, bcryptSpan := hasherTracer().Start(ctx, "hasher.CheckPasswordHash.bcrypt")
	bcryptErr := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	bcryptSpan.SetAttributes(
		attribute.Bool("password.bcrypt.match", bcryptErr == nil),
		attribute.String("password.bcrypt.error", errString(bcryptErr)),
		attribute.String("password.bcrypt.error_kind", bcryptErrorKind(bcryptErr)),
	)
	if bcryptErr != nil {
		// Not a logical "error" if the hash simply isn't a bcrypt hash — that's the expected
		// case when the stored hash is argon2id and the password didn't match. Don't mark the
		// span as failed for that.
		if !errors.Is(bcryptErr, bcrypt.ErrMismatchedHashAndPassword) && !errors.Is(bcryptErr, bcrypt.ErrHashTooShort) {
			recordSpanError(bcryptSpan, bcryptErr)
		}
	}
	bcryptSpan.End()

	if bcryptErr == nil {
		span.SetAttributes(
			attribute.String("password.algorithm", "bcrypt"),
			attribute.Bool("password.match", true),
			attribute.Bool("password.rehash_needed", true),
			attribute.String("password.outcome", "bcrypt_match_needs_rehash"),
		)
		return true, true
	}

	var outcome string
	switch {
	case decodeErr != nil:
		outcome = "rejected_argon2_decode_failed"
	case compareErr != nil:
		outcome = "rejected_argon2_compare_failed"
	default:
		outcome = "rejected_argon2_mismatch_and_not_bcrypt"
	}
	span.SetAttributes(
		attribute.Bool("password.match", false),
		attribute.Bool("password.rehash_needed", false),
		attribute.String("password.outcome", outcome),
	)
	return false, false
}

// comparePasswordAndHash returns (match, decodeErr, compareErr). Splitting the
// errors lets callers distinguish "the stored hash is malformed" (a real bug)
// from "the password didn't match" (expected for wrong passwords).
func comparePasswordAndHash(ctx context.Context, password, encodedHash string) (match bool, decodeErr error, compareErr error) {
	ctx, span := hasherTracer().Start(ctx, "hasher.comparePasswordAndHash")
	defer span.End()

	decodeCtx, decodeSpan := hasherTracer().Start(ctx, "hasher.comparePasswordAndHash.decode")
	storedParams, salt, hash, err := decodeHash(encodedHash)
	if err != nil {
		decodeSpan.SetAttributes(
			attribute.String("decode.error", err.Error()),
			attribute.String("decode.error_kind", classifyDecodeError(err)),
		)
		recordSpanError(decodeSpan, err)
		decodeSpan.End()
		span.SetAttributes(attribute.Bool("decode.ok", false))
		return false, err, nil
	}
	decodeSpan.SetAttributes(
		attribute.Bool("decode.ok", true),
		attribute.Int64("decode.argon2.memory_kib", int64(storedParams.memory)),
		attribute.Int64("decode.argon2.iterations", int64(storedParams.iterations)),
		attribute.Int("decode.argon2.parallelism", int(storedParams.parallelism)),
		attribute.Int64("decode.argon2.salt_length", int64(storedParams.saltLength)),
		attribute.Int64("decode.argon2.key_length", int64(storedParams.keyLength)),
	)
	decodeSpan.End()
	_ = decodeCtx

	span.SetAttributes(
		attribute.Bool("decode.ok", true),
		attribute.Bool("params.match_current",
			storedParams.memory == p.memory &&
				storedParams.iterations == p.iterations &&
				storedParams.parallelism == p.parallelism &&
				storedParams.saltLength == p.saltLength &&
				storedParams.keyLength == p.keyLength),
	)

	_, deriveSpan := hasherTracer().Start(ctx, "hasher.comparePasswordAndHash.derive")
	otherHash := argon2.IDKey([]byte(password), salt, storedParams.iterations, storedParams.memory, storedParams.parallelism, storedParams.keyLength)
	deriveSpan.End()

	_, cmpSpan := hasherTracer().Start(ctx, "hasher.comparePasswordAndHash.constantTimeCompare")
	matched := subtle.ConstantTimeCompare(hash, otherHash) == 1
	cmpSpan.SetAttributes(attribute.Bool("compare.match", matched))
	cmpSpan.End()

	span.SetAttributes(attribute.Bool("compare.match", matched))
	return matched, nil, nil
}

func decodeHash(encodedHash string) (out *params, salt, hash []byte, err error) {
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		return nil, nil, nil, fmt.Errorf("invalid hash format: expected 6 segments, got %d", len(vals))
	}

	var version int
	_, err = fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("invalid version segment %q: %w", vals[2], err)
	}
	if version != argon2.Version {
		return nil, nil, nil, fmt.Errorf("unsupported argon2 version: got %d, want %d", version, argon2.Version)
	}

	out = &params{}
	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &out.memory, &out.iterations, &out.parallelism)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("invalid params segment %q: %w", vals[3], err)
	}

	salt, err = base64.RawStdEncoding.Strict().DecodeString(vals[4])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("invalid salt: %w", err)
	}
	out.saltLength = uint32(len(salt))

	hash, err = base64.RawStdEncoding.Strict().DecodeString(vals[5])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("invalid key: %w", err)
	}
	out.keyLength = uint32(len(hash))

	return out, salt, hash, nil
}

// hashPrefix returns the algorithm prefix of an encoded hash if recognizable, without
// leaking any of the actual key material. This is safe to attach to spans for diagnosis.
func hashPrefix(hash string) string {
	switch {
	case strings.HasPrefix(hash, "$argon2id$"):
		return "argon2id"
	case strings.HasPrefix(hash, "$argon2i$"):
		return "argon2i"
	case strings.HasPrefix(hash, "$argon2d$"):
		return "argon2d"
	case strings.HasPrefix(hash, "$2a$"), strings.HasPrefix(hash, "$2b$"), strings.HasPrefix(hash, "$2y$"):
		return "bcrypt"
	case hash == "":
		return "empty"
	default:
		return "unknown"
	}
}

func classifyDecodeError(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, "expected 6 segments"):
		return "format_segment_count"
	case strings.Contains(msg, "invalid version"):
		return "version_parse"
	case strings.Contains(msg, "unsupported argon2 version"):
		return "version_unsupported"
	case strings.Contains(msg, "invalid params"):
		return "params_parse"
	case strings.Contains(msg, "invalid salt"):
		return "salt_b64"
	case strings.Contains(msg, "invalid key"):
		return "key_b64"
	default:
		return "other"
	}
}

func bcryptErrorKind(err error) string {
	switch {
	case err == nil:
		return ""
	case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
		return "mismatch"
	case errors.Is(err, bcrypt.ErrHashTooShort):
		return "hash_too_short"
	default:
		if _, ok := errors.AsType[bcrypt.HashVersionTooNewError](err); ok {
			return "version_too_new"
		}
		if _, ok := errors.AsType[bcrypt.InvalidHashPrefixError](err); ok {
			return "invalid_prefix"
		}
		return "other"
	}
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func outcomeForMatch(match bool, algorithm string) string {
	if match {
		return algorithm + "_match"
	}
	return algorithm + "_mismatch"
}
