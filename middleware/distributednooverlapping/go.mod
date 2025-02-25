module github.com/flc1125/go-cron/middleware/distributednooverlapping/v4

go 1.23.6

replace (
	github.com/flc1125/go-cron/crontest/v4 => ../../crontest
	github.com/flc1125/go-cron/v4 => ../../
)

require (
	github.com/flc1125/go-cron/crontest/v4 v4.4.1
	github.com/flc1125/go-cron/v4 v4.4.1
	github.com/stretchr/testify v1.10.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
