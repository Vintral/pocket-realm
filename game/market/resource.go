package market

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Vintral/pocket-realm/game/payloads"
	"github.com/Vintral/pocket-realm/models"
	"github.com/Vintral/pocket-realm/utilities"

	"github.com/rs/zerolog/log"
)

func BuyResource(baseContext context.Context) {
	user := baseContext.Value(utilities.KeyUser{}).(*models.User)

	ctx, span := utilities.StartSpan(baseContext, "market-buy-resource")
	defer span.End()

	var payload payloads.MarketTransactionPayload
	err := json.Unmarshal(baseContext.Value(utilities.KeyPayload{}).([]byte), &payload)
	if err != nil {
		log.Warn().AnErr("Err", err).Msg(err.Error())
		return
	}

	log.Info().Msg(fmt.Sprint("BuyResource:", user.ID, payload.GUID, payload.Quantity))

	if result := models.BuyResource(ctx, user, payload.Quantity, payload.GUID); result {
		user.Connection.WriteJSON(payloads.MarketTransactionResult{
			Type:    "MARKET_BOUGHT",
			Success: true,
		})
	} else {
		user.Connection.WriteJSON(payloads.MarketTransactionResult{
			Type:    "MARKET_BOUGHT",
			Success: false,
		})
	}
}

func SellResource(baseContext context.Context) {
	user := baseContext.Value(utilities.KeyUser{}).(*models.User)

	ctx, span := utilities.StartSpan(baseContext, "market-sell-resource")
	defer span.End()

	var payload payloads.MarketTransactionPayload
	err := json.Unmarshal(baseContext.Value(utilities.KeyPayload{}).([]byte), &payload)
	if err != nil {
		log.Warn().AnErr("Err", err).Msg(err.Error())
		return
	}

	log.Info().Msg("SellResource: " + fmt.Sprint(user.ID))

	if result := models.SellResource(ctx, user, payload.Quantity, payload.GUID); result {
		user.Connection.WriteJSON(payloads.MarketTransactionResult{
			Type:    "MARKET_SOLD",
			Success: true,
		})
	} else {
		user.Connection.WriteJSON(payloads.MarketTransactionResult{
			Type:    "MARKET_SOLD",
			Success: false,
		})
	}
}
