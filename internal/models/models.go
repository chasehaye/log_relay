package models

import (
    "time"
    "gorm.io/gorm"
)

type Message struct {
    gorm.Model    
    SenderID      uint   `gorm:"index;not null"` 
    ReceiverID    uint   `gorm:"index;not null"` 
    Header        string `gorm:"type:varchar(255)"`
    Body          string `gorm:"type:text;not null"`
    MessageType   string `gorm:"type:varchar(30);not null"`
    MessageStatus string `gorm:"type:varchar(30)"`
    Importance    string `gorm:"type:varchar(30)"`

    // Relationships
    Sender   Sender   `gorm:"foreignKey:SenderID"`
    Receiver Receiver `gorm:"foreignKey:ReceiverID"`
}

type Sender struct {
    gorm.Model    
    Name         string    `gorm:"type:varchar(255)"`
    OriginEmail  string    `gorm:"uniqueIndex;type:varchar(255);not null"`
    OriginPhone  string    `gorm:"uniqueIndex;type:varchar(255);not null"`
    Messages     []Message `gorm:"foreignKey:SenderID"` // Outbox
}

type Receiver struct {
    gorm.Model    
    Name          string    `gorm:"type:varchar(255)"`
    Password      string    `gorm:"not null" json:"-"`
    Token         string    `gorm:"type:text"`
    ReceiverEmail string    `gorm:"uniqueIndex;type:varchar(255);not null"`
	IsAdmin       bool      `gorm:"default:false"`
    Messages      []Message `gorm:"foreignKey:ReceiverID"` // Inbox
}

type PasswordReset struct {
    gorm.Model
    UserID    uint      `gorm:"index"`
    Token     string    `gorm:"uniqueIndex"`
    ExpiresAt time.Time `gorm:"index"`
    Used      bool      `gorm:"default:false"`
}