package models

type TechnologyLevel struct {
	BaseModel

	ID         uint `gorm:"primaryKey" json:"order"`
	Level      uint `json:"level"`
	Technology uint `json:"technology"`
	Cost       uint `json:"cost"`
}
