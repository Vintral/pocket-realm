package models

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Shout struct {
	BaseModel

	GUID           uuid.UUID `gorm:"uniqueIndex,size:36" json:"guid"`
	UserID         uint      `json:"-"`
	User           string    `gorm:"-" json:"username"`
	Avatar         string    `gorm:"-" json:"avatar"`
	CharacterClass string    `gorm:"-" json:"character_class"`
	Shout          string    `json:"shout"`
	CreatedAt      time.Time `json:"time"`
}

func (shout *Shout) BeforeCreate(tx *gorm.DB) (err error) {
	shout.GUID = uuid.New()
	return
}

func (shout *Shout) AfterFind(tx *gorm.DB) (err error) {
	fmt.Println("shout:AfterFind")

	ctx, sp := Tracer.Start(tx.Statement.Context, "after-find")
	defer sp.End()

	var u *User
	db.WithContext(ctx).Table("users").Select("username").Where("id = ?", shout.UserID).Scan(&u)

	shout.User = u.Username
	shout.Avatar = ""

	return
}

func (shout *Shout) Create(userID uint, text string) error {
	result := db.Create(&Shout{UserID: userID, Shout: text})
	return result.Error
}

type ShoutData struct {
	GUID           uuid.UUID `json:"guid"`
	Username       string    `json:"username"`
	Avatar         string    `json:"avatar"`
	CharacterClass string    `json:"character_class"`
	Shout          string    `json:"shout"`
	CreatedAt      time.Time `json:"created_at"`
}

func (shout *Shout) Load(page int) []*ShoutData {
	var shouts []*ShoutData

	ctx, span := Tracer.Start(context.Background(), "load-shouts")
	defer span.End()

	if page < 1 {
		page = 1
	}
	perPage := 20

	res := db.WithContext(ctx).Table("shouts").
		Select("shouts.guid", "shouts.created_at", "shouts.shout", "users.username", "users.avatar", "user_rounds.character_class").
		Joins("LEFT JOIN users ON users.id = shouts.user_id").
		Joins("LEFT JOIN user_rounds ON ( user_rounds.user_id = users.id AND user_rounds.round_id = users.round_id )").
		Order("shouts.id desc").
		Limit(perPage).Offset((page - 1) * perPage).Scan(&shouts)
	if res.Error != nil {
		fmt.Println("ERROR:", res.Error.Error())
	}

	return shouts
}
