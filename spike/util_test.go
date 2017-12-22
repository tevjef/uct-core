package main

import "testing"

func Test_parseAndroid(t *testing.T) {
	type args struct {
		userAgent string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 string
	}{
		{
			name: "1",
			args: args{
				userAgent: "Rutgers Course Tracker/com.tevinjeffrey.rutgersct (1.0.7.0R; Android 27)",
			},
			want:  "1.0.7.0R",
			want1: "Android 27",
		},
		{
			name: "2",
			args: args{
				userAgent: "(1.0.7.0R; Android 27)",
			},
			want:  "1.0.7.0R",
			want1: "Android 27",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := parseAndroid(tt.args.userAgent)
			if got != tt.want {
				t.Errorf("parseAndroid() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("parseAndroid() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_parseios(t *testing.T) {
	type args struct {
		userAgent string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 string
	}{
		{
			name: "1",
			args: args{
				userAgent: "UniversityCourseTracker/com.tevinjeffrey.uctios (100; iOS 10.2.1) Course Tracker/1.0.7",
			},
			want:  "1.0.7",
			want1: "iOS 10.2.1",
		},
		{
			name: "2",
			args: args{
				userAgent: "UniversityCourseTracker/com.tevinjeffrey.uctios (100; iOS 10.2.1) Course Tracker/1.0.7",
			},
			want:  "1.0.7",
			want1: "iOS 10.2.1",
		},
		{
			name: "3",
			args: args{
				userAgent: "UniversityCourseTracker/1.0.8 (com.tevinjeffrey.uctios; build:100; iOS 10.3.1) Alamofire/4.5.1 Course Tracker/1.0.8",
			},
			want:  "1.0.8",
			want1: "iOS 10.3.1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := parseios(tt.args.userAgent)
			if got != tt.want {
				t.Errorf("parseios() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("parseios() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
