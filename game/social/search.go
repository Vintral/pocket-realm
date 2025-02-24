package social

import (
	"context"
	"encoding/json"

	"github.com/Vintral/pocket-realm/models"
	"github.com/Vintral/pocket-realm/utils"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type searchPayload struct {
	Type   string `json:"type"`
	Search string `json:"search"`
}

type searchResults struct {
	Type    string                 `json:"type"`
	Results []*models.SearchResult `json:"results"`
}

func SearchUsers(baseContext context.Context) {
	ctx, span := utils.StartSpan(baseContext, "SearchUsers")
	defer span.End()

	user := baseContext.Value(utils.KeyUser{}).(*models.User)

	max := 10
	var ret []*models.SearchResult
	found := make(map[uuid.UUID]bool)
	var payload searchPayload
	if err := json.Unmarshal(baseContext.Value(utils.KeyPayload{}).([]byte), &payload); err == nil {
		log.Info().Str("needle", payload.Search).Msg("SearchUsers")

		var data []*models.SearchResult
		if err := db.WithContext(ctx).
			Table("users").
			Select("users.avatar, users.username, MAX(user_rounds.character_class) AS class, users.guid").
			Joins("LEFT JOIN user_rounds ON users.id = user_rounds.user_id").
			Where("users.username LIKE ? AND user_rounds.round_id = ?", payload.Search+"%", user.RoundID).
			Group("username, avatar, guid").
			Order("username ASC").
			Limit(max).
			Scan(&data).Error; err == nil {

			for _, u := range data {
				log.Info().Str("name", u.Username).Str("avatar", u.Avatar).Str("class", u.Class).Str("guid", u.GUID.String()).Msg("User")
				found[u.GUID] = true

				ret = append(ret, u)
			}

			if len(ret) < max {
				if err := db.WithContext(ctx).
					Table("users").
					Select("users.avatar, users.username, MAX(user_rounds.character_class) AS class, users.guid").
					Joins("LEFT JOIN user_rounds ON users.id = user_rounds.user_id").
					Where("users.username LIKE ? AND user_rounds.round_id = ?", "%"+payload.Search+"%", user.RoundID).
					Group("username, avatar, guid").
					Order("username ASC").
					Limit(max).
					Scan(&data).Error; err == nil {

					for _, u := range data {
						if found[u.GUID] {
							log.Info().Str("user", u.Username).Msg("Found user already")
							continue
						} else {
							found[u.GUID] = true
							ret = append(ret, u)
						}
					}
				}
			}
		}
	}

	if len(ret) > max {
		ret = ret[0:max]
	}

	user.Connection.WriteJSON(searchResults{
		Type:    "SEARCH_RESULTS",
		Results: ret,
	})
}
