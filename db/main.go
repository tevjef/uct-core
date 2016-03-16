package db

import (
	"encoding/json"
	"os"
	"io"
	"log"

	"uct-common"
)

func main() {
	decoder := json.NewDecoder(os.Stdin)

	var university uct.University

	for {
		if err := decoder.Decode(&university); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
	}
}
