package spike

import (
	"fmt"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
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
	form := url.Values{}
	form.Add("isSubscribed", "true")
	form.Add("fcmToken", "testToken")
	form.Add("topicName", "topicName")

	f, _ := os.Open("/Users/tevjef/Desktop/out.json")
	req := httptest.NewRequest("GET", "/v2/subscription/rutgers.universitycamden/fall/2020", f)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", "UniversityCourseTracker/1.0.8 (com.tevinjeffrey.uctios; build:100; iOS 10.3.1) Alamofire/4.5.1")

	rr := httptest.NewRecorder()
	Spike(rr, req)

	fmt.Println(rr.Body.String())
	if got := rr.Body.String(); got != "Complete" {
		t.Errorf("Test failed")
	}
}

func TestSpikeSubscription(t *testing.T) {
	form := &url.Values{}
	form.Add("isSubscribed", "true")
	form.Add("fcmToken", "testToken")
	form.Add("topicName", "topicName")

	req := httptest.NewRequest("POST", "/v2/subscription", strings.NewReader(form.Encode()))
	req.Header.Add("User-Agent", "UniversityCourseTracker/1.0.8 (com.tevinjeffrey.uctios; build:100; iOS 10.3.1) Alamofire/4.5.1")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	Spike(rr, req)

	fmt.Println(rr.Body.String())
	if got := rr.Body.String(); got != "Complete" {
		t.Errorf("Test failed")
	}
}
