package models

import "github.com/google/uuid"

type UserItem struct {
	BaseModel

	UserID   uint      `json:"-"`
	ItemID   uint      `json:"-"`
	ItemGuid uuid.UUID `gorm:"->;-:migration" json:"guid"`
	Quantity uint      `json:"quantity"`
}
