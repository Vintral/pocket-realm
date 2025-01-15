package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoundResource struct {
	BaseModel

	GUID       uuid.UUID `gorm:"uniqueIndex,size:36" json:"guid"`
	RoundID    uint      `json:"-"`
	ResourceID uint      `json:"resource_id"`
	CanGather  bool      `json:"can_gather"`
	CanMarket  bool      `json:"can_market"`
	StartWith  uint      `json:"start_with"`
}

func (resource *RoundResource) BeforeCreate(tx *gorm.DB) (err error) {
	resource.GUID = uuid.New()
	return
}
