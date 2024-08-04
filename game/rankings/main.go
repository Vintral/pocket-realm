package rankings

import (
	"context"
	"fmt"
	"math"

	models "github.com/Vintral/pocket-realm/models"
	realmRedis "github.com/Vintral/pocket-realm/redis"
	"github.com/rs/zerolog/log"

	redisDef "github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/sdk/trace"
)

var redisClient *redisDef.Client
var tracerProvider *trace.TracerProvider

func UpdateRank(base context.Context, user models.User) {
	ctx, span := tracerProvider.Tracer("update-rank-tracer").Start(base, "update-rank")
	defer span.End()

	log.Trace().Msg("Update Rank")

	score := math.Floor(user.RoundData.Land * 10)
	log.Warn().Msg("UpdateRank: " + fmt.Sprint(user.ID) + " -- " + fmt.Sprint(score))

	result := redisClient.ZAdd(
		ctx,
		fmt.Sprint(user.RoundID)+"-rankings",
		redisDef.Z{Score: user.RoundData.Land * 10, Member: user.ID},
	)
	if result.Err() != nil {
		log.Warn().AnErr("err", result.Err()).Msg("Error updating redis rank")
		return
	}

	if err := redisClient.Set(
		ctx,
		fmt.Sprint(user.RoundID)+"-snapshot-"+fmt.Sprint(user.ID),
		&models.RankingSnapshot{Username: user.Username, Power: math.Floor(score), Land: math.Floor(user.RoundData.Land)},
		0,
	).Err(); err != nil {
		log.Warn().AnErr("err", err).Msg("Error updating redis snapshot")
	}
}

func Initialize(tp *trace.TracerProvider) {
	tracerProvider = tp

	rdb, err := realmRedis.Instance(tp)
	if err != nil {
		panic(err)
	}

	redisClient = rdb
}
