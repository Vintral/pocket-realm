package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoundTechnology struct {
	BaseModel

	GUID         uuid.UUID          `gorm:"uniqueIndex,size:36" json:"-"`
	RoundID      uint               `json:"-"`
	TechnologyID uint               `json:"-"`
	Available    bool               `json:"-"`
	Levels       []*TechnologyLevel `gorm:"-"`
}

func (technology *RoundTechnology) BeforeCreate(tx *gorm.DB) (err error) {
	technology.GUID = uuid.New()
	return
}
