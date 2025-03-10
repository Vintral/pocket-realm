package models

type Pantheon struct {
	BaseModel

	Category  string      `json:"category"`
	Devotions []*Devotion `gorm:"-" json:"devotions"`
}
