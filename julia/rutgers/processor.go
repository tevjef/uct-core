package rutgers

import (
	"strings"
	"time"

	"github.com/tevjef/uct-core/common/model"
	"github.com/tevjef/uct-core/julia/rutgers/topic"

	log "github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
)

type RutgersProcessor struct {
	in         chan model.UCTNotification
	out        chan model.UCTNotification
	expiration time.Duration
	routines   *Routines
}

func New(expiration time.Duration) *RutgersProcessor {
	rp := &RutgersProcessor{
		in:         make(chan model.UCTNotification),
		out:        make(chan model.UCTNotification),
		expiration: expiration,
		routines:   &Routines{routineMap: make(map[string]*topic.Routine)},
	}

	go rp.process(context.TODO())

	return rp
}

func (rp *RutgersProcessor) IsMatch(topic string) bool {
	return strings.HasPrefix(topic, "rutgers")
}

func (rp *RutgersProcessor) In(notification model.UCTNotification) {
	//log.Printf("In(model.UCTNotification) %+v", notification)
	rp.in <- notification
}

func (rp *RutgersProcessor) Done() <-chan model.UCTNotification {
	//log.Printf("Done() chan model.UCTNotification %+v", rp.out)
	return rp.out
}

func (rp *RutgersProcessor) process(ctx context.Context) {
	for uctNotification := range rp.in {
		uctNotification := uctNotification
		if routine := rp.routines.Get(uctNotification.TopicName); routine == nil {
			// create topic and create expiration routine
			// Manages the communication of the new topic routine and the processor
			// When it starts and completes, as well as any messages sent out of the routine
			go func() {
				defer func(start time.Time) {
					log.WithFields(log.Fields{
						"routines_count": rp.routines.Size(),
						"routine_elapsed": time.Since(start).Seconds(),
						"routine_topic":   uctNotification.TopicName,
						"university_name": uctNotification.University.TopicName,
					}).Infoln("routine_done")
				}(time.Now())

				routine = topic.NewTopicRoutine(rp.expiration)
				rp.routines.Set(uctNotification.TopicName, routine)
				routine.Send(uctNotification)

				for {
					select {
					case uctNotification := <-routine.Out():
						rp.out <- uctNotification
					case topic := <-routine.Done():
						rp.routines.Remove(topic)
						return
					}
				}
			}()
		} else {
			routine.Send(uctNotification)
		}
	}
}
