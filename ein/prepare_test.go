package main

import ()
import (
	"fmt"
	"reflect"
	"testing"
)

func BenchmarkGetTypeSprint(b *testing.B) {
	b.StopTimer()

	toPrint := getType()
	fmt.Println(fmt.Sprintf("%T", toPrint))

	b.ReportAllocs()
	b.StartTimer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fmt.Sprintf("%T", getType())
	}
}

func BenchmarkGetTypeReflect(b *testing.B) {
	b.StopTimer()

	toPrint := getType()
	fmt.Println(reflect.TypeOf(toPrint))

	b.ReportAllocs()
	b.StartTimer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reflect.TypeOf(getType())
	}
}

func getType() interface{} {
	return MockDatabaseHandler{}
}

type MockDatabaseHandler struct {
}

func (dbHandler MockDatabaseHandler) insert(query string, data interface{}) (id int64) {
	return
}

func (dbHandler MockDatabaseHandler) update(query string, data interface{}) (id int64) {
	return
}

func (dbHandler MockDatabaseHandler) upsert(insertQuery, updateQuery string, data interface{}) (id int64) {
	return
}

func (dbHandler MockDatabaseHandler) exists(query string, data interface{}) (id int64) {
	return
}
