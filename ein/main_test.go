package ein

import (
	"net/http/httptest"
	"os"
	"testing"
)

func TestEin(t *testing.T) {
	f, _ := os.Open("/Users/tevjef/Desktop/out.json")
	req := httptest.NewRequest("POST", "/", f)
	req.Header.Add("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	Ein(rr, req)

	if got := rr.Body.String(); got != "Complete" {
		t.Errorf("Test failed")
	}
}
