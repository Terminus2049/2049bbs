package controller

import (
	"net/http"
	"strconv"
	"text/template"

	"github.com/terminus2049/2049bbs/model"
	"goji.io/pat"
)

func (h *BaseHandler) FeedHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/atom+xml; charset=utf-8")

	scf := h.App.Cf.Site

	var feed = `<?xml version="1.0" encoding="utf-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
	<title>` + scf.Name + `</title>
	<link rel="self" type="application/atom+xml" href="` + scf.MainDomain + `/feed"/>
	<link rel="hub" href="https://pubsubhubbub.appspot.com"/>
	<updated>{{.Update}}</updated>
	<id>` + scf.MainDomain + `/feed</id>
	<author>
		<name>` + scf.Name + `</name>
	</author>
	{{range $_, $item := .Items}}
	<entry>
		<title>{{$item.Title}}</title>
		<id>` + scf.MainDomain + `/t/{{$item.Id}}</id>
		<link rel="alternate" type="text/html" href="` + scf.MainDomain + `/t/{{$item.Id}}" />
		<published>{{$item.AddTimeFmt}}</published>
		<updated>{{$item.EditTimeFmt}}</updated>
		<content type="text/plain">
		  {{$item.Cname}} - {{$item.Name}} - {{$item.Des}}
		</content>
    </entry>
	{{end}}
</feed>
`

	db := h.App.Db

	items := model.ArticleFeedList(db, 20, h.App.Cf.Site.TimeZone)

	var upDate string
	if len(items) > 0 {
		upDate = items[0].AddTimeFmt
	}

	t := template.Must(template.New("feed").Parse(feed))
	t.Execute(w, struct {
		Update string
		Items  []model.ArticleFeedListItem
	}{
		Update: upDate,
		Items:  items,
	})
}

func (h *BaseHandler) FeedCategoryHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/atom+xml; charset=utf-8")

	scf := h.App.Cf.Site

	var feed = `<?xml version="1.0" encoding="utf-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
	<title>` + scf.Name + `</title>
	<link rel="self" type="application/atom+xml" href="` + scf.MainDomain + `/feed"/>
	<link rel="hub" href="https://pubsubhubbub.appspot.com"/>
	<updated>{{.Update}}</updated>
	<id>` + scf.MainDomain + `/feed</id>
	<author>
		<name>` + scf.Name + `</name>
	</author>
	{{range $_, $item := .Items}}
	<entry>
		<title>{{$item.Title}}</title>
		<id>` + scf.MainDomain + `/t/{{$item.Id}}</id>
		<link rel="alternate" type="text/html" href="` + scf.MainDomain + `/t/{{$item.Id}}" />
		<published>{{$item.AddTimeFmt}}</published>
		<updated>{{$item.EditTimeFmt}}</updated>
		<content type="text/plain">
		  {{$item.Cname}} - {{$item.Name}} - {{$item.Des}}
		</content>
    </entry>
	{{end}}
</feed>
`

	cid := pat.Param(r, "cid")
	_, err := strconv.Atoi(cid)
	if err != nil {
		w.Write([]byte(`{"retcode":400,"retmsg":"cid type err"}`))
		return
	}

	db := h.App.Db

	items := model.ArticleFeedCategoryList(db, cid, 20, h.App.Cf.Site.TimeZone)

	var upDate string
	if len(items) > 0 {
		upDate = items[0].AddTimeFmt
	}

	t := template.Must(template.New("feed").Parse(feed))
	t.Execute(w, struct {
		Update string
		Items  []model.ArticleFeedListItem
	}{
		Update: upDate,
		Items:  items,
	})
}
