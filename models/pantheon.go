package models

import (
	"context"
	"fmt"

	"github.com/Vintral/pocket-realm/utils"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

var pantheons []*Pantheon
var pantheonsByGuid = make(map[uuid.UUID]*Pantheon)
var pantheonsById = make(map[uint]*Pantheon)

type Pantheon struct {
	BaseModel

	GUID      uuid.UUID   `gorm:"uniqueIndex,size:36" json:"guid"`
	Category  string      `json:"category"`
	Devotions []*Devotion `gorm:"-" json:"devotions"`
}

func (pantheon *Pantheon) BeforeCreate(tx *gorm.DB) (err error) {
	pantheon.GUID = uuid.New()
	return
}

func (pantheon *Pantheon) AfterFind(tx *gorm.DB) (err error) {
	log.Trace().Msg("Round: AfterFind")

	ctx, sp := Tracer.Start(tx.Statement.Context, "Pantheon.AfterFind")
	defer sp.End()

	if err := db.WithContext(ctx).Table("devotions").Where("pantheon = ?", pantheon.ID).Find(&pantheon.Devotions).Error; err != nil {
		log.Error().Err(err).Msg("Error retrieving devotions")
		return err
	}

	return
}

func (pantheon *Pantheon) Dump() {
	val := ""
	for _, devotion := range pantheon.Devotions {
		val += "\n" + devotion.ToString()
	}

	fmt.Println(val)

	log.Info().Msg(`
==========PANTHEON===========
ID: ` + fmt.Sprint(pantheon.ID) + `
Category: ` + pantheon.Category + `
Devotions: ` + val + `
=============================
	`)
}

func loadPantheons(ctx context.Context) {
	ctx, span := utils.StartSpan(ctx, "models.loadPantheons")
	defer span.End()

	if err := db.WithContext(ctx).Find(&pantheons).Error; err != nil {
		log.Error().Err(err).Msg("Error retrieving pantheons")
	} else {
		for _, pantheon := range pantheons {
			pantheonsByGuid[pantheon.GUID] = pantheon
			pantheonsById[pantheon.ID] = pantheon
		}
	}
}

func GetPantheonByGuid(ctx context.Context, guid uuid.UUID) *Pantheon {
	ctx, span := utils.StartSpan(ctx, "models.GetPantheonByGuid")
	defer span.End()

	log.Info().Str("pantheon", guid.String()).Msg("models.GetPantheonByGuid")

	if len(pantheons) == 0 {
		loadPantheons(ctx)
	}

	if pantheon, ok := pantheonsByGuid[guid]; ok {
		return pantheon
	}

	return nil
}

func GetPantheonById(ctx context.Context, id uint) *Pantheon {
	ctx, span := utils.StartSpan(ctx, "models.GetPantheonById")
	defer span.End()

	log.Info().Uint("pantheon", id).Msg("models.GetPantheonById")

	if len(pantheons) == 0 {
		loadPantheons(ctx)
	}

	if pantheon, ok := pantheonsById[id]; ok {
		return pantheon
	}

	return nil
}

func GetPantheons(ctx context.Context, ret chan []*Pantheon) {
	ctx, span := utils.StartSpan(ctx, "models.GetPantheons")
	defer span.End()

	log.Info().Msg("GetPantheons")

	if len(pantheons) == 0 {
		loadPantheons(ctx)
	}

	ret <- pantheons
}
