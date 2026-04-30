package contact

import(
	"net/http"
	"time"
	"os"
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"log_relay/internal/models"
	"log_relay/internal/validation"
	"log_relay/internal/messaging"
	"log_relay/internal/dtos"
	"log_relay/internal/crypt"
)

// ContactSubscribe godoc
// @Summary      Subscribe to mailing list
// @Description  Initiates email verification subscription flow for a mailing list
// @Tags         subscribe
// @Accept       json
// @Produce      json
// @Param        list_id  path      string        true  "Public List ID"
// @Param        request   body      EmailInput    true  "Email subscription input"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  dtos.ValidationErrorResponse
// @Failure      404  {object}  dtos.NotFoundErrorResponse
// @Failure      403  {object}  dtos.ForbiddenResponse
// @Failure      500  {object}  dtos.ServerErrorResponse
// @Failure      503  {object}  dtos.ServerErrorResponse
// @Router       /api/subscriber/signup/{list_id}/subscribe [post]
func ContactSubscribe(c *gin.Context, db *gorm.DB) {
	publicID := c.Param("list_id")
	var input EmailInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ValidationErrorResponse{
			Error: "Invalid email input",
		})
		return
	}

	cleanEmail, ok := validation.CleanAndValidateEmail(c, input.Email)
    if !ok {
        return 
    }

	var list models.List
    if err := db.Where("public_id = ?", publicID).First(&list).Error; err != nil {
        c.JSON(http.StatusNotFound, dtos.NotFoundErrorResponse{
			Error: "Mailing list not found",
		})
        return
    }
	if list.ListType != models.ListTypeMailing {
		c.JSON(http.StatusForbidden, dtos.ForbiddenResponse{
			Error: "This list is not available for public subscription",
		})
		return
	}

	token, err := crypt.GenerateToken()
    if err != nil {
        c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{
			Error: "Could not generate token",
		})
        return
    }

	var contact models.Contact
    err = db.Transaction(func(tx *gorm.DB) error {
        if err := tx.Where(models.Contact{
            Email:  cleanEmail,
            UserID: list.UserID,
        }).FirstOrCreate(&contact).Error; err != nil {
            return err
        }
        return tx.Model(&contact).Updates(models.Contact{
            VerificationToken: token,
            TokenExpiresAt:    time.Now().Add(24 * time.Hour),
            Verified:          false, 
        }).Error
    })
	if err != nil {
        c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{
			Error: "Database error during pre subscribe",
		})
        return
    }

	frontendURL := os.Getenv("FRONTEND_URL") 
    confirmLink := fmt.Sprintf("%s/confirm-subscription?token=%s&list_id=%s", frontendURL, token, publicID)

    if err := messaging.SendConfirmationEmail(cleanEmail, confirmLink); err != nil {
        c.JSON(http.StatusServiceUnavailable, dtos.ServerErrorResponse{
			Error: "Email service failed",
		})
        return
    }


    c.JSON(http.StatusOK, SuccessResponse{
		Message: "Check inbox to confirm subscription",
	})
}

// ContactSubscribeConfirm godoc
// @Summary      Confirm subscription
// @Description  Confirms email subscription after verification
// @Tags         subscribe
// @Accept       json
// @Produce      json
// @Param        list_id  query     string  true  "Public List ID"
// @Param        token    query     string  true  "Verification Token"
// @Success      200  {object}  SubscribeConfirmResponse
// @Failure      401  {object}  dtos.UnauthorizedResponse
// @Failure      403  {object}  dtos.ForbiddenResponse
// @Failure      404  {object}  dtos.NotFoundErrorResponse
// @Failure      410  {object}  dtos.ServerErrorResponse
// @Failure      500  {object}  dtos.ServerErrorResponse
// @Router       /api/subscriber/signup/confirm [get]
func ContactSubscribeConfirm(c *gin.Context, db *gorm.DB) {
	publicID := c.Query("list_id")
	token := c.Query("token")

	var list models.List
    if err := db.Where("public_id = ?", publicID).First(&list).Error; err != nil {
        c.JSON(http.StatusNotFound, dtos.NotFoundErrorResponse{
			Error: "Mailing list not found",
		})
        return
    }
	if list.ListType != "MAILING" {
		c.JSON(http.StatusForbidden, dtos.ForbiddenResponse{
			Error: "This list is not available for public subscription",
		})
		return
	}

	var contact models.Contact
    if err := db.Where("verification_token = ?", token).First(&contact).Error; err != nil {
		c.JSON(http.StatusOK, SubscribeConfirmResponse{
			Message: "Already confirmed or invalid link",
		})
		return
	}

	if time.Now().After(contact.TokenExpiresAt) {
		c.JSON(http.StatusGone, dtos.ServerErrorResponse{
			Error: "Link expired. Please sign up again.",
		})
        return
    }
	if contact.UserID != list.UserID {
		c.JSON(http.StatusForbidden, dtos.ForbiddenResponse{
			Error: "Unauthorized list access",
		})
        return
    }

	err := db.Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&list).Association("Subscribers").Append(&contact)
		if err != nil {
			return err
		}
		
		updates := map[string]interface{}{
			"verified":           true,
			"verification_token": "",
		}

		if contact.UnSubToken == "" {
			newUnSubToken, err := crypt.GenerateToken()
			if err != nil {
				return err
			}
			updates["un_sub_token"] = newUnSubToken
		}
		return tx.Model(&contact).Updates(updates).Error
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{
			Error: "Could not complete subscription",
		})
		return
	}

	c.JSON(http.StatusOK, SubscribeConfirmResponse{
		Message:  "Successfully subscribed!",
		ListName: list.PublicFacingName,
	})
}

// ContactUnSubscribe godoc
// @Summary      Unsubscribe from mailing list
// @Description  Removes a contact from a mailing list using unsubscribe token
// @Tags         subscribe
// @Accept       json
// @Produce      json
// @Param        list_id  path      string  true  "Public List ID"
// @Param        input    body      UnsubscribeInput   true  "Unsubscribe payload"
// @Success      200  {object}  UnsubscribeResponse
// @Failure      401  {object}  dtos.UnauthorizedResponse
// @Failure      404  {object}  dtos.NotFoundErrorResponse
// @Failure      500  {object}  dtos.ServerErrorResponse
// @Router       /api/subscriber/remove/{list_id} [delete]
func ContactUnSubscribe(c *gin.Context, db *gorm.DB) {
	publicID := c.Param("list_id")
	var input UnsubscribeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	var list models.List
    if err := db.Where("public_id = ?", publicID).First(&list).Error; err != nil {
        c.JSON(http.StatusNotFound, dtos.NotFoundErrorResponse{
			Error: "Mailing list not found",
		})
        return
    }
	var contact models.Contact
    if err := db.Where("un_sub_token = ?", input.Token).First(&contact).Error; err != nil {
        c.JSON(http.StatusUnauthorized, dtos.UnauthorizedResponse{
			Error: "Invalid unsubscribe link",
		})
        return
    }
	err := db.Model(&contact).Association("SubscribedTo").Delete(&list)
    
    if err != nil {
        c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{
			Error: "Could not process unsubscribe",
		})
        return
    }

	c.JSON(http.StatusOK, UnsubscribeResponse{
		Message: "You have been successfully removed from the list",
		List:    list.PublicFacingName,
	})
}