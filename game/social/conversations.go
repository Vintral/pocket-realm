package social

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Vintral/pocket-realm/game/models"
	"github.com/Vintral/pocket-realm/game/utilities"
	"github.com/rs/zerolog/log"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	attributes "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type GetMessagesPayload struct {
	Type         string    `json:"type"`
	Conversation uuid.UUID `json:"conversation"`
}

type SendMessagePayload struct {
	Message string `json:"message"`
	To      string `json:"to"`
}

type ErrorSendingMessageResult struct {
	Type string `json:"type"`
}

type MessageSentResult struct {
	Type string `json:"type"`
}

// type ShoutResult struct {
// 	Type    string `json:"type"`
// 	Success bool   `json:"success"`
// }

// func (shout ShoutDataPayload) MarshalBinary() ([]byte, error) {
// 	return json.Marshal(shout)
// }

// func dispatchMessage(shout ShoutDataPayload) {
// 	if redisClient != nil {
// 		err := redisClient.Publish(context.Background(), "SHOUT", shout)
// 		fmt.Println(err)
// 	}
// }

func dispatchErrorSendingMessage(ctx context.Context, user *models.User) {
	user.SendError(models.SendErrorParams{Context: &ctx, Type: "message", Message: "error-sending-message"})
}

func SendMessage(base context.Context) {
	log.Info().Msg("sendMessage")

	ctx, span := utilities.StartSpan(base, "send-message")
	defer span.End()

	var payload SendMessagePayload
	err := json.Unmarshal(base.Value(utilities.KeyPayload{}).([]byte), &payload)
	if err != nil {
		fmt.Println(err)
		return
	}

	user := base.Value(utilities.KeyUser{}).(*models.User)

	span.SetAttributes(attribute.String("From", user.Username), attribute.String("To", payload.To))

	span.AddEvent("Getting other user id")
	if otherId := models.GetUserIdForName(ctx, payload.To); otherId != 0 {
		span.AddEvent("Getting conversation")
		if conversation := models.GetConversation(ctx, user.ID, otherId); conversation != nil {
			log.Trace().Msg("Other User: " + fmt.Sprint(otherId))
			log.Trace().Msg("Conversation Id: " + fmt.Sprint(conversation.ID))

			span.AddEvent("Saving message")
			message := &models.Message{Conversation: conversation.ID, UserID: user.ID, Text: payload.Message}
			if err := message.Save(ctx); err == nil {
				span.AddEvent("Updating conversation")
				if conversation.User1ID == user.ID {
					conversation.User1LastRead = time.Now()
				} else {
					conversation.User2LastRead = time.Now()
				}

				if err := conversation.Save(ctx); err != nil {
					span.SetStatus(codes.Error, "Error updating conversation")
					span.RecordError(err)
				}

				span.AddEvent("Sending success")
				user.Connection.WriteJSON(struct {
					Type string `json:"type"`
				}{
					Type: "MESSAGE_SENT",
				})

				return
			} else {
				span.RecordError(err)
			}
		} else {
			span.RecordError(errors.New("Error finding conversation"))
		}
	} else {
		span.RecordError(errors.New("Error finding other user"))
	}

	span.SetStatus(codes.Error, "Error sending message")
	span.AddEvent("Sending error")
	dispatchErrorSendingMessage(ctx, user)
}

func GetMessages(base context.Context) {
	_, span := utilities.StartSpan(base, "get-messages")
	defer span.End()

	var payload GetMessagesPayload
	err := json.Unmarshal(base.Value(utilities.KeyPayload{}).([]byte), &payload)
	if err != nil {
		fmt.Println(err)
		return
	}

	user := base.Value(utilities.KeyUser{}).(*models.User)
	fmt.Println("Get Messages for:", payload.Conversation)

	span.SetAttributes(attributes.String("conversation", payload.Conversation.String()))
	span.SetAttributes(attributes.Int("user", int(user.ID)))

	conversation := models.LoadConversation(payload.Conversation)
	if conversation == nil {
		fmt.Println("Conversation not found:", payload.Conversation)
		span.RecordError(errors.New("conversation not found"))
		span.SetStatus(codes.Error, "conversation-error")
		return
	}
	if conversation.User1ID != user.ID && conversation.User2ID != user.ID {
		fmt.Println("User not part of conversation:", conversation.ID, ":::", user.ID)
		span.RecordError(errors.New("user not part of conversation"))
		span.SetStatus(codes.Error, "conversation-error")
		return
	}

	fmt.Println(conversation)
	if err := conversation.LoadMessages(); err != nil {
		fmt.Println("Error loading messages:", payload.Conversation)
		span.RecordError(err)
		span.SetStatus(codes.Error, "conversation-error")
		return
	}

	fmt.Println(conversation)
	packet := struct {
		Type         string            `json:"type"`
		Conversation uuid.UUID         `json:"conversation"`
		Messages     []*models.Message `json:"messages"`
	}{
		Type:         "MESSAGES",
		Conversation: conversation.GUID,
		Messages:     conversation.Messages,
	}
	user.Connection.WriteJSON(packet)
}

func GetConversations(base context.Context) {
	ctx, span := utilities.StartSpan(base, "get-conversations")
	defer span.End()

	// var payload payloads.ExplorePayload
	// err := json.Unmarshal(base.Value(utilities.KeyPayload{}).([]byte), &payload)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	user := base.Value(utilities.KeyUser{}).(*models.User)
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
