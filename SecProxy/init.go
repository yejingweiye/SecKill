package main

import (
	"SecKill/service"
	//_"golang.org/x/net/context"
	"context"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	etcd_client "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/garyburd/redigo/redis"
	"time"
)

var (
	redisPool  *redis.Pool
	etcdClient *etcd_client.Client
)

func initRedis() (err error) {
	redisPool = &redis.Pool{
		MaxIdle:     secKillConf.RedisBlackConf.RedisMaxIdle,
		MaxActive:   secKillConf.RedisBlackConf.RedisMaxActive,
		IdleTimeout: time.Duration(secKillConf.RedisBlackConf.RedisIdleTimeout) * time.Second, // 300S
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", secKillConf.RedisBlackConf.RedisAddr)
		},
	}
	conn := redisPool.Get()
	defer conn.Close()
	_, err = conn.Do("ping")
	if err != nil {
		logs.Error("ping redis failed, err:%v", err)
		return
	}

	return
}

func initEtcd() (err error) {
	cli, err := etcd_client.New(etcd_client.Config{
		Endpoints:   []string{secKillConf.EtcdConf.EtcdAddr},
		DialTimeout: time.Duration(secKillConf.EtcdConf.EtcdTimeout) * time.Second,
	})
	if err != nil {
		logs.Error("connect etsc failed,err:", err)
		return
	}
	etcdClient = cli
	return
}

//日志字符串与整型转换
func convertLogLevel(level string) int {
	switch level {
	case "debug":
		return logs.LevelDebug
	case "warn":
		return logs.LevelWarn
	case "info":
		return logs.LevelInfo
	case "trace":
		return logs.LevelTrace
	}
	return logs.LevelDebug

}

func initLog() (err error) {
	config := make(map[string]interface{})
	config["filename"] = secKillConf.LogConf.LogPath
	config["level"] = convertLogLevel(secKillConf.LogConf.LogLevel) // 该值是个int型，需要转换

	configStr, err := json.Marshal(config)
	if err != nil {
		fmt.Println("Marshal failed,err:%v", err)
		return
	}
	logs.SetLogger(logs.AdapterFile, string(configStr))
	return
}

// 从etcd读取
func loadSecConf() (err error) {
	resp, err := etcdClient.Get(context.Background(), secKillConf.EtcdConf.EtcdSecProductKey)
	if err != nil {
		logs.Error("get [%s] from etcd failed,err:%v", secKillConf.EtcdConf.EtcdSecProductKey, err)
		return
	}

	var secProductInfo []service.SecProductInfoConf
	for k, v := range resp.Kvs {
		logs.Debug("key [%v], value [%v]", k, v)
		err = json.Unmarshal(v.Value, &secProductInfo)
		if err != nil {
			logs.Error("Unmarshal sec product info failed,err: %v ", err)
			return
		}
		logs.Debug("sec info conf is [%v]", secProductInfo)
	}
	updateSecProductInfo(secProductInfo) // 更新数值
	return
}

func initSecProductWatcher() {
	go watchSecproductKey(secKillConf.EtcdConf.EtcdSecProductKey)
}

// 监听etcd key 的变化
func watchSecproductKey(key string) {
	cli, err := etcd_client.New(etcd_client.Config{
		Endpoints:   []string{secKillConf.EtcdConf.EtcdAddr},
		DialTimeout: time.Duration(secKillConf.EtcdConf.EtcdTimeout) * time.Second,
	})
	if err != nil {
		logs.Error("connect etcd failed,err:", err)
		return
	}

	logs.Debug("begin watch key;%s", key)
	for {
		rch := cli.Watch(context.Background(), key) // 监听key
		var secProductInfo []service.SecProductInfoConf
		var getConfSucc = true

		for wresp := range rch {
			for _, ev := range wresp.Events {
				if ev.Type == mvccpb.DELETE {
					logs.Warn("key[%s] `s config deleted", key)
					continue
				}
				if ev.Type == mvccpb.PUT && string(ev.Kv.Key) == key {
					err = json.Unmarshal(ev.Kv.Value, &secProductInfo)
					if err != nil {
						logs.Error("key [%s],Unmarshal[%s],err;%v ", err)
						getConfSucc = false
						continue
					}
				}
				logs.Debug("get config from etcd,%s %q : %q\n ", ev.Type, ev.Kv.Key, ev.Kv.Value)
			}
			if getConfSucc {
				logs.Debug("get config fron etcd succ,%v", secProductInfo)
				updateSecProductInfo(secProductInfo) // 更新数值
			}
		}
	}
}

func updateSecProductInfo(secProductInfo []service.SecProductInfoConf) {
	var tmp map[int]*service.SecProductInfoConf = make(map[int]*service.SecProductInfoConf, 1024)
	for _, v := range secProductInfo {
		productInfo := v
		tmp[v.ProductId] = &productInfo
	}
	secKillConf.RWSecProductLock.Lock()
	secKillConf.SecProductInfoMap = tmp
	secKillConf.RWSecProductLock.Unlock()
}

// 初始化
func initSec() (err error) {

	err = initLog()
	if err != nil {
		logs.Error("init log failed, err: %v", err)
		return
	}

	err = initRedis()
	if err != nil {
		logs.Error("init redis failed, err: %v", err)
		return
	}

	err = initEtcd()
	if err != nil {
		logs.Error("init etcd failed, err: %v", err)
		return
	}

	err = loadSecConf()
	if err != nil {
		logs.Error("load secConfig failed,err:%v", err)
		return
	}

	service.InitService(secKillConf)
	initSecProductWatcher()
	logs.Info("init sec success")
	return
}
