package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/pquerna/ffjson/ffjson"
	"strconv"
	uct "uct/common"
)

const (
	jsonContentType     = "application/json; charset=utf-8"
	protobufContentType = "application/x-protobuf"
)

func protobufWriter() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if value, exists := c.Get("response"); exists {
			if response, ok := value.(uct.Response); !ok {
				log.WithField("response", value).Errorln("Runtine type assertion failed")
			} else {
				c.Writer.WriteHeader(int(*response.Meta.Code))
				b, err := response.Marshal()
				uct.LogError(err)
				c.Header("Content-Length", strconv.Itoa(len(b)))
				c.Header("Content-Type", protobufContentType)
				c.Writer.Write(b)
			}
		}
	}
}

func jsonWriter() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if value, exists := c.Get("response"); exists {
			if response, ok := value.(uct.Response); !ok {
				log.WithField("response", value).Errorln("Runtine type assertion failed")
			} else {
				c.Writer.WriteHeader(int(*response.Meta.Code))
				b, err := ffjson.Marshal(response)
				uct.LogError(err)
				c.Header("Content-Length", strconv.Itoa(len(b)))
				c.Header("Content-Type", jsonContentType)
				c.Writer.Write(b)
			}
		} else {
			log.WithField("exists", exists).Debugln("Response data was not set")
			c.String(500, "Response data was not set")
		}
	}
}

func errorWriter() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		debug := c.DefaultQuery("debug", "false")
		debugBool, _ := strconv.ParseBool(debug)
		meta := uct.Meta{}

		if v, exists := c.Get("meta"); exists {
			meta = v.(uct.Meta)
			if len(c.Errors) != 0 && debugBool {
				errorJsonBytes, err := c.Errors.MarshalJSON()
				uct.LogError(err)
				errorJson := string(errorJsonBytes)
				meta.ErrorMessage = &errorJson
			}
		} else {
			if len(c.Errors) != 0 {
				code := int32(500)
				errorType := "Internal Server Error"
				meta.Code = &code
				meta.ErrorType = &errorType
			} else {
				code := int32(200)
				meta.Code = &code
			}
		}
		if value, exists := c.Get("response"); exists {
			response, _ := value.(uct.Response)
			response.Meta = &meta
			c.Set("response", response)
		} else {
			c.Set("response", uct.Response{Meta: &meta})
		}
	}
}
