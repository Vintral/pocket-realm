package main

import (
	"fmt"
	"time"

	realmTest "github.com/Vintral/pocket-realm/test"
	"github.com/go-co-op/gocron/v2"
)

const interval = 15

func _5_minute() {
	fmt.Println("5 Minute")
}

func _1_minute() {
	fmt.Println("1 Minute")

	currentTime := time.Now()
	minute := currentTime.Minute()
	fmt.Println("Minute is:", minute)

	checks := [5]int{1, 5, 10, 15, 30}
	for _, v := range checks {
		fmt.Println(v, "Minute check:", minute%v == 0)
	}
}

func main() {
	realmTest.Test()

	scheduler, err := gocron.NewScheduler()
	defer func() { _ = scheduler.Shutdown() }()

	if err != nil {
		panic("Error creating crons")
	}

	_, _ = scheduler.NewJob(
		gocron.CronJob(
			"*/5 * * * *",
			false,
		),
		gocron.NewTask(
			_5_minute,
		),
	)

	_, _ = scheduler.NewJob(
		gocron.CronJob(
			"* * * * *",
			false,
		),
		gocron.NewTask(
			_1_minute,
		),
	)

	fmt.Println("Starting up...")
	scheduler.Start()

	for {
	}
}
