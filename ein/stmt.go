package main

import (
	"github.com/jmoiron/sqlx"
	log "github.com/Sirupsen/logrus"
	uct "uct/common"

)

var preparedStmts = make(map[string]*sqlx.NamedStmt)


func GetCachedStmt(key string) *sqlx.NamedStmt {
	return preparedStmts[key]
}

func (dbHandler DatabaseHandlerImpl) prepare(query string) *sqlx.NamedStmt {
	if named, err := dbHandler.Database.PrepareNamed(query); err != nil {
		log.Debugln(query)
		uct.CheckError(err)
		return nil
	} else {
		return named
	}
}

func (dbHandler DatabaseHandlerImpl) PrepareAllStmts() {
	queries := []string{UniversityInsertQuery,
		UniversityUpdateQuery,
		SemesterInsertQuery,
		SemesterUpdateQuery,
		SubjectExistQuery,
		SubjectInsertQuery,
		SubjectUpdateQuery,
		CourseUpdateQuery,
		CourseExistQuery,
		CourseInsertQuery,
		SectionInsertQuery,
		SectionUpdateQuery,
		MeetingUpdateQuery,
		MeetingInsertQuery,
		MeetingExistQuery,
		InstructorExistQuery,
		InstructorUpdateQuery,
		InstructorInsertQuery,
		BookUpdateQuery,
		BookInsertQuery,
		RegistrationUpdateQuery,
		RegistrationInsertQuery,
		MetaUniExistQuery,
		MetaUniUpdateQuery,
		MetaUniInsertQuery,
		MetaSubjectExistQuery,
		MetaSubjectUpdateQuery,
		MetaSubjectInsertQuery,
		MetaCourseExistQuery,
		MetaCourseUpdateQuery,
		MetaCourseInsertQuery,
		MetaSectionExistQuery,
		MetaSectionInsertQuery,
		MetaSectionUpdateQuery,
		MetaSectionExistQuery,
		MetaMeetingInsertQuery,
		MetaMeetingUpdateQuery,
		SerialSubjectUpdateQuery,
		SerialCourseUpdateQuery,
		SerialSectionUpdateQuery}

	for _, query := range queries {
		preparedStmts[query] = dbHandler.prepare(query)
	}
}

