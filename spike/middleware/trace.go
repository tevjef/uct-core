package middleware

import (
	"github.com/gin-gonic/gin"
	"cloud.google.com/go/trace"
	"strings"
)

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
		c.Next()
		span.Finish()
	}
}