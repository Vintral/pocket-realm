package models

import (
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type SearchResult struct {
	GUID     uuid.UUID `json:"guid"`
	Username string    `json:"username"`
	Avatar   string    `json:"avatar"`
	Class    string    `json:"class"`
}

func (result *SearchResult) Dump() {
	log.Warn().Msg(`
============================
GUID: ` + result.GUID.String() + `
Username: ` + result.Username + `
Avatar: ` + result.Avatar + `
Class: ` + result.Class + `
============================
	`)
}
