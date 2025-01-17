package models

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gorm"
)

type UndergroundMarketAuction struct {
	BaseModel

	GUID      uuid.UUID `gorm:"uniqueIndex,size:36" json:"-"`
	Auction   uuid.UUID `gorm:"->;-:migration" json:"auction"`
	ItemID    uint      `json:"-"`
	Item      uuid.UUID `gorm:"->;-:migration" json:"item"`
	Cost      uint      `json:"cost"`
	Expires   time.Time `json:"expires"`
	Purchased time.Time `gorm:"->;-:migration" json:"purchased"`
}

func (auction *UndergroundMarketAuction) BeforeCreate(tx *gorm.DB) (err error) {
	auction.GUID = uuid.New()
	return
}

func GetUndergroundMarketAuctions(baseContext context.Context, user *User) []*UndergroundMarketAuction {
	ctx, span := Tracer.Start(baseContext, "get-underground-market-auctions")
	defer span.End()

	var auctions []*UndergroundMarketAuction
	db.WithContext(ctx).Raw(`
		SELECT 
			underground_market_auctions.guid AS auction, items.guid AS item, underground_market_auctions.cost,
			underground_market_auctions.expires, underground_market_purchases.purchased
		FROM 
			underground_market_auctions		
		LEFT JOIN
			items
		ON
			items.id = underground_market_auctions.item_id
		LEFT JOIN 
			underground_market_purchases
		ON 
			underground_market_purchases.user_id = ` + fmt.Sprint(user.ID) + ` AND 
			underground_market_purchases.market_id = underground_market_auctions.id
		WHERE underground_market_auctions.expires > '` + time.Now().Truncate(time.Second).String() + `'`,
	).Scan(&auctions)

	span.SetAttributes(
		attribute.Int("Active auctions", len(auctions)),
	)

	return auctions
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
