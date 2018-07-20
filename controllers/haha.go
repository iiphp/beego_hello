package controllers

import (
	"github.com/astaxie/beego"
)

type HahaController struct {
	beego.Controller
}

func (this *HahaController) Index() {
	rst := make(map[string]interface{})
	rst["Name"] = "Tom"
	rst["Age"]  = this.GetString("age", "21")

	this.Data["json"] = rst

	this.ServeJSON(true)
}
