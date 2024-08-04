package player

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Vintral/pocket-realm//utilities"
	"github.com/Vintral/pocket-realm/models"
	"github.com/google/uuid"

	"github.com/rs/zerolog/log"
)

type GetEventsRequest struct {
	Type string `json:"type"`
	Page int    `json:"page"`
}

type MarkEventSeenRequest struct {
	Event string `json:"event"`
}

type GetEventsResult struct {
	Type   string          `json:"type"`
	Events []*models.Event `json:"events"`
	Page   int             `json:"page"`
	Max    int             `json:"max"`
}

func HandleMarkEventSeen(baseContext context.Context) {
	user := baseContext.Value(utilities.KeyUser{}).(*models.User)

	// ctx, span := utilities.StartSpan(baseContext, "mark-event-seen")
	// defer span.End()

	var payload MarkEventSeenRequest
	err := json.Unmarshal(baseContext.Value(utilities.KeyPayload{}).([]byte), &payload)
	if err != nil {
		fmt.Println(err)
		return
	}

	guid, err := uuid.Parse(payload.Event)
	if err != nil {
		log.Warn().AnErr("error", err).Msg("Error Marking Event Seen: " + payload.Event)
		return
	}

	log.Warn().Msg("MarkEventSeen: " + fmt.Sprint(user.ID) + " - " + payload.Event)
	models.MarkEventSeen(baseContext, guid)

	// log.Info().Msg("GetEvents: " + fmt.Sprint(user.ID))

	// user.Connection.WriteJSON(GetEventsResult{
	// 	Type:   "EVENTS",
	// 	Events: models.LoadEvents(ctx, int(user.ID), user.Round.GUID, payload.Page),
	// 	Page:   payload.Page,
	// 	Max:    models.MaxEventPages(ctx, int(user.ID), user.Round.GUID),
	// })
}

func GetEvents(baseContext context.Context) {
	user := baseContext.Value(utilities.KeyUser{}).(*models.User)

	ctx, span := utilities.StartSpan(baseContext, "player-get-events")
	defer span.End()

	var payload GetEventsRequest
	err := json.Unmarshal(baseContext.Value(utilities.KeyPayload{}).([]byte), &payload)
	if err != nil {
		fmt.Println(err)
		return
	}

	log.Info().Msg("GetEvents: " + fmt.Sprint(user.ID))

	user.Connection.WriteJSON(GetEventsResult{
		Type:   "EVENTS",
		Events: models.LoadEvents(ctx, int(user.ID), user.Round.GUID, payload.Page),
		Page:   payload.Page,
		Max:    models.MaxEventPages(ctx, int(user.ID), user.Round.GUID),
	})
}
