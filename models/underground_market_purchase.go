package models

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type UndergroundMarketPurchase struct {
	BaseModel

	MarketID  uint      `json:"black_market_id"`
	UserID    uint      `json:"user_id"`
	Purchased time.Time `json:"purchased"`
}

func BuyAuction(baseContext context.Context, user *User, guid uuid.UUID) bool {
	ctx, span := Tracer.Start(baseContext, "buy-auction")
	defer span.End()

	log.Info().Uint("user-id", user.ID).Str("auction", guid.String()).Msg("BuyAuction")

	var auction *UndergroundMarketAuction
	db.WithContext(ctx).Where("guid = ?", guid.String()).First(&auction)

	log.Info().Any("auction", auction).Msg("Loaded Auction")
	if auction.GUID == uuid.Nil {
		return false
	}

	var purchase string
	db.WithContext(ctx).Table("underground_market_purchases").Select("purchased").Where("market_id = ? AND user_id = ?", auction.ID, user.ID).Scan(&purchase)
	log.Info().Int("length", len(purchase)).Str("purchased", purchase).Msg("Retrieved Purchase")
	if len(purchase) > 0 {
		return false
	}

	if user.TakeResource(ctx, "gold", int(auction.Cost)) {
		if user.AddItem(ctx, GetItemByID(ctx, int(auction.ItemID))) {
			if result := db.WithContext(ctx).Save(&UndergroundMarketPurchase{
				MarketID:  auction.ID,
				UserID:    user.ID,
				Purchased: time.Now(),
			}); result.Error != nil {
				log.Error().AnErr("error", result.Error).Msg("Error Adding Auction Purchase")
				return false
			}

			log.Info().Int("user-id", int(user.ID)).Int("auction-id", int(auction.ID)).Msg("Bought Auction")
			return true
		} else {
			if user.GiveResource(ctx, "gold", int(auction.Cost)) {
				log.Error().Str("resource", "gold").Uint("user-id", user.ID).Int("amount", int(auction.Cost)).Msg("Error Crediting Resource")
			}
		}
	}

	return false
}

// func BuyResource(baseContext context.Context, user *User, quantity int, resource uuid.UUID) bool {
// 	ctx, span := Tracer.Start(baseContext, "buy-resource")
// 	defer span.End()

// 	log.Warn().Msg(fmt.Sprint("BuyResource:", user.ID, user.RoundID, quantity, resource.String()))

// 	if round, err := LoadRoundById(ctx, user.RoundID); err != nil {
// 		log.Warn().AnErr("err", err).Msg("Error loading round")
// 		user.SendError(SendErrorParams{Context: &ctx, Type: "market", Message: "market-1"})
// 	} else {
// 		log.Info().Str("guid", resource.String()).Msg(fmt.Sprint("Have:", round.ID))

// 		resource := round.GetResourceByGuid(resource.String())
// 		market := round.MarketResources[resource.GUID.String()]
// 		cost := math.Ceil(float64(quantity) * float64(market.Value))

// 		go user.Log(fmt.Sprintf("Trying to buy %d %s for %d gold", quantity, resource.Name, int(math.Ceil(cost))), uint(user.RoundID))
// 		span.SetAttributes(
// 			attribute.String("resource", resource.Name),
// 			attribute.Int("quantity", quantity),
// 			attribute.Float64("cost", cost),
// 			attribute.Float64("user_gold", user.RoundData.Gold),
// 		)

// 		if !user.updateField("gold", -cost) {
// 			span.RecordError(errors.New("cannot afford"))
// 			go user.Log("Cannot afford to buy resource", uint(user.RoundID))
// 		}
// 		user.updateField(resource.Name, float64(quantity))

// 		if err := user.UpdateRound(ctx, nil); !err {
// 			log.Fatal().Msg("Error updating round")

// 			span.RecordError(errors.New("error buying resource"))
// 			go user.Log("Error buying resource", user.RoundData.ID)
// 			user.Load()
// 			return false
// 		}

// 		db.Exec("UPDATE round_market_resources SET bought = bought + ? WHERE round_id = ? AND resource_id = ?", quantity, round.ID, resource.GUID)

// 		log.Info().Msg("Purchased resource")
// 		// go user.Log(fmt.Sprintf("Bought %d %s for %d", quantity, resource.Name, int(math.Ceil(cost))), uint(user.RoundID))
// 		return true
// 	}

// 	return false
// }
