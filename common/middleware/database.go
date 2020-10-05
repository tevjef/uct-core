package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/tevjef/uct-backend/common/database"
)

func Select(ctx context.Context, query string, dest interface{}, args interface{}) error {
	if err := database.FromContext(ctx).Select(query, dest, args); err != nil {
		return err
	}
	return nil
}

func Get(ctx context.Context, query string, dest interface{}, args interface{}) error {
	if err := database.FromContext(ctx).Get(query, dest, args); err != nil {
		return err
	}
	return nil
}

func Insert(ctx context.Context, query string, data interface{}) error {
	database.FromContext(ctx).Insert(query, data)
	return nil
}

// Each handler must either set a meta or response
func Database(db database.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		database.ToContext(c, db)
		c.Next()
	}
}
