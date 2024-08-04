package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"strconv"

	"github.com/Vintral/pocket-realm/game/payloads"
	"github.com/Vintral/pocket-realm/game/utilities"
	"github.com/Vintral/pocket-realm/models"
)

func gatherStat(energy int, before float64, tick float64) (float64, float64, float64) {
	increase := float64(rand.Intn(50)+75) / 100.0 * float64(energy)
	fmt.Println("Increase:", increase)

	increase *= math.Max(1, tick)
	after := before + increase

	return after, math.Floor(after) - math.Floor(before), after - before
}

func Gather(base context.Context) {
	fmt.Println("Gather")

	fmt.Println(base)
	fmt.Println("TraceProvider:", base.Value(utilities.KeyTraceProvider{}))
	fmt.Println("User:", base.Value(utilities.KeyUser{}))
	fmt.Println("Packet:", base.Value(utilities.KeyPayload{}))

	ctx, span := utilities.StartSpan(base, "gather")
	defer span.End()

	user := base.Value(utilities.KeyUser{}).(*models.User)

	var payload payloads.GatherPayload
	err := json.Unmarshal(base.Value(utilities.KeyPayload{}).([]byte), &payload)
	if err != nil {
		fmt.Println(err)
		user.SendError(models.SendErrorParams{Context: &ctx, Type: "gather", Message: "gather-1"})
		return
	}

	if payload.Energy <= 0 {
		fmt.Println("Invalid Energy Value")
		user.SendError(models.SendErrorParams{Context: &ctx, Type: "gather", Message: "gather-2"})
		return
	}
	if payload.Energy > user.RoundData.Energy {
		fmt.Println("Not Enough Energy")
		user.SendError(models.SendErrorParams{Context: &ctx, Type: "gather", Message: "gather-3"})
		return
	}

	round, err := models.LoadRoundById(base, user.RoundID)
	if err != nil {
		fmt.Println("Error Loading Round")
		user.SendError(models.SendErrorParams{Context: &ctx, Type: "gather", Message: "gather-0"})
	}

	resource := round.GetResourceByGuid(payload.Resource)
	if resource == nil || !resource.CanGather {
		user.SendError(models.SendErrorParams{Context: &ctx, Type: "gather", Message: "gather-0"})
		return
	} else {
		fmt.Println("CAN GATHER")
	}
	// if err := round.LoadByGuid(ctx, user.RoundPlaying); err != nil {
	// 	fmt.Println("Error Loading Round:", user.RoundPlaying)
	// 	user.SendMessage(payloads.Response{Type: "ERROR", Data: []byte("{\"type\": \"gather\", \"message\": \"gather-0\"}")})
	// 	return
	// }

	increase := 0.0
	updated := 0.0
	diff := 0.0
	switch resource.Name {
	case "gold":
		user.RoundData.Gold, increase, diff = gatherStat(payload.Energy, user.RoundData.Gold, user.RoundData.TickGold)
		updated = user.RoundData.Gold
	case "food":
		user.RoundData.Food, increase, diff = gatherStat(payload.Energy, user.RoundData.Food, user.RoundData.TickFood)
		updated = user.RoundData.Food
	case "wood":
		user.RoundData.Wood, increase, diff = gatherStat(payload.Energy, user.RoundData.Wood, user.RoundData.TickWood)
		updated = user.RoundData.Wood
	case "stone":
		user.RoundData.Stone, increase, diff = gatherStat(payload.Energy, user.RoundData.Stone, user.RoundData.TickStone)
		updated = user.RoundData.Stone
	case "metal":
		user.RoundData.Metal, increase, diff = gatherStat(payload.Energy, user.RoundData.Metal, user.RoundData.TickMetal)
		updated = user.RoundData.Metal
	case "mana":
		user.RoundData.Mana, increase, diff = gatherStat(payload.Energy, user.RoundData.Mana, user.RoundData.TickMana)
		updated = user.RoundData.Mana
	case "faith":
		user.RoundData.Faith, increase, diff = gatherStat(payload.Energy, user.RoundData.Faith, user.RoundData.TickFaith)
		updated = user.RoundData.Faith
	}

	user.RoundData.Energy -= payload.Energy

	if err := user.UpdateRound(ctx, nil); !err {
		user.SendMessage(payloads.Response{
			Type: "ERROR",
			Data: []byte("{\"type\": \"gather\", \"message\": \"gather-4\"}"),
		})
		go user.Log("Error Gathering", user.RoundData.ID)
		user.Load()
	} else {
		json, _ := json.Marshal(user)
		fmt.Println("User JSON:", json)
		user.SendMessage(payloads.Response{Type: "GATHER_SUCCESS", Data: []byte(
			`{			
					"gains": {
						"` + payload.Resource + `":` + fmt.Sprint(increase) + `
					},
					"spent": {
						"energy":` + fmt.Sprint(payload.Energy) + `
					},
					"user": {
						"energy":` + fmt.Sprint(user.RoundData.Energy) + `,
						"` + resource.Name + `": ` + fmt.Sprint(updated) + `
					}
				}`,
		)})

		go user.Log("Spent: "+strconv.Itoa(payload.Energy)+" energy gathering "+resource.Name+" Found: "+strconv.FormatFloat(diff, 'f', 2, 64), user.RoundData.ID)
	}
}
