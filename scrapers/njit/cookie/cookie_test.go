package cookie

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

func TestBakedCookie_Get(t *testing.T) {
	tests := []struct {
		name string
		b    *BakedCookie
		want http.Cookie
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		if got := tt.b.Get(); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. BakedCookie.Get() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestBakedCookie_SetValue(t *testing.T) {
	type args struct {
		value string
	}
	tests := []struct {
		name string
		b    *BakedCookie
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		tt.b.SetValue(tt.args.value)
	}
}

/*func TestNew(t *testing.T) {

	var queueSize = 10
	var cookies []*http.Cookie
	for i := 0; i < queueSize; i++ {
		cookies = append(cookies, &http.Cookie{
			Name:   "JSESSIONID",
			Path:   "/StudentRegistrationSsb/",
			Domain: "myhub.njit.edu",
		})
	}

	cc := New(cookies, func(bc *BakedCookie) error {
		bc.SetValue(prepareCookie(t, "201690"))
		log.Debugln("initializing cookie", bc.name)
		return nil
	})

	for i := 0; i < 2; i++ {
		go func() {
			bc := cc.Pop(nil)

			time.Sleep(time.Second * time.Duration(rand.Intn(4)))

			cc.Push(bc, func(baked *BakedCookie) error {
				resetCookie(t, *baked.Get())
				return nil
			})
		}()

	}

	time.Sleep(time.Minute)
}*/

func resetCookie(t *testing.T, cookie http.Cookie) {
	req, _ := http.NewRequest(http.MethodPost, "https://myhub.njit.edu/StudentRegistrationSsb/ssb/classSearch/classSearch", strings.NewReader(url.Values{}.Encode()))
	req.AddCookie(&cookie)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	_, err := httpClient.Do(req)

	if err != nil {
		t.Fatalf("Failed to validate cookie err: %s", err.Error())
	}

	log.Debugln("Resetting cookie", cookie.Value, cookie.Domain)
}

func prepareCookie(t *testing.T, term string) string {
	resp, err := httpClient.PostForm("https://myhub.njit.edu/StudentRegistrationSsb/ssb/term/search?mode=search", url.Values{"term": []string{term}})
	if err != nil {
		t.Fatalf("Failed get cookie: err %s", err.Error())
	}

	if len(resp.Cookies()) > 0 {
		cookie := resp.Cookies()[0]
		req, _ := http.NewRequest(http.MethodGet, "https://myhub.njit.edu/StudentRegistrationSsb/ssb/classSearch/classSearch", nil)
		req.AddCookie(cookie)
		_, err := httpClient.Do(req)

		if err != nil {
			t.Fatalf("Failed to validate: err %s", err.Error())
		}
		return cookie.Value
	}

	return ""
}

func TestCookieCutter_Push(t *testing.T) {
	type args struct {
		baked  *BakedCookie
		onPush func(baked *BakedCookie) error
	}
	tests := []struct {
		name string
		cc   *CookieCutter
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		tt.cc.Push(tt.args.baked, tt.args.onPush)
	}
}

var httpClient = &http.Client{
	Timeout:   15 * time.Second,
	Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
}
