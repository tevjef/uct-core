package main

import (
	"bytes"
	"fmt"
	uct "uct/common"
)

func VerboseEqualSubject(this *uct.Subject, that interface{}) error {
	if that == nil {
		if this == nil {
			return nil
		}
		return fmt.Errorf("that == nil && this != nil")
	}

	that1, ok := that.(*uct.Subject)
	if !ok {
		that2, ok := that.(uct.Subject)
		if ok {
			that1 = &that2
		} else {
			return fmt.Errorf("that is not of type *Subject")
		}
	}
	if that1 == nil {
		if this == nil {
			return nil
		}
		return fmt.Errorf("that is type *Subject but is nil && this != nil")
	} else if this == nil {
		return fmt.Errorf("that is type *Subject but is not nil && this == nil")
	}
	if this.Id != that1.Id {
		return fmt.Errorf("Id this(%v) Not Equal that(%v)", this.Id, that1.Id)
	}
	if this.UniversityId != that1.UniversityId {
		return fmt.Errorf("UniversityId this(%v) Not Equal that(%v)", this.UniversityId, that1.UniversityId)
	}
	if this.Name != that1.Name {
		return fmt.Errorf("Name this(%v) Not Equal that(%v)", this.Name, that1.Name)
	}
	if this.Number != that1.Number {
		return fmt.Errorf("Number this(%v) Not Equal that(%v)", this.Number, that1.Number)
	}
	if this.Season != that1.Season {
		return fmt.Errorf("Season this(%v) Not Equal that(%v)", this.Season, that1.Season)
	}
	if this.Year != that1.Year {
		return fmt.Errorf("Year this(%v) Not Equal that(%v)", this.Year, that1.Year)
	}
	if this.Hash != that1.Hash {
		return fmt.Errorf("Hash this(%v) Not Equal that(%v)", this.Hash, that1.Hash)
	}
	if this.TopicName != that1.TopicName {
		return fmt.Errorf("TopicName this(%v) Not Equal that(%v)", this.TopicName, that1.TopicName)
	}
	/*if len(this.Courses) != len(that1.Courses) {
		return fmt.Errorf("Courses this(%v) Not Equal that(%v)", len(this.Courses), len(that1.Courses))
	}*/
	if len(this.Metadata) != len(that1.Metadata) {
		return fmt.Errorf("Metadata this(%v) Not Equal that(%v)", len(this.Metadata), len(that1.Metadata))
	}
	for i := range this.Metadata {
		if !this.Metadata[i].Equal(&that1.Metadata[i]) {
			return fmt.Errorf("Metadata this[%v](%v) Not Equal that[%v](%v)", i, this.Metadata[i], i, that1.Metadata[i])
		}
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return fmt.Errorf("XXX_unrecognized this(%v) Not Equal that(%v)", this.XXX_unrecognized, that1.XXX_unrecognized)
	}
	return nil
}
func EqualSubject(this *uct.Subject, that interface{}) bool {
	if that == nil {
		if this == nil {
			return true
		}
		return false
	}

	that1, ok := that.(*uct.Subject)
	if !ok {
		that2, ok := that.(uct.Subject)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		if this == nil {
			return true
		}
		return false
	} else if this == nil {
		return false
	}
	if this.Id != that1.Id {
		return false
	}
	if this.UniversityId != that1.UniversityId {
		return false
	}
	if this.Name != that1.Name {
		return false
	}
	if this.Number != that1.Number {
		return false
	}
	if this.Season != that1.Season {
		return false
	}
	if this.Year != that1.Year {
		return false
	}
	if this.Hash != that1.Hash {
		return false
	}
	if this.TopicName != that1.TopicName {
		return false
	}
	/*if len(this.Courses) != len(that1.Courses) {
		return false
	}
	for i := range this.Courses {
		if !this.Courses[i].Equal(&that1.Courses[i]) {
			return false
		}
	}*/
	if len(this.Metadata) != len(that1.Metadata) {
		return false
	}
	for i := range this.Metadata {
		if !this.Metadata[i].Equal(&that1.Metadata[i]) {
			return false
		}
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}

func VerboseEqualCourse(this *uct.Course, that interface{}) error {
	if that == nil {
		if this == nil {
			return nil
		}
		return fmt.Errorf("that == nil && this != nil")
	}

	that1, ok := that.(*uct.Course)
	if !ok {
		that2, ok := that.(uct.Course)
		if ok {
			that1 = &that2
		} else {
			return fmt.Errorf("that is not of type *Course")
		}
	}
	if that1 == nil {
		if this == nil {
			return nil
		}
		return fmt.Errorf("that is type *Course but is nil && this != nil")
	} else if this == nil {
		return fmt.Errorf("that is type *Course but is not nil && this == nil")
	}
	if this.Id != that1.Id {
		return fmt.Errorf("Id this(%v) Not Equal that(%v)", this.Id, that1.Id)
	}
	if this.SubjectId != that1.SubjectId {
		return fmt.Errorf("SubjectId this(%v) Not Equal that(%v)", this.SubjectId, that1.SubjectId)
	}
	if this.Name != that1.Name {
		return fmt.Errorf("Name this(%v) Not Equal that(%v)", this.Name, that1.Name)
	}
	if this.Number != that1.Number {
		return fmt.Errorf("Number this(%v) Not Equal that(%v)", this.Number, that1.Number)
	}
	if this.Synopsis != nil && that1.Synopsis != nil {
		if *this.Synopsis != *that1.Synopsis {
			return fmt.Errorf("Synopsis this(%v) Not Equal that(%v)", *this.Synopsis, *that1.Synopsis)
		}
	} else if this.Synopsis != nil {
		return fmt.Errorf("this.Synopsis == nil && that.Synopsis != nil")
	} else if that1.Synopsis != nil {
		return fmt.Errorf("Synopsis this(%v) Not Equal that(%v)", this.Synopsis, that1.Synopsis)
	}
	if this.Hash != that1.Hash {
		return fmt.Errorf("Hash this(%v) Not Equal that(%v)", this.Hash, that1.Hash)
	}
	if this.TopicName != that1.TopicName {
		return fmt.Errorf("TopicName this(%v) Not Equal that(%v)", this.TopicName, that1.TopicName)
	}
	/*if len(this.Sections) != len(that1.Sections) {
		return fmt.Errorf("Sections this(%v) Not Equal that(%v)", len(this.Sections), len(that1.Sections))
	}
	for i := range this.Sections {
		if !this.Sections[i].Equal(&that1.Sections[i]) {
			return fmt.Errorf("Sections this[%v](%v) Not Equal that[%v](%v)", i, this.Sections[i], i, that1.Sections[i])
		}
	}*/
	if len(this.Metadata) != len(that1.Metadata) {
		return fmt.Errorf("Metadata this(%v) Not Equal that(%v)", len(this.Metadata), len(that1.Metadata))
	}
	for i := range this.Metadata {
		if !this.Metadata[i].Equal(&that1.Metadata[i]) {
			return fmt.Errorf("Metadata this[%v](%v) Not Equal that[%v](%v)", i, this.Metadata[i], i, that1.Metadata[i])
		}
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return fmt.Errorf("XXX_unrecognized this(%v) Not Equal that(%v)", this.XXX_unrecognized, that1.XXX_unrecognized)
	}
	return nil
}

func EqualCourse(this *uct.Course, that interface{}) bool {
	if that == nil {
		if this == nil {
			return true
		}
		return false
	}

	that1, ok := that.(*uct.Course)
	if !ok {
		that2, ok := that.(uct.Course)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		if this == nil {
			return true
		}
		return false
	} else if this == nil {
		return false
	}
	if this.Id != that1.Id {
		return false
	}
	if this.SubjectId != that1.SubjectId {
		return false
	}
	if this.Name != that1.Name {
		return false
	}
	if this.Number != that1.Number {
		return false
	}
	if this.Synopsis != nil && that1.Synopsis != nil {
		if *this.Synopsis != *that1.Synopsis {
			return false
		}
	} else if this.Synopsis != nil {
		return false
	} else if that1.Synopsis != nil {
		return false
	}
	if this.Hash != that1.Hash {
		return false
	}
	if this.TopicName != that1.TopicName {
		return false
	}
	/*	if len(this.Sections) != len(that1.Sections) {
			return false
		}
		for i := range this.Sections {
			if !this.Sections[i].Equal(&that1.Sections[i]) {
				return false
			}
		}*/
	if len(this.Metadata) != len(that1.Metadata) {
		return false
	}
	for i := range this.Metadata {
		if !this.Metadata[i].Equal(&that1.Metadata[i]) {
			return false
		}
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
func VerboseEqualSection(this *uct.Section, that interface{}) error {
	if that == nil {
		if this == nil {
			return nil
		}
		return fmt.Errorf("that == nil && this != nil")
	}

	that1, ok := that.(*uct.Section)
	if !ok {
		that2, ok := that.(uct.Section)
		if ok {
			that1 = &that2
		} else {
			return fmt.Errorf("that is not of type *Section")
		}
	}
	if that1 == nil {
		if this == nil {
			return nil
		}
		return fmt.Errorf("that is type *Section but is nil && this != nil")
	} else if this == nil {
		return fmt.Errorf("that is type *Section but is not nil && this == nil")
	}
	if this.Id != that1.Id {
		return fmt.Errorf("Id this(%v) Not Equal that(%v)", this.Id, that1.Id)
	}
	if this.CourseId != that1.CourseId {
		return fmt.Errorf("CourseId this(%v) Not Equal that(%v)", this.CourseId, that1.CourseId)
	}
	if this.Number != that1.Number {
		return fmt.Errorf("Number this(%v) Not Equal that(%v)", this.Number, that1.Number)
	}
	if this.CallNumber != that1.CallNumber {
		return fmt.Errorf("CallNumber this(%v) Not Equal that(%v)", this.CallNumber, that1.CallNumber)
	}
	if this.Max != that1.Max {
		return fmt.Errorf("Max this(%v) Not Equal that(%v)", this.Max, that1.Max)
	}
	if this.Now != that1.Now {
		return fmt.Errorf("Now this(%v) Not Equal that(%v)", this.Now, that1.Now)
	}
	if this.Status != that1.Status {
		return fmt.Errorf("Status this(%v) Not Equal that(%v)", this.Status, that1.Status)
	}
	if this.Credits != that1.Credits {
		return fmt.Errorf("Credits this(%v) Not Equal that(%v)", this.Credits, that1.Credits)
	}
	if this.TopicName != that1.TopicName {
		return fmt.Errorf("TopicName this(%v) Not Equal that(%v)", this.TopicName, that1.TopicName)
	}
	if len(this.Meetings) != len(that1.Meetings) {
		return fmt.Errorf("Meetings this(%v) Not Equal that(%v)", len(this.Meetings), len(that1.Meetings))
	}
	for i := range this.Meetings {
		if !this.Meetings[i].Equal(&that1.Meetings[i]) {
			return fmt.Errorf("Meetings this[%v](%v) Not Equal that[%v](%v)", i, this.Meetings[i], i, that1.Meetings[i])
		}
	}
	if len(this.Instructors) != len(that1.Instructors) {
		return fmt.Errorf("Instructors this(%v) Not Equal that(%v)", len(this.Instructors), len(that1.Instructors))
	}
	for i := range this.Instructors {
		if !this.Instructors[i].Equal(&that1.Instructors[i]) {
			return fmt.Errorf("Instructors this[%v](%v) Not Equal that[%v](%v)", i, this.Instructors[i], i, that1.Instructors[i])
		}
	}
	if len(this.Books) != len(that1.Books) {
		return fmt.Errorf("Books this(%v) Not Equal that(%v)", len(this.Books), len(that1.Books))
	}
	for i := range this.Books {
		if !this.Books[i].Equal(&that1.Books[i]) {
			return fmt.Errorf("Books this[%v](%v) Not Equal that[%v](%v)", i, this.Books[i], i, that1.Books[i])
		}
	}
	if len(this.Metadata) != len(that1.Metadata) {
		return fmt.Errorf("Metadata this(%v) Not Equal that(%v)", len(this.Metadata), len(that1.Metadata))
	}
	for i := range this.Metadata {
		if !this.Metadata[i].Equal(&that1.Metadata[i]) {
			return fmt.Errorf("Metadata this[%v](%v) Not Equal that[%v](%v)", i, this.Metadata[i], i, that1.Metadata[i])
		}
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return fmt.Errorf("XXX_unrecognized this(%v) Not Equal that(%v)", this.XXX_unrecognized, that1.XXX_unrecognized)
	}
	return nil
}
func EqualSection(this *uct.Section, that interface{}) bool {
	if that == nil {
		if this == nil {
			return true
		}
		return false
	}

	that1, ok := that.(*uct.Section)
	if !ok {
		that2, ok := that.(uct.Section)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		if this == nil {
			return true
		}
		return false
	} else if this == nil {
		return false
	}
	if this.Id != that1.Id {
		return false
	}
	if this.CourseId != that1.CourseId {
		return false
	}
	if this.Number != that1.Number {
		return false
	}
	if this.CallNumber != that1.CallNumber {
		return false
	}
	if this.Max != that1.Max {
		return false
	}
	if this.Now != that1.Now {
		return false
	}
	if this.Status != that1.Status {
		return false
	}
	if this.Credits != that1.Credits {
		return false
	}
	if this.TopicName != that1.TopicName {
		return false
	}
	if len(this.Meetings) != len(that1.Meetings) {
		return false
	}
	for i := range this.Meetings {
		if !this.Meetings[i].Equal(&that1.Meetings[i]) {
			return false
		}
	}
	if len(this.Instructors) != len(that1.Instructors) {
		return false
	}
	for i := range this.Instructors {
		if !this.Instructors[i].Equal(&that1.Instructors[i]) {
			return false
		}
	}
	if len(this.Books) != len(that1.Books) {
		return false
	}
	for i := range this.Books {
		if !this.Books[i].Equal(&that1.Books[i]) {
			return false
		}
	}
	if len(this.Metadata) != len(that1.Metadata) {
		return false
	}
	for i := range this.Metadata {
		if !this.Metadata[i].Equal(&that1.Metadata[i]) {
			return false
		}
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
