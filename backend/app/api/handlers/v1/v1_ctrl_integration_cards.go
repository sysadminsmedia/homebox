package v1

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hay-kot/httpkit/errchain"
	"github.com/hay-kot/httpkit/server"
	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"
	"github.com/sysadminsmedia/homebox/backend/internal/web/adapters"
	"go.opentelemetry.io/otel/attribute"
)

type integrationCardOut struct {
	AttachmentID uuid.UUID              `json:"attachmentId"`
	Provider     string                 `json:"provider"`
	Scope        string                 `json:"scope"`
	Title        string                 `json:"title"`
	OpenURL      string                 `json:"openUrl"`
	ThumbnailURL string                 `json:"thumbnailUrl,omitempty"`
	State        string                 `json:"state"`
	Error        string                 `json:"error,omitempty"`
	Fields       map[string]interface{} `json:"fields,omitempty"`
}

type integrationLabel struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type integrationTag struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Color     string `json:"color"`
	TextColor string `json:"textColor"`
}

type cachedIntegrationLabel struct {
	label integrationLabel
	ok    bool
}

type cachedIntegrationTag struct {
	tag integrationTag
	ok  bool
}

type paperlessCardCache struct {
	correspondents map[int]cachedIntegrationLabel
	documentTypes  map[int]cachedIntegrationLabel
	tags           map[int]cachedIntegrationTag
}

type paperlessRef struct {
	docID   string
	openURL string
	scope   string
}

type paperlessConfig struct {
	baseURL string
	token   string
	scope   string
}

var paperlessEndpointPattern = regexp.MustCompile(`^/(?:documents/(\d+)(?:/details)?|api/documents/(\d+)(?:/(?:preview|download))?)/?$`)

const integrationCardBuildTimeout = 15 * time.Second

const (
	thumbnailFallbackContentType = "application/octet-stream"
	thumbnailPNGContentType      = "image/png"
)

func newPaperlessCardCache() *paperlessCardCache {
	return &paperlessCardCache{
		correspondents: make(map[int]cachedIntegrationLabel),
		documentTypes:  make(map[int]cachedIntegrationLabel),
		tags:           make(map[int]cachedIntegrationTag),
	}
}

func (ctrl *V1Controller) integrationHTTPClient() *http.Client {
	transport := ctrl.integrationTransport
	if transport == nil {
		transport = validate.NewOutboundHTTPTransport(&ctrl.config.Notifier)
	}

	return &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
		CheckRedirect: func(req *http.Request, _ []*http.Request) error {
			if err := validate.ValidateOutboundHTTPURLWithContext(req.Context(), req.URL.String(), &ctrl.config.Notifier); err != nil {
				return fmt.Errorf("integration redirect blocked: %w", err)
			}
			return nil
		},
	}
}

func integrationScope(provider, baseURL string) string {
	sum := sha256.Sum256([]byte(strings.TrimRight(strings.TrimSpace(baseURL), "/")))
	return provider + ":" + hex.EncodeToString(sum[:8])
}

func paperlessConfigFromSettings(settings map[string]interface{}) (paperlessConfig, bool) {
	baseURL, _ := settings["paperless_url"].(string)
	token, _ := settings["paperless_token"].(string)
	if strings.TrimSpace(baseURL) == "" {
		return paperlessConfig{}, false
	}
	return paperlessConfig{
		baseURL: baseURL,
		token:   token,
		scope:   integrationScope("paperless", baseURL),
	}, true
}

func parseHTTPURL(raw string) (*url.URL, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return nil, err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("URL must use http or https")
	}
	if u.Host == "" || u.User != nil {
		return nil, fmt.Errorf("URL must include a host and no userinfo")
	}
	return u, nil
}

func parsePaperlessAttachmentLink(rawURL string, cfg paperlessConfig) (paperlessRef, bool) {
	target, targetErr := parseHTTPURL(rawURL)
	base, baseErr := parseHTTPURL(cfg.baseURL)
	if targetErr != nil || baseErr != nil {
		return paperlessRef{}, false
	}
	if !strings.EqualFold(target.Scheme, base.Scheme) || !strings.EqualFold(target.Host, base.Host) {
		return paperlessRef{}, false
	}

	basePath := strings.TrimRight(base.Path, "/")
	var relPath string
	switch {
	case basePath == "":
		relPath = target.Path
	case target.Path == basePath:
		relPath = "/"
	case strings.HasPrefix(target.Path, basePath+"/"):
		relPath = strings.TrimPrefix(target.Path, basePath)
	default:
		return paperlessRef{}, false
	}

	matches := paperlessEndpointPattern.FindStringSubmatch(relPath)
	if matches == nil {
		return paperlessRef{}, false
	}
	docID := matches[1]
	if docID == "" {
		docID = matches[2]
	}

	endpoint := relPath
	if target.RawQuery != "" {
		endpoint += "?" + target.RawQuery
	}
	if target.Fragment != "" {
		endpoint += "#" + target.Fragment
	}

	return paperlessRef{
		docID:   docID,
		openURL: fmt.Sprintf("%s://%s%s%s", base.Scheme, base.Host, basePath, endpoint),
		scope:   cfg.scope,
	}, true
}

func paperlessAPIURL(baseURL, endpoint string) (string, error) {
	base, err := parseHTTPURL(baseURL)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s://%s%s%s", base.Scheme, base.Host, strings.TrimRight(base.Path, "/"), endpoint), nil
}

func (ctrl *V1Controller) paperlessRequest(ctx services.Context, cfg paperlessConfig, endpoint string) (*http.Response, error) {
	if strings.TrimSpace(cfg.baseURL) == "" || strings.TrimSpace(cfg.token) == "" {
		return nil, fmt.Errorf("paperless integration is not configured")
	}

	upstream, err := paperlessAPIURL(cfg.baseURL, endpoint)
	if err != nil {
		return nil, err
	}
	// Use the shared operator-controlled outbound policy as-is. Self-hosted
	// Paperless instances commonly live on private networks, so integrations do
	// not impose stricter defaults than the configured outbound policy.
	if err := validate.ValidateOutboundHTTPURLWithContext(ctx.Context, upstream, &ctrl.config.Notifier); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx.Context, http.MethodGet, upstream, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Token "+cfg.token)
	return ctrl.integrationHTTPClient().Do(req)
}

func (ctrl *V1Controller) paperlessJSON(ctx services.Context, cfg paperlessConfig, endpoint string, target interface{}) error {
	resp, err := ctrl.paperlessRequest(ctx, cfg, endpoint)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("upstream returned %d", resp.StatusCode)
	}

	const maxJSONResponseSize int64 = 2 * 1024 * 1024
	return json.NewDecoder(io.LimitReader(resp.Body, maxJSONResponseSize)).Decode(target)
}

func (ctrl *V1Controller) paperlessCachedLabel(ctx services.Context, cfg paperlessConfig, cache map[int]cachedIntegrationLabel, endpoint string, id int) (integrationLabel, bool) {
	if cached, ok := cache[id]; ok {
		return cached.label, cached.ok
	}

	var label integrationLabel
	err := ctrl.paperlessJSON(ctx, cfg, fmt.Sprintf(endpoint, id), &label)
	cached := cachedIntegrationLabel{label: label, ok: err == nil}
	cache[id] = cached
	return cached.label, cached.ok
}

func (ctrl *V1Controller) paperlessCachedTag(ctx services.Context, cfg paperlessConfig, cache map[int]cachedIntegrationTag, id int) (integrationTag, bool) {
	if cached, ok := cache[id]; ok {
		return cached.tag, cached.ok
	}

	var raw struct {
		ID        int    `json:"id"`
		Name      string `json:"name"`
		Color     string `json:"color"`
		TextColor string `json:"text_color"`
	}
	err := ctrl.paperlessJSON(ctx, cfg, fmt.Sprintf("/api/tags/%d/", id), &raw)
	tag := integrationTag{
		ID:        raw.ID,
		Name:      raw.Name,
		Color:     raw.Color,
		TextColor: raw.TextColor,
	}
	cached := cachedIntegrationTag{tag: tag, ok: err == nil}
	cache[id] = cached
	return cached.tag, cached.ok
}

func (ctrl *V1Controller) buildPaperlessCard(ctx services.Context, cfg paperlessConfig, entityID uuid.UUID, attachment repo.ItemAttachment, ref paperlessRef, cache *paperlessCardCache) integrationCardOut {
	card := integrationCardOut{
		AttachmentID: attachment.ID,
		Provider:     "paperless",
		Scope:        ref.scope,
		Title:        attachment.Title,
		OpenURL:      ref.openURL,
		State:        "loading",
	}
	var raw struct {
		ID            int    `json:"id"`
		Title         string `json:"title"`
		CreatedDate   string `json:"created_date"`
		PageCount     int    `json:"page_count"`
		Correspondent int    `json:"correspondent"`
		DocumentType  int    `json:"document_type"`
		Tags          []int  `json:"tags"`
	}
	if err := ctrl.paperlessJSON(ctx, cfg, fmt.Sprintf("/api/documents/%s/", ref.docID), &raw); err != nil {
		card.State = "error"
		card.Error = err.Error()
		return card
	}

	card.State = "ok"
	card.ThumbnailURL = fmt.Sprintf("/entities/%s/attachments/%s/integration-thumbnail?scope=%s", entityID, attachment.ID, url.QueryEscape(ref.scope))
	if raw.Title != "" {
		card.Title = raw.Title
	}
	card.Fields = map[string]interface{}{
		"createdDate": raw.CreatedDate,
		"pageCount":   raw.PageCount,
	}

	if raw.Correspondent != 0 {
		if c, ok := ctrl.paperlessCachedLabel(ctx, cfg, cache.correspondents, "/api/correspondents/%d/", raw.Correspondent); ok {
			card.Fields["correspondent"] = c
		}
	}
	if raw.DocumentType != 0 {
		if d, ok := ctrl.paperlessCachedLabel(ctx, cfg, cache.documentTypes, "/api/document_types/%d/", raw.DocumentType); ok {
			card.Fields["documentType"] = d
		}
	}

	tags := make([]integrationTag, 0, len(raw.Tags))
	for _, tagID := range raw.Tags {
		if tag, ok := ctrl.paperlessCachedTag(ctx, cfg, cache.tags, tagID); ok {
			tags = append(tags, tag)
		}
	}
	if len(tags) > 0 {
		card.Fields["tags"] = tags
	}

	return card
}

// HandleEntityAttachmentIntegrationCards returns derived cards for link attachments
// that match configured integrations. The attachment itself remains a normal link.
func (ctrl *V1Controller) HandleEntityAttachmentIntegrationCards() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		id, err := adapters.RouteUUID(r, "id")
		if err != nil {
			return err
		}

		spanCtx, span := startEntityCtrlSpan(r.Context(), "controller.V1.HandleEntityAttachmentIntegrationCards")
		defer span.End()
		ctx := services.NewContext(spanCtx)

		entity, err := ctrl.repo.Entities.GetOneByGroup(ctx.Context, ctx.GID, id)
		if err != nil {
			return validate.NewRequestError(err, http.StatusNotFound)
		}

		settings, err := ctrl.svc.User.GetSettings(ctx.Context, services.UseUserCtx(ctx.Context).ID)
		if err != nil {
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}
		cfg, ok := paperlessConfigFromSettings(settings)
		if !ok {
			return server.JSON(w, http.StatusOK, Results[integrationCardOut]{Items: []integrationCardOut{}})
		}

		cardCtx, cancel := context.WithTimeout(ctx.Context, integrationCardBuildTimeout)
		defer cancel()
		cardSvcCtx := services.NewContext(cardCtx)
		cache := newPaperlessCardCache()
		cards := make([]integrationCardOut, 0)
		for _, attachment := range entity.Attachments {
			if attachment.MimeType != repo.MimeTypeLinkURL && attachment.MimeType != repo.MimeTypePaperlessDocument {
				continue
			}
			ref, ok := parsePaperlessAttachmentLink(attachment.Path, cfg)
			if !ok {
				continue
			}
			if cardCtx.Err() != nil {
				break
			}
			cards = append(cards, ctrl.buildPaperlessCard(cardSvcCtx, cfg, id, attachment, ref, cache))
		}

		span.SetAttributes(attribute.Int("integration.cards.count", len(cards)))
		return server.JSON(w, http.StatusOK, Results[integrationCardOut]{Items: cards})
	}
}

func thumbnailContentType(contentType string) string {
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return thumbnailFallbackContentType
	}

	switch strings.ToLower(mediaType) {
	case "image/avif", "image/bmp", "image/gif", "image/jpeg", thumbnailPNGContentType, "image/webp":
		return mediaType
	default:
		return thumbnailFallbackContentType
	}
}

func (ctrl *V1Controller) HandleEntityAttachmentIntegrationThumbnail() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		entityID, err := adapters.RouteUUID(r, "id")
		if err != nil {
			return err
		}
		attachmentID, err := adapters.RouteUUID(r, "attachment_id")
		if err != nil {
			return err
		}

		ctx := services.NewContext(r.Context())
		entity, err := ctrl.repo.Entities.GetOneByGroup(ctx.Context, ctx.GID, entityID)
		if err != nil {
			return validate.NewRequestError(err, http.StatusNotFound)
		}

		var attachment repo.ItemAttachment
		found := false
		for _, candidate := range entity.Attachments {
			if candidate.ID == attachmentID {
				attachment = candidate
				found = true
				break
			}
		}
		if !found {
			return validate.NewRequestError(fmt.Errorf("attachment not found"), http.StatusNotFound)
		}

		settings, err := ctrl.svc.User.GetSettings(ctx.Context, services.UseUserCtx(ctx.Context).ID)
		if err != nil {
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}
		cfg, configured := paperlessConfigFromSettings(settings)
		if !configured {
			return validate.NewRequestError(fmt.Errorf("paperless integration is not configured"), http.StatusNotFound)
		}
		ref, ok := parsePaperlessAttachmentLink(attachment.Path, cfg)
		if !ok {
			return validate.NewRequestError(fmt.Errorf("attachment is not a configured Paperless link"), http.StatusNotFound)
		}
		if scope := r.URL.Query().Get("scope"); scope != "" && scope != ref.scope {
			return validate.NewRequestError(fmt.Errorf("integration configuration changed"), http.StatusNotFound)
		}

		resp, err := ctrl.paperlessRequest(ctx, cfg, fmt.Sprintf("/api/documents/%s/thumb/", ref.docID))
		if err != nil {
			return validate.NewRequestError(err, http.StatusBadGateway)
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode >= 400 {
			return validate.NewRequestError(fmt.Errorf("upstream returned %d", resp.StatusCode), http.StatusBadGateway)
		}
		w.Header().Set("Content-Type", thumbnailContentType(resp.Header.Get("Content-Type")))
		w.Header().Set("X-Content-Type-Options", "nosniff")
		_, err = io.Copy(w, resp.Body)
		if err != nil {
			log.Err(err).Msg("failed to stream integration thumbnail")
		}
		return nil
	}
}
