package models

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Message struct {
	BaseModel

	GUID         uuid.UUID `gorm:"uniqueIndex,size:36" json:"guid"`
	Conversation uint      `json:"-"`
	UserID       uint      `json:"-"`
	User         string    `gorm:"->:" json:"username"`
	Avatar       string    `gorm:"-" json:"avatar"`
	Text         string    `json:"message"`
	CreatedAt    time.Time `json:"time"`
}

func (message *Message) BeforeCreate(tx *gorm.DB) (err error) {
	message.GUID = uuid.New()
	return
}

func (message *Message) Create(userID uint, text string) error {
	result := db.Create(&Shout{UserID: userID, Shout: text})
	return result.Error
}

func (message *Message) Save(ctx context.Context) error {
	result := db.WithContext(ctx).Save(&message)
	return result.Error
}
