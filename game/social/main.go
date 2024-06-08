package social

import (
	"context"
	"fmt"

	realmRedis "github.com/Vintral/pocket-realm/game/redis"
	"github.com/Vintral/pocket-realm/models"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/sdk/trace"
)

var redisClient *redis.Client
var tracerProvider *trace.TracerProvider

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

func Initialize(tp *trace.TracerProvider) {
	rdb, err := realmRedis.Instance(tp)
	if err != nil {
		panic(err)
	}

	redisClient = rdb

	tracerProvider = tp
	go handleShouts()

	subscribers = make(map[uint]*models.User)
}
