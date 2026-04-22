package contact

type EmailInput struct {
	Email string `json:"email" binding:"required"`
}

type SuccessResponse struct {
	Message string `json:"message" example:"Operation successful"`
}

type SubscribeConfirmResponse struct {
	Message  string `json:"message"`
	ListName string `json:"list_name"`
}

type UnsubscribeResponse struct {
	Message string `json:"message"`
	List    string `json:"list"`
}