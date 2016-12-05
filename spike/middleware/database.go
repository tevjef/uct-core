package middleware

import (
	"github.com/tevjef/uct-core/common/database"

	"github.com/gin-gonic/gin"
)

// Each handler must either set a meta or response
func Database(db database.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		database.ToContext(c, db)
		c.Next()
	}
}
