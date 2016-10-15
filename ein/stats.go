package main

import (
	"time"
	log "github.com/Sirupsen/logrus"
)

func audit(university string) {
	var err error

	start := time.Now()

	if err != nil {
		log.Fatalf("Error while creating the hook: %v", err)
	}

	var insertions int
	var updates int
	var upserts int
	var existential int
	var subjectCount int
	var courseCount int
	var sectionCount int
	var meetingCount int
	var metadataCount int
	var serialCourse int
	var serialSection int
	var serialSubject int

	Outerloop:
	for {
		select {
		case count := <-insertionsCh:
			insertions += count
		case count := <-updatesCh:
			updates += count
		case count := <-upsertsCh:
			upserts += count
		case count := <-existentialCh:
			existential += count
		case count := <-subjectCountCh:
			subjectCount += count
		case count := <-courseCountCh:
			courseCount += count
		case count := <-sectionCountCh:
			sectionCount += count
		case count := <-meetingCountCh:
			meetingCount += count
		case count := <-metadataCountCh:
			metadataCount += count
		case count := <-serialSubjectCh:
			serialSubject += count
		case count := <-serialCourseCh:
			serialCourse += count
		case count := <-serialSectionCh:
			serialSection += count
		case <-doneAudit:

			log.WithFields(log.Fields{
				"university_name": university,
				"insertions":    insertions,
				"updates":       updates,
				"upserts":       upserts,
				"existential":   existential,
				"subjectCount":  subjectCount,
				"courseCount":   courseCount,
				"sectionCount":  sectionCount,
				"meetingCount":  meetingCount,
				"metadataCount": metadataCount,
				"serialSubject": serialSubject,
				"serialCourse":  serialCourse,
				"serialSection": serialSection,
				"elapsed":       time.Since(start).Seconds(),
			}).Info("done!")

			doneAudit <- true
			break Outerloop // Break out of loop to end goroutine
		}
	}
}

var (
	insertionsCh    = make(chan int)
	updatesCh       = make(chan int)
	upsertsCh       = make(chan int)
	existentialCh   = make(chan int)
	subjectCountCh  = make(chan int)
	courseCountCh   = make(chan int)
	sectionCountCh  = make(chan int)
	meetingCountCh  = make(chan int)
	metadataCountCh = make(chan int)

	serialCourseCh  = make(chan int)
	serialSectionCh = make(chan int)
	serialSubjectCh = make(chan int)

	doneAudit = make(chan bool)
)
