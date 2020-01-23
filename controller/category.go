package controller

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/terminus2049/2049bbs/model"
	"goji.io/pat"
)

func (h *BaseHandler) CategoryDetail(w http.ResponseWriter, r *http.Request) {
	btn, key, score := r.FormValue("btn"), r.FormValue("key"), r.FormValue("score")
	if len(key) > 0 {
		_, err := strconv.ParseUint(key, 10, 64)
		if err != nil {
			w.Write([]byte(`{"retcode":400,"retmsg":"key type err"}`))
			return
		}
	}
	if len(score) > 0 {
		_, err := strconv.ParseUint(score, 10, 64)
		if err != nil {
			w.Write([]byte(`{"retcode":400,"retmsg":"score type err"}`))
			return
		}
	}

	cid := pat.Param(r, "cid")
	_, err := strconv.Atoi(cid)
	if err != nil {
		w.Write([]byte(`{"retcode":400,"retmsg":"cid type err"}`))
		return
	}

	cmd := "zrscan"
	if btn == "prev" {
		cmd = "zscan"
	}

	db := h.App.Db
	scf := h.App.Cf.Site
	cobj, err := model.CategoryGetById(db, cid)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	currentUser, _ := h.CurrentUser(w, r)

	if cobj.Hidden && currentUser.Flag < 1 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"retcode":404,"retmsg":"仅登录用户可见"}`))
		return
	}
	pageInfo := model.ArticleList(db, cmd, "category_article_timeline:"+cid, key, score, scf.HomeShowNum, scf.TimeZone, "")

	type pageData struct {
		PageData
		Cobj     model.Category
		PageInfo model.ArticlePageInfo
	}

	if currentUser.IgnoreUser != "" {
		for _, uid := range strings.Split(currentUser.IgnoreUser, ",") {
			uid, err := strconv.ParseUint(uid, 10, 64)

			if err != nil {
				w.Write([]byte(`{"retcode":400,"retmsg":"type err"}`))
				return
			}

			for i := 0; i < len(pageInfo.Items); i++ {
				if pageInfo.Items[i].Uid == uid {
					pageInfo.Items = append(pageInfo.Items[:i], pageInfo.Items[i+1:]...)
					i--
				}
			}
		}
	}

	tpl := h.CurrentTpl(r)

	evn := &pageData{}
	evn.SiteCf = scf
	evn.Title = cobj.Name + " - " + scf.Name
	evn.Keywords = cobj.Name
	evn.Description = cobj.About
	evn.IsMobile = tpl == "mobile"

	evn.CurrentUser = currentUser
	evn.ShowSideAd = true
	evn.PageName = "category_detail"
	evn.HotNodes = model.CategoryHot(db, scf.CategoryShowNum, scf.MustLoginNodeIds)

	evn.Cobj = cobj
	evn.PageInfo = pageInfo

	h.Render(w, tpl, evn, "layout.html", "category.html")
}
