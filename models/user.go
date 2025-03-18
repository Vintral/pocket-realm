package models

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	realmRedis "github.com/Vintral/pocket-realm/redis"
	"github.com/Vintral/pocket-realm/utils"
	realmUtils "github.com/Vintral/pocket-realm/utils"
	"github.com/rs/zerolog/log"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	redisDef "github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	gorm "gorm.io/gorm"
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
	Buffs        []*UserBuff     `gorm:"-" json:"buffs"`
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
	log.Warn().Int("id", roundId).Msg("getRound")

	if roundId != 0 {
		round, err := LoadRoundById(ctx, roundId)
		if err != nil {
			log.Error().AnErr("err", err).Int("round", user.RoundID).Msg("Error Loading Round")
			return err
		}

		user.Round = round

		//user.loadSynchronous(ctx)
		wg := new(sync.WaitGroup)
		wg.Add(5)
		go user.loadRound(ctx, wg)
		go user.loadUnits(ctx, wg)
		go user.loadBuildings(ctx, wg)
		go user.loadItems(ctx, wg)
		go user.loadBuffs(ctx, wg)
		wg.Wait()
	}

	return
}

func (user *User) AfterUpdate(tx *gorm.DB) (err error) {
	ctx, sp := Tracer.Start(tx.Statement.Context, "after-update")
	defer sp.End()

	wg := new(sync.WaitGroup)
	wg.Add(5)
	go user.UpdateRound(ctx, wg)
	go user.updateUnits(ctx, wg)
	go user.updateBuildings(ctx, wg)
	go user.updateItems(ctx, wg)
	go user.updateBuffs(ctx, wg)
	wg.Wait()

	return
}

func (user *User) updateUnits(ctx context.Context, wg *sync.WaitGroup) bool {
	ctx, span := Tracer.Start(ctx, "user-update-units")
	defer span.End()
	if wg != nil {
		defer wg.Done()
	}

	if len(user.Units) != 0 {
		if err := db.WithContext(ctx).Save(&user.Units).Error; err != nil {
			log.Error().AnErr("error", err).Msg("Error Updating Units")
			span.SetStatus(codes.Error, err.Error())
			return false
		}
	}

	log.Trace().Msg("Updated units")
	return true
}

func (user *User) updateBuffs(ctx context.Context, wg *sync.WaitGroup) bool {
	ctx, span := Tracer.Start(ctx, "user-update-buffs")
	defer span.End()
	if wg != nil {
		defer wg.Done()
	}

	if len(user.Buffs) != 0 {
		if err := db.WithContext(ctx).Save(&user.Buffs).Error; err != nil {
			log.Error().AnErr("error", err).Msg("Error Updating Buffs")
			span.SetStatus(codes.Error, err.Error())
			return false
		}
	}

	log.Trace().Msg("Updated buggs")
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
	log.Trace().Msg("updateTickField: " + field + " -- " + fmt.Sprint(val))

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
		log.Warn().Msg("Invalid Bonus Field:" + field)
		//panic("INVALID BONUS FIELD")
	}
}

func (user *User) updateTicks(ctx context.Context) {
	// user.resetTicks()

	//log.Warn().Msg("Population: " + fmt.Sprint(user.RoundData.Population))

	CostFaith := 0.0
	CostFood := 0.0
	CostGold := 0.0
	CostMana := 0.0
	CostMetal := 0.0
	CostStone := 0.0
	CostWood := 0.0

	BaseFaith := 0.0
	FaithPercentModifier := 1.0
	BaseFood := 0.0
	FoodPercentModifier := 1.0
	BaseGold := user.RoundData.Population
	GoldPercentModifier := 1.0
	BaseMana := 0.0
	ManaPercentModifier := 1.0
	BaseMetal := 0.0
	MetalPercentModifier := 1.0
	BaseStone := 0.0
	StonePercentModifier := 1.0
	BaseWood := 0.0
	WoodPercentModifier := 1.0
	BaseBuildPower := 5.0
	BuildPowerPercentModifier := 1.0
	BaseRecruitPower := 5.0
	RecruitPowerPercentModifier := 1.0
	BaseDefense := 0.0
	DefensePercentModifier := 1.0
	BaseHousing := 0.0
	HousingPercentModifier := 1.0

	for _, unit := range user.Units {
		baseUnit := user.Round.MapUnitsById[unit.UnitID]
		quantity := math.Floor(unit.Quantity)

		if quantity > 0 {
			CostFaith -= realmUtils.RoundFloat(quantity*float64(baseUnit.UpkeepFaith), 2)
			CostFood -= realmUtils.RoundFloat(quantity*float64(baseUnit.UpkeepFood), 2)
			CostGold -= realmUtils.RoundFloat(quantity*float64(baseUnit.UpkeepGold), 2)
			CostMana -= realmUtils.RoundFloat(quantity*float64(baseUnit.UpkeepMana), 2)
			CostMetal -= realmUtils.RoundFloat(quantity*float64(baseUnit.UpkeepMetal), 2)
			CostStone -= realmUtils.RoundFloat(quantity*float64(baseUnit.UpkeepStone), 2)
			CostWood -= realmUtils.RoundFloat(quantity*float64(baseUnit.UpkeepWood), 2)
		}
	}

	for _, building := range user.Buildings {
		baseBuilding := user.Round.MapBuildingsById[building.BuildingID]
		quantity := math.Floor(building.Quantity)

		if quantity > 0 {
			val := realmUtils.RoundFloat(math.Floor(building.Quantity*float64(baseBuilding.BonusValue)), 2)
			switch baseBuilding.BonusField {
			case "build_power":
				BaseBuildPower += val
			case "recruit_power":
				BaseRecruitPower += val
			case "defense":
				BaseDefense += val
			case "food_tick":
				BaseFood += val
			case "wood_tick":
				BaseWood += val
			case "stone_tick":
				BaseStone += val
			case "metal_tick":
				BaseMetal += val
			case "gold_tick":
				BaseGold += val
			case "mana_tick":
				BaseMana += val
			case "faith_tick":
				BaseFaith += val
			case "housing":
				BaseHousing += val
			}

			CostFaith -= realmUtils.RoundFloat(quantity*float64(baseBuilding.UpkeepFaith), 2)
			CostFood -= realmUtils.RoundFloat(quantity*float64(baseBuilding.UpkeepFood), 2)
			CostGold -= realmUtils.RoundFloat(quantity*float64(baseBuilding.UpkeepGold), 2)
			CostMana -= realmUtils.RoundFloat(quantity*float64(baseBuilding.UpkeepMana), 2)
			CostMetal -= realmUtils.RoundFloat(quantity*float64(baseBuilding.UpkeepMetal), 2)
			CostStone -= realmUtils.RoundFloat(quantity*float64(baseBuilding.UpkeepStone), 2)
			CostWood -= realmUtils.RoundFloat(quantity*float64(baseBuilding.UpkeepWood), 2)
		}
	}

	for _, userBuff := range user.Buffs {
		if buff, err := LoadBuffById(ctx, int(userBuff.BuffID)); err == nil {
			for _, effect := range buff.Effects {
				var field *float64
				var fieldPercent *float64

				switch effect.Field {
				case "build_power":
					field = &BaseBuildPower
					fieldPercent = &BuildPowerPercentModifier
				case "recruit_power":
					field = &BaseRecruitPower
					fieldPercent = &RecruitPowerPercentModifier
				case "defense":
					field = &BaseDefense
					fieldPercent = &DefensePercentModifier
				case "food_tick":
					field = &BaseFood
					fieldPercent = &FoodPercentModifier
				case "wood_tick":
					field = &BaseWood
					fieldPercent = &WoodPercentModifier
				case "stone_tick":
					field = &BaseStone
					fieldPercent = &StonePercentModifier
				case "metal_tick":
					field = &BaseMetal
					fieldPercent = &MetalPercentModifier
				case "gold_tick":
					field = &BaseGold
					fieldPercent = &GoldPercentModifier
				case "mana_tick":
					field = &BaseMana
					fieldPercent = &ManaPercentModifier
				case "faith_tick":
					field = &BaseFaith
					fieldPercent = &FaithPercentModifier
				case "housing":
					field = &BaseHousing
					fieldPercent = &HousingPercentModifier
				default:
					log.Warn().Msg("Unexpected buff field")
				}

				if effect.Percent {
					*fieldPercent += float64(effect.Amount*int(userBuff.Stacks)) / 100.0
				} else {
					*field += float64(effect.Amount * int(userBuff.Stacks))
				}
			}
		} else {
			log.Error().AnErr("err", err).Msg("Buff not found")
		}
	}

	user.RoundData.TickFood = (BaseFood * FoodPercentModifier) - CostFood
	user.RoundData.TickWood = (BaseWood * WoodPercentModifier) - CostWood
	user.RoundData.TickGold = (BaseGold * GoldPercentModifier) - CostGold
	user.RoundData.TickFaith = (BaseFaith * FaithPercentModifier) - CostFaith
	user.RoundData.TickMana = (BaseMana * ManaPercentModifier) - CostMana
	user.RoundData.TickMetal = (BaseMetal * MetalPercentModifier) - CostMetal
	user.RoundData.TickStone = (BaseStone * StonePercentModifier) - CostStone
	user.RoundData.BuildPower = BaseBuildPower * BuildPowerPercentModifier
	user.RoundData.RecruitPower = BaseRecruitPower * RecruitPowerPercentModifier
	user.RoundData.Housing = BaseHousing * HousingPercentModifier
	user.RoundData.Defense = BaseDefense * DefensePercentModifier
}

func (user *User) UpdateRank(base context.Context) {
	ctx, span := Tracer.Start(base, "update-rank")
	defer span.End()

	log.Trace().Msg("Update Rank")

	redisClient, err := realmRedis.Instance(nil)
	if err != nil {
		log.Warn().AnErr("error", err).Msg("Error getting redis client")
		return
	}

	score := math.Floor(user.RoundData.Land * 10)
	log.Warn().Msg("UpdateRank: " + fmt.Sprint(user.ID) + " -- " + fmt.Sprint(score))

	result := redisClient.ZAdd(
		ctx,
		fmt.Sprint(user.RoundID)+"-rankings",
		redisDef.Z{Score: user.RoundData.Land * 10, Member: user.ID},
	)
	if result.Err() != nil {
		log.Warn().AnErr("err", result.Err()).Msg("Error updating redis rank")
		return
	}
}

func (user *User) UpdateRound(ctx context.Context, wg *sync.WaitGroup) bool {
	ctx, span := Tracer.Start(ctx, "user-update-round")
	defer span.End()
	if wg != nil {
		defer wg.Done()
	}

	if user.RoundData.ID != 0 {
		user.updateTicks(ctx)
		if err := db.WithContext(ctx).Save(&user.RoundData).Error; err != nil {
			span.RecordError(err)
			return false
		}

		user.UpdateRank(ctx)

		log.Warn().Msg("Round Data Saved")
		log.Warn().Msg("Gold Tick:" + fmt.Sprint(user.RoundData.TickGold))
	}

	return true
}

func (user *User) updateBuildings(ctx context.Context, wg *sync.WaitGroup) bool {
	ctx, span := Tracer.Start(ctx, "user-update-buildings")
	defer span.End()
	if wg != nil {
		defer wg.Done()
	}

	if len(user.Buildings) > 0 {
		if err := db.WithContext(ctx).Save(&user.Buildings).Error; err != nil {
			log.Error().AnErr("error", err).Msg("Error Updating Buildings")
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return false
		}
	}

	log.Trace().Msg("Updated buildings")
	return true
}

func (user *User) updateItems(ctx context.Context, wg *sync.WaitGroup) bool {
	ctx, span := Tracer.Start(ctx, "user-update-items")
	defer span.End()
	if wg != nil {
		defer wg.Done()
	}

	if len(user.Items) > 0 {
		if err := db.WithContext(ctx).Save(&user.Items).Error; err != nil {
			log.Error().AnErr("error", err).Msg("Error Updating Items")
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return false
		}
	}

	log.Trace().Msg("Updated items")
	return true
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

	// log.Warn().Msg("MetalTick Before: " + fmt.Sprint(user.RoundData.TickMetal))
	db.WithContext(ctx).Where("user_id = ? and round_id = ?", user.ID, getRound(user)).Find(&user.RoundData)
	// log.Warn().Msg("MetalTick After: " + fmt.Sprint(user.RoundData.TickMetal))
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

func (user *User) loadBuffs(ctx context.Context, wg *sync.WaitGroup) {
	ctx, span := Tracer.Start(ctx, "user-load-buffs")
	defer span.End()
	defer wg.Done()

	log.Warn().Msg("loadBuffs")

	db.WithContext(ctx).Where("user_id = ? and round_id = ?", user.ID, getRound(user)).Find(&user.Buffs)
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

func (user *User) Dump() {
	log.Trace().Msg("============USER=============")
	log.Trace().Msg("GUID:" + fmt.Sprint(user.GUID))
	log.Trace().Msg("Email:" + fmt.Sprint(user.Email))
	log.Trace().Msg("Round:" + fmt.Sprint(user.RoundID))
	log.Trace().Msg("RoundLoading:" + fmt.Sprint(user.RoundLoading))
	log.Trace().Msg("Password:" + fmt.Sprint(user.Password))
	if user.Admin {
		log.Trace().Msg("Admin: Yes")
	} else {
		log.Trace().Msg("Admin: No")
	}

	log.Trace().Msg("============ROUND============")
	log.Trace().Msg("Energy:" + fmt.Sprint(user.RoundData.Energy))
	log.Trace().Msg("RecruitPower:" + fmt.Sprint(user.RoundData.RecruitPower))
	log.Trace().Msg("BuildPower:" + fmt.Sprint(user.RoundData.BuildPower))
	log.Trace().Float64("have", user.RoundData.Gold).Float64("tick", user.RoundData.TickGold).Msg("gold")
	log.Trace().Float64("have", user.RoundData.Food).Float64("tick", user.RoundData.TickFood).Msg("food")
	log.Trace().Float64("have", user.RoundData.Wood).Float64("tick", user.RoundData.TickWood).Msg("wood")
	log.Trace().Float64("have", user.RoundData.Metal).Float64("tick", user.RoundData.TickMetal).Msg("metal")
	log.Trace().Float64("have", user.RoundData.Faith).Float64("tick", user.RoundData.TickFaith).Msg("faith")
	log.Trace().Float64("have", user.RoundData.Stone).Float64("tick", user.RoundData.TickStone).Msg("stone")
	log.Trace().Float64("have", user.RoundData.Mana).Float64("tick", user.RoundData.TickMana).Msg("mana")

	log.Trace().Msg("============UNITS============")
	for i := 0; i < len(user.Units); i++ {
		log.Trace().Float64("quantity", user.Units[i].Quantity).Msg("ID: " + fmt.Sprint(user.Units[i].UnitID))
	}

	log.Trace().Msg("==========BUILDINGS==========")
	for i := 0; i < len(user.Buildings); i++ {
		log.Trace().Float64("quantity", user.Buildings[i].Quantity).Msg("ID: " + fmt.Sprint(user.Buildings[i].BuildingID))
	}

	log.Trace().Msg("============ITEMS============")
	for i := 0; i < len(user.Items); i++ {
		log.Trace().Msg("ID: " + fmt.Sprint(user.Items[i].ID))
	}

	log.Trace().Msg("============BUFFS============")
	for i := 0; i < len(user.Buffs); i++ {
		log.Trace().Msg("ID: " + fmt.Sprint(user.Buffs[i].ID))
	}

	log.Trace().Msg("=============================")
}

func (user *User) sendUserData() {
	// userData, err := json.Marshal(user)
	user.SendMessage(struct {
		Type string `json:"type"`
		User *User  `json:"user"`
	}{
		Type: "USER_DATA",
		User: user,
	})

	log.Info().Msg("Sent User Data")
}

type SendErrorParams struct {
	Context *context.Context
	Type    string
	Message string
	Err     error
}

func (user *User) updateField(field string, val float64) bool {
	switch field {
	case "wood":
		if val < 0 && user.RoundData.Wood < val {
			return false
		}
		user.RoundData.Wood += val
	case "gold":
		if val < 0 && user.RoundData.Gold < val {
			return false
		}
		user.RoundData.Gold += val
	case "food":
		if val < 0 && user.RoundData.Food < val {
			return false
		}
		user.RoundData.Food += val
	case "stone":
		if val < 0 && user.RoundData.Stone < val {
			return false
		}
		user.RoundData.Stone += val
	case "metal":
		if val < 0 && user.RoundData.Metal < val {
			return false
		}
		user.RoundData.Metal += val
	case "mana":
		if val < 0 && user.RoundData.Mana < val {
			return false
		}
		user.RoundData.Mana += val
	case "faith":
		if val < 0 && user.RoundData.Faith < val {
			return false
		}
		user.RoundData.Faith += val
	default:
		log.Warn().Msg(fmt.Sprint("Unexpected field: ", field))
		return false
	}

	return true
}

// func (user *User) SendError(ctx context.Context, errorType string, message string) {
func (user *User) SendError(params SendErrorParams) {
	if span := realmUtils.GetSpan(*params.Context); span != nil {
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
			log.Error().AnErr("err", err).Msg("Error Sending Error: " + params.Message)
		}
	} else {
		log.Error().Msg("Connection is nil")
	}
}

func (user *User) SendMessage(packet any) {
	log.Trace().Msg("SendMessage")

	if user.Connection != nil {
		if payload, err := json.Marshal(packet); err == nil {
			user.Connection.WriteMessage(1, payload)
		} else {
			log.Error().AnErr("err", err).Msg("Error in user.SendMessage")
		}
	} else {
		log.Warn().Msg("Connection is nil")
	}
}

func (user *User) Log(message string, round uint) {
	ctx, span := Tracer.Start(context.Background(), "log")
	defer span.End()

	userlog := UserLog{Message: message, RoundID: round, UserID: user.ID}

	if err := db.WithContext(ctx).Save(&userlog).Error; err != nil {
		log.Warn().Msg("Error logging: " + message)
	}
}

func (user *User) LogEvent(eventText string, round uuid.UUID) {
	log.Info().Any("round", round).Msg("LogEvent: " + eventText)

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
	have := user.getField(field)
	tick := user.getTick(field)
	log.Warn().Str("field", field).Int("have", have).Int("tick", tick).Msg("getDeficit: " + fmt.Sprint(have+tick))

	return have + tick
}

func (user *User) ProcessBankruptcy(ctx context.Context, field string) bool {
	log.Warn().Msg("ProcessBankruptcy: " + field)

	if user.getDeficit(field) >= 0 {
		user.zeroField(field)

		wg := new(sync.WaitGroup)
		wg.Add(3)
		user.updateUnits(ctx, wg)
		user.updateBuildings(ctx, wg)
		user.UpdateRound(ctx, wg)
		wg.Wait()
		return true
	}

	user.Dump()

	picker := realmUtils.Picker{}
	for _, u := range user.Units {
		unit := user.Round.GetUnitById(u.UnitID)
		picker.Add(unit.GetUpkeep(field)*uint(u.Quantity), u.UnitID)
	}

	choice := picker.Choose()
	for _, u := range user.Units {
		log.Trace().Msg("Choice " + fmt.Sprint(choice) + " ::: " + fmt.Sprint(u.UnitID))
		if u.UnitID == choice {
			unit := user.Round.GetUnitById(u.UnitID)
			deficit := user.getDeficit(field)

			count := int(math.Ceil(-float64(deficit) / float64(unit.GetUpkeep(field))))
			log.Warn().Msg("Deficit: " + fmt.Sprint(deficit))
			log.Warn().Msg(fmt.Sprint(-float64(deficit) / float64(unit.GetUpkeep(field))))
			log.Warn().Msg("Unit Upkeep: " + fmt.Sprint(unit.GetUpkeep(field)))

			if count == 0 {
				log.Panic().Msg("Count is 0")
			}

			log.Warn().Msg("Get rid of " + fmt.Sprint(count) + " " + unit.Name)

			taken := user.takeUnit(ctx, int(u.UnitID), count)
			if taken {
				user.updateTicks(ctx)
				return user.ProcessBankruptcy(ctx, field)
			}
		}
	}

	picker = realmUtils.Picker{}
	for _, b := range user.Buildings {
		building := user.Round.GetUnitById(b.BuildingID)
		picker.Add(building.GetUpkeep(field)*uint(b.Quantity), b.BuildingID)
	}

	choice = picker.Choose()
	for _, b := range user.Buildings {
		log.Trace().Msg("Choice " + fmt.Sprint(choice) + " ::: " + fmt.Sprint(b.BuildingID))
		if b.BuildingID == choice {
			building := user.Round.GetBuildingById(b.BuildingID)
			deficit := user.getDeficit(field)

			count := int(math.Ceil(-float64(deficit) / float64(building.GetUpkeep(field))))
			log.Warn().Msg("Deficit: " + fmt.Sprint(deficit))
			log.Warn().Msg(fmt.Sprint(-float64(deficit) / float64(building.GetUpkeep(field))))
			log.Warn().Msg("Unit Upkeep: " + fmt.Sprint(building.GetUpkeep(field)))

			if count == 0 {
				log.Panic().Msg("Count is 0")
			}

			log.Warn().Msg("Get rid of " + fmt.Sprint(count) + " " + building.Name)

			taken := user.takeBuilding(ctx, int(b.BuildingID), count)
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
	_, span := Tracer.Start(ctx, "take-unit")
	defer span.End()

	span.SetAttributes(
		attribute.Int("unit", unitid),
		attribute.Int("amount", amount),
	)

	log.Info().Int("unit", unitid).Int("amount", amount).Msg("user.takeUnit")

	for _, u := range user.Units {
		if u.UnitID == uint(unitid) {
			u.Quantity -= float64(amount)

			if u.Quantity <= 0 {
				u.Quantity = 0
			}

			go user.LogEvent("Took "+fmt.Sprint(amount)+" "+user.Round.GetUnitById(u.UnitID).Name, user.Round.GUID)
			span.SetAttributes(
				attribute.Bool("success", true),
			)
			return true
		}
	}

	span.SetAttributes(
		attribute.Bool("success", false),
	)
	return false
}

func (user *User) takeBuilding(ctx context.Context, buildingid int, amount int) bool {
	_, span := Tracer.Start(ctx, "take-building")
	defer span.End()

	span.SetAttributes(
		attribute.Int("building", buildingid),
		attribute.Int("amount", amount),
	)
	log.Info().Int("building", buildingid).Int("amount", amount).Uint("user", user.ID).Msg("user.takeBuilding")

	for _, b := range user.Buildings {
		if b.BuildingID == uint(buildingid) {
			b.Quantity -= float64(amount)

			if b.Quantity <= 0 {
				b.Quantity = 0
			}

			go user.LogEvent("Took "+fmt.Sprint(amount)+" "+user.Round.GetBuildingById(b.BuildingID).Name, user.Round.GUID)
			span.SetAttributes(
				attribute.Bool("success", true),
			)
			return true
		}
	}

	span.SetAttributes(
		attribute.Bool("success", false),
	)
	return false
}

func (user *User) ChangeAvatar(baseContext context.Context, avatar string) bool {
	ctx, span := Tracer.Start(baseContext, "user.ChangeAvatar")
	defer span.End()

	log.Info().Int("user", int(user.ID)).Str("avatar", avatar).Msg("user.ChangeAvatar")

	user.Avatar = avatar
	err := db.WithContext(ctx).Save(&user).Error
	if err != nil {
		log.Warn().AnErr("err", err).Msg("Error saving user")
	}

	return err == nil
}

func (user *User) IsPlayingRound(ctx context.Context, round int) bool {
	log.Info().Uint("user", user.ID).Int("round", round).Msg("IsPlayingRound")

	var data *UserRound
	if err := db.WithContext(ctx).Where("user_id = ? AND round_id = ?", user.ID, round).Find(&data).Error; err != nil {
		log.Info().AnErr("error", err).Any("data", data).Msg("No User Found!")

		return false
	}

	if data.ID == 0 {
		log.Info().Any("user", data).Msg("No user found!")
		return false
	}

	log.Info().Any("user", data).Msg("Found User!")
	return true
}

func (user *User) Join(ctx context.Context, round *Round, classType string) *User {
	log.Info().Str("round", round.GUID.String()).Msg("Joining round")

	ctx, span := Tracer.Start(ctx, "user.Join")
	defer span.End()

	data := &UserRound{
		UserID:         user.ID,
		RoundID:        round.ID,
		CharacterClass: classType,
		// Energy:   int(round.EnergyMax),
		// Land:     float64(round.StartLand),
		// FreeLand: float64(round.StartLand),
	}
	db.WithContext(ctx).Create(&data)
	temp := LoadUserForRound(int(user.ID), int(round.ID))
	temp.RoundData = *data
	temp.RoundData.Energy = int(round.EnergyMax)
	temp.RoundData.Land = float64(round.StartLand)
	temp.RoundData.FreeLand = float64(round.StartLand)

	// temp.RoundID = int(round.ID)
	// temp.RoundPlaying = round.GUID

	for _, res := range round.Resources {
		log.Warn().Any("resource", res).Msg("Resource: " + res.Name)
		switch res.Name {
		case "gold":
			data.Gold = float64(res.StartWith)
		case "wood":
			data.Wood = float64(res.StartWith)
		case "food":
			data.Food = float64(res.StartWith)
		case "stone":
			data.Stone = float64(res.StartWith)
		case "metal":
			data.Metal = float64(res.StartWith)
		case "faith":
			data.Faith = float64(res.StartWith)
		case "mana":
			data.Mana = float64(res.StartWith)
		default:
			log.Warn().Msg("Unexpected Field: " + res.Name)
		}
	}

	buildings := []*UserBuilding{}
	for _, building := range round.Buildings {
		if building.StartWith > 0 {
			log.Warn().Any("building", building).Msg("Building: " + building.Name + " - " + string(building.StartWith))

			log.Warn().Msg(fmt.Sprintf("Slice Info: %d -- %d", len(buildings), cap(buildings)))
			buildings = append(buildings, &UserBuilding{
				UserID:     user.ID,
				BuildingID: building.ID,
				RoundID:    round.ID,
				Quantity:   float64(building.StartWith),
			})
			log.Warn().Msg(fmt.Sprintf("Slice Info: %d -- %d", len(buildings), cap(buildings)))
		}
	}
	temp.Buildings = buildings

	units := []*UserUnit{}
	for _, unit := range round.Units {
		if unit.StartWith > 0 {
			log.Warn().Any("unit", unit).Msg("Unit: " + unit.Name + " - " + string(unit.StartWith))

			units = append(units, &UserUnit{
				UserID:   user.ID,
				UnitID:   unit.ID,
				RoundID:  round.ID,
				Quantity: float64(unit.StartWith),
			})
		}
	}
	temp.Units = units

	log.Warn().Msg(fmt.Sprintf("User Buildings Length: %d", len(temp.Buildings)))
	log.Warn().Msg(fmt.Sprintf("User Units Length: %d", len(temp.Units)))
	db.WithContext(ctx).Save(&temp)

	user.LogEvent("Joined Round", round.GUID)
	return temp
}

func (user *User) SwitchRound(round *Round) bool {
	log.Info().Uint("round", round.ID).Msg(fmt.Sprintf("SwitchRound: %d", round.ID))

	log.Warn().Any("guid", round.GUID).Uint("id", round.ID).Uint("user", user.ID).Msg("TRYING TO SWITCH PLAYING ROUND")

	res := db.Model(&User{}).Where("id = ?", user.ID).Updates(User{RoundID: int(round.ID), RoundPlaying: round.GUID})
	if res.Error != nil {
		log.Warn().AnErr("error", res.Error).Msg("Error Updating User Round Info")
		return false
	} else {
		log.Warn().Msg("Updated user")
	}

	return true
}

func (user *User) Refresh() {
	log.Info().Msg("Refresh:" + fmt.Sprint(user.ID))

	user.RoundData = UserRound{}
	user.Load()
}

func (user *User) AddItem(baseContext context.Context, item *Item) bool {
	ctx, span := Tracer.Start(baseContext, "add-item")
	defer span.End()

	log.Info().Msg("AddItem")

	var temp *UserItem
	found := false
	for i := 0; i < len(user.Items) && !found; i++ {
		if user.Items[i].ItemID == item.ID {
			temp = user.Items[i]
			found = true
		}
	}
	if temp == nil {
		log.Debug().Msg("Create New UserItem")
		temp = &UserItem{
			UserID:   user.ID,
			ItemID:   item.ID,
			ItemGuid: item.GUID,
		}
		user.Items = append(user.Items, temp)
	}

	temp.Quantity++
	return user.updateItems(ctx, nil)
}

func (user *User) TakeItem(baseContext context.Context, item *Item) bool {
	ctx, span := Tracer.Start(baseContext, "take-item")
	defer span.End()

	log.Info().Msg("user.TakeItem")

	var temp *UserItem
	found := false
	for i := 0; i < len(user.Items) && !found; i++ {
		if user.Items[i].ItemID == item.ID {
			temp = user.Items[i]
			found = true
		}
	}

	if !found {
		log.Error().Any("item", item).Uint("user-id", user.ID).Msg("Error taking item")
	}

	temp.Quantity--
	return user.updateItems(ctx, nil)
}

func (user *User) AddUnit(baseContext context.Context, unit *Unit, quantity int) bool {
	_, span := Tracer.Start(baseContext, "add-unit")
	defer span.End()

	span.SetAttributes(
		attribute.Int("user", int(user.ID)),
		attribute.Int("unit", int(unit.ID)),
		attribute.Int("quantity", quantity),
	)

	found := false
	var temp *UserUnit
	for i := 0; i < len(user.Units) && !found; i++ {
		if user.Units[i].UnitID == unit.ID {
			temp = user.Units[i]
		}
	}

	if temp == nil {
		temp = &UserUnit{
			UserID:   user.ID,
			RoundID:  uint(user.RoundID),
			UnitID:   unit.ID,
			UnitGuid: user.Round.MapUnitsById[unit.ID].GUID,
		}
		user.Units = append(user.Units, temp)
	}

	temp.Quantity += float64(quantity)
	return true
}

func (user *User) AddBuilding(baseContext context.Context, building *Building, quantity float64) bool {
	_, span := Tracer.Start(baseContext, "user.AddBuilding")
	defer span.End()

	span.SetAttributes(
		attribute.Int("user", int(user.ID)),
		attribute.Int("building", int(building.ID)),
		attribute.Float64("quantity", quantity),
	)

	found := false
	var temp *UserBuilding
	for i := 0; i < len(user.Buildings) && !found; i++ {
		if user.Buildings[i].BuildingID == building.ID {
			temp = user.Buildings[i]
		}
	}

	if temp == nil {
		temp = &UserBuilding{
			UserID:       user.ID,
			RoundID:      uint(user.RoundID),
			BuildingID:   building.ID,
			BuildingGuid: user.Round.MapUnitsById[building.ID].GUID,
		}
		user.Buildings = append(user.Buildings, temp)
	}

	temp.Quantity += float64(quantity)
	return true
}

func (user *User) AddBuff(baseContext context.Context, buff *Buff) bool {
	ctx, span := Tracer.Start(baseContext, "User.AddBuff")
	defer span.End()

	log.Warn().Int("user", int(user.ID)).Int("buff", int(buff.ID)).Msg("User.AddBuff")

	span.SetAttributes(
		attribute.Int("user", int(user.ID)),
		attribute.Int("buff", int(buff.ID)),
	)

	found := false
	var temp *UserBuff
	for i := 0; i < len(user.Buffs) && !found; i++ {
		if user.Buffs[i].BuffID == buff.ID {
			temp = user.Buffs[i]
			temp.Stacks = temp.Stacks + 1
			if temp.Stacks > buff.MaxStacks {
				temp.Stacks = buff.MaxStacks
			}
		}
	}

	if temp == nil {
		log.Warn().Msg("Creating new UserBuff: " + fmt.Sprint(getRound(user)))
		temp = &UserBuff{
			UserID:  user.ID,
			RoundID: uint(getRound(user)),
			BuffID:  buff.ID,
			Stacks:  1,
		}

		if buff.Duration != 0 {
			temp.Expires = time.Now()
		} else {
			log.Warn().Msg("Duration is 0")
			if round, err := LoadRoundById(ctx, getRound(user)); err == nil {
				fmt.Println(round)
				temp.Expires = round.Ends

				log.Warn().Msg("Set expires: " + fmt.Sprint(temp.Expires))
			} else {
				log.Error().AnErr("err", err).Msg("Error loading round")
			}
		}

		user.Buffs = append(user.Buffs, temp)
	}

	if buff.Duration != 0 {
		temp.Expires = time.Now().Add(buff.Duration)
	}

	return true
}

func (user *User) RemoveExpiredBuffs(baseContext context.Context) bool {
	ctx, span := Tracer.Start(baseContext, "remove-expired-buffs")
	defer span.End()

	var buffIDs []int
	log.Info().Msg("Get Buff IDs for round: " + fmt.Sprint(user.RoundLoading))
	db.WithContext(ctx).Model(&UserBuff{}).Where("round_id = ? AND ( expires <= ? AND expires <> 0 ) AND user_id = ?", user.RoundLoading, time.Now(), user.ID).Select("buff_id").Scan(&buffIDs)

	for _, buff := range buffIDs {
		log.Warn().Msg("Remove Buff: " + fmt.Sprint(buff))

		found := false
		for i := 0; i < len(user.Buffs) && !found; i++ {
			if user.Buffs[i].BuffID == uint(buff) {
				found = true
				user.Buffs = append(user.Buffs[:i], user.Buffs[i+1:]...)
			}
		}

		if !found {
			log.Warn().Int("buff", buff).Uint("user", user.ID).Msg("Buff not found on user")
		}
	}

	return true
}

func (user *User) TakeResource(ctx context.Context, resource string, amount int) bool {
	ctx, span := Tracer.Start(ctx, "take-resource")
	defer span.End()

	log.Info().Str("resource", resource).Int("amount", amount).Any("user-resource", user.RoundData).Msg("user.TakeResource")

	switch resource {
	case "gold":
		if user.RoundData.Gold < float64(amount) {
			return false
		}
		user.RoundData.Gold -= float64(amount)
	case "food":
		if user.RoundData.Food < float64(amount) {
			return false
		}
		user.RoundData.Food -= float64(amount)
	case "wood":
		if user.RoundData.Wood < float64(amount) {
			return false
		}
		user.RoundData.Wood -= float64(amount)
	case "stone":
		if user.RoundData.Stone < float64(amount) {
			return false
		}
		user.RoundData.Stone -= float64(amount)
	case "metal":
		if user.RoundData.Metal < float64(amount) {
			return false
		}
		user.RoundData.Metal -= float64(amount)
	case "mana":
		if user.RoundData.Mana < float64(amount) {
			return false
		}
		user.RoundData.Mana -= float64(amount)
	case "faith":
		if user.RoundData.Faith < float64(amount) {
			return false
		}
		user.RoundData.Faith -= float64(amount)
	}

	log.Info().Msg("Resource Taken")
	return user.UpdateRound(ctx, nil)
}

func (user *User) GiveResource(ctx context.Context, resource string, amount int) bool {
	ctx, span := Tracer.Start(ctx, "give-resource")
	defer span.End()

	log.Info().Str("resource", resource).Int("amount", amount).Any("user-resource", user.RoundData).Msg("user.GiveResource")

	switch resource {
	case "gold":
		user.RoundData.Gold += float64(amount)
	case "food":
		user.RoundData.Food += float64(amount)
	case "wood":
		user.RoundData.Wood += float64(amount)
	case "stone":
		user.RoundData.Stone += float64(amount)
	case "metal":
		user.RoundData.Metal += float64(amount)
	case "mana":
		user.RoundData.Mana += float64(amount)
	case "faith":
		user.RoundData.Faith += float64(amount)
	}

	log.Info().Msg("Resource Given")
	return user.UpdateRound(ctx, nil)
}

func (user *User) PurchaseTechnology(baseContext context.Context, technology *Technology) bool {
	if technology.Cost == 0 {
		return false
	}

	ctx, span := Tracer.Start(baseContext, "user.PurchaseTechnology")
	defer span.End()

	log.Warn().Int("technology", int(technology.ID)).Msg("user.PurchaseTechnology")

	var tech UserTechnology
	db.WithContext(ctx).Where("user_id = ? AND round_id = ? AND technology_id = ?", user.ID, user.RoundID, technology.ID).Find(&tech)

	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if tech.UserID == 0 {
			tech = UserTechnology{UserID: user.ID, RoundID: uint(user.RoundID), TechnologyID: technology.ID}
			if err := tx.WithContext(ctx).Create(&tech).Error; err != nil {
				return err
			}
		}

		if user.RoundData.Research > float64(technology.Cost) {
			user.RoundData.Research -= float64(technology.Cost)

			if buff, err := LoadBuffById(ctx, int(technology.Buff)); err == nil && buff != nil {
				buff.Dump()
				user.AddBuff(ctx, buff)
			} else {
				log.Error().Msg("Buff not found")
				return errors.New("buff not found")
			}

			if err := tx.WithContext(ctx).Save(&user).Error; err != nil {
				return err
			}

			tech.Level++
			if err := tx.WithContext(ctx).Save(&tech).Error; err != nil {
				return err
			}
		} else {
			log.Warn().Int("current_research", int(user.RoundData.Research)).Int("user", int(user.ID)).Int("technology", int(technology.ID)).Msg("Cannot afford research")
			return errors.New("cannot afford")
		}

		return nil
	}); err == nil {
		return true
	}

	return false
}

func (user *User) BuyResource(baseContext context.Context, quantity uint, resourceGuid uuid.UUID) bool {
	ctx, span := Tracer.Start(baseContext, "user.BuyResource")
	defer span.End()

	span.SetAttributes(
		attribute.Int("user", int(user.ID)),
		attribute.Int("round", user.RoundID),
		attribute.Int("quantity", int(quantity)),
		attribute.String("resource", resourceGuid.String()),
	)

	if round, err := LoadRoundById(ctx, user.RoundID); err == nil {
		resource := round.MarketResources[resourceGuid]
		cost := float64(resource.Value) * float64(quantity)
		name := round.GetResourceById(resource.ResourceID).Name
		if user.RoundData.Gold >= cost {
			user.RoundData.Gold -= cost

			switch name {
			case "food":
				user.RoundData.Food += float64(quantity)
			case "wood":
				user.RoundData.Wood += float64(quantity)
			case "stone":
				user.RoundData.Stone += float64(quantity)
			case "metal":
				user.RoundData.Metal += float64(quantity)
			}

			if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
				tx.WithContext(ctx).Exec("UPDATE round_market_resources SET bought = bought + ? WHERE round_id = ? AND resource_id = ?", quantity, round.ID, resource.ResourceID)

				if err := tx.Save(&user).Error; err != nil {
					return err
				}

				return nil
			}); err == nil {
				go user.Log(fmt.Sprintf("Bought %d %s for %f gold", quantity, name, cost), uint(user.RoundID))
				log.Info().Str("resource", resourceGuid.String()).Int("quantity", int(quantity)).Int("user", int(user.ID)).Msg("Bought resource")
				return true
			}
		}

		go user.Log(fmt.Sprintf("Failed to buy %d %s for %f gold", quantity, name, cost), uint(user.RoundID))
		log.Warn().Str("resource", resourceGuid.String()).Int("quantity", int(quantity)).Int("user", int(user.ID)).Msg("Failed to buy resource")
		return false
	} else {
		log.Warn().AnErr("err", err).Msg("Error loading round")
	}

	log.Warn().Msg("Failed to load round")
	return false
}

func (user *User) SellResource(baseContext context.Context, quantity uint, resourceGuid uuid.UUID) bool {
	ctx, span := Tracer.Start(baseContext, "user.SellResource")
	defer span.End()

	span.SetAttributes(
		attribute.Int("user", int(user.ID)),
		attribute.Int("round", user.RoundID),
		attribute.Int("quantity", int(quantity)),
		attribute.String("resource", resourceGuid.String()),
	)

	if round, err := LoadRoundById(ctx, user.RoundID); err == nil {
		resource := round.MarketResources[resourceGuid]

		cost := float64(1/resource.Value) * float64(quantity)
		user.RoundData.Gold += cost
		name := round.GetResourceById(resource.ResourceID).Name

		switch name {
		case "food":
			if user.RoundData.Food < float64(quantity) {
				go user.Log(fmt.Sprintf("Don't have enough %s to sell %d - only have %f", name, quantity, user.RoundData.Food), uint(user.RoundID))
				return false
			}
			user.RoundData.Food -= float64(quantity)
		case "wood":
			if user.RoundData.Wood < float64(quantity) {
				go user.Log(fmt.Sprintf("Don't have enough %s to sell %d - only have %f", name, quantity, user.RoundData.Wood), uint(user.RoundID))
				return false
			}
			user.RoundData.Wood -= float64(quantity)
		case "stone":
			if user.RoundData.Stone < float64(quantity) {
				go user.Log(fmt.Sprintf("Don't have enough %s to sell %d - only have %f", name, quantity, user.RoundData.Stone), uint(user.RoundID))
				return false
			}
			user.RoundData.Stone -= float64(quantity)
		case "metal":
			if user.RoundData.Metal < float64(quantity) {
				go user.Log(fmt.Sprintf("Don't have enough %s to sell %d - only have %f", name, quantity, user.RoundData.Metal), uint(user.RoundID))
				return false
			}
			user.RoundData.Metal -= float64(quantity)
		}

		if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			tx.WithContext(ctx).Exec("UPDATE round_market_resources SET sold = sold + ? WHERE round_id = ? AND resource_id = ?", quantity, round.ID, resource.ResourceID)

			if err := tx.Save(&user).Error; err != nil {
				return err
			}

			return nil
		}); err == nil {
			go user.Log(fmt.Sprintf("Sold %d %s for %f gold", quantity, name, cost), uint(user.RoundID))
			log.Info().Str("resource", resourceGuid.String()).Int("quantity", int(quantity)).Int("user", int(user.ID)).Msg("Sold resource")
			return true
		}

		go user.Log(fmt.Sprintf("Failed to sell %d %s for %f gold", quantity, name, cost), uint(user.RoundID))
		log.Warn().Str("resource", resourceGuid.String()).Int("quantity", int(quantity)).Int("user", int(user.ID)).Msg("Failed to sell resource")
		return false
	}

	log.Warn().Msg("Failed to load round")
	return false
}

func GetUserIdForName(ctx context.Context, name string) uint {
	var user *User
	if err := db.WithContext(ctx).Where("username = ?", name).First(&user).Error; err != nil {
		log.Warn().Err(err).Str("name", name).Msg("GetUserIdForName: No user found")
		return 0
	}

	return user.ID
}

func GetUserIdForGuid(ctx context.Context, guid uuid.UUID) uint {
	ctx, span := utils.StartSpan(ctx, "User.GetUserIdForGuid")
	defer span.End()

	span.SetAttributes(attribute.String("guid", guid.String()))

	var userId *uint
	if err := db.WithContext(ctx).Table("users").Select("id").Where("guid = ?", guid).Scan(&userId).Error; err != nil {
		log.Warn().Err(err).Str("guid", guid.String()).Msg("GetUserIdForGuid: No user found")
		return 0
	}

	return *userId
}

func LoadUserForRound(userid int, roundid int) *User {
	log.Debug().
		Int("userid", userid).
		Int("roundid", roundid).
		Msg("LoadForRound")

	ctx, sp := Tracer.Start(context.Background(), "loading-user")
	defer sp.End()

	user := &User{RoundLoading: roundid}
	user.ID = uint(userid)
	if err := db.WithContext(ctx).First(&user).Error; err != nil {
		return nil
	}

	return user
}
