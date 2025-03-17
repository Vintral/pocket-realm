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
	log.Warn().Msg(`
=================================
UserID: ` + fmt.Sprint(tech.UserID) + `
RoundID: ` + fmt.Sprint(tech.RoundID) + `
TechnologyID: ` + fmt.Sprint(tech.TechnologyID) + `
Level: ` + fmt.Sprint(tech.Level) + `
=================================`)
}
