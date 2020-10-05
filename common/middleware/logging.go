package middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func Ginrus() gin.HandlerFunc {
	return func(c *gin.Context) {

		start := time.Now()
		// some evil middlewares modify this values
		path := c.Request.URL.Path

		parts := strings.Split(path, "/")
		var handler string
		if len(parts) > 2 {
			handler = strings.Join(parts[:3], "/")
		}

		if handler == "" {
			handler = "/"
		}

		c.Next()

		if header := c.Request.Header.Get("X-Health-Check"); header != "" {
			return
		}

		latency := time.Since(start)

		entry := log.WithFields(log.Fields{
			"status":     c.Writer.Status(),
			"method":     c.Request.Method,
			"path":       path,
			"ip":         c.ClientIP(),
			"latency":    latency,
			"handler":    handler,
			"user-agent": c.Request.UserAgent(),
		})

		if len(c.Errors) > 0 {
			// Append error field if this is an erroneous request.
			entry.Error(c.Errors.String())
		} else {
			entry.Infof("%s %s", c.Request.Method, path)
		}
	}
}
