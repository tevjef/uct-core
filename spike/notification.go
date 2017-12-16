package main

import (
	"database/sql"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/tevjef/uct-core/common/model"
	"github.com/tevjef/uct-core/spike/middleware"
	"github.com/tevjef/uct-core/spike/middleware/httperror"
	mtrace "github.com/tevjef/uct-core/spike/middleware/trace"
	"github.com/tevjef/uct-core/spike/store"
	"golang.org/x/net/context"
)

func notificationHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		receiveAtStr := c.FormValue("receiveAt")
		if receiveAtStr == "" {
			httperror.BadRequest(c, errors.New("empty receiveAt"))
			return
		}

		fcmToken := c.FormValue("fcmToken")
		if fcmToken == "" {
			httperror.BadRequest(c, errors.New("empty fcmToken"))
			return
		}

		receiveAt, err := time.Parse(time.RFC3339, receiveAtStr)
		if err != nil {
			httperror.BadRequest(c, err)
			return
		}

		topicName := c.FormValue("topicName")
		if topicName == "" {
			httperror.BadRequest(c, errors.New("empty topicName"))
			return
		}

		notificationId := c.FormValue("notificationId")
		if notificationId == "" {
			httperror.BadRequest(c, errors.New("empty notificationId"))
			return
		}

		os, osVersion, appVersion := deviceInfo(c)
		if err := InsertNotification(c, topicName, fcmToken, receiveAt, os, osVersion, appVersion); err != nil {
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
	os string,
	osVersion string,
	appVersion string) (err error) {

	defer model.TimeTrack(time.Now(), "SelectSection")
	span := mtrace.NewSpan(ctx, "database.InsertNotification")
	span.SetLabel("topicName", topicName)
	defer span.Finish()

	m := map[string]interface{}{
		"topic_name":  topicName,
		"receive_at":  receiveAt,
		"fcm_token":   fcmToken,
		"os":          os,
		"os_version":  osVersion,
		"app_version": appVersion,
	}

	if err = store.Insert(ctx, store.InsertNotificationQuery, m); err != nil {
		return
	}

	return
}
