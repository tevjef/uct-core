package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/tevjef/uct-core/common/database"
)

// Each handler must either set a meta or response
func Database(db database.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		database.ToContext(c, db)
		c.Next()
	}
}
