package redis

import (
	"github.com/tevjef/uct-core/common/conf"
	"github.com/tevjef/uct-core/common/model"

	"github.com/Sirupsen/logrus"
	redis "gopkg.in/redis.v5"
)

type Helper struct {
	NameSpace string
	Client    *redis.Client
}

const (
	BaseNamespace = "uct:"
	ScraperQueue  = BaseNamespace + "scraper:queue"
)

func nameSpaceForApp(appName string) string {
	return BaseNamespace + appName
}

func NewHelper(config conf.Config, appName string) *Helper {
	if client, err := model.OpenRedis(config.RedisAddr(), config.Redis.Password, config.Redis.Db); err != nil {
		logrus.WithError(err).Fatalln()
		return nil
	} else {
		return &Helper{
			NameSpace: nameSpaceForApp(appName),
			Client:    client,
		}
	}

}

func (r Helper) FindAll(key string) ([]string, error) {
	if keys, err := r.Client.Keys(key).Result(); err != nil {
		return nil, err
	} else {
		return keys, nil
	}
}

func (r Helper) Count(key string) (int64, error) {
	if keys, err := r.FindAll(key); err != nil {
		return -1, err
	} else {
		//log.WithFields(log.Fields{"key":key, "result": len(keys)}).Debugln("Count")
		return int64(len(keys)), nil
	}
}

func (r Helper) ListSize(list string) (int64, error) {
	if keys, err := r.Client.LRange(list, 0, -1).Result(); err != nil {
		return -1, err
	} else {
		//log.WithFields(log.Fields{"key":key, "result": len(keys)}).Debugln("ListSize")
		return int64(len(keys)), nil
	}
}

func (r Helper) RPush(list, key string) (int64, error) {
	if result, err := r.Client.RPush(list, key).Result(); err != nil {
		return -1, err
	} else {
		//log.WithFields(log.Fields{"result":result}).Debugln("RPush")
		return result, nil
	}
}

func (r Helper) RPushNotExist(list, key string) (int64, error) {
	if i, err := r.Exists(list, key); err != nil {
		//log.WithError(err).Panic("failed if exists on list", list)
	} else {
		if i >= 0 {
			return i, nil
		}
	}

	if result, err := r.Client.RPush(list, key).Result(); err != nil {
		return -1, err
	} else {
		//log.WithFields(log.Fields{"result":result}).Println("RPushNotExist")
		return result, nil
	}
}

func (r Helper) LPushNotExist(list, key string) (int64, error) {
	if i, err := r.Exists(list, key); err != nil {
		//log.WithError(err).Panic("failed if exists on list:", list)
	} else {
		if i >= 0 {
			return i, nil
		}
	}

	if result, err := r.Client.LPush(list, key).Result(); err != nil {
		return -1, err
	} else {
		//log.WithFields(log.Fields{"result":result}).Debugln("RPushNotExist")
		return result, nil
	}
}

func (r Helper) LPush(list, key string) (int64, error) {
	if result, err := r.Client.LPush(list, key).Result(); err != nil {
		return -1, err
	} else {
		//log.WithFields(log.Fields{"result":result}).Debugln("LPush")
		return result, nil
	}
}

func (r Helper) Exists(list, key string) (int64, error) {
	if result, err := r.Client.LRange(list, 0, -1).Result(); err != nil {
		return -1, err
	} else {
		//log.WithFields(log.Fields{"result":result}).Debugln("Exist")

		for i, val := range result {
			//log.WithFields(log.Fields{"val":val, "key":key}).Debugln("Exist test")
			if val == key {
				return int64(i), nil
			}
		}
	}

	return -1, nil
}
