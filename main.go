package main

import (
	"fmt"
	"time"
	"context"
	"encoding/json"
	_ "hello/routers"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/gomodule/redigo/redis"
	"github.com/coreos/etcd/mvcc/mvccpb"
	etcd_client "github.com/coreos/etcd/clientv3"
)

var (
	redisPool *redis.Pool
	etcdClient *etcd_client.Client
)

func main() {
	var err error

	// 上来必须先初始化日志
	err = initLogger()
	if nil != err {
		panic(err)
		return
	}
	logs.Info("initLogger success.")

	// 从 app.conf 文件中，加载配置信息
	err = loadConf()
	if nil != err {
		panic(err)
		return
	}
	logs.Info("loadConf success.")

	redisPool, err = initRedis(secConf.redisConf)
	if nil != err {
		panic(err)
		return
	}
	logs.Info("initRedis success.")

	etcdClient, err = initEtcd(secConf.etcdConf)
	if nil != err {
		panic(err)
		return
	}
	logs.Info("initEtcd success, but have not connect ectd server.")

	err = initSecPrd()
	if nil != err {
		panic(err)
		return
	}
	logs.Info("initSecPrd success.")

	// 加载秒杀商品信息
	secPrd, err := loadSecPrd()
	if nil != err {
		panic(err)
		return
	}
	updateSecPrd(&secPrd)
	logs.Info("loadSecPrd success.")

	go watchEtcd(secConf.etcdConf.secPrdKey)

	beego.Run()
}

func initSecPrd() (err error) {
	secPrdConf := SecPrdConf{PrdId:777, Status:0}
	encodeJsonSecPrdConf, err := json.Marshal(secPrdConf)

	if nil != err {
		err = fmt.Errorf("initSecPrd json_encode failed. err=%s", err)
		logs.Error(err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	_, err = etcdClient.Put(ctx, secConf.etcdConf.secPrdKey, string(encodeJsonSecPrdConf))
	cancel()
	if nil != err {
		err = fmt.Errorf("initSecPrd put failed. err=%s", err)
		logs.Error(err.Error())
		return
	}
	return
}

func loadSecPrd() (secPrd SecPrdConf, err error) {
	ctx, concel   := context.WithTimeout(context.Background(), time.Second)
	response, err := etcdClient.Get(ctx, secConf.etcdConf.secPrdKey)
	concel()
	if nil != err {
		err = fmt.Errorf("load etcd_sec_prd failed. err=%s", err)
		logs.Error(err.Error())
		return
	}

	for _, ev := range response.Kvs {
		err = json.Unmarshal([]byte(string(ev.Value)), &secPrd)
		if nil != err {
			err = fmt.Errorf("etcd_sec_prd json_decode failed. err=%s", err)
			logs.Error(err.Error())
			return
		}
		logs.Debug("loadSecPrd secPrd=%+v", secPrd)
	}

	return
}

func updateSecPrd(secPrdConf *SecPrdConf)  {
	secConf.prdConfLock.Lock()
	defer secConf.prdConfLock.Unlock()
	secConf.prdConf[secPrdConf.PrdId] = secPrdConf
	logs.Debug("updateSecPrd secConf.prdConf=%+v", secPrdConf)
}

func watchEtcd(key string) {
	logs.Debug("start watchEtcd key=%s", key)
	var secPrd    SecPrdConf
	var watchSucc bool

	for {
		rch := etcdClient.Watch(context.Background(), key)
		for watchResp := range rch {
			logs.Debug("watchResp=%+v", watchResp)
			watchSucc = false
			for ek, ev := range watchResp.Events {
				logs.Debug("watchEtcd key=%s ek=%s ev=%+v", key, ek, ev)
				if mvccpb.DELETE == ev.Type {
					logs.Warning("watchEtcd key=%s delete", key)
					continue
				}

				if mvccpb.PUT == ev.Type && key == string(ev.Kv.Key) {
					err := json.Unmarshal(ev.Kv.Value, &secPrd)
					if nil != err {
						watchSucc = false
						logs.Error("watchEtcd json_decode failed key=%s ek=%s ev=%+v", key, ek, ev)
						continue
					}
					watchSucc = true
				}
			}

			if watchSucc {
				updateSecPrd(&secPrd)
			}
		}
	}
	logs.Debug("finish watchEtcd key=%s", key)
}
