package main

import (
	"github.com/gogo/protobuf/proto"
	"github.com/pquerna/ffjson/ffjson"
	"io/ioutil"
	"log"
	"os"
	"testing"
	"uct/common/model"
)

var emptyUniversity = model.University{}
var scrapedUni *model.University
var university model.University
var jsonBytes []byte
var protoBytes []byte
var err error

func TestMain(m *testing.M) {
	if _, err := os.Stat("json.out"); os.IsNotExist(err) {
		writeJsonData()
	}
	if _, err := os.Stat("protobuf.out"); os.IsNotExist(err) {
		writeProtoData()
	}
	jsonBytes, err = ioutil.ReadFile("json.out")
	if err != nil {
		log.Fatalln("Error reading file:", err)
	}

	protoBytes, err = ioutil.ReadFile("protobuf.out")
	if err != nil {
		log.Fatalln("Error reading file:", err)
	}

	os.Exit(m.Run())
}

func writeJsonData() []byte {
	if scrapedUni == nil {
		t := getCampus("CM")
		scrapedUni = &t
	}
	data, err := ffjson.Marshal(scrapedUni)
	if err != nil {
		log.Fatalln("Failed to encode university:", err)
	}
	if err := ioutil.WriteFile("json.out", data, 0644); err != nil {
		log.Fatalln("Failed to write university:", err)
	}
	return data
}

func writeProtoData() []byte {
	if scrapedUni == nil {
		t := getCampus("CM")
		scrapedUni = &t
	}
	out, err := proto.Marshal(scrapedUni)
	if err != nil {
		log.Fatalln("Failed to encode university:", err)
	}
	if err := ioutil.WriteFile("protobuf.out", out, 0644); err != nil {
		log.Fatalln("Failed to write university:", err)
	}
	return out
}

func TestMarshalJsonUniversity(t *testing.T) {
	_, err := ffjson.Marshal(university)
	model.CheckError(err)
	//fmt.Println(string(data))
}

func TestMarshalProtoUniversity(t *testing.T) {
	_, err := proto.Marshal(&university)
	model.CheckError(err)
	//fmt.Println(string(data))
}

func TestUnmarshalJsonUniversity(t *testing.T) {
	in, err := ioutil.ReadFile("json.out")
	if err != nil {
		log.Fatalln("Error reading file:", err)
	}
	school := &model.University{}
	if err := ffjson.UnmarshalFast(in, school); err != nil {
		log.Fatalln("Failed to parse university:", err)
	}
	//fmt.Println(school.String())
}

func TestUnmarshalProtoUniversity(t *testing.T) {
	in, err := ioutil.ReadFile("protobuf.out")
	if err != nil {
		log.Fatalln("Error reading file:", err)
	}
	school := &model.University{}
	if err := proto.Unmarshal(in, school); err != nil {
		log.Fatalln("Failed to parse address book:", err)
	}
	//fmt.Println(school.String())
}

func TestUnmarshalProtoEqualUniversity(t *testing.T) {
	in, err := ioutil.ReadFile("json.out")
	if err != nil {
		log.Fatalln("Error reading file:", err)
	}
	s := &model.University{}
	if err := ffjson.UnmarshalFast(in, s); err != nil {

	}

	in, err = ioutil.ReadFile("protobuf.out")
	if err != nil {
		log.Fatalln("Error reading file:", err)
	}
	school := &model.University{}
	if err := proto.Unmarshal(in, school); err != nil {
		log.Fatalln("Failed to parse university:", err)
	}

	err = school.VerboseEqual(s)

	//fmt.Println(school.GoString())
	//school.GoString()

	model.CheckError(err)
}

func BenchmarkMarshalJsonUniversity(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ffjson.Marshal(university)
	}
}

func BenchmarkMarshalProtoUniversity(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		proto.Marshal(&university)
	}
}

func BenchmarkUnmarshalJsonUniversity(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err = ffjson.UnmarshalFast(jsonBytes, &emptyUniversity); err != nil {
			log.Fatalln("Failed to parse university:", err)
		}
		emptyUniversity.Reset()
	}
}

func BenchmarkUnmarshalProtoUniversity(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err = proto.Unmarshal(protoBytes, &emptyUniversity); err != nil {
			log.Fatalln("Failed to parse university:", err)
		}
		emptyUniversity.Reset()
	}
}
