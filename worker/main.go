package main

import (
	"flag"
	"fmt"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/uber-go/tally/v4"
	"github.com/uber-go/tally/v4/prometheus"
	"go.temporal.io/sdk/client"
	sdktally "go.temporal.io/sdk/contrib/tally"
	"go.temporal.io/sdk/worker"
	"hello-world-project-template-go/app/app"
	"log"
	"time"
)

const workerConcurrency = 10

func main() {
	fePtr := flag.String("fe", "10.0.3.4:7233", "frontend")
	nsPtr := flag.String("ns", "default", "namespace")
	metricPort := flag.String("port", "9001", "metric port")

	flag.Parse()
	fmt.Println("connecting to =", *fePtr)
	fmt.Println("namespace =", &nsPtr)

	// Create the client object just once per process
	c, err := client.NewClient(client.Options{
		HostPort:  *fePtr,
		Namespace: *nsPtr,
		MetricsHandler: sdktally.NewMetricsHandler(newPrometheusScope(prometheus.Configuration{
			ListenAddress: "0.0.0.0:" + *metricPort,
			TimerType:     "histogram",
		})),
		Logger: &myLogger{},
		//HostPort: constants.FrontEndHostPort,
	})
	if err != nil {
		log.Fatalln("unable to create Temporal client", err)
	}
	defer c.Close()

	worker.SetStickyWorkflowCacheSize(100000)
	worker.EnableVerboseLogging(false)

	// This worker hosts both Workflow and Activity functions
	w := worker.New(c, app.GreetingTaskQueue, worker.Options{
		MaxConcurrentActivityTaskPollers:        50,
		MaxConcurrentWorkflowTaskPollers:        50,
		MaxConcurrentWorkflowTaskExecutionSize:  10000,
		MaxConcurrentLocalActivityExecutionSize: 10000,
		MaxConcurrentActivityExecutionSize:      10000,
		MaxConcurrentSessionExecutionSize:       10000,
		//MaxConcurrentWorkflowTaskExecutionSize:  10,
		//MaxConcurrentActivityExecutionSize:      1000,
		//MaxConcurrentSessionExecutionSize:       1000,
		//MaxConcurrentLocalActivityExecutionSize: 1000,
	})
	w.RegisterWorkflow(app.GreetingWorkflow)
	w.RegisterWorkflow(app.YourWorkflowDefinition)
	w.RegisterActivity(app.ComposeGreeting)
	w.RegisterActivity(app.DelayHello)
	// Start listening to the Task Queue
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("unable to start Worker", err)
	}
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

type myLogger struct {
}

func (m myLogger) Debug(msg string, keyvals ...interface{}) {
	//skip
}

func (m myLogger) Info(msg string, keyvals ...interface{}) {
	//skip
}

func (m myLogger) Warn(msg string, keyvals ...interface{}) {
	//skip
}

func (m myLogger) Error(msg string, keyvals ...interface{}) {
	//skip
}
