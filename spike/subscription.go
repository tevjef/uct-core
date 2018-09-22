package main

import (
	"database/sql"
	"strconv"
	"time"

	"context"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/tevjef/uct-backend/common/model"
	"github.com/tevjef/uct-backend/spike/middleware"
	"github.com/tevjef/uct-backend/spike/middleware/httperror"
	mtrace "github.com/tevjef/uct-backend/spike/middleware/trace"
	"github.com/tevjef/uct-backend/spike/store"
)

func subscriptionHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		subscribed, err := strconv.ParseBool(c.PostForm("isSubscribed"))
		if err != nil {
			httperror.BadRequest(c, errors.New("invalid isSubscribed"+err.Error()))
			return
		}

		fcmToken, exists := c.GetPostForm("fcmToken")
		if !exists {
			httperror.BadRequest(c, errors.New("empty fcmToken"))
			return
		}

		topicName, exists := c.GetPostForm("topicName")
		if !exists {
			httperror.BadRequest(c, errors.New("empty topicName"))
			return
		}

		os, osVersion, appVersion := deviceInfo(c.Request.Header)
		if err := InsertSubscription(c, topicName, fcmToken, subscribed, os, osVersion, appVersion); err != nil {
			if err == sql.ErrNoRows {
				httperror.NotFound(c, err)
				return
			}
			httperror.ServerError(c, err)
		} else {
			response := model.Response{
				Data: &model.Data{},
			}
			c.Set(middleware.ResponseKey, response)
		}
	}
}

func InsertSubscription(
	ctx context.Context,
	topicName,
	fcmToken string,
	subscribed bool,
	os string,
	osVersion string,
	appVersion string) (err error) {

	defer model.TimeTrack(time.Now(), "SelectSection")
	span := mtrace.NewSpan(ctx, "database.InsertSubscription")
	span.SetLabel("topicName", topicName)
	defer span.Finish()

	m := map[string]interface{}{
		"topic_name":    topicName,
		"fcm_token":     fcmToken,
		"is_subscribed": subscribed,
		"os":            os,
		"os_version":    osVersion,
		"app_version":   appVersion,
	}

	if err = store.Insert(ctx, store.InsertSubscriptionQuery, m); err != nil {
		return
	}
	return
}
