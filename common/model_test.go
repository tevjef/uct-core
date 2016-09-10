package common

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"sort"
	"testing"
	"time"
)

var rutgers = []*Registration{
	&Registration{
		Period:     SEM_FALL.String(),
		PeriodDate: time.Date(0000, time.September, 6, 0, 0, 0, 0, time.UTC).Unix(),
	},
	&Registration{
		Period:     SEM_SPRING.String(),
		PeriodDate: time.Date(0000, time.January, 17, 0, 0, 0, 0, time.UTC).Unix(),
	},
	&Registration{
		Period:     SEM_SUMMER.String(),
		PeriodDate: time.Date(0000, time.May, 30, 0, 0, 0, 0, time.UTC).Unix(),
	},
	&Registration{
		Period:     SEM_WINTER.String(),
		PeriodDate: time.Date(0000, time.December, 23, 0, 0, 0, 0, time.UTC).Unix(),
	},
	&Registration{
		Period:     START_FALL.String(),
		PeriodDate: time.Date(0000, time.March, 20, 0, 0, 0, 0, time.UTC).Unix(),
	},
	&Registration{
		Period:     START_SPRING.String(),
		PeriodDate: time.Date(0000, time.October, 18, 0, 0, 0, 0, time.UTC).Unix(),
	},
	&Registration{
		Period:     START_SUMMER.String(),
		PeriodDate: time.Date(0000, time.January, 14, 0, 0, 0, 0, time.UTC).Unix(),
	},
	&Registration{
		Period:     START_WINTER.String(),
		PeriodDate: time.Date(0000, time.September, 21, 0, 0, 0, 0, time.UTC).Unix(),
	},
	&Registration{
		Period:     END_FALL.String(),
		PeriodDate: time.Date(0000, time.September, 13, 0, 0, 0, 0, time.UTC).Unix(),
	},
	&Registration{
		Period:     END_SPRING.String(),
		PeriodDate: time.Date(0000, time.January, 27, 0, 0, 0, 0, time.UTC).Unix(),
	},
	&Registration{
		Period:     END_SUMMER.String(),
		PeriodDate: time.Date(0000, time.June, 15, 0, 0, 0, 0, time.UTC).Unix(),
	},
	&Registration{
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
	assert.Equal(t, SUMMER, semesters.Last.Season)
	assert.Equal(t, 2016, int(semesters.Current.Year))
	assert.Equal(t, FALL, semesters.Current.Season)
	assert.Equal(t, 2016, int(semesters.Next.Year))
	assert.Equal(t, WINTER, semesters.Next.Season)

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

	semesters = ResolveSemesters(time.Date(2016, time.June, 14, 0, 0, 0, 0, time.UTC), rutgers)
	assert.Equal(t, 2016, int(semesters.Last.Year))
	assert.Equal(t, SPRING, semesters.Last.Season)
	assert.Equal(t, 2016, int(semesters.Current.Year))
	assert.Equal(t, SUMMER, semesters.Current.Season)
	assert.Equal(t, 2016, int(semesters.Next.Year))
	assert.Equal(t, FALL, semesters.Next.Season)
	//fmt.Printf("%#v\n", semesters)

	semesters = ResolveSemesters(time.Date(2016, time.June, 15, 0, 0, 0, 0, time.UTC), rutgers)
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
}

func TestToTitle(t *testing.T) {
	str := "ART APPRECIATION VIIIII"
	expect := "Art Appreciation VIIIII"

	result := toTitle(str)
	assert.Equal(t, expect, result)

	str = "ART APPRECIATION V"
	expect = "Art Appreciation V"

	result = toTitle(str)
	assert.Equal(t, expect, result)

	str = "ART APPRECIATION VI"
	expect = "Art Appreciation VI"

	result = toTitle(str)
	assert.Equal(t, expect, result)

	str = "ART APPRECIATION II"
	expect = "Art Appreciation II"

	result = toTitle(str)
	assert.Equal(t, expect, result)

	str = "ART APPRECIATION I"
	expect = "Art Appreciation I"

	result = toTitle(str)
	assert.Equal(t, expect, result)

}

func TestSwapChar(t *testing.T) {
	str := "ART APPRECIATION VII"
	expect := "ART aPPERCIATION VII"

	result := swapChar(str, "a", 4)
	assert.Equal(t, expect, result)

	expect = "ART APPRECIATION VIi"

	result = swapChar(str, "i", len(str)-1)
	assert.Equal(t, expect, result)

	expect = "aRT APPRECIATION VII"

	result = swapChar(str, "a", 0)
	assert.Equal(t, expect, result)
}

func TestTopicName(t *testing.T) {
	topic1 := "Rutgers University–New Brunswick"
	topic2 := "AFRICAN, M. EAST. & S. ASIAN LANG & LIT $ __ "
	topic3 := "Res Proposal In A....H.!@#$%^&*()_?><.02.87ASDA"

	fmt.Printf("%s\n", ToTopicName(topic1))
	fmt.Printf("%s\n", ToTopicName(topic1))

	fmt.Printf("%s\n", ToTopicName(topic2))
	fmt.Printf("%s\n", ToTopicName(topic2))

	fmt.Printf("%s\n", ToTopicName(topic3))
	fmt.Printf("%s\n", ToTopicName(topic3))

	fmt.Println("\n")

	fmt.Printf("%s\n", toTopicName(topic1))
	fmt.Printf("%s\n", toTopicName(topic1))

	fmt.Printf("%s\n", toTopicName(topic2))
	fmt.Printf("%s\n", toTopicName(topic2))

	fmt.Printf("%s\n", toTopicName(topic3))
	fmt.Printf("%s\n", toTopicName(topic3))


}

func BenchmarkToTopicName(b *testing.B) {
	str := "AFRICAN, M. EAST. & S. ASIAN LANG & LIT $ __ "

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ToTopicName(str)
	}
}