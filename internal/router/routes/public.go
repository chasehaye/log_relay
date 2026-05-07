package routes

import (
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"

    "log_relay/internal/middleware"
    "log_relay/internal/handlers/messages"
)

func Public(r *gin.Engine, db *gorm.DB) {
    public := r.Group("/api/public")
    public.Use(middleware.ApiMiddleware(db))
    {
        public.POST("/send/message/:list_id", func(c *gin.Context) { messages.SendInboundMessage(c, db) })
    }
}