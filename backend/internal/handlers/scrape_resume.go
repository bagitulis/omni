package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/services/scraper/shopee"
)

// maxJobIDLen caps an accepted job id. Job ids are UUIDs or a derived
// "scrape_<nanos>_<suffix>", both far below this.
const maxJobIDLen = 128

// Resume handles POST /api/extensions/scrape/:job_id/resume.
//
// The endpoint takes no body: the job id is in the path and the tenant comes
// from the authenticated context, so there is nothing a caller could supply that
// would not widen what they can reach.
//
// Resuming a job that is not blocked answers 200 with resumed=false rather than
// an error. The dashboard polls and several operators may watch the same job, so
// a second Resume is expected traffic, not a fault.
func (h *ScrapeHandler) Resume(c *gin.Context) {
	tenantID := tenantIDFromContext(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	jobID := c.Param("job_id")
	if err := validateResumeJobID(jobID); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	if !h.available(c) {
		return
	}

	outcome, err := h.scraper.Resume(c.Request.Context(), tenantID, jobID)
	if err != nil {
		if errors.Is(err, shopee.ErrJobNotFound) {
			c.JSON(http.StatusNotFound, response.Error("job not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, response.Error("Failed to resume the scrape job"))
		return
	}

	if !outcome.Resumed {
		c.JSON(http.StatusOK, response.Success(gin.H{
			"job_id":  outcome.JobID,
			"status":  outcome.Status,
			"resumed": false,
		}))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"job_id":     outcome.JobID,
		"status":     outcome.Status,
		"start_page": outcome.StartPage,
	}))
}

// validateResumeJobID rejects a job id that could be abused once it reaches a
// query, a path, or a log line. The id arrives from the browser, so it is
// untrusted input even though it is only ever a lookup key.
func validateResumeJobID(id string) error {
	if id == "" {
		return errText("job_id is required")
	}
	if len(id) > maxJobIDLen {
		return errText("job_id is too long")
	}
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-' || r == '_':
		default:
			return errText("job_id contains an invalid character")
		}
	}
	return nil
}
