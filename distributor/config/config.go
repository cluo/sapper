package config

import (
	"flag"

	"github.com/dearcode/crab/config"
)

var (
	// Distributor 配置信息.
	Distributor distributorConfig
	configPath  = flag.String("c", "./config/distributor.ini", "config file")
)

type serverConfig struct {
	Timeout   int
	Domain    string
	Script    string
	BuildPath string
	SecretKey string
}

type dbConfig struct {
	IP      string
	Port    int
	Name    string
	User    string
	Passwd  string
	Charset string
}

type etcdConfig struct {
	Hosts string
}

type distributorConfig struct {
	Server serverConfig
	DB     dbConfig
	ETCD   etcdConfig
}

//Load 加载配置文件.
func Load() error {
	return config.LoadConfig(*configPath, &Distributor)
}
