package player

import (
	"context"
	"encoding/json"

	"github.com/Vintral/pocket-realm/models"
	"github.com/Vintral/pocket-realm/utils"

	"github.com/rs/zerolog/log"
)

type ChangeAvatarRequest struct {
	Type   string `json:"type"`
	Avatar string `json:"avatar"`
}

type ChangeAvatarResult struct {
	Type    string `json:"type"`
	Success bool   `json:"success"`
}

func ChangeAvatar(baseContext context.Context) {
	ctx, span := utils.StartSpan(baseContext, "change-avatar")
	defer span.End()

	user := baseContext.Value(utils.KeyUser{}).(*models.User)
	success := false

	var payload ChangeAvatarRequest
	if err := json.Unmarshal(baseContext.Value(utils.KeyPayload{}).([]byte), &payload); err == nil {
		success = user.ChangeAvatar(ctx, payload.Avatar)
		if success {
			log.Info().Msg("Refresh User")
			user.Refresh()
		}

		log.Info().Bool("success", success).Int("user", int(user.ID)).Str("avatar", payload.Avatar).Msg("Updated user avatar")
	} else {
		log.Warn().AnErr("err", err).Int("user", int(user.ID)).Msg("Error changing avatar")
	}

	user.Connection.WriteJSON(ChangeAvatarResult{
		Type:    "CHANGE_AVATAR",
		Success: success,
	})
}
