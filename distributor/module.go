package distributor

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dearcode/crab/http/server"
	"github.com/dearcode/crab/orm"
	"github.com/zssky/log"

	"github.com/dearcode/sapper/distributor/config"
	"github.com/dearcode/sapper/util/etcd"
)

type module struct {
	ID         int64
	URL        string
	Name       string
	Interface  string
	Deploy     []deploy `db_table:"one2more"`
	CreateTime string   `db_default:"now()"`
}

//GET 获取module的部署情况.
func (m *module) GET(w http.ResponseWriter, r *http.Request) {
	vars := struct {
		ID int64
	}{}

	if err := server.ParseURLVars(r, &vars); err != nil {
		server.SendResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	db, err := mdb.GetConnection()
	if err != nil {
		log.Errorf("connect db error:%v", err.Error())
		server.SendResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer db.Close()

	if err = orm.NewStmt(db, "module").Where("id=%v", vars.ID).Query(m); err != nil {
		log.Errorf("query module:%v error:%v", vars.ID, err.Error())
		server.SendResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Debugf("module:%+v, id:%v", m, vars.ID)

	c, err := etcd.New(strings.Split(config.Distributor.ETCD.Hosts, ","))
	if err != nil {
		log.Errorf("etcd connect:%s error:%v", config.Distributor.ETCD.Hosts, err.Error())
		server.SendResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer c.Close()

	prefix := fmt.Sprintf("/api%s", m.Interface)
	keys, err := c.List(prefix)
	if err != nil {
		log.Errorf("etcd list:%s error:%v", prefix, err.Error())
		server.SendResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	log.Debugf("keys:%v, prefix:%v", keys, prefix)

	server.SendResponseData(w, keys)
}
