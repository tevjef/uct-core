package main

import (
	"context"
	"time"

	"github.com/tevjef/uct-backend/common/middleware"
	mtrace "github.com/tevjef/uct-backend/common/middleware/trace"
	"github.com/tevjef/uct-backend/common/model"
	"github.com/tevjef/uct-backend/edward/store"
	"go.opencensus.io/trace"
)

func GetSubscriberCount(ctx context.Context, sectionTopicName string) (count int, err error) {
	defer model.TimeTrack(time.Now(), "GetSubscriberCount")
	span := mtrace.NewSpan(ctx, "database.GetSubscriberCount")
	span.AddAttributes(trace.StringAttribute("topicName", sectionTopicName))
	defer span.End()

	m := map[string]interface{}{"topic_name": sectionTopicName}
	if err = middleware.Get(ctx, store.CurrentSubscribers, &count, m); err != nil {
		return
	}

	return
}
