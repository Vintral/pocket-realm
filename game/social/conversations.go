package social

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Vintral/pocket-realm/models"
	"github.com/Vintral/pocket-realm/utils"
	"github.com/rs/zerolog/log"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
)

type getMessagesPayload struct {
	Type         string    `json:"type"`
	Conversation uuid.UUID `json:"conversation"`
}

type sendMessagePayload struct {
	Message string    `json:"message"`
	To      uuid.UUID `json:"to"`
}

type sendSupportMessagePayload struct {
	Message string `json:"message"`
}

type sendMessageResult struct {
	Type    string `json:"type"`
	Success bool   `json:"success"`
}

type getMessagesResult struct {
	Type     string            `json:"type"`
	Messages []*models.Message `json:"messages"`
}

type getConversationsResult struct {
	Type          string                 `json:"type"`
	Conversations []*models.Conversation `json:"conversations"`
}

func saveMessage(ctx context.Context, user *models.User, toUser uint, message string) error {
	ctx, span := utils.StartSpan(ctx, "social.saveMessage")
	defer span.End()

	errorMsg := ""

	if conversation := models.GetConversation(ctx, user.ID, toUser); conversation != nil {
		span.AddEvent("Saving message")
		if err := db.WithContext(ctx).Save(&models.Message{Conversation: conversation.ID, UserID: user.ID, Text: message}).Error; err == nil {
			span.AddEvent("Updating conversation")
			if conversation.User1ID == user.ID {
				conversation.User1LastRead = time.Now()
			} else {
				conversation.User2LastRead = time.Now()
			}

			if err := conversation.Save(ctx); err != nil {
				errorMsg = "error updating conversation"
			}
		}
	} else {
		errorMsg = "error getting conversation"
	}

	if errorMsg != "" {
		err := errors.New(errorMsg)
		log.Error().Err(err).Msg("Error saving message")
		span.RecordError(err)

		return err
	}

	return nil
}

func SendMessage(base context.Context) {
	ctx, span := utils.StartSpan(base, "SendMessage")
	defer span.End()

	user := base.Value(utils.KeyUser{}).(*models.User)
	success := false
	errorMsg := ""

	var payload sendMessagePayload
	if err := json.Unmarshal(base.Value(utils.KeyPayload{}).([]byte), &payload); err == nil {

		span.SetAttributes(attribute.String("From", user.Username), attribute.String("To", payload.To.String()))

		span.AddEvent("Getting other user id")
		if otherId := models.GetUserIdForGuid(ctx, payload.To); otherId != 0 {
			span.AddEvent("Getting conversation")
			if conversation := models.GetConversation(ctx, user.ID, otherId); conversation == nil {
				models.CreateConversation(ctx, user.ID, otherId)
			}

			if err := saveMessage(ctx, user, otherId, payload.Message); err != nil {
				errorMsg = "error sending message"
			} else {
				success = true
			}
		} else {
			errorMsg = "error finding other user"
		}
	} else {
		errorMsg = "error parsing SendMessagePayload"
	}

	if errorMsg != "" {
		err := errors.New(errorMsg)
		span.RecordError(err)
		log.Error().Err(err).Msg("Error: SendMEssage")
	}

	user.Connection.WriteJSON(sendMessageResult{
		Type:    "SEND_MESSAGE",
		Success: success,
	})
}

func GetMessages(base context.Context) {
	ctx, span := utils.StartSpan(base, "social.GetMessages")
	defer span.End()

	log.Info().Msg("WAT")

	span.AddEvent("Grabbing user")
	user := base.Value(utils.KeyUser{}).(*models.User)
	span.SetAttributes(attribute.Int("user", int(user.ID)))

	log.Info().Uint("user", user.ID).Msg("Have User")

	span.AddEvent("Parsing Payload")
	var conversation *models.Conversation
	var payload getMessagesPayload
	if err := json.Unmarshal(base.Value(utils.KeyPayload{}).([]byte), &payload); err != nil {
		log.Error().Err(err).Msg("Error parsing payload")
	} else {
		span.AddEvent("Getting other userId")
		otherId := models.GetUserIdForGuid(ctx, payload.Conversation)
		span.SetAttributes(attribute.Int("otherId", int(otherId)))
		log.Info().Int("OtherID", int(otherId)).Msg("Have OtherID")

		span.AddEvent("Getting conversation")
		if conversation = models.GetConversation(ctx, user.ID, otherId); conversation != nil {
			span.AddEvent("Getting messages")
			conversation.GetMessages(ctx)
		}
	}

	if conversation == nil {
		conversation = &models.Conversation{}
	}
	user.Connection.WriteJSON(getMessagesResult{
		Type:     "MESSAGES",
		Messages: conversation.Messages,
	})
}

func SendSupportMessage(baseContext context.Context) {
	ctx, span := utils.StartSpan(baseContext, "social.SendSupportMessage")
	defer span.End()

	user := baseContext.Value(utils.KeyUser{}).(*models.User)
	span.SetAttributes(attribute.Int("user", int(user.ID)))

	success := false
	var payload sendSupportMessagePayload
	if err := json.Unmarshal(baseContext.Value(utils.KeyPayload{}).([]byte), &payload); err == nil {
		span.SetAttributes(attribute.String("message", payload.Message))

		if err := saveMessage(ctx, user, 0, payload.Message); err == nil {
			success = true
		}
	}

	user.Connection.WriteJSON(sendMessageResult{
		Type:    "SEND_SUPPORT_MESSAGE",
		Success: success,
	})
}

func GetSupportMessages(baseContext context.Context) {
	ctx, span := utils.StartSpan(baseContext, "social.GetSupportMessages")
	defer span.End()

	user := baseContext.Value(utils.KeyUser{}).(*models.User)
	span.SetAttributes(attribute.Int("user", int(user.ID)))

	var conversation *models.Conversation
	if conversation = models.GetConversation(ctx, user.ID, 0); conversation != nil {
		conversation.GetMessages(ctx)
	}

	user.Connection.WriteJSON(getMessagesResult{
		Type:     "MESSAGES",
		Messages: conversation.Messages,
	})
}

func GetConversations(base context.Context) {
	ctx, span := utils.StartSpan(base, "get-conversations")
	defer span.End()

	user := base.Value(utils.KeyUser{}).(*models.User)
	fmt.Println("Get Conversations for:", user.ID)

	user.Connection.WriteJSON(getConversationsResult{
		Type:          "CONVERSATIONS",
		Conversations: models.LoadConversations(ctx, user, 1),
	})
}
