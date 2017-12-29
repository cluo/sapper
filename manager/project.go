package manager

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/juju/errors"
	"github.com/zssky/log"

	"github.com/dearcode/sapper/meta"
	"github.com/dearcode/sapper/util"
)

type projectInfo struct {
	ID int64 `json:"id"`
}

func (pi *projectInfo) GET(w http.ResponseWriter, r *http.Request) {
	if err := util.DecodeRequestValue(r, pi); err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	var ps []meta.Project
	_, err := query("project", fmt.Sprintf("id=%d", pi.ID), "", "", 0, 0, &ps)
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	buf, err := json.Marshal(ps[0])
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(buf)
}

func (p *project) GET(w http.ResponseWriter, r *http.Request) {
	vars := struct {
		Sort  string `json:"sort"`
		Order string `json:"order"`
		Page  int    `json:"offset"`
		Size  int    `json:"limit"`
	}{}
	u, err := session.User(r)
	if err != nil {
		log.Errorf("session.User error:%v, req:%v", errors.ErrorStack(err), r)
		response(w, Response{Status: http.StatusBadRequest, Message: err.Error()})
		return
	}

	if err := util.DecodeRequestValue(r, &vars); err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	var where string
	if !u.IsAdmin {
		where = fmt.Sprintf(" project.resource_id in (%s)", u.ResKey)
	}

	var ps []meta.Project
	total, err := query("project", where, vars.Sort, vars.Order, vars.Page, vars.Size, &ps)
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	if len(ps) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"total":0,"rows":[]}`))
		log.Debugf("project not found")
		return
	}

	result := struct {
		Total int            `json:"total"`
		Rows  []meta.Project `json:"rows"`
	}{total, ps}

	buf, err := json.Marshal(result)
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(buf)
}

func (p *project) DELETE(w http.ResponseWriter, r *http.Request) {
	vars := struct {
		ID int64 `json:"id"`
	}{}

	if err := util.DecodeRequestValue(r, &vars); err != nil {
		util.SendResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := del("project", vars.ID); err != nil {
		util.SendResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	util.SendResponse(w, 0, "")

	log.Debugf("delete project:%v, success", vars.ID)
}

func (p *project) POST(w http.ResponseWriter, r *http.Request) {
	vars := meta.Project{}
	u, err := session.User(r)
	if err != nil {
		log.Errorf("session.User error:%v, req:%v", errors.ErrorStack(err), r)
		response(w, Response{Status: http.StatusBadRequest, Message: err.Error()})
		return
	}

	if err := util.DecodeRequestValue(r, &vars); err != nil {
		util.SendResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	if !u.IsAdmin {
		vars.Email = u.Email
		vars.User = u.User
	}

	resID, err := rbacClient.PostResource(vars.Name, vars.Comments)
	if err != nil {
		log.Errorf("ResourceAdd req:%+v, error:%s", r, errors.ErrorStack(err))
		util.SendResponse(w, http.StatusInternalServerError, "添加资源出错")
		return
	}

	roleID, err := rbacClient.PostRole(vars.Name, "默认添加的管理组", vars.User, vars.Email)
	if err != nil {
		log.Errorf("RoleAdd req:%+v, error:%s", r, errors.ErrorStack(err))
		util.SendResponse(w, http.StatusInternalServerError, "添加角色出错")
		return
	}

	if _, err = rbacClient.PostRoleResource(roleID, resID); err != nil {
		log.Errorf("RelationResourceRoleAdd req:%+v, error:%s", r, errors.ErrorStack(err))
		util.SendResponse(w, http.StatusInternalServerError, "为项目授权角色出错")
		return
	}

	vars.ResourceID = resID
	vars.RoleID = roleID

	id, err := add("project", vars)
	if err != nil {
		if strings.Contains(err.Error(), "1062") {
			log.Errorf("add req:%+v, error:%s", r, errors.ErrorStack(err))
			util.SendResponse(w, http.StatusInternalServerError, "项目路径已存在, 项目路径在接口平台中是唯一的，不能重用")
			return
		}
		util.SendResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	util.SendResponse(w, 0, fmt.Sprintf(`{"id":%d}`, id))

	log.Debugf("add project:%v, id:%v, role:%d, resource:%d", vars, id, roleID, resID)
}

type project struct {
}

func (p *project) PUT(w http.ResponseWriter, r *http.Request) {
	vars := struct {
		ID       int64  `json:"id" valid:"Required"`
		Name     string `json:"name"  valid:"Required"`
		User     string `json:"user"  valid:"Required"`
		Email    string `json:"email"  valid:"Email"`
		Path     string `json:"path"  valid:"AlphaNumeric"`
		Comments string `json:"comments"  valid:"Required"`
	}{}
	u, err := session.User(r)
	if err != nil {
		log.Errorf("session.User error:%v, req:%v", errors.ErrorStack(err), r)
		response(w, Response{Status: http.StatusBadRequest, Message: err.Error()})
		return
	}

	if err := util.DecodeRequestValue(r, &vars); err != nil {
		util.SendResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	if !u.IsAdmin {
		vars.Email = u.Email
		vars.User = u.User
	}

	if err := updateProject(vars.ID, vars.Name, vars.User, vars.Email, vars.Path, vars.Comments); err != nil {
		util.SendResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	util.SendResponse(w, 0, "")

	log.Debugf("update project success, new:%+v", vars)
}

func getProjectResourceID(projectID int64) (int64, error) {
	return getResourceID("project", projectID)
}
