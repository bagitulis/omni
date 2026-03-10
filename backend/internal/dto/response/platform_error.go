package response

import "fmt"

// PlatformError represents a structured error from a third-party platform API
type PlatformError struct {
	Platform string `json:"platform"`
	Code     string `json:"code"`
	Message  string `json:"message"`
}

// Error implements the error interface
func (e *PlatformError) Error() string {
	if e.Code == "" {
		return fmt.Sprintf("%s API error: %s", e.Platform, e.Message)
	}
	return fmt.Sprintf("%s API error (code %s): %s", e.Platform, e.Code, e.Message)
}

// ErrorWithPlatform creates a platform-specific error response with structured platform error data
func ErrorWithPlatform(platform, code, message string) APIResponse {
	pe := &PlatformError{
		Platform: platform,
		Code:     code,
		Message:  message,
	}
	return APIResponse{
		Success: false,
		Error:   pe.Error(),
		Data:    pe,
	}
}
