package manager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/dearcode/crab/http/server"
	"github.com/dearcode/crab/orm"
	"github.com/juju/errors"
	"github.com/zssky/log"

	"github.com/dearcode/sapper/meta"
	"github.com/dearcode/sapper/util/etcd"
)

type service struct {
}

func (s *service) GET(w http.ResponseWriter, r *http.Request) {
	vars := struct {
		ProjectID int64 `json:"projectID" valid:"Required"`
	}{}

	if err := util.DecodeRequestValue(r, &vars); err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	db, err := mdb.GetConnection()
	if err != nil {
		log.Errorf("GetConnection error:%v", errors.ErrorStack(err))
		response(w, Response{Status: http.StatusInternalServerError, Message: err.Error()})
		return
	}
	defer db.Close()

	var p meta.Project

	if err = orm.NewStmt(db, "project").Where("id=%d", vars.ProjectID).Query(&p); err != nil {
		log.Errorf("query project:%d error:%v", vars.ProjectID, errors.ErrorStack(err))
		response(w, Response{Status: http.StatusInternalServerError, Message: err.Error()})
		return
	}

    // source http://git.jd.com/dbs/faas_test_001
    key := p.Source[7:]




}
