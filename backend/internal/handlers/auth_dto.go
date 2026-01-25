package handlers

// LoginRequest represents login request body
type LoginRequest struct {
	Username       string `json:"username" binding:"required"`
	Password       string `json:"password" binding:"required"`
	CaptchaToken   string `json:"captchaToken,omitempty"`
	RecaptchaToken string `json:"recaptchaToken,omitempty"`
}

// RefreshTokenRequest represents refresh token request body
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// ChangePasswordRequest represents change password request body
type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=8"`
}

// RegisterRequest represents registration request body
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	ShopName string `json:"shopName,omitempty"`
}

// SwitchTenantRequest represents switch tenant request body
type SwitchTenantRequest struct {
	TenantID string `json:"tenant_id" binding:"required"`
}
