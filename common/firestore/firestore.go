package firestore

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
)

const key = "firestoreclient"

// Setter defines a context that enables setting values.
type Setter interface {
	Set(string, interface{})
}

func FromContext(ctx context.Context) *firestore.Client {
	return ctx.Value(key).(*firestore.Client)
}

func ToContext(s Setter, client *firestore.Client) {
	s.Set(key, client)
}

func Firestore(client *firestore.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ToContext(c, client)
		c.Next()
	}
}
