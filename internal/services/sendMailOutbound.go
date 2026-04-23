package services

import(
    "gorm.io/gorm"
	"log_relay/internal/models"
)

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
        SendMail(c.Email, message.Header, message.Body)
    }
}