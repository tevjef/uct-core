package main

import (
	"fmt"
	"testing"

	"bytes"
)

func TestFo(t *testing.T) {
	str := "\u0000\u00015\u0000\u00011\u0000\u00011\u0000\u0001N\u0000 \u0000\u0001N\u0000"
	str = string(bytes.Replace([]byte(str), []byte("\x00"), []byte(""), -1))
	str = string(bytes.Replace([]byte(str), []byte("\x01"), []byte(""), -1))

	/*
		sub := common.Section{Number: "Tevin Jeffrey", Status:common.OPEN.String()}
		out, _ := json.Marshal(sub)*/
	fmt.Println(string(str))
	fmt.Printf("%x", string(str))

}
