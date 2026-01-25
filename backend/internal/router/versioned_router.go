package router

import (
	"github.com/gin-gonic/gin"
)

// APIVersion represents an API version
type APIVersion string

const (
	V1 APIVersion = "v1"
	V2 APIVersion = "v2"
)

// VersionedRouter provides API versioning support
type VersionedRouter struct {
	engine   *gin.Engine
	versions map[APIVersion]*gin.RouterGroup
}

// NewVersionedRouter creates a new versioned router
func NewVersionedRouter(engine *gin.Engine) *VersionedRouter {
	return &VersionedRouter{
		engine:   engine,
		versions: make(map[APIVersion]*gin.RouterGroup),
	}
}

// Version returns the router group for a specific API version
func (vr *VersionedRouter) Version(v APIVersion) *gin.RouterGroup {
	if group, exists := vr.versions[v]; exists {
		return group
	}

	group := vr.engine.Group("/api/" + string(v))
	vr.versions[v] = group
	return group
}

// V1 returns the v1 API router group
func (vr *VersionedRouter) V1() *gin.RouterGroup {
	return vr.Version(V1)
}

// V2 returns the v2 API router group
func (vr *VersionedRouter) V2() *gin.RouterGroup {
	return vr.Version(V2)
}

// Legacy returns the unversioned /api group for backward compatibility
func (vr *VersionedRouter) Legacy() *gin.RouterGroup {
	return vr.engine.Group("/api")
}

// DeprecationMiddleware adds deprecation headers for old API versions
func DeprecationMiddleware(deprecatedVersion APIVersion, sunset string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Deprecation", "true")
		c.Header("Sunset", sunset)
		c.Header("Link", "</api/v2>; rel=\"successor-version\"")
		c.Next()
	}
}

// VersionNegotiation middleware for Accept-Version header support
func VersionNegotiation(defaultVersion APIVersion) gin.HandlerFunc {
	return func(c *gin.Context) {
		version := c.GetHeader("Accept-Version")
		if version == "" {
			version = c.GetHeader("X-API-Version")
		}
		if version == "" {
			version = string(defaultVersion)
		}

		c.Set("api_version", version)
		c.Header("X-API-Version", version)
		c.Next()
	}
}

// GetAPIVersion returns the API version from context
func GetAPIVersion(c *gin.Context) string {
	if v, exists := c.Get("api_version"); exists {
		return v.(string)
	}
	return string(V1)
}
