package models

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	provider "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB
var Tracer trace.Tracer

func SetTracerProvider(t *provider.TracerProvider) {
	fmt.Println("SetTracerProvider")
	Tracer = t.Tracer("game-server")
}

func Database(retry bool) (*gorm.DB, error) {
	// Use cached value if we can
	if db != nil {
		return db, nil
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
			fmt.Println(err)
			panic("Failed to connect to database @ " + DB_HOST + ":" + DB_PORT)
		} else {
			time.Sleep(3 * time.Second)
			return Database(true)
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
		&RoundUnit{},
		&RoundBuilding{},
		&Conversation{},
		&Message{},
	)
	fmt.Println("Ran Migrations")
}
