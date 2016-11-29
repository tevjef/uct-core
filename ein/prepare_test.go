package main

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"testing"
)

func BenchmarkGetTypeSprint(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fmt.Fprintf(ioutil.Discard, "%T", MockDatabaseHandler{})
	}
}

func BenchmarkGetTypeReflect(b *testing.B) {
	t := reflect.TypeOf(MockDatabaseHandler{})
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t = reflect.TypeOf(MockDatabaseHandler{})
	}

	fmt.Println(t)
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
