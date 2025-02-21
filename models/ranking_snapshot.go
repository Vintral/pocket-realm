package models

import (
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"
)

type RankingSnapshot struct {
	Rank     int     `gorm:"-" json:"rank"`
	Username string  `gorm:"-" json:"username"`
	Avatar   int     `gorm:"-" json:"avatar"`
	Score    float64 `gorm:"-" json:"score"`
	Class    string  `gorm:"column:character_class" json:"class"`
}

func (ranking *RankingSnapshot) MarshalBinary() ([]byte, error) {
	return json.Marshal(ranking)
}

func (ranking *RankingSnapshot) UnMarshalBinary(data []byte, resp interface{}) error {
	return json.Unmarshal(data, resp)
}

func (ranking *RankingSnapshot) Dump() {
	log.Warn().Msg(`
============================
Rank: ` + fmt.Sprint(ranking.Rank) + `
Username: ` + ranking.Username + `
Avatar: ` + fmt.Sprint(ranking.Avatar) + `
Score: ` + fmt.Sprint(ranking.Score) + `
Class: ` + ranking.Class + `
============================
	`)
}
