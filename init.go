package main

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	etcd_client "github.com/coreos/etcd/clientv3"
	"github.com/gomodule/redigo/redis"
	"time"
	"encoding/json"
)

func loadConf() (err error) {
	err = loadRedisConf()
	if nil != err {
		return
	}
	logs.Info("loadRedisConf success.")

	err = loadEtcdConf()
	if nil != err {
		return
	}
	logs.Info("loadEtcdConf success.")

	return
}

func loadRedisConf() (err error) {
	redisAddr := beego.AppConfig.String("redis_addr")

	redisMaxIdle, err := beego.AppConfig.Int("redis_max_idle")
	if nil != err {
		err = fmt.Errorf("load app.conf failed redis_max_idle=%s", err)
		logs.Error(err.Error())
		return
	}

	redisMaxActive, err := beego.AppConfig.Int("redis_max_active")
	if nil != err {
		err = fmt.Errorf("load app.conf failed redis_max_active=%s", err)
		logs.Error(err.Error())
		return
	}

	redisIdleTimeout, err := beego.AppConfig.Int("redis_idle_timeout")
	if nil != err {
		err = fmt.Errorf("load app.conf failed redis_idle_timeout=%s", err)
		logs.Error(err.Error())
		return
	}

	secConf.redisConf.addr = redisAddr
	secConf.redisConf.maxIdleConn = redisMaxIdle
	secConf.redisConf.maxActiveConn = redisMaxActive
	secConf.redisConf.idleTimeout = time.Duration(redisIdleTimeout) * time.Millisecond
	return
}

func loadEtcdConf() (err error) {
	etcdAddr := beego.AppConfig.String("etcd_addr")

	etcdTimeout, err := beego.AppConfig.Int("etcd_timeout")
	if nil != err {
		err = fmt.Errorf("load app.conf failed etcd_timeout=%s", err)
		logs.Error(err.Error())
		return
	}

	secConf.etcdConf.addr = etcdAddr
	secConf.etcdConf.timeout = time.Duration(etcdTimeout) * time.Millisecond

	secPrdKey := beego.AppConfig.String("etcd_sec_prd")
	if 0 == len(secPrdKey) {
		err = fmt.Errorf("app.conf etcd_sec_prd is nil")
		logs.Error(err.Error())
		return
	}
	secConf.etcdConf.secPrdKey = secPrdKey
	return
}


// Todo: etcd 没有测试是否能够成功连接
func initEtcd(conf etcdConf) (etcdCli *etcd_client.Client, err error) {
	etcdCli, err = etcd_client.New(etcd_client.Config{
		Endpoints:   []string{conf.addr},
		DialTimeout: conf.timeout,
	})

	if nil != err {
		err = fmt.Errorf("etcd_client.New failed. err=%s", err)
		logs.Error(err.Error())
		return
	}

	return
}

func initRedis(conf redisConf) (*redis.Pool, error) {
	pool := &redis.Pool{
		MaxIdle:     conf.maxIdleConn,
		MaxActive:   conf.maxActiveConn,
		IdleTimeout: conf.idleTimeout,
		Dial: func() (conn redis.Conn, err error) {
			return redis.Dial("tcp", conf.addr)
		},
	}

	conn := pool.Get()
	defer conn.Close()

	// 正常情况下，返回 PONG, nil
	_, err := redis.DoWithTimeout(conn, time.Millisecond*10, "ping")
	if nil != err {
		err = fmt.Errorf("redis ping failed. err=%s", err)
		logs.Error(err.Error())
	}
	return pool, err
}

func initLogger() (err error) {
	conf := make(map[string]interface{})
	conf["filename"] = beego.AppConfig.String("log_path")
	conf["level"]    = convertLogLevel(beego.AppConfig.String("log_level"))

	encodeJsonConf, err := json.Marshal(conf)
	if nil != err {
		err = fmt.Errorf("initLogger encodeJsonConf failed. err=%s", err)
		logs.Error(err.Error())
		return
	}

	err = logs.SetLogger(logs.AdapterFile, string(encodeJsonConf))
	if nil != err {
		err = fmt.Errorf("logs.SetLogger failed. err=%s", err)
		logs.Error(err.Error())
		return
	}

	return
}

func convertLogLevel(log_level string) int {
	switch log_level {
	case "debug":
		return logs.LevelDebug
	case "warning":
		return logs.LevelWarning
	case "info":
		return logs.LevelInfo
	case "error":
		return logs.LevelError
	case "notice":
		return logs.LevelNotice
	case "trace":
		return logs.LevelTrace
	}
	return logs.LevelDebug
}