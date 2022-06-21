package app

import (
	"fmt"
	"go.temporal.io/sdk/temporal"
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

	return result, err
}

func YourWorkflowDefinition(ctx workflow.Context) error {
	//...
	signal := 1994
	err := workflow.SignalExternalWorkflow(ctx, "greeting-workflow", "3ff7d1cf-8008-454a-beac-de284b742c16", "test-signal", signal).Get(ctx, nil)

	return err
}

func LocalSipWorkflow(ctx workflow.Context, overseaOrder uint64) error {
	var localOrderId uint64
	err := workflow.ExecuteChildWorkflow(ctx, CreatePrimaryPOrderWorkflow).Get(ctx, &localOrderId)
	if err != nil {
		fmt.Println("create P order failed", err)
		return err
	}

	signalChan := workflow.GetSignalChannel(ctx, "logistic-updated")
	selector := workflow.NewSelector(ctx)
	var signal uint64
	selector.AddReceive(signalChan, func(channel workflow.ReceiveChannel, more bool) {
		channel.Receive(ctx, &signal)
	})
	selector.Select(ctx)
	if signal > 0 {
		err = workflow.ExecuteChildWorkflow(ctx, AutoPushWorkflow).Get(ctx, nil)
		if err != nil {
			fmt.Println("auto push failed")
			return err
		}
	}
	return nil
}

func CreatePrimaryPOrderWorkflow(ctx workflow.Context, overseaOrderId uint64) (uint64, error) {
	options := workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 5,
		RetryPolicy:         &temporal.RetryPolicy{},
	}
	var createdPOrder uint64
	err := workflow.ExecuteActivity(ctx, CreateOrderActivity, overseaOrderId, options).Get(ctx, &createdPOrder)
	if err != nil {
		fmt.Println("create order failed")
	}

	executePayOption := workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 5,
		RetryPolicy:         &temporal.RetryPolicy{},
	}
	err = workflow.ExecuteActivity(ctx, PayOrderActivity, createdPOrder, executePayOption).Get(ctx, nil)
	if err != nil {
		fmt.Println("pay order failed")
	}
	return createdPOrder, err
}

func AutoPushWorkflow(ctx workflow.Context, localOrderId uint64, overseaOrderId uint64) error {
	err := workflow.ExecuteActivity(ctx, AutoPushActivity, localOrderId).Get(ctx, nil)
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = workflow.Sleep(ctx, 5*time.Second)
	if err != nil {
		return err
	}
	err = workflow.ExecuteActivity(ctx, CheckTn, overseaOrderId).Get(ctx, nil)
	return err
}

func CreateOrderActivity(overseaOrderId uint64) {}

func PayOrderActivity(orderId uint64) {}

func AutoPushActivity(localOrderId uint64) {}

func CheckTn(overseaOrderId uint64) {}
