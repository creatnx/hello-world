package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/uuid"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/uber-go/tally/v4"
	"github.com/uber-go/tally/v4/prometheus"
	"go.temporal.io/sdk/client"
	sdktally "go.temporal.io/sdk/contrib/tally"
	"go.temporal.io/sdk/temporal"
	"hello-world-project-template-go/app/app"
	"log"
	"strconv"
	"time"
)

const concurrency = 5000

var nsPtr *string
var metricPort *string

func main() {
	concurPtr := flag.Int("c", concurrency, "concurrency")
	fePtr := flag.String("fe", "10.0.3.4:7233", "frontend")
	nsPtr = flag.String("ns", "default", "namespace")
	metricPort = flag.String("port", "9901", "metric port")

	flag.Parse()
	fmt.Println("testing with concurrency= ", *concurPtr)
	fmt.Println("connecting to =", *fePtr)
	fmt.Println("namespace =", *nsPtr)

	// Create the client object just once per process
	//hosts := []string{"35.247.177.37:7233"}
	//hosts := []string{"34.143.228.169:7233", "34.124.218.203:7233"}
	//hosts := []string{"34.143.202.235:7233", "34.143.239.158:7233"}
	hosts := []string{"10.0.4.4:7233"}
	for _, host := range hosts {
		go testWithConcur(host, *concurPtr/len(hosts))
	}
	for {
		time.Sleep(time.Second)
	}
}

func testWithConcur(hostPort string, concur int) {
	c, err := client.NewClient(client.Options{
		//HostPort: "34.143.202.235:7233",
		//HostPort: constants.FrontEndHostPort,
		HostPort:  hostPort,
		Namespace: *nsPtr,
		MetricsHandler: sdktally.NewMetricsHandler(newPrometheusScope(prometheus.Configuration{
			ListenAddress: "0.0.0.0:" + *metricPort,
			TimerType:     "histogram",
		})),
	})
	if err != nil {
		log.Fatalln("unable to create Temporal client", err)
	}
	defer c.Close()

	buffer := make(chan struct{}, concur)
	go putBuffer(buffer, concur)
	pullBuffer(buffer, c)
}

func putBuffer(buffer chan struct{}, concur int) {
	go func() {
		for {
			for i := 0; i < concur; i++ {
				buffer <- struct{}{}
			}
			time.Sleep(time.Second)
		}
	}()
}

func pullBuffer(buffer chan struct{}, c client.Client) {
	temp, _ := uuid.NewUUID()
	var i uint64 = 0
	for _ = range buffer {
		tempId := temp.String()[:10]
		wfId := strconv.FormatUint(i, 16) + "-" + tempId
		i += 1
		go startWorkflow(c, wfId, buffer)
	}
}

func startWorkflow(c client.Client, wfId string, buf chan struct{}) {
	defer func() {
		if f := recover(); f != nil {
			fmt.Println(f)
		}
	}()
	options := client.StartWorkflowOptions{
		ID:        wfId,
		TaskQueue: app.GreetingTaskQueue,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	}
	name := "World " + wfId
	//startTime := time.Now()
	_, err := c.ExecuteWorkflow(context.Background(), options, app.GreetingWorkflow, name)
	if err != nil {
		//fmt.Println("unable to complete Workflow", err.Error(), wfId)
		return
	} else {
		//fmt.Println("workflow done: ", wfId)
	}
	//fmt.Println("elapse:", time.Now().Sub(startTime), startTime)
}

func printResults(greeting string, workflowID, runID string) {
	fmt.Printf("\nWorkflowID: %s RunID: %s\n", workflowID, runID)
	fmt.Printf("\n%s\n\n", greeting)
}

func newPrometheusScope(c prometheus.Configuration) tally.Scope {
	reporter, err := c.NewReporter(
		prometheus.ConfigurationOptions{
			Registry: prom.NewRegistry(),
			OnError: func(err error) {
				log.Println("error in prometheus reporter", err)
			},
		},
	)
	if err != nil {
		log.Fatalln("error creating prometheus reporter", err)
	}
	scopeOpts := tally.ScopeOptions{
		CachedReporter:  reporter,
		Separator:       prometheus.DefaultSeparator,
		SanitizeOptions: &sanitizeOptions,
		Prefix:          "",
	}
	scope, _ := tally.NewRootScope(scopeOpts, time.Second)

	log.Println("prometheus metrics scope created")
	return scope
}

// tally sanitizer options that satisfy Prometheus restrictions.
// This will rename metrics at the tally emission level, so metrics name we
// use maybe different from what gets emitted. In the current implementation
// it will replace - and . with _
var (
	safeCharacters = []rune{'_'}

	sanitizeOptions = tally.SanitizeOptions{
		NameCharacters: tally.ValidCharacters{
			Ranges:     tally.AlphanumericRange,
			Characters: safeCharacters,
		},
		KeyCharacters: tally.ValidCharacters{
			Ranges:     tally.AlphanumericRange,
			Characters: safeCharacters,
		},
		ValueCharacters: tally.ValidCharacters{
			Ranges:     tally.AlphanumericRange,
			Characters: safeCharacters,
		},
		ReplacementCharacter: tally.DefaultReplacementCharacter,
	}
)
