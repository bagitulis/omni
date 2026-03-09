package response

// APIResponse is the standard API response format
// Matches Node.js backend format for compatibility
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// Meta contains pagination and metadata
type Meta struct {
	Total      int `json:"total,omitempty"`
	Page       int `json:"page,omitempty"`
	PageSize   int `json:"page_size,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

// Success creates a successful response
func Success(data interface{}) APIResponse {
	return APIResponse{
		Success: true,
		Data:    data,
	}
}

// SuccessWithMeta creates a successful response with pagination
func SuccessWithMeta(data interface{}, meta *Meta) APIResponse {
	return APIResponse{
		Success: true,
		Data:    data,
		Meta:    meta,
	}
}

// Error creates an error response
func Error(msg string) APIResponse {
	return APIResponse{
		Success: false,
		Error:   msg,
	}
}

// ErrorWithMessage creates an error response with message
func ErrorWithMessage(err, msg string) APIResponse {
	return APIResponse{
		Success: false,
		Error:   err,
		Message: msg,
	}
}

// ErrorWithDetail creates an error response with a summary error and raw detail
func ErrorWithDetail(summary string, detail string) APIResponse {
	return APIResponse{
		Success: false,
		Error:   summary,
		Message: detail,
	}
}
