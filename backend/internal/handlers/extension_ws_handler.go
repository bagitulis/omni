package handlers

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/extensions"
)

// ExtensionWSHandler upgrades a request to the extensions WebSocket.
//
// This handler sits before Auth/Tenant middleware because a browser cannot send
// an Authorization header on a WebSocket handshake. Authentication happens on
// the first frame instead (the token in the connect message), which keeps the
// token out of the URL and therefore out of access logs and proxy logs.
type ExtensionWSHandler struct {
	cfg extensions.ConnConfig
}

// NewExtensionWSHandler creates the WebSocket handler.
func NewExtensionWSHandler(cfg extensions.ConnConfig) *ExtensionWSHandler {
	return &ExtensionWSHandler{cfg: cfg}
}

// Handle upgrades and serves the connection.
func (h *ExtensionWSHandler) Handle(c *gin.Context) {
	h.cfg.ServeUpgrade(c.Writer, c.Request)
}

// BuildExtensionWSConfig wires the connection dependencies to the service.
//
// The authenticate callback maps a pairing token to a tenant-scoped identity.
// The service owns that lookup, so the handler layer stays free of tenancy
// logic.
func BuildExtensionWSConfig(svc *extensions.Service, hub *extensions.Hub, allowedOrigins []string) extensions.ConnConfig {
	return extensions.ConnConfig{
		Hub:            hub,
		AllowedOrigins: allowedOrigins,
		Authenticate: func(ctx context.Context, req extensions.ConnectRequest) (*extensions.ConnectionIdentity, error) {
			identity, err := svc.Authenticate(ctx, req)
			if err != nil {
				return nil, err
			}
			// Record liveness so the dashboard reflects reality rather than the
			// last persisted state. A failure here must not reject the socket:
			// the extension is authenticated and usable either way.
			if markErr := svc.MarkConnected(ctx, identity.TenantID, identity.ExtensionID); markErr != nil {
				log.Printf("extensions: mark connected for %s: %v", identity.ExtensionID, markErr)
			}
			return identity, nil
		},
		OnDisconnect: func(extensionID string) {
			// The tenant is not available here, so the status is refreshed on
			// the next connect and by the startup reset. Logged for operators.
			log.Printf("extensions: %s disconnected", extensionID)
		},
	}
}
