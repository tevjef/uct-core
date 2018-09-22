package main

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/tevjef/uct-backend/common/model"
	"github.com/tevjef/uct-backend/julia/rutgers"
)

type Processor interface {
	IsMatch(topic string) bool
	In(model.UCTNotification)
	Done() <-chan model.UCTNotification
}

type Process struct {
	in  chan model.UCTNotification
	out chan model.UCTNotification
}

func (p *Process) Run(fn DispatchFunc) {
	var rutgersProcessor = rutgers.New(4 * time.Minute)

	for {
		select {
		case uctNotification := <-rutgersProcessor.Done():
			go func() { p.out <- uctNotification }()
		case uctNotification := <-p.in:
			log.WithFields(log.Fields{
				"topic":           uctNotification.TopicName,
				"university_name": uctNotification.University.TopicName}).Infoln("processor_in")
			if rutgersProcessor.IsMatch(uctNotification.TopicName) {
				rutgersProcessor.In(uctNotification)
			} else {
				go func() { p.out <- uctNotification }()
			}
		case uctNotification := <-p.out:
			log.WithFields(log.Fields{
				"topic":           uctNotification.TopicName,
				"university_name": uctNotification.University.TopicName}).Infoln("processor_out")
			fn(uctNotification)
		}
	}
}

func (p *Process) Recv(uctNotification *model.UCTNotification) {
	p.in <- *uctNotification
}

type DispatchFunc func(uctNotification model.UCTNotification)
