package handlers

// LoginRequest represents login request body
type LoginRequest struct {
	Username       string `json:"username" binding:"required"`
	Password       string `json:"password" binding:"required"`
	CaptchaToken   string `json:"captcha_token,omitempty"`
	RecaptchaToken string `json:"recaptcha_token,omitempty"`
}

// RefreshTokenRequest represents refresh token request body
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ChangePasswordRequest represents change password request body
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// RegisterRequest represents registration request body
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	ShopName string `json:"shop_name,omitempty"`
}

// SwitchTenantRequest represents switch tenant request body
type SwitchTenantRequest struct {
	TenantID string `json:"tenant_id" binding:"required"`
}

// UpdateProfileRequest represents update profile request body
type UpdateProfileRequest struct {
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty"`
}
