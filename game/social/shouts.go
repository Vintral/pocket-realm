package social

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Vintral/pocket-realm/models"
	"github.com/Vintral/pocket-realm/utils"
	"github.com/rs/zerolog/log"

	"go.opentelemetry.io/otel/codes"
)

var subscribers map[uint]*models.User

type ShoutPayload struct {
	Type  string `json:"type"`
	Shout string `json:"shout"`
}

type ShoutResult struct {
	Type    string `json:"type"`
	Success bool   `json:"success"`
}

type ShoutDataPayload struct {
	Type  string       `json:"type"`
	Shout models.Shout `json:"shout"`
}

func (shout ShoutDataPayload) MarshalBinary() ([]byte, error) {
	return json.Marshal(shout)
}

func dispatchMessage(shout ShoutDataPayload) {
	if redisClient != nil {
		err := redisClient.Publish(context.Background(), "SHOUT", shout)
		fmt.Println(err)
	}
}

func SubscribeShouts(base context.Context) {
	_, span := utils.StartSpan(base, "subscribe-shouts")
	defer span.End()

	user := base.Value(utils.KeyUser{}).(*models.User)
	subscribers[user.ID] = user

	fmt.Println("Shout Subscribers:", len(subscribers))
}

func UnsubscribeShouts(base context.Context) {
	_, span := utils.StartSpan(base, "unsubscribe-shouts")
	defer span.End()

	user := base.Value(utils.KeyUser{}).(*models.User)
	delete(subscribers, user.ID)

	fmt.Println("Shout Subscribers:", len(subscribers))
}

func SendShout(base context.Context) {
	fmt.Println("SendShout")

	ctx, span := utils.StartSpan(base, "send-shout")
	defer span.End()

	var payload ShoutPayload
	err := json.Unmarshal(base.Value(utils.KeyPayload{}).([]byte), &payload)
	if err != nil {
		fmt.Println(err)
		return
	}

	user := base.Value(utils.KeyUser{}).(*models.User)

	var shout *models.Shout
	if err = shout.Create(ctx, user.ID, payload.Shout); err == nil {
		span.SetStatus(codes.Ok, "OK")
		log.Info().Int("user", int(user.ID)).Str("shout", payload.Shout).Msg("Shout sent")
	} else {
		span.SetStatus(codes.Error, err.Error())
		log.Warn().Err(err).Msg("Error sending shout")
	}

	user.Connection.WriteJSON(ShoutResult{
		Type:    "SEND_SHOUT",
		Success: err == nil,
	})

	go dispatchMessage(ShoutDataPayload{
		Type: "SHOUT",
		Shout: models.Shout{
			User:           user.Username,
			Avatar:         user.Avatar,
			CharacterClass: user.RoundData.CharacterClass,
			CreatedAt:      time.Now(),
			Shout:          payload.Shout,
		},
	})
}

func GetShouts(user *models.User) {
	fmt.Println("Get Shouts:", user.ID)

	var shout *models.Shout
	payload := struct {
		Type   string              `json:"type"`
		Shouts []*models.ShoutData `json:"shouts"`
	}{
		Type:   "SHOUTS",
		Shouts: shout.Load(1),
	}
	user.Connection.WriteJSON(payload)
}
