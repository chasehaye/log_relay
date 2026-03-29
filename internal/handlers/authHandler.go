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

func CreateUser(c *gin.Context, db *gorm.DB) {
	var input struct {
        Name     string `json:"name"`
        Email    string `json:"email" binding:"required,email"`
        Password string `json:"password" binding:"required,min=8"`
    }
	if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cleanEmail := strings.ToLower(strings.TrimSpace(input.Email))
	displayName := strings.TrimSpace(input.Name)
    if err := services.ValidateEmail(cleanEmail); err != nil {
        code, msg := services.HandleEmailError(err)
        c.JSON(code, gin.H{"error": msg})
        return
    }
    var existingUser models.User
    result := db.Where("email = ?", cleanEmail).Limit(1).Find(&existingUser)
    if result.Error != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database lookup failed"})
        return
    }

    if result.RowsAffected > 0 {
        c.JSON(http.StatusConflict, gin.H{"error": "Email already in use"})
        return
    }

    if displayName == "" {
		displayName = "User" 
	}
	hashedPassword, err := services.HashPassword(input.Password)
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

    user := models.User{
        Name:     displayName,
        Email:    cleanEmail,
        Password: string(hashedPassword),
        Token:    token,
        IsAdmin:  isAdmin,
    }
    if err := db.Create(&user).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: could not complete registration"})
        return
    }

    jwtToken, err := services.GenerateJWT(user.ID, user.Email)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
        return
    }

    c.SetCookie(
        "token",    // Name
        jwtToken,   // Value
        86400,      // MaxAge (24 hours in seconds)
        "/",        // Path
        "",         // Domain (leave empty for current domain)
        services.IsProduction(),      // Secure (SET TO TRUE IN PRODUCTION/HTTPS)
        true,       // HttpOnly (CRITICAL: prevents JS access)
    )

	c.JSON(http.StatusCreated, gin.H{
		"message": "Registration successful",
		"api_token": token,
		"is_admin": isAdmin,
        "user_email": cleanEmail,
        "user_name":  displayName,
	})
}

func LoginUser(c *gin.Context, db *gorm.DB) {
	var input struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"` 
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cleanEmail := strings.ToLower(strings.TrimSpace(input.Email))
	
	var user models.User
	if err := db.Where("email = ?", cleanEmail).First(&user).Error; err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
        return
    }

	if err := services.ComparePassword(user.Password, input.Password); err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
        return
    }

	jwtToken, err := services.GenerateJWT(user.ID, user.Email)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
        return
    }

    c.SetCookie(
        "token",    // Name
        jwtToken,   // Value
        86400,      // MaxAge
        "/",        // Path
        "",         // Domain
        services.IsProduction(),      // SET TO TRUE IN PROD
        true,       // HttpOnly
    )

	c.JSON(http.StatusOK, gin.H{
        "message": "Login successful",
        "is_admin": user.IsAdmin,
		"user_email": user.Email,
        "user_name":  user.Name,
    })
}

func CycleToken(c *gin.Context, db *gorm.DB) {
	var input struct {
        Email    string `json:"email" binding:"required,email"`
        Password string `json:"password" binding:"required"`
    }
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cleanEmail := strings.ToLower(strings.TrimSpace(input.Email))
	
	var user models.User
	if err := db.Where("email = ?", cleanEmail).First(&user).Error; err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
        return
    }

	if err := services.ComparePassword(user.Password, input.Password); err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
        return
    }

	newToken, err := services.GenerateToken()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate secure token"})
        return
    }

	if err := db.Model(&user).Update("token", newToken).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update token in database"})
        return
    }

	c.JSON(http.StatusOK, gin.H{
        "message": "API Token updated successfully",
        "api_token": newToken,
    })
}

func ForgotPassword(c *gin.Context, db *gorm.DB){
	var input struct {
        Email string `json:"email" binding:"required,email"`
    }
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cleanEmail := strings.ToLower(strings.TrimSpace(input.Email))

	var user models.User
	if err := db.Where("email = ?", cleanEmail).First(&user).Error; err != nil {
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
        ExpiresAt: time.Now().Add(5 * time.Minute),
    }
    if err := db.Create(&resetRecord).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create reset link"})
        return
    }
    frontendURL := strings.TrimSuffix(os.Getenv("FRONTEND_URL"), "/")
    resetLink := fmt.Sprintf("%s/reset-password/%s", frontendURL, token)
    
    if err := messaging.SendResetEmail(cleanEmail, resetLink); err != nil {
        db.Delete(&resetRecord)
        c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Email service unavailable"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "Check your inbox for a reset link"})
}

func ResetPassword(c *gin.Context, db *gorm.DB) {
    token := c.Param("token")
    var input struct {
        NewPassword string `json:"new_password" binding:"required,min=8"`
    }
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid format"})
        return
    }

    var resetRecord models.PasswordReset
    if err := db.Preload("User").Where("token = ? AND used = ?", token, false).First(&resetRecord).Error; err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
        return
    }
    if time.Now().After(resetRecord.ExpiresAt) {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Token expired"})
        return
    }

    hashedPassword, err := services.HashPassword(input.NewPassword)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating password"})
        return
    }

    err = db.Transaction(func(action *gorm.DB) error {
        if err := action.Model(&models.User{}).Where("id = ?", resetRecord.UserID).Update("password", hashedPassword).Error; err != nil {
            return err
        }
        if err := action.Model(&resetRecord).Update("used", true).Error; err != nil {
            return err
        }
        return nil
    })

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not complete password reset"})
        return
    }

	jwtToken, err := services.GenerateJWT(resetRecord.User.ID, resetRecord.User.Email)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
        return
    }

    c.SetCookie(
        "token",    // Name
        jwtToken,   // Value
        86400,      // MaxAge (24 hours in seconds)
        "/",        // Path
        "",         // Domain (leave empty for current domain)
        services.IsProduction(),      // Secure (SET TO TRUE IN PRODUCTION/HTTPS)
        true,       // HttpOnly (CRITICAL: prevents JS access)
    )

    c.JSON(http.StatusOK, gin.H{
        "message": "Password updated successfully",
    })
}

func LogOut(c *gin.Context, db *gorm.DB) {
    c.SetCookie("token", "", -1, "/", "", services.IsProduction(), true)

    c.JSON(http.StatusOK, gin.H{
        "message": "Successfully logged out",
    })
}

