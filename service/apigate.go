package service

import (
	"encoding/json"
	"flag"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/juju/errors"
	"github.com/zssky/log"

	"github.com/dearcode/sapper/service/debug"
	"github.com/dearcode/sapper/util"
	"github.com/dearcode/sapper/util/etcd"
)

const (
	apigatePrefix = "/api"
)

var (
	etcdAddrs = flag.String("etcd", "192.168.180.104:12379 , 192.168.180.104:22379 , 192.168.180.104:32379", "etcd Endpoints.")
)

type keepalive struct {
	etcd *etcd.Client
}

type MicroAPP struct {
	Version string
	Host    string
	Port    int
	PID     int
}

func newMicroAPP(version, host string, port, pid int) *MicroAPP {
	return &MicroAPP{
		Version: version,
		PID:     pid,
		Host:    host,
		Port:    port,
	}
}

func (m MicroAPP) String() string {
	b, _ := json.Marshal(m)
	return string(b)
}

func newKeepalive() *keepalive {
	if *etcdAddrs == "" {
		return nil
	}

	//清理输入ip列表.
	addrs := strings.Split(*etcdAddrs, ",")
	for i := range addrs {
		addrs[i] = strings.TrimSpace(addrs[i])
	}

	//连接etcd.
	c, err := etcd.New(addrs)
	if err != nil {
		panic(err.Error())
	}

	return &keepalive{etcd: c}
}

// register 服务上线，注册到接口平台的etcd.
func (k *keepalive) start(ln net.Listener, doc document) error {
	if k == nil {
		return nil
	}

	// 获取本机服务地址
	local := util.LocalAddr()

	//因为服务是异步启动的，所以有一次获取不到绑定的端口.
	la := ln.Addr().String()
	port := la[strings.LastIndex(la, ":"):]
	addr := local + port
	log.Infof("listener addr:%v", addr)

	//注册所有接口
	for _, m := range doc.methods() {
		key := apigatePrefix + m + "/" + addr
		p, _ := strconv.Atoi(port)
		val := newMicroAPP(debug.GitHash, local, p, os.Getpid()).String()

		_, err := k.etcd.Keepalive(key, val)
		if err != nil {
			log.Errorf("etcd Keepalive key:%v, val:%v, error:%v", key, val, errors.ErrorStack(err))
			return errors.Trace(err)
		}

		log.Debugf("etcd put key:%v val:%v", key, val)
	}

	return nil
}

func (k *keepalive) stop() {
	if k == nil {
		return
	}

	k.etcd.Close()
}
