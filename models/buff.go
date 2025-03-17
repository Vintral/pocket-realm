package models

import (
	"context"
	"fmt"
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
	log.Trace().Msg("Buff.AfterFind")

	ctx, sp := Tracer.Start(tx.Statement.Context, "Buff.AfterFind")
	defer sp.End()

	effects := strings.Split(buff.EffectList, ",")
	var effect *Effect
	for _, effectId := range effects {
		tx.WithContext(ctx).Where("id = ?", effectId).Find(&effect)
		buff.Effects = append(buff.Effects, effect)
	}

	for _, e := range buff.Effects {
		e.Dump()
	}

	return
}

func (buff *Buff) Dump() {
	effects := ""

	for _, effect := range buff.Effects {
		effects += fmt.Sprintf("| %d ", effect.ID)
	}
	effects += "|"

	log.Trace().Msg(`
=============================
ID: ` + fmt.Sprint(buff.ID) + `
Name: ` + buff.Name + `
EffectList: ` + buff.EffectList + `
Effects: ` + effects + `
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

	ctx, span := Tracer.Start(baseContext, "add-buff")
	defer span.End()

	log.Info().Int("buff_id", buffID).Msg("Loading Buff")

	var buff Buff
	if err := db.WithContext(ctx).Where("id = ?", buffID).Find(&buff).Error; err != nil {
		log.Error().AnErr("err", err).Msg("Error loading buff")
		return nil, err
	}

	log.Debug().Any("buff", buff).Send()
	buff.Dump()

	fmt.Println(buff)

	buffsById[buffID] = &buff
	return &buff, nil
}
