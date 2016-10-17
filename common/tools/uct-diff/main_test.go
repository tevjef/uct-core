package main

import (
	"bufio"
	"flag"
	"github.com/pquerna/ffjson/ffjson"
	"io/ioutil"
	"log"
	"os"
	"testing"
	"uct/common/model"
)

var uni1 model.University
var uni2 model.University

func TestMain(m *testing.M) {
	flag.Parse()

	uni1 = openFile("C:\\Users\\Tevin\\Desktop\\Development\\Go\\src\\uct\\common\\diff\\old.out")
	uni2 = openFile("C:\\Users\\Tevin\\Desktop\\Development\\Go\\src\\uct\\common\\diff\\new.out")

	os.Exit(m.Run())
}

func TestDiffUniversity(t *testing.T) {
	funi := diffAndFilter(uni1, uni2)

	data, err := ffjson.Marshal(funi)
	model.CheckError(err)

	ioutil.WriteFile("filtered.out", data, 0644)
}

func openFile(filePath string) (uni model.University) {
	file, err := os.Open(filePath)
	model.CheckError(err)

	input := bufio.NewReader(file)
	dec := ffjson.NewDecoder()
	if err := dec.DecodeReader(input, &uni); err != nil {
		log.Fatal(err)
	}
	return
}
