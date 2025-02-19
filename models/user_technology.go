package models

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type UserTechnology struct {
	BaseModel

	UserID       uint      `json:"-"`
	RoundID      uint      `json:"-"`
	TechnologyID uint      `json:"-"`
	Technology   uuid.UUID `gorm:"-" json:"technology"`
	Level        uint      `json:"level"`
}

func (tech *UserTechnology) Dump() {
	log.Warn().Msg("=================================")
	log.Warn().Msg("UserID: " + fmt.Sprint(tech.UserID))
	log.Warn().Msg("RoundID: " + fmt.Sprint(tech.RoundID))
	log.Warn().Msg("TechnologyID: " + fmt.Sprint(tech.TechnologyID))
	log.Warn().Msg("Level: " + fmt.Sprint(tech.Level))
	log.Warn().Msg("=================================")
}
