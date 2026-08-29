// Package notifier is the one outbound path for notifier delivery, so URL
// validation and the hardened HTTP client can't be used apart. Both call sites
// used to validate and then hand the URL to shoutrrr.Send, which builds its own
// unguarded client — the policy covered the saved URL and nothing after it.
package notifier

import (
	"fmt"
	"net/http"
	"time"

	"github.com/nicholas-fedor/shoutrrr"
	"github.com/nicholas-fedor/shoutrrr/pkg/types"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"
)

// For services the SSRF policy doesn't cover. shoutrrr's fallback client has no
// timeout, so a hung receiver would pin the maintenance job indefinitely.
const plainTimeout = 30 * time.Second

// Sender holds two clients because the policy is scoped to generic webhooks. A
// generic:// URL can name any host, so it gets the hardened client. Everything
// else points at a fixed vendor endpoint or a self-hosted instance the operator
// picked, and gets a plain client with a timeout — applying the network rules
// there would break e.g. a LAN Gotify on a bogon address for no benefit, and
// notifier.* is documented as covering generic notifiers only.
//
// Reuse a Sender across a batch to keep connection pooling; one per send is fine
// too.
type Sender struct {
	cfg      *config.NotifierConf
	guarded  *http.Client
	fallback *http.Client
}

// NewSender builds a Sender for the given notifier policy.
func NewSender(cfg *config.NotifierConf) *Sender {
	return &Sender{
		cfg:      cfg,
		guarded:  validate.NotifierHTTPClient(cfg),
		fallback: &http.Client{Timeout: plainTimeout},
	}
}

// Send validates rawURL and, if it passes, delivers through the right client.
// A rejected URL comes back as *ValidationError so callers can tell "refused to
// send" from "tried and failed".
func (s *Sender) Send(rawURL, message string) error {
	if err := validate.ValidateNotifierURL(rawURL, s.cfg); err != nil {
		return &ValidationError{URL: rawURL, Err: err}
	}
	return s.deliver(s.clientFor(rawURL), rawURL, message)
}

// clientFor picks the hardened client for generic webhooks, plain otherwise.
func (s *Sender) clientFor(rawURL string) *http.Client {
	if validate.IsGenericNotifier(rawURL) {
		return s.guarded
	}
	return s.fallback
}

// deliver sends with an explicit client, injected via SenderOptions rather than
// by mutating http.DefaultClient. Since v0.17.1 the generic service builds a
// fresh http.Client per send when none is supplied, so a client configured
// anywhere else is never consulted.
func (s *Sender) deliver(client *http.Client, rawURL, message string) error {
	sender, err := shoutrrr.CreateSenderWithOptions(
		types.SenderOptions{HTTPClient: client},
		rawURL,
	)
	if err != nil {
		return fmt.Errorf("creating notifier sender: %w", err)
	}

	// One error per service URL; iterate so a future multi-URL caller can't
	// silently drop failures.
	for _, sendErr := range sender.Send(message, nil) {
		if sendErr != nil {
			return sendErr
		}
	}
	return nil
}

// Send is the one-shot form of Sender.Send.
func Send(cfg *config.NotifierConf, rawURL, message string) error {
	return NewSender(cfg).Send(rawURL, message)
}

// ValidationError means the URL was rejected by policy before any request.
type ValidationError struct {
	URL string
	Err error
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("notifier URL failed validation: %v", e.Err)
}

func (e *ValidationError) Unwrap() error { return e.Err }
