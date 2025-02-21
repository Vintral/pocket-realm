package models

type Ranking struct {
	BaseModel

	Rank    uint   `json:"rank"`
	RoundID uint   `json:"-"`
	UserID  uint   `json:"-"`
	Score   uint   `json:"score"`
	Avatar  string `gorm:"->;-:migration" json:"avatar"`
	Name    string `gorm:"->;-:migration" json:"name"`
}
