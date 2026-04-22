package list


// --- Request DTOs ---
type ListInput struct {
    Name  string `json:"name" binding:"required"`
    ListType string `json:"list_type" binding:"required"`
    PublicFacingName string `json:"public_facing_name" binding:"required"`
}

type listIndexQuery struct {
    CountPerPage int `form:"count_per_page" binding:"required,min=1,max=50"`
    Page         int `form:"page" binding:"required,min=1"`
}



// --- 200 Response DTOs ------------------------------------------------------------------------------------
type ListResponse struct {
	ID                uint   `json:"id"`
	Name              string `json:"name"`
	ListType          string `json:"list_type"`
	PublicFacingName  string `json:"public_facing_name"`
	PublicID          string `json:"public_id"`
	UserID            uint   `json:"user_id"`
}

type SuccessMessageResponse struct {
	Message string `json:"message" example:"Operation successful"`
}

type ListIndexResponse struct {
	Lists       []ListResponse `json:"lists"`
	TotalCount  int64          `json:"total_count"`
	TotalPages  int            `json:"total_pages"`
	CurrentPage int            `json:"current_page"`
}

type ListDetailResponse struct {
	ID               uint   `json:"id"`
	Name             string `json:"name"`
	ListType         string `json:"list_type"`
	PublicFacingName string `json:"public_facing_name"`
	PublicID         string `json:"public_id"`
	UserID           uint   `json:"user_id"`

	Messages    []MessageResponse   `json:"messages,omitempty"`
	Subscribers []ContactResponse   `json:"subscribers,omitempty"`
}

type MessageResponse struct {
	ID     uint   `json:"id"`
	Header string `json:"header"`
	Body   string `json:"body"`
}

type ContactResponse struct {
	Email string `json:"email"`
}