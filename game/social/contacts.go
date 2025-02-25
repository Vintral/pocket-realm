package social

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/Vintral/pocket-realm/models"
	"github.com/Vintral/pocket-realm/utils"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type getContactsPayload struct {
	Type string `json:"type"`
}

type getContactsResult struct {
	Type    string           `json:"type"`
	Friends []models.Contact `json:"friends"`
	Enemies []models.Contact `json:"enemies"`
}

type addContactPayload struct {
	Type        string    `json:"type"`
	Category    string    `json:"category"`
	ContactGuid uuid.UUID `json:"contact"`
	Note        string    `json:"note"`
}

type addContactResult struct {
	Type    string `json:"type"`
	Success bool   `json:"success"`
}

func loadContacts(baseContext context.Context, user *models.User, contactType string, wg *sync.WaitGroup, c chan []models.Contact) {
	ctx, span := utils.StartSpan(baseContext, "loadContacts")
	defer span.End()

	log.Info().Int("user", int(user.ID)).Str("type", contactType).Msg("loadContacts")

	var ret []models.Contact

	if err := db.WithContext(ctx).
		Table("contacts").
		Select("contact_id, user_id, type, note, users.guid AS contact_guid, users.avatar").
		Joins("INNER JOIN users ON users.id = contacts.contact_id").
		Where("user_id = ? AND type = ?", user.ID, contactType).
		Scan(&ret).Error; err != nil {
		log.Error().Err(err).Msg("Error retrieving contacts")
	} else {
		log.Info().Int("records", len(ret)).Msg("Retrieved contacts")

		for _, contact := range ret {
			contact.Dump()
		}
	}

	wg.Done()
	c <- ret
}

func AddContact(baseContext context.Context) {
	ctx, span := utils.StartSpan(baseContext, "AddContact")
	defer span.End()

	user := baseContext.Value(utils.KeyUser{}).(*models.User)
	success := false

	var payload addContactPayload
	if err := json.Unmarshal(baseContext.Value(utils.KeyPayload{}).([]byte), &payload); err == nil {
		var contactId *int
		if err := db.WithContext(ctx).Table("users").Select("id").Where("guid = ?", payload.ContactGuid).Scan(&contactId).Limit(1).Error; err == nil {
			var contact models.Contact
			if err = db.WithContext(ctx).Table("contacts").Where("user_id = ? AND contact_id = ?", user.ID, *contactId).First(&contact).Error; err != nil {
				if err = db.WithContext(ctx).Create(&models.Contact{
					ContactID: *contactId,
					UserID:    int(user.ID),
					Category:  payload.Category,
					Note:      payload.Note,
				}).Error; err == nil {
					success = true
					GetContacts(baseContext)
					log.Info().Int("user", int(user.ID)).Int("contact", *contactId).Str("type", payload.Type).Msg("Added Contact")
				}
			} else {
				log.Warn().Msg("Contact already exists")
			}
		} else {
			log.Error().Err(err).Str("guid", payload.ContactGuid.String()).Msg("Error getting contact id from GUID")
		}
	} else {
		log.Warn().Msg("Error getting payload: getContactsPayload")
	}

	user.Connection.WriteJSON(addContactResult{
		Type:    "ADD_CONTACT",
		Success: success,
	})
}

func GetContacts(baseContext context.Context) {
	_, span := utils.StartSpan(baseContext, "GetContacts")
	defer span.End()

	user := baseContext.Value(utils.KeyUser{}).(*models.User)

	var payload getContactsPayload
	if err := json.Unmarshal(baseContext.Value(utils.KeyPayload{}).([]byte), &payload); err == nil {
		f := make(chan []models.Contact)
		e := make(chan []models.Contact)

		wg := new(sync.WaitGroup)
		wg.Add(2)
		go loadContacts(baseContext, user, "friend", wg, f)
		go loadContacts(baseContext, user, "enemy", wg, e)
		wg.Wait()

		log.Info().Msg("*******************   Wait Groups Done")

		friends, enemies := <-f, <-e

		user.Connection.WriteJSON(getContactsResult{
			Type:    "GET_CONTACTS",
			Friends: friends,
			Enemies: enemies,
		})
	} else {
		log.Warn().Msg("Error getting payload: getContactsPayload")
	}
}
