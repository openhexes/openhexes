module github.com/openhexes/openhexes/api

go 1.24.3

require (
	cloud.google.com/go/auth v0.16.3
	connectrpc.com/connect v1.17.0
	github.com/deckarep/golang-set/v2 v2.8.0
	github.com/openhexes/proto v0.0.0
	github.com/stretchr/testify v1.10.0
	go.opentelemetry.io/otel v1.37.0
	go.opentelemetry.io/otel/sdk v1.37.0
	google.golang.org/protobuf v1.36.6
)

require (
	cloud.google.com/go/compute/metadata v0.7.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.6 // indirect
	github.com/googleapis/gax-go/v2 v2.15.0 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel/metric v1.37.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250728155136-f173205681a0 // indirect
	google.golang.org/grpc v1.74.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require (
	connectrpc.com/cors v0.1.0
	connectrpc.com/otelconnect v0.7.2
	github.com/caarlos0/env/v11 v11.3.1
	github.com/exaring/otelpgx v0.9.3
	github.com/google/uuid v1.6.0
	github.com/hashicorp/golang-lru/v2 v2.0.7
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx-zap v0.0.0-20221202020421-94b1cb2f889f
	github.com/jackc/pgx/v5 v5.7.5
	github.com/rs/cors v1.11.1
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.62.0
	go.opentelemetry.io/otel/sdk/metric v1.37.0
	go.opentelemetry.io/otel/trace v1.37.0
	go.uber.org/zap v1.27.0
	golang.org/x/crypto v0.40.0 // indirect
	golang.org/x/net v0.42.0
	golang.org/x/text v0.28.0 // indirect
)

replace github.com/openhexes/proto v0.0.0 => ../proto/go
