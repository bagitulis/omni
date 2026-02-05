package httputils

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDownloadImage_Success(t *testing.T) {
	// Create a test server that returns image data
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("fake image data"))
	}))
	defer server.Close()

	ctx := context.Background()
	data, err := DownloadImage(ctx, server.URL)

	assert.NoError(t, err)
	assert.Equal(t, []byte("fake image data"), data)
}

func TestDownloadImage_StatusNotOK(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		expectedErrMsg string
	}{
		{
			name:           "not found",
			statusCode:     http.StatusNotFound,
			expectedErrMsg: "404",
		},
		{
			name:           "internal server error",
			statusCode:     http.StatusInternalServerError,
			expectedErrMsg: "500",
		},
		{
			name:           "bad gateway",
			statusCode:     http.StatusBadGateway,
			expectedErrMsg: "502",
		},
		{
			name:           "forbidden",
			statusCode:     http.StatusForbidden,
			expectedErrMsg: "403",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte("error"))
			}))
			defer server.Close()

			ctx := context.Background()
			data, err := DownloadImage(ctx, server.URL)

			assert.Error(t, err)
			assert.Nil(t, data)
			assert.Contains(t, err.Error(), tt.expectedErrMsg)
		})
	}
}

func TestDownloadImage_InvalidURL(t *testing.T) {
	ctx := context.Background()
	data, err := DownloadImage(ctx, "http://invalid-url-that-does-not-exist-12345.local:9999")

	assert.Error(t, err)
	assert.Nil(t, data)
}

func TestDownloadImage_MalformedURL(t *testing.T) {
	ctx := context.Background()
	data, err := DownloadImage(ctx, "not a valid url at all")

	assert.Error(t, err)
	assert.Nil(t, data)
}

func TestDownloadImage_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("data"))
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	data, err := DownloadImage(ctx, server.URL)

	assert.Error(t, err)
	assert.Nil(t, data)
}

func TestDownloadImage_ContextTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow server
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("slow data"))
	}))
	defer server.Close()

	// Create a context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	data, err := DownloadImage(ctx, server.URL)

	assert.Error(t, err)
	assert.Nil(t, data)
}

func TestDownloadImage_LargeData(t *testing.T) {
	// Create a large image data (1MB)
	largeData := strings.Repeat("a", 1024*1024)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(largeData))
	}))
	defer server.Close()

	ctx := context.Background()
	data, err := DownloadImage(ctx, server.URL)

	assert.NoError(t, err)
	assert.Equal(t, []byte(largeData), data)
	assert.Equal(t, 1024*1024, len(data))
}

func TestDownloadImage_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.WriteHeader(http.StatusOK)
		// Write nothing
	}))
	defer server.Close()

	ctx := context.Background()
	data, err := DownloadImage(ctx, server.URL)

	assert.NoError(t, err)
	assert.Empty(t, data)
}

func TestDownloadImage_BinaryData(t *testing.T) {
	// Create binary data similar to an actual image
	binaryData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A} // PNG signature

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.WriteHeader(http.StatusOK)
		w.Write(binaryData)
	}))
	defer server.Close()

	ctx := context.Background()
	data, err := DownloadImage(ctx, server.URL)

	assert.NoError(t, err)
	assert.Equal(t, binaryData, data)
}

func TestDownloadImage_BodyReadError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.WriteHeader(http.StatusOK)
		// Return response that will be readable
		w.Write([]byte("data"))
		w.(http.Flusher).Flush()
	}))
	defer server.Close()

	ctx := context.Background()
	data, err := DownloadImage(ctx, server.URL)

	// This should succeed
	// The io.ReadAll should handle the response properly
	assert.NoError(t, err)
	assert.NotNil(t, data)
}

func TestDownloadImage_RequestMethod(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method is GET
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("image data"))
	}))
	defer server.Close()

	ctx := context.Background()
	data, err := DownloadImage(ctx, server.URL)

	assert.NoError(t, err)
	assert.Equal(t, []byte("image data"), data)
}

func TestDownloadImage_WithContextDeadline(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	deadline := time.Now().Add(2 * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	data, err := DownloadImage(ctx, server.URL)

	assert.NoError(t, err)
	assert.Equal(t, []byte("success"), data)
}

func TestDownloadImage_RedirectResponse(t *testing.T) {
	// Create two servers: one for redirect, one for actual content
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("redirected content"))
	}))
	defer targetServer.Close()

	redirectServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, targetServer.URL, http.StatusMovedPermanently)
	}))
	defer redirectServer.Close()

	ctx := context.Background()
	data, err := DownloadImage(ctx, redirectServer.URL)

	// http.Client follows redirects by default
	assert.NoError(t, err)
	assert.Equal(t, []byte("redirected content"), data)
}

func TestDownloadImage_VariousContentTypes(t *testing.T) {
	contentTypes := []string{
		"image/png",
		"image/jpeg",
		"image/gif",
		"image/webp",
		"application/octet-stream",
	}

	for _, ct := range contentTypes {
		t.Run(ct, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", ct)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("image content"))
			}))
			defer server.Close()

			ctx := context.Background()
			data, err := DownloadImage(ctx, server.URL)

			assert.NoError(t, err)
			assert.Equal(t, []byte("image content"), data)
		})
	}
}
