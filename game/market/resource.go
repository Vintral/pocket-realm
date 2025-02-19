package market

import (
	"context"
	"encoding/json"

	"github.com/Vintral/pocket-realm/models"
	"github.com/Vintral/pocket-realm/utils"
	"github.com/google/uuid"

	"github.com/rs/zerolog/log"
)

type MarketTransactionPayload struct {
	Type     string    `json:"type"`
	Quantity uint      `json:"quantity"`
	Resource uuid.UUID `json:"resource"`
}

type MarketTransactionResult struct {
	Type    string `json:"type"`
	Success bool   `json:"success"`
}

func BuyResource(baseContext context.Context) {
	user := baseContext.Value(utils.KeyUser{}).(*models.User)

	ctx, span := utils.StartSpan(baseContext, "market-buy-resource")
	defer span.End()

	var payload MarketTransactionPayload
	err := json.Unmarshal(baseContext.Value(utils.KeyPayload{}).([]byte), &payload)
	if err != nil {
		log.Warn().AnErr("Err", err).Msg(err.Error())
		return
	}

	if success := user.BuyResource(ctx, payload.Quantity, payload.Resource); success {
		user.Connection.WriteJSON(MarketTransactionResult{
			Type:    "MARKET_BOUGHT",
			Success: true,
		})
	} else {
		user.Connection.WriteJSON(MarketTransactionResult{
			Type:    "MARKET_BOUGHT",
			Success: false,
		})
	}
}

func SellResource(baseContext context.Context) {
	user := baseContext.Value(utils.KeyUser{}).(*models.User)

	ctx, span := utils.StartSpan(baseContext, "market-sell-resource")
	defer span.End()

	var payload MarketTransactionPayload
	err := json.Unmarshal(baseContext.Value(utils.KeyPayload{}).([]byte), &payload)
	if err != nil {
		log.Warn().AnErr("Err", err).Msg(err.Error())
		return
	}

	if result := user.SellResource(ctx, payload.Quantity, payload.Resource); result {
		user.Connection.WriteJSON(MarketTransactionResult{
			Type:    "MARKET_SOLD",
			Success: true,
		})
	} else {
		user.Connection.WriteJSON(MarketTransactionResult{
			Type:    "MARKET_SOLD",
			Success: false,
		})
	}
}
