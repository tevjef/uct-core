package proxy

import (
	"net/http"
	"net/url"
	log "github.com/Sirupsen/logrus"
	"os"
)

type proxyConfig struct {
	parsedUrl *url.URL
	user string
	pass string
	basic string
}

type ProxyClient struct {
	config proxyConfig
	*http.Client
}

var proxyUrl = os.Getenv("HTTP_PROXY_URL")
var proxyUser = os.Getenv("HTTP_PROXY_USER")
var proxyPass = os.Getenv("HTTP_PROXY_PASS")

func GetProxyUrl() func(*http.Request) (*url.URL, error) {
	var proxy proxyConfig
	proxy.user = proxyUser
	proxy.pass = proxyPass

	if proxyURL, err := url.Parse(proxyUrl); err != nil {
		log.Fatalln(err)
	} else {
		if proxyUser != "" && proxyPass != "" {
			userInfo := url.UserPassword(proxyUser, proxyPass)
			proxyURL.User = userInfo
		}
		proxy.parsedUrl = proxyURL
	}

	return http.ProxyURL(proxy.parsedUrl)
}