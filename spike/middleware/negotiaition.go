package middleware

import (
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/tevjef/uct-core/common/model"
)

const (
	JsonContentType     = "application/json"
	ProtobufContentType = "application/x-protobuf"

	TextPlainContentType  = "text/plain"
	TextHtmlContentType   = "text/html"
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
