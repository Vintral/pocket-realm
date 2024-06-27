package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"

	"github.com/Vintral/pocket-realm/game/models"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/sdk/trace"
	span "go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

var round *models.Round
var db *gorm.DB
var sp span.Span
var ctx context.Context

func seedUsers(dbase *gorm.DB, tp *trace.TracerProvider, numUsers int) {
	ctx, sp = tp.Tracer("").Start(context.Background(), "seed-users")
	defer sp.End()

	db = dbase

	log.Warn().Msg("seedUsers: " + fmt.Sprint(numUsers))

	db.WithContext(ctx).First(&round)

	wg := new(sync.WaitGroup)
	wg.Add(numUsers)
	for i := 1; i <= numUsers; i++ {
		createUser(i, wg)
	}
	wg.Wait()

	log.Info().Msg("Done Seeding Users")
}

func createUser(id int, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Info().Msg("Seeding Bot: " + fmt.Sprint(id))

	user := &models.User{
		Email:        "email-" + fmt.Sprint(id) + "@email.com",
		Admin:        false,
		Username:     "user-" + fmt.Sprint(id),
		Avatar:       "1",
		RoundID:      int(round.ID),
		RoundPlaying: round.GUID,
	}
	db.Create(user)

	units := 15
	buildings := 30

	db.Create(&models.UserRound{
		UserID:         user.ID,
		RoundID:        round.ID,
		CharacterClass: "mage",
		Energy:         int(round.EnergyMax),
		Gold:           200,
		Housing:        5,
		Population:     5,
		Food:           200,
		Wood:           200,
		Metal:          200,
		Faith:          200,
		Stone:          200,
		Mana:           200,
		Land:           200,
		FreeLand:       200,
		BuildPower:     1,
		RecruitPower:   1,
	})

	for _, u := range round.Units {
		quantity := float64(units + rand.Intn(units) - units/2)
		// quantity = 1

		db.Create(&models.UserUnit{
			UserID:   user.ID,
			UnitID:   u.ID,
			RoundID:  round.ID,
			Quantity: quantity,
		})
	}

	for _, b := range round.Buildings {
		quantity := float64(buildings + rand.Intn(buildings) - buildings/2)

		fmt.Println(b.ID)
		if b.ID == 1 {
			quantity *= 5.5
		}

		if b.ID == 8 {
			quantity *= 2.7
		}

		db.Create(&models.UserBuilding{
			UserID:     user.ID,
			BuildingID: b.ID,
			RoundID:    round.ID,
			Quantity:   quantity,
		})
	}

	db.WithContext(ctx).First(&user)
	db.WithContext(ctx).Save(&user)

	log.Info().Msg("Housing: " + fmt.Sprint(user.RoundData.Housing) + " ::: Population: " + fmt.Sprint(user.RoundData.Population))
	user.RoundData.Population = user.RoundData.Housing
	db.WithContext(ctx).Save(&user)

	log.Warn().Msg("User Id: " + fmt.Sprint(user.ID))
	log.Warn().Msg("Housing: " + fmt.Sprint(user.RoundData.Housing) + " ::: Population: " + fmt.Sprint(user.RoundData.Population))
}
