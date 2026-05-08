package list

import (
	"net/http"
    "errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
    "math"

	"log_relay/internal/models"
	"log_relay/internal/crypt"
    "log_relay/internal/dtos"
)


// CreateList godoc
// @Summary      Create New List
// @Description  Create a new list and returns the created list and a success message
// @Tags         lists
// @Accept       json
// @Produce      json
// @Param        user  body      ListInput true "List Creation Data"
// @Success      201   {object}  ListResponse
// @Failure      400   {object}  dtos.ValidationErrorResponse
// @Failure      401   {object}  dtos.UnauthorizedResponse
// @Failure      409   {object}  dtos.AlreadyExistsResponse
// @Failure      500   {object}  dtos.ServerErrorResponse
// @Router       /api/list/create [post]
func CreateList(c *gin.Context, db *gorm.DB) {
	var input ListInput
	if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, dtos.ValidationErrorResponse{
            Error: "Failed: Incorrect input format follow the below specificatons",
            Details: map[string]string{
                "name": "name is required",
                "list_type": "list_type is required must be of type: 'MAILING', 'INQUIRY', 'BUG', or 'CATCH_ALL'",
                "public_facing_name": "pubic_facing_name is a required field",
            },
        })
        return
    }

	uidValue, exists := c.Get("userID")
    if !exists {
        c.JSON(http.StatusUnauthorized, dtos.UnauthorizedResponse{
            Error: "Invalid session",
        })
        return
    }
    userID := uidValue.(uint)

    publicID, err := crypt.GenerateToken()
    if err != nil {
        c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{
            Error: "Failed to generate list identifier",
        })
        return
    }

	var existing models.List
    result := db.Where("name = ? AND user_id = ?", input.Name, userID).Limit(1).Find(&existing)
    if result.RowsAffected > 0 {
        c.JSON(http.StatusConflict, dtos.AlreadyExistsResponse{
            Error: "List name alreadt exists select a different name",
        })
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
		c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{
            Error: "Database error",
        })
		return
	}
	c.JSON(http.StatusCreated, ListResponse{
        ID:               newList.ID,
        Name:             newList.Name,
        ListType:         string(newList.ListType),
        PublicFacingName: newList.PublicFacingName,
        PublicID:         newList.PublicID,
        UserID:           newList.UserID,
        CreatedAt:        newList.CreatedAt, 
    })
}

// DeleteList godoc
// @Summary      Delete a list
// @Description  Deletes a list and all associated messages
// @Tags         lists
// @Accept       json
// @Produce      json
// @Param        id path string true "List ID"
// @Success      200 {object} SuccessMessageResponse
// @Failure      401 {object} dtos.UnauthorizedResponse
// @Failure      404 {object} dtos.NotFoundErrorResponse
// @Failure      500 {object} dtos.ServerErrorResponse
// @Router       /api/list/delete/{id} [delete]
func DeleteList(c *gin.Context, db *gorm.DB) {
    listID := c.Param("id")
    uidValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dtos.UnauthorizedResponse{
			Error: "Invalid session",
		})
		return
	}
    userID := uidValue.(uint)

    result := db.Where("id = ? AND user_id = ?", listID, userID).Delete(&models.List{})
    if result.Error != nil {
        c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{
			Error: "Database error",
		})
        return
    }
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, dtos.NotFoundErrorResponse{
			Error: "List not found or you don't have permission",
		})
		return
	}

    c.JSON(http.StatusOK, SuccessMessageResponse{
		Message: "List deleted successfully",
	})
}

// IndexList godoc
// @Summary      Get paginated lists
// @Description  Returns a paginated list of user lists with metadata
// @Tags         lists
// @Accept       json
// @Produce      json
// @Param        count_per_page  query     int  true  "Number of items per page (1-50)"
// @Param        page            query     int  true  "Page number (starts from 1)"
// @Success      200 {object} ListIndexResponse
// @Failure      400 {object} dtos.ValidationErrorResponse
// @Failure      401 {object} dtos.UnauthorizedResponse
// @Failure      500 {object} dtos.ServerErrorResponse
// @Router       /api/lists [get]
func IndexList(c *gin.Context, db *gorm.DB) {
    var input IndexQuery
    if err := c.ShouldBindQuery(&input); err != nil {
        c.JSON(http.StatusBadRequest, dtos.ValidationErrorResponse{
            Error: "Invalid query parameters",
            Details: map[string]string{
                "page":           "page is required and must be >= 1",
                "count_per_page": "count_per_page is required and must be between 1 and 50",
            },
        })
        return
    }

    uidValue, exists := c.Get("userID")
    if !exists {
        c.JSON(http.StatusUnauthorized, dtos.UnauthorizedResponse{
			Error: "Invalid session",
		})
        return
    }
    userID := uidValue.(uint)

    var totalCount int64
    db.Model(&models.List{}).Where("user_id = ? AND list_type = ?", userID, models.ListTypeMailing).Count(&totalCount)

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
    
    result := db.Where("user_id = ? AND list_type = ?", userID, models.ListTypeMailing).
		Select("id", "name", "list_type", "public_facing_name", "public_id", "user_id", "created_at", "updated_at", countSubQuery).
		Limit(input.CountPerPage).
		Offset(offset).
		Order("updated_at DESC").
		Find(&lists)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{
			Error: "Database error",
		})
		return
	}
    converted := make([]ListResponse, len(lists))
	for i, l := range lists {
		converted[i] = ListResponse{
			ID:               l.ID,
			Name:             l.Name,
			ListType:         string(l.ListType),
			PublicFacingName: l.PublicFacingName,
			PublicID:         l.PublicID,
			UserID:           l.UserID,
            CreatedAt:        l.CreatedAt, 
		}
	}

	c.JSON(http.StatusOK, ListIndexResponse{
		Lists:       converted,
		TotalCount:  totalCount,
		TotalPages:  totalPages,
		CurrentPage: requestedPage,
	})
}

// GetListDetail godoc
// @Summary      Get list details
// @Description  Returns a single list with messages and subscribers
// @Tags         lists
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "List ID"
// @Success      200  {object}  ListDetailResponse
// @Failure      401  {object}  dtos.UnauthorizedResponse
// @Failure      404  {object}  dtos.NotFoundErrorResponse
// @Failure      500  {object}  dtos.ServerErrorResponse
// @Router       /api/list/detail/{id} [get]
func GetListDetail(c *gin.Context, db *gorm.DB){
    listID := c.Param("id")

    uid, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dtos.UnauthorizedResponse{
			Error: "Invalid session",
		})
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
            c.JSON(http.StatusNotFound, dtos.NotFoundErrorResponse{
				Error: "List not found",
			})
            return
        }
        c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{
			Error: "Database error",
		})
        return
    }
    response := ListDetailResponse{
		ID:               list.ID,
		Name:             list.Name,
		ListType:         string(list.ListType),
		PublicFacingName: list.PublicFacingName,
		PublicID:         list.PublicID,
		UserID:           list.UserID,
        CreatedAt:        list.CreatedAt,
        UpdatedAt: list.UpdatedAt,
	}

    response.Messages = make([]MessageResponse, 0, len(list.Messages))
    response.Subscribers = make([]ContactResponse, 0, len(list.Subscribers))

	for _, m := range list.Messages {
        response.Messages = append(response.Messages, MessageResponse{
            ID:     m.ID,
            Header: m.Header,
            Body:   m.Body,
        })
    }

    for _, s := range list.Subscribers {
        response.Subscribers = append(response.Subscribers, ContactResponse{
            Email: s.Email,
        })
    }


	c.JSON(http.StatusOK, response)
}

// GetListPublicName godoc
// @Summary      Get public mailing list name
// @Description  Returns the public-facing name of a mailing list by its public ID
// @Tags         lists
// @Accept       json
// @Produce      json
// @Param        list_id  path      string  true  "Public List ID"
// @Success      200  {object}  ListPublicNameResponse
// @Failure      404  {object}  dtos.NotFoundErrorResponse
// @Failure      500  {object}  dtos.ServerErrorResponse
// @Router       /api/lists/{list_id} [get]
func GetListPublicName(c *gin.Context, db *gorm.DB){
    listID := c.Param("list_id")

    var list models.List
    err := db.Select("public_facing_name").Where("public_id = ?", listID).First(&list).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            c.JSON(http.StatusNotFound, dtos.NotFoundErrorResponse{
                Error: "List name not found",
            })
            return
        }

        c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{
            Error: "Database error",
        })
        return
    }

    c.JSON(http.StatusOK, ListPublicNameResponse{
        Name: list.PublicFacingName,
    })
}

// IndexProjectBugReport godoc
// @Summary      Get paginated bug report lists
// @Description  Returns a paginated list of user bug report lists with metadata
// @Tags         lists
// @Accept       json
// @Produce      json
// @Param        count_per_page  query     int  true  "Number of items per page (1-50)"
// @Param        page            query     int  true  "Page number (starts from 1)"
// @Success      200 {object} ListIndexResponse
// @Failure      400 {object} dtos.ValidationErrorResponse
// @Failure      401 {object} dtos.UnauthorizedResponse
// @Failure      500 {object} dtos.ServerErrorResponse
// @Router       /api/list/index/project/bug-report [get]
func IndexProjectBugReport(c *gin.Context, db *gorm.DB) {
    var input IndexQuery
    if err := c.ShouldBindQuery(&input); err != nil {
        c.JSON(http.StatusBadRequest, dtos.ValidationErrorResponse{
            Error: "Invalid query parameters",
            Details: map[string]string{
                "page":           "page is required and must be >= 1",
                "count_per_page": "count_per_page is required and must be between 1 and 50",
            },
        })
        return
    }

    uidValue, exists := c.Get("userID")
    if !exists {
        c.JSON(http.StatusUnauthorized, dtos.UnauthorizedResponse{
			Error: "Invalid session",
		})
        return
    }
    userID := uidValue.(uint)

    var totalCount int64
    db.Model(&models.List{}).Where("user_id = ? AND list_type = ?", userID, models.ListTypeBug).Count(&totalCount)

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

    result := db.Where("user_id = ? AND list_type = ?", userID, models.ListTypeBug).
		Select("id", "name", "list_type", "public_facing_name", "public_id", "user_id", "created_at", "updated_at").
		Limit(input.CountPerPage).
		Offset(offset).
		Order("updated_at DESC").
		Find(&lists)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{
			Error: "Database error",
		})
		return
	}
    converted := make([]ListResponse, len(lists))
	for i, l := range lists {
		converted[i] = ListResponse{
			ID:               l.ID,
			Name:             l.Name,
			ListType:         string(l.ListType),
			PublicFacingName: l.PublicFacingName,
			PublicID:         l.PublicID,
			UserID:           l.UserID,
            CreatedAt:        l.CreatedAt, 
		}
	}

	c.JSON(http.StatusOK, ListIndexResponse{
		Lists:       converted,
		TotalCount:  totalCount,
		TotalPages:  totalPages,
		CurrentPage: requestedPage,
	})
}

// IndexProjectInquiry godoc
// @Summary      Get paginated inquiry lists
// @Description  Returns a paginated list of user inquiry lists with metadata
// @Tags         lists
// @Accept       json
// @Produce      json
// @Param        count_per_page  query     int  true  "Number of items per page (1-50)"
// @Param        page            query     int  true  "Page number (starts from 1)"
// @Success      200 {object} ListIndexResponse
// @Failure      400 {object} dtos.ValidationErrorResponse
// @Failure      401 {object} dtos.UnauthorizedResponse
// @Failure      500 {object} dtos.ServerErrorResponse
// @Router       /api/list/index/project/inquiry [get]
func IndexProjectInquiry(c *gin.Context, db *gorm.DB) {
    var input IndexQuery
    if err := c.ShouldBindQuery(&input); err != nil {
        c.JSON(http.StatusBadRequest, dtos.ValidationErrorResponse{
            Error: "Invalid query parameters",
            Details: map[string]string{
                "page":           "page is required and must be >= 1",
                "count_per_page": "count_per_page is required and must be between 1 and 50",
            },
        })
        return
    }

    uidValue, exists := c.Get("userID")
    if !exists {
        c.JSON(http.StatusUnauthorized, dtos.UnauthorizedResponse{
            Error: "Invalid session",
        })
        return
    }
    userID := uidValue.(uint)

    var totalCount int64
    db.Model(&models.List{}).Where("user_id = ? AND list_type = ?", userID, models.ListTypeInquiry).Count(&totalCount)

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

    result := db.Where("user_id = ? AND list_type = ?", userID, models.ListTypeInquiry).
        Select("id", "name", "list_type", "public_facing_name", "public_id", "user_id", "created_at", "updated_at").
        Limit(input.CountPerPage).
        Offset(offset).
        Order("updated_at DESC").
        Find(&lists)

    if result.Error != nil {
        c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{
            Error: "Database error",
        })
        return
    }

    converted := make([]ListResponse, len(lists))
    for i, l := range lists {
        converted[i] = ListResponse{
            ID:               l.ID,
            Name:             l.Name,
            ListType:         string(l.ListType),
            PublicFacingName: l.PublicFacingName,
            PublicID:         l.PublicID,
            UserID:           l.UserID,
            CreatedAt:        l.CreatedAt,
        }
    }

    c.JSON(http.StatusOK, ListIndexResponse{
        Lists:       converted,
        TotalCount:  totalCount,
        TotalPages:  totalPages,
        CurrentPage: requestedPage,
    })
}