package models

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type Devotion struct {
	BaseModel

	GUID       uuid.UUID `gorm:"uniqueIndex,size:36" json:"-"`
	Pantheon   uint      `json:"-"`
	Level      uint      `json:"level"`
	Upkeep     uint      `json:"upkeep"`
	BuffId     uint      `gorm:"column:buff" json:"-"`
	Buff       Buff      `gorm:"-" json:"-"`
	BuffName   string    `gorm:"-" json:"buff"`
	BuffEffect string    `gorm:"-" json:"effect"`
}

func (devotion *Devotion) BeforeCreate(tx *gorm.DB) (err error) {
	devotion.GUID = uuid.New()
	return
}

func (devotion *Devotion) AfterFind(tx *gorm.DB) (err error) {
	ctx, sp := Tracer.Start(tx.Statement.Context, "Devotion.AfterFind")
	defer sp.End()

	log.Info().Msg("Devotion.AfterFind")

	// Load Buff
	if devotion.BuffId != 0 {
		if buff, err := LoadBuffById(ctx, int(devotion.BuffId)); err != nil {
			log.Error().Err(err).Msg("Error loading buff")
		} else {
			devotion.Buff = *buff
			devotion.BuffName = buff.Name
			devotion.BuffEffect = buff.Effect()
		}
	}

	return
}

func (devotion *Devotion) Dump() {
	log.Info().Msg(`
=============================
ID: ` + fmt.Sprint(devotion.ID) + `
GUID: ` + devotion.GUID.String() + `
Level: ` + fmt.Sprint(devotion.Level) + `
Upkeep: ` + fmt.Sprint(devotion.Upkeep) + `
BuffId: ` + fmt.Sprint(devotion.BuffId) + `
Buff: ` + devotion.Buff.String() + `
=============================
	`)
}

func (devotion *Devotion) ToString() string {
	val := fmt.Sprintf("%d | %d | %s", devotion.Level, devotion.Upkeep, devotion.BuffId)

	log.Info().Msg(val)

	return val
}
