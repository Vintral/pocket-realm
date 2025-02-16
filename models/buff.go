package models

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

var buffsById = make(map[int]*Buff)

type Buff struct {
	BaseModel

	ID       uint          `gorm:"primaryKey" json:"order"`
	Name     string        `json:"name"`
	Type     string        `json:"type"`
	Field    string        `json:"field"`
	Item     uint          `jons:"item"`
	Bonus    float64       `json:"bonus"`
	Percent  bool          `json:"percent"`
	Duration time.Duration `json:"duration"`
}

func (buff *Buff) applyTo(user *User) {
	log.Trace().Uint("buff", buff.ID).Uint("user", user.ID).Msg("applyTo")
}

func (buff *Buff) Dump() {
	log.Trace().Msg("=============================")
	log.Trace().Msg("ID:" + fmt.Sprint(buff.ID))
	log.Trace().Msg("Name:" + buff.Name)
	log.Trace().Msg("Type: " + buff.Type)
	log.Trace().Msg("Field: " + buff.Field)
	log.Trace().Msg("Item: " + fmt.Sprint(buff.Item))
	log.Trace().Msg("Bonus: " + fmt.Sprint(buff.Bonus))
	log.Trace().Msg("Percent: " + fmt.Sprint(buff.Percent))
	log.Trace().Msg("Duration: " + fmt.Sprint(buff.Duration))
	log.Trace().Msg("=============================")
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
