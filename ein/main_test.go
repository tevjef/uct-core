package ein

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEin(t *testing.T) {
	req := httptest.NewRequest("GET", "/", strings.NewReader("{}"))
	req.Header.Add("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	Ein(rr, req)

	if got := rr.Body.String(); got != "Complete" {
		t.Errorf("Test failed")
	}
}
