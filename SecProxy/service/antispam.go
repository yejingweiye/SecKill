package service

import (
	"fmt"
	"sync"
)

var (
	secLimitMgr = &SecLimitMgr{
		UserLimitMap: make(map[int]*SecLimit, 10000),
	}
)

type SecLimitMgr struct {
	UserLimitMap map[int]*SecLimit // userId ,limit
	IpLimitMap   map[string]*SecLimit
	lock         sync.Mutex
}

// 访问次数控制
func antiSpam(req *SecRequest) (err error) { // 用户过来开始计数
	secLimitMgr.lock.Lock()
	secLimit, ok := secLimitMgr.UserLimitMap[req.UserId]
	if !ok { // 用户第一次来，初始化
		secLimit = &SecLimit{}
		secLimitMgr.UserLimitMap[req.UserId] = secLimit

	}
	count := secLimit.Count(req.AccessTime.Unix()) //

	ipLimit, ok := secLimitMgr.IpLimitMap[req.ClientAddr]
	if !ok { // 用户第一次来，初始化
		secLimit = &SecLimit{}
		secLimitMgr.IpLimitMap[req.ClientAddr] = ipLimit
	}
	ipcount := secLimit.Count(req.AccessTime.Unix())
	secLimitMgr.lock.Unlock()
	if count > secKillConf.UserSecAccessLimit {
		err = fmt.Errorf("invalid request,system busy")
		return
	}
	if ipcount > secKillConf.IpSecAccessLimit {
		err = fmt.Errorf("invalid request,system busy")
		return
	}

	return
}

type SecLimit struct {
	count   int
	curTime int64
}

// 精确到秒，访问次数统计
func (p *SecLimit) Count(nowTime int64) (curCount int) {
	if p.curTime != nowTime {
		p.count = 1
		p.curTime = nowTime
		curCount = p.count
		return
	}
	p.count++
	curCount = p.count
	return
}

func (p *SecLimit) Check(nowTime int64) int {
	if p.curTime != nowTime {
		return 0
	}
	return p.count
}
