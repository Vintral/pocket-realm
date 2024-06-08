package models

type UserRound struct {
	BaseModel

	CharacterClass string  `json:"character_class"`
	Land           float64 `json:"land"`
	FreeLand       float64 `json:"free_land"`
	UserID         uint    `json:"-"`
	RoundID        uint    `json:"-"`
	Defense        float64 `json:"defense"`
	Energy         int     `json:"energy"`
	RecruitPower   float64 `json:"recruit_power"`
	BuildPower     float64 `json:"build_power"`
	Gold           float64 `json:"gold"`
	TickGold       float64 `json:"tick_gold"`
	Food           float64 `json:"food"`
	TickFood       float64 `json:"tick_food"`
	Wood           float64 `json:"wood"`
	TickWood       float64 `json:"tick_wood"`
	Metal          float64 `json:"metal"`
	TickMetal      float64 `json:"tick_metal"`
	Faith          float64 `json:"faith"`
	TickFaith      float64 `json:"tick_faith"`
	Stone          float64 `json:"stone"`
	TickStone      float64 `json:"tick_stone"`
	Mana           float64 `json:"mana"`
	TickMana       float64 `json:"tick_mana"`
}
