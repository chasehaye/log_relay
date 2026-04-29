package services

import(
    "os"
    "strings"
    "fmt"

    "gorm.io/gorm"
	"log_relay/internal/models"
)

var (
    frontendURL string
)

func init() {
    frontendURL = strings.TrimSuffix(os.Getenv("FRONTEND_URL"), "/")
}

func SendMailingList(messageID uint, db *gorm.DB) {
    var message models.Message
    if err := db.First(&message, messageID).Error; err != nil {
        return
    }

    var list models.List
    if err := db.Preload("Subscribers").
        First(&list, message.ListID).Error; err != nil {
        return
    }

    for _, c := range list.Subscribers {
        unsubLink := fmt.Sprintf(
            "%s/unsubscribe?list=%s&token=%s",
            frontendURL,
            list.PublicID,
            c.UnSubToken,
        )
        SendMailListItem(c.Email, message.Header, message.Body, unsubLink)
    }
}