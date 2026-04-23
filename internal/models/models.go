package models

import (
    "time"
    "gorm.io/gorm"
)

type ListType string

const (
	ListTypeMailing   ListType = "MAILING"
	ListTypeInquiry   ListType = "INQUIRY"
    ListTypeSupport   ListType = "SUPPORT"
    ListTypeCatchAll  ListType = "CATCH_ALL"
)

type MessageType string

const (
    MessageTypeOutbound MessageType = "OUTBOUND"
    MessageTypeInbound  MessageType = "INBOUND"
)

type User struct {
    ID        uint      `gorm:"primarykey"`
    CreatedAt time.Time
    UpdatedAt time.Time  

    Name      string    `gorm:"type:varchar(255)"`
    Password  string    `gorm:"not null"`
    Token     string    `gorm:"type:text"`
    Email     string    `gorm:"uniqueIndex;type:varchar(255);not null"`
	IsAdmin   bool      `gorm:"default:false"`
    Lists     []List    `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type PasswordReset struct {
    gorm.Model
    UserID    uint      `gorm:"index"`
    Token     string    `gorm:"uniqueIndex"`
    ExpiresAt time.Time `gorm:"index"`
    Used      bool      `gorm:"default:false"`
    User      User      `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type List struct {
    ID                uint       `gorm:"primarykey"`
    CreatedAt         time.Time
    UpdatedAt         time.Time
    PublicID          string     `gorm:"uniqueIndex;type:varchar(50);not null"`
    PublicFacingName  string     `gorm:"type:varchar(255)"`
    ListType          ListType   `gorm:"type:varchar(30);not null"`
    
    Name              string     `gorm:"index:idx_user_list_name,unique;type:varchar(255);not null"`
    UserID            uint       `gorm:"index:idx_user_list_name,unique;not null"`
    SubscriberCount   int        `gorm:"-"`

    Messages          []Message  `gorm:"foreignKey:ListID;constraint:OnDelete:CASCADE;"`
    Subscribers       []Contact  `gorm:"many2many:subscriber_list;constraint:OnDelete:CASCADE;"`
}

type Message struct {
    gorm.Model
    Header        string `gorm:"type:varchar(255)"`
    Body          string `gorm:"type:text;not null"`
    ListID        uint   `gorm:"index"` 
    Type          MessageType  `gorm:"type:varchar(20);not null"`

    
    Importance    string `gorm:"type:varchar(30)"` // for inbound
    ContactID     uint    `gorm:"index"` // contact association, for inbound
}

type Contact struct {
    gorm.Model
    UserID            uint   `gorm:"uniqueIndex:idx_user_email;not null"`
    Email             string `gorm:"uniqueIndex:idx_user_email;type:varchar(255);not null"`
    Name              string `gorm:"type:varchar(255)"`
    
    // subscribe to mailing list
    Verified          bool   `gorm:"default:false"`
    VerificationToken string `gorm:"index"`
    TokenExpiresAt    time.Time
    // unsubscribe to mailing list
    UnSubToken        string `gorm:"index"`





    // Messages     []Message `gorm:"foreignKey:ContactID"` // for messages to user type, not for mailing list type
    SubscribedTo        []List    `gorm:"many2many:subscriber_list;"`
}