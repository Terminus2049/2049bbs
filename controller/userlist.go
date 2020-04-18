package controller

import (
	"net/http"
	"strconv"

	"github.com/rs/xid"
	"github.com/terminus2049/2049bbs/model"
)

func (h *BaseHandler) UserList(w http.ResponseWriter, r *http.Request) {
	flag, btn, key := r.FormValue("flag"), r.FormValue("btn"), r.FormValue("key")
	if len(key) > 0 {
		_, err := strconv.ParseUint(key, 10, 64)
		if err != nil {
			w.Write([]byte(`{"retcode":400,"retmsg":"key type err"}`))
			return
		}
	}

	cmd := "hrscan"
	if btn == "prev" {
		cmd = "hscan"
	}

	db := h.App.Db

	if len(flag) == 0 {
		flag = "6"
	}

	if flag != "6" && flag != "0" && flag != "7" {
		w.Write([]byte(`{"retcode":403,"retmsg":"flag forbidden}`))
		return
	}

	pageInfo := model.UserListByFlag(db, cmd, "user_flag:"+flag, key, h.App.Cf.Site.PageShowNum)

	type pageData struct {
		PageData
		PageInfo model.UserPageInfo
		Flag     string
	}

	tpl := h.CurrentTpl(r)
	evn := &pageData{}
	evn.SiteCf = h.App.Cf.Site
	evn.Title = "用户列表"
	evn.IsMobile = tpl == "mobile"
	evn.ShowSideAd = true
	evn.PageName = "user_list"

	evn.PageInfo = pageInfo
	evn.Flag = flag

	token := h.GetCookie(r, "token")
	if len(token) == 0 {
		token := xid.New().String()
		h.SetCookie(w, "token", token, 1)
	}

	h.Render(w, tpl, evn, "layout.html", "userlist.html")
}
