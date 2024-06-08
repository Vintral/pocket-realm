//module realm
module github.com/Vintral/pocket-realm/game

go 1.21.5

require (
	go.opentelemetry.io/otel v1.25.0
	go.opentelemetry.io/otel/sdk v1.25.0
	google.golang.org/grpc v1.63.2
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/Vintral/pocket-realm/models v0.0.0-20240608100744-7e43754634e7 // indirect
	github.com/Vintral/pocket-realm/payloads v0.0.0-20240608101223-c30295234780 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-redis/redis v6.15.9+incompatible // indirect
	github.com/go-sql-driver/mysql v1.8.1 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.19.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/redis/go-redis v6.15.9+incompatible // indirect
	github.com/uptrace/opentelemetry-go-extra/otelsql v0.2.4 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.25.0 // indirect
	go.opentelemetry.io/proto/otlp v1.2.0 // indirect
	golang.org/x/sys v0.19.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240415180920-8c6c420018be // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240415180920-8c6c420018be // indirect
	google.golang.org/protobuf v1.33.0 // indirect
)

require (
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/uuid v1.6.0
	github.com/gorilla/websocket v1.5.1
	github.com/joho/godotenv v1.5.1
	github.com/redis/go-redis/v9 v9.5.1
	github.com/uptrace/opentelemetry-go-extra/otelgorm v0.2.4
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.25.0
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v1.25.0
	go.opentelemetry.io/otel/metric v1.25.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.25.0
	go.opentelemetry.io/otel/trace v1.25.0
	golang.org/x/net v0.24.0 // indirect
	gorm.io/driver/mysql v1.5.6
	gorm.io/gorm v1.25.9
)

replace github.com/Vintral/pocket-realm/models => ../models

replace github.com/Vintral/pocket-realm/payloads => ../payloads
