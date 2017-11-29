package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/tevjef/uct-core/common/model"
	"fmt"
	log "github.com/Sirupsen/logrus"
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
		log.WithError(fmt.Errorf("%s", *meta.Message)).Error("response does not exist")
		c.Set(ResponseKey, model.Response{Meta: &meta})
	}
}
