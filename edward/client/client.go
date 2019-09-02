package client

import (
	"net/http"
	"net/url"
)

type Client struct {
	BaseURL   *url.URL
	UserAgent string

	HttpClient *http.Client
}
