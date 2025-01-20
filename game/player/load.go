package player

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Vintral/pocket-realm/models"
	"github.com/Vintral/pocket-realm/utils"
)

type LoadUserRequest struct {
	Type string `json:"type"`
	User string `json:"user"`
}

type LoadUserResult struct {
	Type string      `json:"type"`
	User models.User `json:"user"`
}

func Load(baseContext context.Context) {
	fmt.Println("Load User")

	fmt.Println("TraceProvider:", baseContext.Value(utils.KeyTraceProvider{}))
	fmt.Println("User:", baseContext.Value(utils.KeyUser{}))
	fmt.Println("Packet:", baseContext.Value(utils.KeyPayload{}))

	user := baseContext.Value(utils.KeyUser{}).(*models.User)

	_, span := utils.StartSpan(baseContext, "user-load")
	defer span.End()

	var payload LoadUserRequest
	err := json.Unmarshal(baseContext.Value(utils.KeyPayload{}).([]byte), &payload)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(payload)
	fmt.Println(payload.User)

	if payload.User == "" {
		user.Connection.WriteJSON(LoadUserResult{
			Type: "USER_DATA",
			User: *user,
		})
	}

	// var rule *models.Rule
	// payload := struct {
	// 	Type  string         `json:"type"`
	// 	Rules []*models.Rule `json:"rules"`
	// }{
	// 	Type:  "RULES",
	// 	Rules: rule.Load(),
	// }
	// user.Connection.WriteJSON(payload)
}
