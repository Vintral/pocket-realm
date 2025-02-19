package library

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/Vintral/pocket-realm/models"
	"github.com/Vintral/pocket-realm/utils"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type PurchaseTechnologyPayload struct {
	Technology uuid.UUID `json:"technology"`
}

type PurchaseTechnologyResult struct {
	Type    string `json:"type"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type GetTechnologiesResult struct {
	Type         string               `json:"type"`
	Technologies []*models.Technology `json:"technologies"`
}

func GetTechnologies(baseContext context.Context) {
	ctx, span := utils.StartSpan(baseContext, "library.get-technologies")
	defer span.End()

	log.Warn().Msg("GetTechnologies")

	user := baseContext.Value(utils.KeyUser{}).(*models.User)
	if round, err := models.LoadRoundByGuid(ctx, user.RoundPlaying); err == nil {
		fmt.Println(round.Technologies)

		log.Warn().Int("technology_count", len(round.Technologies)).Msg("Library: Get Technologies")

		techs := make([]*models.Technology, 0)

		wg := new(sync.WaitGroup)
		wg.Add(len(round.Technologies))
		for _, technology := range round.Technologies {
			tech := &models.Technology{}
			techs = append(techs, tech)
			go technology.LoadForUser(ctx, wg, user, tech)
		}
		wg.Wait()

		user.Connection.WriteJSON(GetTechnologiesResult{
			Type:         "GET_TECHNOLOGIES",
			Technologies: techs,
		})

	} else {
		log.Error().AnErr("err", err).Msg("Error Loading Round")
	}
}

func PurchaseTechnology(baseContext context.Context) {
	ctx, span := utils.StartSpan(baseContext, "purchase-technology")
	defer span.End()

	user := baseContext.Value(utils.KeyUser{}).(*models.User)

	var payload PurchaseTechnologyPayload
	if err := json.Unmarshal(baseContext.Value(utils.KeyPayload{}).([]byte), &payload); err == nil {
		techId := models.GetTechnologyIdForGuid(ctx, payload.Technology)

		if round, err := models.LoadRoundById(ctx, user.RoundID); err == nil {
			technology := round.GetTechnologyById(techId)

			tech := &models.Technology{}
			technology.LoadForUser(ctx, nil, user, tech)
			if user.PurchaseTechnology(ctx, tech) {
				log.Info().Int("user", int(user.ID)).Int("technology", int(tech.ID)).Int("level", int(tech.Level+1)).Msg("Purchased technology level")
				user.Connection.WriteJSON(PurchaseTechnologyResult{
					Type:    "PURCHASE_TECHNOLOGY",
					Success: true,
				})
				return
			} else {
				log.Warn().Int("user", int(user.ID)).Int("technology", int(tech.ID)).Int("level", int(tech.Level+1)).Msg("Failed to purchase technology level")
			}
		}
	}

	user.Connection.WriteJSON(PurchaseTechnologyResult{
		Type:    "PURCHASE_TECHNOLOGY",
		Success: false,
		Message: "purchase-technology-error-1",
	})
}
