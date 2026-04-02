// phone mailing in the future
package models

import (
    "time"
    "gorm.io/gorm"
)

type ListType string

const (
	ListTypeMailing ListType = "MAILING"
	ListTypeInquiry ListType = "INQUIRY"
    ListTypeSupport ListType = "SUPPORT"
    ListTypeCatchAll  ListType = "CATCH_ALL"
)

type User struct {
    gorm.Model    
    Name      string    `gorm:"type:varchar(255)"`
    Password  string    `gorm:"not null" json:"-"`
    Token     string    `gorm:"type:text"`
    Email     string    `gorm:"uniqueIndex;type:varchar(255);not null"`
	IsAdmin   bool      `gorm:"default:false"`
    Lists    []List    `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
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
    ID        uint           `gorm:"primarykey" json:"id"`
    CreatedAt time.Time      `json:"created_at,omitempty"` 
    UpdatedAt time.Time      `json:"updated_at,omitempty"` 
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

    PublicID string `gorm:"uniqueIndex;type:varchar(50);not null"`
    PublicFacingName         string `gorm:"type:varchar(255)" json:"public_facing_name"`
    Name      string   `gorm:"index:idx_user_list_name,unique;type:varchar(255);not null" json:"name"`
    ListType  ListType `gorm:"type:varchar(30);not null" json:"list_type"`
    UserID    uint     `gorm:"index:idx_user_list_name,unique;not null" json:"-"`
    
    Messages []Message `gorm:"foreignKey:ListID" json:"messages,omitempty"` // outbound
    Subscribers  []Contact `gorm:"many2many:subscriber_list;" json:"subscribers,omitempty"`
}
// should store percentage outbound success rate
// should also have contacts sent to for outbound
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
    gorm.Model `json:"-"`
    UserID       uint   `gorm:"uniqueIndex:idx_user_email;not null" json:"user_id"`
    Email        string `gorm:"uniqueIndex:idx_user_email;type:varchar(255);not null" json:"email"`
 
    Name         string `gorm:"type:varchar(255)" json:"name"`
    // unsubscribe to mailing list
    UnSubToken string `gorm:"index" json:"-"`
    // subscribe to mailing list
    Verified          bool   `gorm:"default:false" json:"verified"`
    VerificationToken string `gorm:"index" json:"-"`
    TokenExpiresAt    time.Time `json:"-"`

    Unsubscribed      bool      `gorm:"default:false" json:"unsubscribed"`
    // Messages     []Message `gorm:"foreignKey:ContactID"` // 0 for message to user type
    SubscribedTo        []List    `gorm:"many2many:subscriber_list;" json:"subscribed_to,omitempty"` // 1 for mailing list type
}