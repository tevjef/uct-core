package spike

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	uctfirestore "github.com/tevjef/uct-backend/common/firestore"
	"github.com/tevjef/uct-backend/common/middleware"
	"github.com/tevjef/uct-backend/common/middleware/cache"
	"github.com/tevjef/uct-backend/common/middleware/httperror"
	"github.com/tevjef/uct-backend/common/model"
)

func sectionHandler(expire time.Duration) gin.HandlerFunc {
	return cache.CachePage(func(c *gin.Context) {
		sectionTopicName := strings.ToLower(c.Param("topic"))
		firestore := uctfirestore.FromContext(c)

		if s, err := firestore.GetSection(c, sectionTopicName); err != nil {
			httperror.ServerError(c, err)
			return
		} else {
			response := model.Response{
				Data: &model.Data{Section: s},
			}
			c.Set(middleware.ResponseKey, response)
		}
	}, expire)
}
