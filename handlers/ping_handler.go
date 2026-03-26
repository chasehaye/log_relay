package handlers

import (
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    "net/http"
)

func Ping(c *gin.Context, db *gorm.DB) {
    sqlDB, err := db.DB()
    if err != nil || sqlDB.Ping() != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"status": "unhealthy", "database": "disconnected"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "status":  "healthy",
    })
}