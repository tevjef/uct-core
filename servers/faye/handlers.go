package main

import (
	"errors"
	"github.com/gin-gonic/gin"
	"uct/common/model"
	"uct/servers"
)

func notificationHandler(c *gin.Context) {
	topicName := c.PostFormValue("topic")
	messageId := c.PostFormValue("message_id")

	if topicName != "" {
		servers.ResolveErr(servers.ErrMissingParam{"Missing topic parameter in POST request"}, c)
		return
	} else if messageId != "" {
		servers.ResolveErr(servers.ErrMissingParam{"Missing messageId parameter in POST request"}, c)
		return
	}

	if id := acknowledgeNotification(topicName, messageId); id == 0 {
		servers.ResolveErr(errors.New("Failed to acknowlege notification on server"), c)
	} else {
		response := model.Response{}
		c.Set(servers.ResponseKey, response)
	}
}
