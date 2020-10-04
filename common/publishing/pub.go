package publishing

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"cloud.google.com/go/pubsub"
	log "github.com/sirupsen/logrus"
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

func PublishToHttp(context context.Context, url string, body io.Reader) error {
	//http.DefaultTransport = &ochttp.Transport{
	//	// Use Google Cloud propagation format.
	//	Propagation: &propagation.HTTPFormat{},
	//}

	client, err := idtoken.NewClient(context, url)
	if err != nil {
		return fmt.Errorf("idtoken.NewClient: %v", err)
	}

	req, err := http.NewRequestWithContext(context, http.MethodPost, url, body)
	resp, err := client.Do(req)
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
