package middleware

import (
	"net"
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
			"httpRequest": log.Fields{
				"requestMethod": c.Request.Method,
				"requestUrl":    c.Request.URL.String(),
				"requestSize":   c.Request.ContentLength,
				"responseSize":  c.Writer.Size(),
				"status":        c.Writer.Status(),
				"userAgent":     c.Request.UserAgent(),
				"remoteIp":      c.ClientIP(),
				"serverIp":      GetOutboundIP().String(),
				"latency":       latency.String(),
			},
		})

		if len(c.Errors) > 0 {
			// Append error field if this is an erroneous request.
			entry.Error(c.Errors.String())
		} else {
			entry.Infof("%s %s", c.Request.Method, path)
		}
	}
}

// Get preferred outbound ip of this machine
func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}
