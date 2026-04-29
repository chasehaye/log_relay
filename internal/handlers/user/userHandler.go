package user

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"log_relay/internal/crypt"
	"log_relay/internal/dtos"
	"log_relay/internal/messaging"
	"log_relay/internal/models"
	"log_relay/internal/validation"
)

var (
    frontendURL string
)

func init() {
    frontendURL = strings.TrimSuffix(os.Getenv("FRONTEND_URL"), "/")
}


// ChangeUsername godoc
// @Summary Change username
// @Description Updates the authenticated user's display name
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body ChangeUsernameInput true "New username"
// @Success 200 {object} SuccessMessageResponse
// @Failure 400 {object} dtos.ValidationErrorResponse
// @Failure 401 {object} dtos.UnauthorizedResponse
// @Failure 404 {object} dtos.NotFoundErrorResponse
// @Failure 500 {object} dtos.ServerErrorResponse
// @Router /user/change/username [patch]
func ChangeUsername(c *gin.Context, db *gorm.DB) {
	var input ChangeUsernameInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ValidationErrorResponse{
			Error: "Invalid username format",
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

	result := db.Model(&models.User{}).
		Where("id = ?", userID).
		Update("name", input.Name)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{
			Error: "Failed to update username",
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, dtos.NotFoundErrorResponse{
			Error: "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessMessageResponse{
		Message: "username updated successfully",
	})
}

// ChangeEmailRequest godoc
// @Summary      Request email change
// @Description  Sends a confirmation email to complete the email change process
// @Tags         user
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        input  body      ChangeEmailInput  true  "New email address"
// @Success 200 {object} SuccessMessageResponse
// @Failure 400 {object} dtos.ValidationErrorResponse
// @Failure 401 {object} dtos.UnauthorizedResponse
// @Failure 404 {object} dtos.NotFoundErrorResponse
// @Failure 409 {object} dtos.AlreadyExistsResponse
// @Failure 500 {object} dtos.ServerErrorResponse
// @Router /user/change/email [patch]
func ChangeEmail(c *gin.Context, db *gorm.DB) {
	var input ChangeEmailInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ValidationErrorResponse{
			Error: "Invalid username format",
		})
		return
	}
	cleanEmail, ok := validation.CleanAndValidateEmail(c, input.Email)
	if !ok {
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

	var count int64
	db.Model(&models.User{}).
		Where("email = ?", cleanEmail).
		Where("id != ?", userID).
		Count(&count)

	if count > 0 {
		c.JSON(http.StatusConflict, dtos.AlreadyExistsResponse{
			Error: "Email already in use or invalid",
		})
		return
	}

	token, err := crypt.GenerateToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{
			Error: "Server Error",
		})
		return
	}
	req := models.EmailChangeRequest{
		UserID:    userID,
		NewEmail:  cleanEmail,
		Token:     token,
		ExpiresAt: time.Now().Add(30 * time.Minute),
	}

	if err := db.Create(&req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{
			Error: "Failed to create change request",
		})
		return
	}
	link := fmt.Sprintf("%s/change-email/confirm?token=%s", frontendURL, token)

	err = messaging.SendResetEmailChangeEmail(cleanEmail, link)

	c.JSON(http.StatusOK, SuccessMessageResponse{
		Message: "Confirmation sent check inbox",
	})
}


func ChangeEmailConfirm(c *gin.Context, db *gorm.DB) {
	var input ChangeEmailConfirmInput

	if err := c.ShouldBindJSON(&input); err != nil || input.Token == "" {
		c.JSON(http.StatusBadRequest, dtos.ValidationErrorResponse{
			Error: "Missing token",
		})
		return
	}

	token := input.Token

	var req models.EmailChangeRequest
	err := db.Where("token = ?", token).First(&req).Error
	if err != nil {
		c.JSON(http.StatusNotFound, dtos.NotFoundErrorResponse{
			Error: "Invalid or expired token",
		})
		return
	}

	if req.Used {
		c.JSON(http.StatusBadRequest, dtos.ValidationErrorResponse{
			Error: "Token already used",
		})
		return
	}

	if time.Now().After(req.ExpiresAt) {
		c.JSON(http.StatusBadRequest, dtos.ValidationErrorResponse{
			Error: "Token expired",
		})
		return
	}

	var user models.User
	if err := db.First(&user, req.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, dtos.NotFoundErrorResponse{
			Error: "User not found",
		})
		return
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&user).Update("email", req.NewEmail).Error; err != nil {
			return err
		}

		if err := tx.Model(&req).Update("used", true).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{
			Error: "Failed to update email",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessMessageResponse{
		Message: "Email updated successfully",
	})
}

// DeleteAccount godoc
// @Summary Delete user account
// @Description Permanently deletes the authenticated user's account after verifying password
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body PasswordInput true "User password confirmation"
// @Success 200 {object} SuccessMessageResponse
// @Failure 400 {object} dtos.ValidationErrorResponse
// @Failure 401 {object} dtos.UnauthorizedResponse
// @Failure 404 {object} dtos.NotFoundErrorResponse
// @Failure 500 {object} dtos.ServerErrorResponse
// @Router /user/account/delete [delete]
func DeleteAccount(c *gin.Context, db *gorm.DB) {
	var input PasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ValidationErrorResponse{
			Error: "Invalid password format",
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

	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
        log.Printf("Failed to find user during account deletion, userID=%d: %v", userID, err)
        c.JSON(http.StatusNotFound, dtos.NotFoundErrorResponse{Error: "Resource not found",})
        return
    }

	if err := validation.ComparePassword(user.Password, input.Password); err != nil {
        c.JSON(http.StatusUnauthorized, dtos.UnauthorizedResponse{Error: "Invalid credentials",})
        return
    }

	if err := db.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{
			Error: "Failed to delete account",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessMessageResponse{
		Message: "account deleted successfully",
	})
}