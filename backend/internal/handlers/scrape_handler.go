package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/extensions"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/services/jobs"
	"github.com/omni/backend/internal/services/scraper/shopee"
	"gorm.io/gorm"
)

// ScrapeHandler exposes Shopee scrape control and results.
//
// Like ExtensionHandler, it stays thin: parse and validate, delegate, format.
// The one exception is reading scraped products, which is a plain paginated
// query with no business rules — routing that through a service would add a
// layer without adding meaning.
type ScrapeHandler struct {
	tenantDB func(tenantID string) (*gorm.DB, error)
	scraper  *shopee.ScrapeService
}

// NewScrapeHandler creates a ScrapeHandler.
func NewScrapeHandler(tenantDB func(tenantID string) (*gorm.DB, error), scraper *shopee.ScrapeService) *ScrapeHandler {
	return &ScrapeHandler{tenantDB: tenantDB, scraper: scraper}
}

func (h *ScrapeHandler) available(c *gin.Context) bool {
	if h == nil || h.scraper == nil || h.tenantDB == nil {
		c.JSON(http.StatusServiceUnavailable, response.Error("Scraper service is not available"))
		return false
	}
	return true
}

// startScrapeRequest is the body of POST /api/extensions/scrape.
//
// No tenant_id: the tenant is taken from the authenticated request context, and
// accepting one here would let a caller choose which schema the results land in.
type startScrapeRequest struct {
	Mode        string `json:"mode" binding:"required"`
	Query       string `json:"query"`
	ShopURL     string `json:"shop_url"`
	ShopID      string `json:"shop_id"`
	ProductURL  string `json:"product_url"`
	MaxPages    int    `json:"max_pages"`
	MaxProducts int    `json:"max_products"`
	ExtensionID string `json:"extension_id" binding:"required"`
}

// Start handles POST /api/extensions/scrape.
//
// A scrape is queued as a job rather than run inline:
//
//   - It is long-running and drives a browser, so holding a request open for its
//     whole duration would tie up a connection and time out behind a proxy.
//   - Queuing gives status and cancellation for free through the existing
//     /api/jobs endpoints (GET /api/jobs/:id, POST /api/jobs/cancel/:jobId),
//     instead of duplicating a second control surface here.
//
// The response carries the job id, which is also the id scraped products are
// grouped under.
func (h *ScrapeHandler) Start(c *gin.Context) {
	tenantID := tenantIDFromContext(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	var body startScrapeRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	if err := validateScrapeRequest(body); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	if !h.available(c) {
		return
	}

	db, err := h.tenantDB(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to resolve tenant database"))
		return
	}

	data := shopee.ScrapeJobData{
		Mode:        body.Mode,
		Query:       body.Query,
		ShopURL:     body.ShopURL,
		ShopID:      body.ShopID,
		ProductURL:  body.ProductURL,
		MaxPages:    body.MaxPages,
		MaxProducts: body.MaxProducts,
		ExtensionID: body.ExtensionID,
	}

	// The job id doubles as the results group id, so the caller can go straight
	// from "queued" to "show me what it collected".
	jobID := uuid.New().String()
	data.ResultsJobID = jobID

	payload, err := json.Marshal(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to build the job payload"))
		return
	}

	queue := jobs.NewQueueManager(db, "")
	job, err := queue.AddJob(models.CreateJobRequest{
		ID:       jobID,
		Type:     models.JobTypeShopeeScrape,
		Data:     string(payload),
		Priority: "normal",
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to queue the scrape job"))
		return
	}

	c.JSON(http.StatusAccepted, response.Success(gin.H{
		"job_id":      job.ID,
		"status":      job.Status,
		"results_url": "/api/extensions/scraped-products?job_id=" + job.ID,
	}))
}

// validateScrapeRequest enforces the mode-specific requirements before any work
// is dispatched, so an invalid request never opens a browser tab.
func validateScrapeRequest(body startScrapeRequest) error {
	switch body.Mode {
	case "search":
		if body.Query == "" {
			return errText("query is required for search mode")
		}
	case "shop":
		if body.ShopURL == "" {
			return errText("shop_url is required for shop mode")
		}
	case "product":
		if body.ProductURL == "" {
			return errText("product_url is required for product mode")
		}
	default:
		return errText("mode must be one of: search, shop, product")
	}

	// binding:"required" only proves the field is non-empty. The identifier is
	// browser-supplied and ends up in lookups and log lines, so it goes through
	// the same validation as every other path that accepts one.
	if err := extensions.ValidateExtensionID(body.ExtensionID); err != nil {
		return errText(err.Error())
	}
	return nil
}

// ListProducts handles GET /api/extensions/scraped-products.
func (h *ScrapeHandler) ListProducts(c *gin.Context) {
	tenantID := tenantIDFromContext(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	jobID := c.Query("job_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))

	if !h.available(c) {
		return
	}

	db, err := h.tenantDB(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to resolve tenant database"))
		return
	}

	products, total, err := repositories.NewExtensionRepository(db).
		ListScrapedProducts(c.Request.Context(), jobID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to list scraped products"))
		return
	}

	out := make([]scrapedProductDTO, 0, len(products))
	for _, p := range products {
		out = append(out, toScrapedProductDTO(p))
	}

	c.JSON(http.StatusOK, response.SuccessWithMeta(out, &response.Meta{
		Total:      int(total),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages(int(total), pageSize),
	}))
}

// scrapedProductDTO is the API shape of a scraped product.
type scrapedProductDTO struct {
	ID           int64  `json:"id"`
	JobID        string `json:"job_id"`
	Platform     string `json:"platform"`
	ScrapeMode   string `json:"scrape_mode"`
	Query        string `json:"query,omitempty"`
	ShopID       string `json:"shop_id,omitempty"`
	ProductName  string `json:"product_name"`
	Price        string `json:"price"`
	Sold         string `json:"sold"`
	Link         string `json:"link"`
	ImageURL     string `json:"image_url"`
	ShopeeItemID string `json:"shopee_item_id,omitempty"`
	PageNumber   int    `json:"page_number"`
	Source       string `json:"source,omitempty"`
	CreatedAt    string `json:"created_at"`
}

func toScrapedProductDTO(p models.ScrapedProduct) scrapedProductDTO {
	return scrapedProductDTO{
		ID:           p.ID,
		JobID:        p.JobID,
		Platform:     p.Platform,
		ScrapeMode:   p.ScrapeMode,
		Query:        p.Query,
		ShopID:       p.ShopID,
		ProductName:  p.ProductName,
		Price:        p.Price,
		Sold:         p.Sold,
		Link:         p.Link,
		ImageURL:     p.ImageURL,
		ShopeeItemID: p.ShopeeItemID,
		PageNumber:   p.PageNumber,
		Source:       p.Source,
		CreatedAt:    p.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
}

func totalPages(total, pageSize int) int {
	if pageSize <= 0 {
		return 0
	}
	pages := total / pageSize
	if total%pageSize != 0 {
		pages++
	}
	return pages
}

// errText is a local error type for small validators, so the message reaches the
// client verbatim without pulling a formatting dependency into this file.
type errText string

func (e errText) Error() string { return string(e) }

// mustJSON marshals a value built from plain request types. A marshal failure is
// a programming error, so it degrades to an empty object rather than panicking
// in a request handler.
func mustJSON(v any) string {
	raw, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(raw)
}
