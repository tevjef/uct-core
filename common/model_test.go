package common

import (
	"github.com/stretchr/testify/assert"
	"log"
	"sort"
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

	semesters := ResolveSemesters(time.Now(), rutgers)
	assert.Equal(t, 2016, int(semesters.Last.Year))
	assert.Equal(t, SPRING, semesters.Last.Season)
	assert.Equal(t, 2016, int(semesters.Current.Year))
	assert.Equal(t, SUMMER, semesters.Current.Season)
	assert.Equal(t, 2016, int(semesters.Next.Year))
	assert.Equal(t, FALL, semesters.Next.Season)

	semesters = ResolveSemesters(time.Date(2015, time.December, 24, 0, 0, 0, 0, time.UTC), rutgers)
	assert.Equal(t, 2015, int(semesters.Last.Year))
	assert.Equal(t, WINTER, semesters.Last.Season)
	assert.Equal(t, 2016, int(semesters.Current.Year))
	assert.Equal(t, SPRING, semesters.Current.Season)
	assert.Equal(t, 2016, int(semesters.Next.Year))
	assert.Equal(t, SUMMER, semesters.Next.Season)
	//fmt.Printf("%#v\n", semesters)

	semesters = ResolveSemesters(time.Date(2016, time.January, 8, 0, 0, 0, 0, time.UTC), rutgers)
	assert.Equal(t, 2015, int(semesters.Last.Year))
	assert.Equal(t, WINTER, semesters.Last.Season)
	assert.Equal(t, 2016, int(semesters.Current.Year))
	assert.Equal(t, SPRING, semesters.Current.Season)
	assert.Equal(t, 2016, int(semesters.Next.Year))
	assert.Equal(t, SUMMER, semesters.Next.Season)
	//fmt.Printf("%#v\n", semesters)

	semesters = ResolveSemesters(time.Date(2016, time.March, 19, 0, 0, 0, 0, time.UTC), rutgers)
	assert.Equal(t, 2015, int(semesters.Last.Year))
	assert.Equal(t, WINTER, semesters.Last.Season)
	assert.Equal(t, 2016, int(semesters.Current.Year))
	assert.Equal(t, SPRING, semesters.Current.Season)
	assert.Equal(t, 2016, int(semesters.Next.Year))
	assert.Equal(t, SUMMER, semesters.Next.Season)
	//fmt.Printf("%#v\n", semesters)

	semesters = ResolveSemesters(time.Date(2016, time.March, 20, 0, 0, 0, 0, time.UTC), rutgers)
	assert.Equal(t, 2016, int(semesters.Last.Year))
	assert.Equal(t, SPRING, semesters.Last.Season)
	assert.Equal(t, 2016, int(semesters.Current.Year))
	assert.Equal(t, SUMMER, semesters.Current.Season)
	assert.Equal(t, 2016, int(semesters.Next.Year))
	assert.Equal(t, FALL, semesters.Next.Season)
	//fmt.Printf("%#v\n", semesters)

	semesters = ResolveSemesters(time.Date(2016, time.April, 30, 0, 0, 0, 0, time.UTC), rutgers)
	assert.Equal(t, 2016, int(semesters.Last.Year))
	assert.Equal(t, SPRING, semesters.Last.Season)
	assert.Equal(t, 2016, int(semesters.Current.Year))
	assert.Equal(t, SUMMER, semesters.Current.Season)
	assert.Equal(t, 2016, int(semesters.Next.Year))
	assert.Equal(t, FALL, semesters.Next.Season)
	//fmt.Printf("%#v\n", semesters)

	semesters = ResolveSemesters(time.Date(2016, time.August, 14, 0, 0, 0, 0, time.UTC), rutgers)
	assert.Equal(t, 2016, int(semesters.Last.Year))
	assert.Equal(t, SPRING, semesters.Last.Season)
	assert.Equal(t, 2016, int(semesters.Current.Year))
	assert.Equal(t, SUMMER, semesters.Current.Season)
	assert.Equal(t, 2016, int(semesters.Next.Year))
	assert.Equal(t, FALL, semesters.Next.Season)
	//fmt.Printf("%#v\n", semesters)

	semesters = ResolveSemesters(time.Date(2016, time.August, 15, 0, 0, 0, 0, time.UTC), rutgers)
	assert.Equal(t, 2016, int(semesters.Last.Year))
	assert.Equal(t, SUMMER, semesters.Last.Season)
	assert.Equal(t, 2016, int(semesters.Current.Year))
	assert.Equal(t, FALL, semesters.Current.Season)
	assert.Equal(t, 2016, int(semesters.Next.Year))
	assert.Equal(t, WINTER, semesters.Next.Season)
	//fmt.Printf("%#v\n", semesters)

	semesters = ResolveSemesters(time.Date(2016, time.September, 15, 0, 0, 0, 0, time.UTC), rutgers)
	assert.Equal(t, 2016, int(semesters.Last.Year))
	assert.Equal(t, SUMMER, semesters.Last.Season)
	assert.Equal(t, 2016, int(semesters.Current.Year))
	assert.Equal(t, FALL, semesters.Current.Season)
	assert.Equal(t, 2016, int(semesters.Next.Year))
	assert.Equal(t, WINTER, semesters.Next.Season)
	//fmt.Printf("%#v\n", semesters)

	semesters = ResolveSemesters(time.Date(2016, time.October, 17, 0, 0, 0, 0, time.UTC), rutgers)
	assert.Equal(t, 2016, int(semesters.Last.Year))
	assert.Equal(t, SUMMER, semesters.Last.Season)
	assert.Equal(t, 2016, int(semesters.Current.Year))
	assert.Equal(t, FALL, semesters.Current.Season)
	assert.Equal(t, 2016, int(semesters.Next.Year))
	assert.Equal(t, WINTER, semesters.Next.Season)

	semesters = ResolveSemesters(time.Date(2016, time.October, 18, 0, 0, 0, 0, time.UTC), rutgers)
	assert.Equal(t, 2016, int(semesters.Last.Year))
	assert.Equal(t, FALL, semesters.Last.Season)
	assert.Equal(t, 2016, int(semesters.Current.Year))
	assert.Equal(t, WINTER, semesters.Current.Season)
	assert.Equal(t, 2017, int(semesters.Next.Year))
	assert.Equal(t, SPRING, semesters.Next.Season)
	//fmt.Printf("%#v\n", semesters)

	semesters = ResolveSemesters(time.Date(2016, time.November, 1, 0, 0, 0, 0, time.UTC), rutgers)
	assert.Equal(t, 2016, int(semesters.Last.Year))
	assert.Equal(t, FALL, semesters.Last.Season)
	assert.Equal(t, 2016, int(semesters.Current.Year))
	assert.Equal(t, WINTER, semesters.Current.Season)
	assert.Equal(t, 2017, int(semesters.Next.Year))
	assert.Equal(t, SPRING, semesters.Next.Season)
	//fmt.Printf("%#v\n", semesters)

	semesters = ResolveSemesters(time.Date(2016, time.December, 22, 0, 0, 0, 0, time.UTC), rutgers)
	assert.Equal(t, 2016, int(semesters.Last.Year))
	assert.Equal(t, FALL, semesters.Last.Season)
	assert.Equal(t, 2016, int(semesters.Current.Year))
	assert.Equal(t, WINTER, semesters.Current.Season)
	assert.Equal(t, 2017, int(semesters.Next.Year))
	assert.Equal(t, SPRING, semesters.Next.Season)
	//fmt.Printf("%#v\n", semesters)

}

func TestMeetingSorter(t *testing.T) {
	m := "Monday"
	tu := "Tuesday"
	w := "Wednesday"
	th := "Thursday"
	f := "Friday"

	meetings := []Meeting{Meeting{Day: &tu},
		Meeting{Day: &w}, Meeting{Day: &th}, Meeting{Day: &m}, Meeting{Day: &f}}

	sort.Sort(meetingSorter{meetings})

	if meetings[0].Day != &m {
		log.Printf("%s != %s", meetings[0].Day, &m)
		t.Fail()
	}
	/*for i := range meetings {
		switch i {
		case 0:
			if meetings[0].Day != &m {
				log.Printf("%s != %s", meetings[0].Day,&m)
				t.Fail()
			}
		case 1:
			if meetings[1].Day != &tu {
				log.Printf("%s != %s", meetings[1].Day,&tu)
				t.Fail()
			}
		case 2:
			if meetings[2].Day != &w {
				log.Printf("%s != %s", meetings[2].Day,&w)
				t.Fail()
			}
		}
	}*/
}
