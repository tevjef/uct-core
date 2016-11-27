package middleware

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/pquerna/ffjson/ffjson"
	"strconv"
	"time"
	"uct/common/model"
	"strings"
)

const (
	JsonContentType = "application/json"
	ProtobufContentType = "application/x-protobuf"

	TextPlainContentType = "text/plain"
	TextHtmlContentType = "text/html"
	JavascriptContentType = "application/javascript"

	ServingFromCache = "servingFromCache"
	ResponseKey      = "response"
	MetaKey          = "meta"

	contentTypeHeader   = "Content-Type"
	contentLengthHeader = "Content-Length"
)

func ContentNegotiation(contentType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if _, exists := c.Get(ServingFromCache); exists {
			return
		}

		var responseType string

		for _, val := range c.Request.Header["Accept"] {
			if strings.Contains(val, ProtobufContentType) {
				responseType = ProtobufContentType
			} else if strings.Contains(val, TextHtmlContentType) {
				responseType = JsonContentType
			} else if strings.Contains(val, JsonContentType) {
				responseType = JsonContentType
			}
		}

		if responseType == "" {
			responseType = contentType
		}

		if value, exists := c.Get(ResponseKey); exists {
			if response, ok := value.(model.Response); ok {
				// Write status header
				c.Writer.WriteHeader(int(*response.Meta.Code))

				var responseData []byte
				var err error

				if responseType == ProtobufContentType {
					if responseData, err = response.Marshal(); err != nil {
						log.WithError(err).Errorln("error while parsing protobuf response")
					}
				} else if responseType == JsonContentType {
					if responseData, err = ffjson.Marshal(response); err != nil {
						log.WithError(err).Errorln("error while parsing json response")
					}
				}

				// Write Headers
				c.Header(contentLengthHeader, strconv.Itoa(len(responseData)))
				c.Header(contentTypeHeader, responseType)

				// Write data and flush
				c.Writer.Write(responseData)
				c.Writer.Flush()
			}
		}
	}
}

// Each handler must either set a meta or response
func Decorator(c *gin.Context) {
	c.Next()
	meta := model.Meta{}
	var metaExists bool
	var responseExists bool
	var value interface{}

	if value, metaExists = c.Get(MetaKey); metaExists {
		meta = value.(model.Meta)
	}

	if value, responseExists = c.Get(ResponseKey); responseExists {
		response, _ := value.(model.Response)
		code := int32(200)
		meta.Code = &code
		response.Meta = &meta
		c.Set(ResponseKey, response)
	} else {
		c.Set(ResponseKey, model.Response{Meta: &meta})
	}

	if !metaExists && !responseExists {
		c.Set(ServingFromCache, true)
	}
}

func Ginrus(logger *log.Logger, timeFormat string, utc bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		// some evil middlewares modify this values
		path := c.Request.URL.Path

		c.Next()

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
			"elapsed":    latency.Seconds() * 1e3,
			"latency":    latency,
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
