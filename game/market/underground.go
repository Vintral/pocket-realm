package market

import (
	"context"

	"github.com/Vintral/pocket-realm/models"
	"github.com/Vintral/pocket-realm/utilities"

	"github.com/rs/zerolog/log"
)

type UndergroundMarketInfo struct {
	Type     string                             `json:"type"`
	Auctions []*models.UndergroundMarketAuction `json:"auctions"`
}

func GetUndergroundAuctions(baseContext context.Context) {
	ctx, span := utilities.StartSpan(baseContext, "get-underground-auctions")
	defer span.End()

	user := baseContext.Value(utilities.KeyUser{}).(*models.User)

	log.Info().Uint("user-id", user.ID).Msg("getUndergroundAuctions")

	auctions := models.GetUndergroundMarketAuctions(ctx, user)

	user.Connection.WriteJSON(UndergroundMarketInfo{
		Type:     "UNDERGROUND_MARKET",
		Auctions: auctions,
	})
}
