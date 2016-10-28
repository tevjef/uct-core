package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var courses = []NCourse{
	{CourseTitle: "a"},
	{CourseTitle: "a"},
	{CourseTitle: "a"},
	{CourseTitle: "a"},
	{CourseTitle: "b"},
	{CourseTitle: "b"},
	{CourseTitle: "b"},
	{CourseTitle: "b"},
	{CourseTitle: "c"},
	{CourseTitle: "c"},
	{CourseTitle: "c"},
	{CourseTitle: "d"},
	{CourseTitle: "e"},
	{CourseTitle: "e"},
	{CourseTitle: "e"},
	{CourseTitle: "e"},
	{CourseTitle: "f"},
	{CourseTitle: "f"},
	{CourseTitle: "f"},
	{CourseTitle: "f"},
	{CourseTitle: "f"},
}

func Test_collapseCourses(t *testing.T) {
	result := collapseCourses(courses)
	expected := 6
	assert.True(t, len(result) == expected)
}

func Benchmark_collapseCourses(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		collapseCourses(courses)
	}
}

func Test_formatMeetingHour(t *testing.T) {
	type args struct {
		timeStr string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{args:args{"1099"}, want:"10:99 AM"},
		{args:args{"0830"}, want:"8:30 AM"},
		{args:args{"1000"}, want:"10:00 AM"},
		{args:args{"1200"}, want:"12:00 PM"},
		{args:args{"1230"}, want:"12:30 PM"},
		{args:args{"1300"}, want:"1:00 PM"},
	}
	for _, tt := range tests {
		if got := formatMeetingHour(tt.args.timeStr); got != tt.want {
			t.Errorf("%q. formatMeetingHour() = %v, want %v", tt.name, got, tt.want)
		}
	}
}
