// Package centrifuge provides a Centrifuge broker for real-time event broadcasting.
package centrifuge

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/centrifugal/centrifuge"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services/reporting/eventbus"
)

const (
	// Channel format: {tenantID}:mutations
	channelFormat = "%s:mutations"

	// History configuration
	historySize = 100
	historyTTL  = 5 * time.Minute
)

type EventMessage struct {
	Event string `json:"event"`
}

type Broker struct {
	node   *centrifuge.Node
	secret string
	bus    *eventbus.EventBus
}

// New creates a new Centrifuge broker instance.
func New(secret string, bus *eventbus.EventBus) (*Broker, error) {
	if secret == "" {
		return nil, fmt.Errorf("centrifuge secret cannot be empty")
	}

	node, err := centrifuge.New(centrifuge.Config{
		LogLevel:   centrifuge.LogLevelError,
		LogHandler: handleLog,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create centrifuge node: %w", err)
	}

	broker := &Broker{
		node:   node,
		secret: secret,
		bus:    bus,
	}

	// Configure the node
	node.OnConnecting(broker.onConnecting)
	node.OnConnect(broker.onConnect)

	return broker, nil
}

// handleLog is a custom log handler for Centrifuge.
func handleLog(entry centrifuge.LogEntry) {
	switch entry.Level {
	case centrifuge.LogLevelError:
		log.Error().Msg(entry.Message)
	case centrifuge.LogLevelWarn:
		log.Warn().Msg(entry.Message)
	case centrifuge.LogLevelInfo:
		log.Info().Msg(entry.Message)
	case centrifuge.LogLevelDebug:
		log.Debug().Msg(entry.Message)
	}
}

// onConnecting validates the JWT token.
func (b *Broker) onConnecting(ctx context.Context, event centrifuge.ConnectEvent) (centrifuge.ConnectReply, error) {
	log.Debug().Msg("onConnecting called")

	if event.Token == "" {
		log.Warn().Msg("empty token received in onConnecting")
		return centrifuge.ConnectReply{}, centrifuge.ErrorTokenExpired
	}

	// Parse and validate JWT token
	token, err := jwt.Parse(event.Token, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Warn().Msg("unexpected signing method")
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(b.secret), nil
	})

	if err != nil {
		log.Warn().Err(err).Msg("failed to parse JWT token")
		return centrifuge.ConnectReply{}, centrifuge.ErrorTokenExpired
	}

	if !token.Valid {
		log.Warn().Msg("invalid JWT token")
		return centrifuge.ConnectReply{}, centrifuge.ErrorTokenExpired
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Warn().Msg("invalid JWT claims type")
		return centrifuge.ConnectReply{}, centrifuge.ErrorTokenExpired
	}

	log.Debug().Msg("parsed JWT claims")

	// Extract user ID and tenant ID from claims
	userIDStr, ok := claims["sub"].(string)
	if !ok {
		log.Warn().Msg("missing or invalid user ID in token")
		return centrifuge.ConnectReply{}, centrifuge.ErrorTokenExpired
	}

	tenantIDStr, ok := claims["tenant"].(string)
	if !ok {
		log.Warn().Msg("missing or invalid tenant ID in token")
		return centrifuge.ConnectReply{}, centrifuge.ErrorTokenExpired
	}

	log.Debug().
		Str("user_id", userIDStr).
		Str("tenant_id", tenantIDStr).
		Msg("successfully validated JWT token")

	// Store tenant ID in credentials for subscription validation
	return centrifuge.ConnectReply{
		Credentials: &centrifuge.Credentials{
			UserID: userIDStr,
			Info:   []byte(fmt.Sprintf(`{"tenant":"%s"}`, tenantIDStr)),
		},
		Subscriptions: map[string]centrifuge.SubscribeOptions{
			fmt.Sprintf(channelFormat, tenantIDStr): {
				EnableRecovery: true,
			},
		},
	}, nil
}

// onConnect is called when a client successfully connects.
func (b *Broker) onConnect(client *centrifuge.Client) {
	// Log transport type in debug mode
	transport := client.Transport()
	log.Debug().
		Str("user", client.UserID()).
		Str("transport", transport.Name()).
		Msg("client connected")

	client.OnSubscribe(func(e centrifuge.SubscribeEvent, cb centrifuge.SubscribeCallback) {
		// Extract tenant from client info
		var info map[string]string
		if err := json.Unmarshal(client.Info(), &info); err != nil {
			log.Debug().Err(err).Msg("failed to parse client info")
			cb(centrifuge.SubscribeReply{}, centrifuge.ErrorPermissionDenied)
			return
		}

		tenantID, ok := info["tenant"]
		if !ok {
			log.Debug().Msg("tenant not found in client info")
			cb(centrifuge.SubscribeReply{}, centrifuge.ErrorPermissionDenied)
			return
		}

		// Validate channel format: {tenantID}:mutations
		expectedChannel := fmt.Sprintf(channelFormat, tenantID)
		if e.Channel != expectedChannel {
			log.Debug().
				Str("expected", expectedChannel).
				Str("actual", e.Channel).
				Msg("invalid channel subscription attempt")
			cb(centrifuge.SubscribeReply{}, centrifuge.ErrorPermissionDenied)
			return
		}

		log.Debug().
			Str("user", client.UserID()).
			Str("channel", e.Channel).
			Msg("client subscribed to channel")

		cb(centrifuge.SubscribeReply{
			Options: centrifuge.SubscribeOptions{
				EnableRecovery: true,
				RecoveryMode:   centrifuge.RecoveryModeCache,
			},
		}, nil)
	})
}

// Run starts the broker and subscribes to event bus events.
func (b *Broker) Run(ctx context.Context) error {
	if err := b.node.Run(); err != nil {
		return fmt.Errorf("failed to run centrifuge node: %w", err)
	}

	// Subscribe to event bus events
	b.subscribeToEvents()

	log.Info().Msg("centrifuge broker started")

	<-ctx.Done()

	log.Info().Msg("shutting down centrifuge broker")
	return b.node.Shutdown(context.Background())
}

// subscribeToEvents subscribes to all event bus events and publishes to Centrifuge.
func (b *Broker) subscribeToEvents() {
	factory := func(eventName string) func(data any) {
		return func(data any) {
			eventData, ok := data.(eventbus.GroupMutationEvent)
			if !ok {
				log.Debug().Msgf("invalid event data: %v", data)
				return
			}

			msg := &EventMessage{Event: eventName}
			jsonBytes, err := json.Marshal(msg)
			if err != nil {
				log.Error().Err(err).Msgf("error marshaling event data %v", data)
				return
			}

			// Publish to tenant-specific channel
			channel := fmt.Sprintf(channelFormat, eventData.GID.String())

			_, err = b.node.Publish(channel, jsonBytes, centrifuge.WithHistory(historySize, historyTTL))
			if err != nil {
				log.Error().Err(err).
					Str("channel", channel).
					Str("event", eventName).
					Msg("failed to publish event")
			} else {
				log.Debug().
					Str("channel", channel).
					Str("event", eventName).
					Msg("published event")
			}
		}
	}

	b.bus.Subscribe(eventbus.EventTagMutation, factory("tag.mutation"))
	b.bus.Subscribe(eventbus.EventLocationMutation, factory("location.mutation"))
	b.bus.Subscribe(eventbus.EventItemMutation, factory("item.mutation"))

	log.Info().Msg("subscribed to event bus events")
}

// HTTPHandler returns the HTTP handler supporting multiple transports (WebSocket, SSE, HTTP Streaming).
// Auto-negotiation allows clients to automatically fall back to alternative transports when WebSocket fails.
func (b *Broker) HTTPHandler() http.Handler {
	// Shared configuration for all transports
	checkOrigin := func(r *http.Request) bool {
		// Allow all origins for all transports
		// The JWT token validation provides the actual security
		log.Debug().
			Str("origin", r.Header.Get("Origin")).
			Str("host", r.Host).
			Str("user_agent", r.Header.Get("User-Agent")).
			Msg("transport origin check")
		return true
	}

	// Create handlers for each transport
	wsHandler := centrifuge.NewWebsocketHandler(b.node, centrifuge.WebsocketConfig{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     checkOrigin,
	})

	sseHandler := centrifuge.NewSSEHandler(b.node, centrifuge.SSEConfig{
		MaxRequestBodySize: 64 * 1024, // 64KB
	})

	httpStreamHandler := centrifuge.NewHTTPStreamHandler(b.node, centrifuge.HTTPStreamConfig{
		MaxRequestBodySize: 64 * 1024, // 64KB
	})

	// Create emulate handler for SSE/HTTP Stream command endpoint
	emulateHandler := centrifuge.NewEmulationHandler(b.node, centrifuge.EmulationConfig{
		MaxRequestBodySize: 64 * 1024, // 64KB
	})

	// Create a multiplexer to route requests to appropriate handlers
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle CORS for SSE/HTTP Stream
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			if headers := r.Header.Get("Access-Control-Request-Headers"); headers != "" {
				w.Header().Set("Access-Control-Allow-Headers", headers)
			}
			w.Header().Set("Access-Control-Max-Age", "86400")
			// Override potentially restrictive security headers set by middleware
			w.Header().Set("Content-Origin-Resource-Policy", "cross-origin")
		}

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Check for POST requests - these are emulation commands for SSE/HTTP Stream
		if r.Method == http.MethodPost {
			log.Debug().
				Str("transport", "emulate").
				Str("path", r.URL.Path).
				Msg("routing POST request to emulation handler")
			emulateHandler.ServeHTTP(w, r)
			return
		}

		// Check Upgrade header for WebSocket
		if r.Header.Get("Upgrade") == "websocket" {
			log.Debug().Str("transport", "websocket").Msg("routing to WebSocket handler")
			wsHandler.ServeHTTP(w, r)
			return
		}

		// Check Accept header for SSE
		accept := r.Header.Get("Accept")
		if strings.Contains(accept, "text/event-stream") {
			log.Debug().Str("transport", "sse").Msg("routing to SSE handler")
			sseHandler.ServeHTTP(w, r)
			return
		}

		// Default to HTTP streaming
		log.Debug().Str("transport", "http_stream").Msg("routing to HTTP Stream handler")
		httpStreamHandler.ServeHTTP(w, r)
	})
}

// GenerateToken generates a JWT token for Centrifuge connection.
func (b *Broker) GenerateToken(userID uuid.UUID, tenantID uuid.UUID, expiration time.Duration) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":    userID.String(),
		"tenant": tenantID.String(),
		"iat":    now.Unix(),
		"exp":    now.Add(expiration).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(b.secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}
