package contact

import(
	"net/http"
	"time"
	"os"
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"log_relay/internal/models"
	"log_relay/internal/services"
	"log_relay/internal/messaging"
)

func ContactSubscribe(c *gin.Context, db *gorm.DB) {
	publicID := c.Param("list_id")
	var input struct {
		Email        string `json:"email" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cleanEmail, ok := services.CleanAndValidateEmail(c, input.Email)
    if !ok {
        return 
    }

	var list models.List
    if err := db.Where("public_id = ?", publicID).First(&list).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Mailing list not found"})
        return
    }
	if list.ListType != "MAILING" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "This list is not available for public subscription",
		})
		return
	}

	token, err := services.GenerateToken()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
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
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error during pre subscribe"})
        return
    }

	frontendURL := os.Getenv("FRONTEND_URL") 
    confirmLink := fmt.Sprintf("%s/confirm-subscription?token=%s&list_id=%s", frontendURL, token, publicID)

    if err := messaging.SendConfirmationEmail(cleanEmail, confirmLink); err != nil {
        c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Email service failed"})
        return
    }


    c.JSON(http.StatusOK, gin.H{"message": "Check inbox to confirm subscription"})
}

func ContactSubscribeConfirm(c *gin.Context, db *gorm.DB) {
	publicID := c.Query("list_id")
	token := c.Query("token")

	var list models.List
    if err := db.Where("public_id = ?", publicID).First(&list).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Mailing list not found"})
        return
    }
	if list.ListType != "MAILING" {
		c.JSON(http.StatusForbidden, gin.H{
        "error": "This list is not available for public subscription",
		})
		return
	}

	var contact models.Contact
    if err := db.Where("verification_token = ?", token).First(&contact).Error; err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired confirmation link"})
        return
    }

	if time.Now().After(contact.TokenExpiresAt) {
        c.JSON(http.StatusGone, gin.H{"error": "Link expired. Please sign up again."})
        return
    }
	if contact.UserID != list.UserID {
        c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized list access"})
        return
    }

	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&list).Association("Subscribers").Append(&contact); err != nil {
			return err
		}
		updates := map[string]interface{}{
			"verified":           true,
			"verification_token": "",
		}

		if contact.UnSubToken == "" {
			newUnSubToken, err := services.GenerateToken()
			if err != nil {
				return err
			}
			updates["un_sub_token"] = newUnSubToken
		}
		return tx.Model(&contact).Updates(updates).Error
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not complete subscription"})
		return
	}

    c.JSON(http.StatusOK, gin.H{
		"message": "Successfully subscribed!",
		"list_name": list.PublicFacingName,
	})
}

func ContactUnSubscribe(c *gin.Context, db *gorm.DB) {
	publicID := c.Param("list_id")
	token := c.Param("token")

	var list models.List
    if err := db.Where("public_id = ?", publicID).First(&list).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Mailing list not found"})
        return
    }
	var contact models.Contact
    if err := db.Where("un_sub_token = ?", token).First(&contact).Error; err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid unsubscribe link"})
        return
    }
	err := db.Model(&contact).Association("SubscribedTo").Delete(&list)
    
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not process unsubscribe"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message": "You have been successfully removed from the list",
        "list":    list.PublicFacingName,
    })
}