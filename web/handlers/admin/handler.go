// SPDX-License-Identifier: MIT

// Package admin implements the dashboard for admins and moderators to change and control aspects of the room.
// Including aliases, allow/deny list managment, invites and settings of the room.
package admin

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/vcraescu/go-paginator/v2"
	"github.com/vcraescu/go-paginator/v2/adapter"
	"github.com/vcraescu/go-paginator/v2/view"
	"go.mindeco.de/http/render"

	"github.com/ssb-ngi-pointer/go-ssb-room/internal/network"
	"github.com/ssb-ngi-pointer/go-ssb-room/roomdb"
	"github.com/ssb-ngi-pointer/go-ssb-room/roomstate"
	"github.com/ssb-ngi-pointer/go-ssb-room/web"
	weberrors "github.com/ssb-ngi-pointer/go-ssb-room/web/errors"
	"github.com/ssb-ngi-pointer/go-ssb-room/web/i18n"
	"github.com/ssb-ngi-pointer/go-ssb-room/web/router"
)

// HTMLTemplates define the list of files the template system should load.
var HTMLTemplates = []string{
	"admin/dashboard.tmpl",
	"admin/menu.tmpl",

	"admin/settings.tmpl",

	"admin/aliases-revoke-confirm.tmpl",

	"admin/denied-keys.tmpl",
	"admin/denied-keys-remove-confirm.tmpl",

	"admin/invite-list.tmpl",
	"admin/invite-revoke-confirm.tmpl",
	"admin/invite-created.tmpl",

	"admin/notice-edit.tmpl",

	"admin/member.tmpl",
	"admin/member-list.tmpl",
	"admin/members-remove-confirm.tmpl",
	"admin/members-show-password-reset-token.tmpl",
}

// Databases is an option struct that encapsulates the required database services
type Databases struct {
	Aliases       roomdb.AliasesService
	AuthFallback  roomdb.AuthFallbackService
	Config        roomdb.RoomConfig
	DeniedKeys    roomdb.DeniedKeysService
	Invites       roomdb.InvitesService
	Notices       roomdb.NoticesService
	Members       roomdb.MembersService
	PinnedNotices roomdb.PinnedNoticesService
}

// Handler supplies the elevated access pages to known users.
// It is not registering on the mux router like other pages to clean up the authorize flow.
func Handler(
	netInfo network.ServerEndpointDetails,
	r *render.Renderer,
	roomState *roomstate.Manager,
	fh *weberrors.FlashHelper,
	locHelper *i18n.Helper,
	dbs Databases,
) http.Handler {
	mux := &http.ServeMux{}

	urlTo := web.NewURLTo(router.CompleteApp(), netInfo)

	var dashboardHandler = dashboardHandler{
		r:       r,
		flashes: fh,
		netInfo: netInfo,

		dbs:       dbs,
		roomState: roomState,
	}
	mux.HandleFunc("/dashboard", r.HTML("admin/dashboard.tmpl", dashboardHandler.overview))

	var sh = settingsHandler{
		r:     r,
		urlTo: urlTo,
		db:    dbs.Config,
		loc:   locHelper,
	}
	mux.HandleFunc("/settings", r.HTML("admin/settings.tmpl", sh.overview))
	mux.HandleFunc("/settings/set-privacy", sh.setPrivacy)
	mux.HandleFunc("/settings/set-language", sh.setLanguage)

	mux.HandleFunc("/menu", r.HTML("admin/menu.tmpl", func(w http.ResponseWriter, req *http.Request) (interface{}, error) {
		return map[string]interface{}{}, nil
	}))

	var ah = aliasesHandler{
		r:       r,
		flashes: fh,

		db: dbs.Aliases,
	}
	mux.HandleFunc("/aliases/revoke/confirm", r.HTML("admin/aliases-revoke-confirm.tmpl", ah.revokeConfirm))
	mux.HandleFunc("/aliases/revoke", ah.revoke)

	var dh = deniedKeysHandler{
		r:       r,
		flashes: fh,

		db: dbs.DeniedKeys,

		roomCfg: dbs.Config,
	}
	mux.HandleFunc("/denied", r.HTML("admin/denied-keys.tmpl", dh.overview))
	mux.HandleFunc("/denied/add", dh.add)
	mux.HandleFunc("/denied/remove/confirm", r.HTML("admin/denied-keys-remove-confirm.tmpl", dh.removeConfirm))
	mux.HandleFunc("/denied/remove", dh.remove)

	var mh = membersHandler{
		r:       r,
		flashes: fh,
		urlTo:   urlTo,
		netInfo: netInfo,

		db: dbs.Members,

		fallbackAuthDB: dbs.AuthFallback,
		roomCfgDB:      dbs.Config,
	}
	mux.HandleFunc("/member", r.HTML("admin/member.tmpl", mh.details))
	mux.HandleFunc("/members", r.HTML("admin/member-list.tmpl", mh.overview))
	mux.HandleFunc("/members/add", mh.add)
	mux.HandleFunc("/members/change-role", mh.changeRole)
	mux.HandleFunc("/members/remove/confirm", r.HTML("admin/members-remove-confirm.tmpl", mh.removeConfirm))
	mux.HandleFunc("/members/remove", mh.remove)
	mux.HandleFunc("/members/create-fallback-reset-link", r.HTML("admin/members-show-password-reset-token.tmpl", mh.createPasswordResetToken))

	var ih = invitesHandler{
		r:       r,
		flashes: fh,
		urlTo:   urlTo,

		db:     dbs.Invites,
		config: dbs.Config,
	}

	mux.HandleFunc("/invites", r.HTML("admin/invite-list.tmpl", ih.overview))
	mux.HandleFunc("/invites/create", r.HTML("admin/invite-created.tmpl", ih.create))
	mux.HandleFunc("/invites/revoke/confirm", r.HTML("admin/invite-revoke-confirm.tmpl", ih.revokeConfirm))
	mux.HandleFunc("/invites/revoke", ih.revoke)

	var nh = noticeHandler{
		r:       r,
		urlTo:   urlTo,
		flashes: fh,

		noticeDB: dbs.Notices,
		pinnedDB: dbs.PinnedNotices,
		roomCfg:  dbs.Config,
	}
	mux.Handle("/notice/edit", r.HTML("admin/notice-edit.tmpl", nh.edit))
	mux.Handle("/notice/translation/draft", r.HTML("admin/notice-edit.tmpl", nh.draftTranslation))
	mux.Handle("/notice/translation/add", http.HandlerFunc(nh.addTranslation))
	mux.Handle("/notice/save", http.HandlerFunc(nh.save))

	// path:/ matches everything that isn't registerd (ie. its the "Not Found handler")
	mux.HandleFunc("/", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		r.Error(rw, req, 404, weberrors.PageNotFound{Path: req.URL.Path})
	}))

	return customStripPrefix("/admin", mux)
}

// how many elements does a paginated page have by default
const defaultPageSize = 20

// paginate receives the total slice and it's length/count, a URL query for the 'limit' and which 'page'.
//
// The members of the map are:
//	Entries: the paginated slice
//	Count: the total number of the whole, unpaginated list
//	FirstInView: a bool thats true if you render the first page
//	LastInView: a bool thats true if you render the last page
//	Paginator and View: helpers for rendering the page accessor (see github.com/vcraescu/go-paginator)
//
// TODO: we could return a struct instead but then need to re-think how we embedd it into all the pages where we need it.
//  Maybe renderData["Pages"] = paginatedData
func paginate(total interface{}, count int, qry url.Values) (map[string]interface{}, error) {
	pageSize, err := strconv.Atoi(qry.Get("limit"))
	if err != nil {
		pageSize = defaultPageSize
	}

	page, err := strconv.Atoi(qry.Get("page"))
	if err != nil || page < 1 {
		page = 1
	}

	paginator := paginator.New(adapter.NewSliceAdapter(total), pageSize)
	paginator.SetPage(page)

	var entries []interface{}
	if err = paginator.Results(&entries); err != nil {
		return nil, fmt.Errorf("paginator failed with %w", err)
	}

	view := view.New(paginator)
	pagesSlice, err := view.Pages()
	if err != nil {
		return nil, fmt.Errorf("paginator view.Pages failed with %w", err)
	}
	if len(pagesSlice) == 0 {
		pagesSlice = []int{1}
	}

	last, err := view.Last()
	if err != nil {
		return nil, fmt.Errorf("paginator view.Last failed with %w", err)
	}

	return map[string]interface{}{
		"Entries":     entries,
		"Count":       count,
		"Paginator":   paginator,
		"View":        view,
		"FirstInView": pagesSlice[0] == 1,
		"LastInView":  pagesSlice[len(pagesSlice)-1] == last,
	}, nil
}

// trim prefix if exists (workaround for named router problem)
func customStripPrefix(prefix string, h http.Handler) http.Handler {
	if prefix == "" {
		return h
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(r.URL.Path, prefix)
		rp := strings.TrimPrefix(r.URL.RawPath, prefix)
		if len(p) < len(r.URL.Path) && (r.URL.RawPath == "" || len(rp) < len(r.URL.RawPath)) {
			r2 := new(http.Request)
			*r2 = *r
			r2.URL = new(url.URL)
			*r2.URL = *r.URL
			r2.URL.Path = p
			r2.URL.RawPath = rp
			h.ServeHTTP(w, r2)
		} else {
			h.ServeHTTP(w, r)
		}
	})
}
