package proxy

import (
	"net/http"
	"net/url"
	"os"

	log "github.com/sirupsen/logrus"
)

var proxyUrl = os.Getenv("HTTP_PROXY_URL")
var proxyUser = os.Getenv("HTTP_PROXY_USER")
var proxyPass = os.Getenv("HTTP_PROXY_PASS")

func ProxyUrl() func(*http.Request) (*url.URL, error) {
	if proxyUrl == "" {
		return nil
	}

	if proxyURL, err := url.Parse(proxyUrl); err != nil {
		log.Fatalln(err)
	} else {
		if proxyUser != "" && proxyPass != "" {
			proxyURL.User = url.UserPassword(proxyUser, proxyPass)
		}
		return http.ProxyURL(proxyURL)
	}

	return nil
}
