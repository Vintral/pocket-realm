package temple

import (
	"context"

	"github.com/Vintral/pocket-realm/models"
	"github.com/Vintral/pocket-realm/utils"
	"github.com/rs/zerolog/log"
)

type getPantheonsResult struct {
	Type      string               `json:"type"`
	Pantheons []*models.Pantheon   `json:"pantheons"`
	Current   *models.UserDevotion `json:"current"`
}

func getUserDevotion(baseContext context.Context, user *models.User, ch chan *models.UserDevotion) {
	ctx, span := utils.StartSpan(baseContext, "temple.getUserDevotion")
	defer span.End()

	ch <- user.GetDevotion(ctx)
}

func GetPantheons(baseContext context.Context) {
	ctx, span := utils.StartSpan(baseContext, "temple.GetPantheons")
	defer span.End()

	user := baseContext.Value(utils.KeyUser{}).(*models.User)

	log.Info().Uint("user", user.ID).Msg("GetPantheons")

	pc := make(chan []*models.Pantheon)
	go models.GetPantheons(ctx, pc)

	dc := make(chan *models.UserDevotion)
	go getUserDevotion(ctx, user, dc)

	pantheons, current := <-pc, <-dc

	user.Connection.WriteJSON(getPantheonsResult{
		Type:      "GET_PANTHEONS",
		Pantheons: pantheons,
		Current:   current,
	})
}
