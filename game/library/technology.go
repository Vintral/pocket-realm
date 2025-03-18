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
	ctx, span := utils.StartSpan(baseContext, "library.getTechnologies")
	defer span.End()

	log.Info().Msg("library.GetTechnologies")

	user := baseContext.Value(utils.KeyUser{}).(*models.User)
	if round, err := models.LoadRoundById(ctx, user.RoundID); err == nil {
		fmt.Println(round.Technologies)

		techs := make([]*models.Technology, 0)

		wg := new(sync.WaitGroup)
		wg.Add(len(round.Technologies))
		for _, technology := range round.Technologies {
			tech := &models.Technology{}
			techs = append(techs, tech)
			go technology.LoadForUser(ctx, wg, user, tech)
		}
		wg.Wait()

		for _, tech := range techs {
			tech.Dump()
		}

		user.Connection.WriteJSON(GetTechnologiesResult{
			Type:         "GET_TECHNOLOGIES",
			Technologies: techs,
		})

	} else {
		log.Error().AnErr("err", err).Msg("Error Loading Round")
	}
}

func PurchaseTechnology(baseContext context.Context) {
	ctx, span := utils.StartSpan(baseContext, "library.PurchaseTechnology")
	defer span.End()

	user := baseContext.Value(utils.KeyUser{}).(*models.User)
	success := false

	var payload PurchaseTechnologyPayload
	if err := json.Unmarshal(baseContext.Value(utils.KeyPayload{}).([]byte), &payload); err == nil {
		if techId := models.GetTechnologyIdForGuid(ctx, payload.Technology); techId != 0 {
			if round, err := models.LoadRoundById(ctx, user.RoundID); err == nil {
				technology := round.GetTechnologyById(techId)

				tech := &models.Technology{}
				technology.LoadForUser(ctx, nil, user, tech)

				if user.PurchaseTechnology(ctx, tech) {
					log.Info().Int("user", int(user.ID)).Int("technology", int(tech.ID)).Int("level", int(tech.Level+1)).Msg("Purchased technology level")

					user.Refresh()
					GetTechnologies(baseContext)
					success = true
				} else {
					log.Warn().Int("user", int(user.ID)).Int("technology", int(tech.ID)).Int("level", int(tech.Level+1)).Msg("Failed to purchase technology level")
				}
			}
		}
	}

	user.Connection.WriteJSON(PurchaseTechnologyResult{
		Type:    "PURCHASE_TECHNOLOGY",
		Success: success,
	})
}
