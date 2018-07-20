package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"hello/models"
)

type SecController struct {
	beego.Controller
}

func (this *SecController) SecKill() {
	rst := make(map[string]interface{})
	rst["Name"] = "Tom"
	rst["Age"]  = this.GetString("age", "21")

	this.Data["json"] = rst

	this.ServeJSON(true)
}

func (this *SecController) SecShow() {
	_, err := this.GetInt("product_id")
	resp := make(map[string]interface{})
	resp["code"] = 0
	resp["message"] = ""

	defer func() {
		logs.Debug("SecShow response=%+v", resp)
		this.Data["json"] = resp
		this.ServeJSON(true)
	}()

	if nil != err {
		resp["code"] = 100001
		resp["message"] = "invalid product_id"
		logs.Error("invalid request, GetInt product_id failed, err=%s", err)
		return
	}

	// 每一层自己负责对外的 err_code 和 err_message，内部的错误记日志，不要让内部的函数，考虑对最外层用户看到什么
	secModel := models.SecModel{}
	secPrd, err := secModel.AllProducts()
	if nil != err {
		resp["code"] = 100002
		resp["message"] = "all products failed"
		logs.Error("models.SecModel.AllProducts failed, err=%s", err)
		return
	}

	respData := make(map[string]interface{})
	respData["product"] = secPrd
	resp["data"] = respData
	return
}
