package middleware

import (
	"time"

	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
)

func Ginrus(logger *log.Logger, timeFormat string, utc bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		// some evil middlewares modify this values
		path := c.Request.URL.Path

		parts := strings.Split(path, "/")
		var handler string
		if len(parts) > 2 {
			handler = strings.Join(parts[:3], "/")
		}

		c.Next()

		if header := c.Request.Header.Get("X-Health-Check"); header != "" {
			return
		}

		end := time.Now()
		latency := end.Sub(start)
		if utc {
			end = end.UTC()
		}

		entry := logger.WithFields(log.Fields{
			"status":     c.Writer.Status(),
			"method":     c.Request.Method,
			"path":       path,
			"ip":         c.ClientIP(),
			"latency":    latency,
			"handler":    handler,
			"user-agent": c.Request.UserAgent(),
			"time":       end.Format(timeFormat),
		})

		if len(c.Errors) > 0 {
			// Append error field if this is an erroneous request.
			entry.Error(c.Errors.String())
		} else {
			entry.Info()
		}
	}
}
