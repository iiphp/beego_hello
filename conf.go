package main

import (
	"sync"
	"time"
)

var (
	secConf = &SecConf{
		etcdConf:    etcdConf{addr: "", timeout: time.Second, secPrdKey: ""},
		redisConf:   redisConf{addr: "", maxIdleConn: 0, maxActiveConn: 0, idleTimeout: time.Second},
		prdConf:     make(map[int]*SecPrdConf, 8),
		prdConfLock: sync.RWMutex{},
	}
)

type SecConf struct {
	etcdConf    etcdConf
	redisConf   redisConf
	prdConf     map[int]*SecPrdConf
	prdConfLock sync.RWMutex
}

type redisConf struct {
	addr          string
	maxIdleConn   int
	maxActiveConn int
	idleTimeout   time.Duration
}

type etcdConf struct {
	addr      string
	timeout   time.Duration
	secPrdKey string
}

// 小写的属性，json 就不能反序列化了
type SecPrdConf struct {
	PrdId  int `json:"prd_id"`
	Status int `json:"status"`
}
