package models

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/Vintral/pocket-realm/game/utilities"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gorm"
)

type Event struct {
	BaseModel

	GUID      uuid.UUID `gorm:"uniqueIndex,size:36" json:"guid"`
	UserID    uint      `json:"-"`
	Round     uuid.UUID `json:"round"`
	Event     string    `json:"event"`
	Seen      bool      `json:"seen"`
	UpdatedAt time.Time `json:"time"`
}

func (event *Event) BeforeCreate(tx *gorm.DB) (err error) {
	event.GUID = uuid.New()
	return
}

func markEventSeen(event *Event) {
	ctx, span := Tracer.Start(context.Background(), "mark-event-seen")
	defer span.End()

	fmt.Println(event)
	event.Seen = true
	res := db.WithContext(ctx).Save(&event)
	if res.Error != nil {
		log.Warn().Int("event-id", int(event.ID)).Msg("Error Marking Event Seen: " + fmt.Sprint(event.ID))
	}
}

func markEventsSeen(events []*Event) {
	time.Sleep(3 * time.Second)

	_, span := Tracer.Start(context.Background(), "mark-events-seen")
	defer span.End()

	for _, e := range events {
		go markEventSeen(e)
	}
}

func MaxEventPages(baseContext context.Context, userid int, round uuid.UUID) int {
	ctx, span := Tracer.Start(baseContext, "max-event-pages")
	defer span.End()

	perPage := 20

	var count int64
	db.WithContext(ctx).Table("events").
		Select("COUNT(id) AS total").
		Where("user_id = ? AND ( round = ? OR round = ?)", userid, round, uuid.Nil).
		Count(&count)

	log.Debug().Msg("Retrieved Event Count: " + fmt.Sprint(count))

	return int(math.Ceil(float64(count) / float64(perPage)))
}

func LoadEvents(baseContext context.Context, userid int, round uuid.UUID, page int) []*Event {
	ctx, span := Tracer.Start(baseContext, "load-events")
	defer span.End()

	var events []*Event
	perPage := 20

	fmt.Println("ROUND:", round)
	fmt.Println("USER:", userid)

	res := db.WithContext(ctx).Table("events").
		Order("id desc").
		Where("user_id = ? AND ( round = ? OR round = ?)", userid, round, uuid.Nil).
		Limit(perPage).Offset((page - 1) * perPage).Find(&events)
	if res.Error != nil {
		log.Error().AnErr("Error Loading Events", res.Error).Msg("Error Loading Events")
	}

	return events
}

func MarkEventSeen(baseContext context.Context, evt uuid.UUID) {
	user := baseContext.Value(utilities.KeyUser{}).(*User)

	ctx, span := Tracer.Start(baseContext, "mark-event-seen")
	defer span.End()

	log.Info().Msg("Mark Event Seen: " + evt.String())
	span.SetAttributes(attribute.String("evt", evt.String()))
	span.SetAttributes(attribute.Int("user", int(user.ID)))

	var event *Event
	db.WithContext(ctx).Where("guid = ? AND user_id = ?", evt, user.ID).First(&event)

	fmt.Println(event)
	event.Seen = true
	db.WithContext(ctx).Save(&event)
}
