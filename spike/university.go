package main

import (
	"database/sql"
	"sort"
	"strings"
	"sync"
	"time"

	"context"

	"github.com/gin-gonic/gin"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/tevjef/uct-backend/common/model"
	"github.com/tevjef/uct-backend/spike/middleware"
	"github.com/tevjef/uct-backend/spike/middleware/cache"
	"github.com/tevjef/uct-backend/spike/middleware/httperror"
	mtrace "github.com/tevjef/uct-backend/spike/middleware/trace"
	"github.com/tevjef/uct-backend/spike/store"
)

func universityHandler(expire time.Duration) gin.HandlerFunc {
	return cache.CachePage(func(c *gin.Context) {
		topicName := strings.ToLower(c.Param("topic"))

		if u, err := SelectUniversity(c, topicName); err != nil {
			if err == sql.ErrNoRows {
				httperror.NotFound(c, err)
				return
			}
			httperror.ServerError(c, err)
			return
		} else {
			response := model.Response{
				Data: &model.Data{University: &u},
			}
			c.Set(middleware.ResponseKey, response)
		}
	}, expire)
}

func universitiesHandler(expire time.Duration) gin.HandlerFunc {
	return cache.CachePage(func(c *gin.Context) {
		if universities, err := SelectUniversities(c); err != nil {
			if err == sql.ErrNoRows {
				httperror.NotFound(c, err)
				return
			}
			httperror.ServerError(c, err)
			return
		} else {
			response := model.Response{
				Data: &model.Data{Universities: universities},
			}
			c.Set(middleware.ResponseKey, response)
		}
	}, expire)
}

func SelectUniversity(ctx context.Context, topicName string) (university model.University, err error) {
	defer model.TimeTrack(time.Now(), "SelectUniversity")
	span := mtrace.NewSpan(ctx, "database.SelectUniversity")
	span.SetLabel("topicName", topicName)
	defer span.Finish()

	m := map[string]interface{}{"topic_name": topicName}
	d := store.Data{}
	if err = store.Get(ctx, store.SelectUniversityCTE, &d, m); err != nil {
		return
	}

	if err = ffjson.Unmarshal([]byte(d.Data), &university); err != nil {
		return
	}
	return
}

func SelectUniversities(ctx context.Context) (universities []*model.University, err error) {
	span := mtrace.NewSpan(ctx, "database.SelectUniversities")
	defer span.Finish()

	var topics []string
	m := map[string]interface{}{}
	if err = store.Select(ctx, store.ListUniversitiesQuery, &topics, m); err != nil {
		return
	}

	if err == nil && len(topics) == 0 {
		err = sql.ErrNoRows
	}

	uniChan := make(chan model.University)
	go func() {
		var wg sync.WaitGroup
		for i := range topics {
			wg.Add(1)
			u := topics[i]
			go func() {
				defer wg.Done()
				var uni model.University
				uni, err = SelectUniversity(ctx, u)
				if err != nil {
					return
				}
				uniChan <- uni
			}()
		}
		wg.Wait()
		close(uniChan)
	}()

	for uni := range uniChan {
		u := uni
		universities = append(universities, &u)
	}

	sort.Sort(model.UniversityByName(universities))
	return
}
