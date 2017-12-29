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
	"github.com/dearcode/sapper/util"
)

type interfaceRun struct {
}

func (i *interfaceRun) PUT(w http.ResponseWriter, r *http.Request) {
	vars := struct {
		ID int64 `json:"id" valid:"Required"`
	}{}

	if err := util.DecodeRequestValue(r, &vars); err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	iface, err := queryInterfaceInfo(vars.ID)
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	var vs []meta.Variable

	if _, err = query("variable", fmt.Sprintf("interface_id=%d", vars.ID), "id", "asc", 0, 0, &vs); err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}
	log.Debugf("load var id:%d data:%+v", vars.ID, vs)

	req := &http.Request{}
	backend := iface.Backend
	if !strings.Contains(backend, "?") {
		backend += fmt.Sprintf("?t=%d", time.Now().Unix())
	}
	req.Header = make(map[string][]string)

	reqBody := bytes.NewBuffer([]byte{})

	for _, v := range vs {
		val := r.FormValue(v.Name)
		if val == "" && v.IsRequired {
			fmt.Fprintf(w, "字段:"+v.Name+"不能为空")
			return
		}
		switch v.Postion {
		case server.URI:
			backend += fmt.Sprintf("&%s=%s", v.Name, val)
		case server.HEADER:
			if req.Header == nil {
				req.Header = make(map[string][]string)
			}
			req.Header.Add(v.Name, val)
			log.Debugf("header add name:%v, val:%v", v.Name, val)
		case server.FORM:
			if reqBody.Len() > 0 {
				reqBody.WriteString("&")
			}
			reqBody.WriteString(fmt.Sprintf("%s=%s", v.Name, url.QueryEscape(val)))
		}
	}

	if req.URL, err = req.URL.Parse(backend); err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	if iface.Method == server.GET {
		req.Method = "GET"
	} else {
		req.Method = "POST"
	}

	req.ContentLength = int64(reqBody.Len())
	req.Body = ioutil.NopCloser(reqBody)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	c := &http.Client{Timeout: time.Second * 5}
	log.Debugf("test req:%+v", req)
	resp, err := c.Do(req)
	if err != nil {
		log.Debugf("do error:%v", err.Error())
		fmt.Fprintf(w, err.Error())
		return
	}

	buf, err := httputil.DumpResponse(resp, true)
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	//w.Header().Set("Content-Type", "application/json")
	w.Write(buf)
}

type interfaceInfo struct {
}

//GET interfaceInfo.
func (ii *interfaceInfo) GET(w http.ResponseWriter, r *http.Request) {
	vars := struct {
		ID int64 `json:"id" valid:"Required"`
	}{}

	if err := util.DecodeRequestValue(r, &vars); err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	iface, err := queryInterfaceInfo(vars.ID)
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}
	iface.ID = vars.ID

	buf, err := json.Marshal(iface)
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(buf)

	log.Debugf("find id:%d, interface:%+v\n", vars.ID, iface)
}

type interfaceAction struct {
	ID        int64  `json:"id"`
	State     int    `json:"state"`
	ProjectID int64  `json:"pid"`
	Sort      string `json:"sort"`
	Order     string `json:"order"`
	Page      int    `json:"offset"`
	Size      int    `json:"limit"`
	Search    string `json:"search"`
}

func (i *interfaceAction) GET(w http.ResponseWriter, r *http.Request) {
	u, err := session.User(r)
	if err != nil {
		log.Errorf("session.User error:%v, req:%v", errors.ErrorStack(err), r)
		response(w, Response{Status: http.StatusBadRequest, Message: err.Error()})
		return
	}

	if err = util.DecodeRequestValue(r, i); err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	resID, err := getProjectResourceID(i.ProjectID)
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	if err = u.assert(resID); err != nil {
		log.Errorf("resourceID:%d, vars:%+v, err:%v", resID, i, errors.ErrorStack(err))
		fmt.Fprintf(w, err.Error())
		return
	}

	var is []meta.Interface

	where := fmt.Sprintf("project_id=%d and project.id = interface.project_id", i.ProjectID)

	if i.State == 1 {
		where += " and state=1 "
	}

	if i.Search != "" {
		where += " and (interface.name like '%" + i.Search + "%'"
		where += " or interface.user like '%" + i.Search + "%'"
		where += " or interface.comments like '%" + i.Search + "%'"
		where += " or interface.path like '%" + i.Search + "%'"
		where += " or interface.backend like '%" + i.Search + "%')"
	}

	if i.Sort != "" {
		i.Sort = "interface." + i.Sort
	}

	total, err := query("interface, project", where, i.Sort, i.Order, i.Page, i.Size, &is)
	if err != nil {
		log.Errorf("query interface error:%v", errors.ErrorStack(err))
		fmt.Fprintf(w, err.Error())
		return
	}

	if len(is) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"total":0,"rows":[]}`))
		log.Debugf("interface with projectID:%d, not found", i.ProjectID)
		return
	}

	result := struct {
		Total int              `json:"total"`
		Rows  []meta.Interface `json:"rows"`
	}{total, is}

	buf, err := json.Marshal(result)
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(buf)

	log.Debugf(":%+v\n", string(buf))
}

func (i *interfaceAction) DELETE(w http.ResponseWriter, r *http.Request) {
	if err := util.DecodeRequestValue(r, i); err != nil {
		log.Errorf("%v DecodeRequestValue error:%v", r.RemoteAddr, errors.ErrorStack(err))
		util.SendResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := del("interface", i.ID); err != nil {
		log.Errorf("%v interface error:%v", i, errors.ErrorStack(err))
		util.SendResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	util.SendResponse(w, 0, "")

	log.Debugf("delete Interface:%v, success", i.ID)
}

func (i *interfaceAction) POST(w http.ResponseWriter, r *http.Request) {
	vars := iface{}

	u, err := session.User(r)
	if err != nil {
		log.Errorf("session.User error:%v, req:%v", errors.ErrorStack(err), r)
		response(w, Response{Status: http.StatusBadRequest, Message: err.Error()})
		return
	}

	if err = util.DecodeRequestValue(r, &vars); err != nil {
		log.Errorf("invalid req:%+v", r)
		util.SendResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	resID, err := getProjectResourceID(vars.ProjectID)
	if err != nil {
		log.Errorf("invalid req:%+v", r)
		util.SendResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if err = u.assert(resID); err != nil {
		log.Errorf("resourceID:%d, vars:%+v, err:%v", resID, vars, errors.ErrorStack(err))
		fmt.Fprintf(w, err.Error())
		return
	}

	vars.State = 0
	vars.User = u.User
	vars.Email = u.Email

	db, err := mdb.GetConnection()
	if err != nil {
		log.Errorf("GetConnection error:%v", errors.ErrorStack(err))
		response(w, Response{Status: http.StatusInternalServerError, Message: err.Error()})
		return
	}
	defer db.Close()

	id, err := orm.NewStmt(db, "interface").Insert(vars)
	if err != nil {
		if strings.Contains(err.Error(), "1062") {
			log.Errorf("add req:%+v, error:%s", r, errors.ErrorStack(err))
			util.SendResponse(w, http.StatusInternalServerError, "接口路径已存在, 接口路径在整个项目中是唯一的，不可重复")
			return
		}
		util.SendResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	util.SendResponseJSON(w, &id)

	log.Debugf("add Interface success, id:%v", id)
}

func (i *interfaceAction) PUT(w http.ResponseWriter, r *http.Request) {
	vars := struct {
		ID       int64  `json:"id" valid:"Required"`
		Name     string `json:"name"  valid:"Required"`
		User     string `json:"user"`
		Email    string `json:"email"`
		Method   int    `json:"method"`
		Path     string `json:"path"  valid:"AlphaNumeric"`
		Backend  string `json:"backend"  valid:"Required"`
		Comments string `json:"comments"  valid:"Required"`
		Level    int    `json:"level"`
	}{}

	if err := util.DecodeRequestValue(r, &vars); err != nil {
		log.Errorf("invalid req:%+v", r)
		util.SendResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := updateInterface(vars.ID, vars.Method, vars.Level, vars.Name, vars.Path, vars.Backend, vars.Comments, vars.User, vars.Email); err != nil {
		util.SendResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	util.SendResponse(w, 0, "")

	log.Debugf("update Interface success, new:%+v", vars)
}

type interfaceDeploy struct {
}

func (id *interfaceDeploy) PUT(w http.ResponseWriter, r *http.Request) {
	vars := struct {
		ID int64 `json:"id" valid:"Required"`
	}{}

	if err := util.DecodeRequestValue(r, &vars); err != nil {
		log.Errorf("invalid req:%+v", r)
		util.SendResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := deployInterface(vars.ID); err != nil {
		util.SendResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	util.SendResponse(w, 0, "")

	log.Debugf("deploy Interface:%d success", vars.ID)
}
