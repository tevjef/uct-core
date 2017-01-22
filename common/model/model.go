package model

import (
	"time"

	log "github.com/Sirupsen/logrus"
)

type (
	Period int
	Status int

	DBResolvedSemester struct {
		Id            int64  `db:"id"`
		UniversityId  int64  `db:"university_id"`
		CurrentSeason string `db:"current_season"`
		CurrentYear   string `db:"current_year"`
		LastSeason    string `db:"last_season"`
		LastYear      string `db:"last_year"`
		NextSeason    string `db:"next_season"`
		NextYear      string `db:"next_year"`
	}
)

const (
	Protobuf = "protobuf"
	Json     = "json"
)

const (
	InFall Period = iota
	InSpring
	InSummer
	InWinter
	StartFall
	StartSpring
	StartSummer
	StartWinter
	EndFall
	EndSpring
	EndSummer
	EndWinter
)

var period = [...]string{
	"fall",
	"spring",
	"summer",
	"winter",
	"start_fall",
	"start_spring",
	"start_summer",
	"start_winter",
	"end_fall",
	"end_spring",
	"end_summer",
	"end_winter",
}

func (s Period) String() string {
	return period[s]
}

const (
	Fall   = "fall"
	Spring = "spring"
	Summer = "summer"
	Winter = "winter"
)

const (
	Open Status = 1 + iota
	Closed
)

var status = [...]string{
	"Open",
	"Closed",
}

func (s Status) String() string {
	return status[s-1]
}

func (r Registration) month() time.Month {
	return time.Unix(r.PeriodDate, 0).UTC().Month()
}

func (r Registration) day() int {
	return time.Unix(r.PeriodDate, 0).UTC().Day()
}

func (r Registration) dayOfYear() int {
	return time.Unix(r.PeriodDate, 0).UTC().YearDay()
}

func (r Registration) season() string {
	switch r.Period {
	case InFall.String():
		return Fall
	case InSpring.String():
		return Spring
	case InSummer.String():
		return Summer
	case InWinter.String():
		return Winter
	default:
		return Summer
	}
}

func ResolveSemesters(t time.Time, registration []*Registration) *ResolvedSemester {
	month := t.Month()
	day := t.Day()
	year := t.Year()

	yearDay := t.YearDay()

	//var springReg = registration[SEM_SPRING];
	var winterReg = registration[InWinter]
	//var summerReg = registration[SEM_SUMMER];
	//var fallReg  = registration[SEM_FALL];
	var startFallReg = registration[StartFall]
	var startSpringReg = registration[StartSpring]
	var endSummerReg = registration[EndSummer]
	//var startSummerReg  = registration[START_SUMMER];

	fall := &Semester{
		Year:   int32(year),
		Season: Fall}

	winter := &Semester{
		Year:   int32(year),
		Season: Winter}

	spring := &Semester{
		Year:   int32(year),
		Season: Spring}

	summer := &Semester{
		Year:   int32(year),
		Season: Summer}

	// Spring: Winter - StartFall
	log.Println(winterReg.month(), winterReg.day(), startFallReg.month(), startFallReg.day(), month, day)
	if (month >= winterReg.month() && day >= winterReg.day()) || month <= startFallReg.month() || (month <= startFallReg.month() && day <= startFallReg.day()) {
		if winterReg.month()-month <= 0 {
			spring.Year = spring.Year + 1
			summer.Year = summer.Year + 1
		} else {
			winter.Year = winter.Year - 1
			fall.Year = fall.Year - 1
		}
		log.Debugln("Spring: Winter - StartFall ", winterReg.month(), winterReg.day(), "--", startFallReg.month(), startFallReg.day(), "--", month, day)
		return &ResolvedSemester{
			Last:    winter,
			Current: spring,
			Next:    summer}

	} else if yearDay >= startFallReg.dayOfYear() && yearDay < endSummerReg.dayOfYear() {
		log.Debugln("StartFall: StartFall -- EndSummer ", startFallReg.dayOfYear(), "--", endSummerReg.dayOfYear(), "--", yearDay)
		return &ResolvedSemester{
			Last:    spring,
			Current: summer,
			Next:    fall,
		}
	} else if yearDay >= endSummerReg.dayOfYear() && yearDay < startSpringReg.dayOfYear() {
		log.Debugln("Fall: EndSummer -- StartSpring ", endSummerReg.dayOfYear(), "--", yearDay < startSpringReg.dayOfYear(), "--", yearDay)
		return &ResolvedSemester{
			Last:    summer,
			Current: fall,
			Next:    winter,
		}
	} else if yearDay >= startSpringReg.dayOfYear() && yearDay < winterReg.dayOfYear() {
		spring.Year = spring.Year + 1
		log.Debugln("StartSpring: StartSpring -- Winter ", startSpringReg.dayOfYear(), "--", winterReg.dayOfYear(), "--", yearDay)
		return &ResolvedSemester{
			Last:    fall,
			Current: winter,
			Next:    spring,
		}
	}

	return &ResolvedSemester{}
}
