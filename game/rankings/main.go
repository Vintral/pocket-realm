package rankings

import (
	realmRedis "github.com/Vintral/pocket-realm/redis"

	redisDef "github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/sdk/trace"
)

var redisClient *redisDef.Client
var tracerProvider *trace.TracerProvider

func Initialize(tp *trace.TracerProvider) {
	tracerProvider = tp

	rdb, err := realmRedis.Instance(tp)
	if err != nil {
		panic(err)
	}

	redisClient = rdb
}
