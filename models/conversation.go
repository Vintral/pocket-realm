package models

import (
	"context"
	"fmt"
	"time"

	"github.com/Vintral/pocket-realm/utils"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gorm"
)

type ConversationUser struct {
	Username string    `json:"username"`
	Avatar   string    `json:"avatar"`
	GUID     uuid.UUID `json:"guid"`
}

type LastMessage struct {
	UserID uint   `json:"-"`
	Reply  bool   `json:"reply"`
	Text   string `json:"text"`
}

type Conversation struct {
	BaseModel

	GUID          uuid.UUID  `gorm:"uniqueIndex,size:36" json:"guid"`
	User1ID       uint       `json:"-"`
	User1LastRead time.Time  `gorm:"default:1970-01-01" json:"-"`
	User2ID       uint       `json:"-"`
	User2LastRead time.Time  `gorm:"default:1970-01-01" json:"-"`
	Username      string     `gorm:"->;-:migration" json:"username"`
	Avatar        string     `gorm:"->;-:migration" json:"avatar"`
	LastRead      time.Time  `gorm:"->;-:migration" json:"last_read"`
	UpdatedAt     time.Time  `json:"updated"`
	Replied       bool       `gorm:"->;-:migration" json:"replied"`
	Message       string     `gorm:"->;-:migration" json:"message"`
	Messages      []*Message `gorm:"-" json:"-"`
}

func (conversation *Conversation) BeforeCreate(tx *gorm.DB) (err error) {
	conversation.GUID = uuid.New()
	return
}

func (conversation *Conversation) GetMessages(ctx context.Context) error {
	ctx, span := utils.StartSpan(ctx, "conversation.GetMessages")
	defer span.End()

	if err := db.WithContext(ctx).
		Table("messages").
		Select("username", "avatar", "user_id", "text", "messages.created_at").
		Joins("INNER JOIN users ON users.id = messages.user_id").
		Where("conversation = ?", conversation.ID).
		Order("messages.id DESC").Limit(50).
		Find(&conversation.Messages).Error; err != nil {
		log.Error().Err(err).Msg("Error retrieving messages")
		return err
	}

	return nil
}

func CreateConversation(ctx context.Context, user1 uint, user2 uint) *Conversation {
	ctx, span := utils.StartSpan(ctx, "conversation.createConversation")
	defer span.End()

	conversation := Conversation{User1ID: user1, User2ID: user2}
	if err := db.WithContext(ctx).Save(&conversation).Error; err != nil {
		log.Error().Err(err).Uint("from", user1).Uint("to", user2).Msg("Error creating conversation")
		span.RecordError(err)
		return nil
	}

	return &conversation
}

func GetConversation(ctx context.Context, user1 uint, user2 uint) *Conversation {
	ctx, span := utils.StartSpan(ctx, "Conversation.GetConversation")
	defer span.End()

	log.Info().Uint("user1", user1).Uint("user2", user2).Msg("GetConversation")
	span.SetAttributes(attribute.Int("user1", int(user1)), attribute.Int("user2", int(user2)))

	var conversation *Conversation
	if err := db.WithContext(ctx).Table("conversations").Where("(user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?)", user1, user2, user2, user1).Scan(&conversation).Error; err != nil {
		log.Error().Err(err).Msg("Error finding conversation")
	}

	return conversation
}

func LoadConversation(guid uuid.UUID) *Conversation {
	fmt.Println("conversation:Load:", guid)

	var conversation *Conversation
	res := db.Table("conversations").Where("guid = ?", guid).Find(&conversation)
	if res.Error == nil {
		return conversation
	} else {
		return nil
	}
}

func LoadConversations(base context.Context, user *User, page int) []*Conversation {
	ctx, span := Tracer.Start(base, "conversation.LoadConversations")
	defer span.End()

	log.Trace().Msg("LoadConversations")

	perPage := 20
	var conversations []*Conversation
	if err := db.WithContext(ctx).Raw(`
		SELECT 
			conversations.id, avatar, username, users.guid, 
			CASE WHEN ? = user1_id THEN user1_last_read ELSE user2_last_read END AS last_read, 
			conversations.updated_at, msg.message, 
			CASE WHEN ? = msg.user_id THEN true ELSE false END as replied 
		FROM conversations
		INNER JOIN 
			( SELECT user_id, text AS message, conversation 
				FROM messages 
				INNER JOIN 
				( SELECT MAX(id) AS max_id 
					FROM messages 
					GROUP BY conversation ) AS m 
				ON m.max_id = messages.id ) AS msg 
			ON msg.conversation = conversations.id 
		INNER JOIN users 
			ON users.id = CASE WHEN user1_id = ? THEN user2_id ELSE user1_id END 
		WHERE (user1_id = ? AND user2_id <> 0 ) OR ( user1_id <> 0 AND user2_id = ?) 
		ORDER BY conversations.updated_at DESC LIMIT ?
	`, user.ID, user.ID, user.ID, user.ID, user.ID, perPage).Scan(&conversations).Error; err != nil {
		log.Error().Err(err).Uint("user", user.ID).Msg("Error getting conversations")
	}

	return conversations
}

func (conversation *Conversation) Save(ctx context.Context) error {
	result := db.WithContext(ctx).Save(&conversation)
	return result.Error
}

func (conversation *Conversation) Dump() {
	log.Trace().Msg(`
=============================")
ID: ` + fmt.Sprint(conversation.ID) + `
GUID: ` + fmt.Sprint(conversation.GUID) + `
Username: ` + conversation.Username + `
Avatar: ` + conversation.Avatar + `
Message: ` + conversation.Message + `
LastRead: ` + fmt.Sprint(conversation.LastRead) + `
UpdatedAt: ` + fmt.Sprint(conversation.UpdatedAt) + `
=============================
	`)
}
