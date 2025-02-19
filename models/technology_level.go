package models

type TechnologyLevel struct {
	BaseModel

	ID         uint `gorm:"primaryKey" json:"-"`
	Level      uint `json:"level"`
	Technology uint `json:"-"`
	Cost       uint `json:"cost"`
}
