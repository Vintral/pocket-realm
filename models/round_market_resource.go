package models

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type RoundMarketResource struct {
	BaseModel

	GUID       uuid.UUID `gorm:"uniqueIndex,size:36" json:"resource"`
	RoundID    uint      `json:"-"`
	ResourceID uint      `json:"-"`
	Sold       uint      `json:"-"`
	Bought     uint      `json:"-"`
	Value      float32   `json:"value"`
}

func (resource *RoundMarketResource) BeforeCreate(tx *gorm.DB) (err error) {
	resource.GUID = uuid.New()
	return
}

func (resource *RoundMarketResource) Dump() {
	log.Trace().Msg("=====================================")
	log.Trace().Msg("GUID: " + resource.GUID.String())
	log.Trace().Msg("RoundID: " + fmt.Sprint(resource.RoundID))
	log.Trace().Msg("ResourceID: " + fmt.Sprint(resource.ResourceID))
	log.Trace().Msg("Sold: " + fmt.Sprint(resource.Sold))
	log.Trace().Msg("Bought: " + fmt.Sprint(resource.Bought))
	log.Trace().Msg("Value: " + fmt.Sprint(resource.Value))
	log.Trace().Msg("=====================================")
}
