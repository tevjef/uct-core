package model

import (
	"testing"
	"sort"
	"fmt"
	"bytes"
)

func TestMeetingSort(t *testing.T) {
	meetings := []RMeetingTime{
		{StartTime: "6:34 PM", EndTime: "6:35 PM", MeetingDay: "Monday", MeetingModeCode: "03"},
		{StartTime: "8:30 PM", EndTime: "12:35 PM", MeetingDay: "Saturday", MeetingModeCode: "02"},
		{StartTime: "12:30 AM", EndTime: "12:35 PM", MeetingDay: "Friday", MeetingModeCode: "02"},
		{StartTime: "4:30 AM", EndTime: "6:35 PM", MeetingDay: "Monday", MeetingModeCode: "03"},
		{StartTime: "1:30 PM", EndTime: "4:35 PM", MeetingDay: "Tuesday", MeetingModeCode: "07"},
		{StartTime: "11:30 AM", EndTime: "12:35 PM", MeetingDay: "Monday", MeetingModeCode: "02"},
		{StartTime: "6:30 PM", EndTime: "6:35 PM", MeetingDay: "Monday", MeetingModeCode: "03"},
	}

	expected := []RMeetingTime{
		{StartTime: "4:30 AM", EndTime: "6:35 PM", MeetingDay: "Monday", MeetingModeCode: "03"},
		{StartTime: "11:30 AM", EndTime: "12:35 PM", MeetingDay: "Monday", MeetingModeCode: "02"},
		{StartTime: "6:30 PM", EndTime: "6:35 PM", MeetingDay: "Monday", MeetingModeCode: "03"},
		{StartTime: "6:34 PM", EndTime: "6:35 PM", MeetingDay: "Monday", MeetingModeCode: "03"},
		{StartTime: "1:30 PM", EndTime: "4:35 PM", MeetingDay: "Tuesday", MeetingModeCode: "07"},
		{StartTime: "12:30 AM", EndTime: "12:35 PM", MeetingDay: "Friday", MeetingModeCode: "02"},
		{StartTime: "8:30 PM", EndTime: "12:35 PM", MeetingDay: "Saturday", MeetingModeCode: "02"},
	}

	printMeeting(meetings)
	sort.Sort(MeetingByClass(meetings))
	printMeeting(meetings)
	printMeeting(expected)
}

func printMeeting(meetings []RMeetingTime) {
	for i := range meetings {
		m := meetings[i]
		fmt.Printf("%-9s %-8s %-8s %s\n", m.MeetingDay, m.StartTime, m.EndTime, *m.ClassType())
	}
	fmt.Printf("\n\n")
}

func TestFo(t *testing.T) {
	str := "\u0000\u00015\u0000\u00011\u0000\u00011\u0000\u0001N\u0000 \u0000\u0001N\u0000"
	str = string(bytes.Replace([]byte(str), []byte("\x00"), []byte(""), -1))
	str = string(bytes.Replace([]byte(str), []byte("\x01"), []byte(""), -1))

	/*
		sub := common.Section{Number: "Tevin Jeffrey", Status:common.OPEN.String()}
		out, _ := json.Marshal(sub)*/
	fmt.Println(string(str))
	fmt.Printf("%x", string(str))

}
