package social

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Vintral/pocket-realm-test-access/game/utilities"
	"github.com/Vintral/pocket-realm-test-access/models"

	"github.com/google/uuid"
	attributes "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type GetMessagesPayload struct {
	Type         string    `json:"type"`
	Conversation uuid.UUID `json:"conversation"`
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
