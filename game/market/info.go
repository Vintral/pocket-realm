package market

import (
	"context"
	"fmt"

	"github.com/Vintral/pocket-realm/models"
	"github.com/Vintral/pocket-realm/utils"

	"github.com/rs/zerolog/log"
)

type GetMarketInfoResult struct {
	Type   string          `json:"type"`
	Events []*models.Event `json:"events"`
	Page   int             `json:"page"`
	Max    int             `json:"max"`
}

func GetInfo(baseContext context.Context) {
	user := baseContext.Value(utils.KeyUser{}).(*models.User)

	ctx, span := utils.StartSpan(baseContext, "market-info")
	defer span.End()

	log.Info().Msg("GetMarketInfo: " + fmt.Sprint(user.ID))

	if round, err := models.LoadRoundById(ctx, user.RoundID); err != nil {
		log.Warn().AnErr("err", err).Msg("Error loading round")
		user.SendError(models.SendErrorParams{Context: &ctx, Type: "market", Message: "market-1"})
	} else {
		log.Info().Any("round", round).Msg(fmt.Sprint("Have:", round.ID))

		user.Connection.WriteJSON(struct {
			Type      string                        `json:"type"`
			Resources []*models.RoundMarketResource `json:"resources"`
		}{
			Type:      "MARKET_INFO",
			Resources: round.GetMarketInfo(ctx),
		})
	}
}
