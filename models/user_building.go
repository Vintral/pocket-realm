package models

import "github.com/google/uuid"

type UserBuilding struct {
	BaseModel

	UserID       uint      `json:"-"`
	BuildingID   uint      `json:"-"`
	BuildingGuid uuid.UUID `gorm:"->;-:migration" json:"guid"`
	RoundID      uint      `json:"-"`
	Quantity     float64   `json:"quantity"`
}
