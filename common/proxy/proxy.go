package proxy

import (
	log "github.com/Sirupsen/logrus"
	"net/http"
	"net/url"
	"os"
	"math/rand"
)

type proxyConfig struct {
	parsedUrl *url.URL
	user      string
	pass      string
	basic     string
}

type ProxyClient struct {
	config proxyConfig
	*http.Client
}

var proxyUrl = os.Getenv("HTTP_PROXY_URL")
var proxyUser = os.Getenv("HTTP_PROXY_USER")
var proxyPass = os.Getenv("HTTP_PROXY_PASS")

var hardProxy = []string{"http://192.225.174.195:60099",
	"http://192.225.174.197:60099",
	"http://192.225.174.198:60099",
	"http://192.225.174.200:60099",
	"http://192.225.174.201:60099",
	"http://192.225.174.202:60099",
	"http://192.225.174.204:60099",
	"http://192.225.174.206:60099",
	"http://192.225.174.207:60099",
	"http://192.225.174.210:60099",
	"http://192.225.174.211:60099",
	"http://192.225.174.212:60099",
	"http://192.225.174.213:60099",
	"http://192.225.174.215:60099",
	"http://192.225.174.219:60099",
	"http://192.225.174.226:60099",
	"http://192.225.174.227:60099",
	"http://192.225.174.234:60099",
	"http://192.225.174.236:60099",
	"http://192.225.174.237:60099",
	"http://192.225.174.241:60099",
	"http://192.225.174.242:60099",
	"http://192.225.174.243:60099",
	"http://192.225.174.247:60099",
	"http://192.225.174.248:60099"}

func GetProxyUrl() func(*http.Request) (*url.URL, error) {
	var proxy proxyConfig
	proxy.user = proxyUser
	proxy.pass = proxyPass

	if proxyURL, err := url.Parse(hardProxy[rand.Intn(len(hardProxy)-1)]); err != nil {
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
