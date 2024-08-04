package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Item struct {
	BaseModel

	ID          uint      `gorm:"primaryKey" json:"order"`
	GUID        uuid.UUID `gorm:"uniqueIndex,size:36" json:"guid"`
	Name        string    `json:"name"`
	Plural      string    `json:"plural"`
	Description string    `json:"description"`
}

func (item *Item) BeforeCreate(tx *gorm.DB) (err error) {
	item.GUID = uuid.New()
	return
}
