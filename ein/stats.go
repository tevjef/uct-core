package main

import (
	log "github.com/Sirupsen/logrus"
	"time"
	"uct/common/model"
)

func audit(university string) {
	start := time.Now()

	var insertions int
	var updates int
	var upserts int
	var existential int

	var subject int
	var course int
	var section int
	var meeting int
	var metadata int

	var diffSubject int
	var diffCourse int
	var diffSection int
	var diffMeeting int
	var diffMetadata int

	var diffSerialCourse int
	var diffSerialSection int
	var diffSerialSubject int
	var diffSerialMeeting int
	var diffSerialMetadata int

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

		case count := <-subjectCh:
			subject += count
		case count := <-courseCh:
			course += count
		case count := <-sectionCh:
			section += count
		case count := <-meetingCh:
			meeting += count
		case count := <-metadataCh:
			metadata += count

		case count := <-diffSubjectCh:
			diffSubject += count
		case count := <-diffCourseCh:
			diffCourse += count
		case count := <-diffSectionCh:
			diffSection += count
		case count := <-diffMeetingCh:
			diffMeeting += count
		case count := <-diffMetadataCh:
			diffMetadata += count

		case count := <-diffSerialSubjectCh:
			diffSerialSubject += count
		case count := <-diffSerialCourseCh:
			diffSerialCourse += count
		case count := <-diffSerialSectionCh:
			diffSerialSection += count
		case count := <-diffSerialMeetingCountCh:
			diffSerialMeeting += count
		case count := <-diffSerialMetadataCountCh:
			diffSerialMetadata += count
		case <-doneAudit:

			log.WithFields(log.Fields{
				"university_name": university,
				"insertions":      insertions,
				"updates":         updates,
				"upserts":         upserts,
				"existential":     existential,

				"subjectCount":  subject,
				"courseCount":   course,
				"sectionCount":  section,
				"meetingCount":  meeting,
				"metadataCount": metadata,

				"diffSubjectCount":  diffSubject,
				"diffCourseCount":   diffCourse,
				"diffSectionCount":  diffSection,
				"diffMeetingCount":  diffMeeting,
				"diffMetadataCount": diffMetadata,

				"diffSerialSubject":  diffSerialSubject,
				"diffSerialCourse":   diffSerialCourse,
				"diffSerialSection":  diffSerialSection,
				"diffSerialMeeting":  diffSerialMeeting,
				"diffSerialMetadata": diffSerialMetadata,
				"elapsed":            time.Since(start).Seconds(),
			}).Info("done!")

			doneAudit <- true
			return // Break out of loop to end goroutine
		}
	}
}

var (
	insertionsCh  = make(chan int)
	updatesCh     = make(chan int)
	upsertsCh     = make(chan int)
	existentialCh = make(chan int)

	subjectCh  = make(chan int)
	courseCh   = make(chan int)
	sectionCh  = make(chan int)
	meetingCh  = make(chan int)
	metadataCh = make(chan int)

	diffSubjectCh  = make(chan int)
	diffCourseCh   = make(chan int)
	diffSectionCh  = make(chan int)
	diffMeetingCh  = make(chan int)
	diffMetadataCh = make(chan int)

	diffSerialCourseCh        = make(chan int)
	diffSerialSectionCh       = make(chan int)
	diffSerialSubjectCh       = make(chan int)
	diffSerialMeetingCountCh  = make(chan int)
	diffSerialMetadataCountCh = make(chan int)

	doneAudit = make(chan bool)
)

func countUniversity(uni model.University, subjectCh, courseCh, sectionCh, meetingCh, metadataCh chan int) {
	subjectCh <- len(uni.Subjects)
	metadataCh <- len(uni.Metadata)
	for i := range uni.Subjects {
		courseCh <- len(uni.Subjects[i].Courses)
		metadataCh <- len(uni.Subjects[i].Metadata)
		for j := range uni.Subjects[i].Courses {
			sectionCh <- len(uni.Subjects[i].Courses[j].Sections)
			metadataCh <- len(uni.Subjects[i].Courses[j].Metadata)
			for k := range uni.Subjects[i].Courses[j].Sections {
				metadataCh <- len(uni.Subjects[i].Courses[j].Sections[k].Metadata)
				meetingCh <- len(uni.Subjects[i].Courses[j].Sections[k].Meetings)
			}
		}
	}
}

func countSubjects(subjects []*model.Subject, courses []*model.Course, subjectCh, courseCh, sectionCh, meetingCh, metadataCh chan int) {
	subjectCh <- len(subjects)
	courseCh <- len(courses)
	for i := range subjects {
		metadataCh <- len(subjects[i].Metadata)
	}
	for j := range courses {
		sectionCh <- len(courses[j].Sections)
		metadataCh <- len(courses[j].Metadata)
		for k := range courses[j].Sections {
			metadataCh <- len(courses[j].Sections[k].Metadata)
			meetingCh <- len(courses[j].Sections[k].Meetings)
		}
	}
}
