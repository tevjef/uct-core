package trace

import (
	"context"
	"log"
	"os"
	"strings"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"github.com/gin-gonic/gin"
	"go.opencensus.io/trace"
)

// Setter defines a context that enables setting values.
type Setter interface {
	Set(string, interface{})
}

const key = "tracespan"

func FromContext(ctx context.Context) *trace.Span {
	if span := ctx.Value(key); span == nil {
		return nil
	} else {
		return span.(*trace.Span)
	}
}

func ToContext(s Setter, span *trace.Span) {
	s.Set(key, span)
}

func NewSpan(ctx context.Context, spanName string) *trace.Span {
	span := FromContext(ctx)
	span.SetName(spanName)
	return span
}

func Trace(traceClient trace.Exporter) gin.HandlerFunc {
	exporter, err := stackdriver.NewExporter(stackdriver.Options{
		ProjectID: os.Getenv("GOOGLE_CLOUD_PROJECT"),
	})
	if err != nil {
		log.Fatal(err)
	}
	trace.RegisterExporter(exporter)

	return func(c *gin.Context) {
		path := c.Request.URL.Path

		parts := strings.Split(path, "/")
		var handler string
		if len(parts) > 2 {
			handler = strings.Join(parts[:3], "/")
		}

		if handler == "" {
			handler = "/"
		}

		//span := traceClient.NewSpan(handler)
		//ToContext(c, span)
		//c.Next()
		//span.Finish()
	}
}
