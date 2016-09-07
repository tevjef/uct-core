package v1

import (
	uct "uct/common"
	"testing"
)

func TestClientConnection(t *testing.T) {
	c := uct.Config{}
	c.Redis.Host = "localhost:32768"
	c.Redis.Db = 0
	c.Redis.Password = ""
	wrapper := New(c, "snanitycheck")
	if _, err := wrapper.LPush("My-new-list", "some shit"); err != nil {
		t.Error(err)
	}

}