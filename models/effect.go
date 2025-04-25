package models

import (
	"context"
	"fmt"

	"github.com/Vintral/pocket-realm/utils"
	"github.com/rs/zerolog/log"
)

var effectsById = make(map[int]*Effect)

type Effect struct {
	BaseModel

	Type    string `json:"type"`
	Field   string `json:"field"`
	Amount  int    `json:"amount"`
	Percent bool   `json:"percent"`
}

func (effect *Effect) Dump() {
	log.Warn().Msg(`
============EFFECT===========
ID: ` + fmt.Sprint(effect.ID) + `
Type: ` + effect.Type + `
Field: ` + effect.Field + `
Amount: ` + fmt.Sprint(effect.Amount) + `
Percent: ` + fmt.Sprint(effect.Percent) + `
=============================`)
}

func (effect *Effect) String() string {
	var ret = ""

	if effect.Amount > 0 {
		ret += "+"
	}
	ret += fmt.Sprint(effect.Amount)
	if effect.Percent {
		ret += "%"
	}

	ret += "," + effect.Field

	return ret
}

func LoadEffectById(ctx context.Context, effectId int) *Effect {
	ctx, span := utils.StartSpan(ctx, "models.LoadEffectById")
	defer span.End()

	log.Info().Int("effect", effectId).Msg("models.LoadEffectById")

	if effect, ok := effectsById[effectId]; ok {
		return effect
	} else {
		var effect *Effect
		if err := db.WithContext(ctx).Where("id = ?", effectId).Find(&effect).Error; err == nil {
			effectsById[effectId] = effect
			return effectsById[effectId]
		} else {
			log.Error().Err(err).Int("effect", effectId).Msg("Error loading effect")
		}
	}

	return nil
}
