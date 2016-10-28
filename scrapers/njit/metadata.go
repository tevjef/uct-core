package main

import (
	"uct/common/model"
	"time"
)

var njit = model.University{
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
			Title: "About", Content: "",
		},
		},
	}
