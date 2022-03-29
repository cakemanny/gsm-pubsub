package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/pubsub"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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

	// Run findSecret once to ensure permissions are set correctly
	_, err := findSecret(ctx, "")
	if err != nil {
		log.Fatal(err)
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

	err = sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		log.Println("Got message. Attrs:", msg.Attributes)
		eventType, _ := msg.Attributes["eventType"]
		if eventType != "SECRET_VERSION_ADD" {
			msg.Ack()
			return
		}
		fullSecretID, ok := msg.Attributes["secretId"]
		if !ok {
			msg.Ack()
			log.Println("Received SECRET_VERSION_ADD with no secretId")
			return
		}

		// have
		//   projects/827585297303/secrets/test-secret
		// want
		//   test-secret
		var projectNumber int
		var secretID string
		_, err := fmt.Sscanf(fullSecretID, "projects/%d/secrets/%s",
			&projectNumber, &secretID)
		if err != nil {
			msg.Ack()
			log.Printf("Failed to extract final portion of secretId %q\n", fullSecretID)
			return
		}
		log.Println("secretID =", secretID)

		secret, err := findSecret(ctx, secretID)
		if err != nil {
			msg.Ack()
			log.Printf("Error finding secret %q: %v\n", secretID, err)
			return
		}
		if secret == nil {
			log.Printf("No matching secret in kubernetes for %q\n", secretID)
			msg.Ack()
			return
		}

		log.Printf(
			"Found %q in namespace %q for gsm secret %q",
			secret.Name, secret.Namespace, secretID,
		)

		msg.Ack()
		// At this point we would access the secret and do something useful
		// with it
	})
	if err != nil {
		return fmt.Errorf("sub.Receive: %v", err)
	}

	return nil
}

func findSecret(ctx context.Context, gsmSecretID string) (*v1.Secret, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	kubeclient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	secretsList, err := kubeclient.CoreV1().Secrets(v1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %v", err)
	}
	var _ = secretsList
	for _, secretItem := range secretsList.Items {
		anno, ok := secretItem.Annotations["jenkins-x.io/gsm-secret-id"]
		if !ok {
			continue
		}
		if anno == gsmSecretID {
			return &secretItem, nil
		}
	}

	return nil, nil
}
