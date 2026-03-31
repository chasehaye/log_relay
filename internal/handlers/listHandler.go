package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"


	"log_relay/internal/models"

)

func CreateList(c *gin.Context, db *gorm.DB) {
	var input struct {
		Name  string `json:"name" binding:"required"`
		ListType string `json:"listtype" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

	uidValue, exists := c.Get("userID")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Not a valid user session"})
        return
    }
    userID := uidValue.(uint)

	var existing models.List
    result := db.Where("name = ? AND user_id = ?", input.Name, userID).First(&existing)
    if result.Error == nil {
        c.JSON(http.StatusConflict, gin.H{
            "error": "You already have a list named '" + input.Name + "'",
        })
        return
    }

	newList := models.List{
        Name:     input.Name,
        ListType: models.ListType(input.ListType),
        UserID:   userID,
    }
	if err := db.Create(&newList).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	c.JSON(http.StatusCreated, newList)
}