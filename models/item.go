package models

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type Item struct {
	BaseModel

	ID          uint      `gorm:"primaryKey" json:"order"`
	GUID        uuid.UUID `gorm:"uniqueIndex,size:36" json:"guid"`
	Name        string    `json:"name"`
	EffectList  string    `json:"-"`
	Effects     []*Effect `gorm:"-" json:"effects"`
	Plural      string    `json:"plural"`
	Description string    `json:"description"`
}

var items []*Item
var itemsById = make(map[int]*Item)
var itemsByGuid = make(map[uuid.UUID]*Item)

func (item *Item) BeforeCreate(tx *gorm.DB) (err error) {
	item.GUID = uuid.New()
	return
}

func (item *Item) AfterFind(tx *gorm.DB) (err error) {
	ctx, sp := Tracer.Start(tx.Statement.Context, "item.AfterFind")
	defer sp.End()

	log.Trace().Msg("item.AfterFind")

	for _, effectId := range strings.Split(item.EffectList, ",") {
		log.Info().Str("effect", effectId).Msg("Load Effect")
		var effect *Effect
		if err := db.WithContext(ctx).Table("effects").Where("id = ?", effectId).Scan(&effect).Error; err != nil {
			log.Error().Err(err).Str("effect", effectId).Msg("Error loading effect")
		} else {
			item.Effects = append(item.Effects, effect)
		}
	}

	return
}

func (item *Item) Use(baseContext context.Context, user *User) bool {
	ctx, span := Tracer.Start(baseContext, "item-use")
	defer span.End()

	log.Info().Int("effects", len(item.Effects)).Msg("item.Use")

	roundUpdated := false

	for _, effect := range item.Effects {
		log.Info().Any("effect", effect).Msg("Process Effect")

		switch effect.Type {
		case "resource":
			switch effect.Field {
			case "energy":
				user.RoundData.Energy += int(effect.Amount)
			case "food":
				user.RoundData.Food += float64(effect.Amount)
			}

			roundUpdated = true
		}
	}

	log.Info().Int("energy", user.RoundData.Energy).Int("food", int(user.RoundData.Food)).Msg("======== > UserInfo")

	if roundUpdated {
		if !user.UpdateRound(ctx, nil) {
			log.Error().Uint("id", user.ID).Msg("Error updating round for user")
			return false
		}
	}

	return user.TakeItem(ctx, item)
}

func GetItemByID(baseContext context.Context, id int) *Item {
	ctx, span := Tracer.Start(baseContext, "models.GetItemById")
	defer span.End()

	log.Info().Msg(fmt.Sprint("GetItemByID: ", id))

	var val *Item
	if _, ok := itemsById[id]; !ok {
		log.Info().Int("item", id).Msg("Load Item")
		if err := db.WithContext(ctx).Table("items").Where("id = ?", id).Find(&val).Error; err == nil {
			if val == nil {
				log.Warn().Int("id", int(id)).Msg("Failed to load item")
				return nil
			}
		} else {
			log.Error().AnErr("err", err).Int("item", id).Msg("Error loading item")
		}

		log.Info().Int("id", int(id)).Msg("Grabbed item")
		itemsById[id] = val
		itemsByGuid[val.GUID] = val
	}

	val = itemsById[id]
	return val
}

func GetItemByGUID(baseContext context.Context, guid uuid.UUID) *Item {
	ctx, span := Tracer.Start(baseContext, "get-item-by-id")
	defer span.End()

	log.Info().Msg(fmt.Sprint("GetItemByGUID: ", guid.String()))

	var val *Item
	if _, ok := itemsByGuid[guid]; !ok {
		db.WithContext(ctx).Table("items").Where("guid = ?", guid).Scan(&val)
		if val == nil {
			log.Warn().Str("guid", guid.String()).Msg("Failed to load item")
			return nil
		}

		log.Info().Str("guid", guid.String()).Msg("Grabbed item")
		itemsByGuid[guid] = val
		itemsById[int(val.ID)] = val
	}

	val = itemsByGuid[guid]
	return val
}

func GetItems(baseContext context.Context) []*Item {
	ctx, span := Tracer.Start(baseContext, "get-items")
	defer span.End()

	if len(items) == 0 {
		log.Info().Msg("GetItems")
		db.WithContext(ctx).Find(&items)
		log.Info().Int("item-count", len(items)).Msg("Loaded Items")
	}

	return items
}
