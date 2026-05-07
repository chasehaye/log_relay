package handlers

import (
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    "net/http"
)

type StatusResponse struct {
    Status   string `json:"status" example:"healthy?"`
    Database string `json:"database,omitempty" example:"conntect?"`
}

// Ping godoc
// @Summary      Health Check
// @Description  Checks if the API and Database are reachable.
// @Tags         system
// @Produce      json
// @Success      200  {object}  StatusResponse
// @Failure      500  {object}  StatusResponse
// @Router       /api/status [get]
func Ping(c *gin.Context, db *gorm.DB) {
    sqlDB, err := db.DB()
    if err != nil || sqlDB.Ping() != nil {
        c.JSON(http.StatusInternalServerError, StatusResponse{
            Status:   "unhealthy",
            Database: "disconnected",
        })
        return
    }

    c.JSON(http.StatusOK, StatusResponse{
        Status: "healthy",
        Database: "connected",
    })
}