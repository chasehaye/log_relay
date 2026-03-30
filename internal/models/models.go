package models

import (
    "time"
    "gorm.io/gorm"
)

type User struct {
    gorm.Model    
    Name      string    `gorm:"type:varchar(255)"`
    Password  string    `gorm:"not null" json:"-"`
    Token     string    `gorm:"type:text"`
    Email     string    `gorm:"uniqueIndex;type:varchar(255);not null"`
	IsAdmin   bool      `gorm:"default:false"`
    Lists    []List `gorm:"foreignKey:UserID"`
    Contacts  []Contact `gorm:"foreignKey:UserID"` // members of business user
}

type PasswordReset struct {
    gorm.Model
    UserID    uint      `gorm:"index"`
    Token     string    `gorm:"uniqueIndex"`
    ExpiresAt time.Time `gorm:"index"`
    Used      bool      `gorm:"default:false"`
    User      User      `gorm:"foreignKey:UserID"`
}

type List struct {
    gorm.Model
    Name string        `gorm:"uniqueIndex;type:varchar(255);not null"`
    ListType   string `gorm:"type:varchar(30);not null"`
    UserID    uint   `gorm:"index;not null"`
    Messages []Message `gorm:"foreignKey:ListID"` // both
    Contacts  []Contact `gorm:"many2many:contact_lists;"` // 1 for mailing list (list of contacts)
}

type Message struct {
    gorm.Model
    Header        string `gorm:"type:varchar(255)"`
    Body          string `gorm:"type:text;not null"`
    Importance    string `gorm:"type:varchar(30)"`
    MessageStatus string `gorm:"type:varchar(30)"`
    Type    string  `gorm:"type:varchar(20)"` // 0 in or 1 out
    ContactID     uint    `gorm:"index"`  // 0
    ListID        uint   `gorm:"index"` // both
}

type Contact struct {
    gorm.Model
    UserID       uint      `gorm:"index"` // members of business user
    Name         string    `gorm:"type:varchar(255)"`
    OriginEmail  string    `gorm:"uniqueIndex;type:varchar(255);not null"`
    OriginPhone  string    `gorm:"uniqueIndex;type:varchar(255);not null"`
    Messages     []Message `gorm:"foreignKey:ContactID"` // 0 for message to user type
    Lists        []List    `gorm:"many2many:contact_lists;"` // 1 for mailing list type
}