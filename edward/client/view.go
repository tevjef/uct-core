package client

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/tevjef/uct-backend/common/model"
)

func (c *Client) ListSubscriptionView(topicName string) ([]*model.SubscriptionView, error) {
	rel := &url.URL{Path: "/v1/hotness/" + topicName}
	u := c.BaseURL.ResolveReference(rel)
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/x-protobuf")
	req.Header.Set("User-Agent", c.UserAgent)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	response := model.Response{}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	log.Println(string(b))
	err = response.Unmarshal(b)
	if err != nil {
		return nil, err
	}

	return response.Data.SubscriptionView, err
}
