package models

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
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

	GUID          uuid.UUID        `gorm:"uniqueIndex,size:36" json:"guid"`
	User1ID       uint             `json:"-"`
	User1LastRead time.Time        `gorm:"default:1970-01-01" json:"-"`
	User1         ConversationUser `gorm:"-" json:"-"`
	User2ID       uint             `json:"-"`
	User2LastRead time.Time        `gorm:"default:1970-01-01" json:"-"`
	User2         ConversationUser `gorm:"-" json:"-"`
	Username      string           `gorm:"-" json:"username"`
	Avatar        string           `gorm:"-" json:"avatar"`
	LastRead      time.Time        `gorm:"-" json:"last_read"`
	UpdatedAt     time.Time        `json:"updated"`
	Messages      []*Message       `gorm:"-" json:"messages"`
	LastMessage   LastMessage      `gorm:"-" json:"last_message"`
}

func (conversation *Conversation) BeforeCreate(tx *gorm.DB) (err error) {
	conversation.GUID = uuid.New()
	return
}

func (conversation *Conversation) AfterFind(tx *gorm.DB) (err error) {
	fmt.Println("conversation:AfterFind")

	ctx, sp := Tracer.Start(tx.Statement.Context, "after-find")
	defer sp.End()

	db.WithContext(ctx).Table("users").Select("username", "avatar", "guid").Where("id = ?", conversation.User1ID).Scan(&conversation.User1)
	db.WithContext(ctx).Table("users").Select("username", "avatar", "guid").Where("id = ?", conversation.User2ID).Scan(&conversation.User2)

	return
}

func GetConversation(base context.Context, user1 uint, user2 uint) *Conversation {
	ctx, span := Tracer.Start(base, "Convsersation.GetConversation")
	defer span.End()

	log.Info().Uint("user1", user1).Uint("user2", user2).Msg("GetConversation")
	span.SetAttributes(attribute.Int("user1", int(user1)), attribute.Int("user2", int(user2)))

	var conversation *Conversation
	if err := db.WithContext(ctx).Table("conversations").Where("(user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?)", user1, user2, user2, user1).Scan(&conversation).Error; err != nil || conversation == nil {
		log.Info().Msg("Creating Conversation")

		conversation = &Conversation{User1ID: user1, User2ID: user2}
		if err = db.WithContext(ctx).Save(&conversation).Error; err != nil {
			log.Error().Err(err).Uint("from", user1).Uint("to", user2).Msg("Error creating conversation")
			span.RecordError(err)
			return nil
		}
	} else {
		log.Info().Msg("Found conversation?")
		fmt.Println(conversation)
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
	log.Trace().Msg("LoadConversations")

	ctx, span := Tracer.Start(base, "load-conversations")
	defer span.End()

	perPage := 20
	var conversations []*Conversation

	res := db.WithContext(ctx).
		Table("conversations").
		// Select("users.guid").
		Where("user1_id = ? OR user2_id = ?", user.ID, user.ID).
		Joins("INNER JOIN ( SELECT messages.conversation FROM messages GROUP BY conversation DESC) AS msg ON msg.conversation = conversations.id ").
		Joins("INNER JOIN users ON users.id = CASE WHEN user1_id = ? THEN user1_id ELSE user2_id END", user.ID).
		Limit(20).Offset((page - 1) * perPage).
		Find(&conversations)
	if res.Error != nil {
		span.SetStatus(codes.Error, "loading-conversation-error")
		span.RecordError(res.Error)
	} else {
		span.SetStatus(codes.Ok, "processed")

		for _, conversation := range conversations {
			db.WithContext(ctx).Table("messages").Where("conversation = ?", conversation.ID).Order("ID desc").Limit(1).Scan(&conversation.LastMessage)

			conversation.LastMessage.Reply = user.ID == conversation.LastMessage.UserID

			if conversation.User2ID == user.ID {
				conversation.Username = conversation.User1.Username
				conversation.Avatar = conversation.User1.Avatar
				conversation.LastRead = conversation.User1LastRead
				conversation.GUID = conversation.User1.GUID
			} else {
				conversation.Username = conversation.User2.Username
				conversation.Avatar = conversation.User2.Avatar
				conversation.LastRead = conversation.User2LastRead
				conversation.GUID = conversation.User2.GUID
			}

			conversation.Dump()
		}
	}

	return conversations
}

func (conversation *Conversation) Save(ctx context.Context) error {
	result := db.WithContext(ctx).Save(&conversation)
	return result.Error
}

func (conversation *Conversation) Dump() {
	log.Warn().Msg(`
=============================")
ID: ` + fmt.Sprint(conversation.ID) + `
GUID: ` + fmt.Sprint(conversation.GUID) + `
User1.ID: ` + fmt.Sprint(conversation.User1ID) + `
User1.Username: ` + conversation.User1.Username + `
User1.Avatar: ` + conversation.User1.Avatar + `
User2.ID: ` + fmt.Sprint(conversation.User2ID) + `
User2.Username: ` + conversation.User2.Username + `
User2.Avatar: ` + conversation.User2.Avatar + `
Username: ` + conversation.Username + `
Avatar: ` + conversation.Avatar + `
LastRead: ` + fmt.Sprint(conversation.LastRead) + `
UpdatedAt: ` + fmt.Sprint(conversation.UpdatedAt) + `
LastMessage: ` + conversation.LastMessage.Text + `
=============================
	`)
}
