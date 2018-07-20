package routers

import (
	"hello/controllers"
	"github.com/astaxie/beego"
)

func init() {
    beego.Router("/", &controllers.MainController{})
    beego.Router("haha", &controllers.HahaController{}, "*:Index")

    beego.Router("seckill", &controllers.SecController{}, "*:SecKill")
    beego.Router("secshow", &controllers.SecController{}, "*:SecShow")
}
