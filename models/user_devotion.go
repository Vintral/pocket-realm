package models

import (
	"fmt"

	"github.com/rs/zerolog/log"
)

type UserDevotion struct {
	BaseModel

	UserID       uint   `json:"-"`
	RoundID      uint   `json:"-"`
	Pantheon     uint   `json:"-"`
	PantheonName string `gorm:"->;-:migration" json:"pantheon"`
	Level        uint   `json:"level"`
}

func (devotion *UserDevotion) Dump() {
	log.Trace().Msg(`
========USER_DEVOTION========
ID: ` + fmt.Sprint(devotion.ID) + `
UserID: ` + fmt.Sprint(devotion.UserID) + `
RoundID: ` + fmt.Sprint(devotion.RoundID) + `
Pantheon: ` + fmt.Sprint(devotion.Pantheon) + `
PantheonName: ` + devotion.PantheonName + `
Level: ` + fmt.Sprint(devotion.Level) + `
=============================
	`)
}
