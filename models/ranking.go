package models

type Ranking struct {
	BaseModel

	Place   uint   `json:"place"`
	RoundID uint   `json:"-"`
	UserID  uint   `json:"-"`
	Power   uint   `json:"power"`
	Land    uint   `json:"land"`
	Avatar  string `gorm:"->;-:migration" json:"avatar"`
	Name    string `gorm:"->;-:migration" json:"name"`
}
