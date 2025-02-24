package social

import (
	"context"
	"fmt"

	"github.com/Vintral/pocket-realm/models"
	realmRedis "github.com/Vintral/pocket-realm/redis"
	"gorm.io/gorm"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/sdk/trace"
)

var redisClient *redis.Client
var tracerProvider *trace.TracerProvider
var db *gorm.DB

func Initialize(tp *trace.TracerProvider, dbase *gorm.DB) {
	rdb, err := realmRedis.Instance(tp)
	if err != nil {
		panic(err)
	}

	redisClient = rdb
	db = dbase

	tracerProvider = tp
	go handleShouts()

	subscribers = make(map[uint]*models.User)
}

func handleShouts() {
	pubsub := redisClient.Subscribe(context.Background(), "SHOUT")
	defer pubsub.Close()

	tracer := tracerProvider.Tracer("social-tracer")
	for {
		ctx, span := tracer.Start(context.Background(), "process-shout-message")

		msg, err := pubsub.ReceiveMessage(ctx)
		if err != nil {
			panic(err)
		}

		fmt.Println("RECEIVED MESSAGE :::", msg.Channel, msg.Payload)

		for key, value := range subscribers {
			fmt.Println("Sending shout to", key)
			value.Connection.WriteJSON(msg.Payload)
		}

		span.End()
	}
}
