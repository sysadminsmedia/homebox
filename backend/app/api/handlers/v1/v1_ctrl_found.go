package v1

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/hay-kot/httpkit/errchain"
	"github.com/hay-kot/httpkit/server"
	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"
	"go.opentelemetry.io/otel/attribute"
)

type (
	FoundItemResponse struct {
		ItemID  string `json:"itemId"`
		Message string `json:"message,omitempty"`
		Mode    string `json:"mode"              enums:"form,mailto"`
		Email   string `json:"email,omitempty"`
	}

	// FoundContactRequest's validate tags feed the OpenAPI spec via swaggo;
	// enforcement is manual in HandleFoundContact, and max=2000 must stay in
	// sync with foundContactMaxMessageLen.
	FoundContactRequest struct {
		Message string `json:"message"           validate:"required,max=2000"`
		ReplyTo string `json:"replyTo,omitempty" validate:"omitempty,email"`
	}
)

// foundContactMaxMessageLen bounds the finder's message. The inbound body is
// already capped by the body-size middleware; this keeps the relayed email a
// sane size.
const foundContactMaxMessageLen = 2000

// foundModeForm is returned when the SMTP mailer is configured and the
// frontend should render an in-page contact form; foundModeMailto is the
// fallback where the owner's address is exposed for a mailto: link.
const (
	foundModeForm   = "form"
	foundModeMailto = "mailto"
)

// foundKindItem and foundKindAsset are the two accepted values of the {kind}
// route segment.
const (
	foundKindItem  = "item"
	foundKindAsset = "asset"
)

// errFoundNotFound is the single error produced for every failed resolution:
// unknown kind, malformed ID, missing item, archived item, opted-out group,
// or ambiguous asset ID. Keeping them indistinguishable prevents an
// unauthenticated caller from probing which of those states applies.
var errFoundNotFound = errors.New("found item not found")

// foundKindAttr bounds the route's {kind} value for span attributes so
// arbitrary client input never becomes an attribute value.
func foundKindAttr(kind string) string {
	switch kind {
	case foundKindItem, foundKindAsset:
		return kind
	default:
		return "other"
	}
}

// foundLookup resolves the {kind}/{id} route params to a FoundContact.
// Parse failures and unknown kinds return the same error shape as a failed
// repository lookup so all not-found paths are identical to the client.
func (ctrl *V1Controller) foundLookup(r *http.Request) (repo.FoundContact, error) {
	kind := chi.URLParam(r, "kind")
	id := chi.URLParam(r, "id")

	switch kind {
	case foundKindItem:
		itemID, err := uuid.Parse(id)
		if err != nil {
			return repo.FoundContact{}, errFoundNotFound
		}
		return ctrl.repo.Groups.FoundContactByItemID(r.Context(), itemID)
	case foundKindAsset:
		assetID, ok := repo.ParseAssetID(id)
		// Every entity defaults to asset_id 0, so "000-000" (and anything
		// non-positive) must never resolve; reject before touching the DB.
		if !ok || assetID <= 0 {
			return repo.FoundContact{}, errFoundNotFound
		}
		return ctrl.repo.Groups.FoundContactByAssetID(r.Context(), assetID)
	default:
		return repo.FoundContact{}, errFoundNotFound
	}
}

// sendFoundContact performs the SMTP send in the background, mirroring
// UserService.processResetRequest: fresh root span, errors logged rather than
// returned so the always-204 response never reflects the outcome. Unlike
// forgot-password, the goroutine lives here in the handler rather than in the
// service, deliberately: FoundService.SendContact stays synchronous and
// directly testable.
func (ctrl *V1Controller) sendFoundContact(contact repo.FoundContact, message, replyTo string) {
	_, span := startEntityCtrlSpan(context.Background(), "controller.V1.sendFoundContact",
		attribute.String("item.id", contact.ItemID.String()),
	)
	defer span.End()

	if err := ctrl.svc.Found.SendContact(contact, message, replyTo); err != nil {
		recordCtrlSpanError(span, err)
		span.SetAttributes(attribute.String("found.outcome", "send_failed"))
		log.Err(err).Msg("found-item contact message failed to send")
		return
	}
	span.SetAttributes(attribute.String("found.outcome", "sent"))
}

// HandleFoundGet godoc
//
//	@Summary		Get Found Item Contact Page
//	@Description	Public, unauthenticated lookup for the found-item contact page. Resolves an
//	@Description	item by UUID (kind "item") or by asset ID (kind "asset") when the owning
//	@Description	group has opted in. Returns 404 for anything that does not resolve, with no
//	@Description	distinction between missing, archived, opted-out, or ambiguous.
//	@Tags			Found
//	@Produce		json
//	@Param			kind	path		string	true	"ID kind"	Enums(item, asset)
//	@Param			id		path		string	true	"Item UUID or asset ID"
//	@Success		200		{object}	FoundItemResponse
//	@Failure		404		{string}	string	"item not found or found-item contact not enabled"
//	@Router			/v1/found/{kind}/{id} [GET]
func (ctrl *V1Controller) HandleFoundGet() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		spanCtx, span := startEntityCtrlSpan(r.Context(), "controller.V1.HandleFoundGet",
			attribute.String("found.kind", foundKindAttr(chi.URLParam(r, "kind"))),
		)
		defer span.End()

		contact, err := ctrl.foundLookup(r.WithContext(spanCtx))
		if err != nil {
			// SECURITY: never return the underlying error; a raw ent error
			// would leak schema details and let callers distinguish the
			// not-found cases. The static sentinel keeps every 404 identical.
			// (validate.NewRequestError(nil, ...) is not used because
			// RequestError.Error() dereferences Err unconditionally and the
			// error middleware calls it.)
			span.SetAttributes(attribute.String("found.outcome", "not_found"))
			return validate.NewRequestError(errFoundNotFound, http.StatusNotFound)
		}

		resp := FoundItemResponse{
			ItemID:  contact.ItemID.String(),
			Message: contact.Message,
		}
		if ctrl.svc.Found.MailerReady() {
			resp.Mode = foundModeForm
		} else {
			// SECURITY: the owner's email is exposed only in mailto mode,
			// where it is the sole way for the finder to make contact. When
			// the mailer is configured the address stays server-side.
			resp.Mode = foundModeMailto
			resp.Email = contact.OwnerEmail
		}

		span.SetAttributes(
			attribute.String("found.outcome", "ok"),
			attribute.String("found.mode", resp.Mode),
		)
		return server.JSON(w, http.StatusOK, resp)
	}
}

// HandleFoundContact godoc
//
//	@Summary		Send Found Item Contact Message
//	@Description	Relays a finder's message to the item owner by email. After basic input
//	@Description	validation this always returns 204, whether or not the item resolved or
//	@Description	the email was sent, so the endpoint cannot be used to probe for items.
//	@Tags			Found
//	@Accept			application/json
//	@Produce		json
//	@Param			kind	path	string				true	"ID kind"	Enums(item, asset)
//	@Param			id		path	string				true	"Item UUID or asset ID"
//	@Param			payload	body	FoundContactRequest	true	"Message"
//	@Success		204
//	@Failure		400	{string}	string	"invalid request body, empty or oversized message, or invalid reply-to address"
//	@Router			/v1/found/{kind}/{id}/contact [POST]
func (ctrl *V1Controller) HandleFoundContact() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		spanCtx, span := startEntityCtrlSpan(r.Context(), "controller.V1.HandleFoundContact",
			attribute.String("found.kind", foundKindAttr(chi.URLParam(r, "kind"))),
		)
		defer span.End()

		var body FoundContactRequest
		if err := server.Decode(r, &body); err != nil {
			span.SetAttributes(attribute.String("found.outcome", "decode_failed"))
			return validate.NewRequestError(err, http.StatusBadRequest)
		}

		body.Message = strings.TrimSpace(body.Message)
		if body.Message == "" {
			span.SetAttributes(attribute.String("found.outcome", "missing_message"))
			return validate.NewRequestError(errors.New("message is required"), http.StatusBadRequest)
		}
		if len(body.Message) > foundContactMaxMessageLen {
			span.SetAttributes(attribute.String("found.outcome", "message_too_long"))
			return validate.NewRequestError(
				fmt.Errorf("message is too long (max %d)", foundContactMaxMessageLen),
				http.StatusBadRequest,
			)
		}

		body.ReplyTo = strings.TrimSpace(body.ReplyTo)
		if body.ReplyTo != "" {
			if _, err := mail.ParseAddress(body.ReplyTo); err != nil {
				span.SetAttributes(attribute.String("found.outcome", "invalid_reply_to"))
				return validate.NewRequestError(errors.New("replyTo must be a valid email address"), http.StatusBadRequest)
			}
		}

		// SECURITY: From here on the response is ALWAYS 204, mirroring
		// HandleForgotPassword. Whether the ID resolved, whether the group
		// opted in, and whether the mailer is configured or the send
		// succeeded are all server state an unauthenticated caller must not
		// be able to probe. Failures are logged for operators instead. The
		// mailer-readiness check below MUST NOT change the response either;
		// on a mailto-mode instance a direct POST is simply dropped. The
		// SMTP send runs in a background goroutine (same as the
		// forgot-password flow), so response timing is uniform for hit and
		// miss: both paths do one lookup and return immediately, and an
		// attacker cannot time-distinguish resolved items from unresolved
		// ones.
		if !ctrl.svc.Found.MailerReady() {
			span.SetAttributes(attribute.String("found.outcome", "mailer_not_configured"))
			log.Warn().Msg("found-item contact posted but SMTP mailer is not configured; no email will be sent")
			return server.JSON(w, http.StatusNoContent, nil)
		}

		contact, err := ctrl.foundLookup(r.WithContext(spanCtx))
		if err != nil {
			span.SetAttributes(attribute.String("found.outcome", "not_found"))
			return server.JSON(w, http.StatusNoContent, nil)
		}

		// Per-item send cap, keyed on the resolved item ID so the bound holds
		// regardless of source IP (the per-IP foundLimiter middleware cannot
		// stop distributed mailbombing of one owner). Gated only after a
		// successful lookup to preserve the uniform 204/404 behavior, and the
		// response stays 204 when denied so the cap is never observable to the
		// caller (anti-probing).
		if ctrl.foundSendLimiter != nil && !ctrl.foundSendLimiter.Allow(contact.ItemID.String()) {
			span.SetAttributes(attribute.String("found.outcome", "send_rate_limited"))
			log.Warn().Str("item.id", contact.ItemID.String()).Msg("found-item contact send rate limit exceeded; dropping message")
			return server.JSON(w, http.StatusNoContent, nil)
		}

		// Detached from the request context so a client disconnect doesn't
		// abort the send; SendContact takes no context, and the goroutine
		// starts its own root span. Errors are logged, never propagated —
		// the response is already committed to 204.
		go ctrl.sendFoundContact(contact, body.Message, body.ReplyTo)

		span.SetAttributes(attribute.String("found.outcome", "queued"))
		return server.JSON(w, http.StatusNoContent, nil)
	}
}
