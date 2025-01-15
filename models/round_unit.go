package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoundUnit struct {
	BaseModel

	GUID            uuid.UUID `gorm:"uniqueIndex,size:36" json:"guid"`
	RoundID         uint      `json:"-"`
	UnitID          uint      `json:"unit_id"`
	Attack          uint      `gorm:"default:1" json:"attack"`
	Defense         uint      `gorm:"default:1" json:"defense"`
	Power           uint      `gorm:"default:1" json:"power"`
	Health          uint      `gorm:"default:1" json:"health"`
	Ranged          bool      `gorm:"default:false" json:"ranged"`
	CostGold        uint      `gorm:"default:1" json:"cost_gold"`
	CostPoints      uint      `gorm:"default:1" json:"cost_points"`
	CostFood        uint      `gorm:"default:0" json:"cost_food"`
	CostWood        uint      `gorm:"default:0" json:"cost_wood"`
	CostMetal       uint      `gorm:"default:0" json:"cost_metal"`
	CostStone       uint      `gorm:"default:0" json:"cost_stone"`
	CostFaith       uint      `gorm:"default:0" json:"cost_faith"`
	CostMana        uint      `gorm:"default:0" json:"cost_mana"`
	UpkeepGold      uint      `gorm:"default:0" json:"upkeep_gold"`
	UpkeepFood      uint      `gorm:"default:0" json:"upkeep_food"`
	UpkeepWood      uint      `gorm:"default:0" json:"upkeep_wood"`
	UpkeepStone     uint      `gorm:"default:0" json:"upkeep_stone"`
	UpkeepMetal     uint      `gorm:"default:0" json:"upkeep_metal"`
	UpkeepFaith     uint      `gorm:"default:0" json:"upkeep_faith"`
	UpkeepMana      uint      `gorm:"default:0" json:"upkeep_mana"`
	Available       bool      `gorm:"default:false" json:"available"`
	Recruitable     bool      `gorm:"default:false" json:"recruitable"`
	SupportsPartial bool      `gorm:"default:false" json:"supports_partial"`
	StartWith       uint      `gorm:"default:0" json:"start_with"`
}

func (unit *RoundUnit) BeforeCreate(tx *gorm.DB) (err error) {
	unit.GUID = uuid.New()
	return
}
