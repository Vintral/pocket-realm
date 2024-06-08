package models

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/codes"
	"gorm.io/gorm"
)

type ConversationUser struct {
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
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

	db.WithContext(ctx).Table("users").Select("username", "avatar").Where("id = ?", conversation.User1ID).Scan(&conversation.User1)
	db.WithContext(ctx).Table("users").Select("username", "avatar").Where("id = ?", conversation.User2ID).Scan(&conversation.User2)

	return
}

func (conversation *Conversation) LoadMessages() (err error) {
	fmt.Println("conversation:AfterFind")

	res := db.Table("messages").Where("conversation = ?", conversation.ID).Order("id DESC").Limit(50).Find(&conversation.Messages)
	return res.Error
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
	fmt.Println("LoadConversations")

	ctx, span := Tracer.Start(base, "load-conversations")
	defer span.End()

	perPage := 20
	var conversations []*Conversation
	res := db.WithContext(ctx).Table("conversations").Where("user1_id = ? OR user2_id = ?", user.ID, user.ID).Limit(20).Offset((page - 1) * perPage).Find(&conversations)
	if res.Error != nil {
		span.SetStatus(codes.Error, "loading-conversation-error")
		span.RecordError(res.Error)
	} else {
		span.SetStatus(codes.Ok, "processed")

		for _, conversation := range conversations {
			db.WithContext(ctx).Table("messages").Where("conversation = ?", conversation.GUID).Order("ID desc").Limit(1).Scan(&conversation.LastMessage)
			conversation.LastMessage.Reply = user.ID == conversation.LastMessage.UserID

			if conversation.User2ID == user.ID {
				conversation.Username = conversation.User1.Username
				conversation.Avatar = conversation.User1.Avatar
				conversation.LastRead = conversation.User1LastRead
			} else {
				conversation.Username = conversation.User2.Username
				conversation.Avatar = conversation.User2.Avatar
				conversation.LastRead = conversation.User2LastRead
			}
		}
	}

	return conversations
}
