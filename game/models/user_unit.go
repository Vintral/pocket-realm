package models

import "github.com/google/uuid"

type UserUnit struct {
	BaseModel

	UserID   uint      `json:"-"`
	UnitID   uint      `json:"-"`
	UnitGuid uuid.UUID `gorm:"->;-:migration" json:"guid"`
	RoundID  uint      `json:"-"`
	Quantity float64   `json:"quantity"`
}
