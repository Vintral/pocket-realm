package models

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NewsItem struct {
	BaseModel

	GUID      uuid.UUID `gorm:"uniqueIndex,size:36" json:"-"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Active    bool      `json:"-"`
	UpdatedAt time.Time `json:"posted"`
}

func (news *NewsItem) BeforeCreate(tx *gorm.DB) (err error) {
	news.GUID = uuid.New()
	return
}

func (news *NewsItem) Load() []*NewsItem {
	var newsItems []*NewsItem

	ctx, span := Tracer.Start(context.Background(), "news")
	defer span.End()

	db.WithContext(ctx).Where("active = ?", true).Find(&newsItems)
	return newsItems
}
