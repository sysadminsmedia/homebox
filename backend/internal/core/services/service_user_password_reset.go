package services

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/pkgs/hasher"
	"github.com/sysadminsmedia/homebox/backend/pkgs/mailer"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// MailerReady reports whether the SMTP mailer is configured and usable. The
// HTTP forgot-password handler uses this to short-circuit with a clear error
// when SMTP is missing, so the user sees something actionable instead of a
// generic success message followed by no email arriving.
func (svc *UserService) MailerReady() bool {
	return svc.mailer != nil && svc.mailer.Ready()
}

// RequestPasswordReset issues a single-use reset token and emails the link.
// All work — user lookup, token creation, SMTP send — runs in a background
// goroutine so the HTTP response time does not depend on whether the email
// is registered. Without this, the SMTP call (tens of ms) would defeat the
// "always 204" enumeration defense by letting an attacker time-distinguish
// known accounts from unknown ones. Only the up-front "is the mailer
// configured" check runs synchronously.
//
// baseURL is the absolute URL prefix (scheme + host) the reset link is built
// against. The handler MUST resolve it via SecureBaseURL — never via
// GetHBURL — so a forged Referer or untrusted X-Forwarded-Host can't poison
// the link in the victim's email.
func (svc *UserService) RequestPasswordReset(ctx context.Context, email, baseURL string) error {
	_, span := entityServiceTracer().Start(ctx, "service.UserService.RequestPasswordReset",
		trace.WithAttributes(
			attribute.Int("user.email.length", len(email)),
			attribute.Int("base_url.length", len(baseURL)),
		))
	defer span.End()

	if !svc.MailerReady() {
		span.SetAttributes(attribute.String("reset.outcome", "mailer_not_configured"))
		return ErrorMailerNotConfigured
	}

	// Detach from the request context so a client disconnect or timeout
	// doesn't abort the email send halfway through. Errors are logged, never
	// propagated — surfacing them would re-introduce the enumeration leak.
	go svc.processResetRequest(email, baseURL)

	span.SetAttributes(attribute.String("reset.outcome", "queued"))
	return nil
}

func (svc *UserService) processResetRequest(email, baseURL string) {
	ctx, span := entityServiceTracer().Start(context.Background(), "service.UserService.processResetRequest",
		trace.WithAttributes(attribute.Int("user.email.length", len(email))))
	defer span.End()

	rawToken, usr, err := svc.createResetToken(ctx, email)
	if err != nil {
		switch {
		case ent.IsNotFound(err):
			span.SetAttributes(attribute.String("reset.outcome", "user_not_found"))
		case errors.Is(err, errResetUserHasNoPassword):
			span.SetAttributes(attribute.String("reset.outcome", "user_no_password"))
		default:
			recordServiceSpanError(span, err)
			span.SetAttributes(attribute.String("reset.outcome", "create_failed"))
			log.Err(err).Msg("failed to create password reset token")
		}
		return
	}
	span.SetAttributes(attribute.String("user.id", usr.ID.String()))

	link := buildResetLink(baseURL, rawToken)
	if err := svc.sendResetEmail(usr, link); err != nil {
		recordServiceSpanError(span, err)
		span.SetAttributes(attribute.String("reset.outcome", "email_send_failed"))
		log.Err(err).Str("user.id", usr.ID.String()).Msg("failed to send password reset email")
		return
	}
	span.SetAttributes(attribute.String("reset.outcome", "sent"))
}

// GenerateResetLink mints a token and returns the reset URL without sending
// email. The CLI subcommand calls this so an admin can recover an account when
// SMTP is not configured (e.g. a small self-hosted instance, or to debug the
// password matching path itself). Unlike the HTTP path, this returns
// ent.NotFound for an unknown email — the caller is the operator, who needs
// to know if they typed the address wrong.
func (svc *UserService) GenerateResetLink(ctx context.Context, email, baseURL string) (string, error) {
	ctx, span := entityServiceTracer().Start(ctx, "service.UserService.GenerateResetLink",
		trace.WithAttributes(
			attribute.Int("user.email.length", len(email)),
			attribute.Int("base_url.length", len(baseURL)),
		))
	defer span.End()

	rawToken, usr, err := svc.createResetToken(ctx, email)
	if err != nil {
		if !ent.IsNotFound(err) && !errors.Is(err, errResetUserHasNoPassword) {
			recordServiceSpanError(span, err)
		}
		return "", err
	}
	span.SetAttributes(attribute.String("user.id", usr.ID.String()))

	return buildResetLink(baseURL, rawToken), nil
}

// ResetPassword consumes a reset token and changes the password. On success it
// also revokes every active session token for the user — the password change
// must invalidate any session an attacker (or anyone on a shared device) might
// be holding. Returns ErrorPasswordResetInvalid for any invalid/expired/used
// token, deliberately conflating the three cases.
func (svc *UserService) ResetPassword(ctx context.Context, rawToken, newPassword string) error {
	ctx, span := entityServiceTracer().Start(ctx, "service.UserService.ResetPassword",
		trace.WithAttributes(
			attribute.Int("token.length", len(rawToken)),
			attribute.Int("password.new.length", len(newPassword)),
		))
	defer span.End()

	if rawToken == "" || newPassword == "" {
		span.SetAttributes(attribute.String("reset.outcome", "missing_input"))
		return ErrorPasswordResetInvalid
	}

	hash := hasher.HashToken(rawToken)
	tok, err := svc.repos.PasswordResetTokens.GetValidByHash(ctx, hash)
	if err != nil {
		if ent.IsNotFound(err) {
			// SECURITY: equalize response time with the valid-token path.
			// The valid-token path runs argon2id (~50ms) + DB writes; without
			// a dummy hash here, an attacker could distinguish valid from
			// invalid tokens by response time. argon2id dominates the cost
			// of the success path, so equalizing it brings the residual gap
			// well below typical network jitter and below practical
			// distinguishability. The dummy hash result is discarded.
			_, _ = hasher.HashPasswordCtx(ctx, newPassword)
			span.SetAttributes(attribute.String("reset.outcome", "token_invalid"))
			return ErrorPasswordResetInvalid
		}
		recordServiceSpanError(span, err)
		span.SetAttributes(attribute.String("reset.outcome", "lookup_failed"))
		return err
	}
	span.SetAttributes(
		attribute.String("user.id", tok.UserID.String()),
		attribute.String("token.id", tok.ID.String()),
	)

	hashed, err := hasher.HashPasswordCtx(ctx, newPassword)
	if err != nil {
		recordServiceSpanError(span, err)
		span.SetAttributes(attribute.String("reset.outcome", "hash_failed"))
		return err
	}

	// ConsumeAndChangePassword runs the atomic conditional claim of the token
	// (used_at IS NULL → now) and the password update in a single ent
	// transaction, so a transient DB failure can never leave us in the bad
	// state where the token is burned but the password is unchanged. The
	// conditional UPDATE still defends against the concurrent-reset race —
	// a racer sees 0 rows affected and gets ErrPasswordResetTokenAlreadyClaimed.
	// Session revocation runs only after the commit succeeds.
	if err := svc.repos.PasswordResetTokens.ConsumeAndChangePassword(ctx, tok.ID, tok.UserID, hashed); err != nil {
		if errors.Is(err, repo.ErrPasswordResetTokenAlreadyClaimed) {
			span.SetAttributes(attribute.String("reset.outcome", "claim_race_lost"))
			return ErrorPasswordResetInvalid
		}
		recordServiceSpanError(span, err)
		span.SetAttributes(attribute.String("reset.outcome", "persist_failed"))
		return err
	}

	revoked, err := svc.repos.AuthTokens.DeleteAllByUser(ctx, tok.UserID)
	if err != nil {
		// The password is already changed and the token consumed; logging the
		// session-revocation failure is more useful than failing the whole
		// reset and leaving the user unable to log in with their new password.
		log.Err(err).Str("user.id", tok.UserID.String()).Msg("failed to revoke sessions after password reset")
		span.SetAttributes(attribute.String("reset.outcome", "success_revoke_failed"))
		return nil
	}

	span.SetAttributes(
		attribute.String("reset.outcome", "success"),
		attribute.Int("sessions.revoked.count", revoked),
	)
	return nil
}

// errResetUserHasNoPassword is an internal sentinel for the OIDC-only-user
// case. Callers translate it to either a silent success (HTTP) or a real
// error (CLI).
var errResetUserHasNoPassword = errors.New("user has no password set")

func (svc *UserService) createResetToken(ctx context.Context, email string) (string, repo.UserOut, error) {
	ctx, span := entityServiceTracer().Start(ctx, "service.UserService.createResetToken",
		trace.WithAttributes(attribute.Int("user.email.length", len(email))))
	defer span.End()

	usr, err := svc.repos.Users.GetOneEmail(ctx, email)
	if err != nil {
		span.SetAttributes(attribute.Bool("user.found", false))
		return "", repo.UserOut{}, err
	}
	span.SetAttributes(
		attribute.Bool("user.found", true),
		attribute.String("user.id", usr.ID.String()),
		attribute.Bool("user.has_password_hash", usr.PasswordHash != ""),
	)

	if usr.PasswordHash == "" {
		// OIDC-only user — there's no local password to reset. Don't issue a
		// token; the caller decides whether to surface this.
		return "", usr, errResetUserHasNoPassword
	}

	tok := hasher.GenerateTokenCtx(ctx)
	expiresAt := time.Now().Add(passwordResetTokenTTL)

	if _, err := svc.repos.PasswordResetTokens.Create(ctx, usr.ID, tok.Hash, expiresAt); err != nil {
		recordServiceSpanError(span, err)
		return "", usr, err
	}
	return tok.Raw, usr, nil
}

func (svc *UserService) sendResetEmail(usr repo.UserOut, link string) error {
	subject := "Reset your Homebox password"
	body := buildResetEmailBody(usr.Name, link)

	msg := mailer.NewMessageBuilder().
		SetTo(usr.Name, usr.Email).
		SetFrom("Homebox", svc.mailer.From).
		SetSubject(subject).
		SetBody(body).
		Build()

	return svc.mailer.Send(msg)
}

func buildResetLink(baseURL, rawToken string) string {
	base := strings.TrimSuffix(baseURL, "/")
	return fmt.Sprintf("%s/reset-password?token=%s", base, url.QueryEscape(rawToken))
}

func buildResetEmailBody(name, link string) string {
	if name == "" {
		name = "there"
	}
	// Plain HTML; the mailer sends as text/html. The link is also rendered as
	// plain text below in case the user's client mangles the anchor.
	return fmt.Sprintf(`<!doctype html>
<html><body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; line-height: 1.5; color: #1f2937;">
<p>Hi %s,</p>
<p>Someone (hopefully you) requested a password reset for your Homebox account. Click the link below to choose a new password. The link will expire in one hour and can only be used once.</p>
<p><a href="%s" style="display: inline-block; padding: 10px 18px; background: #0ea5e9; color: white; text-decoration: none; border-radius: 6px;">Reset password</a></p>
<p>If the button doesn't work, paste this URL into your browser:</p>
<p style="word-break: break-all;"><code>%s</code></p>
<p>If you didn't request this, you can ignore this email — your password won't change.</p>
<p>— Homebox</p>
</body></html>`, htmlEscape(name), htmlEscape(link), htmlEscape(link))
}

// htmlEscape is a minimal escaper for the few values we interpolate into the
// reset email. We don't pull in html/template here because the body is a fixed
// fragment with two substitutions.
func htmlEscape(s string) string {
	r := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&#39;",
	)
	return r.Replace(s)
}
