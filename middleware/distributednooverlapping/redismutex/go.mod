module github.com/flc1125/go-cron/middleware/distributednooverlapping/redismutex/v4

go 1.25.0

replace (
	github.com/flc1125/go-cron/middleware/distributednooverlapping/v4 => ../
	github.com/flc1125/go-cron/v4 => ../../../
)

require (
	github.com/flc1125/go-cron/middleware/distributednooverlapping/v4 v4.10.0
	github.com/flc1125/go-cron/v4 v4.10.0
	github.com/redis/go-redis/v9 v9.20.0
	github.com/stretchr/testify v1.11.1
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
