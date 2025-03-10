package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	models "github.com/Vintral/pocket-realm/models"
	realmRedis "github.com/Vintral/pocket-realm/redis"
	"github.com/google/uuid"
	redisDef "github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

type BuffInfo struct {
	Name     string
	Field    string
	Category string
	Item     uint
	Bonus    float64
	Percent  bool
}

func main() {
	setupLogs()

	fmt.Println("Loading Environment")
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	//==============================//
	//	Setup Telemetry							//
	//==============================//
	fmt.Println("Setting up telemetry")
	otelShutdown, tp, err := setupOTelSDK(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}

	defer func() {
		fmt.Println("In shutdown")
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	fmt.Println("Setting Trace Provider")
	models.SetTracerProvider(tp)

	fmt.Println("Setting up database")
	db, err := models.Database(false, nil)
	if err != nil {
		time.Sleep(3 * time.Second)
		panic(err)
	}

	if len(os.Args) > 1 && os.Args[1] == "users" {
		numUsers := 0

		if numUsers, err = strconv.Atoi(os.Args[2]); err != nil {
			numUsers = 100
		}

		seedUsers(db, tp, numUsers)
		log.Info().Msg("Done seeding users")
	} else {
		dropTables(db)
		runMigrations(db)

		round, finished := createRounds(db)
		item1, item2 := createItems(db)
		createUsers(db, round, item1, item2)
		createUserTables(db, round)
		createContacts(db)

		var user *models.User
		db.First(&user)
		log.Info().Any("user", user).Msg("Loaded User")
		models.GetUndergroundMarketAuctions(context.Background(), user)

		createRules(db)
		createNews(db)
		unit := createUnits(db)
		createBuildings(db)
		createResources(db)
		createBuffs(db, user)
		createTechnologies(db, user)

		createTemple(db)

		createOverrides(db)

		createShouts(db)
		createConversations(db)
		createEvents(db, round)
		createRankings(db, finished, round)
		createResourceMarket(db, round)
		createBlackMarket(db)
		createMercenaryMarket(db, unit, round)
	}

	log.Info().Msg("Done Seeding")
}

func createUserTables(db *gorm.DB, round *models.Round) {
	//================================//
	// Users Units										//
	//================================//
	fmt.Println("Seeding User's Units")
	// db.Create(&models.UserUnit{
	// 	UserID:   1,
	// 	UnitID:   1,
	// 	RoundID:  1,
	// 	Quantity: 15,
	// })
	// db.Create(&models.UserUnit{
	// 	UserID:   1,
	// 	UnitID:   2,
	// 	RoundID:  1,
	// 	Quantity: 20,
	// })
	// db.Create(&models.UserUnit{
	// 	UserID:   1,
	// 	UnitID:   3,
	// 	RoundID:  1,
	// 	Quantity: 30,
	// })
	// db.Create(&models.UserUnit{
	// 	UserID:   1,
	// 	UnitID:   4,
	// 	RoundID:  1,
	// 	Quantity: 40,
	// })
	// db.Create(&models.UserUnit{
	// 	UserID:   1,
	// 	UnitID:   5,
	// 	RoundID:  2,
	// 	Quantity: 45,
	// })

	//================================//
	// User Buildings									//
	//================================//
	fmt.Println("Seeding User's Buildings")
	// db.Create(&models.UserBuilding{
	// 	UserID:     1,
	// 	BuildingID: 1,
	// 	RoundID:    1,
	// 	Quantity:   10,
	// })
	// db.Create(&models.UserBuilding{
	// 	UserID:     1,
	// 	BuildingID: 2,
	// 	RoundID:    1,
	// 	Quantity:   10,
	// })
	// db.Create(&models.UserBuilding{
	// 	UserID:     1,
	// 	BuildingID: 3,
	// 	RoundID:    1,
	// 	Quantity:   10,
	// })
	// db.Create(&models.UserBuilding{
	// 	UserID:     1,
	// 	BuildingID: 4,
	// 	RoundID:    1,
	// 	Quantity:   10,
	// })
	// db.Create(&models.UserBuilding{
	// 	UserID:     1,
	// 	BuildingID: 5,
	// 	RoundID:    1,
	// 	Quantity:   10,
	// })
	// db.Create(&models.UserBuilding{
	// 	UserID:     1,
	// 	BuildingID: 6,
	// 	RoundID:    1,
	// 	Quantity:   10,
	// })

	//================================//
	// User Items											//
	//================================//
	fmt.Println("Seeding User's Items")
	// db.Create(&models.UserItem{
	// 	UserID: 1,
	// 	ItemID: 1,
	// })

	//================================//
	// Users Rounds										//
	//================================//
	fmt.Println("Seeding User's Round")

	db.Create(&models.UserRound{
		UserID:         1,
		RoundID:        1,
		CharacterClass: "mage",
		Energy:         int(round.EnergyMax),
		Gold:           10,
		TickGold:       0,
		Housing:        5,
		Population:     5,
		Food:           15,
		TickFood:       0,
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

	db.Create(&models.UserRound{
		UserID:         1,
		RoundID:        2,
		CharacterClass: "warlord",
		Energy:         int(round.EnergyMax),
		Gold:           1,
		TickGold:       1,
		Housing:        1,
		Population:     1,
		Food:           1,
		TickFood:       1,
		Wood:           1,
		Metal:          1,
		Faith:          1,
		Stone:          1,
		Mana:           1,
		Land:           1,
		FreeLand:       1,
		BuildPower:     1,
		RecruitPower:   1,
	})

	db.Create(&models.UserRound{
		UserID:         2,
		RoundID:        2,
		CharacterClass: "merchant",
		Energy:         int(round.EnergyMax),
		Gold:           1,
		TickGold:       1,
		Housing:        1,
		Population:     1,
		Food:           1,
		TickFood:       1,
		Wood:           1,
		Metal:          1,
		Faith:          1,
		Stone:          1,
		Mana:           1,
		Land:           1,
		FreeLand:       1,
		BuildPower:     1,
		RecruitPower:   1,
	})

	db.Create(&models.UserRound{
		UserID:         3,
		RoundID:        2,
		CharacterClass: "thief",
		Energy:         int(round.EnergyMax),
		Gold:           1,
		TickGold:       1,
		Housing:        1,
		Population:     1,
		Food:           1,
		TickFood:       1,
		Wood:           1,
		Metal:          1,
		Faith:          1,
		Stone:          1,
		Mana:           1,
		Land:           1,
		FreeLand:       1,
		BuildPower:     1,
		RecruitPower:   1,
	})

	db.Create(&models.UserRound{
		UserID:         4,
		RoundID:        2,
		CharacterClass: "priest",
		Energy:         int(round.EnergyMax),
		Gold:           1,
		TickGold:       1,
		Housing:        1,
		Population:     1,
		Food:           1,
		TickFood:       1,
		Wood:           1,
		Metal:          1,
		Faith:          1,
		Stone:          1,
		Mana:           1,
		Land:           1,
		FreeLand:       1,
		BuildPower:     1,
		RecruitPower:   1,
	})

	// db.Create(&models.UserRound{
	// 	UserID:         1,
	// 	RoundID:        2,
	// 	CharacterClass: "mage",
	// 	Energy:         int(round.EnergyMax),
	// 	Gold:           200,
	// 	TickGold:       5,
	// 	Food:           200,
	// 	TickFood:       5,
	// 	Wood:           200,
	// 	TickWood:       5,
	// 	Metal:          200,
	// 	TickMetal:      5,
	// 	Faith:          200,
	// 	TickFaith:      5,
	// 	Stone:          200,
	// 	TickStone:      5,
	// 	Mana:           200,
	// 	TickMana:       5,
	// 	Land:           200,
	// 	FreeLand:       200,
	// 	BuildPower:     25,
	// 	RecruitPower:   25,
	// })
	// db.Create(&models.UserRound{
	// 	UserID:         3,
	// 	RoundID:        1,
	// 	CharacterClass: "priest",
	// })
	// db.Create(&models.UserRound{
	// 	UserID:         4,
	// 	RoundID:        1,
	// 	CharacterClass: "warlord",
	// })
	// db.Create(&models.UserRound{
	// 	UserID:         5,
	// 	RoundID:        1,
	// 	CharacterClass: "necromancer",
	// })
	// db.Create(&models.UserRound{
	// 	UserID:         6,
	// 	RoundID:        1,
	// 	CharacterClass: "merchant",
	// })
	// db.Create(&models.UserRound{
	// 	UserID:         7,
	// 	RoundID:        1,
	// 	CharacterClass: "druid",
	// })
}

func createTechnologies(db *gorm.DB, user *models.User) {
	log.Info().Msg("Creating Technologies")

	costs := [...]uint{250, 500, 1000, 2500}
	fields := [...]string{"gold", "food", "research", "metal", "wood", "faith", "stone"}
	for i, field := range fields {
		f := cases.Title(language.English).String(field)

		technology := &models.Technology{
			Name:        fmt.Sprintf("Improved %s", f),
			Description: fmt.Sprintf("+25%% %s Tick", f),
			Buff:        uint(i + 1),
		}
		db.Create(technology)

		db.Create(&models.RoundTechnology{
			RoundID:      0,
			TechnologyID: technology.ID,
			Available:    true,
		})

		for n, cost := range costs {
			db.Create(&models.TechnologyLevel{
				Technology: technology.ID,
				Level:      uint(n + 1),
				Cost:       cost,
			})
		}
	}

	db.Create(&models.RoundTechnology{
		RoundID:      1,
		TechnologyID: 7,
		Available:    false,
	})

	db.Create(&models.UserTechnology{
		RoundID:      1,
		UserID:       1,
		TechnologyID: 1,
		Level:        2,
	})
}

func createBuffs(db *gorm.DB, user *models.User) {
	log.Info().Msg("Creating Buffs")

	fields := [...]string{"gold", "food", "research", "metal", "wood", "faith", "stone"}
	maxStacks := 4

	var goldBuff *models.Buff
	for i, field := range fields {
		buff :=
			&models.Buff{
				Name:      fmt.Sprintf("%s Production Percent Buff", cases.Title(language.English).String(field)),
				Field:     fmt.Sprintf("%s_tick", field),
				Bonus:     float64(100 / maxStacks),
				Type:      "player",
				MaxStacks: uint(maxStacks),
				Percent:   true,
			}
		db.Create(buff)

		if i == 0 {
			goldBuff = buff
		}
	}

	ctx := context.Background()
	user.AddBuff(ctx, goldBuff)
	user.AddBuff(ctx, goldBuff)
	user.AddBuff(ctx, goldBuff)
	user.AddBuff(ctx, goldBuff)
	user.AddBuff(ctx, goldBuff)
	db.WithContext(ctx).Save(&user)
}

func createRounds(db *gorm.DB) (current *models.Round, finished *models.Round) {
	//================================//
	// Rounds													//
	//================================//
	fmt.Println("Seeding Round")
	round := &models.Round{
		EnergyMax:   250,
		EnergyRegen: 10,
		Tick:        1,
		Starts:      time.Now(),
		Ends:        time.Now().Add(14 * 24 * time.Hour),
		StartLand:   100,
	}
	db.Create(round)
	ret := round

	round = &models.Round{
		EnergyMax:   250,
		EnergyRegen: 10,
		Tick:        10,
		Starts:      time.Now().Add(7 * 24 * time.Hour),
		Ends:        time.Now().Add(14 * 24 * time.Hour),
		StartLand:   10,
	}
	db.Create(round)

	round = &models.Round{
		EnergyMax:   250,
		EnergyRegen: 10,
		Tick:        10,
		Starts:      time.Now().Add(-21 * 24 * time.Hour),
		Ends:        time.Now().Add(-14 * 24 * time.Hour),
		StartLand:   10,
	}
	db.Create(round)

	return ret, round
}

func createUsers(db *gorm.DB, r *models.Round, i1 *models.Item, i2 *models.Item) {
	//================================//
	// Users													//
	//================================//
	fmt.Println("Seeding users")

	var round models.Round
	db.Model(models.Round{GUID: r.GUID}).First(&round)

	user := &models.User{
		Email:        "jeffrey.heater@gmail.com",
		Admin:        true,
		Username:     "Vintral",
		Avatar:       "m1",
		RoundID:      int(round.ID),
		RoundPlaying: round.GUID,
	}
	db.FirstOrCreate(&user)
	// user.Join(context.Background(), &round)
	user.AddItem(ctx, i1)
	user.AddItem(ctx, i1)
	user.AddItem(ctx, i2)

	user = &models.User{
		Email:    "jeffrey.heater0@gmail.com",
		Admin:    true,
		Username: "Trilanni",
		Avatar:   "f2",
	}
	db.Create(&user)
	user.Join(context.Background(), &round, "mage")

	user = &models.User{
		Email:        "jeffrey.heater1@gmail.com",
		Admin:        true,
		Username:     "Vintral1",
		Avatar:       "m3",
		RoundID:      1,
		RoundPlaying: round.GUID,
	}
	db.Create(&user)
	user.Join(context.Background(), &round, "priest")

	user = &models.User{
		Email:        "jeffrey.heater2@gmail.com",
		Admin:        true,
		Username:     "Vintral2",
		Avatar:       "f4",
		RoundID:      1,
		RoundPlaying: round.GUID,
	}
	db.Create(&user)
	user.Join(context.Background(), &round, "merchant")

	user = &models.User{
		Email:        "jeffrey.heater3@gmail.com",
		Admin:        true,
		Username:     "Vintral3",
		Avatar:       "m5",
		RoundID:      1,
		RoundPlaying: round.GUID,
	}
	db.Create(&user)
	user.Join(context.Background(), &round, "warlord")

	user = &models.User{
		Email:        "jeffrey.heater4@gmail.com",
		Admin:        true,
		Username:     "Vintral4",
		Avatar:       "f6",
		RoundID:      1,
		RoundPlaying: round.GUID,
	}
	db.Create(&user)
	user.Join(context.Background(), &round, "thief")

	user = &models.User{
		Email:        "jeffrey.heater5@gmail.com",
		Admin:        true,
		Username:     "Vintral5",
		Avatar:       "m1",
		RoundID:      1,
		RoundPlaying: round.GUID,
	}
	db.Create(&user)
	user.Join(context.Background(), &round, "mage")
}

func createShouts(db *gorm.DB) {
	//================================//
	// Shouts													//
	//================================//
	fmt.Println("Seeding shouts")
	db.Create(&models.Shout{
		UserID: 1,
		Shout:  "Mage Shout",
	})
	db.Create(&models.Shout{
		UserID: 2,
		Shout:  "Not Playing Shout",
	})
	db.Create(&models.Shout{
		UserID: 3,
		Shout:  "Priest Shout",
	})
	db.Create(&models.Shout{
		UserID: 4,
		Shout:  "Warlord Shout",
	})
	db.Create(&models.Shout{
		UserID: 5,
		Shout:  "Necromacer Shout",
	})
	db.Create(&models.Shout{
		UserID: 6,
		Shout:  "Merchant shout",
	})
	db.Create(&models.Shout{
		UserID: 7,
		Shout:  "Druid shout",
	})
	db.Create(&models.Shout{
		UserID: 1,
		Shout:  "Mage shout",
	})
	db.Create(&models.Shout{
		UserID: 2,
		Shout:  "Not Playing shout",
	})
	db.Create(&models.Shout{
		UserID: 3,
		Shout:  "Priest shout",
	})
}

func createRankings(db *gorm.DB, round *models.Round, current *models.Round) {
	log.Info().Msg("createRankings")

	var users []*models.User
	result := db.Order("id desc").Find(&users)
	log.Info().Msg("Users: " + fmt.Sprint(result.RowsAffected))

	redis, err := realmRedis.Instance(nil)
	if err != nil {
		log.Panic().AnErr("err", err).Msg("Error getting redis instance")
	}

	log.Warn().Msg("Redis loaded")
	if redis == nil {
		log.Panic().Msg("Redis instance is nil")
	}

	for i, user := range users {
		log.Info().Msg("User: " + user.Username + " -- " + fmt.Sprint(user.RoundID) + " -- " + fmt.Sprint(user.ID))

		land := (uint(result.RowsAffected) - uint(i)) * 25
		power := land * 10

		db.Create(&models.Ranking{
			UserID:  user.ID,
			RoundID: round.ID,
			Rank:    uint(i) + 1,
			Score:   power,
		})

		result := redis.ZAdd(
			context.Background(),
			fmt.Sprint(current.ID)+"-rankings",
			redisDef.Z{Score: float64(land * 10), Member: user.ID},
		)
		if result.Err() != nil {
			log.Warn().AnErr("err", result.Err()).Msg("Error adding ranking")
		}

		// if err := redis.Set(
		// 	context.Background(),
		// 	fmt.Sprint(current.ID)+"-snapshot-"+fmt.Sprint(user.ID),
		// 	&models.RankingSnapshot{Username: user.Username, Score: math.Floor(float64(power))},
		// 	0,
		// ).Err(); err != nil {
		// 	log.Warn().AnErr("err", err).Msg("Error updating redis snapshot")
		// }
	}
}

func createEvents(db *gorm.DB, round *models.Round) {
	fmt.Println("Seeding Events")

	db.Create(&models.Event{
		UserID: 1,
		Round:  round.GUID,
		Event:  "Test Event Round 1",
	})
	db.Create(&models.Event{
		UserID: 1,
		Round:  round.GUID,
		Event:  "Test Event 2 Round 1",
	})
	db.Create(&models.Event{
		UserID: 1,
		Round:  uuid.Nil,
		Event:  "Test Event Account",
	})
}

func createBlackMarket(db *gorm.DB) {
	log.Info().Msg("createBlackMarket")

	db.Create(&models.UndergroundMarketAuction{
		ItemID:  1,
		Cost:    50,
		Expires: time.Now().AddDate(0, 0, -1),
	})

	db.Create(&models.UndergroundMarketAuction{
		ItemID:  1,
		Cost:    50,
		Expires: time.Now().AddDate(0, 0, 3),
	})

	db.Create(&models.UndergroundMarketAuction{
		ItemID:  1,
		Cost:    50,
		Expires: time.Now().Add(time.Hour * 5), // * time.Duration(5)),
	})

	db.Create(&models.UndergroundMarketPurchase{
		MarketID:  2,
		UserID:    1,
		Purchased: time.Now(),
	})
}

func createResourceMarket(db *gorm.DB, round *models.Round) {
	log.Info().Int("round", int(round.ID)).Msg("createResourceMarket")

	vals := [...]uint{2, 3, 4, 5}

	for _, resource := range vals {
		db.Create(&models.RoundMarketResource{
			RoundID:    round.ID,
			ResourceID: resource,
			Value:      2,
		})
	}
}

func createMercenaryMarket(db *gorm.DB, unit *models.RoundUnit, round *models.Round) {
	log.Info().Msg("createMercenaryMarket")

	db.Create(&models.MercenaryMarket{
		Cost:    5,
		Unit:    unit.GUID,
		Round:   round.ID,
		Expires: time.Now().AddDate(0, 0, 2),
	})
}

func createContacts(db *gorm.DB) {
	log.Info().Msg("createContacts")

	db.Create(&models.Contact{
		ContactID: 2,
		UserID:    1,
		Category:  "friend",
		Note:      "Test note for friend",
	})

	db.Create(&models.Contact{
		ContactID: 3,
		UserID:    1,
		Category:  "enemy",
		Note:      "They attacked ME!!",
	})
}

func dropTables(db *gorm.DB) {
	db.Exec("DROP TABLE user_units")
	db.Exec("DROP TABLE user_rounds")
	db.Exec("DROP TABLE user_buildings")
	db.Exec("DROP TABLE user_items")
	db.Exec("DROP TABLE round_resources")
	db.Exec("DROP TABLE round_buildings")
	db.Exec("DROP TABLE round_market_resources")
	db.Exec("DROP TABLE round_units")
	db.Exec("DROP TABLE units")
	db.Exec("DROP TABLE users")
	db.Exec("DROP TABLE buildings")
	db.Exec("DROP TABLE effects")
	db.Exec("DROP TABLE items")
	db.Exec("DROP TABLE rounds")
	db.Exec("DROP TABLE resources")
	db.Exec("DROP TABLE news_items")
	db.Exec("DROP TABLE rules")
	db.Exec("DROP TABLE shouts")
	db.Exec("DROP TABLE user_logs")
	db.Exec("DROP TABLE conversations")
	db.Exec("DROP TABLE messages")
	db.Exec("DROP TABLE events")
	db.Exec("DROP TABLE rankings")
	db.Exec("DROP TABLE underground_market_purchases")
	db.Exec("DROP TABLE underground_market_auctions")
	db.Exec("DROP TABLE mercenary_markets")
	db.Exec("DROP TABLE buffs")
	db.Exec("DROP TABLE user_buffs")
	db.Exec("DROP TABLE technologies")
	db.Exec("DROP TABLE technology_levels")
	db.Exec("DROP TABLE round_technologies")
	db.Exec("DROP TABLE user_technologies")
	db.Exec("DROP TABLE contacts")
	db.Exec("DROP TABLE pantheons")
	db.Exec("DROP TABLE devotions")
	db.Exec("DROP TABLE user_devotions")
}

func createConversations(db *gorm.DB) {
	fmt.Println("Seeding conversations")

	conversation := &models.Conversation{
		User1ID:       1,
		User2ID:       2,
		User2LastRead: time.Now(),
	}
	db.Create(conversation)

	for i := 0; i < 15; i++ {
		db.Create(&models.Message{
			Conversation: conversation.ID,
			UserID:       1 + uint(i%2),
			Text:         "Message should show",
		})
	}

	conversation = &models.Conversation{
		User1ID:       2,
		User2ID:       3,
		User2LastRead: time.Now(),
	}
	db.Create(conversation)

	for i := 0; i < 15; i++ {
		db.Create(&models.Message{
			Conversation: conversation.ID,
			UserID:       1 + uint(i%2),
			Text:         "Message should not show",
		})
	}
}

func runMigrations(db *gorm.DB) {
	models.RunMigrations(db)
}

func createOverrides(db *gorm.DB) {
	db.Create(&models.RoundUnit{
		RoundID:     1,
		UnitID:      1,
		Attack:      1.00,
		Defense:     1.00,
		Power:       1.00,
		Health:      5,
		Ranged:      false,
		CostGold:    1,
		CostPoints:  1,
		CostFood:    1,
		CostWood:    1,
		CostMetal:   1,
		CostStone:   1,
		CostFaith:   1,
		CostMana:    1,
		UpkeepGold:  1,
		UpkeepFood:  1,
		UpkeepWood:  1,
		UpkeepStone: 1,
		UpkeepMetal: 1,
		UpkeepFaith: 1,
		UpkeepMana:  1,
		Available:   true,
		Recruitable: true,
		StartWith:   5,
	})

	db.Create(&models.RoundResource{RoundID: 1, ResourceID: 6, StartWith: 400, CanGather: false, CanMarket: false})
	db.Create(&models.RoundResource{RoundID: 1, ResourceID: 7, StartWith: 400, CanGather: false, CanMarket: false})

	db.Create(&models.RoundResource{RoundID: 2, ResourceID: 7, StartWith: 400, CanGather: false, CanMarket: false})

	db.Create(&models.RoundBuilding{
		BuildingID:  1,
		RoundID:     1,
		CostPoints:  1,
		CostWood:    1,
		CostStone:   1,
		CostGold:    1,
		CostFood:    1,
		CostMetal:   1,
		CostFaith:   1,
		CostMana:    1,
		BonusValue:  1,
		UpkeepGold:  1,
		UpkeepFood:  1,
		UpkeepWood:  1,
		UpkeepStone: 1,
		UpkeepMetal: 1,
		UpkeepFaith: 1,
		UpkeepMana:  1,
		Buildable:   true,
		Available:   true,
		StartWith:   5,
	})

	db.Create(&models.RoundBuilding{
		BuildingID:  1,
		RoundID:     2,
		CostPoints:  1,
		CostWood:    1,
		CostStone:   1,
		CostGold:    1,
		CostFood:    1,
		CostMetal:   1,
		CostFaith:   1,
		CostMana:    1,
		BonusValue:  1,
		UpkeepGold:  1,
		UpkeepFood:  1,
		UpkeepWood:  1,
		UpkeepStone: 1,
		UpkeepMetal: 1,
		UpkeepFaith: 1,
		UpkeepMana:  1,
		Buildable:   true,
		Available:   true,
		StartWith:   5,
	})
}

func createBuildings(db *gorm.DB) {
	//================================//
	// Buildings											//
	//================================//
	fmt.Println("Seeding Buildings")
	db.Create(&models.Building{Name: "farm", BonusField: "food_tick"})
	db.Create(&models.Building{Name: "barracks", BonusField: "recruit_power"})
	db.Create(&models.Building{Name: "lumber-mill", BonusField: "wood_tick"})
	db.Create(&models.Building{Name: "quarry", BonusField: "stone_tick"})
	db.Create(&models.Building{Name: "wall", BonusField: "defense"})
	db.Create(&models.Building{Name: "workshop", BonusField: "build_power"})
	db.Create(&models.Building{Name: "mine", BonusField: "metal_tick"})
	db.Create(&models.Building{Name: "house", BonusField: "housing"})
	db.Create(&models.Building{Name: "library", BonusField: "research_tick"})
	db.Create(&models.Building{Name: "shrine", BonusField: "faith_tick"})

	//================================//
	// Building Defaults							//
	//================================//
	fmt.Println("Seeding Building Defaults")
	db.Create(&models.RoundBuilding{
		BuildingID:      1,
		RoundID:         0,
		CostPoints:      1,
		CostWood:        1,
		BonusValue:      1,
		Available:       true,
		Buildable:       true,
		SupportsPartial: false,
		StartWith:       0,
	})
	db.Create(&models.RoundBuilding{
		BuildingID:      2,
		RoundID:         0,
		CostWood:        100,
		CostStone:       100,
		CostPoints:      10,
		BonusValue:      1,
		Available:       true,
		Buildable:       true,
		SupportsPartial: false,
		StartWith:       0,
	})
	db.Create(&models.RoundBuilding{
		BuildingID:      3,
		RoundID:         0,
		CostWood:        15,
		CostStone:       0,
		CostPoints:      2,
		BonusValue:      1,
		Available:       true,
		Buildable:       true,
		SupportsPartial: false,
		StartWith:       0,
	})
	db.Create(&models.RoundBuilding{
		BuildingID:      4,
		RoundID:         0,
		CostWood:        5,
		CostStone:       10,
		CostPoints:      2,
		BonusValue:      1,
		Available:       true,
		Buildable:       true,
		SupportsPartial: false,
		StartWith:       0,
	})
	db.Create(&models.RoundBuilding{
		BuildingID:      5,
		RoundID:         0,
		CostWood:        0,
		CostStone:       25,
		CostPoints:      2,
		BonusValue:      1,
		Available:       true,
		Buildable:       true,
		SupportsPartial: false,
		StartWith:       0,
	})
	db.Create(&models.RoundBuilding{
		BuildingID:      6,
		RoundID:         0,
		CostWood:        20,
		CostStone:       25,
		CostPoints:      2,
		BonusValue:      1,
		Available:       true,
		Buildable:       true,
		SupportsPartial: false,
		StartWith:       0,
	})
	db.Create(&models.RoundBuilding{
		BuildingID:      7,
		RoundID:         0,
		CostWood:        20,
		CostStone:       25,
		CostPoints:      2,
		BonusValue:      1,
		Available:       true,
		Buildable:       true,
		SupportsPartial: false,
		StartWith:       0,
	})
	db.Create(&models.RoundBuilding{
		BuildingID:      8,
		RoundID:         0,
		CostWood:        5,
		CostStone:       2,
		CostPoints:      2,
		BonusValue:      2,
		Available:       true,
		Buildable:       true,
		SupportsPartial: false,
		StartWith:       0,
	})
	db.Create(&models.RoundBuilding{
		BuildingID:      9,
		RoundID:         0,
		CostWood:        5,
		CostStone:       2,
		CostPoints:      2,
		BonusValue:      3,
		Available:       true,
		Buildable:       true,
		SupportsPartial: false,
		StartWith:       0,
	})
	db.Create(&models.RoundBuilding{
		BuildingID:      10,
		RoundID:         0,
		CostWood:        5,
		CostStone:       2,
		CostPoints:      2,
		BonusValue:      3,
		Available:       true,
		Buildable:       true,
		SupportsPartial: false,
		StartWith:       0,
	})
}

func createResources(db *gorm.DB) {
	//================================//
	// Resources											//
	//================================//
	fmt.Println("Seeding Resources")
	db.Create(&models.Resource{Name: "gold"})
	db.Create(&models.Resource{Name: "wood"})
	db.Create(&models.Resource{Name: "food"})
	db.Create(&models.Resource{Name: "stone"})
	db.Create(&models.Resource{Name: "metal"})
	db.Create(&models.Resource{Name: "faith"})
	db.Create(&models.Resource{Name: "mana"})
	db.Create(&models.Resource{Name: "research"})

	//================================//
	// Resource Defaults							//
	//================================//
	db.Create(&models.RoundResource{RoundID: 0, ResourceID: 1, StartWith: 200, CanGather: true, CanMarket: false})
	db.Create(&models.RoundResource{RoundID: 0, ResourceID: 2, StartWith: 200, CanGather: true, CanMarket: true})
	db.Create(&models.RoundResource{RoundID: 0, ResourceID: 3, StartWith: 200, CanGather: true, CanMarket: true})
	db.Create(&models.RoundResource{RoundID: 0, ResourceID: 4, StartWith: 200, CanGather: true, CanMarket: true})
	db.Create(&models.RoundResource{RoundID: 0, ResourceID: 5, StartWith: 200, CanGather: true, CanMarket: true})
	db.Create(&models.RoundResource{RoundID: 0, ResourceID: 6, StartWith: 200, CanGather: true, CanMarket: false})
	db.Create(&models.RoundResource{RoundID: 0, ResourceID: 7, StartWith: 200, CanGather: true, CanMarket: false})
	db.Create(&models.RoundResource{RoundID: 0, ResourceID: 8, StartWith: 0, CanGather: false, CanMarket: false})
}

func createDevotionBuff(db *gorm.DB, data BuffInfo) *models.Buff {
	buff := models.Buff{
		Name:      data.Name,
		Field:     data.Field,
		Bonus:     data.Bonus,
		Type:      data.Category,
		Item:      data.Item,
		MaxStacks: 1,
		Percent:   data.Percent,
	}
	db.Save(&buff)

	return &buff
}

func createLifePantheon(db *gorm.DB) {
	pantheon := &models.Pantheon{
		Category: "Life",
	}
	db.Create(pantheon)
	buff := createDevotionBuff(db, BuffInfo{
		Name: "+25% Population Growth", Field: "population_growth", Category: "player", Bonus: 25, Percent: true,
	})
	db.Create(&models.Devotion{
		Pantheon:        pantheon.ID,
		Level:           1,
		Upkeep:          25,
		Buff:            buff.ID,
		BuffDescription: buff.Name,
	})
	buff = createDevotionBuff(db, BuffInfo{
		Name: "+25% Population Growth", Field: "population_growth", Category: "player", Bonus: 25, Percent: true,
	})
	db.Create(&models.Devotion{
		Pantheon:        pantheon.ID,
		Level:           2,
		Upkeep:          50,
		Buff:            buff.ID,
		BuffDescription: buff.Name,
	})
	buff = createDevotionBuff(db, BuffInfo{
		Name: "+50% Population Growth", Field: "population_growth", Category: "player", Bonus: 50, Percent: true,
	})
	db.Create(&models.Devotion{
		Pantheon:        pantheon.ID,
		Level:           3,
		Upkeep:          100,
		Buff:            buff.ID,
		BuffDescription: buff.Name,
	})
}

func createWarPantheon(db *gorm.DB) {
	pantheon := models.Pantheon{
		Category: "War",
	}
	db.Create(&pantheon)

	buff := createDevotionBuff(db, BuffInfo{
		Name: "+5 Attack for Units", Field: "attack", Category: "unit", Bonus: 5,
	})
	db.Create(&models.Devotion{
		Pantheon:        pantheon.ID,
		Level:           1,
		Upkeep:          25,
		Buff:            buff.ID,
		BuffDescription: buff.Name,
	})
	buff = createDevotionBuff(db, BuffInfo{
		Name: "+5 Defense for Units", Field: "defense", Category: "unit", Bonus: 5,
	})
	db.Create(&models.Devotion{
		Pantheon:        pantheon.ID,
		Level:           2,
		Upkeep:          50,
		Buff:            buff.ID,
		BuffDescription: buff.Name,
	})
	buff = createDevotionBuff(db, BuffInfo{
		Name: "+5 Health for Units", Field: "health", Category: "unit", Bonus: 5,
	})
	db.Create(&models.Devotion{
		Pantheon:        pantheon.ID,
		Level:           3,
		Upkeep:          100,
		Buff:            buff.ID,
		BuffDescription: buff.Name,
	})
}

func createDeathPantheon(db *gorm.DB) {
	pantheon := models.Pantheon{
		Category: "Death",
	}
	db.Create(&pantheon)

	buff := createDevotionBuff(db, BuffInfo{
		Name: "-25% Population Growth", Field: "population_growth", Category: "player", Bonus: -25, Percent: true,
	})
	db.Create(&models.Devotion{
		Pantheon:        pantheon.ID,
		Level:           1,
		Upkeep:          25,
		Buff:            buff.ID,
		BuffDescription: buff.Name,
	})
	buff = createDevotionBuff(db, BuffInfo{
		Name: "-25% Population Growth", Field: "population_growth", Category: "player", Bonus: -25, Percent: true,
	})
	db.Create(&models.Devotion{
		Pantheon:        pantheon.ID,
		Level:           2,
		Upkeep:          50,
		Buff:            buff.ID,
		BuffDescription: buff.Name,
	})
	buff = createDevotionBuff(db, BuffInfo{
		Name: "-50% Population Growth", Field: "population_growth", Category: "player", Bonus: -50, Percent: true,
	})
	db.Create(&models.Devotion{
		Pantheon:        pantheon.ID,
		Level:           3,
		Upkeep:          100,
		Buff:            buff.ID,
		BuffDescription: buff.Name,
	})
}

func createEmpirePantheon(db *gorm.DB) {
	pantheon := models.Pantheon{
		Category: "Empire",
	}
	db.Create(&pantheon)

	buff := createDevotionBuff(db, BuffInfo{
		Name: "+5 Build Power", Field: "build_power", Category: "player", Bonus: 5,
	})
	db.Create(&models.Devotion{
		Pantheon:        pantheon.ID,
		Level:           1,
		Upkeep:          25,
		Buff:            buff.ID,
		BuffDescription: buff.Name,
	})
	buff = createDevotionBuff(db, BuffInfo{
		Name: "+50% Land Found Exploring", Field: "exploring_land_gain", Category: "player", Bonus: 50, Percent: true,
	})
	db.Create(&models.Devotion{
		Pantheon:        pantheon.ID,
		Level:           2,
		Upkeep:          50,
		Buff:            buff.ID,
		BuffDescription: buff.Name,
	})
	buff = createDevotionBuff(db, BuffInfo{
		Name: "+50% Gathering Returns", Field: "gathering_returns", Category: "player", Bonus: 50, Percent: true,
	})
	db.Create(&models.Devotion{
		Pantheon:        pantheon.ID,
		Level:           3,
		Upkeep:          100,
		Buff:            buff.ID,
		BuffDescription: buff.Name,
	})
}

func createTemple(db *gorm.DB) {
	log.Info().Msg("Creating Temple")

	createLifePantheon(db)
	createWarPantheon(db)
	createDeathPantheon(db)
	createEmpirePantheon(db)
}

func createUnits(db *gorm.DB) *models.RoundUnit {
	//================================//
	// Units													//
	//================================//
	fmt.Println("Seeding Units")
	db.Create(&models.Unit{Name: "peasant"})
	db.Create(&models.Unit{Name: "footman"})
	db.Create(&models.Unit{Name: "archer"})
	db.Create(&models.Unit{Name: "crusader"})
	db.Create(&models.Unit{Name: "cavalry"})

	//================================//
	// Unit Defaults  								//
	//================================//
	fmt.Println("Seeding Unit Defaults")
	db.Create(&models.RoundUnit{
		RoundID:         0,
		UnitID:          1,
		Attack:          1.00,
		Defense:         1.00,
		Power:           1.00,
		Health:          5,
		Ranged:          false,
		CostGold:        1,
		CostPoints:      1,
		CostFood:        1,
		UpkeepFood:      1,
		Available:       true,
		Recruitable:     true,
		SupportsPartial: false,
	})
	db.Create(&models.RoundUnit{
		RoundID:         0,
		UnitID:          2,
		Attack:          2.00,
		Defense:         2.00,
		Power:           2.00,
		Health:          15,
		Ranged:          false,
		CostGold:        2,
		CostPoints:      2,
		CostFood:        2,
		UpkeepGold:      1,
		UpkeepFood:      1,
		Available:       true,
		Recruitable:     true,
		SupportsPartial: false,
		StartWith:       0,
	})
	db.Create(&models.RoundUnit{
		RoundID:         0,
		UnitID:          3,
		Attack:          3.00,
		Defense:         1.00,
		Power:           3.00,
		Health:          15,
		Ranged:          true,
		CostGold:        5,
		CostPoints:      5,
		UpkeepGold:      2,
		UpkeepFood:      1,
		UpkeepWood:      1,
		Available:       true,
		Recruitable:     true,
		SupportsPartial: false,
		StartWith:       0,
	})
	db.Create(&models.RoundUnit{
		RoundID:         0,
		UnitID:          4,
		Attack:          5.00,
		Defense:         5.00,
		Power:           10.00,
		Health:          30,
		Ranged:          false,
		CostGold:        10,
		CostPoints:      10,
		UpkeepGold:      3,
		UpkeepFood:      2,
		UpkeepMetal:     2,
		Available:       true,
		Recruitable:     true,
		SupportsPartial: false,
		StartWith:       0,
	})

	unit := models.RoundUnit{
		RoundID:         0,
		UnitID:          5,
		Attack:          10.00,
		Defense:         5.00,
		Power:           20.00,
		Health:          50,
		Ranged:          false,
		CostGold:        25,
		CostPoints:      20,
		UpkeepGold:      5,
		UpkeepFood:      5,
		Available:       true,
		Recruitable:     true,
		SupportsPartial: false,
		StartWith:       0,
	}
	db.Create(&unit)

	return &unit
}

func createItems(db *gorm.DB) (item1 *models.Item, item2 *models.Item) {
	//================================//
	// Items													//
	//================================//
	fmt.Println("Seeding Items")

	item1 = &models.Item{
		Name:   "Small Hourglass",
		Plural: "Small Hourglasses",
	}
	db.Create(item1)
	db.Create(&models.Effect{
		ItemID: item1.ID,
		Type:   "resource",
		Name:   "energy",
		Amount: 10,
	})

	item2 = &models.Item{
		Name:   "Crate of Grain",
		Plural: "Crates of Grain",
	}
	db.Create(item2)
	db.Create(&models.Effect{
		ItemID: item2.ID,
		Type:   "resource",
		Name:   "food",
		Amount: 50,
	})

	return
}

func createRules(db *gorm.DB) {
	fmt.Println("Seeding Rules")

	db.Create(&models.Rule{
		Rule:   "rule-1",
		Active: true,
	})
	db.Create(&models.Rule{
		Rule:   "rule-2",
		Active: false,
	})
	db.Create(&models.Rule{
		Rule:   "rule-3",
		Active: true,
	})
}

func createNews(db *gorm.DB) {
	fmt.Println("Seeding news")

	db.Create(&models.NewsItem{
		Title:  "Test Title",
		Body:   "News body goes here",
		Active: true,
	})

	db.Create(&models.NewsItem{
		Title:  "Test Title 2",
		Body:   "News body goes here 2",
		Active: true,
	})
}
