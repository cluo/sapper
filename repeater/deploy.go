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
				log.Errorf("invalid k:%s, data:%s, create:%v, modify:%v", e.Kv.Key, e.Kv.Value, e.IsCreate(), e.IsModify())
				continue
			}

			log.Debugf("key:%s, app:%#v, create:%v, modify:%v", e.Kv.Key, app, e.IsCreate(), e.IsModify())
			d.register(app)
		}
	}
}

func (d *deploy) register(app meta.MicroAPP) {

}

func (d *deploy) stop() {
	d.c.Close()
}
