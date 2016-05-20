package main

import (
	"testing"
	"flag"
	"os"
	"bufio"
	"github.com/pquerna/ffjson/ffjson"
	"log"
	uct "uct/common"
	"io/ioutil"
)

var uni1 uct.University
var uni2 uct.University

func TestMain(m *testing.M) {
	flag.Parse()

	uni1 = openFile("C:\\Users\\Tevin\\Desktop\\Development\\Go\\src\\uct\\common\\diff\\old.out")
	uni2 = openFile("C:\\Users\\Tevin\\Desktop\\Development\\Go\\src\\uct\\common\\diff\\new.out")

	os.Exit(m.Run())
}

func TestDiffUniversity(t *testing.T) {
	funi := diffAndFilter(uni1, uni2)

	data, err := ffjson.Marshal(funi)
	uct.CheckError(err)

	ioutil.WriteFile("filtered.out", data, 0644)
}

func openFile(filePath string) (uni uct.University) {
	file, err := os.Open(filePath)
	uct.CheckError(err)

	input := bufio.NewReader(file)
	dec := ffjson.NewDecoder()
	if err := dec.DecodeReader(input, &uni); err != nil {
		log.Fatal(err)
	}
	return
}
