package main

import (
	"context"
	"go.temporal.io/sdk/temporal"
	"hello-world-project-template-go/app/app"
	"hello-world-project-template-go/app/constants"
	"log"

	"go.temporal.io/sdk/client"
)

func main() {
	// Create the client object just once per process
	c, err := client.NewClient(client.Options{
		HostPort: constants.FrontEndHostPort,
	})
	if err != nil {
		log.Fatalln("unable to create Temporal client", err)
	}
	defer c.Close()
	options := client.StartWorkflowOptions{
		ID:        "test-signal-workflow",
		TaskQueue: app.GreetingTaskQueue,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 5,
		},
	}
	_, err = c.ExecuteWorkflow(context.Background(), options, app.YourWorkflowDefinition)
	if err != nil {
		log.Fatalln("unable to complete Workflow", err)
	}
}
