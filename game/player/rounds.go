package player

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Vintral/pocket-realm/models"
	"github.com/Vintral/pocket-realm/utils"
	"github.com/google/uuid"

	"github.com/rs/zerolog/log"
)

type PlayRoundRequest struct {
	Round string `json:"round"`
}

// type MarkEventSeenRequest struct {
// 	Event string `json:"event"`
// }

type PlayRoundResult struct {
	Type  string      `json:"type"`
	User  models.User `json:"user"`
	Round string      `json:"round"`
}

func joinRoundIfNeeded(baseContext context.Context, user *models.User, round *models.Round) {
	if user.IsPlayingRound(baseContext, int(round.ID)) {
		log.Info().Msg("User already playing round")
	} else {
		log.Info().Msg("User needs to join round")

		user.Join(baseContext, round)
	}
}

func PlayRound(baseContext context.Context) {
	user := baseContext.Value(utils.KeyUser{}).(*models.User)

	ctx, span := utils.StartSpan(baseContext, "player-play-round")
	defer span.End()

	var payload PlayRoundRequest
	err := json.Unmarshal(baseContext.Value(utils.KeyPayload{}).([]byte), &payload)
	if err != nil {
		fmt.Println(err)
		return
	}

	log.Info().Msg("PlayRound: " + fmt.Sprint(user.ID) + " - " + payload.Round)

	if roundGuid, err := uuid.Parse(payload.Round); err == nil {
		if round, err := models.LoadRoundByGuid(ctx, roundGuid); err == nil {
			joinRoundIfNeeded(ctx, user, round)
			if user.SwitchRound(round) {
				user.Refresh()
				user.Connection.WriteJSON(PlayRoundResult{
					Type:  "PLAY_ROUND_SUCCESS",
					User:  *user,
					Round: payload.Round,
				})
			} else {
				log.Warn().Str("guid", payload.Round).Msg("Error switching to round")
				user.Connection.WriteJSON(PlayRoundResult{
					Type: "PLAY_ROUND_ERROR",
				})
			}
		} else {
			log.Warn().Str("guid", payload.Round).Msg("Error loading round")
			user.Connection.WriteJSON(PlayRoundResult{
				Type: "PLAY_ROUND_ERROR",
			})
		}
	} else {
		log.Warn().Str("guid", payload.Round).Msg("Error parsing round guid")
		user.Connection.WriteJSON(PlayRoundResult{
			Type: "PLAY_ROUND_ERROR",
		})
	}
}
