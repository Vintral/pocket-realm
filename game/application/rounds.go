package application

import (
	"context"
	"fmt"

	"github.com/Vintral/pocket-realm/game/utilities"
	"github.com/Vintral/pocket-realm/models"
	"github.com/rs/zerolog/log"
)

func GetRounds(baseCtx context.Context) {
	ctx, span := utilities.StartSpan(baseCtx, "explore")
	defer span.End()

	user := baseCtx.Value(utilities.KeyUser{}).(*models.User)

	log.Info().Msg("GetRounds: " + fmt.Sprint(user.ID))

	c := make(chan []*models.Round)
	d := make(chan []*models.Round)

	log.Info().Msg("Getting active rounds")
	go models.GetActiveRounds(ctx, c)

	log.Info().Msg("Getting past rounds")
	go models.GetPastRounds(ctx, user, d)

	log.Info().Msg("Sent requests")
	active, past := <-c, <-d

	for _, r := range active {
		r.Buildings = nil
		r.Units = nil
		r.Resources = nil

		log.Info().Msg("Cleared Data")
	}

	log.Info().Msg("Have results")

	payload := struct {
		Type   string          `json:"type"`
		Active []*models.Round `json:"active"`
		Past   []*models.Round `json:"past"`
	}{
		Type:   "ROUNDS",
		Active: active,
		Past:   past,
	}

	fmt.Println(payload)
	user.Connection.WriteJSON(payload)
}
