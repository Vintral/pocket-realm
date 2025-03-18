package models

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

var buffsById = make(map[int]*Buff)

type Buff struct {
	BaseModel

	ID         uint          `gorm:"primaryKey" json:"order"`
	Name       string        `json:"name"`
	EffectList string        `json:"-"`
	Effects    []*Effect     `gorm:"-" json:"effects"`
	Item       uint          `jons:"item"`
	Duration   time.Duration `json:"duration"`
	MaxStacks  uint          `json:"max_stacks"`
}

func (buff *Buff) AfterFind(tx *gorm.DB) (err error) {
	ctx, sp := Tracer.Start(tx.Statement.Context, "Buff.AfterFind")
	defer sp.End()

	log.Info().Msg("Buff.AfterFind")

	effects := strings.Split(buff.EffectList, ",")
	for _, effectId := range effects {
		if id, err := strconv.ParseInt(effectId, 10, 0); err == nil {
			effect := LoadEffectById(ctx, int(id))
			buff.Effects = append(buff.Effects, effect)
		} else {
			log.Error().Err(err).Msg("Error parsing effectId")
		}
	}

	return
}

func (buff *Buff) Dump() {

	log.Trace().Msg(`
============BUFF=============
ID: ` + fmt.Sprint(buff.ID) + `
Name: ` + buff.Name + `
EffectList: ` + buff.EffectList + `
LoadedEffects: ` + fmt.Sprint(len(buff.Effects)) + `
Item: ` + fmt.Sprint(buff.Item) + `
Duration: ` + fmt.Sprint(buff.Duration) + `
MaxStacks: ` + fmt.Sprint(buff.MaxStacks) + `
============================`)
}

func LoadBuffById(baseContext context.Context, buffID int) (*Buff, error) {
	b := buffsById[buffID]
	if b != nil {
		return b, nil
	}

	ctx, span := Tracer.Start(baseContext, "models.LoadBuffById")
	defer span.End()

	log.Info().Int("buff_id", buffID).Msg("models.LoadByffById")

	var buff Buff
	if err := db.WithContext(ctx).Where("id = ?", buffID).Find(&buff).Error; err != nil {
		log.Error().AnErr("err", err).Msg("Error loading buff")
		return nil, err
	}

	buffsById[buffID] = &buff
	return &buff, nil
}
