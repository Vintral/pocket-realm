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
		WHERE underground_market_auctions.expires > '` + time.Now().Truncate(time.Second).String() + `'
		ORDER BY purchased`,
	).Scan(&auctions)

	span.SetAttributes(
		attribute.Int("Active auctions", len(auctions)),
	)

	return auctions
}
