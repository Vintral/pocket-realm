package models

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
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

func (devotion *Devotion) Dump() {
	log.Info().Msg(`
=============================
ID: ` + fmt.Sprint(devotion.ID) + `
GUID: ` + devotion.GUID.String() + `
Level: ` + fmt.Sprint(devotion.Level) + `
Upkeep: ` + fmt.Sprint(devotion.Upkeep) + `
Buff: ` + fmt.Sprint(devotion.Buff) + `
BuffDescription: ` + devotion.BuffDescription + `
=============================
	`)
}

func (devotion *Devotion) ToString() string {
	val := fmt.Sprintf("%d | %d | %s", devotion.Level, devotion.Upkeep, devotion.BuffDescription)

	log.Info().Msg(val)

	return val
}
