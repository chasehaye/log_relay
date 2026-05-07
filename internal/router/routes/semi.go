package routes

import (
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"

    "log_relay/internal/handlers"
    "log_relay/internal/handlers/auth"
    "log_relay/internal/handlers/contact"
    "log_relay/internal/handlers/list"
    "log_relay/internal/handlers/user"
)

func Semi(r *gin.Engine, db *gorm.DB) {
    semi := r.Group("/api")
    {
        semi.GET("/status", func(c *gin.Context) { handlers.Ping(c, db) })
        semi.POST("/user/register", func(c *gin.Context) { auth.CreateUser(c, db) })
        semi.POST("/user/login", func(c *gin.Context) { auth.LoginUser(c, db) })
        semi.POST("/user/forgot-password", func(c *gin.Context) { auth.ForgotPassword(c, db) })
        semi.POST("/user/change-password/:token", func(c *gin.Context) { auth.ResetPassword(c, db) })

        semi.POST("/subscriber/signup/:list_id", func(c *gin.Context) { contact.ContactSubscribe(c, db) })
        semi.GET("/subscriber/signup/confirm", func(c *gin.Context) { contact.ContactSubscribeConfirm(c, db) })
        semi.DELETE("/subscriber/remove/:list_id", func(c *gin.Context) { contact.ContactUnSubscribe(c, db) })

        semi.GET("/list/:list_id", func(c *gin.Context) { list.GetListPublicName(c, db) })
        semi.PUT("/user/change/email/confirm", func(c *gin.Context) { user.ChangeEmailConfirm(c, db) })
    }
}