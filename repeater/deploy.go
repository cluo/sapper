package repeater

import (
	"encoding/json"
	"strings"

	"github.com/juju/errors"
	"github.com/zssky/log"

	"github.com/dearcode/sapper/meta"
	"github.com/dearcode/sapper/repeater/config"
	"github.com/dearcode/sapper/util/etcd"
)

const (
	apigatePrefix = "/api"
)

type deploy struct {
	c *etcd.Client
}

func newDeploy() (*deploy, error) {
	c, err := etcd.New(strings.Split(config.Repeater.ETCD.Hosts, ","))
	if err != nil {
		return nil, errors.Annotatef(err, config.Repeater.ETCD.Hosts)
	}

	return &deploy{c: c}, nil
}

func (d *deploy) start() {
	for {
		log.Debugf("begin watch key:%s", apigatePrefix)
		es := d.c.WatchPrefix(apigatePrefix)
		for _, e := range es {
			app := meta.MicroAPP{}
			if err := json.Unmarshal(e.Kv.Value, &app); err != nil {
				log.Errorf("invalid k:%s, data:%s, event:%v", e.Kv.Key, e.Kv.Value, e.Type.String())
				continue
			}

			log.Debugf("key:%s, app:%#v, event:%v", e.Kv.Key, app, e.Type.String())
			d.register(app)
		}
	}
}

//unregister 如果etcd中事务是删除，这里就去管理处删除.
func (d *deploy) unregister(app meta.MicroAPP) {
	log.Debugf("will unregister:%v", app)

}

//register 到管理处添加接口, 肯定是多个repeater同时上报的，所以添加操作要指定版本信息.
func (d *deploy) register(app meta.MicroAPP) {
	log.Debugf("will register:%v", app)

}

func (d *deploy) stop() {
	d.c.Close()
}
