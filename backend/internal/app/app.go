package app

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/handlers"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/services"
	"github.com/omni/backend/internal/services/cache"
	"github.com/omni/backend/internal/services/oauth"
	"github.com/omni/backend/internal/services/platform"
	"github.com/omni/backend/internal/services/webhooks"
	"github.com/omni/backend/internal/utils"
	zlog "github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// App contains all application dependencies
type App struct {
	Config     *config.Config
	BasePath   string
	SystemDB   *gorm.DB
	Encryption *utils.EncryptionService
	JWTService *utils.JWTService

	// Cache Service
	CacheService cache.CacheManager

	// Repositories
	UserRepo           *repositories.UserRepository
	AuditRepo          *repositories.AuditRepository
	OAuthRepo          *repositories.OAuthRepository
	WebhookRepo        *repositories.WebhookRepository
	AnalyticsRepo      *repositories.AnalyticsRepository
	GlobalConfigRepo   *repositories.GlobalConfigRepository
	PlatformRepo       *repositories.PlatformConfigRepository
	RefreshSessionRepo *repositories.RefreshSessionRepository

	// OAuth Services
	ShopeeOAuth *oauth.ShopeeOAuthService
	LazadaOAuth *oauth.LazadaOAuthService
	TiktokOAuth *oauth.TiktokOAuthService

	// Services
	AuthService      *services.AuthService
	MultiTenantAuth  *services.MultiTenantAuthService
	TenantService    *services.TenantService
	UserService      *services.UserManagementService
	AuditService     *services.AuditService
	CaptchaService   *services.CaptchaService
	TokenManager     *services.TokenManager
	AnalyticsService *services.AnalyticsService

	// Webhook Processors
	ShopeeProcessor *webhooks.ShopeeWebhookProcessor
	LazadaProcessor *webhooks.LazadaWebhookProcessor
	TiktokProcessor *webhooks.TiktokWebhookProcessor

	// Handlers
	AuthHandler         *handlers.AuthHandler
	UserHandler         *handlers.UserHandler
	AuditHandler        *handlers.AuditHandler
	CaptchaHandler      *handlers.CaptchaHandler
	TokenHandler        *handlers.TokenHandler
	OAuthHandler        *handlers.OAuthHandler
	AnalyticsHandler    *handlers.AnalyticsHandler
	WebhookHandler      *handlers.WebhookHandler
	PlatformAuthHandler *handlers.PlatformAuthHandler
	AdsHandler          *handlers.AdsHandler
}

// New creates and initializes the application
func New(cfg *config.Config) (*App, error) {
	app := &App{
		Config:   cfg,
		BasePath: cfg.DatabasePath,
	}

	if err := app.initCore(); err != nil {
		return nil, err
	}

	app.initRepositories()
	app.initOAuthServices()
	app.initServices()
	app.initProcessors()
	app.initHandlers()

	return app, nil
}

// initCore initializes core dependencies
func (a *App) initCore() error {
	isProd := os.Getenv("GO_ENV") == "production"

	// Initialize database driver based on config
	if err := a.initDatabaseDriver(); err != nil {
		return err
	}

	// Get system database
	systemDB, err := config.GetSystemDB(a.Config.DatabasePath)
	if err != nil {
		return err
	}
	a.SystemDB = systemDB

	// Initialize GlobalConfigService with existing system DB connection
	// This prevents creating new connections for each config fetch
	config.InitGlobalConfigService(systemDB)

	// Run system database migrations
	if err := config.MigrateSystemDatabase(systemDB); err != nil {
		if isProd {
			return fmt.Errorf("system database migration failed: %w", err)
		}
		zlog.Warn().Err(err).Msg("System database migration failed")
	}

	// Run tenant database migrations for all known tenants
	tenantSvc := services.NewTenantService(a.Config.DatabasePath)
	if err := tenantSvc.MigrateAllTenants(context.Background()); err != nil {
		if isProd {
			return fmt.Errorf("tenant database migration failed: %w", err)
		}
		zlog.Warn().Err(err).Msg("Tenant database migration failed")
	}

	// Initialize encryption
	encKey := os.Getenv("ENCRYPTION_KEY")
	if encKey == "" {
		if isProd {
			return fmt.Errorf("ENCRYPTION_KEY must be set in production")
		}
		zlog.Warn().Msg("ENCRYPTION_KEY not set, using dev default")
		encKey = "default-dev-key-32-bytes-long!!"
	}
	a.Encryption, err = utils.NewEncryptionService(encKey)
	if err != nil {
		return fmt.Errorf("encryption init failed: %w", err)
	}

	// Initialize JWT
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		if isProd {
			return fmt.Errorf("JWT_SECRET must be set in production")
		}
		zlog.Warn().Msg("JWT_SECRET not set, using dev default")
		jwtSecret = "default-jwt-secret-change-in-prod"
	}
	a.JWTService = utils.NewJWTService(jwtSecret)

	// Initialize Cache Service
	// Default: 5 minute expiration, 10 minute cleanup interval
	a.CacheService = cache.New(5*time.Minute, 10*time.Minute)
	zlog.Info().Msg("Cache service initialized (in-memory, multi-tenant)")

	return nil
}

// initDatabaseDriver configures the database driver (PostgreSQL only)
func (a *App) initDatabaseDriver() error {
	pgConfig := &config.PostgresConfig{
		Host:     a.Config.PGHost,
		Port:     a.Config.PGPort,
		User:     a.Config.PGUser,
		Password: a.Config.PGPassword,
		DBName:   a.Config.PGDatabase,
		SSLMode:  a.Config.PGSSLMode,
	}

	config.SetDatabaseDriver(config.DriverPostgres, pgConfig)
	models.SetDatabaseDriver(models.DriverPostgres)
	zlog.Info().Str("host", a.Config.PGHost).Str("db", a.Config.PGDatabase).Msg("Database driver: PostgreSQL")

	return nil
}

// initRepositories initializes all repositories
func (a *App) initRepositories() {
	a.UserRepo = repositories.NewUserRepository(a.SystemDB)
	a.AuditRepo = repositories.NewAuditRepository(a.SystemDB)
	a.OAuthRepo = repositories.NewOAuthRepository(a.SystemDB)
	a.WebhookRepo = repositories.NewWebhookRepository(a.SystemDB)
	a.AnalyticsRepo = repositories.NewAnalyticsRepository(a.SystemDB)
	a.GlobalConfigRepo = repositories.NewGlobalConfigRepository(a.SystemDB)
	a.PlatformRepo = repositories.NewPlatformConfigRepository(a.SystemDB)
	a.RefreshSessionRepo = repositories.NewRefreshSessionRepository(a.SystemDB)
}

// initOAuthServices initializes OAuth services for each platform
func (a *App) initOAuthServices() {
	// These will be initialized with actual credentials when needed
	// For now, create placeholder instances
	a.ShopeeOAuth = oauth.NewShopeeOAuthService(0, "", "", false)
	a.LazadaOAuth = oauth.NewLazadaOAuthService("", "", "", false)
	a.TiktokOAuth = oauth.NewTiktokOAuthService("", "", "", false)
}

// initServices initializes all services
func (a *App) initServices() {
	a.AuthService = services.NewAuthServiceWithRefresh(a.UserRepo, a.AuditRepo, a.RefreshSessionRepo, a.JWTService)
	a.UserService = services.NewUserManagementService(a.UserRepo, a.AuditRepo, a.AuthService)
	a.AuditService = services.NewAuditService(a.AuditRepo)

	// Initialize tenant service for multi-tenant operations
	a.TenantService = services.NewTenantService(a.BasePath)
	a.MultiTenantAuth = services.NewMultiTenantAuthService(a.TenantService, a.JWTService, a.BasePath)

	captchaConfig := &services.CaptchaConfig{
		SecretKey: os.Getenv("RECAPTCHA_SECRET"),
		SiteKey:   os.Getenv("RECAPTCHA_SITE_KEY"),
		Enabled:   os.Getenv("RECAPTCHA_ENABLED") == "true",
	}
	a.CaptchaService = services.NewCaptchaService(captchaConfig)

	// TokenManager uses tenant-scoped database for PlatformConfig
	a.TokenManager = services.NewTokenManager(a.GlobalConfigRepo, a.Encryption, a.BasePath)

	// Register TokenManager globally so all CredentialService instances
	// (created ad-hoc across handlers/services) auto-refresh expired tokens
	services.RegisterGlobalTokenManager(a.TokenManager)

	// Register token refresh adapter for platform coordination service
	// This enables auto-refresh of expired tokens before API calls
	tokenRefreshAdapter := services.NewTokenRefreshAdapter(a.TokenManager)
	platform.RegisterTokenRefreshService(tokenRefreshAdapter)

	a.AnalyticsService = services.NewAnalyticsService(a.AnalyticsRepo)
}

// initProcessors initializes webhook processors
func (a *App) initProcessors() {
	a.ShopeeProcessor = webhooks.NewShopeeWebhookProcessor(a.WebhookRepo, a.ShopeeOAuth)
	a.LazadaProcessor = webhooks.NewLazadaWebhookProcessor(a.WebhookRepo, a.LazadaOAuth)
	a.TiktokProcessor = webhooks.NewTiktokWebhookProcessor(a.WebhookRepo, a.TiktokOAuth)
}

// initHandlers initializes all handlers
func (a *App) initHandlers() {
	a.AuthHandler = handlers.NewAuthHandler(a.AuthService, a.MultiTenantAuth, a.UserService, a.BasePath)
	a.UserHandler = handlers.NewUserHandler(a.UserService)
	a.AuditHandler = handlers.NewAuditHandler(a.AuditService)
	a.CaptchaHandler = handlers.NewCaptchaHandler(a.CaptchaService)
	a.TokenHandler = handlers.NewTokenHandler(a.TokenManager)
	a.AnalyticsHandler = handlers.NewAnalyticsHandler(a.SystemDB)

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		if os.Getenv("GO_ENV") == "production" {
			zlog.Warn().Msg("FRONTEND_URL not set in production, defaulting to localhost")
		}
		frontendURL = "http://localhost:5173"
	}
	a.OAuthHandler = handlers.NewOAuthHandler(
		a.OAuthRepo,
		a.GlobalConfigRepo,
		a.PlatformRepo,
		frontendURL,
	)

	a.WebhookHandler = handlers.NewWebhookHandlerWithCache(
		a.ShopeeProcessor,
		a.LazadaProcessor,
		a.TiktokProcessor,
		a.BasePath,
		a.CacheService,
	)

	a.PlatformAuthHandler = handlers.NewPlatformAuthHandler(
		a.GlobalConfigRepo,
		frontendURL,
		a.BasePath,
	)

	a.AdsHandler = handlers.NewAdsHandler(a.SystemDB)
}

// Close closes all database connections
func (a *App) Close() {
	config.CloseAllDBs()
}
