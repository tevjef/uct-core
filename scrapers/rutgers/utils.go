package main

func formatMeetingHours(time string) string {
	if len(time) > 1 {
		if time[:1] == "0" {
			return time[1:2] + ":" + time[2:]
		}
		return time[:2] + ":" + time[2:]
	}
	return ""
}

func (meeting RMeetingTime) isByArrangement() bool {
	return meeting.BaClassHours == "B"
}

func (meeting RMeetingTime) isStudio() bool {
	return meeting.BaClassHours == "07"
}

func (meeting RMeetingTime) isLab() bool {
	return meeting.BaClassHours == "05"
}

func (meeting RMeetingTime) isRecitation() bool {
	return meeting.BaClassHours == "03"
}

func (meeting RMeetingTime) isLecture() bool {
	return meeting.BaClassHours == "02"
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
