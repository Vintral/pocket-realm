package market

import (
	"context"
	"encoding/json"

	"github.com/Vintral/pocket-realm/models"
	"github.com/Vintral/pocket-realm/utils"
	"github.com/google/uuid"

	"github.com/rs/zerolog/log"
)

type UndergroundMarketInfo struct {
	Type     string                             `json:"type"`
	Auctions []*models.UndergroundMarketAuction `json:"auctions"`
}

type BuyAuctionPayload struct {
	Type    string    `json:"type"`
	Auction uuid.UUID `json:"auction"`
}

type BuyAuctionResponse struct {
	Type    string `json:"type"`
	Success bool   `json:"success"`
}

func GetUndergroundAuctions(baseContext context.Context) {
	ctx, span := utils.StartSpan(baseContext, "get-underground-auctions")
	defer span.End()

	user := baseContext.Value(utils.KeyUser{}).(*models.User)

	log.Info().Uint("user-id", user.ID).Msg("getUndergroundAuctions")

	auctions := models.GetUndergroundMarketAuctions(ctx, user)

	user.Connection.WriteJSON(UndergroundMarketInfo{
		Type:     "UNDERGROUND_MARKET",
		Auctions: auctions,
	})
}

func BuyAuction(baseContext context.Context) {
	ctx, span := utils.StartSpan(baseContext, "buy-auction")
	defer span.End()

	user := baseContext.Value(utils.KeyUser{}).(*models.User)

	var payload BuyAuctionPayload
	err := json.Unmarshal(baseContext.Value(utils.KeyPayload{}).([]byte), &payload)
	if err != nil {
		log.Warn().AnErr("Err", err).Msg(err.Error())
		return
	}

	log.Info().Uint("user-id", user.ID).Any("payload", payload).Msg("BuyAuction")

	success := models.BuyAuction(ctx, user, payload.Auction)
	log.Info().Bool("success", success).Msg("BuyAuction Call Finished")

	user.SendMessage(BuyAuctionResponse{
		Type:    "BUY_AUCTION_RESULT",
		Success: success,
	})
}
