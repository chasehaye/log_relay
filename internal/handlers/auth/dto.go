package auth

// --- Request DTOs ---

// RegisterInput defines the data needed to create a new account
type RegisterInput struct {
	Name     string `json:"name" binding:"max=255" example:"User Name"`
	Email    string `json:"email" binding:"required,email,max=255" example:"user@example.com"`
	Password string `json:"password" binding:"required,min=8,max=72" example:"SecurePass123!"`
}

// LoginInput defines the data needed to authenticate
type LoginInput struct {
	Email    string `json:"email" binding:"required,email,max=255" example:"user@example.com"`
	Password string `json:"password" binding:"required,min=8,max=72" example:"SecurePass123!"`
}

// CycleToken and ResetPassword requires the password for security confirmation
type PasswordInput struct {
	Password string `json:"password" binding:"required,min=8,max=72" example:"SecurePass123!"`
}

// ForgotPasswordInput defines the data needed to send an email to the user
type ForgotPasswordInput struct {
	Email    string `json:"email" binding:"required,email,max=255" example:"user@example.com"`
}

// --- 200 Response DTOs ------------------------------------------------------------------------------------

// RegisterResponse defines the successful registration output
type RegisterResponse struct {
	Message   string `json:"message" example:"Registration successful"`
	APIToken  string `json:"api_token" example:"abcdef_gh1234!@#zxcv"`
	IsAdmin   bool   `json:"is_admin" example:"false"`
	UserEmail string `json:"user_email" example:"user@example.com"`
	UserName  string `json:"user_name" example:"User Name"`
}

// RegisterResponse defines the successful login output
type LoginResponse struct {
    Message   string `json:"message" example:"Login successful"`
    IsAdmin   bool   `json:"is_admin" example:"false"`
    UserEmail string `json:"user_email" example:"user@example.com"`
    UserName  string `json:"user_name" example:"User Name"`
}

// CycleTokenResponse returns the brand new token
type CycleTokenResponse struct {
	Message  string `json:"message" example:"API Token updated successfully"`
	APIToken string `json:"api_token" example:"new_lr_7f8e9d0a1b2c3d4e5f..."`
}

// ForgotPasswordResponse returns a success message with instructions
type ForgotPasswordResponse struct {
	Message  string `json:"message" example:"Check your inbox for a reset link"`
}

// ResetPasswordResponse returns a success message
type ResetPasswordResponse struct {
	Message  string `json:"message" example:"Password updated successfully"`
}

// LogOutResponse resturns a success message
type LogOutResponse struct {
	Message  string `json:"message" example:"Successfully logged out"`
}

// GetMeResponse retruns user information
type GetMeResponse struct {
    ID        uint   `json:"id" example:"1"`
    UserName  string `json:"user_name" example:"User Name"`
    UserEmail string `json:"user_email" example:"user@example.com"`
    IsAdmin   bool   `json:"is_admin" example:"false"`
}