package store

type Data struct {
	Data []byte `db:"data"`
}

var Queries = []string{
	SelectCourseQuery,
	SelectSubscription,
	CurrentSubscribers,
}

const (
	SelectCourseQuery = `SELECT data FROM course WHERE course.topic_name = :topic_name ORDER BY course.id`

	SelectSubscription = `SELECT id, os, is_subscribed, topic_name, fcm_token, created_at 
							FROM subscription 
							WHERE subscription.topic_name = :topic_name 
							ORDER BY subscription.created_at`
	CurrentSubscribers = `SELECT
    (SELECT count(is_subscribed) as the_count
    FROM subscription
    WHERE is_subscribed = 'true'
      AND topic_name = :topic_name 
    GROUP BY is_subscribed, topic_name
) -
    (SELECT count(is_subscribed) as the_count
     FROM subscription
     WHERE is_subscribed = 'false'
       AND topic_name = :topic_name 
     GROUP BY is_subscribed, topic_name
    ) as difference`
)
