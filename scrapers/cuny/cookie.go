package main

import (
	"net/http"
	"net/url"
)

var portalUrl, _ = url.Parse("hrsa.cunyfirst.cuny.edu")

func saveCookie(client *http.Client, response http.Response) {
	client.Jar.SetCookies(portalUrl, response.Cookies())
}
