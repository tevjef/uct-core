package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/tevjef/uct-backend/common/database"
	mtrace "github.com/tevjef/uct-backend/common/middleware/trace"
)

func Select(ctx context.Context, query string, dest interface{}, args interface{}) error {
	span := mtrace.NewSpan(ctx, "database.Select")
	span.SetLabel("query", query)
	defer span.Finish()

	if err := database.FromContext(ctx).Select(query, dest, args); err != nil {
		return err
	}
	return nil
}

func Get(ctx context.Context, query string, dest interface{}, args interface{}) error {
	span := mtrace.NewSpan(ctx, "database.Get")
	span.SetLabel("query", query)
	defer span.Finish()

	if err := database.FromContext(ctx).Get(query, dest, args); err != nil {
		return err
	}
	return nil
}

func Insert(ctx context.Context, query string, data interface{}) error {
	span := mtrace.NewSpan(ctx, "database.Insert")
	span.SetLabel("query", query)
	defer span.Finish()

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
