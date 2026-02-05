package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/handlers"
	"github.com/omni/backend/internal/middleware"
)

// RegisterAuthRoutes registers authentication routes
func RegisterAuthRoutes(router *gin.RouterGroup, authHandler *handlers.AuthHandler) {
	auth := router.Group("/auth")
	{
		// Public auth routes (no middleware)
		auth.POST("/login", authHandler.Login)
		auth.POST("/register", authHandler.Register)
		auth.POST("/refresh", authHandler.RefreshToken)
		auth.GET("/verify", authHandler.VerifyToken)

		// Dev-only routes (ONLY work when GO_ENV != "production")
		auth.GET("/dev-info", authHandler.DevLoginInfo)
		auth.POST("/dev-login", authHandler.DevLogin)
	}
}

// RegisterProtectedAuthRoutes registers protected auth routes
func RegisterProtectedAuthRoutes(router *gin.RouterGroup, authHandler *handlers.AuthHandler) {
	auth := router.Group("/auth")
	auth.Use(middleware.Auth())
	{
		auth.POST("/logout", authHandler.Logout)
		auth.POST("/change-password", authHandler.ChangePassword)
		auth.GET("/me", authHandler.GetCurrentUser)
		auth.GET("/tenants", authHandler.GetTenants)
		auth.POST("/switch-tenant", authHandler.SwitchTenant)
	}
}

// RegisterUserRoutes registers user management routes
func RegisterUserRoutes(router *gin.RouterGroup, userHandler *handlers.UserHandler) {
	users := router.Group("/users")
	users.Use(middleware.Auth())
	users.Use(middleware.Tenant())
	{
		users.GET("", userHandler.ListUsers)
		users.POST("", userHandler.CreateUser)
		users.GET("/:id", userHandler.GetUser)
		users.PUT("/:id", userHandler.UpdateUser)
		users.DELETE("/:id", userHandler.DeleteUser)
		users.POST("/:id/unlock", userHandler.UnlockUser)
	}
}

// RegisterAuditRoutes registers audit log routes
func RegisterAuditRoutes(router *gin.RouterGroup, auditHandler *handlers.AuditHandler) {
	audit := router.Group("/audit")
	audit.Use(middleware.Auth())
	audit.Use(middleware.Tenant())
	{
		audit.GET("", auditHandler.GetAuditLogs)
		audit.GET("/logs/tenant", auditHandler.GetAuditLogs) // Alias for frontend
		audit.GET("/user/:userId", auditHandler.GetAuditLogsByUser)
		audit.GET("/action/:action", auditHandler.GetAuditLogsByAction)
		audit.GET("/range", auditHandler.GetAuditLogsByDateRange)
	}
}

// RegisterCaptchaRoutes registers captcha routes
func RegisterCaptchaRoutes(router *gin.RouterGroup, captchaHandler *handlers.CaptchaHandler) {
	captcha := router.Group("/captcha")
	{
		captcha.GET("/status", captchaHandler.GetCaptchaStatus)
		captcha.GET("/site-key", captchaHandler.GetSiteKey)
		captcha.POST("/verify", captchaHandler.VerifyCaptcha)
	}
}

// RegisterOAuthRoutes registers OAuth routes
func RegisterOAuthRoutes(router *gin.RouterGroup, oauthHandler *handlers.OAuthHandler) {
	// Protected routes for initiating OAuth
	platformAuth := router.Group("/platform-auth")
	platformAuth.Use(middleware.Auth())
	platformAuth.Use(middleware.Tenant())
	{
		platformAuth.GET("/initiate/:platform", oauthHandler.InitiateAuth)
		platformAuth.GET("/logs", oauthHandler.GetOAuthLogs)

		// Alias routes for frontend compatibility (frontend calls /lazada/authorize instead of /initiate/lazada)
		platformAuth.GET("/shopee/authorize", func(c *gin.Context) {
			c.Params = append(c.Params, gin.Param{Key: "platform", Value: "shopee"})
			oauthHandler.InitiateAuth(c)
		})
		platformAuth.GET("/lazada/authorize", func(c *gin.Context) {
			c.Params = append(c.Params, gin.Param{Key: "platform", Value: "lazada"})
			oauthHandler.InitiateAuth(c)
		})
		platformAuth.GET("/tiktok/authorize", func(c *gin.Context) {
			c.Params = append(c.Params, gin.Param{Key: "platform", Value: "tiktok"})
			oauthHandler.InitiateAuth(c)
		})
	}

	// Public callback routes (no auth middleware)
	callback := router.Group("/platform-auth/callback")
	{
		callback.GET("/:platform", oauthHandler.HandleCallback)
	}
}

// RegisterTokenRoutes registers token management routes
func RegisterTokenRoutes(router *gin.RouterGroup, tokenHandler *handlers.TokenHandler) {
	tokens := router.Group("/tokens")
	tokens.Use(middleware.Auth())
	tokens.Use(middleware.Tenant())
	{
		tokens.GET("/status", tokenHandler.GetAllTokenStatus)
		tokens.GET("/status/:platform", tokenHandler.GetTokenStatus)
		tokens.POST("/refresh/:platform", tokenHandler.RefreshToken)
		tokens.POST("/refresh-all", tokenHandler.RefreshAllTokens) // For frontend compatibility
	}

	// Alias routes for frontend compatibility (frontend calls /api/token-status)
	tokenAlias := router.Group("")
	tokenAlias.Use(middleware.Auth())
	tokenAlias.Use(middleware.Tenant())
	{
		tokenAlias.GET("/token-status", tokenHandler.GetAllTokenStatus)                // Alias: /api/token-status
		tokenAlias.GET("/:platform/token-status", tokenHandler.GetPlatformTokenStatus) // Alias: /api/shopee/token-status
	}
}

// RegisterAnalyticsRoutes registers analytics routes
func RegisterAnalyticsRoutes(router *gin.RouterGroup, analyticsHandler *handlers.AnalyticsHandler) {
	analytics := router.Group("/analytics")
	analytics.Use(middleware.Auth())
	analytics.Use(middleware.Tenant())
	{
		analytics.GET("/dashboard", analyticsHandler.GetDashboardSummary)
		analytics.GET("/orders", analyticsHandler.GetOrderAnalytics)
		analytics.GET("/revenue", analyticsHandler.GetRevenueAnalytics)
		analytics.GET("/settings", analyticsHandler.GetAnalyticsSettings)
		analytics.PUT("/settings", analyticsHandler.UpdateAnalyticsSettings)
		analytics.GET("/escrow", analyticsHandler.GetEscrowSyncStatus)
	}
}

// RegisterWebhookRoutes registers webhook routes (public, no auth)
func RegisterWebhookRoutes(router *gin.RouterGroup, webhookHandler *handlers.WebhookHandler) {

	webhooks := router.Group("/webhooks")
	{
		// Public webhook endpoints (no auth)
		webhooks.POST("/shopee", webhookHandler.ShopeeWebhook)
		webhooks.POST("/lazada", webhookHandler.LazadaWebhook)
		webhooks.POST("/tiktok", webhookHandler.TiktokWebhook)
	}
}

// RegisterWebhookAdminRoutes registers webhook admin routes (protected)
func RegisterWebhookAdminRoutes(router *gin.RouterGroup, webhookHandler *handlers.WebhookHandler) {
	webhooks := router.Group("/webhooks")
	webhooks.Use(middleware.Auth())
	webhooks.Use(middleware.Tenant())
	{
		webhooks.GET("/logs", webhookHandler.GetLogs)
		webhooks.GET("/logs/:platform", webhookHandler.GetLogsByPlatform)
		webhooks.GET("/stats", webhookHandler.GetStats)
	}
}
