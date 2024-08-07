package models

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"time"

	"github.com/Vintral/pocket-realm/game/payloads"
	"github.com/Vintral/pocket-realm/utilities"
	"github.com/rs/zerolog/log"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ContextKey string
type KeyUser struct{}

var rounds = make(map[uuid.UUID]*Round)
var roundsById = make(map[int]*Round)

var activeRounds []*Round
var pastRounds []*Round
var rankingsByRoundId = make(map[int][]*Ranking)

type Round struct {
	BaseModel

	GUID             uuid.UUID            `gorm:"uniqueIndex,size:36" json:"guid"`
	EnergyMax        uint                 `gorm:"default:250" json:"energy_max"`
	EnergyRegen      uint                 `gorm:"default:10" json:"energy_regen"`
	Ends             time.Time            `json:"ends"`
	Resources        []*Resource          `gorm:"-" json:"resources"`
	MapResources     map[string]*Resource `gorm:"-" json:"-"`
	MapResourcesById map[uint]*Resource   `gorm:"-" json:"-"`
	Units            []*Unit              `gorm:"-" json:"units"`
	MapUnits         map[string]*Unit     `gorm:"-" json:"-"`
	MapUnitsById     map[uint]*Unit       `gorm:"-" json:"-"`
	Buildings        []*Building          `gorm:"-" json:"buildings"`
	MapBuildings     map[string]*Building `gorm:"-" json:"-"`
	MapBuildingsById map[uint]*Building   `gorm:"-" json:"-"`
	Top              []*Ranking           `gorm:"-" json:"top"`
	User             []*Ranking           `gorm:"-" json:"finished"`
	Tick             uint                 `gorm:"default:5" json:"tick"`
}

func (round *Round) BeforeCreate(tx *gorm.DB) (err error) {
	round.GUID = uuid.New()
	return
}

func (round *Round) AfterFind(tx *gorm.DB) (err error) {
	log.Trace().Msg("Round: AfterFind")

	ctx, sp := Tracer.Start(tx.Statement.Context, "after-find")
	defer sp.End()

	wg := new(sync.WaitGroup)
	wg.Add(3)
	go round.loadResources(ctx, wg)
	go round.loadUnits(ctx, wg)
	go round.loadBuildings(ctx, wg)
	// go user.loadItems(ctx, wg)
	wg.Wait()

	return
}

func (round *Round) loadBuildings(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Info().Msg("loadBuildings")

	db.WithContext(ctx).Raw(`
		SELECT 
			buildings.id, buildings.name, round_buildings.created_at, round_buildings.updated_at, round_buildings.deleted_at, round_buildings.guid, 
			round_buildings.building_id, round_buildings.cost_wood, round_buildings.cost_stone, round_buildings.cost_points, 
			round_buildings.cost_gold, round_buildings.cost_metal, round_buildings.cost_faith, round_buildings.cost_mana,
			round_buildings.cost_food, buildings.bonus_field, round_buildings.bonus_value, round_buildings.available,
			round_buildings.upkeep_gold, round_buildings.upkeep_food, round_buildings.upkeep_wood, round_buildings.upkeep_faith, 
			round_buildings.upkeep_metal, round_buildings.upkeep_stone, round_buildings.upkeep_mana, round_buildings.buildable
		FROM 
			round_buildings 
		INNER JOIN 
			( 
				SELECT building_id, MAX(round_id) AS round_id 
				FROM round_buildings
				WHERE round_id = 0 OR round_id = ` + fmt.Sprint(round.ID) + ` 
				GROUP BY building_id				
				ORDER BY building_id DESC 
			) AS A 
		ON 
			A.round_id = round_buildings.round_id
		INNER JOIN 
			buildings
		ON 
			buildings.id = round_buildings.building_id			
		WHERE round_buildings.building_id = A.building_id`,
	).Scan(&round.Buildings)

	round.MapBuildings = make(map[string]*Building)
	round.MapBuildingsById = make(map[uint]*Building)
	for _, building := range round.Buildings {
		log.Debug().
			Str("guid", building.GUID.String()).
			Int("id", int(building.ID)).
			Msg("Saved: " + building.Name)

		round.MapBuildings[building.GUID.String()] = building
		round.MapBuildingsById[building.ID] = building

		building.Dump()
	}
}

func (round *Round) loadUnits(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Info().Msg("loadUnits")

	db.WithContext(ctx).Raw(`
		SELECT 
			units.id, units.name, round_units.created_at, round_units.updated_at, round_units.deleted_at, round_units.guid, round_units.unit_id,
			round_units.attack, round_units.defense, round_units.power, round_units.health, round_units.ranged, round_units.cost_gold, 
			round_units.cost_points, round_units.cost_food, round_units.cost_wood, round_units.cost_stone, round_units.cost_metal, round_units.cost_mana, 
			round_units.cost_faith, round_units.upkeep_gold, round_units.upkeep_food, round_units.upkeep_wood, round_units.upkeep_faith, 
			round_units.upkeep_metal, round_units.upkeep_stone, round_units.upkeep_mana, round_units.available, round_units.recruitable
		FROM 
			round_units 
		INNER JOIN 
			( 
				SELECT unit_id, MAX(round_id) AS round_id 
				FROM round_units
				WHERE round_id = 0 OR round_id = ` + fmt.Sprint(round.ID) + ` 
				GROUP BY unit_id				
				ORDER BY unit_id DESC 
			) AS A 
		ON 
			A.round_id = round_units.round_id
		INNER JOIN 
			units
		ON 
			units.id = round_units.unit_id			
		WHERE round_units.unit_id = A.unit_id`,
	).Scan(&round.Units)

	round.MapUnits = make(map[string]*Unit)
	round.MapUnitsById = make(map[uint]*Unit)
	for _, unit := range round.Units {
		log.Debug().
			Str("guid", unit.GUID.String()).
			Int("id", int(unit.ID)).
			Msg("Saved: " + unit.Name)

		round.MapUnits[unit.GUID.String()] = unit
		round.MapUnitsById[unit.ID] = unit
	}
}

func (round *Round) loadResources(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Info().Msg("loadResources")

	db.WithContext(ctx).Raw(`
		SELECT 
			resources.id, resources.name, round_resources.created_at, round_resources.updated_at, round_resources.deleted_at, round_resources.guid, round_resources.resource_id,
			round_resources.can_gather, round_resources.can_market 
		FROM 
			round_resources 
		INNER JOIN 
			( 
				SELECT resource_id, MAX(round_id) AS round_id 
				FROM round_resources
				WHERE round_id = 0 OR round_id = ` + fmt.Sprint(round.ID) + ` 
				GROUP BY resource_id				
				ORDER BY resource_id DESC 
			) AS A 
		ON 
			A.round_id = round_resources.round_id
		INNER JOIN 
			resources
		ON 
			resources.id = round_resources.resource_id			
		WHERE round_resources.resource_id = A.resource_id`,
	).Scan(&round.Resources)

	round.MapResources = make(map[string]*Resource)
	for _, r := range round.Resources {
		log.Debug().
			Str("guid", r.GUID.String()).
			Int("id", int(r.ID)).
			Msg("Saved: " + r.Name)

		round.MapResources[r.GUID.String()] = r
	}
}

func (round *Round) Load(packet []byte) ([]byte, error) {
	ctx, span := Tracer.Start(context.Background(), "load-round-data")
	defer span.End()

	var payload payloads.RoundDataPayload
	err := json.Unmarshal(packet, &payload)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	fmt.Println(payload)

	fmt.Println("Round: Load -", payload.Round)

	r, err := LoadRoundByGuid(ctx, payload.Round)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return json.Marshal(r)
}

func (round *Round) GetResourceByGuid(guid string) *Resource {
	log.Debug().Msg("GetResourceByGuid:" + guid)

	if resource, ok := round.MapResources[guid]; ok {
		return resource
	}

	return nil
}

func (round *Round) GetUnitByGuid(guid string) *Unit {
	log.Debug().Msg("GetUnitByGuid:" + guid)

	if unit, ok := round.MapUnits[guid]; ok {
		return unit
	}

	return nil
}

func (round *Round) GetUnitById(id uint) *Unit {
	log.Debug().Msg("GetUnitById:" + fmt.Sprint(id))

	if unit, ok := round.MapUnitsById[id]; ok {
		return unit
	}

	return nil
}

func (round *Round) GetBuildingByGuid(guid string) *Building {
	log.Debug().Msg("GetBuildingByGuid:" + guid)

	if building, ok := round.MapBuildings[guid]; ok {
		return building
	}

	return nil
}

func (round *Round) CanGather(guid string) bool {
	fmt.Println("CanGather:", guid)

	if resource, ok := round.MapResources[guid]; ok {
		return resource.CanGather
	}

	return false
}

func (round *Round) LoadUserResults(user *User) {
	log.Debug().Msg("LoadUserResults: " + fmt.Sprint(round.ID) + " -- " + fmt.Sprint(user.ID))
}

func (round *Round) GetRankings(user *User) (top []*Ranking, personal *Ranking) {
	log.Debug().Msg("GetRankings: " + fmt.Sprint(round.ID) + " -- " + fmt.Sprint(user.ID))

	return nil, nil
}

func LoadRoundById(ctx context.Context, roundID int) (*Round, error) {
	r := roundsById[roundID]
	if r != nil {
		return r, nil
	}

	log.Info().Int("round_id", roundID).Msg("Loading Round")

	var round Round
	if err := db.WithContext(ctx).Where("id = ?", roundID).Find(&round).Error; err != nil {
		fmt.Println("Error loading round")
		fmt.Println(err)
		return nil, err
	}

	log.Debug().Any("round", round).Send()

	roundsById[roundID] = &round
	rounds[round.GUID] = &round
	return &round, nil
}

func LoadRoundByGuid(ctx context.Context, guid uuid.UUID) (*Round, error) {
	r := rounds[guid]
	if r != nil {
		return r, nil
	}

	log.Info().Msg("Loading Round: " + guid.String())

	var round Round
	if err := db.WithContext(ctx).Where("guid = ?", guid).Find(&round).Error; err != nil {
		fmt.Println("Error loading round")
		fmt.Println(err)
		return nil, err
	}

	log.Debug().Any("round", round).Send()

	rounds[guid] = &round
	roundsById[int(round.ID)] = &round
	return &round, nil
}

func LoadRoundForUser(base context.Context) {
	ctx, span := utilities.StartSpan(base, "load-round-by-user")
	defer span.End()

	user := base.Value(utilities.KeyUser{}).(*User)

	if round, err := LoadRoundById(ctx, user.RoundID); err == nil {
		user.Connection.WriteJSON(struct {
			Type  string `json:"type"`
			Round *Round `json:"round"`
		}{
			Type:  "ROUND",
			Round: round,
		})

		user.Round = round
	} else {
		user.SendError(SendErrorParams{Context: &ctx, Type: "round", Message: "round-0"})
	}
}

func ResetActiveRounds(baseContext context.Context) {
	_, span := Tracer.Start(baseContext, "reset-active-rounds")
	defer span.End()

	activeRounds = nil
}

func GetPastRounds(baseContext context.Context, user *User, c chan []*Round) {
	log.Debug().Msg("GetPastRounds")

	if pastRounds != nil {
		log.Debug().Msg("Re-using past rounds")
	} else {
		ctx, span := Tracer.Start(baseContext, "get-past-rounds")
		defer span.End()

		if err := db.WithContext(ctx).Model(&Round{}).Where("ends < ?", time.Now()).Scan(&pastRounds).Error; err != nil {
			log.Warn().Err(err).Msg("Error loading rounds")
			c <- nil
		}
	}

	c <- pastRounds
}

func GetActiveRoundsForTick(baseContext context.Context, tick int) []*Round {
	log.Debug().Msg("GetActiveRoundsForTick")

	ctx, span := Tracer.Start(baseContext, "get-active-rounds-for-tick")
	defer span.End()

	c := make(chan []*Round)
	go GetActiveRounds(ctx, c)
	rounds := <-c

	var ret []*Round
	for _, r := range rounds {
		log.Debug().Msg(fmt.Sprintf("%d - %d = %t", r.Tick, tick, tick%int(r.Tick) == 0))

		if tick == 0 || tick%int(r.Tick) == 0 {
			ret = append(ret, r)
		}
	}

	return ret
}

func GetActiveRounds(baseContext context.Context, c chan []*Round) {
	if activeRounds != nil {
		log.Debug().Msg("Re-using active rounds")
	} else {
		ctx, span := Tracer.Start(baseContext, "get-active-rounds")
		defer span.End()

		if err := db.WithContext(ctx).Model(&Round{}).Where("ends > ?", time.Now()).Find(&activeRounds).Error; err != nil {
			log.Warn().Err(err).Msg("Error loading rounds")
			c <- nil
		}
	}

	c <- activeRounds
}
