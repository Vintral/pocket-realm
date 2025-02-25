package models

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type Contact struct {
	BaseModel

	ContactID   int       `json:"-"`
	Avatar      string    `gorm:"->;-:migration" json:"avatar"`
	ContactGuid uuid.UUID `gorm:"->;-:migration" json:"guid"`
	UserID      int       `json:"-"`
	Category    string    `json:"category"`
	Note        string    `json:"note"`
}

func (contact *Contact) Dump() {
	log.Warn().Msg(`
============================
ContactID: ` + fmt.Sprint(contact.ContactID) + `
ContactGuid: ` + contact.ContactGuid.String() + `
UserID: ` + fmt.Sprint(contact.UserID) + `
Avatar: ` + contact.Avatar + `
Category: ` + contact.Category + `
Note: ` + contact.Note + `
============================
	`)
}
