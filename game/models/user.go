package models

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"sync"

	"github.com/Vintral/pocket-realm/game/payloads"
	"github.com/Vintral/pocket-realm/game/utilities"
	"github.com/rs/zerolog/log"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.opentelemetry.io/otel/codes"
	"gorm.io/gorm"
)

// var (
// 	tracer  = otel.Tracer("rolldice")
// 	meter   = otel.Meter("rolldice")
// 	rollCnt metric.Int64Counter
// )

type User struct {
	BaseModel

	GUID         uuid.UUID       `gorm:"size:36" json:"guid"`
	Username     string          `gorm:"uniqueIndex,size:32" json:"username"`
	Avatar       string          `json:"avatar"`
	Email        string          `gorm:"uniqueIndex,size:64" json:"-"`
	Password     string          `json:"-"`
	Admin        bool            `gorm:"default:false" json:"-"`
	RoundID      int             `json:"-"`
	RoundLoading int             `gorm:"-" json:"-"`
	RoundPlaying uuid.UUID       `gorm:"size:36" json:"round_playing"`
	RoundData    UserRound       `gorm:"-" json:"round"`
	Round        *Round          `gorm:"-" json:"-"`
	Units        []*UserUnit     `gorm:"-" json:"units"`
	Buildings    []*UserBuilding `gorm:"-" json:"buildings"`
	Items        []*UserItem     `gorm:"-" json:"items"`
	Context      context.Context `gorm:"-:all" json:"-"`
	Connection   *websocket.Conn `gorm:"-" json:"-"`
	DB           *gorm.DB        `gorm:"-" json:"-"`
}

func (user *User) BeforeCreate(tx *gorm.DB) (err error) {
	user.GUID = uuid.New()
	return
}

func getRound(user *User) int {
	if user.RoundLoading != 0 {
		return user.RoundLoading
	}

	return user.RoundID
}

func (user *User) AfterFind(tx *gorm.DB) (err error) {
	log.Trace().Msg("user:AfterFind")

	ctx, sp := Tracer.Start(tx.Statement.Context, "after-find")
	defer sp.End()

	roundId := getRound(user)

	if roundId != 0 {
		round, err := LoadRoundById(ctx, roundId)
		if err != nil {
			fmt.Println("Error Loading Round:", user.RoundID)
			return err
		}

		user.Round = round

		//user.loadSynchronous(ctx)
		wg := new(sync.WaitGroup)
		wg.Add(4)
		go user.loadRound(ctx, wg)
		go user.loadUnits(ctx, wg)
		go user.loadBuildings(ctx, wg)
		go user.loadItems(ctx, wg)
		wg.Wait()
	}

	return
}

func (user *User) AfterUpdate(tx *gorm.DB) (err error) {
	ctx, sp := Tracer.Start(tx.Statement.Context, "after-update")
	defer sp.End()

	wg := new(sync.WaitGroup)
	wg.Add(4)
	go user.UpdateRound(ctx, wg)
	go user.updateUnits(ctx, wg)
	go user.updateBuildings(ctx, wg)
	go user.updateItems(ctx, wg)
	wg.Wait()

	return
}

func (user *User) updateUnits(ctx context.Context, wg *sync.WaitGroup) bool {
	ctx, span := Tracer.Start(ctx, "user-update-units")
	defer span.End()
	if wg != nil {
		defer wg.Done()
	}

	fmt.Println("Saving Quantity:", user.Units[0].Quantity)
	if err := db.WithContext(ctx).Save(&user.Units).Error; err != nil {
		fmt.Println("Error updatingUnits")
		span.SetStatus(codes.Error, err.Error())
		return false
	}

	fmt.Println("Units Updated")

	return true
}

func (user *User) resetTicks() {
	user.RoundData.TickFaith = 0
	user.RoundData.TickFood = 0
	user.RoundData.TickGold = 0
	user.RoundData.TickMana = 0
	user.RoundData.TickStone = 0
	user.RoundData.TickWood = 0
	user.RoundData.TickMetal = 0
}

func (user *User) updateTickField(field string, val float64) {
	fmt.Println("updateTickField: " + field + " -- " + fmt.Sprint(val))
	switch field {
	case "build_power":
		user.RoundData.BuildPower += val
	case "recruit_power":
		user.RoundData.RecruitPower += val
	case "defense":
		user.RoundData.Defense += val
	case "food_tick":
		user.RoundData.TickFood += val
	case "wood_tick":
		user.RoundData.TickWood += val
	case "stone_tick":
		user.RoundData.TickStone += val
	case "metal_tick":
		user.RoundData.TickMetal += val
	case "gold_tick":
		user.RoundData.TickGold += val
	case "mana_tick":
		user.RoundData.TickMana += val
	case "faith_tick":
		user.RoundData.TickFaith += val
	case "housing":
		user.RoundData.Housing += val
	default:
		fmt.Println("Invalid Bonus Field:", field)
		//panic("INVALID BONUS FIELD")
	}
}

func (user *User) updateTicks(ctx context.Context) {
	_, span := Tracer.Start(ctx, "user-update-ticks")
	defer span.End()

	fmt.Println("updateTicks")
	user.resetTicks()

	log.Warn().Msg("Population: " + fmt.Sprint(user.RoundData.Population))
	user.RoundData.TickGold = user.RoundData.Population

	for _, unit := range user.Units {

		baseUnit := user.Round.MapUnitsById[unit.UnitID]
		quantity := math.Floor(unit.Quantity)
		fmt.Println("Upkeep: " + fmt.Sprint(baseUnit.UpkeepGold))
		fmt.Println("Quantity: " + fmt.Sprint(quantity))

		if quantity > 0 {
			user.RoundData.TickFaith -= utilities.RoundFloat(quantity*float64(baseUnit.UpkeepFaith), 2)
			user.RoundData.TickFood -= utilities.RoundFloat(quantity*float64(baseUnit.UpkeepFood), 2)
			user.RoundData.TickGold -= utilities.RoundFloat(quantity*float64(baseUnit.UpkeepGold), 2)
			user.RoundData.TickMana -= utilities.RoundFloat(quantity*float64(baseUnit.UpkeepMana), 2)
			user.RoundData.TickMetal -= utilities.RoundFloat(quantity*float64(baseUnit.UpkeepMetal), 2)
			user.RoundData.TickStone -= utilities.RoundFloat(quantity*float64(baseUnit.UpkeepStone), 2)
			user.RoundData.TickWood -= utilities.RoundFloat(quantity*float64(baseUnit.UpkeepWood), 2)
		}

		fmt.Println("Current Gold Tick(units): " + fmt.Sprint(user.RoundData.TickGold))
	}

	for _, building := range user.Buildings {
		baseBuilding := user.Round.MapBuildingsById[building.BuildingID]
		quantity := math.Floor(building.Quantity)

		if quantity > 0 {
			val := utilities.RoundFloat(math.Floor(building.Quantity*float64(baseBuilding.BonusValue)), 2)
			user.updateTickField(baseBuilding.BonusField, val)

			log.Warn().Msg("Field: " + baseBuilding.BonusField)
			fmt.Println("Current Gold Tick(buildings): " + fmt.Sprint(user.RoundData.TickGold))

			user.RoundData.TickFaith -= utilities.RoundFloat(quantity*float64(baseBuilding.UpkeepFaith), 2)
			user.RoundData.TickFood -= utilities.RoundFloat(quantity*float64(baseBuilding.UpkeepFood), 2)
			user.RoundData.TickGold -= utilities.RoundFloat(quantity*float64(baseBuilding.UpkeepGold), 2)
			user.RoundData.TickMana -= utilities.RoundFloat(quantity*float64(baseBuilding.UpkeepMana), 2)
			user.RoundData.TickMetal -= utilities.RoundFloat(quantity*float64(baseBuilding.UpkeepMetal), 2)
			user.RoundData.TickStone -= utilities.RoundFloat(quantity*float64(baseBuilding.UpkeepStone), 2)
			user.RoundData.TickWood -= utilities.RoundFloat(quantity*float64(baseBuilding.UpkeepWood), 2)
		}

		fmt.Println("Current Gold Tick(buildings): " + fmt.Sprint(user.RoundData.TickGold))
	}

	fmt.Println("Tick Gold:::", user.RoundData.TickGold)
}

func (user *User) UpdateRound(ctx context.Context, wg *sync.WaitGroup) bool {
	ctx, span := Tracer.Start(ctx, "user-update-round")
	defer span.End()
	if wg != nil {
		defer wg.Done()
	}

	user.updateTicks(ctx)
	if err := db.WithContext(ctx).Save(&user.RoundData).Error; err != nil {
		span.RecordError(err)
		return false
	}

	log.Warn().Msg("Round Data Saved")
	log.Warn().Msg("Gold Tick:" + fmt.Sprint(user.RoundData.TickGold))
	return true
}

func (user *User) updateBuildings(ctx context.Context, wg *sync.WaitGroup) bool {
	ctx, span := Tracer.Start(ctx, "user-update-buildings")
	defer span.End()
	if wg != nil {
		defer wg.Done()
	}

	if err := db.WithContext(ctx).Save(&user.Buildings).Error; err != nil {
		span.SetName("user-update-buildings-ERROR")
		span.RecordError(err)
		return false
	}
	return true
}

func (user *User) updateItems(ctx context.Context, wg *sync.WaitGroup) {
	wg.Done()

	// ctx, span := Tracer.Start(ctx, "user-update-items")
	// defer span.End()
	// defer wg.Done()

	// db.WithContext(ctx).Save(&user.Items)
}

func (user *User) loadUnits(ctx context.Context, wg *sync.WaitGroup) {
	ctx, span := Tracer.Start(ctx, "user-load-units")
	defer span.End()
	defer wg.Done()

	db.WithContext(ctx).Where("user_id = ? and round_id = ?", user.ID, getRound(user)).Find(&user.Units)
	for _, unit := range user.Units {
		unit.UnitGuid = user.Round.MapUnitsById[unit.UnitID].GUID
	}
}

func (user *User) loadRound(ctx context.Context, wg *sync.WaitGroup) {
	ctx, span := Tracer.Start(ctx, "user-load-round")
	defer span.End()
	defer wg.Done()

	db.WithContext(ctx).Where("user_id = ? and round_id = ?", user.ID, getRound(user)).Find(&user.RoundData)
}

func (user *User) loadBuildings(ctx context.Context, wg *sync.WaitGroup) {
	ctx, span := Tracer.Start(ctx, "user-load-buildings")
	defer span.End()
	defer wg.Done()

	db.WithContext(ctx).Where("user_id = ? and round_id = ?", user.ID, getRound(user)).Find(&user.Buildings)
	for _, building := range user.Buildings {
		building.BuildingGuid = user.Round.MapBuildingsById[building.BuildingID].GUID
	}
}

func (user *User) loadItems(ctx context.Context, wg *sync.WaitGroup) {
	ctx, span := Tracer.Start(ctx, "user-load-items")
	defer span.End()
	defer wg.Done()

	db.WithContext(ctx).Where("user_id = ?", user.ID).Find(&user.Items)
}

func (user *User) Load() *User {
	log.Debug().Msg("Load")

	ctx, sp := Tracer.Start(context.Background(), "loading-user")
	defer sp.End()

	if err := db.WithContext(ctx).First(&user).Error; err != nil {
		return nil
	}

	user.sendUserData()
	return user
}

func (user *User) LoadForRound(userid int, roundid int) *User {
	log.Debug().
		Int("userid", userid).
		Int("roundid", roundid).
		Msg("LoadForRound")

	ctx, sp := Tracer.Start(context.Background(), "loading-user")
	defer sp.End()

	user = &User{RoundLoading: roundid}
	user.ID = uint(userid)
	if err := db.WithContext(ctx).First(&user).Error; err != nil {
		return nil
	}

	return user
}

func (user *User) Dump() {
	fmt.Println("============USER=============")
	fmt.Println("GUID:", user.GUID)
	fmt.Println("Email:", user.Email)
	fmt.Println("Round:", user.RoundID)
	fmt.Println("RoundLoading:", user.RoundLoading)
	fmt.Println("Password:", user.Password)
	fmt.Print("Admin:")
	if user.Admin {
		fmt.Println("Yes")
	} else {
		fmt.Println("No")
	}

	fmt.Println("============ROUND============")
	fmt.Println("Energy:", user.RoundData.Energy)
	fmt.Println("RecruitPower:", user.RoundData.RecruitPower)
	fmt.Println("BuildPower:", user.RoundData.BuildPower)
	fmt.Println("Gold:", user.RoundData.Gold)
	fmt.Println("GoldTick:", user.RoundData.TickGold)
	fmt.Println("Food:", user.RoundData.Food)
	fmt.Println("FoodTick:", user.RoundData.TickFood)
	fmt.Println("Wood:", user.RoundData.Wood)
	fmt.Println("WoodTick:", user.RoundData.TickWood)
	fmt.Println("Faith:", user.RoundData.Faith)
	fmt.Println("FaithTick:", user.RoundData.TickFaith)
	fmt.Println("Stone:", user.RoundData.Stone)
	fmt.Println("StoneTick:", user.RoundData.TickStone)
	fmt.Println("Mana:", user.RoundData.Mana)
	fmt.Println("ManaTick:", user.RoundData.TickMana)

	fmt.Println("============UNITS============")
	for i := 0; i < len(user.Units); i++ {
		fmt.Println("ID", user.Units[i].UnitID, ":", user.Units[i].Quantity)
	}

	fmt.Println("==========BUILDINGS==========")
	for i := 0; i < len(user.Buildings); i++ {
		fmt.Println("ID", user.Buildings[i].BuildingID, ":", user.Buildings[i].Quantity)
	}

	fmt.Println("============ITEMS============")
	for i := 0; i < len(user.Items); i++ {
		fmt.Println("ID", user.Items[i].ID)
	}

	fmt.Println("=============================")
}

func (user *User) sendUserData() {
	userData, err := json.Marshal(user)
	if err != nil {
		user.SendMessage(payloads.Response{
			Type: "USER_DATA",
			Data: append([]byte("\"user\":"), userData...),
		})

		log.Info().Msg("Sent User Data")
	}
}

type SendErrorParams struct {
	Context *context.Context
	Type    string
	Message string
	Err     error
}

// func (user *User) SendError(ctx context.Context, errorType string, message string) {
func (user *User) SendError(params SendErrorParams) {
	if span := utilities.GetSpan(*params.Context); span != nil {
		span.SetStatus(codes.Error, params.Message)
		if params.Err != nil {
			span.RecordError(params.Err)
		}
	}

	if user.Connection != nil {
		payload, err := json.Marshal(struct {
			Type    string `json:"type"`
			Message string `json:"message"`
		}{
			Type:    strings.ToUpper("error_" + params.Type),
			Message: params.Message,
		})

		if err == nil {
			user.Connection.WriteMessage(1, payload)
		} else {
			fmt.Println(err)
			fmt.Println("Error Sending Error:", params.Message)
		}
	} else {
		fmt.Println("Connection is nil")
	}
}

func (user *User) SendMessage(packet any) {
	fmt.Println("-----------------------------")
	fmt.Println("SendMessage:", packet)

	if user.Connection != nil {
		//fmt.Println("Packet:", packet)
		if payload, err := json.Marshal(packet); err == nil {
			//fmt.Println("Payload:", payload)
			user.Connection.WriteMessage(1, payload)
			//fmt.Println("Sent:", packet)
		} else {
			fmt.Println(err)
			fmt.Println("Error Sending:", packet)
		}
	} else {
		fmt.Println("Connection is nil")
	}
	fmt.Println("-----------------------------")
}

func (user *User) Log(message string, round uint) {
	ctx, span := Tracer.Start(context.Background(), "log")
	defer span.End()

	fmt.Println("Log: " + message)

	log := UserLog{Message: message, RoundID: round}

	if err := db.WithContext(ctx).Save(&log).Error; err != nil {
		fmt.Println("Error logging: " + message)
	} else {
		fmt.Println("Wrote log")
	}
}

func (user *User) LogEvent(eventText string, round uuid.UUID) {
	log.Info().Msg("LogEvent: " + eventText)

	ctx, span := Tracer.Start(context.Background(), "log-event")
	defer span.End()

	event := Event{Event: eventText, Round: round, UserID: user.ID, Seen: false}

	if err := db.WithContext(ctx).Save(&event).Error; err != nil {
		log.Error().AnErr("Error saving event", err).Msg("Error saving event")
	}
}

func (user *User) getField(field string) int {
	switch field {
	case "gold":
		return int(math.Floor(user.RoundData.Gold))
	case "food":
		return int(math.Floor(user.RoundData.Food))
	case "wood":
		return int(math.Floor(user.RoundData.Wood))
	case "stone":
		return int(math.Floor(user.RoundData.Stone))
	case "metal":
		return int(math.Floor(user.RoundData.Metal))
	case "faith":
		return int(math.Floor(user.RoundData.Faith))
	case "mana":
		return int(math.Floor(user.RoundData.Mana))
	}

	return 0
}

func (user *User) zeroField(field string) {
	switch field {
	case "gold":
		user.RoundData.Gold = 0
	case "food":
		user.RoundData.Food = 0
	case "wood":
		user.RoundData.Wood = 0
	case "stone":
		user.RoundData.Stone = 0
	case "metal":
		user.RoundData.Metal = 0
	case "faith":
		user.RoundData.Faith = 0
	case "mana":
		user.RoundData.Mana = 0
	}
}

func (user *User) getTick(field string) int {
	switch field {
	case "gold":
		return int(math.Floor(user.RoundData.TickGold))
	case "food":
		return int(math.Floor(user.RoundData.TickFood))
	case "wood":
		return int(math.Floor(user.RoundData.TickWood))
	case "stone":
		return int(math.Floor(user.RoundData.TickStone))
	case "metal":
		return int(math.Floor(user.RoundData.TickMetal))
	case "faith":
		return int(math.Floor(user.RoundData.TickFaith))
	case "mana":
		return int(math.Floor(user.RoundData.TickMana))
	}

	return 0
}

func (user *User) getDeficit(field string) int {
	return user.getField(field) + user.getTick(field)
}

func (user *User) ProcessBankruptcy(ctx context.Context, field string) bool {
	log.Info().Msg("ProcessBankruptcy: " + field)

	picker := utilities.Picker{}
	for _, u := range user.Units {
		unit := user.Round.GetUnitById(u.UnitID)
		picker.Add(unit.GetUpkeep(field)*uint(u.Quantity), u.UnitID)
	}

	if user.getDeficit(field) == 0 {
		user.zeroField(field)
		user.updateUnits(ctx, nil)
		user.UpdateRound(ctx, nil)

		return true
	}

	choice := picker.Choose()
	for _, u := range user.Units {
		log.Warn().Msg("Choice " + fmt.Sprint(choice) + " ::: " + fmt.Sprint(u.UnitID))
		if u.UnitID == choice {
			unit := user.Round.GetUnitById(u.UnitID)
			deficit := user.getDeficit(field)

			fmt.Print(-float64(deficit) / float64(unit.GetUpkeep(field)))
			count := int(math.Ceil(-float64(deficit) / float64(unit.GetUpkeep(field))))
			log.Info().Msg("Deficit: " + fmt.Sprint(deficit))
			log.Info().Msg("Unit Upkeep: " + fmt.Sprint(unit.GetUpkeep(field)))
			log.Info().Msg("Get rid of " + fmt.Sprint(count) + " " + unit.Name)

			taken := user.takeUnit(ctx, int(u.UnitID), count)
			if taken {
				user.updateTicks(ctx)
				return user.ProcessBankruptcy(ctx, field)
			}
		}
	}

	log.Info().Msg("Deficit Not handled: " + fmt.Sprint(choice))
	return false
}

func (user *User) takeUnit(ctx context.Context, unitid int, amount int) bool {
	for _, u := range user.Units {
		if u.UnitID == uint(unitid) {
			u.Quantity -= float64(amount)

			if u.Quantity <= 0 {
				u.Quantity = 0
			}

			user.LogEvent("Took "+fmt.Sprint(amount)+" "+user.Round.GetUnitById(u.UnitID).Name, user.Round.GUID)
			return true
		}
	}

	return false
}

func (user *User) Refresh() {
	log.Info().Msg("Refresh:" + fmt.Sprint(user.ID))

	user.Load()
}
