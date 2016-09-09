package v1

import (
	"gopkg.in/redis.v4"
	uct "uct/common"
	//log "github.com/Sirupsen/logrus"
)

type RedisWrapper struct {
	NameSpace string
	Client *redis.Client
}

const (
	BaseNamespace = "uct:scraper"
)

func nameSpaceForApp(appName string) string {
	return BaseNamespace + ":" + appName
}

func New(config uct.Config, appName string) *RedisWrapper {
	return &RedisWrapper{
		NameSpace: nameSpaceForApp(appName),
		Client: redis.NewClient(&redis.Options{
			Addr:    config.Redis.Host + ":" + config.Redis.Port,
			Password: config.Redis.Password,
			DB:       config.Redis.Db}),
	}
}

func (r RedisWrapper) Find(key string) ([]string, error) {
	if keys, err := r.Client.Keys(key).Result(); err != nil {
		return nil, err
	} else {
		//log.WithFields(log.Fields{"key":key, "result": keys}).Debugln("Find")
		return keys, nil
	}
}

func (r RedisWrapper) Count(key string) (int64, error) {
	if keys, err := r.Find(key); err != nil {
		return -1, err
	} else {
		//log.WithFields(log.Fields{"key":key, "result": len(keys)}).Debugln("Count")
		return int64(len(keys)), nil
	}
}

func (r RedisWrapper) RPush(list, key string) (int64, error) {
	if result, err := r.Client.RPush(list, key).Result(); err != nil {
		return -1, err
	} else {
		//log.WithFields(log.Fields{"result":result}).Debugln("RPush")
		return result, nil
	}
}

func (r RedisWrapper) RPushNotExist(list, key string) (int64, error) {
	if i, err := r.Exists(list, key); err != nil {
		//log.WithError(err).Panic("failed if exists on list:", list)
	} else {
		if i >= 0 {
			return i, nil
		}
	}

	if result, err := r.Client.RPush(list, key).Result(); err != nil {
		return -1, err
	} else {
		//log.WithFields(log.Fields{"result":result}).Debugln("RPushNotExist")
		return result, nil
	}
}

func (r RedisWrapper) LPushNotExist(list, key string) (int64, error) {
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

func (r RedisWrapper) LPush(list, key string) (int64, error) {
	if result, err := r.Client.LPush(list, key).Result(); err != nil {
		return -1, err
	} else {
		//log.WithFields(log.Fields{"result":result}).Debugln("LPush")
		return result, nil
	}
}

func (r RedisWrapper) Exists(list, key string) (int64, error) {
	if result, err := r.Client.LRange(list, 0, -1).Result(); err != nil {
		return -1, err
	} else {
		//log.WithFields(log.Fields{"result":result}).Debugln("Exist")

		for i, val := range result {
			//log.WithFields(log.Fields{"val":val, "key":key}).Debugln("Exist test")
			if (val == key) {
				return int64(i), nil
			}
		}
	}

	return -1, nil
}