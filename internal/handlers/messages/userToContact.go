package messages

import(
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"log_relay/internal/models"
	"log_relay/internal/dtos"
	"log_relay/internal/services"
)

// SendMailingListMessage godoc
// @Summary      Send message to mailing list
// @Description  Creates a message and queues it for sending to all subscribers in a list
// @Tags         messages
// @Accept       json
// @Produce      json
// @Param        list_id path int true "Mailing List ID"
// @Param        message body SendMessageInput true "Message payload"
// @Success      201 {object} SuccessMessageResponse
// @Failure      400 {object} dtos.ValidationErrorResponse
// @Failure      401 {object} dtos.UnauthorizedResponse
// @Failure      404 {object} dtos.NotFoundErrorResponse
// @Failure      500 {object} dtos.ServerErrorResponse
// @Router       /api/mail/send/{list_id} [post]
func SendMailingListMessage(c *gin.Context, db *gorm.DB) {
	var input SendMessageInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ValidationErrorResponse{
			Error: "Invalid Email input",
		})
		return
	}

	uidValue, exists := c.Get("userID")
    if !exists {
        c.JSON(http.StatusUnauthorized, dtos.UnauthorizedResponse{
            Error: "Invalid session",
        })
        return
    }

	userID := uidValue.(uint)

	privateID, err := strconv.Atoi(c.Param("list_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid list id",
		})
		return
	}

	var list models.List
	if err := db.Where("id = ? AND user_id = ?", privateID, userID).First(&list).Error; err != nil {
		c.JSON(http.StatusNotFound, dtos.NotFoundErrorResponse{
			Error: "Not found or no permission",
		})
		return
	}

	message := models.Message{
		Header: input.Header,
		Body: input.Body,
		ListID: list.ID,
		Type: "OUTBOUND",
	}

	if err := db.Create(&message).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{
			Error: "Failed to add message to the database",
		})
		return
	}

	c.JSON(http.StatusCreated, SuccessMessageResponse{
		Message: "Message created and queued for sending",
	})

	go func(messageID uint) {
		services.SendMailingList(messageID, db)
	}(message.ID)

}