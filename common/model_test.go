package common

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var rutgers = []Registration{
	Registration{
		Period:     SEM_FALL.String(),
		PeriodDate: time.Date(0000, time.September, 6, 0, 0, 0, 0, time.UTC).Unix(),
	},
	Registration{
		Period:     SEM_SPRING.String(),
		PeriodDate: time.Date(0000, time.January, 17, 0, 0, 0, 0, time.UTC).Unix(),
	},
	Registration{
		Period:     SEM_SUMMER.String(),
		PeriodDate: time.Date(0000, time.May, 30, 0, 0, 0, 0, time.UTC).Unix(),
	},
	Registration{
		Period:     SEM_WINTER.String(),
		PeriodDate: time.Date(0000, time.December, 23, 0, 0, 0, 0, time.UTC).Unix(),
	},
	Registration{
		Period:     START_FALL.String(),
		PeriodDate: time.Date(0000, time.March, 20, 0, 0, 0, 0, time.UTC).Unix(),
	},
	Registration{
		Period:     START_SPRING.String(),
		PeriodDate: time.Date(0000, time.October, 18, 0, 0, 0, 0, time.UTC).Unix(),
	},
	Registration{
		Period:     START_SUMMER.String(),
		PeriodDate: time.Date(0000, time.January, 14, 0, 0, 0, 0, time.UTC).Unix(),
	},
	Registration{
		Period:     START_WINTER.String(),
		PeriodDate: time.Date(0000, time.September, 21, 0, 0, 0, 0, time.UTC).Unix(),
	},
	Registration{
		Period:     END_FALL.String(),
		PeriodDate: time.Date(0000, time.September, 13, 0, 0, 0, 0, time.UTC).Unix(),
	},
	Registration{
		Period:     END_SPRING.String(),
		PeriodDate: time.Date(0000, time.January, 27, 0, 0, 0, 0, time.UTC).Unix(),
	},
	Registration{
		Period:     END_SUMMER.String(),
		PeriodDate: time.Date(0000, time.August, 15, 0, 0, 0, 0, time.UTC).Unix(),
	},
	Registration{
		Period:     END_WINTER.String(),
		PeriodDate: time.Date(0000, time.December, 22, 0, 0, 0, 0, time.UTC).Unix(),
	}}

/*
const (
	FALL Season = iota
	SPRING
	SUMMER
	WINTER
)
*/

func TestResolveSemesters(t *testing.T) {

	semesters := ResolveSemesters(time.Date(2015, time.December, 24, 0, 0, 0, 0, time.UTC), rutgers)
	assert.Equal(t, semesters.Last.Year, 2015)
	assert.Equal(t, semesters.Last.Season, WINTER)
	assert.Equal(t, semesters.Current.Year, 2016)
	assert.Equal(t, semesters.Current.Season, SPRING)
	assert.Equal(t, semesters.Next.Year, 2016)
	assert.Equal(t, semesters.Next.Season, SUMMER)
	//fmt.Printf("%#v\n", semesters)

	semesters = ResolveSemesters(time.Date(2016, time.January, 8, 0, 0, 0, 0, time.UTC), rutgers)
	assert.Equal(t, semesters.Last.Year, 2015)
	assert.Equal(t, semesters.Last.Season, WINTER)
	assert.Equal(t, semesters.Current.Year, 2016)
	assert.Equal(t, semesters.Current.Season, SPRING)
	assert.Equal(t, semesters.Next.Year, 2016)
	assert.Equal(t, semesters.Next.Season, SUMMER)
	//fmt.Printf("%#v\n", semesters)

	semesters = ResolveSemesters(time.Date(2016, time.March, 19, 0, 0, 0, 0, time.UTC), rutgers)
	assert.Equal(t, semesters.Last.Year, 2015)
	assert.Equal(t, semesters.Last.Season, WINTER)
	assert.Equal(t, semesters.Current.Year, 2016)
	assert.Equal(t, semesters.Current.Season, SPRING)
	assert.Equal(t, semesters.Next.Year, 2016)
	assert.Equal(t, semesters.Next.Season, SUMMER)
	//fmt.Printf("%#v\n", semesters)

	semesters = ResolveSemesters(time.Date(2016, time.March, 20, 0, 0, 0, 0, time.UTC), rutgers)
	assert.Equal(t, semesters.Last.Year, 2016)
	assert.Equal(t, semesters.Last.Season, SPRING)
	assert.Equal(t, semesters.Current.Year, 2016)
	assert.Equal(t, semesters.Current.Season, SUMMER)
	assert.Equal(t, semesters.Next.Year, 2016)
	assert.Equal(t, semesters.Next.Season, FALL)
	//fmt.Printf("%#v\n", semesters)

	semesters = ResolveSemesters(time.Date(2016, time.April, 30, 0, 0, 0, 0, time.UTC), rutgers)
	assert.Equal(t, semesters.Last.Year, 2016)
	assert.Equal(t, semesters.Last.Season, SPRING)
	assert.Equal(t, semesters.Current.Year, 2016)
	assert.Equal(t, semesters.Current.Season, SUMMER)
	assert.Equal(t, semesters.Next.Year, 2016)
	assert.Equal(t, semesters.Next.Season, FALL)
	//fmt.Printf("%#v\n", semesters)

	semesters = ResolveSemesters(time.Date(2016, time.August, 14, 0, 0, 0, 0, time.UTC), rutgers)
	assert.Equal(t, semesters.Last.Year, 2016)
	assert.Equal(t, semesters.Last.Season, SPRING)
	assert.Equal(t, semesters.Current.Year, 2016)
	assert.Equal(t, semesters.Current.Season, SUMMER)
	assert.Equal(t, semesters.Next.Year, 2016)
	assert.Equal(t, semesters.Next.Season, FALL)
	//fmt.Printf("%#v\n", semesters)

	semesters = ResolveSemesters(time.Date(2016, time.August, 15, 0, 0, 0, 0, time.UTC), rutgers)
	assert.Equal(t, semesters.Last.Year, 2016)
	assert.Equal(t, semesters.Last.Season, SUMMER)
	assert.Equal(t, semesters.Current.Year, 2016)
	assert.Equal(t, semesters.Current.Season, FALL)
	assert.Equal(t, semesters.Next.Year, 2016)
	assert.Equal(t, semesters.Next.Season, WINTER)
	//fmt.Printf("%#v\n", semesters)

	semesters = ResolveSemesters(time.Date(2016, time.September, 15, 0, 0, 0, 0, time.UTC), rutgers)
	assert.Equal(t, semesters.Last.Year, 2016)
	assert.Equal(t, semesters.Last.Season, SUMMER)
	assert.Equal(t, semesters.Current.Year, 2016)
	assert.Equal(t, semesters.Current.Season, FALL)
	assert.Equal(t, semesters.Next.Year, 2016)
	assert.Equal(t, semesters.Next.Season, WINTER)
	//fmt.Printf("%#v\n", semesters)

	semesters = ResolveSemesters(time.Date(2016, time.October, 17, 0, 0, 0, 0, time.UTC), rutgers)
	assert.Equal(t, semesters.Last.Year, 2016)
	assert.Equal(t, semesters.Last.Season, SUMMER)
	assert.Equal(t, semesters.Current.Year, 2016)
	assert.Equal(t, semesters.Current.Season, FALL)
	assert.Equal(t, semesters.Next.Year, 2016)
	assert.Equal(t, semesters.Next.Season, WINTER)

	semesters = ResolveSemesters(time.Date(2016, time.October, 18, 0, 0, 0, 0, time.UTC), rutgers)
	assert.Equal(t, semesters.Last.Year, 2016)
	assert.Equal(t, semesters.Last.Season, FALL)
	assert.Equal(t, semesters.Current.Year, 2016)
	assert.Equal(t, semesters.Current.Season, WINTER)
	assert.Equal(t, semesters.Next.Year, 2017)
	assert.Equal(t, semesters.Next.Season, SPRING)
	//fmt.Printf("%#v\n", semesters)

	semesters = ResolveSemesters(time.Date(2016, time.November, 1, 0, 0, 0, 0, time.UTC), rutgers)
	assert.Equal(t, semesters.Last.Year, 2016)
	assert.Equal(t, semesters.Last.Season, FALL)
	assert.Equal(t, semesters.Current.Year, 2016)
	assert.Equal(t, semesters.Current.Season, WINTER)
	assert.Equal(t, semesters.Next.Year, 2017)
	assert.Equal(t, semesters.Next.Season, SPRING)
	//fmt.Printf("%#v\n", semesters)

	semesters = ResolveSemesters(time.Date(2016, time.December, 22, 0, 0, 0, 0, time.UTC), rutgers)
	assert.Equal(t, semesters.Last.Year, 2016)
	assert.Equal(t, semesters.Last.Season, FALL)
	assert.Equal(t, semesters.Current.Year, 2016)
	assert.Equal(t, semesters.Current.Season, WINTER)
	assert.Equal(t, semesters.Next.Year, 2017)
	assert.Equal(t, semesters.Next.Season, SPRING)
	//fmt.Printf("%#v\n", semesters)

}
