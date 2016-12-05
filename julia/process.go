package main

import (
	"time"

	"github.com/tevjef/uct-core/common/model"
	"github.com/tevjef/uct-core/julia/rutgers"

	log "github.com/Sirupsen/logrus"
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
			log.WithFields(log.Fields{"topic": uctNotification.TopicName}).Infoln("processor_out")
			go func() { p.out <- uctNotification }()
		case uctNotification := <-p.in:
			log.WithFields(log.Fields{"topic": uctNotification.TopicName}).Infoln("processor_in")
			if rutgersProcessor.IsMatch(uctNotification.TopicName) {
				rutgersProcessor.In(uctNotification)
			}
		case uctNotification := <-p.out:
			fn(uctNotification)
		}
	}
}

func (p *Process) Recv(uctNotification *model.UCTNotification) {
	p.in <- *uctNotification
}

type DispatchFunc func(uctNotification model.UCTNotification)
