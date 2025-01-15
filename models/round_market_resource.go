package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoundMarketResource struct {
	BaseModel

	GUID       uuid.UUID `gorm:"uniqueIndex,size:36" json:"-"`
	RoundID    uint      `json:"-"`
	ResourceID uuid.UUID `json:"resource"`
	Sold       uint      `json:"-"`
	Bought     uint      `json:"-"`
	Value      float32   `json:"value"`
}

func (resource *RoundMarketResource) BeforeCreate(tx *gorm.DB) (err error) {
	resource.GUID = uuid.New()
	return
}
