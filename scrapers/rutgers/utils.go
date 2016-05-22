package main

import (
	"strconv"
	"time"
	uct "uct/common"
)

func formatMeetingHours(time string) string {
	if len(time) > 1 {
		if time[:1] == "0" {
			return time[1:2] + ":" + time[2:]
		}
		return time[:2] + ":" + time[2:]
	}
	return ""
}

func (meetingTime RMeetingTime) getMeetingHourEndTime() time.Time {
	if len(uct.TrimAll(meetingTime.StartTime)) > 1 || len(uct.TrimAll(meetingTime.EndTime)) > 1 {
		var meridian string
		starttime := meetingTime.StartTime
		endtime := meetingTime.EndTime
		pmcode := meetingTime.PmCode

		end, _ := strconv.Atoi(endtime[:2])
		start, _ := strconv.Atoi(starttime[:2])

		if pmcode != "A" {
			meridian = "PM"
		} else if end < start {
			meridian = "PM"
		} else if endtime[:2] == "12" {
			meridian = "AM"
		} else {
			meridian = "AM"
		}

		time, err := time.Parse(time.Kitchen, formatMeetingHours(meetingTime.EndTime)+meridian)
		uct.CheckError(err)
		return time
	}
	return time.Unix(0, 0)
}

func (meeting RMeetingTime) isByArrangement() bool {
	return meeting.MeetingModeCode == "B"
}

func (meeting RMeetingTime) isStudio() bool {
	return meeting.MeetingModeCode == "07"
}

func (meeting RMeetingTime) isLab() bool {
	return meeting.MeetingModeCode == "05"
}

func (meeting RMeetingTime) isRecitation() bool {
	return meeting.MeetingModeCode == "03"
}

func (meeting RMeetingTime) isLecture() bool {
	return meeting.MeetingModeCode == "02"
}

func (meeting RMeetingTime) day() string {
	var day string
	switch meeting.MeetingDay {
	case "M":
		day = "Monday"
	case "T":
		day = "Tuesday"
	case "W":
		day = "Wednesday"
	case "TH":
		day = "Thursday"
	case "F":
		day = "Friday"
	case "S":
		day = "Saturday"
	case "U":
		day = "Sunday"
	}
	if len(day) == 0 {
		return ""
	} else {
		return day
	}
}

func (meeting RMeetingTime) dayPointer() *string {
	if meeting.MeetingDay == "" {
		return nil
	} else {
		return &meeting.MeetingDay
	}
}

func (meeting RMeetingTime) classType() string {
	if meeting.isLab() {
		return "Lab"
	} else if meeting.isStudio() {
		return "Studio"
	} else if meeting.isByArrangement() {
		return "Hours By Arrangement"
	} else if meeting.isLecture() {
		return "Lecture"
	} else if meeting.isRecitation() {
		return "Recitation"
	}
	return meeting.MeetingModeDesc
}

func IsAfter(t1, t2 string) bool {
	if l1 := len(t1); l1 == 7 {
		t1 = "0" + t1
	} else if l1 == 0 {
		return false
	}
	if l2 := len(t2); l2 == 7 {
		t2 = "0" + t2
	} else if l2 == 0 {
		return false
	}
	if t1[:2] == "12" {
		t1 = t1[2:]
		t1 = "00" + t1
	}
	if t2[:2] == "12" {
		t2 = t2[2:]
		t2 = "00" + t2
	}
	if t1[6:] == "AM" && t2[6:] == "PM" {
		return true
	}
	if t1[:2] == t2[:2] {
		return t1[3:5] < t2[3:5]
	}
	return t1[:2] < t2[:2]
}
