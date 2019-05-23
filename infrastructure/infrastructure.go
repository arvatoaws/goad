package infrastructure

import (
	"github.com/JanBerktold/goad/api"
	"math"
	"time"

	"github.com/JanBerktold/goad/goad/types"
	"github.com/JanBerktold/goad/result"
)

const DefaultRunnerAsset = "data/lambda.zip"

type Infrastructure interface {
	Setup() (teardown func(), err error)
	Run(settings api.LambdaSettings)
	GetQueueURL() string
	Receive(chan *result.LambdaResults)
	GetSettings() *types.TestConfig
}

type InvokeArgs struct {
	File string   `json:"file"`
	Args []string `json:"args"`
}

func InvokeLambdas(inf Infrastructure) {
	t := inf.GetSettings()
	currentID := 0
	for i := 0; i < t.Lambdas; i++ {
		region := t.Regions[i%len(t.Regions)]
		requests, requestsRemainder := divide(t.Requests, t.Lambdas)
		concurrency, _ := divide(t.Concurrency, t.Lambdas)
		execTimeout := t.Timelimit

		if requestsRemainder > 0 && i == t.Lambdas-1 {
			requests += requestsRemainder
		}

		settings := api.LambdaSettings{
			SqsURL: inf.GetQueueURL(),
			ConcurrencyCount:concurrency,
			MaxRequestCount: requests,
			StresstestTimeout:execTimeout,
			QueueRegion:t.Regions[0],
			LambdaRegion:region,
			ReportingFrequency:reportingFrequency(t.Lambdas),
			ClientTimeout:time.Duration(t.Timeout)*time.Second,
			RunnerID:currentID,
			RequestParameters:api.RequestParameters{
				URL: t.URL,
				RequestMethod: t.Method,
				RequestBody: t.Body,
				RequestHeaders: t.Headers,
			},
		}

		currentID++

		go inf.Run(settings)
	}
}

func Aggregate(i Infrastructure) chan *result.LambdaResults {
	results := make(chan *result.LambdaResults)
	go i.Receive(results)
	return results
}

func divide(dividend int, divisor int) (quotient, remainder int) {
	return dividend / divisor, dividend % divisor
}

func reportingFrequency(numberOfLambdas int) time.Duration {
	return time.Duration((math.Log2(float64(numberOfLambdas)) + 1)) * time.Second
}
