package main

import (
	"SecKill/service"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"strings"
)

var (
	secKillConf = &service.SecKillConf{
		SecProductInfoMap: make(map[int]*service.SecProductInfoConf, 1024),
	}
)

// 读取配置信息
func initConfig() (err error) {
	redisBlackAddr := beego.AppConfig.String("redis_black_addr")
	etcdAddr := beego.AppConfig.String("etcd_addr")
	logs.Debug("redis config addr :%v", redisBlackAddr)
	logs.Debug("etcd config addr :%v", etcdAddr)
	secKillConf.RedisBlackConf.RedisAddr = redisBlackAddr
	secKillConf.EtcdConf.EtcdAddr = etcdAddr
	if len(redisBlackAddr) == 0 || len(etcdAddr) == 0 {
		err = fmt.Errorf("init config failed ,redis[%s] or etcd[%s] conf is null", redisBlackAddr, etcdAddr)
		return
	}
	redisMaxIdle, err := beego.AppConfig.Int("redis_black_idle")
	if err != nil {
		err = fmt.Errorf("init config failed ,read redis_black_idle error:%v", err)
	}
	redisMaxActive, err := beego.AppConfig.Int("redis_black_active")
	if err != nil {
		err = fmt.Errorf("init config failed ,read redis_black_active error:%v", err)
	}
	redisIdleTimeout, err := beego.AppConfig.Int("redis_black_idle_timeout")
	if err != nil {
		err = fmt.Errorf("init config failed ,read redis_black_idle_timeout error:%v", err)
	}

	secKillConf.RedisBlackConf.RedisMaxIdle = redisMaxIdle
	secKillConf.RedisBlackConf.RedisMaxActive = redisMaxActive
	secKillConf.RedisBlackConf.RedisIdleTimeout = redisIdleTimeout

	etcdTimeout, err := beego.AppConfig.Int("etcd_timeout")
	if err != nil {
		err = fmt.Errorf("init config failed ,read etcd_timeout error:%v", err)
	}
	etcdSecKeyPrefix := beego.AppConfig.String("etcd_sec_key_prefix")
	if len(etcdSecKeyPrefix) == 0 {
		err = fmt.Errorf("init config failed ,read etcd_sec_key_prefix error:%v", err)
		return
	}
	secKillConf.EtcdConf.EtcdSecKeyPrefix = etcdSecKeyPrefix

	etcdProductKey := beego.AppConfig.String("etcd_product_key")
	if len(etcdProductKey) == 0 {
		err = fmt.Errorf("init config failed ,read etcd_product_key error:%v", err)
		return
	}

	if strings.HasSuffix(secKillConf.EtcdConf.EtcdSecKeyPrefix, "/") == false {
		secKillConf.EtcdConf.EtcdSecKeyPrefix = secKillConf.EtcdConf.EtcdSecKeyPrefix + "/"
	}
	secKillConf.EtcdConf.EtcdSecProductKey = fmt.Sprintf("%s%s", secKillConf.EtcdConf.EtcdSecKeyPrefix, etcdProductKey)
	secKillConf.EtcdConf.EtcdTimeout = etcdTimeout
	secKillConf.EtcdConf.EtcdSecKeyPrefix = etcdSecKeyPrefix

	logPath := beego.AppConfig.String("log_path")
	loglevel := beego.AppConfig.String("log_level")
	secKillConf.LogConf.LogPath = logPath
	secKillConf.LogConf.LogLevel = loglevel

	secKillConf.CookieSecretKey = beego.AppConfig.String("cookie_secretkey")
	seclimit, err := beego.AppConfig.Int("user_sec_access_limit")
	if err != nil {
		err = fmt.Errorf("read user_sec_access_limit error:%v", err)
	}
	secKillConf.UserSecAccessLimit = seclimit
	referList := beego.AppConfig.String("refer_whitelist")
	if len(referList) > 0 {
		secKillConf.ReferWhiteList = strings.Split(referList, ",")
	}
	iplimit, err := beego.AppConfig.Int("ip_sec_access_limit")
	if err != nil {
		err = fmt.Errorf("read ip_sec_access_limit error:%v", err)
	}
	secKillConf.IpSecAccessLimit = iplimit

	//
	redisProxy2LayerAddr := beego.AppConfig.String("redis_proxy2layer_addr")
	logs.Debug("redis config addr :%v", redisProxy2LayerAddr)
	secKillConf.RedisProxy2LayerConf.RedisAddr = redisProxy2LayerAddr

	if len(redisProxy2LayerAddr) == 0 {
		err = fmt.Errorf("init config failed ,redis[%s] conf is null", redisProxy2LayerAddr)
		return
	}
	redisMaxIdle, err = beego.AppConfig.Int("redis_proxy2layer_idle")
	if err != nil {
		err = fmt.Errorf("init config failed ,read redis_proxy2layer_idle error:%v", err)
	}
	redisMaxActive, err = beego.AppConfig.Int("redis_proxy2layer_active")
	if err != nil {
		err = fmt.Errorf("init config failed ,read redis_proxy2layer_active error:%v", err)
	}
	redisIdleTimeout, err = beego.AppConfig.Int("redis_proxy2layer_idle_timeout")
	if err != nil {
		err = fmt.Errorf("init config failed ,read redis_proxy2layer_idle_timeout error:%v", err)
	}

	secKillConf.RedisProxy2LayerConf.RedisMaxIdle = redisMaxIdle
	secKillConf.RedisProxy2LayerConf.RedisMaxActive = redisMaxActive
	secKillConf.RedisProxy2LayerConf.RedisIdleTimeout = redisIdleTimeout

	writeGoNums, err := beego.AppConfig.Int("write_proxy2layer_goroutine_num")
	if err != nil {
		err = fmt.Errorf("init config failed ,read write_proxy2layer_goroutine_num error:%v", err)
	}
	secKillConf.WriteProxy2layerGoroutineNum = writeGoNums

	readGoNums, err := beego.AppConfig.Int("read_layer2proxy_goroutine_num")
	if err != nil {
		err = fmt.Errorf("init config failed ,read read_layer2proxy_goroutine_num error:%v", err)
	}
	secKillConf.ReadLayer2proxyGoroutineNum = readGoNums
	return
}
