package service

import (
	"encoding/json"
	"github.com/astaxie/beego/logs"
)

func WriteHandler() {
	for {
		req := <-secKillConf.SecReqChan
		conn := secKillConf.Proxy2LayerRedisPool.Get()
		data, err := json.Marshal(req)
		if err != nil {
			logs.Error("json.Marshal failed,err:[%v]", err)
			conn.Close()
			continue
		}
		_, err = conn.Do("LPUSH", "sec_queue", data)
		if err != nil {
			logs.Error("LPUSH data failed,err:[%v]", err)
			conn.Close()
			continue
		}
		conn.Close()
	}

}

func ReadHandler() { // 这块代码暂时没有
}
