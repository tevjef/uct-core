package ein

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/tevjef/uct-backend/common/model"
)

type UniversityView struct {
	Data []byte `firestore:"data"`
}

func (ein *ein) slimInsertUniversity(university model.University) {

	defer model.TimeTrack(time.Now(), "insertUniversity")

	data, err := university.Marshal()

	university.Subjects = nil
	universityView := UniversityView{data}
	//collections := ein.firestoreClient.Collection("universities")
	//docRef := collections.Doc(university.TopicName)
	//result, err := docRef.Set(ein.ctx, universityView)
	if err != nil {
		log.Fatalln(err)
	}

	log.Infoln(universityView)

}

func (ein *ein) insertUniversity(university model.University) {

	defer model.TimeTrack(time.Now(), "insertUniversity")

	//university.Id = ein.postgres.Upsert(UniversityInsertQuery, UniversityUpdateQuery, university)

	/*	ein.insertSubjects(&university)

		// ResolvedSemesters
		ein.insertSemester(&university)

		// Registrations
		for _, registrations := range university.Registrations {
			registrations.UniversityId = university.Id
			ein.insertRegistration(registrations)
		}

		// university []Metadata
		metadata := university.Metadata
		for metadataIndex := range metadata {
			metadata := metadata[metadataIndex]

			metadata.UniversityId = &university.Id
			ein.insertMetadata(metadata)
		}*/
}
