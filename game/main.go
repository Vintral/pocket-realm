package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/Vintral/pocket-realm/game/actions"
	"github.com/Vintral/pocket-realm/game/application"
	"github.com/Vintral/pocket-realm/game/player"
	realmRedis "github.com/Vintral/pocket-realm/game/redis"
	"github.com/Vintral/pocket-realm/game/social"
	"github.com/Vintral/pocket-realm/game/utilities"
	"github.com/Vintral/pocket-realm/models"
	"github.com/Vintral/pocket-realm/payloads"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"go.opentelemetry.io/otel/sdk/trace"
)

var traceProvider *trace.TracerProvider

func Testing() {
	fmt.Println("HEYO")
}

func main() {
	//==============================//
	//	Setup ENV variables					//
	//==============================//
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println(err)
	}

	var port = "3001"
	if os.Getenv("DEPLOYED") == "1" {
		port = "8080"
	}

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	ctx := context.Background()
	_, cancel := context.WithCancel(ctx)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	defer func() {
		signal.Stop(signalChan)
		cancel()
	}()
	go func() {
		select {
		case <-signalChan:
			db, err := models.Database(false)
			if err != nil {
				sql, err := db.DB()
				if err != nil {
					sql.Close()
				}
			}

			cancel()
			os.Exit(1)
		}
	}()

	//==============================//
	//	Setup Telemetry							//
	//==============================//
	otelShutdown, tp, err := setupOTelSDK(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		fmt.Println("In shutdown")
		err = errors.Join(err, otelShutdown(context.Background()))
	}()
	models.SetTracerProvider(tp)
	traceProvider = tp

	//==============================//
	//	Setup Metrics								//
	//==============================//
	// userCounter, err = metrics.Int64UpDownCounter("users.online")
	// if err != nil {
	// 	panic(err)
	// }

	//==============================//
	//	Setup Redis									//
	//==============================//
	redisClient, err := realmRedis.Instance(tp)
	if err != nil {
		panic(err)
	}
	fmt.Println("Connected to Redis")
	fmt.Println(redisClient)

	social.Initialize(tp)

	//==============================//
	//	Run Database migrations			//
	//==============================//
	db, err := models.Database(false)
	if err != nil {
		panic(err)
	}
	models.RunMigrations(db)

	// fmt.Println("Env Variables:")
	// fmt.Println(os.Environ())

	// var user = models.User{Email: "jeffrey.heater@gmail.com"}
	// user.Load(db)
	// user.Dump()

	// uuid, _ := uuid.Parse("be31cd63-5608-438a-bf8c-65157e944558")
	// round, _ := models.LoadRoundByGuid(context.Background(), uuid)
	// fmt.Println(round)

	//==============================//
	//	Setup Websocket Server			//
	//==============================//
	http.HandleFunc("/testing", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Received Request!")
		//io.WriteString(w, "OK")
		w.Write([]byte("OK"))
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Print("Received Request")

		websocket, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		listen(websocket)
	})
	fmt.Println("Server started on", port)
	http.ListenAndServe(":"+port, nil)
}

func listen(conn *websocket.Conn) {
	var u *models.User
	user := u.Load()
	if user == nil {
		fmt.Print("ERROR LOADING USER")
		return
	}
	user.Connection = conn

	for {
		// read a message
		_, messageContent, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				fmt.Println("Unexpected Close")
			} else {
				fmt.Println("STill here")
				fmt.Println(err)
			}

			return
		}

		fmt.Println("Message:", string(messageContent))

		ctx := context.WithValue(context.Background(), utilities.KeyTraceProvider{}, traceProvider)
		ctx = context.WithValue(ctx, utilities.KeyUser{}, user)
		ctx = context.WithValue(ctx, utilities.KeyPayload{}, messageContent)

		var payload payloads.Payload
		err = json.Unmarshal(messageContent, &payload)
		if err != nil {
			fmt.Println(err)
			return
		}

		switch payload.Type {
		case "TESTING":
			conn.WriteMessage(1, []byte("OK"))
		case "PING":
			conn.WriteMessage(1, []byte("{ \"type\":\"PONG\"}"))
		case "ROUND":
			models.LoadRoundForUser(ctx)
		case "EXPLORE":
			actions.Explore(ctx)
		case "GATHER":
			actions.Gather(ctx)
		case "RECRUIT":
			actions.Recruit(ctx)
		case "BUILD":
			actions.Build(ctx)
		case "GET_CONVERSATIONS":
			social.GetConversations(ctx)
		case "GET_MESSAGES":
			social.GetMessages(ctx)
		case "SHOUT":
			social.SendShout(ctx)
		case "SHOUTS":
			social.GetShouts(user)
		case "SUBSCRIBE_SHOUTS":
			social.SubscribeShouts(ctx)
		case "UNSUBSCRIBE_SHOUTS":
			social.UnsubscribeShouts(ctx)
		case "RULES":
			application.GetRules(user)
		case "NEWS":
			application.GetNews(user)
		case "GET_SELF":
			fallthrough
		case "LOAD_USER":
			player.Load(ctx)
			// player.Get
			// fmt.Println("Get User")
			// if data, err := json.Marshal(user); err == nil {
			// 	before := []byte("{\"type\":\"USER_DATA\",\"data\":{\"user\":")
			// 	after := []byte("}}")
			// 	conn.WriteMessage(1, []byte(append(append(before, data...), after...)))
			// } else {
			// 	fmt.Println(err)
			// 	conn.WriteMessage(1, []byte("{\"type\":\"ERROR_USER\"}"))
			// }
		default:
			fmt.Println("Unhandled Command:", payload.Type)
		}
	}
}
