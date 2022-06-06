package main

import (
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"hello-world-project-template-go/app/app"
	"hello-world-project-template-go/app/constants"
	"log"
)

const workerConcurrency = 10

func main() {
	// Create the client object just once per process
	c, err := client.NewClient(client.Options{
		HostPort: constants.FrontEndHostPort,
	})
	if err != nil {
		log.Fatalln("unable to create Temporal client", err)
	}
	defer c.Close()

	// This worker hosts both Workflow and Activity functions
	w := worker.New(c, app.GreetingTaskQueue, worker.Options{
		MaxConcurrentActivityTaskPollers:        100,
		MaxConcurrentWorkflowTaskPollers:        100,
		MaxConcurrentWorkflowTaskExecutionSize:  100000,
		MaxConcurrentLocalActivityExecutionSize: 100000,
		MaxConcurrentActivityExecutionSize:      100000,
		MaxConcurrentSessionExecutionSize:       100000,
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
