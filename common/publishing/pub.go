package publishing

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"

	"cloud.google.com/go/pubsub"
	log "github.com/Sirupsen/logrus"
	"google.golang.org/api/idtoken"
)

func PublishMessage(token string, projectId string, topicId string, data string) {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectId)
	if err != nil {
		log.Errorf("pubsub.NewClient: %v", err)
		return
	}

	t := client.Topic(topicId)

	result := t.Publish(ctx, &pubsub.Message{
		Data: []byte(data),
	})

	fmt.Print(data)

	// The Get method blocks until a server-generated ID or
	// an error is returned for the published message.
	id, err := result.Get(ctx)
	if err != nil {
		// Error handling code can be added here.
		log.Debugln("failed to publish: %v", err)
		return
	}
	log.Debugln("Published message %d; msg ID: %v\n", id)
}

func PublishToHttp(url string, body io.Reader) error {
	ctx := context.Background()
	client, err := idtoken.NewClient(ctx, url)
	if err != nil {
		return fmt.Errorf("idtoken.NewClient: %v", err)
	}

	resp, err := client.Post(url, "", body)
	if err != nil {
		return err
	}

	if resp.StatusCode > 300 {
		log.WithError(err).WithField("status", resp.Status).Debugln("api call failed")
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.WithError(err).Fatal("failed to read response")
		return err
	}

	log.Debugln(string(b))

	return nil
}
