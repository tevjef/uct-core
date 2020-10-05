package spike

import (
	"fmt"
	"net/http/httptest"
	"os"
	"testing"
)

func TestSpike(t *testing.T) {
	f, _ := os.Open("/Users/tevjef/Desktop/out.json")
	req := httptest.NewRequest("GET", "/v2/universities", f)
	req.Header.Add("Accept", "application/json")

	rr := httptest.NewRecorder()
	Spike(rr, req)

	fmt.Println(rr.Body.String())
	if got := rr.Body.String(); got != "Complete" {
		t.Errorf("Test failed")
	}
}

func TestSpikeUniversity(t *testing.T) {
	f, _ := os.Open("/Users/tevjef/Desktop/out.json")
	req := httptest.NewRequest("GET", "/v2/subjects/rutgers.universitycamden/fall/2020", f)
	req.Header.Add("Accept", "application/json")

	rr := httptest.NewRecorder()
	Spike(rr, req)

	fmt.Println(rr.Body.String())
	if got := rr.Body.String(); got != "Complete" {
		t.Errorf("Test failed")
	}
}
