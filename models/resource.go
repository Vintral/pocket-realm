package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Resource struct {
	BaseModel

	ID        uint      `gorm:"primaryKey" json:"order"`
	GUID      uuid.UUID `gorm:"uniqueIndex,size:36" json:"guid"`
	Name      string    `json:"name"`
	CanGather bool      `gorm:"->;-:migration" json:"can_gather"`
	CanMarket bool      `gorm:"->;-:migration" json:"can_market"`
	StartWith uint      `gorm:"->;-:migration" json:"start_with"`
}

func (resource *Resource) BeforeCreate(tx *gorm.DB) (err error) {
	resource.GUID = uuid.New()
	return
}
