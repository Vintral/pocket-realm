package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Vintral/pocket-realm/models"
	"github.com/Vintral/pocket-realm/utils"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/sdk/trace"
	tracerDefinition "go.opentelemetry.io/otel/trace"

	realmRedis "github.com/Vintral/pocket-realm/redis"
	redisDef "github.com/redis/go-redis/v9"

	"github.com/go-co-op/gocron/v2"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

var db *gorm.DB
var tracer tracerDefinition.Tracer
var tp *trace.TracerProvider
var redisClient *redisDef.Client

func buildTickField(field string) string {
	var builder strings.Builder
	builder.WriteString("tick_")
	builder.WriteString(field)

	return builder.String()
}

func getBankruptQuery(field string) string {
	tickField := buildTickField(field)

	var sb strings.Builder
	sb.WriteString("round_id = ? AND ")
	sb.WriteString(tickField)
	sb.WriteString(" < 0 AND -")
	sb.WriteString(tickField)
	sb.WriteString(" > ")
	sb.WriteString(field)

	return sb.String()
}

func dealWithBankruptUser(ctx context.Context, field string, roundid uint, userid int) {
	ctx, span := utils.StartCronSpan(ctx, "dealWithBankruptUser")
	defer span.End()

	user := models.LoadUserForRound(userid, int(roundid))

	if ok := user.ProcessBankruptcy(ctx, field); !ok {
		log.Warn().Uint("roundid", roundid).Uint("userid", user.ID).Str("field", field).Msg("Error processing Bankruptcy")
	}
}

func handleBankruptcies(baseContext context.Context, roundid uint, field string) {
	var sb strings.Builder
	sb.WriteString("handle-bankruptcies-")
	sb.WriteString(field)

	ctx, span := utils.StartCronSpan(baseContext, sb.String())
	defer span.End()

	log.Warn().Msg("=====================================")
	log.Warn().Msg(getBankruptQuery(field))
	log.Warn().Msg("=====================================")

	var userIDs []int
	db.WithContext(ctx).Model(&models.UserRound{}).Where(getBankruptQuery(field), roundid).Select("user_id").Scan(&userIDs)

	log.Warn().
		Int("bankrupt_users", len(userIDs)).
		Str("type", field).
		Msg("Users going bankrupt: " + fmt.Sprint(len(userIDs)) + " ::: " + field)

	wg := new(sync.WaitGroup)
	wg.Add(len(userIDs))
	for _, u := range userIDs {
		go func(userid int) {
			dealWithBankruptUser(ctx, field, roundid, userid)
			wg.Done()
		}(u)
	}
	wg.Wait()
}

func processField(baseContext context.Context, roundid uint, field string, wg *sync.WaitGroup) {
	log.Warn().Msg("processField: " + field)

	// ctx, span := tracer.Start(baseContext, "process-field-"+field)
	ctx, span := utils.StartCronSpan(baseContext, "process-field-"+field)
	defer span.End()
	if wg != nil {
		defer wg.Done()
	}

	tickField := buildTickField(field)

	handleBankruptcies(ctx, roundid, field)

	var query strings.Builder
	query.WriteString("round_id = ? AND (")
	query.WriteString("(" + tickField + " >= 0)")
	query.WriteString(" OR ")
	query.WriteString("(" + field + " >= " + tickField + ")")
	query.WriteString(")")


	res := db.WithContext(ctx).Model(&models.UserRound{}).Where(query.String(), roundid).Update(field, gorm.Expr(field+" + "+tickField))
	log.Trace().Msg("Rows Affected: " + fmt.Sprint(res.RowsAffected))
}

func growPopulations(baseContext context.Context, roundid uint) {
	ctx, span := tracer.Start(baseContext, "grow-populations")
	defer span.End()

	log.Debug().Msg("Grow Populations")

	var userIDs []int
	db.WithContext(ctx).Table("user_rounds").Select("user_id").Where("population < housing AND round_id = ?", roundid).Scan(&userIDs)
	db.WithContext(ctx).Exec("UPDATE user_rounds SET population = population + 1, tick_gold = tick_gold + 1 WHERE population < housing AND round_id = ?", roundid)

	log.Warn().Msg("User Count with population updates(" + fmt.Sprint(roundid) + "):" + fmt.Sprint(len(userIDs)))
}

func processRound(baseContext context.Context, roundid uint, waitgroup *sync.WaitGroup) {
	label := "process-round-" + fmt.Sprint(roundid)

	// ctx, span := tracer.Start(baseContext, label)
	ctx, span := utils.StartCronSpan(baseContext, label)
	defer span.End()
	defer waitgroup.Done()
	defer log.Warn().Msg("Done processRound")

	log.Info().Uint("roundid", roundid).Msg("processRound: " + fmt.Sprint(roundid))

	fields := [...]string{"gold", "food", "wood", "metal", "stone", "mana", "faith"}

	growPopulations(ctx, roundid)

	//wg := new(sync.WaitGroup)
	//wg.Add(len(fields))
	for _, f := range fields {
		processField(ctx, roundid, f, nil)
	}
	//wg.Wait()

	var userIDs []int
	log.Info().Msg("Current date: " + fmt.Sprint(time.Now()))
	db.WithContext(ctx).Model(&models.UserBuff{}).Distinct("user_id").Where("round_id = ? AND expires <= ?", roundid, time.Now()).Select("user_id").Scan(&userIDs)
	log.Info().Int("users", len(userIDs)).Msg("Remove expired buffs")
	for _, userID := range userIDs {
		log.Info().Int("user", userID).Msg("Remove expired buffs for user")

		user := models.LoadUserForRound(userID, int(roundid))
		user.RemoveExpiredBuffs(ctx)
		db.WithContext(ctx).Save(&user)
	}

	if round, err := models.LoadRoundById(ctx, int(roundid)); err == nil {
		if err = db.WithContext(ctx).Exec("UPDATE user_rounds SET energy =  energy + ? WHERE energy < ? AND round_id = ?", round.EnergyRegen, round.EnergyMax, round.ID).Error; err != nil {
			log.Error().Int("round", int(round.ID)).Err(err).Msg("Error updating user energy for round")
		} else {
			if err = db.WithContext(ctx).Table("user_rounds").Where("energy > ?", round.EnergyMax).Update("energy", round.EnergyMax).Error; err != nil {
				log.Error().Int("round", int(round.ID)).Err(err).Msg("Error correcting over the cap energy")
			} else {
				log.Info().Msg("ROUND ENERGY UPDATED")
			}
		}
	} else {
		log.Error().Err(err).Msg("Error retrieving round")
	}

	db.WithContext(ctx).Unscoped().Where("expires <= ? AND expires <> 0", time.Now()).Delete(&models.UserBuff{})
	db.WithContext(ctx).Unscoped().Where("quantity = ?", 0).Delete(&models.UserUnit{})
	db.WithContext(ctx).Unscoped().Where("quantity = ?", 0).Delete(&models.UserBuilding{})

	redisClient.Publish(ctx, "ROUND_UPDATE", roundid)
}

func process() {
	defer log.Warn().Msg("Finished Process")
	minute := time.Now().Minute()

	ctx, span := tracer.Start(context.Background(), "process")
	defer span.End()

	ctx = context.WithValue(ctx, utils.KeyTraceProvider{}, tp)

	if time.Now().Hour() == 0 && time.Now().Minute() == 0 {
		fmt.Println("Reseting active rounds")
		models.ResetActiveRounds(ctx)
	}

	log.Info().Int("minute", minute).Msg("Cron: " + fmt.Sprint(minute))

	rounds := models.GetActiveRoundsForTick(ctx, minute)
	if rounds == nil {
		log.Warn().Msg("No Active Rounds")
		return
	}

	wg := new(sync.WaitGroup)
	wg.Add(len(rounds))
	for _, r := range rounds {
		go processRound(ctx, r.ID, wg)
	}
	log.Warn().Msg("Waiting...")
	wg.Wait()
	log.Warn().Msg("Done Waiting!")
}

func setupDbase() {
	log.Info().Msg("Loading Environment")
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	log.Info().Msg("Setting up database")
	db, err = models.Database(false, nil)
	if err != nil {
		time.Sleep(3 * time.Second)
		panic(err)
	}
}

func setupLogs() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	output := zerolog.ConsoleWriter{
		Out:           os.Stderr,
		FieldsExclude: []string{zerolog.TimestampFieldName},
	}

	log.Logger = log.Output(output).With().Logger()
}

func setupRedis(tp *trace.TracerProvider) {
	log.Info().Msg("Setup Redis")

	rdb, err := realmRedis.Instance(tp)
	if err != nil {
		panic(err)
	}

	redisClient = rdb
}

func main() {
	setupLogs()
	log.Info().Msg("Running Cron")

	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println(err)
	}

	//==============================//
	//	Setup Telemetry							//
	//==============================//
	log.Info().Msg("Setting up telemetry")
	otelShutdown, provider, err := setupOTelSDK(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}

	defer func() {
		fmt.Println("In shutdown")
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	log.Info().Msg("Setting Trace Provider")
	tp = provider
	if tp == nil {
		panic("Trace Provider is nil")
	}
	tracer = tp.Tracer("realm-cron")
	models.SetTracerProvider(tp)

	setupRedis(tp)
	setupDbase()

	scheduler, err := gocron.NewScheduler()
	defer func() { _ = scheduler.Shutdown() }()

	if err != nil {
		panic("Error creating crons")
	}

	if _, err = scheduler.NewJob(
		gocron.CronJob(
			"* * * * *",
			false,
		),
		gocron.NewTask(
			process,
		),
	); err != nil {
		log.Error().AnErr("err", err).Msg("Error setting up crons")
	} else {
		log.Info().Msg("Starting up...")
		scheduler.Start()
	}

	process()

	// for {
	// 	time.Sleep(60 * time.Second)
	// 	log.Trace().Msg("Tick")
	// }
}
