package models

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type Technology struct {
	BaseModel

	GUID      uuid.UUID `gorm:"uniqueIndex,size:36"`
	Name      string
	Buff      uint
	Available bool               `gorm:"->;-:migration" json:"available"`
	Levels    []*TechnologyLevel `gorm:"-"`
}

func (technology *Technology) BeforeCreate(tx *gorm.DB) (err error) {
	technology.GUID = uuid.New()
	return
}

func (technology *Technology) AfterFind(tx *gorm.DB) (err error) {
	tx.Raw(`
		SELECT
			level, cost
		FROM
			technology_levels
		WHERE technology = ` + fmt.Sprint(technology.ID) + `
		ORDER BY level ASC		
	`).Scan(&technology.Levels)

	return
}

func (technology *Technology) Dump() {
	log.Trace().Msg("==================================")
	log.Trace().Msg("Name: " + technology.Name)
	log.Trace().Msg("Buff: " + fmt.Sprint(technology.Buff))
	log.Trace().Msg("Available: " + fmt.Sprint(technology.Available))
	log.Trace().Msg("--------------LEVELS--------------")
	for _, level := range technology.Levels {
		log.Trace().Msg(fmt.Sprint(level.Level) + " -- " + fmt.Sprint(level.Cost))
	}
	log.Trace().Msg("==================================")
}
