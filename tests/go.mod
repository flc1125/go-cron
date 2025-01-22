module github.com/flc1125/go-cron/tests/v4

go 1.22.0

replace (
	github.com/flc1125/go-cron/middleware/distributednooverlapping/redismutex/v4 => ../middleware/distributednooverlapping/redismutex
	github.com/flc1125/go-cron/middleware/distributednooverlapping/v4 => ../middleware/distributednooverlapping
	github.com/flc1125/go-cron/middleware/nooverlapping/v4 => ../middleware/nooverlapping
	github.com/flc1125/go-cron/middleware/otel/v4 => ../middleware/otel
	github.com/flc1125/go-cron/middleware/recovery/v4 => ../middleware/recovery
	github.com/flc1125/go-cron/v4 => ../
)

require (
	github.com/flc1125/go-cron/middleware/distributednooverlapping/redismutex/v4 v4.3.2
	github.com/flc1125/go-cron/middleware/distributednooverlapping/v4 v4.3.2
	github.com/flc1125/go-cron/middleware/nooverlapping/v4 v4.3.2
	github.com/flc1125/go-cron/middleware/otel/v4 v4.3.2
	github.com/flc1125/go-cron/middleware/recovery/v4 v4.3.2
	github.com/flc1125/go-cron/v4 v4.3.2
	github.com/redis/go-redis/v9 v9.7.0
	github.com/stretchr/testify v1.10.0
	go.opentelemetry.io/otel/sdk v1.34.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel v1.34.0 // indirect
	go.opentelemetry.io/otel/metric v1.34.0 // indirect
	go.opentelemetry.io/otel/trace v1.34.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
