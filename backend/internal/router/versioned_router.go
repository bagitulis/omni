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
