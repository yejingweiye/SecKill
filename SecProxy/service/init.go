package service

import (
	"github.com/astaxie/beego/logs"
	"github.com/garyburd/redigo/redis"
	"strconv"
	"time"
)

func InitService(serviceConf *SecKillConf) (err error) { // 拿到main 包的全局变量
	secKillConf = serviceConf
	err = loadBlackList()
	if err != nil {
		logs.Error("load blacklist failed,err:[%v]", err)
		return
	}
	logs.Debug("init service succ,conf:%v", secKillConf)
	err = initProxy2LayerRedis()
	if err != nil {
		logs.Error("initProxy2LayerRedis failed,err:[%v]", err)
		return
	}
	secKillConf.SecReqChan = make(chan *SecRequest, secKillConf.SecReqChanSize)
	initRedisProcessFunc()
	if err != nil {
		logs.Error(" initRedisProcessFunc failed,err:[%v]", err)
		return
	}
	return
}

func initRedisProcessFunc() { // 16goroutine读写redis
	for i := 0; i < secKillConf.WriteProxy2layerGoroutineNum; i++ {
		go WriteHandler()
	}
	for i := 0; i < secKillConf.ReadLayer2proxyGoroutineNum; i++ {
		go ReadHandler()
	}
	return
}
func initProxy2LayerRedis() (err error) {
	secKillConf.Proxy2LayerRedisPool = &redis.Pool{
		MaxIdle:     secKillConf.RedisProxy2LayerConf.RedisMaxIdle,
		MaxActive:   secKillConf.RedisProxy2LayerConf.RedisMaxActive,
		IdleTimeout: time.Duration(secKillConf.RedisProxy2LayerConf.RedisIdleTimeout) * time.Second, // 300S
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", secKillConf.RedisProxy2LayerConf.RedisAddr)
		},
	}
	conn := secKillConf.Proxy2LayerRedisPool.Get()
	defer conn.Close()
	_, err = conn.Do("ping")
	if err != nil {
		logs.Error("ping redis failed, err:%v", err)
		return
	}

	return
}

func initBlackRedis() (err error) {
	secKillConf.BlackRedisPool = &redis.Pool{
		MaxIdle:     secKillConf.RedisBlackConf.RedisMaxIdle,
		MaxActive:   secKillConf.RedisBlackConf.RedisMaxActive,
		IdleTimeout: time.Duration(secKillConf.RedisBlackConf.RedisIdleTimeout) * time.Second, // 300S
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", secKillConf.RedisBlackConf.RedisAddr)
		},
	}
	conn := secKillConf.BlackRedisPool.Get()
	defer conn.Close()
	_, err = conn.Do("ping")
	if err != nil {
		logs.Error("ping redis failed, err:%v", err)
		return
	}

	return
}

func loadBlackList() (err error) {
	err = initBlackRedis()
	if err != nil {
		logs.Error("init black redis failed,err:%v", err)
		return
	}
	conn := secKillConf.BlackRedisPool.Get()
	defer conn.Close()
	reply, err := conn.Do("hgetall", "idblacklist")
	idList, err := redis.Strings(reply, err)
	if err != nil {
		logs.Warn("hgetall failed,err:%v", err)
		return
	}

	for _, v := range idList {
		id, err := strconv.Atoi(v)
		if err != nil {
			logs.Warn("invalid user id[%d]", id)
			continue
		}
		secKillConf.IdBlackMap[id] = true
	}

	reply, err = conn.Do("hgetall", "ipblacklist")
	ipList, err := redis.Strings(reply, err)
	if err != nil {
		logs.Warn("hgetall failed,err:%v", err)
		return
	}

	for _, v := range ipList {
		secKillConf.IpBlackMap[v] = true
	}

	go SyncIpBlackLit()
	go SyncIdBlackList()
	return
}

func SyncIdBlackList() {
	var idList []int
	lastTime := time.Now().Unix()
	for {
		conn := secKillConf.BlackRedisPool.Get()
		defer conn.Close()
		reply, err := conn.Do("BLPOP", "blackidlist", time.Second)
		id, err := redis.Int(reply, err)
		if err != nil {
			continue
		}
		idList = append(idList, id)
		curTime := time.Now().Unix()
		if len(idList) > 100 || curTime-lastTime > 5 {
			secKillConf.RWBlackLock.Lock()
			for _, v := range idList {
				secKillConf.IdBlackMap[v] = true
			}
			secKillConf.RWBlackLock.Unlock()
		}

	}
}

func SyncIpBlackLit() {
	var ipList []string
	lastTime := time.Now().Unix()
	for {
		conn := secKillConf.BlackRedisPool.Get()
		defer conn.Close()
		reply, err := conn.Do("BLPOP", "blackiplist", time.Second) // 没有就阻塞
		ip, err := redis.String(reply, err)
		if err != nil {
			continue
		}
		curTime := time.Now().Unix()
		ipList = append(ipList, ip)
		if len(ipList) > 100 || curTime-lastTime > 5 { // 大于100次才更新，减少加锁的频率或者时间大于5秒
			secKillConf.RWBlackLock.Lock()
			for _, v := range ipList {
				secKillConf.IpBlackMap[v] = true
			}
			secKillConf.RWBlackLock.Unlock()
			lastTime = curTime
			logs.Info("sync ip list from redis succ,ip[%v]", ipList)
		}

	}
}
