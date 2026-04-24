package user

type ChangeUsernameInput struct {
	Name string `json:"name" binding:"max=255" example:"User Name"`
}

type ChangeEmailInput struct {
	Email string `json:"email" binding:"required,email,max=255" example:"user@example.com"`
}

type SuccessMessageResponse struct {
	Message  string `json:"message" example:"Successfully updated the username"`
}

type PasswordInput struct {
	Password string `json:"password" binding:"required,min=8,max=72" example:"SecurePass123!"`
}