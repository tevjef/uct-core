package rutgers

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsAfter(t *testing.T) {
	a := []string{"8:00 AM", "11:59 AM", "12:00 PM"}
	b := []string{"9:00 AM", "12:00 PM", "1:00 PM"}

	for i := range a {
		assert.True(t, isAfter(a[i], b[i]))
	}
}

func TestStripSpaces(t *testing.T) {
	var str string
	var expected string
	var result string

	str = "      Computer     Science     "
	expected = " Computer Science "
	result = stripSpaces(str)

	assert.Equal(t, expected, result)
}

func TestTrimAll(t *testing.T) {
	var str string
	var expected string
	var result string

	str = "\u0000T\u0001\u0000\u0001E\u0000\u0001\u0000\u0001\u0000S\u0000\u0001\u0000T"
	expected = "TEST"
	result = trimAll(str)

	assert.Equal(t, expected, result)
}

func TestFormatMeetingHours(t *testing.T) {
	var str string
	var expected string
	var result string

	str = "0912"
	expected = "9:12"
	result = formatMeetingHours(str)
	assert.Equal(t, expected, result)

	str = "1212"
	expected = "12:12"
	result = formatMeetingHours(str)
	assert.Equal(t, expected, result)
}

func TestMeetingSort(t *testing.T) {
	meetings := []RMeetingTime{
		{StartTime: "6:34 PM", EndTime: "6:35 PM", MeetingDay: "Monday", MeetingModeCode: "90"},
		{StartTime: "", EndTime: "", MeetingDay: "", MeetingModeCode: "91"},
		{StartTime: "", EndTime: "", MeetingDay: "", MeetingModeCode: "90"},
		{StartTime: "8:30 PM", EndTime: "12:35 PM", MeetingDay: "Saturday", MeetingModeCode: "90"},
		{StartTime: "", EndTime: "", MeetingDay: "", MeetingModeCode: "02"},
		{StartTime: "4:30 AM", EndTime: "6:35 PM", MeetingDay: "Monday", MeetingModeCode: "90"},
		{StartTime: "1:30 PM", EndTime: "4:35 PM", MeetingDay: "Tuesday", MeetingModeCode: "90"},
		{StartTime: "11:30 AM", EndTime: "12:35 PM", MeetingDay: "Monday", MeetingModeCode: "90"},
		{StartTime: "", EndTime: "", MeetingDay: "", MeetingModeCode: "03"},
	}

	expected := []RMeetingTime{
		{StartTime: "4:30 AM", EndTime: "6:35 PM", MeetingDay: "Monday", MeetingModeCode: "90"},
		{StartTime: "11:30 AM", EndTime: "12:35 PM", MeetingDay: "Monday", MeetingModeCode: "90"},
		{StartTime: "6:34 PM", EndTime: "6:35 PM", MeetingDay: "Monday", MeetingModeCode: "90"},
		{StartTime: "1:30 PM", EndTime: "4:35 PM", MeetingDay: "Tuesday", MeetingModeCode: "90"},
		{StartTime: "8:30 PM", EndTime: "12:35 PM", MeetingDay: "Saturday", MeetingModeCode: "90"},
		{StartTime: "", EndTime: "", MeetingDay: "", MeetingModeCode: "02"},
		{StartTime: "", EndTime: "", MeetingDay: "", MeetingModeCode: "03"},
		{StartTime: "", EndTime: "", MeetingDay: "", MeetingModeCode: "91"},
		{StartTime: "", EndTime: "", MeetingDay: "", MeetingModeCode: "90"},
	}

	sort.Sort(MeetingByClass(meetings))
	assert.Equal(t, meetings, expected)
}

func TestGetMeetingHourStartTime(t *testing.T) {
	var meeting RMeetingTime
	var expected string
	var result string

	meeting = RMeetingTime{StartTime: "0430", PmCode: "A"}
	expected = "4:30 AM"
	result = meeting.getMeetingHourStart()
	assert.Equal(t, expected, result)

	meeting = RMeetingTime{StartTime: "0200", PmCode: "P"}
	expected = "2:00 PM"
	result = meeting.getMeetingHourStart()
	assert.Equal(t, expected, result)

	meeting = RMeetingTime{StartTime: "0430", PmCode: "P"}
	expected = "4:30 PM"
	result = meeting.getMeetingHourStart()
	assert.Equal(t, expected, result)
}

func TestGetMeetingHourEndTime(t *testing.T) {
	var meeting RMeetingTime
	var expected string
	var result string

	meeting = RMeetingTime{StartTime: "0000", EndTime: "0430", PmCode: "A"}
	expected = "4:30 AM"
	result = meeting.getMeetingHourEnd()
	assert.Equal(t, expected, result)

	meeting = RMeetingTime{StartTime: "0000", EndTime: "0200", PmCode: "P"}
	expected = "2:00 PM"
	result = meeting.getMeetingHourEnd()
	assert.Equal(t, expected, result)

	meeting = RMeetingTime{StartTime: "0000", EndTime: "0430", PmCode: "P"}
	expected = "4:30 PM"
	result = meeting.getMeetingHourEnd()
	assert.Equal(t, expected, result)

	meeting = RMeetingTime{EndTime: "030", PmCode: "P"}
	expected = ""
	result = meeting.getMeetingHourEnd()
	assert.Equal(t, expected, result)

	meeting = RMeetingTime{StartTime: "030", PmCode: "P"}
	expected = ""
	result = meeting.getMeetingHourEnd()
	assert.Equal(t, expected, result)
}
