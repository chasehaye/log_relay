package messages

type SendMessageInput struct {
	Header string `json:"header" binding:"required"`
	Body   string `json:"body" binding:"required"`
}

type SuccessMessageResponse struct {
	Message string `json:"message" example:"Operation successful"`
}

type SendBugReportInput struct {
	Email string `json:"email" binding:"required,email,max=255" example:"user@example.com"`
	Header string `json:"header" binding:"required"`
	Body   string `json:"body" binding:"required"`
}

type SendInquiryInput struct {
	Email string `json:"email" binding:"required,email,max=255" example:"user@example.com"`
	Header string `json:"header" binding:"required"`
	Body   string `json:"body" binding:"required"`
}