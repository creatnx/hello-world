package app

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
)

func GreetingWorkflow(ctx workflow.Context, name string) (string, error) {
	options := workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 5,
	}
	ctx = workflow.WithActivityOptions(ctx, options)
	var result string
	err := workflow.ExecuteActivity(ctx, ComposeGreeting, name).Get(ctx, &result)

	err = workflow.Sleep(ctx, time.Second*10)
	if err != nil {
		fmt.Println(err)
	}
	var delayHello string
	delayOption := workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 10,
	}
	ctx = workflow.WithActivityOptions(ctx, delayOption)
	err = workflow.ExecuteActivity(ctx, DelayHello, name).Get(ctx, &delayHello)

	return result, err
}

func YourWorkflowDefinition(ctx workflow.Context) error {
	//...
	signal := 1994
	err := workflow.SignalExternalWorkflow(ctx, "greeting-workflow", "3ff7d1cf-8008-454a-beac-de284b742c16", "test-signal", signal).Get(ctx, nil)

	return err
}
