package redis

import (
	"testing"

	"github.com/tevjef/uct-backend/common/conf"
)

func TestClientConnection(t *testing.T) {
	c := conf.Config{}
	c.Redis.Host = "localhost:32768"
	c.Redis.Db = 0
	c.Redis.Password = ""
	wrapper := NewHelper(c, "snanitycheck")
	if _, err := wrapper.LPush("My-new-list", "some shit"); err != nil {
		t.Error(err)
	}

}
