package realmRedis

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/sdk/trace"
)

var rdb *redis.Client

func Instance(tp *trace.TracerProvider) (*redis.Client, error) {
	log.Trace().Msg("realmRedis: Instance")

	// Use cached value if we can
	if rdb != nil {
		return rdb, nil
	}

	if tp != nil {
		_, sp := tp.Tracer("realm-redis").Start(context.Background(), "setup-redis")
		defer sp.End()
	}

	REDIS_HOST := os.Getenv("REDIS_HOST")
	REDIS_PORT := os.Getenv("REDIS_PORT")

	log.Warn().Msg("Redis @ " + REDIS_HOST + ":" + fmt.Sprint(REDIS_PORT))
	rdb = redis.NewClient(&redis.Options{
		Addr:     REDIS_HOST + ":" + REDIS_PORT,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	if rdb == nil {
		log.Warn().Msg("Redis client is nil")
	}

	return rdb, nil
}

// func UpdateScore(ctx context.Context, user *models.User) {
// 	score := user.RoundData.Land * 10
// 	log.Trace().Msg("UpdateScore: " + fmt.Sprint(user.ID) + " -- " + fmt.Sprint(score))

// 	result := rdb.ZAdd(
// 		ctx,
// 		fmt.Sprint(user.RoundID)+"-rankings",
// 		redis.Z{Score: user.RoundData.Land * 10, Member: user.ID},
// 	)

// 	fmt.Println(result.Val)
// 	fmt.Println(result.Result())
// 	//log.Warn().Msg("Result: " + result.Val())
// }
