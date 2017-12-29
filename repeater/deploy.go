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
		log.Infof("end watch, event:%+v", es)
		ks, err := d.c.List(apigatePrefix)
		if err != nil {
			log.Errorf("etcd list:%s error:%v", apigatePrefix, errors.ErrorStack(err))
			continue
		}

		for k, v := range ks {
			app := meta.MicroAPP{}
			if err := json.Unmarshal([]byte(v), &app); err != nil {
				log.Errorf("invalid data:%s", v)
				continue
			}

			log.Debugf("key:%v, app:%v", k, app)
		}
	}
}

func (d *deploy) register() {

}

func (d *deploy) stop() {
	d.c.Close()
}
