package models

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type Contact struct {
	BaseModel

	ContactID      int       `json:"-"`
	Avatar         string    `gorm:"->;-:migration" json:"avatar"`
	ContactGuid    uuid.UUID `gorm:"->;-:migration" json:"guid"`
	UserID         int       `json:"-"`
	Category       string    `json:"-"`
	Note           string    `json:"note"`
	CharacterClass string    `gorm:"->;-:migration;column:character_class" json:"characterClass"`
	Username       string    `gorm:"->;-:migration" json:"username"`
}

func (contact *Contact) Dump() {
	log.Warn().Msg(`
============================
ContactID: ` + fmt.Sprint(contact.ContactID) + `
Username: ` + contact.Username + `
ContactGuid: ` + contact.ContactGuid.String() + `
UserID: ` + fmt.Sprint(contact.UserID) + `
CharacterClass: ` + contact.CharacterClass + `
Avatar: ` + contact.Avatar + `
Category: ` + contact.Category + `
Note: ` + contact.Note + `
============================
	`)
}
