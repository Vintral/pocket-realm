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

type sendMessageResult struct {
	Type    string `json:"type"`
	Success bool   `json:"success"`
}

type getMessagesResult struct {
	Type     string            `json:"type"`
	Messages []*models.Message `json:"messages"`
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
			if conversation := models.GetConversation(ctx, user.ID, otherId); conversation != nil {
				span.AddEvent("Saving message")
				if err = db.WithContext(ctx).Save(&models.Message{Conversation: conversation.ID, UserID: user.ID, Text: payload.Message}).Error; err == nil {
					span.AddEvent("Updating conversation")
					if conversation.User1ID == user.ID {
						conversation.User1LastRead = time.Now()
					} else {
						conversation.User2LastRead = time.Now()
					}

					if err := conversation.Save(ctx); err == nil {
						success = true
					} else {
						span.RecordError(err)
					}
				}
			} else {
				errorMsg = "error getting conversation"
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

	var payload getMessagesPayload
	err := json.Unmarshal(base.Value(utils.KeyPayload{}).([]byte), &payload)
	if err != nil {
		fmt.Println(err)
		return
	}

	otherId := models.GetUserIdForGuid(ctx, payload.Conversation)
	span.SetAttributes(attribute.Int("other_user", int(otherId)))

	user := base.Value(utils.KeyUser{}).(*models.User)
	span.SetAttributes(attribute.Int("user", int(user.ID)))

	var conversation *int
	var messages []*models.Message

	if err := db.WithContext(ctx).Table("conversations").Select("id").Where("( user1_id = ? AND user2_id = ? ) OR ( user1_id = ? AND user2_id = ?)", user.ID, otherId, otherId, user.ID).Scan(&conversation).Error; err == nil && conversation != nil {
		span.SetAttributes(attribute.Int("conversation", *conversation))

		log.Info().Int("otherUser", int(otherId)).Int("user", int(user.ID)).Int("conversation", *conversation).Msg("Retrieve messages")

		if err := db.Table("messages").Where("conversation = ?", *conversation).Order("id DESC").Limit(50).Find(&messages).Error; err != nil {
			log.Error().Err(err).Msg("Error retrieving messages")
		}
	}

	user.Connection.WriteJSON(getMessagesResult{
		Type:     "MESSAGES",
		Messages: messages,
	})
}

func GetConversations(base context.Context) {
	ctx, span := utils.StartSpan(base, "get-conversations")
	defer span.End()

	user := base.Value(utils.KeyUser{}).(*models.User)
	fmt.Println("Get Conversations for:", user.ID)

	payload := struct {
		Type          string                 `json:"type"`
		Conversations []*models.Conversation `json:"conversations"`
	}{
		Type:          "CONVERSATIONS",
		Conversations: models.LoadConversations(ctx, user, 1),
	}
	user.Connection.WriteJSON(payload)
}
