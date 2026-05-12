package repo

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/passwordresettokens"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type PasswordResetTokenRepository struct {
	db *ent.Client
}

type PasswordResetToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	ExpiresAt time.Time
}

// Create persists a hashed reset token for the given user.
func (r *PasswordResetTokenRepository) Create(ctx context.Context, userID uuid.UUID, tokenHash []byte, expiresAt time.Time) (PasswordResetToken, error) {
	ctx, span := entityTracer().Start(ctx, "repo.PasswordResetTokenRepository.Create",
		trace.WithAttributes(
			attribute.String("user.id", userID.String()),
			attribute.String("token.expires_at", expiresAt.Format(time.RFC3339)),
		))
	defer span.End()

	row, err := r.db.PasswordResetTokens.Create().
		SetUserID(userID).
		SetToken(tokenHash).
		SetExpiresAt(expiresAt).
		Save(ctx)
	if err != nil {
		recordSpanError(span, err)
		return PasswordResetToken{}, err
	}

	return PasswordResetToken{
		ID:        row.ID,
		UserID:    userID,
		ExpiresAt: row.ExpiresAt,
	}, nil
}

// GetValidByHash returns the token row matching the given hash if it has not
// expired and has not been used. Returns ent.NotFound otherwise so callers
// can't distinguish "wrong token" from "expired" from "used" — all three look
// the same from the outside, which is what we want.
func (r *PasswordResetTokenRepository) GetValidByHash(ctx context.Context, tokenHash []byte) (PasswordResetToken, error) {
	ctx, span := entityTracer().Start(ctx, "repo.PasswordResetTokenRepository.GetValidByHash",
		trace.WithAttributes(attribute.Int("token.hash.length", len(tokenHash))))
	defer span.End()

	row, err := r.db.PasswordResetTokens.Query().
		Where(
			passwordresettokens.Token(tokenHash),
			passwordresettokens.UsedAtIsNil(),
			passwordresettokens.ExpiresAtGT(time.Now()),
		).
		WithUser().
		Only(ctx)
	if err != nil {
		span.SetAttributes(
			attribute.Bool("token.found", false),
			attribute.Bool("token.lookup.not_found", ent.IsNotFound(err)),
		)
		if !ent.IsNotFound(err) {
			recordSpanError(span, err)
		}
		return PasswordResetToken{}, err
	}

	span.SetAttributes(attribute.Bool("token.found", true))
	return PasswordResetToken{
		ID:        row.ID,
		UserID:    row.Edges.User.ID,
		ExpiresAt: row.ExpiresAt,
	}, nil
}

// ErrPasswordResetTokenAlreadyClaimed is returned by MarkUsed when no
// unclaimed row matched the predicate — i.e. the token was used by a
// concurrent request. Callers must treat this as the same kind of failure as
// "token not found" so a race doesn't let two resets succeed against the
// same token.
var ErrPasswordResetTokenAlreadyClaimed = errors.New("password reset token was already claimed")

// MarkUsed atomically sets used_at on the token, succeeding only if used_at
// was previously NULL. Returns ErrPasswordResetTokenAlreadyClaimed if a
// concurrent caller won the race. Callers should claim BEFORE running any
// side effects (changing the password, etc.) so the token can be used at
// most once even under contention.
func (r *PasswordResetTokenRepository) MarkUsed(ctx context.Context, id uuid.UUID, at time.Time) error {
	ctx, span := entityTracer().Start(ctx, "repo.PasswordResetTokenRepository.MarkUsed",
		trace.WithAttributes(attribute.String("token.id", id.String())))
	defer span.End()

	affected, err := r.db.PasswordResetTokens.Update().
		Where(passwordresettokens.ID(id), passwordresettokens.UsedAtIsNil()).
		SetUsedAt(at).
		Save(ctx)
	if err != nil {
		recordSpanError(span, err)
		return err
	}
	span.SetAttributes(attribute.Int("tokens.claimed.count", affected))
	if affected == 0 {
		return ErrPasswordResetTokenAlreadyClaimed
	}
	return nil
}

// PurgeExpired deletes expired and already-used tokens. Run periodically.
func (r *PasswordResetTokenRepository) PurgeExpired(ctx context.Context) (int, error) {
	ctx, span := entityTracer().Start(ctx, "repo.PasswordResetTokenRepository.PurgeExpired")
	defer span.End()

	deleted, err := r.db.PasswordResetTokens.Delete().
		Where(passwordresettokens.Or(
			passwordresettokens.ExpiresAtLTE(time.Now()),
			passwordresettokens.UsedAtNotNil(),
		)).
		Exec(ctx)
	if err != nil {
		recordSpanError(span, err)
		return 0, err
	}
	span.SetAttributes(attribute.Int("tokens.deleted.count", deleted))
	return deleted, nil
}
