package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
	MysqlConfig       MysqlConfig
	Auth              Auth
	RedisConfig       redis.RedisConf
	AdminInviteCode   string
	UserInviteCode    string
	RedisExpire       int64
	RedisLockExpire   int64
	RedisOrderPageNum int
	PageSize          int
	Path              string
}

type MysqlConfig struct {
	DataSource     string
	ConnectTimeout int64
}

type Auth struct {
	AccessSecret string
	Expire       int64
}
