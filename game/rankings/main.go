package rankings

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"sync"

	"github.com/Vintral/pocket-realm/models"
	realmRedis "github.com/Vintral/pocket-realm/redis"
	"github.com/Vintral/pocket-realm/utils"
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

func getRank(base context.Context, round int, userId int) int {
	result := redisClient.ZRevRank(base, getRankingsKey(round), strconv.Itoa(userId)).Val()
	fmt.Println("UserID: " + fmt.Sprint(userId))
	fmt.Println("getRank: " + fmt.Sprint(result))
	return int(result)
}

func getRankings(base context.Context, round int, start int64, count int64, c chan []models.RankingSnapshot) {
	ctx, span := Tracer.Start(base, "rankings.getRankings")
	defer span.End()

	log.Warn().Msg("getRankings: " + fmt.Sprint(round) + " - " + fmt.Sprint(start) + " - " + fmt.Sprint(count))
	log.Warn().Str("rankings-key", getRankingsKey(round)).Send()

	if result, err := redisClient.ZRangeArgsWithScores(
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
	).Result(); err == nil {
		wg := new(sync.WaitGroup)
		wg.Add(len(result))
		ret := make([]models.RankingSnapshot, len(result))

		firstRank := -1
		for i, v := range result {
			if i == 0 {
				if firstId, err := strconv.Atoi(v.Member.(string)); err != nil {
					log.Error().AnErr("err", err).Any("user", v.Member).Msg("Error converting userId to int")
					c <- nil
					return
				} else {
					firstRank = getRank(ctx, round, firstId)
				}
			}

			go func() {
				defer wg.Done()

				log.Info().Int("i", i).Any("v", v).Msg("Retrieving User")
				userId, _ := strconv.Atoi(v.Member.(string))

				var data *struct {
					Username string `json:"username"`
					Avatar   string `json:"avatar"`
					Class    string `gorm:"column:character_class" json:"class"`
				}
				if err := db.WithContext(ctx).Table("users").Select("users.avatar, users.username, user_rounds.character_class").
					Joins("INNER JOIN user_rounds ON users.id = user_rounds.user_id").
					Where("user_rounds.user_id = ? AND user_rounds.round_id = ?", userId, round).
					Scan(&data).Error; err == nil {

					ret[i].Class = data.Class
					ret[i].Avatar = data.Avatar
					ret[i].Username = data.Username
					ret[i].Score = v.Score
				} else {
					log.Warn().AnErr("err", err).Str("user", v.Member.(string)).Msg("Error getting rank user's info")
				}

				ret[i].Rank = i + firstRank + 1
			}()
			log.Warn().Int("i", i).Msg("Fired off go routine")
		}
		wg.Wait()

		c <- ret
		return
	} else {
		log.Warn().AnErr("err", err).Msg("Error loading rank")
	}

	c <- nil
}

func getNearBounds(baseContext context.Context, user *models.User) int64 {
	ctx, span := Tracer.Start(baseContext, "rankings.getNearBounds")
	defer span.End()

	rank := getRank(ctx, user.RoundID, int(user.ID))
	spots := 10

	return int64(math.Max(float64(rank-spots), 0))
}

func RetrieveRankings(base context.Context) {
	ctx, span := Tracer.Start(base, "rankings.RetrieveRankings")
	defer span.End()

	log.Info().Msg("rankings.RetrieveRankings")

	user := base.Value(utils.KeyUser{}).(*models.User)

	c := make(chan []models.RankingSnapshot)
	d := make(chan []models.RankingSnapshot)

	start := getNearBounds(ctx, user)

	go getRankings(ctx, user.RoundID, 0, 15, c)
	go getRankings(ctx, user.RoundID, start, 15, d)

	top, near := <-c, <-d

	user.Connection.WriteJSON(RankingsResult{
		Type: "RANKINGS",
		Top:  top,
		Near: near,
	})
}
