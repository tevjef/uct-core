package main

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tevjef/uct-backend/common/middleware"
	"github.com/tevjef/uct-backend/common/middleware/cache"
	"github.com/tevjef/uct-backend/common/middleware/httperror"
	mtrace "github.com/tevjef/uct-backend/common/middleware/trace"
	"github.com/tevjef/uct-backend/common/model"
	"github.com/tevjef/uct-backend/spike/store"
)

func sectionHandler(expire time.Duration) gin.HandlerFunc {
	return cache.CachePage(func(c *gin.Context) {
		sectionTopicName := strings.ToLower(c.Param("topic"))

		if s, _, err := SelectSection(c, sectionTopicName); err != nil {
			if err == sql.ErrNoRows {
				httperror.NotFound(c, err)
				return
			}
			httperror.ServerError(c, err)
			return
		} else {
			response := model.Response{
				Data: &model.Data{Section: &s},
			}
			c.Set(middleware.ResponseKey, response)
		}
	}, expire)
}

func SelectSection(ctx context.Context, sectionTopicName string) (section model.Section, b []byte, err error) {
	defer model.TimeTrack(time.Now(), "SelectSection")
	span := mtrace.NewSpan(ctx, "database.SelectSection")
	span.SetLabel("topicName", sectionTopicName)
	defer span.Finish()

	d := store.Data{}
	m := map[string]interface{}{"topic_name": sectionTopicName}
	if err = middleware.Get(ctx, store.SelectProtoSectionQuery, &d, m); err != nil {
		return
	}
	b = d.Data
	err = section.Unmarshal(b)
	return
}
