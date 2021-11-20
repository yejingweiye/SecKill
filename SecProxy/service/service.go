package service

import (
	"crypto/md5"
	"fmt"
	"github.com/astaxie/beego/logs"
	"time"
)

var (
	secKillConf *SecKillConf
)

func SecInfoList() (data []map[string]interface{}, code int, err error) {
	secKillConf.RWSecProductLock.RLock()
	defer secKillConf.RWSecProductLock.RUnlock()
	for _, v := range secKillConf.SecProductInfoMap {
		item, _, err := SecInfoById(v.ProductId)
		if err != nil {
			logs.Error("get product_id[%d] failed,err: %x ", v.ProductId, err)
			continue
		}
		data = append(data, item)
	}
	return
}

func SecInfo(productId int) (data []map[string]interface{}, code int, err error) {
	secKillConf.RWSecProductLock.RLock()
	defer secKillConf.RWSecProductLock.RUnlock()
	item, code, err := SecInfoById(productId)
	if err != nil {
		return
	}
	data = append(data, item)
	return
}

func SecInfoById(productId int) (data map[string]interface{}, code int, err error) {
	secKillConf.RWSecProductLock.RLock()
	defer secKillConf.RWSecProductLock.RUnlock()
	v, ok := secKillConf.SecProductInfoMap[productId]
	if !ok {
		code = ErrNotFoundProductId
		fmt.Errorf("not found product_id : %d", productId)
		return
	}

	start := false
	end := false
	status := "success"
	now := time.Now().Unix()
	if now-v.StartTime < 0 {
		start = false
		end = false
		status = "sec kill is not start"
		code = ErrActiveNotStart
	}
	if now-v.StartTime >= 0 {
		start = true
	}
	if now-v.EndTime > 0 {
		start = false
		end = true
		status = "sec kill is already end"
		code = ErrActiveAlreadyEnd
	}
	if v.Status == ProductStatusForceSaleOut || v.Status == ProductStatusSaleOut {
		start = false
		end = true
		status = "product is sale out"
		code = ErrActiveSaleOut
	}

	data = make(map[string]interface{})
	data["product_id"] = productId
	data["start"] = start
	data["end"] = end
	data["status"] = status
	return
}

// md5(userId:secret)
func userCheck(req *SecRequest) (err error) {
	found := false
	for _, refer := range secKillConf.ReferWhiteList {
		if refer == req.ClientRefence {
			found = true
			break // 放行
		}
	}
	if !found {
		err = fmt.Errorf("invaild request")
		logs.Warn("user [%d] is reject by refer,req[%v]", req.UserId, req)
		return
	}
	authData := fmt.Sprintf("%d:%s", req.UserId, secKillConf.CookieSecretKey)
	authSign := fmt.Sprintf("%x", md5.Sum([]byte(authData))) // byte数值格式化16进制
	if authSign != req.UserAuthSign {
		err = fmt.Errorf("invalid user cookie auth")
		return
	}

	secKillConf.SecReqChan <- req
	return
}

// 用户校验 频率
func SecKill(req *SecRequest) (data map[string]interface{}, code int, err error) {
	secKillConf.RWSecProductLock.RLock()
	defer secKillConf.RWSecProductLock.RUnlock()
	err = userCheck(req)
	if err != nil {
		code = ErrUserCheckAuthFailed
		logs.Warn("userId[%d] invalid,check failed, req[%v] ", req.UserId, req)
		return
	}

	err = antiSpam(req) //
	if err != nil {
		code = ErrUserServiceBusy
		logs.Warn("anti busy, req[%v] ", req.UserId, req)
		return
	}
	data, code, err = SecInfoById(req.ProductId)
	if err != nil {
		logs.Warn("userId[%d] secInfoby Id failed,req[%v]  ", req.UserId, req)
		return
	}
	if code != 0 {
		logs.Warn("userId[%d] secInfobyid failed,code[%d],req[%v]", req.UserId, code, req)
		return
	}

	return
}
