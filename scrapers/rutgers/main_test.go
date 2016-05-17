package main

import (
	"bufio"
	"github.com/gogo/protobuf/proto"
	"github.com/pquerna/ffjson/ffjson"
	"io/ioutil"
	"log"
	"os"
	"testing"
	uct "uct/common"
)

var emptyUniversity = uct.University{}
var scrapedUni *uct.University
var university uct.University
var jsonBytes []byte
var protoBytes []byte
var err error

func TestMain(m *testing.M) {
	file, err := os.Open("json.out")
	uct.CheckError(err)
	input := bufio.NewReader(file)

	dec := ffjson.NewDecoder()
	if err := dec.DecodeReader(input, &university); err != nil {
		log.Fatal(err)
	}

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
	err := proto.MarshalText(os.Stdout, scrapedUni)
	uct.CheckError(err)

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
	uct.CheckError(err)
	//fmt.Println(string(data))
}

func TestMarshalProtoUniversity(t *testing.T) {
	_, err := proto.Marshal(&university)
	uct.CheckError(err)
	//fmt.Println(string(data))
}

func TestUnmarshalJsonUniversity(t *testing.T) {
	in, err := ioutil.ReadFile("json.out")
	if err != nil {
		log.Fatalln("Error reading file:", err)
	}
	school := &uct.University{}
	if err := ffjson.Unmarshal(in, school); err != nil {
		log.Fatalln("Failed to parse address book:", err)
	}
	//fmt.Println(school.String())
}

func TestUnmarshalProtoUniversity(t *testing.T) {
	in, err := ioutil.ReadFile("protobuf.out")
	if err != nil {
		log.Fatalln("Error reading file:", err)
	}
	school := &uct.University{}
	if err := proto.Unmarshal(in, school); err != nil {
		log.Fatalln("Failed to parse address book:", err)
	}
	//fmt.Println(school.String())
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
		if err = ffjson.Unmarshal(jsonBytes, &emptyUniversity); err != nil {
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

/*
λ benchcmp json_test.txt proto_test.txt
benchmark                          json ns/op    	protobuf ns/op      delta
BenchmarkMarshalUniversity-4       39004106      	7817502       		-79.96% (5x)
BenchmarkUnmarshalUniversity-4     134929520     	12652147      		-90.62% (10.6x)

benchmark                          json allocs/op   protobuf allocs/op  delta
BenchmarkMarshalUniversity-4       23            	1              		-95.65% (23x)
BenchmarkUnmarshalUniversity-4     298394        	124327         		-58.33% (2.4x)

benchmark                          json B/op   		protobuf B/op     	delta
BenchmarkMarshalUniversity-4       17833786     	2121735       		-88.10% (8.4x)
BenchmarkUnmarshalUniversity-4     14871633      	9250024       		-37.80% (1.6x)
*/
