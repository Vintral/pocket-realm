package models

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	provider "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB
var redisClient *redis.Client
var Tracer trace.Tracer

func SetTracerProvider(t *provider.TracerProvider) {
	log.Info().Msg("SetTracerProvider")
	Tracer = t.Tracer("game-server")
}

func Database(retry bool, redis *redis.Client) (*gorm.DB, error) {
	// Use cached value if we can
	if db != nil {
		return db, nil
	}

	if redis != nil {
		redisClient = redis
	}

	_, sp := Tracer.Start(context.Background(), "setup-database")
	defer sp.End()

	DB_HOST := os.Getenv("DB_HOST")
	DB_PORT := os.Getenv("DB_PORT")
	DB_USER := os.Getenv("DB_USER")
	DB_PASSWORD := os.Getenv("DB_PASSWORD")
	DB_NAME := os.Getenv("DB_NAME")

	dsn := DB_USER + ":" + DB_PASSWORD + "@tcp(" + DB_HOST + ":" + DB_PORT + ")/" + DB_NAME + "?charset=utf8mb4&parseTime=True&loc=Local"
	dbase, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		if retry {
			log.Panic().Err(err).Msg("Failed to connect to database @ " + DB_HOST + ":" + DB_PORT)
		} else {
			time.Sleep(3 * time.Second)
			return Database(true, nil)
		}
	}

	// Cache this to re-use next time
	db = dbase

	if err := dbase.Use(otelgorm.NewPlugin()); err != nil {
		panic(err)
	}

	sql, err := dbase.DB()
	if err != nil {
		panic(err)
	}

	sql.SetMaxOpenConns(10)
	sql.SetMaxIdleConns(3)
	sql.SetConnMaxIdleTime(5 * time.Minute)

	return dbase, err
}

func RunMigrations(db *gorm.DB) {
	ctx, sp := Tracer.Start(context.Background(), "run-migrations")
	defer sp.End()

	db.WithContext(ctx).AutoMigrate(
		&User{},
		&Unit{},
		&Round{},
		&Building{},
		&Effect{},
		&Item{},
		&Resource{},
		&NewsItem{},
		&Rule{},
		&Shout{},
		&UserUnit{},
		&UserRound{},
		&UserBuilding{},
		&UserItem{},
		&UserLog{},
		&RoundResource{},
		&RoundMarketResource{},
		&UndergroundMarketAuction{},
		&UndergroundMarketPurchase{},
		&RoundUnit{},
		&RoundBuilding{},
		&Conversation{},
		&Message{},
		&Event{},
		&Ranking{},
	)
	fmt.Println("Ran Migrations")
}
