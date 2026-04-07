package list

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
    "math"

	"log_relay/internal/models"
	"log_relay/internal/services"
)

func CreateList(c *gin.Context, db *gorm.DB) {
	var input struct {
		Name  string `json:"name" binding:"required"`
		ListType string `json:"list_type" binding:"required"`
        PublicFacingName string `json:"public_facing_name" binding:"required"`
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

    publicID, err := services.GenerateToken()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate list identifier"})
        return
    }

	var existing models.List
    result := db.Where("name = ? AND user_id = ?", input.Name, userID).Limit(1).Find(&existing)
    if result.RowsAffected > 0 {
        c.JSON(http.StatusConflict, gin.H{"error": "List name already exists"})
        return
    }

	newList := models.List{
        Name:     input.Name,
        ListType: models.ListType(input.ListType),
        UserID:   userID,
        PublicID: publicID,
        PublicFacingName: input.PublicFacingName,
    }
	if err := db.Create(&newList).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	c.JSON(http.StatusCreated, newList)
}

func DeleteList(c *gin.Context, db *gorm.DB) {
    listID := c.Param("id")
    uidValue, exists := c.Get("userID")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Not a valid user session"})
        return
    }
    userID := uidValue.(uint)

    result := db.Where("id = ? AND user_id = ?", listID, userID).Delete(&models.List{})
    if result.Error != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }
    if result.RowsAffected == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "List not found or you don't have permission"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "List deleted successfully"})
}

func IndexList(c *gin.Context, db *gorm.DB) {
    var input struct {
        CountPerPage int `form:"count_per_page" binding:"required,min=1"`
        Page       int `form:"page" binding:"required,min=1"`
    }
    if err := c.ShouldBindQuery(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameters 'page' and 'cnt_per_page' are required"})
        return
    }

    uidValue, exists := c.Get("userID")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Not a valid user session"})
        return
    }
    userID := uidValue.(uint)

    if input.CountPerPage < 1 || input.CountPerPage > 50 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid count per page"})
        return
    }

    var totalCount int64
    db.Model(&models.List{}).Where("user_id = ?", userID).Count(&totalCount)

    totalPages := 0
    if totalCount > 0 {
        totalPages = int(math.Ceil(float64(totalCount) / float64(input.CountPerPage)))
    }

    requestedPage := input.Page
    if requestedPage > totalPages && totalPages > 0 {
        requestedPage = totalPages
    }

    var lists []models.List
    offset := (requestedPage - 1) * input.CountPerPage

    countSubQuery := "(SELECT COUNT(*) FROM subscriber_list WHERE subscriber_list.list_id = lists.id) AS subscriber_count"
    
    db.Where("user_id = ?", userID).
        Select("id", "name", "list_type", "updated_at", countSubQuery).
        Limit(input.CountPerPage).
        Offset(offset).
        Order("updated_at DESC").
        Find(&lists)
        

    c.JSON(http.StatusOK, gin.H{
        "lists":        lists,
        "total_count":  totalCount,
        "total_pages":  totalPages,
        "current_page": requestedPage,
    })
}

func GetListDetail(c *gin.Context, db *gorm.DB){
    listID := c.Param("id")

    uid, exists := c.Get("userID")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Not a valid user session"})
        return
    }

    userID := uid.(uint)

    var list models.List
    result := db.Preload("Messages").
        Preload("Subscribers").
        Where("id = ? AND user_id = ?", listID, userID).
        First(&list)

    if result.Error != nil {
        if result.Error == gorm.ErrRecordNotFound {
            c.JSON(http.StatusNotFound, gin.H{"error": "List not found"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }
    c.JSON(http.StatusOK, list)
}