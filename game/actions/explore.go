package actions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"time"

	"github.com/Vintral/pocket-realm/game/payloads"
	"github.com/Vintral/pocket-realm/game/utilities"
	"github.com/Vintral/pocket-realm/models"

	"go.opentelemetry.io/otel/attribute"
)

func Explore(baseCtx context.Context) {
	fmt.Println("Explore")

	ctx, span := utilities.StartSpan(baseCtx, "explore")
	defer span.End()

	var payload payloads.ExplorePayload
	err := json.Unmarshal(baseCtx.Value(utilities.KeyPayload{}).([]byte), &payload)
	if err != nil {
		fmt.Println(err)
		return
	}

	user := baseCtx.Value(utilities.KeyUser{}).(*models.User)

	if payload.Energy <= 0 {
		user.SendError(models.SendErrorParams{Context: &ctx, Type: "explore", Message: "explore-1"})
		return
	}
	if payload.Energy > user.RoundData.Energy {
		span.SetAttributes(attribute.Int("Current", user.RoundData.Energy), attribute.Int("Tried", payload.Energy))
		user.SendError(models.SendErrorParams{Context: &ctx, Type: "explore", Message: "explore-1"})
		return
	}

	//Run the calculations for land increase
	land := user.RoundData.Land
	energy := payload.Energy
	gain := 0.0
	increase := 0.0

	rand := rand.New((rand.NewSource(time.Now().UnixNano())))
	for i := 0; i < energy; i++ {
		switch {
		case land <= 100:
			gain = rand.Float64()*10 + 3
		case land <= 200:
			gain = rand.Float64()*7.5 + 2
		case land <= 400:
			gain = rand.Float64()*5 + 1.5
		case land <= 800:
			gain = rand.Float64()*2.5 + 1
		case land <= 1300:
			gain = rand.Float64()*1 + .3
		case land <= 1800:
			gain = rand.Float64()*.5 + .2
		case land <= 2500:
			gain = rand.Float64()*.25 + .1
		default:
			gain = rand.Float64()*.15 + .05
		}

		increase += gain
	}

	before := math.Floor(user.RoundData.Land)
	user.RoundData.Land += increase
	user.RoundData.FreeLand += increase
	user.RoundData.Energy -= energy

	if err := user.UpdateRound(ctx, nil); !err {
		span.RecordError(errors.New("error updating user"))
		user.SendMessage(payloads.Response{
			Type: "ERROR",
			Data: []byte("{\"type\": \"explore\", \"message\": \"Error exploring\"}"),
		})
		user.SendError(models.SendErrorParams{Context: &ctx, Type: "explore", Message: "explore-0"})
		go user.Log("Error Exploring", user.RoundData.ID)
		user.Load()
	} else {
		user.SendMessage(payloads.Response{Type: "EXPLORE_SUCCESS", Data: []byte(
			`{
			"message":"You explored",
			"gains": {
				"land":` + fmt.Sprint(math.Floor(user.RoundData.Land-before)) + `
			},
			"spent": {
				"energy":` + fmt.Sprint(energy) + `
			},
			"user":{
				"energy":` + fmt.Sprint(user.RoundData.Energy) + `,
				"landFree":` + fmt.Sprint(user.RoundData.FreeLand) + `,
				"land":` + fmt.Sprint(user.RoundData.Land) + `
			}
		}`,
		)})
		go user.Log("Spent: "+strconv.Itoa(energy)+" energy -- Found: "+strconv.FormatFloat(increase, 'f', 2, 64)+" acres", user.RoundData.ID)
		//go user.UpdateRank()
	}
}
