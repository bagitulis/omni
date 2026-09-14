package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/extensions"
	"github.com/omni/backend/internal/handlers"
	"github.com/omni/backend/internal/middleware"
)

// RegisterExtensionRoutes registers the extensions API.
//
// Route protection is split deliberately:
//
//   - /pairing/confirm is UNAUTHENTICATED. A browser pairing for the first time
//     has no JWT, so it holds only the pairing code. The tenant is resolved from
//     that code inside the handler. Codes are 40 bits of CSPRNG, single-use, and
//     expire in 5 minutes, which is what keeps this safe.
//
//   - /ws is UNAUTHENTICATED at the HTTP layer for the same reason: a browser
//     cannot set an Authorization header on a WebSocket handshake. It
//     authenticates on the first frame using the pairing token, which keeps the
//     token out of the URL and therefore out of access and proxy logs.
//
//   - everything else requires Auth + Tenant, and fails closed without a tenant.
func RegisterExtensionRoutes(
	router *gin.RouterGroup,
	handler *handlers.ExtensionHandler,
	wsHandler *handlers.ExtensionWSHandler,
	scrapeHandler *handlers.ScrapeHandler,
) {
	group := router.Group("/extensions")

	// Unauthenticated: pairing handshake and socket upgrade.
	group.POST("/pairing/confirm", handler.ConfirmPairing)
	group.GET("/ws", wsHandler.Handle)

	// Authenticated: everything an operator does from the dashboard.
	authed := group.Group("")
	authed.Use(middleware.Auth())
	authed.Use(middleware.Tenant())
	{
		authed.GET("", handler.List)
		authed.POST("/pairing/generate", handler.GeneratePairingCode)
		authed.DELETE("/:extension_id", handler.Unpair)

		if scrapeHandler != nil {
			authed.POST("/scrape", scrapeHandler.Start)
			authed.GET("/scraped-products", scrapeHandler.ListProducts)
		}
	}
}

// ExtensionWSConfigFor builds the WebSocket configuration for the router wiring.
func ExtensionWSConfigFor(svc *extensions.Service, hub *extensions.Hub, allowedOrigins []string) extensions.ConnConfig {
	return handlers.BuildExtensionWSConfig(svc, hub, allowedOrigins)
}
