package spike

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	uctfirestore "github.com/tevjef/uct-backend/common/firestore"
	"github.com/tevjef/uct-backend/common/middleware"
	"github.com/tevjef/uct-backend/common/middleware/httperror"
	"github.com/tevjef/uct-backend/common/model"
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

		firestore := uctfirestore.FromContext(c.Request.Context())

		if err := firestore.InsertDeviceNotification(c, &uctfirestore.DeviceNotification{
			SectionTopicName: topicName,
			FcmToken:         fcmToken,
			ReceivedAt:       receiveAt,
			NotificationId:   notificationId,
			Os:               os,
			OsVersion:        osVersion,
			AppVersion:       appVersion}); err != nil {
			httperror.ServerError(c, err)
		} else {
			response := model.Response{
				Data: &model.Data{},
			}
			c.Set(middleware.ResponseKey, response)
		}
	}
}
