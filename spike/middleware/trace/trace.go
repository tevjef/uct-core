package trace

import (
	"cloud.google.com/go/trace"
	"context"
	"github.com/gin-gonic/gin"
	"strings"
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
	return span.NewChild(spanName)
}

// Each handler must either set a meta or response
func Trace(traceClient *trace.Client) gin.HandlerFunc {
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

		span := traceClient.NewSpan(handler)
		ToContext(c, span)
		c.Next()
		span.Finish()
	}
}
