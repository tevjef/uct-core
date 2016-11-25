package main

import (
	_ "net/http/pprof"
	"testing"
	"uct/common/model"

	_ "github.com/lib/pq"
	"log"
	"uct/common/deepcopier"
	"time"
)

func Test_diffAndMergeCourses(t *testing.T) {
	var u1 model.University
	var u2 model.University

	u1 = generateUni()

	deepcopier.Copy(&u2).From(&u1)

	//a := diffAndMergeCourses(generateUni(), generateUni2())
	//b := diffAndMergeCourses(generateUni(), generateUni())
	log.Printf("u1 %+v\n", u1)
	log.Printf("u2 %+v\n", u2)

	//log.Printf("%+v\n", b)
	time.Sleep(time.Second)
}

func generateUni2() model.University {
	return model.University{
		Subjects: []*model.Subject{
			{Courses: []*model.Course{
				{TopicName: "e"},
				{TopicName: "f"},
			}},
		},
	}
}

func generateUni() model.University {
	return model.University{
		Subjects: []*model.Subject{
			{Courses: []*model.Course{
				{TopicName: "a"},
				{TopicName: "b"},
				{TopicName: "c"},
				{TopicName: "d"},
				{TopicName: "e", Sections: []*model.Section{
					{TopicName:"e1"},
					{TopicName:"e2"},
				}},
				{TopicName: "f", Sections: []*model.Section{
					{TopicName:"f1"},
					{TopicName:"f2"},
				}},
			}},
		},
	}
}