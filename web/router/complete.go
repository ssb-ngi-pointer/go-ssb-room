// SPDX-License-Identifier: MIT

package router

import (
	"github.com/gorilla/mux"
)

// constant names for the named routes
const (
	CompleteIndex = "complete:index"
	CompleteAbout = "complete:about"

	CompleteNoticeShow = "complete:notice:show"
	CompleteNoticeList = "complete:notice:list"

	CompleteAliasResolve = "complete:alias:resolve"

	CompleteInviteAccept  = "complete:invite:accept"
	CompleteInviteConsume = "complete:invite:consume"
)

// CompleteApp constructs a mux.Router containing the routes for batch Complete html frontend
func CompleteApp() *mux.Router {
	m := mux.NewRouter()

	Auth(m)
	Admin(m.PathPrefix("/admin").Subrouter())

	m.Path("/").Methods("GET").Name(CompleteIndex)
	m.Path("/about").Methods("GET").Name(CompleteAbout)

	m.Path("/alias/{alias}").Methods("GET").Name(CompleteAliasResolve)

	m.Path("/invite/accept").Methods("GET").Name(CompleteInviteAccept)
	m.Path("/invite/consume").Methods("POST").Name(CompleteInviteConsume)

	m.Path("/notice/show").Methods("GET").Name(CompleteNoticeShow)
	m.Path("/notice/list").Methods("GET").Name(CompleteNoticeList)

	return m
}
