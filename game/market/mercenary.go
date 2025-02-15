package market

import (
	"context"
	"encoding/json"

	"github.com/Vintral/pocket-realm/models"
	"github.com/Vintral/pocket-realm/utils"

	"github.com/rs/zerolog/log"
)

type BuyMercenaryPayload struct {
	Quantity int `json:"quantity"`
}

type BuyMercenaryResult struct {
	Type    string `json:"type"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func GetMercenaryMarket(baseContext context.Context) {
	ctx, span := utils.StartSpan(baseContext, "market-get-mercenary-market")
	defer span.End()

	user := baseContext.Value(utils.KeyUser{}).(*models.User)

	models.GetMercenaryMarket(ctx, user)
}

func BuyMercenary(baseContext context.Context) {
	ctx, span := utils.StartSpan(baseContext, "market-buy-mercenary")
	defer span.End()

	user := baseContext.Value(utils.KeyUser{}).(*models.User)

	var payload BuyMercenaryPayload
	err := json.Unmarshal(baseContext.Value(utils.KeyPayload{}).([]byte), &payload)
	if err != nil {
		log.Warn().AnErr("Err", err).Msg(err.Error())
		return
	}

	log.Info().Uint("user-id", user.ID).Int("quantity", payload.Quantity).Msg("BuyMercenary")
	if payload.Quantity != 0 {
		if result := models.BuyMercenary(ctx, user, payload.Quantity); result {
			user.Connection.WriteJSON(BuyMercenaryResult{
				Type:    "BUY_MERCENARY",
				Success: true,
			})
		}
	} else {
		user.Connection.WriteJSON(BuyMercenaryResult{
			Type:    "BUY_MERCENARY",
			Success: false,
			Message: "error-mercenary-0",
		})
	}

	user.Refresh()
}
