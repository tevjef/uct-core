package spike

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	uctfirestore "github.com/tevjef/uct-backend/common/firestore"
	"github.com/tevjef/uct-backend/common/middleware"
	"github.com/tevjef/uct-backend/common/middleware/httperror"
	"github.com/tevjef/uct-backend/common/model"
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

		firestore := uctfirestore.FromContext(c)

		if err := firestore.InsertSubscriptionAndUpdateCount(c, &uctfirestore.Subscription{
			SectionTopicName: topicName,
			FcmToken:         fcmToken,
			IsSubscribed:     subscribed,
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
