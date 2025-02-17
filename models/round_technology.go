package models

import (
	"github.com/google/uuid"
)

type RoundTechnology struct {
	BaseModel

	GUID         uuid.UUID          `gorm:"uniqueIndex,size:36" json:"guid"`
	RoundID      uint               `json:"-"`
	TechnologyID uint               `json:"-"`
	Available    bool               `json:"-"`
	Levels       []*TechnologyLevel `gorm:"-"`
}
