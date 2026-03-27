package handlers

import (
	"net/http"
	"os"
	"strings"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"log_relay/internal/models"
	"log_relay/internal/services"
	"log_relay/internal/messaging"
)

type RegisterInput struct {
	Name          string `json:"name"`
	ReceiverEmail string `json:"email" binding:"required,email"`
	Password      string `json:"password" binding:"required,min=8"`
}

// receiver is user for reference pardon the poor naming convention
// IN THE FUTURE MAKE IT SO THAT I HAVE TO APPROVE CREATION ON ADMIN END
func CreateReceiver(c *gin.Context, db *gorm.DB) {
	var inputData RegisterInput

	if err := c.ShouldBindJSON(&inputData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cleanEmail := strings.ToLower(strings.TrimSpace(inputData.ReceiverEmail))
	displayName := strings.TrimSpace(inputData.Name)

	if displayName == "" {
		displayName = "User" 
	}
	if err := services.ValidateEmail(cleanEmail); err != nil {
        switch err {
        case services.ErrInvalidFormat:
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
        case services.ErrInvalidDomain:
            c.JSON(http.StatusBadRequest, gin.H{"error": "Email domain does not exist or cannot receive mail"})
        default:
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Verification service error"})
        }
        return
    }
	hashedPassword, err := services.HashPassword(inputData.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "An unexpected error occurred. Please try again later."})
		return
	}
    token, err := services.GenerateToken()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
        return
    }


	isAdmin := false
	if cleanEmail == os.Getenv("ADMIN_EMAIL") {
		isAdmin = true
	}
	receiver := models.Receiver{
		Name:          displayName,
		ReceiverEmail: cleanEmail,
		Password:      string(hashedPassword),
		Token:         token,
		IsAdmin:       isAdmin,
	}
	if err := db.Create(&receiver).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Registration failed: Email already in use"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Registration successful",
		"api_token": token,
		"is_admin": isAdmin,
		// can return email or name if ever needed
	})
}

type LoginInput struct {
	ReceiverEmail string `json:"email" binding:"required,email"`
	Password      string `json:"password" binding:"required"` 
}

func LoginReceiver(c *gin.Context, db *gorm.DB) {
	var inputData LoginInput

	if err := c.ShouldBindJSON(&inputData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cleanEmail := strings.ToLower(strings.TrimSpace(inputData.ReceiverEmail))
	
	var receiver models.Receiver
	if err := db.Where("receiver_email = ?", cleanEmail).First(&receiver).Error; err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
        return
    }

	if err := services.ComparePassword(receiver.Password, inputData.Password); err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
        return
    }

	jwtToken, err := services.GenerateJWT(receiver.ID, receiver.ReceiverEmail)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
        return
    }

	c.JSON(http.StatusOK, gin.H{
        "message": "Login successful",
		"token": jwtToken,
        "is_admin": receiver.IsAdmin,
		// can return email or name if ever needed
    })
}

type CycleInput struct {
	ReceiverEmail string `json:"email" binding:"required,email"`
	Password      string `json:"password" binding:"required"`
}

func CycleToken(c *gin.Context, db *gorm.DB) {
	var input CycleInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cleanEmail := strings.ToLower(strings.TrimSpace(input.ReceiverEmail))
	
	var receiver models.Receiver
	if err := db.Where("receiver_email = ?", cleanEmail).First(&receiver).Error; err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
        return
    }

	if err := services.ComparePassword(receiver.Password, input.Password); err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
        return
    }

	newToken, err := services.GenerateToken()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate secure token"})
        return
    }

	if err := db.Model(&receiver).Update("token", newToken).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update token in database"})
        return
    }

	c.JSON(http.StatusOK, gin.H{
        "message": "API Token cycled successfully",
        "api_token": newToken,
        "note": "All previous static tokens are now invalid.",
    })
}

type emailInput struct {
	ReceiverEmail string `json:"email" binding:"required,email"`
}

func ForgotPassword(c *gin.Context, db *gorm.DB){
	var input emailInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cleanEmail := strings.ToLower(strings.TrimSpace(input.ReceiverEmail))

	var user models.Receiver
	if err := db.Where("receiver_email = ?", cleanEmail).First(&user).Error; err != nil {
		// Security: don't reveal if email exists
		c.JSON(http.StatusOK, gin.H{"message": "Check your inbox for a reset link"})
		return
	}

	token, err := services.GenerateToken()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
        return
    }

	db.Model(&models.PasswordReset{}).
        Where("user_id = ? AND used = ?", user.ID, false).
        Update("used", true)
    resetRecord := models.PasswordReset{
        UserID:    user.ID,
        Token:     token,
        ExpiresAt: time.Now().Add(15 * time.Minute),
    }
    if err := db.Create(&resetRecord).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create reset link"})
        return
    }
    frontendURL := strings.TrimSuffix(os.Getenv("FRONTEND_URL"), "/")
    resetLink := fmt.Sprintf("%s/reset-password?token=%s", frontendURL, token)

    if err := messaging.SendResetEmail(cleanEmail, resetLink); err != nil {
        db.Delete(&resetRecord)
        c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Email service unavailable"})
        return
    }
	fmt.Print("AFTER")
    c.JSON(http.StatusOK, gin.H{"message": "Check your inbox for a reset link"})
}
func ResetPassword(c *gin.Context, db *gorm.DB){
// use the token given to search the database from all the present tokens
// then reset the password for the user
// return a new jwt session token
}