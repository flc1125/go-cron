module distributednooverlapping

go 1.22

require (
	github.com/flc1125/go-cron/v4 v4.1.0
	github.com/stretchr/testify v1.10.0
)

replace github.com/flc1125/go-cron/v4 => ../../

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
