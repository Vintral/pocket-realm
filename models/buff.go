package models

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Vintral/pocket-realm/utils"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/attribute"
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

func (buff *Buff) AddToUser(baseContext context.Context, user *User) bool {
	ctx, span := utils.StartSpan(baseContext, "Buff.AddToUser")
	defer span.End()

	log.Info().Int("user", int(user.ID)).Int("buff", int(buff.ID)).Msg("Buff.AddToUser")

	span.SetAttributes(
		attribute.Int("user", int(user.ID)),
		attribute.Int("buff", int(buff.ID)),
	)

	found := false
	var temp *UserBuff
	for i := 0; i < len(user.Buffs) && !found; i++ {
		if user.Buffs[i].BuffID == buff.ID {
			temp = user.Buffs[i]
			temp.Stacks = temp.Stacks + 1
			if temp.Stacks > buff.MaxStacks {
				temp.Stacks = buff.MaxStacks
			}
		}
	}

	if temp == nil {
		log.Warn().Msg("Creating new UserBuff: " + fmt.Sprint(getRound(user)))
		temp = &UserBuff{
			UserID:  user.ID,
			RoundID: uint(getRound(user)),
			BuffID:  buff.ID,
			Stacks:  1,
		}

		if buff.Duration != 0 {
			temp.Expires = time.Now()
		} else {
			log.Warn().Msg("Duration is 0")
			if round, err := LoadRoundById(ctx, getRound(user)); err == nil {
				fmt.Println(round)
				temp.Expires = round.Ends

				log.Warn().Msg("Set expires: " + fmt.Sprint(temp.Expires))
			} else {
				log.Error().AnErr("err", err).Msg("Error loading round")
			}
		}

		user.Buffs = append(user.Buffs, temp)
	}

	if buff.Duration != 0 {
		temp.Expires = time.Now().Add(buff.Duration)
	}

	if err := db.WithContext(ctx).Save(&user).Error; err != nil {
		log.Error().Err(err).Msg("Error adding buff")
		return false
	}

	return true
}

func (buff *Buff) RemoveFromUser(baseContext context.Context, user *User) bool {
	ctx, span := utils.StartSpan(baseContext, "Buff.RemoveFromUser")
	defer span.End()

	log.Info().Int("user", int(user.ID)).Int("buff", int(buff.ID)).Msg("Buff.RemoveFromUser")

	span.SetAttributes(
		attribute.Int("user", int(user.ID)),
		attribute.Int("buff", int(buff.ID)),
	)

	found := false
	index := -1
	for i := 0; i < len(user.Buffs) && !found; i++ {
		log.Info().Int("current", int(user.Buffs[i].BuffID)).Int("search", int(buff.ID)).Msg("Search buffs")
		if user.Buffs[i].BuffID == buff.ID {
			index = i
			found = true
		}
	}

	log.Info().Int("index", index).Int("buffs", len(user.Buffs)).Msg("Done searching")
	if index != -1 {
		user.Buffs = append(user.Buffs[:index], user.Buffs[index+1:]...)
		if result := db.WithContext(ctx).Where("user_id = ? AND round_id = ? AND buff_id = ?", user.ID, user.RoundID, buff.ID).Delete(&UserBuff{}); result.Error == nil {
			log.Info().Int("rows", int(result.RowsAffected)).Msg("Rows affected")
			return true
		} else {
			log.Error().Err(result.Error).Int("user", int(user.ID)).Int("round", user.RoundID).Int("buff", int(buff.ID)).Msg("Error removing UserBuff")
		}
	}

	return false
}

func (buff *Buff) Dump() {

	log.Warn().Msg(`
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
