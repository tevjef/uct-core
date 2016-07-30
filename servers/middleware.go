package servers

import (
	"github.com/gin-gonic/gin"
	"github.com/pquerna/ffjson/ffjson"
	"strconv"
	uct "uct/common"
)

const (
	jsonContentType     = "application/json; charset=utf-8"
	protobufContentType = "application/x-protobuf"

	ServingFromCache = "servingFromCache"
	ResponseKey = "response"
	MetaKey = "meta"

	contentTypeHeader   = "Content-Type"
	contentLengthHeader = "Content-Length"
)

func ProtobufWriter() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if _, exists := c.Get(ServingFromCache); exists {
			return
		}

		if value, exists := c.Get(ResponseKey); exists {
			if response, ok := value.(uct.Response); ok {
				// Write status header
				c.Writer.WriteHeader(int(*response.Meta.Code))

				// Serialize response
				b, err := response.Marshal()
				uct.LogError(err)

				// Write Headers
				c.Header(contentLengthHeader, strconv.Itoa(len(b)))
				c.Header(contentTypeHeader, protobufContentType)

				// Write data and flush
				c.Writer.Write(b)
				c.Writer.Flush()
			}
		}
	}
}

func JsonWriter() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if _, exists := c.Get(ServingFromCache); exists {
			return
		}

		if value, exists := c.Get(ResponseKey); exists {
			if response, ok := value.(uct.Response); ok {
				// Write status header
				c.Writer.WriteHeader(int(*response.Meta.Code))
				// Serialize response
				b, err := ffjson.Marshal(response)
				uct.LogError(err)

				// Write Headers
				c.Header(contentLengthHeader, strconv.Itoa(len(b)))
				c.Header(contentTypeHeader, jsonContentType)

				// Write data and flush
				c.Writer.Write(b)
				c.Writer.Flush()
			}
		}
	}
}

// Each handler must either set a meta or response
func ErrorWriter() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		meta := uct.Meta{}
		var metaExists bool
		var responseExists bool
		var value interface{}

		if value, metaExists = c.Get(MetaKey); metaExists {
			meta = value.(uct.Meta)
		}

		if value, responseExists = c.Get(ResponseKey); responseExists {
			response, _ := value.(uct.Response)
			code := int32(200)
			meta.Code = &code
			response.Meta = &meta
			c.Set(ResponseKey, response)
		} else {
			c.Set(ResponseKey, uct.Response{Meta: &meta})
		}

		if !metaExists && !responseExists {
			c.Set(ServingFromCache, true)
		}
	}
}
