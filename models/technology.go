package models

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gorm"
)

type Technology struct {
	BaseModel

	GUID        uuid.UUID          `gorm:"uniqueIndex,size:36" json:"guid"`
	Name        string             `json:"name"`
	Buff        uint               `json:"-"`
	Description string             `json:"description"`
	Available   bool               `gorm:"->;-:migration" json:"-"`
	Levels      []*TechnologyLevel `gorm:"-" json:"-"`
	Level       uint               `gorm:"-" json:"level"`
	Cost        uint               `gorm:"-" json:"cost"`
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

func (technology *Technology) LoadForUser(baseContext context.Context, wg *sync.WaitGroup, user *User, tech *Technology) {
	ctx, sp := Tracer.Start(baseContext, "technology.LoadForUser")
	defer sp.End()

	if wg != nil {
		defer wg.Done()
	}

	tech.ID = technology.ID
	tech.GUID = technology.GUID
	tech.Name = technology.Name
	tech.Buff = technology.Buff
	tech.Description = technology.Description

	log.Trace().Any("guid", tech.GUID).Msg("LoadForUser")

	sp.SetAttributes(
		attribute.Int("user", int(user.ID)),
		attribute.Int("round", user.RoundID),
		attribute.Int("technology", int(technology.ID)),
	)

	db.WithContext(ctx).Table("user_technologies").Select("level").Where("user_id = ? AND round_id = ? AND technology_id = ?", user.ID, user.RoundID, technology.ID).Scan(&tech.Level)

	for i := 0; i < len(technology.Levels); i++ {
		if technology.Levels[i].Level == tech.Level+1 {
			tech.Cost = technology.Levels[i].Cost
			return
		}
	}
}

func (technology *Technology) Dump() {
	levels := ""
	for _, level := range technology.Levels {
		levels += fmt.Sprintf("%d -- %d\n", level.Level, level.Cost)
	}

	log.Trace().Msg(`
============TECHNOLOGY===========")
ID: ` + fmt.Sprint(technology.ID) + `
Name: ` + technology.Name + `
Description: ` + technology.Description + `
Buff: ` + fmt.Sprint(technology.Buff) + `
Available: ` + fmt.Sprint(technology.Available) + `
Player Level: ` + fmt.Sprint(technology.Level) + `
--------------LEVELS--------------
` + levels + `
==================================
	`)
}

func GetTechnologyIdForGuid(baseContext context.Context, tech uuid.UUID) uint {
	ctx, sp := Tracer.Start(baseContext, "technology.GetTechnologyIdForGuid")
	defer sp.End()

	log.Info().Any("guid", tech).Msg("LoadTechnologyByGUID")

	var technology *uint
	if err := db.WithContext(ctx).Table("round_technologies").Where("guid = ?", tech).Select("technology_id").Scan(&technology).Error; err != nil || technology == nil {
		log.Error().AnErr("err", err).Str("technology", tech.String()).Msg("Technology not found")
		return 0
	}

	return *technology
}
