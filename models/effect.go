package models

type Effect struct {
	BaseModel

	ItemID uint   `json:"-"`
	Type   string `json:"type"`
	Name   string `json:"name"`
	Amount uint   `json:"amount"`
}
