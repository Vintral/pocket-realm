package models

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/attribute"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoundMarketResource struct {
	BaseModel

	GUID       uuid.UUID `gorm:"uniqueIndex,size:36" json:"-"`
	RoundID    uint      `json:"-"`
	ResourceID uuid.UUID `json:"resource"`
	Sold       uint      `json:"-"`
	Bought     uint      `json:"-"`
	Value      float32   `json:"value"`
}

func (resource *RoundMarketResource) BeforeCreate(tx *gorm.DB) (err error) {
	resource.GUID = uuid.New()
	return
}

func BuyResource(baseContext context.Context, user *User, quantity int, resource uuid.UUID) bool {
	ctx, span := Tracer.Start(baseContext, "buy-resource")
	defer span.End()

	log.Warn().Msg(fmt.Sprint("BuyResource:", user.ID, user.RoundID, quantity, resource.String()))

	if round, err := LoadRoundById(ctx, user.RoundID); err != nil {
		log.Warn().AnErr("err", err).Msg("Error loading round")
		user.SendError(SendErrorParams{Context: &ctx, Type: "market", Message: "market-1"})
	} else {
		log.Info().Str("guid", resource.String()).Msg(fmt.Sprint("Have:", round.ID))

		resource := round.GetResourceByGuid(resource.String())
		market := round.MarketResources[resource.GUID.String()]
		cost := math.Ceil(float64(quantity) * float64(market.Value))

		go user.Log(fmt.Sprintf("Trying to buy %d %s for %d gold", quantity, resource.Name, int(math.Ceil(cost))), uint(user.RoundID))
		span.SetAttributes(
			attribute.String("resource", resource.Name),
			attribute.Int("quantity", quantity),
			attribute.Float64("cost", cost),
			attribute.Float64("user_gold", user.RoundData.Gold),
		)

		if !user.updateField("gold", -cost) {
			span.RecordError(errors.New("cannot afford"))
			go user.Log("Cannot afford to buy resource", uint(user.RoundID))
		}
		user.updateField(resource.Name, float64(quantity))

		if err := user.UpdateRound(ctx, nil); !err {
			log.Fatal().Msg("Error updating round")

			span.RecordError(errors.New("error buying resource"))
			go user.Log("Error buying resource", user.RoundData.ID)
			user.Load()
			return false
		}

		db.Exec("UPDATE round_market_resources SET bought = bought + ? WHERE round_id = ? AND resource_id = ?", quantity, round.ID, resource.GUID)

		log.Info().Msg("Purchased resource")
		// go user.Log(fmt.Sprintf("Bought %d %s for %d", quantity, resource.Name, int(math.Ceil(cost))), uint(user.RoundID))
		return true
	}

	return false
}

func SellResource(baseContext context.Context, user *User, quantity int, resource uuid.UUID) bool {
	ctx, span := Tracer.Start(baseContext, "sell-resource")
	defer span.End()

	log.Warn().Msg(fmt.Sprint("SellResource:", user.ID, user.RoundID, quantity, resource.String()))

	if round, err := LoadRoundById(ctx, user.RoundID); err != nil {
		log.Warn().AnErr("err", err).Msg("Error loading round")
		user.SendError(SendErrorParams{Context: &ctx, Type: "market", Message: "market-1"})
	} else {
		log.Info().Str("guid", resource.String()).Msg(fmt.Sprint("Have:", round.ID))

		resource := round.GetResourceByGuid(resource.String())
		market := round.MarketResources[resource.GUID.String()]
		cost := math.Floor(float64(quantity) * float64(1/market.Value))

		go user.Log(fmt.Sprintf("Trying to sell %d %s for %d gold", quantity, resource.Name, int(math.Ceil(cost))), uint(user.RoundID))
		span.SetAttributes(
			attribute.String("resource", resource.Name),
			attribute.Int("quantity", quantity),
			attribute.Float64("cost", cost),
			attribute.Float64("user_gold", user.RoundData.Gold),
		)

		user.updateField("gold", -cost)
		if !user.updateField(resource.Name, float64(quantity)) {
			msg := fmt.Sprint("not enough ", resource.Name)
			go user.Log(msg, uint(user.RoundID))
			log.Info().Msg(msg)
			return false
		}

		if err := user.UpdateRound(ctx, nil); !err {
			msg := "error updating round"
			log.Fatal().Msg(msg)
			span.RecordError(errors.New(msg))
			go user.Log(msg, user.RoundData.ID)

			user.Load()
			return false
		}

		db.Exec("UPDATE round_market_resources SET sold = sold + ? WHERE round_id = ? AND resource_id = ?", quantity, round.ID, resource.GUID)

		msg := fmt.Sprintf("Sold %d %s for %d gold", quantity, resource.Name, int(math.Ceil(cost)))
		log.Info().Msg(msg)
		go user.Log(msg, uint(user.RoundID))
		return true
	}

	return false
}
