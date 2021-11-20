package main

import (
	_ "SecKill/routers" // 不调用用函数就是让其路由init()运行
	"github.com/astaxie/beego"
)

func main() {
	err := initConfig()
	if err != nil {
		panic(err) // 打印错误
		return
	}

	//初始化
	err = initSec()
	if err != nil {
		panic(err)
		return
	}

	beego.Run()
}
