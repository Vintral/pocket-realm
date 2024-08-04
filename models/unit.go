package models

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/google/uuid"
	"gorm.io/gorm"

	utils "github.com/Vintral/pocket-realm//utilities"
)

type Unit struct {
	BaseModel

	ID              uint      `gorm:"primaryKey" json:"order"`
	GUID            uuid.UUID `gorm:"uniqueIndex,size:36" json:"guid"`
	Name            string    `json:"name"`
	Attack          uint      `gorm:"->;-:migration" json:"attack"`
	Defense         uint      `gorm:"->;-:migration" json:"defense"`
	Power           uint      `gorm:"->;-:migration" json:"power"`
	Health          uint      `gorm:"->;-:migration" json:"health"`
	Ranged          bool      `gorm:"->;-:migration" json:"ranged"`
	CostPoints      uint      `gorm:"->;-:migration" json:"cost_points"`
	CostGold        uint      `gorm:"->;-:migration" json:"cost_gold"`
	CostFood        uint      `gorm:"->;-:migration" json:"cost_food"`
	CostWood        uint      `gorm:"->;-:migration" json:"cost_wood"`
	CostMetal       uint      `gorm:"->;-:migration" json:"cost_metal"`
	CostStone       uint      `gorm:"->;-:migration" json:"cost_stone"`
	CostFaith       uint      `gorm:"->;-:migration" json:"cost_faith"`
	CostMana        uint      `gorm:"->;-:migration" json:"cost_mana"`
	UpkeepGold      uint      `gorm:"->;-:migration" json:"upkeep_gold"`
	UpkeepFood      uint      `gorm:"->;-:migration" json:"upkeep_food"`
	UpkeepWood      uint      `gorm:"->;-:migration" json:"upkeep_wood"`
	UpkeepMetal     uint      `gorm:"->;-:migration" json:"upkeep_metal"`
	UpkeepFaith     uint      `gorm:"->;-:migration" json:"upkeep_faith"`
	UpkeepStone     uint      `gorm:"->;-:migration" json:"upkeep_stone"`
	UpkeepMana      uint      `gorm:"->;-:migration" json:"upkeep_mana"`
	Available       bool      `gorm:"->;-:migration" json:"available"`
	Recruitable     bool      `gorm:"->;-:migration" json:"recruitable"`
	SupportsPartial bool      `gorm:"->;-:migration" json:"supports_partial"`
}

func (unit *Unit) BeforeCreate(tx *gorm.DB) (err error) {
	unit.GUID = uuid.New()
	return
}

func (unit *Unit) getRecruitAmount(user *User, energy uint) float64 {
	var amount = (user.RoundData.RecruitPower * float64(energy)) / float64(unit.CostPoints)
	fmt.Println("Amount:", utils.RoundFloat(amount, 2))

	if amount > 1 {
		amount = math.Floor(amount)
	}

	return amount
}

func (unit *Unit) canAffordRecruit(user *User, amount float64) error {
	if user.RoundData.Gold < float64(unit.CostGold)*amount {
		return errors.New("recruit-not-enough-gold")
	}
	if user.RoundData.Food < float64(unit.CostFood)*amount {
		return errors.New("recruit-not-enough-food")
	}
	if user.RoundData.Wood < float64(unit.CostWood)*amount {
		return errors.New("recruit-not-enough-wood")
	}
	if user.RoundData.Stone < float64(unit.CostStone)*amount {
		return errors.New("recruit-not-enough-stone")
	}
	if user.RoundData.Metal < float64(unit.CostMetal)*amount {
		return errors.New("recruit-not-enough-metal")
	}
	if user.RoundData.Faith < float64(unit.CostFaith)*amount {
		return errors.New("recruit-not-enough-faith")
	}
	if user.RoundData.Mana < float64(unit.CostMana)*amount {
		return errors.New("recruit-not-enough-mana")
	}

	return nil
}

func (unit *Unit) canRecruit(user *User, energy uint) (float64, error) {
	fmt.Println("canRecruit")
	fmt.Println("Energy:", energy)
	fmt.Println("User:", user)

	if !unit.Recruitable {
		return 0, errors.New("recruit-not-recruitable")
	}

	if energy > uint(user.RoundData.Energy) {
		return 0, errors.New("recruit-not-enough-energy")
	}

	amount := unit.getRecruitAmount(user, energy)
	if err := unit.canAffordRecruit(user, amount); err != nil {
		return 0, err
	}

	return amount, nil
}

func (unit *Unit) takeCost(user *User, amount float64) {
	fmt.Println("takeCost")

	if unit.CostGold > 0 {
		user.RoundData.Gold -= float64(unit.CostGold) * amount
	}
	if unit.CostFood > 0 {
		user.RoundData.Food -= float64(unit.CostFood) * amount
	}
	if unit.CostWood > 0 {
		user.RoundData.Wood -= float64(unit.CostWood) * amount
	}
	if unit.CostMetal > 0 {
		user.RoundData.Metal -= float64(unit.CostMetal) * amount
	}
	if unit.CostStone > 0 {
		user.RoundData.Stone -= float64(unit.CostStone) * amount
	}
	if unit.CostFaith > 0 {
		user.RoundData.Faith -= float64(unit.CostFaith) * amount
	}
	if unit.CostMana > 0 {
		user.RoundData.Mana -= float64(unit.CostMana) * amount
	}
}

func (unit *Unit) Recruit(ctx context.Context, user *User, energy uint) (float64, error) {
	var err error
	var amount float64

	if amount, err = unit.canRecruit(user, energy); err == nil {
		fmt.Println("DO RECRUIT HERE", amount)

		found := false
		var temp *UserUnit
		for i := 0; i < len(user.Units) && !found; i++ {
			fmt.Print("Unit", i)
			if user.Units[i].UnitID == unit.ID {
				temp = user.Units[i]
			}
		}
		if temp == nil {
			fmt.Println("Create New UserUnit")
			temp = &UserUnit{
				UserID:   user.ID,
				RoundID:  uint(user.RoundID),
				UnitID:   unit.ID,
				UnitGuid: user.Round.MapUnitsById[unit.ID].GUID,
			}
			user.Units = append(user.Units, temp)
		}

		temp.Quantity += amount
		if !user.updateUnits(ctx, nil) {
			return 0, errors.New("error-updating-units")
		}

		unit.takeCost(user, amount)
		user.RoundData.Energy -= int(energy)

		if !user.UpdateRound(ctx, nil) {
			fmt.Println(" oh noes")
			temp.Quantity -= amount
			if !user.updateUnits(ctx, nil) {
				fmt.Println("Rolled back")
				return 0, errors.New("error-removing-units")
			}
		}

		return amount, nil
	}

	return 0, err
}

func (unit *Unit) GetUpkeep(field string) uint {
	switch field {
	case "gold":
		return unit.UpkeepGold
	case "food":
		return unit.UpkeepFood
	case "wood":
		return unit.UpkeepWood
	case "stone":
		return unit.UpkeepStone
	case "metal":
		return unit.UpkeepMetal
	case "faith":
		return unit.UpkeepFaith
	case "mana":
		return unit.UpkeepMana
	}

	return 0
}

func (unit *Unit) Dump() {
	fmt.Println("=============================")
	fmt.Println("ID:", unit.ID)
	fmt.Println("GUID:", unit.GUID)
	fmt.Println("Name:", unit.Name)
	fmt.Println("Attack:", unit.Attack)
	fmt.Println("Defense:", unit.Defense)
	fmt.Println("Power:", unit.Power)
	fmt.Println("Health:", unit.Health)
	fmt.Println("Ranged:", unit.Ranged)
	fmt.Println("CostPoints:", unit.CostPoints)
	fmt.Println("CostGold:", unit.CostGold)
	fmt.Println("CostFood:", unit.CostFood)
	fmt.Println("CostWood:", unit.CostWood)
	fmt.Println("CostMetal:", unit.CostMetal)
	fmt.Println("CostStone:", unit.CostStone)
	fmt.Println("CostFaith:", unit.CostFaith)
	fmt.Println("CostMana:", unit.CostMana)
	fmt.Println("UpkeepGold:", unit.UpkeepGold)
	fmt.Println("UpkeepFood:", unit.UpkeepFood)
	fmt.Println("UpkeepWood:", unit.UpkeepWood)
	fmt.Println("UpkeepMetal:", unit.UpkeepMetal)
	fmt.Println("UpkeepFaith:", unit.UpkeepFaith)
	fmt.Println("UpkeepStone:", unit.UpkeepStone)
	fmt.Println("UpkeepMana:", unit.UpkeepMana)
	fmt.Println("Available:", unit.Available)
	fmt.Println("Recruitable:", unit.Recruitable)
	fmt.Println("SupportsPartial:", unit.SupportsPartial)
	fmt.Println("=============================")
}
