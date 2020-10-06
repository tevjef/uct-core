package publishing

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"time"

	"cloud.google.com/go/pubsub"
	log "github.com/sirupsen/logrus"
	"github.com/tevjef/uct-backend/common/middleware"
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

func PublishToHttp(context context.Context, url string, body io.Reader) (*log.Entry, error) {
	startTime := time.Now()
	client, err := idtoken.NewClient(context, url)
	if err != nil {
		return nil, fmt.Errorf("idtoken.NewClient: %v", err)
	}

	req, err := http.NewRequestWithContext(context, http.MethodPost, url, body)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	entry := log.WithFields(log.Fields{
		"httpRequest": log.Fields{
			"requestHeaders":  req.Header,
			"responseHeaders": resp.Header,
			"requestMethod":   resp.Request.Method,
			"requestUrl":      resp.Request.URL.String(),
			"requestSize":     resp.Request.ContentLength,
			"responseSize":    resp.ContentLength,
			"status":          resp.Status,
			"userAgent":       resp.Request.UserAgent(),
			"serverIp":        middleware.GetOutboundIP().String(),
			"latency":         time.Since(startTime),
		}})

	if resp.StatusCode > 300 {
		dump, _ := httputil.DumpResponse(resp, false)
		entry.WithError(err).
			WithField("httpDump", string(dump)).
			WithField("status", resp.Status).Debugln("api call failed")
	}

	return entry, nil
}
