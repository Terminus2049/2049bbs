package controller

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dchest/captcha"
	"github.com/ego008/youdb"
	"github.com/rs/xid"
	"github.com/terminus2049/2049bbs/model"
	"github.com/terminus2049/2049bbs/util"
	"goji.io/pat"
)

func (h *BaseHandler) UserLogin(w http.ResponseWriter, r *http.Request) {
	type pageData struct {
		PageData
		Act       string
		Token     string
		CaptchaId string
	}
	act := strings.TrimLeft(r.RequestURI, "/")
	title := "登录"
	if act == "register" {
		title = "注册"
	}

	tpl := h.CurrentTpl(r)
	evn := &pageData{}
	evn.SiteCf = h.App.Cf.Site
	evn.Title = title
	evn.Keywords = ""
	evn.Description = ""
	evn.IsMobile = tpl == "mobile"

	evn.ShowSideAd = true
	evn.PageName = "user_login_register"

	evn.Act = act
	evn.CaptchaId = captcha.New()

	token := h.GetCookie(r, "token")
	if len(token) == 0 {
		token := xid.New().String()
		h.SetCookie(w, "token", token, 1)
	}

	h.Render(w, tpl, evn, "layout.html", "userlogin.html")
}

func (h *BaseHandler) UserLoginPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	token := h.GetCookie(r, "token")
	if len(token) == 0 {
		w.Write([]byte(`{"retcode":400,"retmsg":"token cookie missed"}`))
		return
	}

	act := strings.TrimLeft(r.RequestURI, "/")

	type recForm struct {
		Name            string `json:"name"`
		Password        string `json:"password"`
		CaptchaId       string `json:"captchaId"`
		CaptchaSolution string `json:"captchaSolution"`
	}

	type response struct {
		normalRsp
	}

	decoder := json.NewDecoder(r.Body)
	var rec recForm
	err := decoder.Decode(&rec)
	if err != nil {
		w.Write([]byte(`{"retcode":400,"retmsg":"json Decode err:` + err.Error() + `"}`))
		return
	}
	defer r.Body.Close()

	if len(rec.Name) == 0 || len(rec.Password) == 0 {
		w.Write([]byte(`{"retcode":400,"retmsg":"name or pw is empty"}`))
		return
	}
	nameLow := strings.ToLower(rec.Name)
	if !util.IsUserName(nameLow) {
		w.Write([]byte(`{"retcode":400,"retmsg":"name fmt err"}`))
		return
	}

	if !captcha.VerifyString(rec.CaptchaId, rec.CaptchaSolution) {
		w.Write([]byte(`{"retcode":405,"retmsg":"验证码错误","newCaptchaId":"` + captcha.New() + `"}`))
		return
	}

	db := h.App.Db
	timeStamp := uint64(time.Now().UTC().Unix())

	if act == "login" {
		bn := "user_login_token"
		key := []byte(token + ":loginerr")
		if db.Zget(bn, key).State == "ok" {
			// todo
			//w.Write([]byte(`{"retcode":400,"retmsg":"name and pw not match"}`))
			//return
		}
		uobj, err := model.UserGetByName(db, nameLow)
		if err != nil {
			w.Write([]byte(`{"retcode":405,"retmsg":"json Decode err:` + err.Error() + `","newCaptchaId":"` + captcha.New() + `"}`))
			return
		}
		if uobj.Password != rec.Password {
			db.Zset(bn, key, uint64(time.Now().UTC().Unix()))
			w.Write([]byte(`{"retcode":405,"retmsg":"name and pw not match","newCaptchaId":"` + captcha.New() + `"}`))
			return
		}
		sessionid := xid.New().String()
		uobj.LastLoginTime = timeStamp
		uobj.Session = sessionid
		jb, _ := json.Marshal(uobj)
		db.Hset("user", youdb.I2b(uobj.Id), jb)
		h.SetCookie(w, "SessionID", strconv.FormatUint(uobj.Id, 10)+":"+sessionid, 365)
	} else {
		// register
		siteCf := h.App.Cf.Site
		if siteCf.CloseReg {
			w.Write([]byte(`{"retcode":400,"retmsg":"stop to new register"}`))
			return
		}
		if db.Hget("user_name2uid", []byte(nameLow)).State == "ok" {
			w.Write([]byte(`{"retcode":405,"retmsg":"name is exist","newCaptchaId":"` + captcha.New() + `"}`))
			return
		}

		userId, _ := db.HnextSequence("user")
		flag := 5
		if siteCf.RegReview {
			flag = 1
		}

		if userId == 1 {
			flag = 99
		}

		uobj := model.User{
			Id:            userId,
			Name:          rec.Name,
			Password:      rec.Password,
			Flag:          flag,
			RegTime:       timeStamp,
			LastLoginTime: timeStamp,
			Session:       xid.New().String(),
		}

		// 从指定的用户中随机选一个头像作为新注册用户头像
		// 指定用户必须连续，取最小id和最大id

		rand.Seed(time.Now().UnixNano())
		min := siteCf.AvatarMinId // 2539
		max := siteCf.AvatarMaxId // 2558
		sampleID := rand.Intn(max-min+1) + min
		uidStr := strconv.FormatUint(uint64(sampleID), 10)
		uobj.Avatar = uidStr
		uobj.IgnoreLimitedUsers = true

		jb, _ := json.Marshal(uobj)
		db.Hset("user", youdb.I2b(uobj.Id), jb)
		db.Hset("user_name2uid", []byte(nameLow), youdb.I2b(userId))
		db.Hset("user_flag:"+strconv.Itoa(flag), youdb.I2b(uobj.Id), []byte(""))

		h.SetCookie(w, "SessionID", strconv.FormatUint(uobj.Id, 10)+":"+uobj.Session, 365)
	}

	h.DelCookie(w, "token")

	rsp := response{}
	rsp.Retcode = 200
	json.NewEncoder(w).Encode(rsp)
}

func (h *BaseHandler) UserNotification(w http.ResponseWriter, r *http.Request) {
	currentUser, _ := h.CurrentUser(w, r)
	if currentUser.Id == 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	type pageData struct {
		PageData
		PageInfo model.ArticlePageInfo
	}

	db := h.App.Db
	scf := h.App.Cf.Site

	tpl := h.CurrentTpl(r)

	evn := &pageData{}
	evn.SiteCf = scf
	evn.Title = "站内提醒 - " + scf.Name
	evn.IsMobile = tpl == "mobile"

	evn.CurrentUser = currentUser
	evn.ShowSideAd = true
	evn.PageName = "user_notification"
	evn.HotNodes = model.CategoryHot(db, scf.CategoryShowNum, scf.MustLoginNodeIds)

	evn.PageInfo = model.ArticleNotificationList(db, currentUser.Notice, scf.TimeZone)

	// fix currentUser.NoticeNum != len(evn.PageInfo.Items)
	if currentUser.NoticeNum != len(evn.PageInfo.Items) {
		var keys [][]byte
		for _, v := range strings.Split(currentUser.Notice, ",") {
			keys = append(keys, youdb.DS2b(v))
		}

		var newKeys []string
		db.Hmget("article", keys).KvEach(func(key, value youdb.BS) {
			newKeys = append(newKeys, youdb.B2ds(key))
		})

		currentUser.Notice = strings.Join(newKeys, ",")
		currentUser.NoticeNum = len(newKeys)

		jb, _ := json.Marshal(currentUser)
		db.Hset("user", youdb.I2b(currentUser.Id), jb)

		evn.CurrentUser = currentUser
	}

	h.Render(w, tpl, evn, "layout.html", "notification.html")
}

func (h *BaseHandler) UserLogout(w http.ResponseWriter, r *http.Request) {
	cks := []string{"SessionID", "QQUrlState", "WeiboUrlState", "token"}
	for _, k := range cks {
		h.DelCookie(w, k)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *BaseHandler) UserDetail(w http.ResponseWriter, r *http.Request) {
	act, btn, key, score := r.FormValue("act"), r.FormValue("btn"), r.FormValue("key"), r.FormValue("score")
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

	db := h.App.Db
	scf := h.App.Cf.Site

	uid := pat.Param(r, "uid")
	uidi, err := strconv.ParseUint(uid, 10, 64)
	if err != nil {
		uid = model.UserGetIdByName(db, strings.ToLower(uid))
		if uid == "" {
			w.Write([]byte(`{"retcode":400,"retmsg":"uid type err"}`))
			return
		}
		http.Redirect(w, r, "/member/"+uid, 301)
		return
	}

	cmd := "rscan"
	if btn == "prev" {
		cmd = "scan"
	}

	uobj, err := model.UserGetById(db, uidi)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	currentUser, _ := h.CurrentUser(w, r)

	if uobj.Flag == -1 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"retcode":404,"retmsg":"你来晚了，用户已注销"}`))
		return
	}

	if uobj.Hidden && currentUser.Flag < 99 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"retcode":404,"retmsg":"用户主页已被管理员隐藏"}`))
		return
	}

	var pageInfo model.ArticlePageInfo

	if act == "reply" {
		tb := "user_article_reply:" + uid
		// pageInfo = model.UserArticleList(db, cmd, tb, key, h.App.Cf.Site.PageShowNum)
		pageInfo = model.ArticleList(db, "z"+cmd, tb, key, score, scf.PageShowNum, scf.TimeZone)
	} else {
		act = "post"
		tb := "user_article_timeline:" + uid
		pageInfo = model.UserArticleList(db, "h"+cmd, tb, key, scf.PageShowNum, scf.TimeZone)
	}

	type userDetail struct {
		model.User
		RegTimeFmt string
	}
	type pageData struct {
		PageData
		Act      string
		Uobj     userDetail
		PageInfo model.ArticlePageInfo
	}

	tpl := h.CurrentTpl(r)

	evn := &pageData{}
	evn.SiteCf = scf
	evn.Title = uobj.Name + " - " + scf.Name
	evn.Keywords = uobj.Name
	evn.Description = uobj.About
	evn.IsMobile = tpl == "mobile"

	evn.CurrentUser = currentUser
	evn.ShowSideAd = true
	evn.PageName = "category_detail"
	evn.HotNodes = model.CategoryHot(db, scf.CategoryShowNum, scf.MustLoginNodeIds)

	evn.Act = act
	evn.Uobj = userDetail{
		User:       uobj,
		RegTimeFmt: util.TimeFmt(uobj.RegTime, "2006-01-02", scf.TimeZone),
	}
	evn.PageInfo = pageInfo

	h.Render(w, tpl, evn, "layout.html", "user.html")
}
