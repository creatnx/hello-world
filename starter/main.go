package main

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"hello-world-project-template-go/app/app"
	"log"
	"strconv"
	"time"
)

const concurrency = 6000

func main() {
	// Create the client object just once per process
	hosts := []string{"34.143.202.235:7233", "34.143.239.158:7233"}
	for _, host := range hosts {
		go testWithConcur(host, concurrency/len(hosts))
	}
	for {
		time.Sleep(time.Second)
	}
}

func testWithConcur(hostPort string, concur int) {
	c, err := client.NewClient(client.Options{
		//HostPort: "34.143.202.235:7233",
		//HostPort: constants.FrontEndHostPort,
		HostPort: hostPort,
	})
	if err != nil {
		log.Fatalln("unable to create Temporal client", err)
	}
	defer c.Close()

	buffer := make(chan struct{}, concur)
	go putBuffer(buffer)
	pullBuffer(buffer, c)
}

func putBuffer(buffer chan struct{}) {
	go func() {
		time.Sleep(time.Second)
		for i := 0; i < concurrency; i++ {
			buffer <- struct{}{}
		}
	}()
}

func pullBuffer(buffer chan struct{}, c client.Client) {

	var i uint64 = 0
	for _ = range buffer {
		temp, _ := uuid.NewUUID()
		tempId := temp.String()[:20]
		wfId := tempId + "-" + strconv.FormatUint(i, 10)
		i += 1
		go startWorkflow(c, wfId, buffer)
	}
}

func startWorkflow(c client.Client, wfId string, buf chan struct{}) {
	//defer func() {
	//	if f := recover(); f != nil {
	//		fmt.Println(f)
	//	}
	//	buf <- struct{}{}
	//}()
	options := client.StartWorkflowOptions{
		ID:        wfId,
		TaskQueue: app.GreetingTaskQueue,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	}
	name := "World " + wfId
	startTime := time.Now()
	_, err := c.ExecuteWorkflow(context.Background(), options, app.GreetingWorkflow, name)
	if err != nil {
		fmt.Println("unable to complete Workflow", err.Error(), wfId)
		return
	} else {
		//fmt.Println("workflow done: ", wfId)
	}
	fmt.Println("elapse:", time.Now().Sub(startTime), startTime)
}

func printResults(greeting string, workflowID, runID string) {
	fmt.Printf("\nWorkflowID: %s RunID: %s\n", workflowID, runID)
	fmt.Printf("\n%s\n\n", greeting)
}
