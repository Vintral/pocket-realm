package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Devotion struct {
	BaseModel

	GUID            uuid.UUID `gorm:"uniqueIndex,size:36" json:"-"`
	Pantheon        uint      `json:"-"`
	Level           uint      `json:"level"`
	Upkeep          uint      `json:"upkeep"`
	Buff            uint      `json:"-"`
	BuffDescription string    `json:"buff"`
}

func (devotion *Devotion) BeforeCreate(tx *gorm.DB) (err error) {
	devotion.GUID = uuid.New()
	return
}
