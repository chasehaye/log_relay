package middleware

import (
	"log_relay/internal/services"

	"github.com/gin-gonic/gin"

)

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        tokenString, err := c.Cookie("token")
        
        if err != nil {
            c.AbortWithStatusJSON(401, gin.H{"error": "Not Authenticated"})
            return
        }
        token, err := services.ValidateJWT(tokenString)
        if err != nil || !token.Valid {
            c.AbortWithStatusJSON(401, gin.H{"error": "Invalid session"})
            return
        }
        uid, _ := services.GetUserIDFromJWT(token)
        c.Set("userID", uid)
        c.Next()
    }
}