package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/scraper/shopee"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// The resume endpoint's contract, as agreed in the hardening spec.
//
// The security-relevant half is pinned here rather than in the service: the
// handler is where a tenant could be defaulted, and where an untrusted job id
// arrives from the browser.

func newResumeTestRouter(t *testing.T, h *ScrapeHandler, withTenant bool) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	r := gin.New()
	group := r.Group("/api/extensions")
	if withTenant {
		group.Use(func(c *gin.Context) {
			c.Set("tenant_id", "test-tenant")
			c.Set("userID", "1")
			c.Next()
		})
	}
	group.POST("/scrape/:job_id/resume", h.Resume)
	return r
}

func postResume(t *testing.T, r *gin.Engine, jobID string) *httptest.ResponseRecorder {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, "/api/extensions/scrape/"+jobID+"/resume", nil)
	require.NoError(t, err)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// TestScrapeHandler_ResumeMissingTenantIsRejected enforces the no-default-tenant
// rule on the resume path. A defaulted tenant here would re-queue another
// tenant's job.
func TestScrapeHandler_ResumeMissingTenantIsRejected(t *testing.T) {
	h := NewScrapeHandler(nil, nil)
	w := postResume(t, newResumeTestRouter(t, h, false), "job-1")

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.False(t, resp.Success, "success must not be true on a rejected request")
	assert.Equal(t, "Missing tenant_id", resp.Error)
}

// TestScrapeHandler_ResumeNilServiceFailsClosed asserts an unwired handler
// returns 503 rather than panicking, since a panic in a Gin handler can take the
// process down.
func TestScrapeHandler_ResumeNilServiceFailsClosed(t *testing.T) {
	h := NewScrapeHandler(nil, nil)
	w := postResume(t, newResumeTestRouter(t, h, true), "job-1")

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

// TestScrapeHandler_ResumeRejectsMalformedJobID keeps an untrusted path segment
// from reaching the tenant schema as a query key.
func TestScrapeHandler_ResumeRejectsMalformedJobID(t *testing.T) {
	h := NewScrapeHandler(nil, nil)
	r := newResumeTestRouter(t, h, true)

	// A job id is a UUID in practice; anything with path or quote characters is
	// a caller bug and must be refused before the service is reached.
	w := postResume(t, r, "job%20one%27")

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestScrapeHandler_ResumeTakesNoBody documents that the endpoint is driven by
// the path alone: a body naming a tenant or a start page must not be honoured.
func TestScrapeHandler_ResumeTakesNoBody(t *testing.T) {
	h := NewScrapeHandler(nil, nil)
	r := newResumeTestRouter(t, h, true)

	req, err := http.NewRequest(http.MethodPost, "/api/extensions/scrape/job-1/resume",
		bytes.NewReader([]byte(`{"tenant_id":"other-tenant","start_page":99}`)))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// The body is ignored, so the request still stops at the availability check
	// rather than being parsed into anything.
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

// resumeHandlerWithDB wires a real ScrapeService over an in-memory schema, so the
// response bodies below are the ones an operator's browser actually receives.
func resumeHandlerWithDB(t *testing.T) (*ScrapeHandler, *gorm.DB) {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", uuid.New().String())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Job{}))

	tenantDB := func(tenantID string) (*gorm.DB, error) {
		if tenantID != "test-tenant" {
			return nil, fmt.Errorf("unknown tenant %q", tenantID)
		}
		return db, nil
	}
	return NewScrapeHandler(tenantDB, shopee.NewScrapeService(tenantDB, nil)), db
}

func seedResumeJob(t *testing.T, db *gorm.DB, status models.JobStatus, resultData string) string {
	t.Helper()
	job := models.Job{
		ID:         uuid.New().String(),
		Type:       models.JobTypeShopeeScrape,
		Status:     status,
		Priority:   "normal",
		Data:       `{"mode":"search","query":"kaos","extension_id":"ext-1"}`,
		ResultData: resultData,
	}
	require.NoError(t, db.Create(&job).Error)
	return job.ID
}

// TestScrapeHandler_ResumeReturnsStartPage pins the re-queued response body from
// the spec, including the snake_case keys the dashboard reads.
func TestScrapeHandler_ResumeReturnsStartPage(t *testing.T) {
	h, db := resumeHandlerWithDB(t)
	jobID := seedResumeJob(t, db, models.JobStatusBlocked,
		`{"reason":"blocked","resume_from_page":4}`)

	w := postResume(t, newResumeTestRouter(t, h, true), jobID)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Success bool `json:"success"`
		Data    struct {
			JobID     string `json:"job_id"`
			Status    string `json:"status"`
			StartPage int    `json:"start_page"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	assert.True(t, resp.Success)
	assert.Equal(t, jobID, resp.Data.JobID)
	assert.Equal(t, "pending", resp.Data.Status)
	assert.Equal(t, 4, resp.Data.StartPage)
}

// TestScrapeHandler_ResumeNotBlockedIsNoOp: a resume on a job that is not blocked
// answers 200 with resumed=false. It is expected traffic from a polling
// dashboard, not a fault.
func TestScrapeHandler_ResumeNotBlockedIsNoOp(t *testing.T) {
	h, db := resumeHandlerWithDB(t)
	jobID := seedResumeJob(t, db, models.JobStatusCompleted, `{"reason":"last_page"}`)

	w := postResume(t, newResumeTestRouter(t, h, true), jobID)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Success bool `json:"success"`
		Data    struct {
			JobID   string `json:"job_id"`
			Status  string `json:"status"`
			Resumed *bool  `json:"resumed"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	assert.True(t, resp.Success)
	assert.Equal(t, jobID, resp.Data.JobID)
	assert.Equal(t, "completed", resp.Data.Status)
	require.NotNil(t, resp.Data.Resumed, "the no-op response must carry resumed")
	assert.False(t, *resp.Data.Resumed)

	var after models.Job
	require.NoError(t, db.Where("id = ?", jobID).First(&after).Error)
	assert.Equal(t, models.JobStatusCompleted, after.Status,
		"a no-op resume must not move the job")
}

// TestScrapeHandler_ResumeUnknownJobIs404 stops a typo'd id from being reported
// as a successful resume.
func TestScrapeHandler_ResumeUnknownJobIs404(t *testing.T) {
	h, _ := resumeHandlerWithDB(t)

	w := postResume(t, newResumeTestRouter(t, h, true), "no-such-job")
	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.False(t, resp.Success)
	assert.Equal(t, "job not found", resp.Error)
}

// TestScrapeHandler_ResumeDoesNotCrossTenants: the tenant comes from the request
// context, so a job that exists in another schema is simply not found here.
func TestScrapeHandler_ResumeDoesNotCrossTenants(t *testing.T) {
	h, db := resumeHandlerWithDB(t)
	jobID := seedResumeJob(t, db, models.JobStatusBlocked, `{"resume_from_page":4}`)

	r := gin.New()
	group := r.Group("/api/extensions")
	group.Use(func(c *gin.Context) {
		c.Set("tenant_id", "other-tenant")
		c.Next()
	})
	group.POST("/scrape/:job_id/resume", h.Resume)

	w := postResume(t, r, jobID)
	assert.NotEqual(t, http.StatusOK, w.Code,
		"a tenant must not resume a job it cannot see")

	var after models.Job
	require.NoError(t, db.Where("id = ?", jobID).First(&after).Error)
	assert.Equal(t, models.JobStatusBlocked, after.Status,
		"the other tenant's job must stay blocked")
}

// TestScrapeHandler_ResumeFailureIsNotAFalsePositive: an internal failure must
// never answer success:true. Product mode has no page after the first, so a
// cursor past it is rejected by the service.
func TestScrapeHandler_ResumeFailureIsNotAFalsePositive(t *testing.T) {
	h, db := resumeHandlerWithDB(t)

	job := models.Job{
		ID:         uuid.New().String(),
		Type:       models.JobTypeShopeeScrape,
		Status:     models.JobStatusBlocked,
		Priority:   "normal",
		Data:       `{"mode":"product","product_url":"https://shopee.co.id/product/1/2","extension_id":"ext-1"}`,
		ResultData: `{"resume_from_page":3}`,
	}
	require.NoError(t, db.Create(&job).Error)

	w := postResume(t, newResumeTestRouter(t, h, true), job.ID)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp struct {
		Success bool `json:"success"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.False(t, resp.Success, "a failure must never report success")
}

func TestValidateResumeJobID(t *testing.T) {
	cases := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{"uuid", "3f4b6c1e-9a2d-4c7f-8e11-2b5d6a0c9f33", false},
		{"generated scrape id", "scrape_1726480000000_ab12cd34", false},
		{"empty", "", true},
		{"path traversal", "../../etc/passwd", true},
		{"quote", "job'1", true},
		{"space", "job 1", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateResumeJobID(tc.id)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
