package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync/atomic"

	"cloud.google.com/go/pubsub"
)

// pubsub docs are here:
// https://pkg.go.dev/cloud.google.com/go/pubsub

var ctx = context.Background()

func main() {
	projectID := os.Getenv("PROJECT_ID")
	subID := os.Getenv("SUBSCRIPTION")

	if projectID == "" {
		log.Fatal("PROJECT_ID not set")
	}
	if subID == "" {
		log.Fatal("SUBSCRIPTION not set")
	}

	log.Fatal(pullMsgs(projectID, subID))
}

func pullMsgs(projectID, subID string) error {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("pubsub.NewClient: %v", err)
	}
	defer client.Close()

	sub := client.Subscription(subID)

	// Receive messages for 10 seconds, which simplifies testing.
	// Comment this out in production, since `Receive` should
	// be used as a long running operation.
	//ctx, cancel := context.WithTimeout(ctx, 360*time.Second)
	//defer cancel()

	var received int32
	err = sub.Receive(ctx, func(_ context.Context, msg *pubsub.Message) {
		log.Printf("Got message: %q\n", string(msg.Data))
		log.Println(msg.Attributes)

		atomic.AddInt32(&received, 1)
		msg.Ack()

		// At this point we would access the secret and do something useful
		// with it
	})
	if err != nil {
		return fmt.Errorf("sub.Receive: %v", err)
	}
	log.Printf("Received %d messages\n", received)

	return nil
}
