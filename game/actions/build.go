package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/Vintral/pocket-realm-test-access/game/utilities"
	"github.com/Vintral/pocket-realm-test-access/models"

	"go.opentelemetry.io/otel/codes"
)

type BuildPayload struct {
	Type     string `json:"type"`
	Energy   uint   `json:"energy"`
	Building string `json:"building"`
}

type BuildResult struct {
	Type string      `json:"type"`
	User models.User `json:"user"`
}

func Build(base context.Context) {
	fmt.Println("Build")

	ctx, span := utilities.StartSpan(base, "build")
	defer span.End()

	var payload BuildPayload
	err := json.Unmarshal(base.Value(utilities.KeyPayload{}).([]byte), &payload)
	if err != nil {
		fmt.Println(err)
		return
	}

	user := base.Value(utilities.KeyUser{}).(*models.User)

	fmt.Println("Get Building")
	if building := user.Round.GetBuildingByGuid(payload.Building); building != nil {
		if amount, err := building.Build(ctx, user, uint(payload.Energy)); err == nil {
			user.Connection.WriteJSON(BuildResult{
				Type: "BUILD_SUCCESS",
				User: *user,
			})

			go user.Log("Spent: "+strconv.Itoa(int(payload.Energy))+" energy building "+strconv.FormatFloat(amount, 'f', 2, 64)+" "+building.Name, user.RoundData.ID)
			span.SetStatus(codes.Ok, "OK")
			return
		} else {
			fmt.Println("Error:", err)
			span.SetStatus(codes.Error, err.Error())
		}
	}

	user.SendError(models.SendErrorParams{
		Context: &ctx,
		Type:    "build",
		Message: "build-0",
	})
}
