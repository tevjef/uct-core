package middleware

import (
	"uct/common/model"

	"github.com/gin-gonic/gin"
)

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
}
