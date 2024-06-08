package social

import (
	"context"
	"encoding/json"
	"fmt"
	"realm/models"
	"time"

	"realm/utilities"

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
	_, span := utilities.StartSpan(base, "subscribe-shouts")
	defer span.End()

	user := base.Value(utilities.KeyUser{}).(*models.User)
	subscribers[user.ID] = user

	fmt.Println("Shout Subscribers:", len(subscribers))
}

func UnsubscribeShouts(base context.Context) {
	_, span := utilities.StartSpan(base, "unsubscribe-shouts")
	defer span.End()

	user := base.Value(utilities.KeyUser{}).(*models.User)
	delete(subscribers, user.ID)

	fmt.Println("Shout Subscribers:", len(subscribers))
}

func SendShout(base context.Context) {
	fmt.Println("SendShout")

	_, span := utilities.StartSpan(base, "send-shout")
	defer span.End()

	var payload ShoutPayload
	err := json.Unmarshal(base.Value(utilities.KeyPayload{}).([]byte), &payload)
	if err != nil {
		fmt.Println(err)
		return
	}

	user := base.Value(utilities.KeyUser{}).(*models.User)

	var shout *models.Shout
	if err = shout.Create(user.ID, payload.Shout); err == nil {
		span.SetStatus(codes.Ok, "OK")
	} else {
		span.SetStatus(codes.Error, err.Error())
	}

	user.Connection.WriteJSON(ShoutResult{
		Type:    "SHOUT",
		Success: err == nil,
	})

	go dispatchMessage(ShoutDataPayload{
		Type: "SHOUT",
		Shout: models.Shout{
			User:           user.Username,
			Avatar:         "",
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
