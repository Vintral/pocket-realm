package rankings

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"sync"

	"github.com/Vintral/pocket-realm/models"
	realmRedis "github.com/Vintral/pocket-realm/redis"
	"github.com/Vintral/pocket-realm/utilities"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	redisDef "github.com/redis/go-redis/v9"
	provider "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

var redisClient *redisDef.Client
var Tracer trace.Tracer
var db *gorm.DB

type RankingsResult struct {
	Type string                   `json:"type"`
	Top  []models.RankingSnapshot `json:"top"`
	Near []models.RankingSnapshot `json:"near"`
}

func Initialize(tp *provider.TracerProvider, dbase *gorm.DB) {
	Tracer = tp.Tracer("rankings-tracer")

	rdb, err := realmRedis.Instance(tp)
	if err != nil {
		panic(err)
	}
	if rdb == nil {
		log.Panic().Msg("Redis instance is nil")
	}

	redisClient = rdb

	db = dbase
}

func getRankingsKey(round int) string {
	return fmt.Sprint(round) + "-rankings"
}

func getNearStart(base context.Context, user *models.User, desired int64) int {
	rank := getRank(base, user.RoundID, int(user.ID))
	fmt.Println("Rank: " + fmt.Sprint(rank))
	fmt.Println(float64(int64(rank) - desired/2))
	return int(math.Max(float64(int64(rank)-desired/2), 0))
}

func getRank(base context.Context, round int, userId int) int {
	result := redisClient.ZRevRank(base, getRankingsKey(round), strconv.Itoa(userId)).Val()
	fmt.Println("UserID: " + fmt.Sprint(userId))
	fmt.Println("getRank: " + fmt.Sprint(result))
	return int(result)
}

func getRankings(base context.Context, round int, start int64, count int64, c chan []models.RankingSnapshot) {
	ctx, span := Tracer.Start(base, "get-rankings")
	defer span.End()

	log.Warn().Msg("getRankings: " + fmt.Sprint(round) + " - " + fmt.Sprint(start) + " - " + fmt.Sprint(count))

	result, err := redisClient.ZRangeArgs(
		ctx,
		redisDef.ZRangeArgs{
			Key:     getRankingsKey(round),
			Stop:    1000000000,
			Start:   0,
			ByScore: true,
			Rev:     true,
			Offset:  start,
			Count:   count,
		},
	).Result()
	if err != nil {
		log.Warn().AnErr("error", err).Msg("Error retrieving rankings")
	}

	if len(result) < 1 {
		c <- nil
	}

	firstRank, err := strconv.Atoi(result[0])
	if err != nil {
		c <- nil
		return
	}
	firstRank = getRank(ctx, round, firstRank)

	wg := new(sync.WaitGroup)
	wg.Add(len(result))
	ret := make([]models.RankingSnapshot, len(result))
	for i, v := range result {
		go func() {
			defer wg.Done()

			res := redisClient.Get(
				ctx,
				fmt.Sprint(round)+"-snapshot-"+v,
			)

			var data models.RankingSnapshot
			err = json.Unmarshal([]byte(res.Val()), &data)
			if err != nil {
				log.Warn().AnErr("error", err).Msg("Error decoding snapshot")
				ret[i] = models.RankingSnapshot{}
			} else {
				ret[i] = data
			}

			if res.Val() == redisDef.Nil.Error() {
				log.Warn().Msg("Key not found")
			}

			var u *models.User
			db.WithContext(ctx).Table("users").Select("avatar").Where("id = ?", v).Scan(&u)

			avatar, _ := strconv.Atoi(u.Avatar)
			ret[i].Avatar = avatar
			ret[i].Rank = i + firstRank + 1
		}()
	}
	wg.Wait()

	c <- ret
}

func RetrieveRankings(base context.Context) {
	ctx, span := Tracer.Start(base, "retrieve-rankings")
	defer span.End()

	log.Info().Msg("RetrieveRankings")

	user := base.Value(utilities.KeyUser{}).(*models.User)

	c := make(chan []models.RankingSnapshot)
	d := make(chan []models.RankingSnapshot)
	nearResults := int64(3)

	go getRankings(ctx, user.RoundID, 0, 20, c)
	go getRankings(ctx, user.RoundID, int64(getNearStart(ctx, user, nearResults)), int64(nearResults), d)

	top, near := <-c, <-d

	user.Connection.WriteJSON(RankingsResult{
		Type: "RANKINGS",
		Top:  top,
		Near: near,
	})
}
