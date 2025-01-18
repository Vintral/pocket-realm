package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/Vintral/pocket-realm/game/actions"
	"github.com/Vintral/pocket-realm/game/application"
	"github.com/Vintral/pocket-realm/game/market"
	"github.com/Vintral/pocket-realm/game/payloads"
	"github.com/Vintral/pocket-realm/game/player"
	"github.com/Vintral/pocket-realm/game/rankings"
	"github.com/Vintral/pocket-realm/game/social"
	"github.com/Vintral/pocket-realm/models"
	realmRedis "github.com/Vintral/pocket-realm/redis"
	"github.com/Vintral/pocket-realm/utilities"
	"github.com/redis/go-redis/v9"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"go.opentelemetry.io/otel/sdk/trace"
)

var traceProvider *trace.TracerProvider

func Testing() {
	fmt.Println("HEYO")
}

var connectedUsersByRound map[int]map[int]*models.User
var connectedUsers map[int]*models.User

func refreshUser(id uint) {
	delay := rand.Intn(5)
	time.Sleep(time.Duration(delay) * time.Second)

	if user, ok := connectedUsers[int(id)]; ok {
		user.Refresh()
	} else {
		log.Warn().Int("userid", int(id)).Msg("User not connected")
	}
}

func handleRoundUpdates(rdb *redis.Client) {
	pubsub := rdb.Subscribe(context.Background(), "ROUND_UPDATE")
	defer pubsub.Close()

	tracer := traceProvider.Tracer("handle-round-updates")
	for {
		msg, err := pubsub.ReceiveMessage(context.Background())
		if err != nil {
			panic(err)
		}

		_, span := tracer.Start(context.Background(), "round-updated")

		fmt.Println("RECEIVED MESSAGE :::", msg.Channel, msg.Payload)
		round, err := strconv.Atoi(msg.Payload)
		if err != nil {
			panic(err)
		}

		if players, ok := connectedUsersByRound[round]; ok {
			for _, p := range players {
				go refreshUser(p.ID)

				// fmt.Println("Send Payload to", p.ID)
				// ctx := context.WithValue(context.Background(), utilities.KeyTraceProvider{}, traceProvider)
				// ctx = context.WithValue(ctx, utilities.KeyUser{}, p)

				// models.LoadRoundForUser(ctx)
				// //player.Load(ctx)
			}
		}

		span.End()
	}
}

func main() {
	//==============================//
	//	Setup ENV variables					//
	//==============================//
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println(err)
	}

	setupLogs()

	connectedUsersByRound = make(map[int]map[int]*models.User)
	connectedUsers = make(map[int]*models.User)

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
			db, err := models.Database(false, nil)
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
	//models.Redis = redisClient

	fmt.Println("Connected to Redis")
	fmt.Println(redisClient)

	go handleRoundUpdates(redisClient)

	//==============================//
	//	Run Database migrations			//
	//==============================//
	db, err := models.Database(false, redisClient)
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
	//	Module Initializations			//
	//==============================//
	social.Initialize(tp)
	rankings.Initialize(tp, db)

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
			log.Warn().AnErr("Error Upgrading Websocket", err).Send()
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

	if user.RoundID > 0 {
		_, ok := connectedUsersByRound[user.RoundID]
		if !ok {
			connectedUsersByRound[user.RoundID] = make(map[int]*models.User)
		}

		connectedUsersByRound[user.RoundID][int(user.ID)] = user
	}
	connectedUsers[int(user.ID)] = user

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

			delete(connectedUsersByRound[user.RoundID], int(user.ID))
			delete(connectedUsers, int(user.ID))
			return
		}

		log.Info().Msg("Message:" + string(messageContent))

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
		case "BUILD":
			actions.Build(ctx)
		case "BUY_RESOURCE":
			market.BuyResource(ctx)
		case "BUY_AUCTION":
			market.BuyAuction(ctx)
		case "EXPLORE":
			actions.Explore(ctx)
		case "GATHER":
			actions.Gather(ctx)
		case "GET_CONVERSATIONS":
			social.GetConversations(ctx)
		case "GET_EVENTS":
			player.GetEvents(ctx)
		case "GET_MESSAGES":
			social.GetMessages(ctx)
		case "GET_RANKINGS":
			rankings.RetrieveRankings(ctx)
		case "GET_ROUNDS":
			application.GetRounds(ctx)
		case "GET_UNDERGROUND_MARKET":
			market.GetUndergroundAuctions(ctx)
		case "MARK_EVENT_SEEN":
			player.HandleMarkEventSeen(ctx)
		case "MARKET_INFO":
			market.GetInfo(ctx)
		case "MESSAGE":
			social.SendMessage(ctx)
		case "NEWS":
			application.GetNews(user)
		case "PING":
			conn.WriteMessage(1, []byte("{ \"type\":\"PONG\"}"))
		case "PLAY_ROUND":
			player.PlayRound(ctx)
		case "RECRUIT":
			actions.Recruit(ctx)
		case "ROUND":
			models.LoadRoundForUser(ctx)
		case "RULES":
			application.GetRules(user)
		case "SELL_RESOURCE":
			market.SellResource(ctx)
		case "SHOUT":
			social.SendShout(ctx)
		case "SHOUTS":
			social.GetShouts(user)
		case "SUBSCRIBE_SHOUTS":
			social.SubscribeShouts(ctx)
		case "TESTING":
			conn.WriteMessage(1, []byte("OK"))
		case "UNSUBSCRIBE_SHOUTS":
			social.UnsubscribeShouts(ctx)
		case "GET_SELF":
			fallthrough
		case "LOAD_USER":
			player.Load(ctx)
		default:
			log.Warn().Msg("Unhandled Command: " + payload.Type)
		}
	}
}

func setupLogs() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	output := zerolog.ConsoleWriter{
		Out:           os.Stderr,
		FieldsExclude: []string{zerolog.TimestampFieldName},
	}

	log.Logger = log.Output(output).With().Logger()
}
