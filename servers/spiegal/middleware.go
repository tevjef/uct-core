package main

import (
	"fmt"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/tevjef/gin"
	"strconv"
)

const (
	jsonContentType     = "application/json; charset=utf-8"
	protobufContentType = "application/x-protobuf"
)

func protobufWriter() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if value, exists := c.Get("protobuf"); exists {
			b, ok := value.([]byte)
			if !ok {
				err := fmt.Errorf("%s", "Error while getting protobuf bytes")
				c.Error(gin.Error{err, gin.ErrorTypePublic, c.Request.URL.String()})
			}
			c.Header("Content-Length", strconv.Itoa(len(b)))
			c.Header("Content-Type", protobufContentType)
			c.Writer.Write(b)
		}

	}
}

func jsonWriter() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if value, exists := c.Get("object"); exists {
			if b, err := ffjson.Marshal(value); err != nil {
				err := fmt.Errorf("%s Error: %s", "Can't retrieve response data", err)
				c.Error(gin.Error{err, gin.ErrorTypePublic, c.Request.URL.String()})
			} else {
				c.Header("Content-Length", strconv.Itoa(len(b)))
				c.Header("Content-Type", jsonContentType)
				c.Writer.Write(b)
			}
		}

	}
}

func errorWriter() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) != 0 {
			debug := c.DefaultQuery("debug", "false")
			debugBool, _ := strconv.ParseBool(debug)
			if debugBool {
				c.String(500, "%s", c.Errors.String())
			} else {
				c.String(500, "%s ", "500 internal server error")
			}
		}
	}
}
