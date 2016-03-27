package main

import (
	"testing"
	"uct/common"
)

func BenchmarkSelectSubject(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		SelectSubjects(1, common.SPRING, "2016", false)
	}
}

func BenchmarkSelectCourse(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		SelectCourses(1, true)
	}
}