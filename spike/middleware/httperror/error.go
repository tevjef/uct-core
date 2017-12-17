package httperror

import (
	"github.com/gin-gonic/gin"
	"github.com/tevjef/uct-core/common/model"
	"github.com/tevjef/uct-core/spike/middleware"
)

type notFound struct {
	message string
}

func (c notFound) Error() string {
	return c.message
}

func NoDataFound(msg string) error {
	return &notFound{msg}
}

func BadRequest(c *gin.Context, err error) {
	code := int32(400)
	message := "Bad Request: " + err.Error()
	c.Set(middleware.MetaKey, model.Meta{Code: &code, Message: &message})
}

func NotFound(c *gin.Context, err error) {
	code := int32(404)
	message := "Not Found: " + err.Error()
	c.Set(middleware.MetaKey, model.Meta{Code: &code, Message: &message})
}

func ServerError(c *gin.Context, err error) {
	if _, ok := err.(notFound); ok {
		NotFound(c, err)
		return
	}

	code := int32(500)
	message := "Internal server error: " + err.Error()
	c.Set(middleware.MetaKey, model.Meta{Code: &code, Message: &message})
}
