package ein

import (
	"fmt"

	uctfirestore "github.com/tevjef/uct-backend/common/firestore"
	"github.com/tevjef/uct-backend/common/model"
)

func (ein *ein) insertUniversity(newUniversity model.University, university model.University) {
	_ = ein.uctFSClient.InsertUniversity(ein.ctx, newUniversity)

	ein.insertSubjects(&newUniversity)

	ein.insertSemester(&newUniversity)
}

func (ein *ein) insertSemester(university *model.University) {
	firestoreSeason := uctfirestore.FirestoreSemesters{
		CurrentSeason: uctfirestore.MakeSemesterKey(university.ResolvedSemesters.Current),
		NextSeason:    uctfirestore.MakeSemesterKey(university.ResolvedSemesters.Next),
		LastSeason:    uctfirestore.MakeSemesterKey(university.ResolvedSemesters.Last),
	}

	collections := ein.firestoreClient.Collection(uctfirestore.CollectionUniversitySemesters)
	docRef := collections.Doc(university.TopicName)
	_, err := docRef.Set(ein.ctx, firestoreSeason)
	if err != nil {
		ein.logger.WithError(err).Fatalln("firestore: failed to set university.semesters")
	}
}

func (ein *ein) insertSubjects(university *model.University) {
	_ = ein.uctFSClient.InsertSubjectsBySemester(ein.ctx, *university, university.ResolvedSemesters.Current)
	_ = ein.uctFSClient.InsertSubjectsBySemester(ein.ctx, *university, university.ResolvedSemesters.Next)
	_ = ein.uctFSClient.InsertSubjectsBySemester(ein.ctx, *university, university.ResolvedSemesters.Last)

	_ = ein.uctFSClient.InsertSubjects(ein.ctx, university.Subjects)
}

func (ein *ein) updateSerialSection(sectionMeta []uctfirestore.SectionMeta) {
	_ = ein.uctFSClient.InsertSections(ein.ctx, sectionMeta)

	field := map[string]interface{}{}
	logSectionMetadata(sectionMeta, field)
	ein.logger.WithFields(field).Infof("firestore: %d sections updated", len(sectionMeta))
}

func logSectionMetadata(sectionMeta []uctfirestore.SectionMeta, field map[string]interface{}) {
	if len(sectionMeta) <= 50 {
		var sectionStatus []string
		for i := range sectionMeta {
			sm := sectionMeta[i]
			sectionStatus = append(sectionStatus, fmt.Sprintf("status: %s topicName: %s", sm.Section.Status, sm.Section.TopicName))
		}
		field["sections"] = sectionStatus
	}
}

func (ein *ein) updateSerialCourse(courses []*model.Course) {
	_ = ein.uctFSClient.InsertCourses(ein.ctx, courses)

	ein.logger.Infof("firestore: %d courses updated", len(courses))
}
