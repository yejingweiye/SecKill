package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	etcd_client "github.com/coreos/etcd/clientv3"
	"time"
)

const (
	EtcdKey = "/oldboy/backbend/secskill/product"
)

type SecInfoConf struct {
	ProductId int
	StartTime int
	EndTime   int
	Status    int
	Total     int
	Left      int // 剩余
}

// 往etcd插入数据
func SetSecInfoConfToEtcd() {
	cli, err := etcd_client.New(etcd_client.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		logs.Error("connect etcd failed,err:", err)
		return
	}
	fmt.Println("etcd conn success")
	defer cli.Close()

	var secInfoConfArr []SecInfoConf
	secInfoConfArr = append(
		secInfoConfArr,
		SecInfoConf{
			ProductId: 1029,
			StartTime: 1505008800, // 2017-9-10 10:00
			EndTime:   1505012400, // 2017-9-10 11:00
			Status:    0,          // 正常
			Total:     1000,
			Left:      1000,
		},
	)

	secInfoConfArr = append(
		secInfoConfArr,
		SecInfoConf{
			ProductId: 1027,
			StartTime: 1505008800, // 2017-9-10 10:00
			EndTime:   1639578869, // 2021-12-15 22:34:29
			Status:    0,          // 正常
			Total:     2000,
			Left:      1000,
		},
	)

	data, err := json.Marshal(secInfoConfArr)
	if err != nil {
		fmt.Println("json failed, ", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	_, err = cli.Put(ctx, EtcdKey, string(data))
	cancel()
	if err != nil {
		fmt.Println("put failed, err: ", err)
		return
	}

	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	resp, err := cli.Get(ctx, EtcdKey)
	cancel()
	if err != nil {
		fmt.Println("get failed, err: ", err)
		return
	}
	for _, ev := range resp.Kvs {
		fmt.Printf("%s : %s\n", ev.Key, ev.Value)
	}
}

func main() {
	SetSecInfoConfToEtcd()
}
