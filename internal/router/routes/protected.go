package routes

import (
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"

    "log_relay/internal/middleware"
    "log_relay/internal/handlers/auth"
    "log_relay/internal/handlers/list"
    "log_relay/internal/handlers/messages"
    "log_relay/internal/handlers/user"
)

func Protected(r *gin.Engine, db *gorm.DB) {
    protected := r.Group("/api")
    protected.Use(middleware.AuthMiddleware())
    {
        protected.GET("/user/me", func(c *gin.Context) { auth.GetMe(c, db) })
        protected.POST("/user/logout", func(c *gin.Context) { auth.LogOut(c, db) })

        protected.POST("/user/cycle-token", func(c *gin.Context) { auth.CycleToken(c, db) })
        protected.PUT("/user/change/username", func(c *gin.Context) { user.ChangeUsername(c, db) })
        protected.PATCH("/user/change/email", func(c *gin.Context) { user.ChangeEmail(c, db) })
        protected.DELETE("/user/account/delete", func(c *gin.Context) { user.DeleteAccount(c, db) })

        protected.POST("/list/create", func(c *gin.Context) { list.CreateList(c, db) })
        protected.DELETE("/list/delete/:id", func(c *gin.Context) { list.DeleteList(c, db) })
        protected.GET("/list/index", func(c *gin.Context) { list.IndexList(c, db) })
        protected.GET("/list/detail/:id", func(c *gin.Context) { list.GetListDetail(c, db) })

        protected.POST("/mail/send/:list_id", func(c *gin.Context) { messages.SendMailingListMessage(c, db) })
    }
}