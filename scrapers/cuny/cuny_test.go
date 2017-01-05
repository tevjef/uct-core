package main

import (
	"reflect"
	"testing"

	log "github.com/Sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.TextFormatter{})
}

func TestSplitMeeting(t *testing.T) {
	type args struct {
		meeting string
	}
	tests := []struct {
		name string
		args args
		want [3]string
	}{
		{args: args{"Saturday 12:30PM - 1:45PM"}, want: [3]string{"Saturday", "12:30PM", "1:45PM"}},
	}
	for _, tt := range tests {
		if got := splitMeeting(tt.args.meeting); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. expandMeeting() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestParseTime(t *testing.T) {
	type args struct {
		t string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{args: args{"1:30PM"}, want: "1:30 PM"},
	}
	for _, tt := range tests {
		if got := parseTime(tt.args.t); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. expandMeeting() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestExpandMeeting(t *testing.T) {
	type args struct {
		meeting string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{args: args{"TuTh 8:50AM - 9:50AM"}, want: []string{"Tuesday 8:50AM - 9:50AM", "Thursday 8:50AM - 9:50AM"}},
	}
	for _, tt := range tests {
		if got := expandMeeting(tt.args.meeting); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. expandMeeting() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestParseMeeting(t *testing.T) {
	type args struct {
		meeting string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{args: args{
			"TuThMoWeFrSa 8:50AM - 9:50AM"},
			want: []string{
				"Tuesday 8:50AM - 9:50AM",
				"Thursday 8:50AM - 9:50AM",
				"Monday 8:50AM - 9:50AM",
				"Wednesday 8:50AM - 9:50AM",
				"Friday 8:50AM - 9:50AM",
				"Saturday 8:50AM - 9:50AM"}},
	}
	for _, tt := range tests {
		if got := expandMeeting(tt.args.meeting); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. expandMeeting() = %v, want %v", tt.name, got, tt.want)
		}
	}
}
