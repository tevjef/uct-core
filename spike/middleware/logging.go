package middleware

import (
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
)

var (
	httpResponsesLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "http_response_latency",
		Help: "Measure http response latencies",
	}, []string{"status", "method", "handler"})

	httpRequestTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_request_total_count",
		Help: "Counts request total",
	}, []string{"status", "method", "handler"})
)

func init() {
	prometheus.MustRegister(httpResponsesLatency)
	prometheus.MustRegister(httpRequestTotal)
}

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

		labels := prometheus.Labels{
			"status":  strconv.Itoa(c.Writer.Status()),
			"method":  c.Request.Method,
			"handler": handler,
		}

		httpResponsesLatency.With(labels).Observe(latency.Seconds())
		httpRequestTotal.With(labels).Inc()

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
			entry.Info()
		}
	}
}
