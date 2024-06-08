package redis

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/sdk/trace"
)

var rdb *redis.Client

func Instance(tp *trace.TracerProvider) (*redis.Client, error) {
	// Use cached value if we can
	if rdb != nil {
		return rdb, nil
	}

	_, sp := tp.Tracer("realm-redis").Start(context.Background(), "setup-redis")
	defer sp.End()

	REDIS_HOST := os.Getenv("REDIS_HOST")
	REDIS_PORT := os.Getenv("REDIS_PORT")

	fmt.Println("Redis @ ", REDIS_HOST+":"+REDIS_PORT)
	rdb = redis.NewClient(&redis.Options{
		Addr:     REDIS_HOST + ":" + REDIS_PORT,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	return rdb, nil
}
