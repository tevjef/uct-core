package main

import (
	"context"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/tevjef/uct-backend/common/middleware"
	mtrace "github.com/tevjef/uct-backend/common/middleware/trace"
	"github.com/tevjef/uct-backend/common/model"
	"github.com/tevjef/uct-backend/edward/store"
)

func GetSubscriberCount(ctx context.Context, sectionTopicName string) (count int, err error) {
	log.Debugln("starting GetSubscriberCount")

	defer model.TimeTrack(time.Now(), "GetSubscriberCount")
	span := mtrace.NewSpan(ctx, "database.GetSubscriberCount")
	span.SetLabel("topicName", sectionTopicName)
	defer span.Finish()

	m := map[string]interface{}{"topic_name": sectionTopicName}
	if err = middleware.Get(ctx, store.CurrentSubscribers, &count, m); err != nil {
		return
	}

	return
}
