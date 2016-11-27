package middleware

import (
	"github.com/gin-gonic/gin"
	"uct/common/model"
)

type ErrNoRows struct {
	Uri string
}

func (e ErrNoRows) Error() string {
	return e.Uri
}

type ErrMissingParam struct {
	String string
}

func (e ErrMissingParam) Error() string {
	return e.String
}

func ResolveErr(err error, c *gin.Context) {
	if v, ok := err.(ErrNoRows); ok {
		c.Set(MetaKey, resolveResNotFound(v.Error()+" URI: "+c.Request.RequestURI))
	} else if v, ok := err.(ErrMissingParam); ok {
		c.Set(MetaKey, resolveMissingParam(v.Error()))
	} else {
		code := int32(500)
		message := "Internal server error: " + err.Error() + " "
		c.Set(MetaKey, model.Meta{Code: &code, Message: &message})
	}
}

func resolveMissingParam(str string) model.Meta {
	code := int32(400)
	message := "Bad Request: " + str
	return model.Meta{Code: &code, Message: &message}
}

func resolveResNotFound(str string) model.Meta {
	code := int32(404)
	message := "Not Found: " + str
	return model.Meta{Code: &code, Message: &message}
}
