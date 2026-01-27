package ml

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// ReportService handles ML report generation and management
type ReportService struct {
	db         *gorm.DB
	executor   *PythonExecutor
	reportsDir string
	scriptsDir string
}

// NewReportService creates a new report service
func NewReportService(db *gorm.DB, reportsDir, scriptsDir string) *ReportService {
	if reportsDir == "" {
		reportsDir = "./output"
	}
	if scriptsDir == "" {
		scriptsDir = "./notebooks"
	}

	// Ensure directories exist
	os.MkdirAll(reportsDir, 0755)

	return &ReportService{
		db:         db,
		executor:   NewPythonExecutor("python", 10*time.Minute),
		reportsDir: reportsDir,
		scriptsDir: scriptsDir,
	}
}

// GenerateReport triggers Python ML script to generate report
func (s *ReportService) GenerateReport(ctx context.Context, tenantID, platform string) (*models.MLReport, error) {
	// Validate platform
	if platform != "shopee" && platform != "tiktok" {
		return nil, fmt.Errorf("invalid platform: %s", platform)
	}

	// Determine script path
	var scriptPath string
	if platform == "tiktok" {
		scriptPath = filepath.Join(s.scriptsDir, "tiktok_ads", "scripts", "generate_report.py")
	} else {
		scriptPath = filepath.Join(s.scriptsDir, "shopee_ads", "scripts", "generate_report.py")
	}

	// Check if script exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("script not found: %s", scriptPath)
	}

	// Execute Python script
	result, err := s.executor.Execute(ctx, scriptPath, "--tenant", tenantID, "--quiet")
	if err != nil {
		return nil, fmt.Errorf("failed to execute script: %w", err)
	}

	if result.ExitCode != 0 {
		return nil, fmt.Errorf("script failed with exit code %d: %s", result.ExitCode, result.Stderr)
	}

	// Find generated HTML file
	htmlFile, err := s.findLatestReport(platform)
	if err != nil {
		return nil, fmt.Errorf("report generated but file not found: %w", err)
	}

	// Get file info
	fileInfo, err := os.Stat(htmlFile)
	if err != nil {
		return nil, err
	}

	// Create report record
	report := &models.MLReport{
		TenantID:    tenantID,
		Platform:    platform,
		ReportType:  "full",
		PeriodStart: time.Now().AddDate(0, -1, 0), // last month
		PeriodEnd:   time.Now(),
		PeriodLabel: time.Now().Format("2006-01"),
		FilePath:    htmlFile,
		FileName:    filepath.Base(htmlFile),
		FileSize:    fileInfo.Size(),
		Status:      "completed",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save to database
	if err := s.db.WithContext(ctx).Create(report).Error; err != nil {
		return nil, err
	}

	return report, nil
}

// GetLatestReport retrieves the most recent report
func (s *ReportService) GetLatestReport(ctx context.Context, tenantID, platform string) (*models.MLReport, error) {
	var report models.MLReport
	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND platform = ? AND status = ?", tenantID, platform, "completed").
		Order("created_at DESC").
		First(&report).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &report, err
}

// ListReports returns all reports for tenant and platform
func (s *ReportService) ListReports(ctx context.Context, tenantID, platform string, limit int) ([]*models.MLReport, error) {
	if limit == 0 {
		limit = 20
	}

	var reports []*models.MLReport
	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND platform = ?", tenantID, platform).
		Order("created_at DESC").
		Limit(limit).
		Find(&reports).Error

	return reports, err
}

// GetReportByFilename retrieves report by filename
func (s *ReportService) GetReportByFilename(ctx context.Context, tenantID, platform, filename string) (*models.MLReport, error) {
	var report models.MLReport
	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND platform = ? AND file_name = ?", tenantID, platform, filename).
		First(&report).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &report, err
}

// ReadReportHTML reads HTML content from file
func (s *ReportService) ReadReportHTML(report *models.MLReport) ([]byte, error) {
	return os.ReadFile(report.FilePath)
}

// findLatestReport finds the most recently generated report file
func (s *ReportService) findLatestReport(platform string) (string, error) {
	pattern := fmt.Sprintf("%s_ADS_REPORT_*.html", platform)
	if platform == "tiktok" {
		pattern = "TIKTOK_ADS_REPORT_*.html"
	} else if platform == "shopee" {
		pattern = "SHOPEE_ADS_REPORT_*.html"
	}

	files, err := filepath.Glob(filepath.Join(s.reportsDir, pattern))
	if err != nil {
		return "", err
	}

	if len(files) == 0 {
		return "", fmt.Errorf("no report files found matching pattern: %s", pattern)
	}

	// Get the most recent file
	latestFile := files[0]
	latestTime := time.Time{}

	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}
		if info.ModTime().After(latestTime) {
			latestTime = info.ModTime()
			latestFile = file
		}
	}

	return latestFile, nil
}

// DeleteOldReports deletes reports older than specified days
func (s *ReportService) DeleteOldReports(ctx context.Context, tenantID, platform string, olderThanDays int) error {
	cutoff := time.Now().AddDate(0, 0, -olderThanDays)

	var reports []*models.MLReport
	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND platform = ? AND created_at < ?", tenantID, platform, cutoff).
		Find(&reports).Error
	if err != nil {
		return err
	}

	// Delete files and database records
	for _, report := range reports {
		// Delete file
		if report.FilePath != "" {
			os.Remove(report.FilePath) // ignore error
		}

		// Delete database record
		s.db.WithContext(ctx).Delete(report)
	}

	return nil
}
