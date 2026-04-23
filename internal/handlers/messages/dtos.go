package messages

type SendMessageInput struct {
	Header string `json:"header" binding:"required"`
	Body   string `json:"body" binding:"required"`
}

type SuccessMessageResponse struct {
	Message string `json:"message" example:"Operation successful"`
}