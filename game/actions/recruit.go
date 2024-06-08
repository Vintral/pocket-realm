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

type RecruitPayload struct {
	Type   string `json:"type"`
	Energy uint   `json:"energy"`
	Unit   string `json:"unit"`
}

type RecruitResult struct {
	Type string      `json:"type"`
	User models.User `json:"user"`
}

func Recruit(base context.Context) {
	fmt.Println("Recruit")

	ctx, span := utilities.StartSpan(base, "build")
	defer span.End()

	var payload RecruitPayload
	err := json.Unmarshal(base.Value(utilities.KeyPayload{}).([]byte), &payload)
	if err != nil {
		fmt.Println(err)
		return
	}

	user := base.Value(utilities.KeyUser{}).(*models.User)

	fmt.Println("Get Unit")
	if unit := user.Round.GetUnitByGuid(payload.Unit); unit != nil {
		if amount, err := unit.Recruit(ctx, user, uint(payload.Energy)); err == nil {
			user.Connection.WriteJSON(BuildResult{
				Type: "RECRUIT_SUCCESS",
				User: *user,
			})

			go user.Log("Spent: "+strconv.Itoa(int(payload.Energy))+" energy recruiting "+strconv.FormatFloat(amount, 'f', 2, 64)+" "+unit.Name, user.RoundData.ID)
			span.SetStatus(codes.Ok, "OK")
			return
		} else {
			span.SetStatus(codes.Error, err.Error())
		}
	}

	user.SendError(models.SendErrorParams{Context: &ctx, Type: "recruit", Message: "recruit-0"})
}
