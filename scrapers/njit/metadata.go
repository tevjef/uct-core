package main

import (
	"time"

	"github.com/tevjef/uct-backend/common/model"
)

var njitBase = model.University{
	Name:             "New Jersey Institute of Technology",
	Abbr:             "NJIT",
	HomePage:         "http://www.njit.edu/",
	RegistrationPage: "https://my.njit.edu/",
	Registrations: []*model.Registration{
		{
			Period:     model.InFall.String(),
			PeriodDate: time.Date(2000, time.September, 6, 0, 0, 0, 0, time.UTC).Unix(),
		},
		{
			Period:     model.InSpring.String(),
			PeriodDate: time.Date(2000, time.January, 17, 0, 0, 0, 0, time.UTC).Unix(),
		},
		{
			Period:     model.InSummer.String(),
			PeriodDate: time.Date(2000, time.May, 30, 0, 0, 0, 0, time.UTC).Unix(),
		},
		{
			Period:     model.InWinter.String(),
			PeriodDate: time.Date(2000, time.December, 23, 0, 0, 0, 0, time.UTC).Unix(),
		},
		{
			Period:     model.StartFall.String(),
			PeriodDate: time.Date(2000, time.March, 20, 0, 0, 0, 0, time.UTC).Unix(),
		},
		{
			Period:     model.StartSpring.String(),
			PeriodDate: time.Date(2000, time.October, 5, 0, 0, 0, 0, time.UTC).Unix(),
		},
		{
			Period:     model.StartSummer.String(),
			PeriodDate: time.Date(2000, time.January, 14, 0, 0, 0, 0, time.UTC).Unix(),
		},
		{
			Period:     model.StartWinter.String(),
			PeriodDate: time.Date(2000, time.September, 21, 0, 0, 0, 0, time.UTC).Unix(),
		},
		{
			Period:     model.EndFall.String(),
			PeriodDate: time.Date(2000, time.September, 13, 0, 0, 0, 0, time.UTC).Unix(),
		},
		{
			Period:     model.EndSpring.String(),
			PeriodDate: time.Date(2000, time.January, 27, 0, 0, 0, 0, time.UTC).Unix(),
		},
		{
			Period:     model.EndSummer.String(),
			PeriodDate: time.Date(2000, time.June, 15, 0, 0, 0, 0, time.UTC).Unix(),
		},
		{
			Period:     model.EndWinter.String(),
			PeriodDate: time.Date(2000, time.December, 22, 0, 0, 0, 0, time.UTC).Unix(),
		},
	},
	Metadata: []*model.Metadata{{
		Title: "About", Content: `The New Jersey Institute of Technology (NJIT) is a public research university in 
			the University Heights neighborhood of Newark, New Jersey. NJIT is New Jersey's Science & Technology University.
			 Centrally located in the New York metropolitan area its campus is within walking distance of downtown Newark.
			 New York City, 9 miles (14.5 km) and under 30 minutes away, is directly accessible from campus via public transit.
Founded in 1881 with the support of local industrialists and inventors, especially Edward Weston (334 US Patents), 
NJIT opened as Newark Technical School in 1884.[b] Application oriented from inception the school grew into a classic engineering college 
– Newark College of Engineering (NCE) – and then, with the addition of a School of Architecture in 1973, into a polytechnic university
that is now home to five colleges and one school. NJIT opened with 88 students.[c] As of fall 2015, the university enrolls over 11,300 students, 2,200 of whom live on campus.
 Architecturally significant buildings include Eberhardt Hall, the Campus Center, and the Central King Building – in the 
Collegiate Gothic style – which is being renovated into a STEM center. Facilities under construction include a Wellness and 
Events Center that will house a 3,500-seat venue for social and sporting events.`,
	},
	},
}
