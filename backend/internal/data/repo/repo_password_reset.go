package repo

import (
	"context"
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

// MarkUsed sets used_at on the token so it cannot be replayed.
func (r *PasswordResetTokenRepository) MarkUsed(ctx context.Context, id uuid.UUID, at time.Time) error {
	ctx, span := entityTracer().Start(ctx, "repo.PasswordResetTokenRepository.MarkUsed",
		trace.WithAttributes(attribute.String("token.id", id.String())))
	defer span.End()

	err := r.db.PasswordResetTokens.UpdateOneID(id).SetUsedAt(at).Exec(ctx)
	recordSpanError(span, err)
	return err
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
