package social

import (
	"context"
	"encoding/json"

	"github.com/Vintral/pocket-realm/models"
	"github.com/Vintral/pocket-realm/utils"
	"github.com/rs/zerolog/log"
)

type profilePayload struct {
	Type     string `json:"type"`
	Username string `json:"username"`
}

type profileResults struct {
	Type    string `json:"type"`
	Success bool   `json:"success"`
}

func GetProfile(baseContext context.Context) {
	_, span := utils.StartSpan(baseContext, "GetProfile")
	defer span.End()

	user := baseContext.Value(utils.KeyUser{}).(*models.User)

	var payload profilePayload
	if err := json.Unmarshal(baseContext.Value(utils.KeyPayload{}).([]byte), &payload); err == nil {
		log.Info().Str("user", payload.Username).Msg("GetProfile")
	}

	user.Connection.WriteJSON(profileResults{
		Type:    "PROFILE",
		Success: true,
	})
}
