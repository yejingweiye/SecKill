package service

import (
	"github.com/garyburd/redigo/redis"
	"sync"
	"time"
)

const (
	ProductStatusNormal       = 0
	ProductStatusSaleOut      = 1
	ProductStatusForceSaleOut = 2
)

type RedisConf struct {
	RedisAddr        string
	RedisMaxIdle     int
	RedisMaxActive   int
	RedisIdleTimeout int
}

type EtcdConf struct {
	EtcdAddr          string
	EtcdTimeout       int
	EtcdSecKeyPrefix  string
	EtcdSecProductKey string
}

type LogConf struct {
	LogPath  string
	LogLevel string
}

type SecKillConf struct {
	RedisBlackConf       RedisConf
	RedisProxy2LayerConf RedisConf
	EtcdConf             EtcdConf
	LogConf              LogConf
	SecProductInfoMap    map[int]*SecProductInfoConf
	RWSecProductLock     sync.RWMutex
	CookieSecretKey      string
	UserSecAccessLimit   int
	ReferWhiteList       []string
	IpSecAccessLimit     int
	IpBlackMap           map[string]bool
	IdBlackMap           map[int]bool

	BlackRedisPool               *redis.Pool
	Proxy2LayerRedisPool         *redis.Pool
	secLimitMgr                  *SecLimitMgr
	RWBlackLock                  sync.RWMutex
	WriteProxy2layerGoroutineNum int
	ReadLayer2proxyGoroutineNum  int

	SecReqChan     chan *SecRequest
	SecReqChanSize int
}

type SecProductInfoConf struct {
	ProductId int
	StartTime int64
	EndTime   int64
	Status    int
	Total     int
	Left      int // 剩余
}

type SecRequest struct {
	ProductId     int
	Source        string
	Authcode      string
	SecTime       string
	Nance         string
	UserId        int
	UserAuthSign  string
	AccessTime    time.Time
	ClientAddr    string
	ClientRefence string //  描述：表示在a连接点连接skill请求。F12，则代表是 a连接
}
