package main

import (
	"github.com/gin-gonic/gin"
	"github.com/pquerna/ffjson/ffjson"
	"strconv"
	uct "uct/common"
)

const (
	jsonContentType     = "application/json; charset=utf-8"
	protobufContentType = "application/x-protobuf"

	servingFromCache = "servingFromCache"
	responseKey      = "response"
	metaKey          = "meta"

	contentTypeHeader   = "Content-Type"
	contentLengthHeader = "Content-Length"
)

func protobufWriter() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if _, exists := c.Get(servingFromCache); exists {
			return
		}

		if value, exists := c.Get(responseKey); exists {
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

func jsonWriter() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if _, exists := c.Get(servingFromCache); exists {
			return
		}

		if value, exists := c.Get(responseKey); exists {
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
func errorWriter() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		meta := uct.Meta{}
		var metaExists bool
		var responseExists bool
		var value interface{}

		if value, metaExists = c.Get(metaKey); metaExists {
			meta = value.(uct.Meta)
			if len(c.Errors) != 0 {
				errorJsonBytes := c.Errors.String()
				errorJson := string(errorJsonBytes)
				meta.ErrorMessage = &errorJson
			}
		}

		if value, responseExists = c.Get(responseKey); responseExists {
			response, _ := value.(uct.Response)
			code := int32(200)
			meta.Code = &code
			response.Meta = &meta
			c.Set(responseKey, response)
		} else {
			c.Set(responseKey, uct.Response{Meta: &meta})
		}

		if !metaExists && !responseExists {
			c.Set(servingFromCache, true)
		}
	}
}
