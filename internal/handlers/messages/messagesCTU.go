package messages

import (
	"net/http"
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"log_relay/internal/dtos"
	"log_relay/internal/models"
)

// SendInboundMessage godoc
// @Summary      Send inbound message to a list
// @Description  Creates or reuses a contact and stores an inbound message associated with a specific list
// @Tags         messages
// @Accept       json
// @Produce      json
// @Param        list_id  path      string               true  "Public List ID"
// @Param        request  body      SendBugReportInput   true  "Inbound message payload"
// @Success      201      {object}  SuccessMessageResponse
// @Failure      400      {object}  dtos.ValidationErrorResponse
// @Failure      401      {object}  dtos.UnauthorizedResponse
// @Failure      403      {object}  dtos.ForbiddenResponse
// @Failure      404      {object}  dtos.NotFoundErrorResponse
// @Failure      500      {object}  dtos.ServerErrorResponse
// @Router       /api/send/message/{list_id} [post]
func SendInboundMessage(c *gin.Context, db *gorm.DB) {
    list := c.MustGet("list").(models.List)

    var input SendBugReportInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ValidationErrorResponse{
			Error: "Invalid format",
			Details: map[string]string{
				"email":  "Valid email is required",
				"header": "Header is required",
				"body":   "Body is required",
			},
		})
		return
	}

    contact := models.Contact{
        Email:  input.Email,
        UserID: list.UserID,
    }
    err := db.Where(&contact).FirstOrCreate(&contact).Error
    if err != nil {
        log.Printf("Contact upsert error: %v", err)
        c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{
            Error: "Internal Server Error",
        })
        return
    }

    message := models.Message{
        Header: input.Header,
        Body:   input.Body,
        ListID: list.ID,
        ContactID: &contact.ID,
        Type:   models.MessageTypeInbound,
    }

    if err := db.Create(&message).Error; err != nil {
        log.Printf("DB insert error: %v", err)

        c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{
            Error: "Internal Server Error",
        })
        return
    }

    c.JSON(http.StatusCreated, SuccessMessageResponse{
        Message: "ok",
    })
}
