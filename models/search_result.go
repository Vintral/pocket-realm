package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type SearchResult struct {
	ID       int       `json:"-"`
	GUID     uuid.UUID `json:"guid"`
	Rank     int       `json:"rank"`
	Score    int       `json:"score"`
	Username string    `json:"username"`
	Avatar   string    `json:"avatar"`
	Class    string    `json:"class"`
	LastSeen time.Time `json:"last_seen"`
}

func (result *SearchResult) Dump() {
	log.Warn().Msg(`
============================
GUID: ` + result.GUID.String() + `
Username: ` + result.Username + `
Avatar: ` + result.Avatar + `
Class: ` + result.Class + `
Rank: ` + fmt.Sprint(result.Rank) + `
Score: ` + fmt.Sprint(result.Score) + `
LastSeen: ` + result.LastSeen.String() + `
============================
	`)
}
