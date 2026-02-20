module github.com/uwork/gorond

go 1.24.7

require (
	github.com/aws/aws-sdk-go v1.55.8
	github.com/robfig/cron v1.2.0
	gopkg.in/gcfg.v1 v1.2.3
	gopkg.in/warnings.v0 v0.1.2
)

replace github.com/aws/aws-sdk-go => ./internal/aws-stub
