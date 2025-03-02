module github.com/flc1125/go-cron/middleware/distributednooverlapping/redismutex/v4

go 1.23.0

replace (
	github.com/flc1125/go-cron/crontest/v4 => ../../../crontest
	github.com/flc1125/go-cron/middleware/distributednooverlapping/v4 => ../
	github.com/flc1125/go-cron/v4 => ../../../
)

require (
	github.com/flc1125/go-cron/middleware/distributednooverlapping/v4 v4.5.0
	github.com/flc1125/go-cron/v4 v4.5.0
	github.com/redis/go-redis/v9 v9.7.1
	github.com/stretchr/testify v1.10.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
