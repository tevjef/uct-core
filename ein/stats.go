package ein

import (
	"github.com/tevjef/uct-backend/common/model"
)

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
