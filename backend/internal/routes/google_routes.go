package routes

import (
	"github.com/gin-gonic/gin"
	googleHandler "github.com/omni/backend/internal/handlers/google"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/services/google"
	"gorm.io/gorm"
)

// GoogleHandlers contains all Google-related handlers
type GoogleHandlers struct {
	Auth           *googleHandler.AuthHandler
	Sheets         *googleHandler.SheetsHandler
	Settings       *googleHandler.SettingsHandler
	ServiceAccount *googleHandler.ServiceAccountHandler
	Quota          *googleHandler.QuotaHandler
	SheetConfig    *googleHandler.SheetConfigHandler
}

// NewGoogleHandlers creates all Google handlers
func NewGoogleHandlers(authService *google.AuthService, quotaService *google.QuotaService, db *gorm.DB) *GoogleHandlers {
	return &GoogleHandlers{
		Auth:           googleHandler.NewAuthHandler(authService),
		Sheets:         googleHandler.NewSheetsHandler(authService),
		Settings:       googleHandler.NewSettingsHandler(authService, db),
		ServiceAccount: googleHandler.NewServiceAccountHandler(authService),
		Quota:          googleHandler.NewQuotaHandler(quotaService),
		SheetConfig:    googleHandler.NewSheetConfigHandler(db, authService),
	}
}

// RegisterGoogleRoutes registers all Google API routes
func RegisterGoogleRoutes(router *gin.RouterGroup, handlers *GoogleHandlers) {
	google := router.Group("/google")

	// Auth routes (some public, some protected)
	RegisterGoogleAuthRoutes(google, handlers.Auth)

	// Sheets routes (protected)
	RegisterGoogleSheetsRoutes(google, handlers.Sheets)

	// Settings routes (protected)
	RegisterGoogleSettingsRoutes(google, handlers.Settings)

	// Service account routes (protected)
	RegisterGoogleServiceAccountRoutes(google, handlers.ServiceAccount)

	// Quota routes (public - for monitoring)
	RegisterGoogleQuotaRoutes(google, handlers.Quota)

	// Sheet config routes (protected)
	RegisterGoogleSheetConfigRoutes(google, handlers.SheetConfig)
}

// RegisterGoogleAuthRoutes registers Google auth routes
func RegisterGoogleAuthRoutes(router *gin.RouterGroup, handler *googleHandler.AuthHandler) {
	auth := router.Group("/auth")
	{
		// Public routes
		auth.GET("/status", handler.GetAuthStatus)
		auth.GET("/callback", handler.HandleCallback)

		// Protected routes
		protected := auth.Group("")
		protected.Use(middleware.Auth())
		protected.Use(middleware.Tenant())
		{
			protected.GET("/url", handler.GetAuthURL)
			protected.POST("/disconnect", handler.Disconnect)
		}
	}
}

// RegisterGoogleSheetsRoutes registers Google Sheets routes
func RegisterGoogleSheetsRoutes(router *gin.RouterGroup, handler *googleHandler.SheetsHandler) {
	sheets := router.Group("/sheets")
	sheets.Use(middleware.Auth())
	sheets.Use(middleware.Tenant())
	{
		sheets.GET("/list", handler.ListSpreadsheets)
		sheets.GET("/data", handler.GetSpreadsheetData)
		sheets.GET("/worksheets/:spreadsheetId", handler.GetWorksheets)
		sheets.GET("/columns/:spreadsheetId/:sheetName", handler.GetColumnHeaders)
		sheets.POST("/create", handler.CreateSpreadsheet)
		sheets.GET("/refresh", handler.RefreshSpreadsheets)
	}
}

// RegisterGoogleSettingsRoutes registers Google settings routes
func RegisterGoogleSettingsRoutes(router *gin.RouterGroup, handler *googleHandler.SettingsHandler) {
	settings := router.Group("/settings")
	settings.Use(middleware.Auth())
	settings.Use(middleware.Tenant())
	{
		settings.GET("/detailed", handler.GetDetailedSettings)
		settings.POST("/update-detailed", handler.UpdateDetailedSettings)
		settings.GET("/test", handler.TestConnection)
		settings.POST("/save-links", handler.SaveLinks)
		settings.GET("/saved-links", handler.GetSavedLinks)
		settings.POST("/validate-link", handler.ValidateLink)
	}
}

// RegisterGoogleServiceAccountRoutes registers service account routes
func RegisterGoogleServiceAccountRoutes(router *gin.RouterGroup, handler *googleHandler.ServiceAccountHandler) {
	serviceAccounts := router.Group("/service-accounts")
	serviceAccounts.Use(middleware.Auth())
	{
		serviceAccounts.GET("", handler.ListAccounts)
		serviceAccounts.GET("/stats", handler.GetStats)
		serviceAccounts.POST("/switch", handler.SwitchAccount)
		serviceAccounts.POST("/detect-sheet", handler.DetectSheet)
	}
}

// RegisterGoogleQuotaRoutes registers quota routes (public)
func RegisterGoogleQuotaRoutes(router *gin.RouterGroup, handler *googleHandler.QuotaHandler) {
	quota := router.Group("/quota")
	{
		quota.GET("/status", handler.GetStatus)
		quota.GET("/stats", handler.GetStats)
	}
}

// RegisterGoogleSheetConfigRoutes registers sheet config routes
func RegisterGoogleSheetConfigRoutes(router *gin.RouterGroup, handler *googleHandler.SheetConfigHandler) {
	sheetConfig := router.Group("/sheet-config")
	sheetConfig.Use(middleware.Auth())
	sheetConfig.Use(middleware.Tenant())
	{
		sheetConfig.POST("/save", handler.SaveConfig)
		sheetConfig.GET("/list", handler.ListConfigs)
		sheetConfig.DELETE("/:sheetId", handler.DeleteConfig)
	}
}
