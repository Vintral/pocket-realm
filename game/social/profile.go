package social

import (
	"context"
	"encoding/json"

	"github.com/Vintral/pocket-realm/models"
	"github.com/Vintral/pocket-realm/utils"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type profilePayload struct {
	Type     string `json:"type"`
	Username string `json:"username"`
}

type profileResults struct {
	Type     string    `json:"type"`
	Success  bool      `json:"success"`
	Username string    `json:"username"`
	UserGuid uuid.UUID `json:"guid"`
	Avatar   string    `json:"avatar"`
	Friend   bool      `json:"friend"`
	Enemy    bool      `json:"enemy"`
	Blocked  bool      `json:"blocked"`
}

func GetProfile(baseContext context.Context) {
	ctx, span := utils.StartSpan(baseContext, "GetProfile")
	defer span.End()

	user := baseContext.Value(utils.KeyUser{}).(*models.User)

	var payload profilePayload
	if err := json.Unmarshal(baseContext.Value(utils.KeyPayload{}).([]byte), &payload); err == nil {
		log.Info().Str("user", payload.Username).Msg("GetProfile")
	}

	var profileUser *models.User
	if err := db.WithContext(ctx).Table("users").Select("id, guid, username, avatar").Where("username = ?", payload.Username).Scan(&profileUser).Error; err == nil {
		ret := profileResults{
			Type:     "PROFILE",
			Success:  profileUser.Username == payload.Username,
			Username: payload.Username,
			UserGuid: profileUser.GUID,
			Avatar:   profileUser.Avatar,
		}

		var data []*models.Contact
		if err = db.WithContext(ctx).Table("contacts").Select("category").Where("contact_id = ? AND user_id = ?", profileUser.ID, user.ID).Scan(&data).Error; err == nil {
			for _, contact := range data {
				switch contact.Category {
				case "friend":
					ret.Friend = true
				case "enemy":
					ret.Enemy = true
				case "blocked":
					ret.Blocked = true
				}
			}
		} else {
			log.Warn().Int("contact", int(profileUser.ID)).Int("user", int(user.ID)).Msg("Error looking up contact entries")
		}

		user.Connection.WriteJSON(&ret)
	} else {
		user.Connection.WriteJSON(profileResults{
			Type:    "PROFILE",
			Success: false,
		})
	}
}
