package main

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/tevjef/uct-backend/common/middleware"
	"github.com/tevjef/uct-backend/common/middleware/httperror"
	mtrace "github.com/tevjef/uct-backend/common/middleware/trace"
	"github.com/tevjef/uct-backend/common/model"
	"github.com/tevjef/uct-backend/spike/store"
)

func notificationHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		receiveAtStr, exists := c.GetPostForm("receiveAt")
		if !exists {
			httperror.BadRequest(c, errors.New("empty receiveAt"))
			return
		}

		fcmToken, exists := c.GetPostForm("fcmToken")
		if !exists {
			httperror.BadRequest(c, errors.New("empty fcmToken"))
			return
		}

		receiveAt, err := time.Parse(time.RFC3339, receiveAtStr)
		if err != nil {
			httperror.BadRequest(c, err)
			return
		}

		topicName, exists := c.GetPostForm("topicName")
		if !exists {
			httperror.BadRequest(c, errors.New("empty topicName"))
			return
		}

		notificationId, exists := c.GetPostForm("notificationId")
		if notificationId == "" {
			httperror.BadRequest(c, errors.New("empty notificationId"))
			return
		} else if _, err := strconv.Atoi(notificationId); err != nil {
			httperror.BadRequest(c, errors.New("invalid notificationId: "+notificationId))
			return
		}

		os, osVersion, appVersion := deviceInfo(c.Request.Header)
		if err := InsertNotification(c, topicName, fcmToken, receiveAt, notificationId, os, osVersion, appVersion); err != nil {
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

func InsertNotification(
	ctx context.Context,
	topicName,
	fcmToken string,
	receiveAt time.Time,
	notificationId string,
	os string,
	osVersion string,
	appVersion string) (err error) {

	defer model.TimeTrack(time.Now(), "SelectSection")
	span := mtrace.NewSpan(ctx, "database.InsertNotification")
	span.SetLabel("topicName", topicName)
	defer span.Finish()

	m := map[string]interface{}{
		"topic_name":      topicName,
		"receive_at":      receiveAt,
		"fcm_token":       fcmToken,
		"notification_id": notificationId,
		"os":              os,
		"os_version":      osVersion,
		"app_version":     appVersion,
	}

	if err = middleware.Insert(ctx, store.InsertNotificationQuery, m); err != nil {
		return
	}

	return
}
