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
	tech.Description = technology.Description

	log.Warn().Str("description", tech.Description).Msg("LOAD TECHNOLOGY")

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
	log.Warn().Msg("==================================")
	log.Warn().Msg("ID: " + fmt.Sprint(technology.ID))
	log.Warn().Msg("Name: " + technology.Name)
	log.Warn().Msg("Description: " + technology.Description)
	log.Warn().Msg("Buff: " + fmt.Sprint(technology.Buff))
	log.Warn().Msg("Available: " + fmt.Sprint(technology.Available))
	log.Warn().Msg("Player Level: " + fmt.Sprint(technology.Level))
	log.Warn().Msg("--------------LEVELS--------------")
	for _, level := range technology.Levels {
		log.Warn().Msg(fmt.Sprint(level.Level) + " -- " + fmt.Sprint(level.Cost))
	}
	log.Warn().Msg("==================================")
}

func GetTechnologyIdForGuid(baseContext context.Context, tech uuid.UUID) uint {
	ctx, sp := Tracer.Start(baseContext, "technology.GetTechnologyIdForGuid")
	defer sp.End()

	log.Info().Any("guid", tech).Msg("LoadTechnologyByGUID")

	var technology *uint
	db.WithContext(ctx).Table("round_technologies").Where("guid = ?", tech).Select("technology_id").Scan(&technology)

	return *technology
}
