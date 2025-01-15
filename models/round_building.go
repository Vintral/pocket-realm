package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoundBuilding struct {
	BaseModel

	GUID            uuid.UUID `gorm:"uniqueIndex,size:36" json:"guid"`
	BuildingID      uint      `json:"building_id"`
	RoundID         uint      `json:"-"`
	CostPoints      uint      `gorm:"default:0" json:"cost_points"`
	CostWood        uint      `gorm:"default:0" json:"cost_wood"`
	CostStone       uint      `gorm:"default:0" json:"cost_stone"`
	CostGold        uint      `gorm:"default:0" json:"cost_gold"`
	CostFood        uint      `gorm:"default:0" json:"cost_food"`
	CostMetal       uint      `gorm:"default:0" json:"cost_metal"`
	CostFaith       uint      `gorm:"default:0" json:"cost_faith"`
	CostMana        uint      `gorm:"default:0" json:"cost_mana"`
	BonusValue      uint      `gorm:"default:0" json:"bonus_value"`
	UpkeepGold      uint      `gorm:"default:0" json:"upkeep_gold"`
	UpkeepFood      uint      `gorm:"default:0" json:"upkeep_food"`
	UpkeepWood      uint      `gorm:"default:0" json:"upkeep_wood"`
	UpkeepStone     uint      `gorm:"default:0" json:"upkeep_stone"`
	UpkeepMetal     uint      `gorm:"default:0" json:"upkeep_metal"`
	UpkeepFaith     uint      `gorm:"default:0" json:"upkeep_faith"`
	UpkeepMana      uint      `gorm:"default:0" json:"upkeep_mana"`
	Buildable       bool      `gorm:"default:false" json:"buildable"`
	Available       bool      `gorm:"default:false" json:"available"`
	SupportsPartial bool      `gorm:"default:false" json:"supports_partial"`
	StartWith       uint      `gorm:"default:0" json:"start_with"`
}

func (building *RoundBuilding) BeforeCreate(tx *gorm.DB) (err error) {
	building.GUID = uuid.New()
	return
}
