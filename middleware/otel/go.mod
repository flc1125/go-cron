module github.com/flc1125/go-cron/v4/middleware/otel

go 1.22

require (
	github.com/flc1125/go-cron/v4 v4.1.0
	go.opentelemetry.io/otel/trace v1.32.0
)

replace github.com/flc1125/go-cron/v4 => ../../

require (
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	go.opentelemetry.io/otel v1.32.0
	go.opentelemetry.io/otel/metric v1.32.0 // indirect
)
