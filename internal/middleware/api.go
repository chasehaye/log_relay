package middleware

import (
	"log_relay/internal/models"
    "net/http"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
    "log_relay/internal/dtos"
	"log"
	"errors"
    "crypto/sha256"
    "encoding/hex"
)


func ApiMiddleware(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {

        apiKey := c.GetHeader("X-API-Key")
        if apiKey == "" {
            c.JSON(http.StatusUnauthorized, dtos.UnauthorizedResponse{
                Error: "missing api key",
            })
            c.Abort()
            return
        }

        hash := sha256.Sum256([]byte(apiKey))
        hashedToken := hex.EncodeToString(hash[:])
        var user models.User
        if err := db.Where("token = ?", hashedToken).First(&user).Error; err != nil {
            c.JSON(http.StatusUnauthorized, dtos.UnauthorizedResponse{
                Error: "invalid api key",
            })
            c.Abort()
            return
        }

        publicListID := c.Param("list_id")

        var list models.List
        if err := db.Where("public_id = ?", publicListID).First(&list).Error; err != nil {
            if errors.Is(err, gorm.ErrRecordNotFound) {
                c.JSON(http.StatusNotFound, dtos.NotFoundErrorResponse{
                    Error: "list not found",
                })
                c.Abort()
                return
            }

            log.Printf("DB error: %v", err)

            c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{
                Error: "Internal Server Error",
            })
            c.Abort()
            return
        }

        if list.UserID != user.ID {
            c.JSON(http.StatusForbidden, dtos.ForbiddenResponse{
                Error: "you do not have access to this list",
            })
            c.Abort()
            return
        }

        c.Set("user", user)
        c.Set("list", list)

        c.Next()
    }
}