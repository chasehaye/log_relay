package auth

import (
	"net/http"
	"os"
	"strings"
	"fmt"
	"time"
    "log"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
    "github.com/joho/godotenv"

	"log_relay/internal/models"
	"log_relay/internal/services"
	"log_relay/internal/messaging"
    "log_relay/internal/dtos"
)

var (
    adminEmail string
)

func init() {
    _ = godotenv.Load()
    adminEmail = strings.ToLower(strings.TrimSpace(os.Getenv("ADMIN_EMAIL")))
}



// CreateUser godoc
// @Summary      Register New User
// @Description  Creates a user account, hashes the password, and hashes and API token.
// @Description  Returns a one-time API token and sets an HttpOnly 'token' cookie.
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Param        user  body      RegisterInput  true  "User Registration Data"
// @Header       201   {string}  Set-Cookie     "Contains the JWT session token (HttpOnly, Secure)"
// @Success      201   {object}  RegisterResponse
// @Failure      400   {object}  dtos.ValidationErrorResponse "Validation failed"
// @Failure      409   {object}  dtos.AlreadyExistsResponse   "User already exists with email"
// @Failure      500   {object}  dtos.ServerErrorResponse     "Server error"
// @Router       /api/user/register [post]
func CreateUser(c *gin.Context, db *gorm.DB) {
    var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, dtos.ValidationErrorResponse{
			Error: "Validation failed, incorrect format follow the specifications",
			Details: map[string]string{
				"email":    "Email is required and must be a valid email",
				"password": "Password is required (min 8, max 72 characters)",
				"name":     "Name is optional but must be less than 255 characters",
			},
		})
        return
	}

	cleanEmail, ok := services.CleanAndValidateEmail(c, input.Email)
    if !ok {
        return 
    }

    var existingUser models.User
    result := db.Where("email = ?", cleanEmail).Limit(1).Find(&existingUser)
    if result.Error != nil {
        log.Printf("Database lookup error: %v", result.Error)
        c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{Error: "An unexpected error occurred. Please try again later."})
        return
    }
    if result.RowsAffected > 0 {
        c.JSON(http.StatusConflict, dtos.AlreadyExistsResponse{Error: "Email is taken by another user try again"})
        return
    }

    displayName := strings.TrimSpace(input.Name)
    if displayName == "" {
		displayName = "User" 
	}

    if err := services.ValidatePassword(input.Password); err != nil {
        c.JSON(http.StatusBadRequest, dtos.ValidationErrorResponse{Error: err.Error(),})
        return
    }
	hashedPassword, err := services.HashPassword(input.Password)
	if err != nil {
        log.Printf("Password hashing error: %v", err)
		c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{Error: "An unexpected error occurred. Please try again later."})
		return
	}
    plainToken, hashedToken, err := services.GenerateHashedToken()
    if err != nil {
        log.Printf("Api token generation and hashing failed: %v", err)
        c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{Error: "An unexpected error occurred. Please try again later."})
        return
    }
	isAdmin := false
	if cleanEmail == adminEmail {
		isAdmin = true
	}

    user := models.User{
        Name:     displayName,
        Email:    cleanEmail,
        Password: string(hashedPassword),
        Token:    hashedToken,
        IsAdmin:  isAdmin,
    }
    if err := db.Create(&user).Error; err != nil {
        log.Printf("Failed to create user in database: %v", err)
        c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{Error: "Could not complete registration"})
        return
    }

    jwtToken, err := services.GenerateJWT(user.ID, user.Email)
    if err != nil {
        c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{Error: "Failed to create session"})
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

    c.JSON(http.StatusCreated, RegisterResponse{
		Message:   "Registration successful",
		APIToken:  plainToken,
		IsAdmin:   isAdmin,
		UserEmail: cleanEmail,
		UserName:  displayName,
	})
}

// LoginUser godoc
// @Summary     Login Existing User
// @Description Login for exisitng user using email and passwrod (compares the stored hash to the input).
// @Description Returns a success message and sets an HttpOnly 'token' cookie.
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Param        user  body      LoginInput  true  "User Login Credentials"
// @Header       201   {string}  Set-Cookie     "Contains the JWT session token (HttpOnly, Secure)"
// @Success      201   {object}  LoginResponse
// @Failure      400   {object}  dtos.ValidationErrorResponse "Invalid input or malformed email"
// @Failure      401   {object}  dtos.UnauthorizedResponse "Validation failed"
// @Failure      500   {object}  dtos.ServerErrorResponse     "Server error"
// @Router       /api/user/login [post]
func LoginUser(c *gin.Context, db *gorm.DB) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ValidationErrorResponse{
			Error: "Validation failed, incorrect format follow the specifications",
			Details: map[string]string{
				"email":    "Email is required and must be a valid email",
				"password": "Password is required (min 8, max 72 characters)",
			},
		})
        return
	}

	cleanEmail, ok := services.CleanAndValidateEmail(c, input.Email)
    if !ok {
        return 
    }
	var user models.User
	result := db.Where("email = ?", cleanEmail).Limit(1).Find(&user)
	if result.Error != nil || result.RowsAffected == 0 {
        log.Printf("Database lookup failed: Login attempt with non-existent email: %s", cleanEmail)
        c.JSON(http.StatusUnauthorized, dtos.UnauthorizedResponse{Error: "Invalid credentials"})
        return
    }

    if err := services.ComparePassword(user.Password, input.Password); err != nil {
        c.JSON(http.StatusUnauthorized, dtos.UnauthorizedResponse{Error: "Invalid credentials"})
        return
    }

	jwtToken, err := services.GenerateJWT(user.ID, user.Email)
    if err != nil {
        c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{Error: "Failed to create session"})
        return
    }

    c.SetCookie(
        "token",
        jwtToken,
        86400,
        "/",
        "",
        services.IsProduction(),
        true,
    )

	c.JSON(http.StatusOK, LoginResponse{
		Message:   "Login successful",
		IsAdmin:   user.IsAdmin,
		UserEmail: user.Email,
		UserName:  user.Name,
	})
}

// CycleToken godoc
// @Summary      Rotate API Token
// @Description  Invalidates the old API token and generates a new one. Requires password confirmation and a valid session cookie.
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Param        input  body      PasswordInput  true  "Confirm password to rotate token"
// @Success      200    {object}  CycleTokenResponse
// @Failure      401    {object}  dtos.UnauthorizedResponse "Invalid credentials"
// @Failure      404    {object}  dtos.NotFoundErrorResponse  "Session not valid or not found"
// @Failure      500    {object}  dtos.ServerErrorResponse  "Internal server error"
// @Router       /api/user/cycle-token [post]
func CycleToken(c *gin.Context, db *gorm.DB) {
	var input PasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ValidationErrorResponse{
            Error: "Validation failed, incorrect format follow the specifications",
            Details: map[string]string{
                "password": "Password is required (min 8, max 72 characters)",
            },
        })
        return
	}

    uidValue, _ := c.Get("userID")
    userID := uidValue.(uint)

	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
        log.Printf("Failed to lookup user for token cycling, userID=%d: %v", userID, err)
        c.JSON(http.StatusNotFound, dtos.NotFoundErrorResponse{Error: "Resource not found",})
        return
    }

	if err := services.ComparePassword(user.Password, input.Password); err != nil {
        c.JSON(http.StatusUnauthorized, dtos.UnauthorizedResponse{Error: "Invalid credentials",})
        return
    }

	plainToken, hashedToken, err := services.GenerateHashedToken()
    if err != nil {
        c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{Error: "Failed to cycle token",})
        return
    }

	if err := db.Model(&user).Update("token", hashedToken).Error; err != nil {
        log.Printf("Failed to update API token for userID=%d: %v", userID, err)
        c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{Error: "Failed to update token in database",})
        return
    }

	c.JSON(http.StatusOK, CycleTokenResponse{
        Message:  "API Token updated successfully",
        APIToken: plainToken,
    })
}

// ForgotPassword godoc
// @Summary      Request Password Reset
// @Description  Sends a password reset link to the provided email if the account exists. 
// @Description  Always returns a success message to prevent email enumeration.
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Param        input  body      ForgotPasswordInput  true  "User Email"
// @Success      200    {object}  ForgotPasswordResponse
// @Failure      400    {object}  dtos.ValidationErrorResponse
// @Failure      500    {object}  dtos.ServerErrorResponse
// @Router       /api/user/forgot-password [post]
func ForgotPassword(c *gin.Context, db *gorm.DB){
	var input ForgotPasswordInput
		if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ValidationErrorResponse{
			Error: "Validation failed, incorrect format follow the specifications",
			Details: map[string]string{
				"email": "Email is required and must be a valid email address",
			},
		})
        return
	}

	cleanEmail, ok := services.CleanAndValidateEmail(c, input.Email)
    if !ok {
        return 
    }
	var user models.User
    result := db.Where("email = ?", cleanEmail).Limit(1).Find(&user)
    if result.Error != nil {
        log.Printf("Database error during forgot password lookup for email=%s: %v", cleanEmail, result.Error)
        c.JSON(http.StatusOK, ForgotPasswordResponse{Message: "Check your inbox for a reset link"})
        return
    }
    if result.RowsAffected == 0 {
        c.JSON(http.StatusOK, ForgotPasswordResponse{Message: "Check your inbox for a reset link"})
        return
    }

	token, err := services.GenerateToken()
    if err != nil {
        c.JSON(http.StatusOK, ForgotPasswordResponse{Message: "Check your inbox for a reset link"})
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
        log.Printf("Failed to create password reset record for userID=%d: %v", user.ID, err)
        c.JSON(http.StatusOK, ForgotPasswordResponse{Message: "Check your inbox for a reset link"})
        return
    }
    frontendURL := strings.TrimSuffix(os.Getenv("FRONTEND_URL"), "/")
    resetLink := fmt.Sprintf("%s/reset-password/%s", frontendURL, token)
    
    if err := messaging.SendResetEmail(cleanEmail, resetLink); err != nil {
        log.Printf("Failed to send reset email to %s: %v", cleanEmail, err)
        db.Delete(&resetRecord)
        c.JSON(http.StatusOK, ForgotPasswordResponse{Message: "Check your inbox for a reset link"})
        return
    }
    c.JSON(http.StatusOK, ForgotPasswordResponse{Message: "Check your inbox for a reset link",})
}

// ResetPassword godoc
// @Summary      Change Password
// @Description  Validates the reset token and updates the user's password.
// @Description  Automatically logs the user in by setting a session cookie upon success.
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Param        token  path      string         true  "Reset Token"
// @Param        input  body      PasswordInput  true  "New Password"
// @Header       200    {string}  Set-Cookie     "token=jwt_value; HttpOnly; Secure; SameSite=Lax"
// @Success      200    {object}  ResetPasswordResponse
// @Failure      400    {object}  dtos.ValidationErrorResponse
// @Failure      401    {object}  dtos.UnauthorizedResponse
// @Failure      500    {object}  dtos.ServerErrorResponse
// @Router       /api/user/change-password/{token} [post]
func ResetPassword(c *gin.Context, db *gorm.DB) {
    token := c.Param("token")
    var input PasswordInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, dtos.ValidationErrorResponse{
			Error: "Invalid format",
			Details: map[string]string{
				"password": "Password is required (min 8, max 72 characters)",
			},
		})
        return
    }
    
    var resetRecord models.PasswordReset
    result := db.Preload("User").Where("token = ? AND used = ?", token, false).Limit(1).Find(&resetRecord)
    if result.Error != nil {
        log.Printf("Database error during password reset lookup for token=%s: %v", token, result.Error)
        c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{Error: "An unexpected error occurred. Please try again later."})
        return
    }
    if result.RowsAffected == 0 {
        c.JSON(http.StatusUnauthorized, dtos.UnauthorizedResponse{Error: "Invalid or expired token"})
        return
    }

    if time.Now().After(resetRecord.ExpiresAt) {
        c.JSON(http.StatusUnauthorized, dtos.UnauthorizedResponse{Error: "Expired token"})
        return
    }

    hashedPassword, err := services.HashPassword(input.Password)
    if err != nil {
        log.Printf("Bcrypt hashing failed: %v", err)
		c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{Error: "Failed to process password"})
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
        log.Printf("Failed to update password for userID=%d: %v", resetRecord.UserID, err)
        c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{Error: "Could not complete password reset"})
        return
    }

	jwtToken, err := services.GenerateJWT(resetRecord.User.ID, resetRecord.User.Email)
    if err != nil {
        c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{Error: "Password updated, please login manually"})
        return
    }

    c.SetCookie(
        "token",
        jwtToken,
        86400,
        "/",
        "",
        services.IsProduction(),
        true,
    )

    c.JSON(http.StatusOK, ResetPasswordResponse{
		Message: "Password updated successfully",
	})
}

// LogOut godoc
// @Summary      User Logout
// @Description  Invalidates the session by clearing the 'token' cookie.
// @Tags         accounts
// @Produce      json
// @Success      200  {object}  LogOutResponse
// @Header       200  {string}  Set-Cookie  "token=; Max-Age=0; Path=/; HttpOnly; Secure"
// @Failure      401  {object}  dtos.UnauthorizedResponse "Session expired or invalid"
// @Failure      500  {object}  dtos.ServerErrorResponse
// @Router       /api/user/logout [post]
func LogOut(c *gin.Context, db *gorm.DB) {
    c.SetCookie("token", "", -1, "/", "", services.IsProduction(), true)

    c.JSON(http.StatusOK, LogOutResponse{Message: "Successfully logged out",})
}

// GetMe godoc
// @Summary      Get Current User Info
// @Description  Returns the details of the authenticated user based on the session cookie.
// @Tags         accounts
// @Produce      json
// @Success      200  {object}  LoginResponse
// @Failure      401  {object}  dtos.UnauthorizedResponse
// @Failure      500  {object}  dtos.ServerErrorResponse
// @Router       /api/user/me [get]
func GetMe(c *gin.Context, db *gorm.DB) {
	uidInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dtos.UnauthorizedResponse{Error: "Session context missing"})
		return
	}
	uid, ok := uidInterface.(uint)
	if !ok {
		log.Printf("Type assertion failed: userID in context is %T, expected uint", uidInterface)
		c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{Error: "Internal configuration error"})
		return
	}
	var user models.User
	if err := db.First(&user, uid).Error; err != nil {
		c.JSON(http.StatusNotFound, dtos.NotFoundErrorResponse{Error: "User not found"})
		return
	}

    c.JSON(http.StatusOK, GetMeResponse{
        ID:        user.ID,
        UserName:  user.Name,
        UserEmail: user.Email,
        IsAdmin:   user.IsAdmin,
    })
}