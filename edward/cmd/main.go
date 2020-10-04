package main

import (
	"net/http"
	"net/url"

	"github.com/sirupsen/logrus"
	"github.com/tevjef/uct-backend/edward/client"
)

func main() {
	url, _ := url.Parse("http://localhost:2058")

	client := client.Client{
		BaseURL:    url,
		UserAgent:  "localhost",
		HttpClient: http.DefaultClient,
	}

	view, err := client.ListSubscriptionView("rutgers.universitynewark.510.history.general.fall.2019.201.history.of.western.civilization.i")
	logrus.WithError(err).Fatal(view)
}
