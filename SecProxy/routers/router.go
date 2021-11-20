package routers

import (
	"SecKill/controllers"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
)

func init() {
	logs.Debug("enter router init ....")
	beego.Router("/", &controllers.MainController{})
	beego.Router("/seckill", &controllers.SkillController{}, "*:SecKill") // * 代表post get 请求
	beego.Router("/secinfo", &controllers.SkillController{}, "*:SecInfo")
}
