package models

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/Vintral/pocket-realm/utilities"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type Building struct {
	BaseModel

	ID              uint      `gorm:"primaryKey" json:"order"`
	GUID            uuid.UUID `gorm:"uniqueIndex,size:36" json:"guid"`
	Name            string    `json:"name"`
	CostPoints      uint      `gorm:"->;-:migration" json:"cost_points"`
	CostWood        uint      `gorm:"->;-:migration" json:"cost_wood"`
	CostStone       uint      `gorm:"->;-:migration" json:"cost_stone"`
	CostGold        uint      `gorm:"->;-:migration" json:"cost_gold"`
	CostFood        uint      `gorm:"->;-:migration" json:"cost_food"`
	CostMetal       uint      `gorm:"->;-:migration" json:"cost_metal"`
	CostFaith       uint      `gorm:"->;-:migration" json:"cost_faith"`
	CostMana        uint      `gorm:"->;-:migration" json:"cost_mana"`
	UpkeepGold      uint      `gorm:"->;-:migration" json:"upkeep_gold"`
	UpkeepFood      uint      `gorm:"->;-:migration" json:"upkeep_food"`
	UpkeepWood      uint      `gorm:"->;-:migration" json:"upkeep_wood"`
	UpkeepMetal     uint      `gorm:"->;-:migration" json:"upkeep_metal"`
	UpkeepFaith     uint      `gorm:"->;-:migration" json:"upkeep_faith"`
	UpkeepStone     uint      `gorm:"->;-:migration" json:"upkeep_stone"`
	UpkeepMana      uint      `gorm:"->;-:migration" json:"upkeep_mana"`
	BonusField      string    `json:"bonus_field"`
	BonusValue      uint      `gorm:"->;-:migration" json:"bonus_value"`
	Available       bool      `gorm:"->;-:migration" json:"available"`
	Buildable       bool      `gorm:"->;-:migration" json:"buildable"`
	SupportsPartial bool      `gorm:"->;-:migration" json:"supports_partial"`
	StartWith       uint      `gorm:"->;-:migration" json:"start_with"`
}

func (building *Building) BeforeCreate(tx *gorm.DB) (err error) {
	building.GUID = uuid.New()
	return
}

func (building *Building) getBuildAmount(user *User, energy uint) float64 {
	var amount = (user.RoundData.BuildPower * float64(energy)) / float64(building.CostPoints)
	fmt.Println("Amount:", utilities.RoundFloat(amount, 2))

	if amount > 1 {
		amount = math.Floor(amount)
	}

	return amount
}

func (building *Building) canAffordBuild(user *User, amount float64) error {
	user.Dump()

	if user.RoundData.Gold < float64(building.CostGold)*amount {
		return errors.New("build-not-enough-gold")
	}
	if user.RoundData.Food < float64(building.CostFood)*amount {
		return errors.New("build-not-enough-food")
	}
	if user.RoundData.Wood < float64(building.CostWood)*amount {
		return errors.New("build-not-enough-wood")
	}
	if user.RoundData.Stone < float64(building.CostStone)*amount {
		return errors.New("build-not-enough-stone")
	}
	if user.RoundData.Metal < float64(building.CostMetal)*amount {
		return errors.New("build-not-enough-metal")
	}
	if user.RoundData.Faith < float64(building.CostFaith)*amount {
		return errors.New("build-not-enough-faith")
	}
	if user.RoundData.Mana < float64(building.CostMana)*amount {
		return errors.New("build-not-enough-mana")
	}

	return nil
}

func (building *Building) canBuild(user *User, energy uint) (float64, error) {
	if !building.Buildable {
		return 0, errors.New("building-not-buildable")
	}

	if energy > uint(user.RoundData.Energy) {
		return 0, errors.New("building-not-enough-energy")
	}

	amount := building.getBuildAmount(user, energy)
	if err := building.canAffordBuild(user, amount); err != nil {
		return 0, err
	}

	return amount, nil
}

func (building *Building) takeCost(user *User, amount float64) {
	fmt.Println("takeCost")

	if building.CostGold > 0 {
		user.RoundData.Gold -= float64(building.CostGold) * amount
	}
	if building.CostFood > 0 {
		user.RoundData.Food -= float64(building.CostFood) * amount
	}
	if building.CostWood > 0 {
		user.RoundData.Wood -= float64(building.CostWood) * amount
	}
	if building.CostMetal > 0 {
		user.RoundData.Metal -= float64(building.CostMetal) * amount
	}
	if building.CostStone > 0 {
		user.RoundData.Stone -= float64(building.CostStone) * amount
	}
	if building.CostFaith > 0 {
		user.RoundData.Faith -= float64(building.CostFaith) * amount
	}
	if building.CostMana > 0 {
		user.RoundData.Mana -= float64(building.CostMana) * amount
	}
}

func (building *Building) Build(ctx context.Context, user *User, energy uint) (float64, error) {
	var err error
	var amount float64

	fmt.Println("We in Build")

	if amount, err = building.canBuild(user, energy); err == nil {
		fmt.Println("DO BUILD HERE", amount)

		found := false
		var temp *UserBuilding
		for i := 0; i < len(user.Buildings) && !found; i++ {
			fmt.Println("Building", i)
			if user.Buildings[i].BuildingID == building.ID {
				temp = user.Buildings[i]
			}
		}
		if temp == nil {
			fmt.Println("Create New UserBuilding")
			temp = &UserBuilding{
				UserID:       user.ID,
				RoundID:      uint(user.RoundID),
				BuildingID:   building.ID,
				BuildingGuid: user.Round.MapBuildingsById[building.ID].GUID,
			}
			user.Buildings = append(user.Buildings, temp)
		}

		fmt.Println("Energy:", energy)
		temp.Quantity += amount
		if !user.updateBuildings(ctx, nil) {
			return 0, errors.New("error-updating-buildings")
		}

		building.takeCost(user, amount)
		user.RoundData.Energy -= int(energy)
		user.Dump()

		if !user.UpdateRound(ctx, nil) {
			fmt.Println(" oh noes")
			temp.Quantity -= amount
			if !user.updateBuildings(ctx, nil) {
				fmt.Println("Rolled back")
				return 0, errors.New("error-removing-buildings")
			}
		}

		return amount, nil
	}

	fmt.Println("After Build")

	user.Load()
	return 0, err
}

func (building *Building) Dump() {
	log.Trace().Msg("=============================")
	log.Trace().Msg("ID:" + fmt.Sprint(building.ID))
	log.Trace().Msg("GUID:" + fmt.Sprint(building.GUID))
	log.Trace().Msg("Name:" + building.Name)
	log.Trace().Msg("CostPoints:" + fmt.Sprint(building.CostPoints))
	log.Trace().Msg("CostGold:" + fmt.Sprint(building.CostGold))
	log.Trace().Msg("CostFood:" + fmt.Sprint(building.CostFood))
	log.Trace().Msg("CostWood:" + fmt.Sprint(building.CostWood))
	log.Trace().Msg("CostMetal:" + fmt.Sprint(building.CostMetal))
	log.Trace().Msg("CostStone:" + fmt.Sprint(building.CostStone))
	log.Trace().Msg("CostFaith:" + fmt.Sprint(building.CostFaith))
	log.Trace().Msg("CostMana:" + fmt.Sprint(building.CostMana))
	log.Trace().Msg("BonusField:" + fmt.Sprint(building.BonusField))
	log.Trace().Msg("BonusValue:" + fmt.Sprint(building.BonusValue))
	log.Trace().Msg("UpkeepGold:" + fmt.Sprint(building.UpkeepGold))
	log.Trace().Msg("UpkeepFood:" + fmt.Sprint(building.UpkeepFood))
	log.Trace().Msg("UpkeepWood:" + fmt.Sprint(building.UpkeepWood))
	log.Trace().Msg("UpkeepMetal:" + fmt.Sprint(building.UpkeepMetal))
	log.Trace().Msg("UpkeepStone:" + fmt.Sprint(building.UpkeepStone))
	log.Trace().Msg("UpkeepFaith:" + fmt.Sprint(building.UpkeepFaith))
	log.Trace().Msg("UpkeepMana:" + fmt.Sprint(building.UpkeepMana))
	log.Trace().Msg("Available:" + fmt.Sprint(building.Available))
	log.Trace().Msg("Buildable:" + fmt.Sprint(building.Buildable))
	log.Trace().Msg("SupportsPartial:" + fmt.Sprint(building.SupportsPartial))
	log.Trace().Msg("=============================")
}
