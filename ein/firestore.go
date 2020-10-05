package ein

import (
	uctfirestore "github.com/tevjef/uct-backend/common/firestore"
	"github.com/tevjef/uct-backend/common/model"
)

func (ein *ein) insertUniversity(newUniversity model.University, university model.University) {
	_ = ein.uctFSClient.InsertUniversity(newUniversity)

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
	_ = ein.uctFSClient.InsertSubjectsBySemester(*university, university.ResolvedSemesters.Current)
	_ = ein.uctFSClient.InsertSubjectsBySemester(*university, university.ResolvedSemesters.Next)
	_ = ein.uctFSClient.InsertSubjectsBySemester(*university, university.ResolvedSemesters.Last)

	_ = ein.uctFSClient.InsertSubjects(university.Subjects)
}

func (ein *ein) updateSerialSection(sectionMeta []uctfirestore.SectionMeta) {
	_ = ein.uctFSClient.InsertSection(sectionMeta)
}

func (ein *ein) updateSerialCourse(courses []*model.Course) {
	_ = ein.uctFSClient.InsertCourses(courses)
}
