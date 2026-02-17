package route

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveFrontendSourcePath_UsesEnvOverride(t *testing.T) {
	overridePath := filepath.Join(t.TempDir(), "frontend-src")
	err := os.MkdirAll(overridePath, 0o755)
	assert.NoError(t, err)

	t.Setenv("FRONTEND_SOURCE_PATH", overridePath)

	scanner := NewScannerService(filepath.Join(t.TempDir(), "internal"))
	assert.Equal(t, overridePath, scanner.resolveFrontendSourcePath())
}

func TestResolveFrontendSourcePath_FrontendSiblingOfBasePath(t *testing.T) {
	t.Setenv("FRONTEND_SOURCE_PATH", "")
	rootDir := t.TempDir()
	internalDir := filepath.Join(rootDir, "internal")
	frontendSrcDir := filepath.Join(rootDir, "frontend", "src")

	err := createDir(internalDir)
	assert.NoError(t, err)
	err = createDir(frontendSrcDir)
	assert.NoError(t, err)

	scanner := NewScannerService(internalDir)
	assert.Equal(t, frontendSrcDir, scanner.resolveFrontendSourcePath())
}

func TestResolveFrontendSourcePath_IgnoresInvalidEnvAndFallsBack(t *testing.T) {
	rootDir := t.TempDir()
	internalDir := filepath.Join(rootDir, "internal")
	frontendSrcDir := filepath.Join(rootDir, "frontend", "src")

	err := createDir(internalDir)
	assert.NoError(t, err)
	err = createDir(frontendSrcDir)
	assert.NoError(t, err)

	t.Setenv("FRONTEND_SOURCE_PATH", filepath.Join(rootDir, "missing", "src"))

	scanner := NewScannerService(internalDir)
	assert.Equal(t, frontendSrcDir, scanner.resolveFrontendSourcePath())
}

func createDir(path string) error {
	return os.MkdirAll(path, 0o755)
}
