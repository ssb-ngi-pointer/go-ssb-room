// SPDX-License-Identifier: MIT

package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/csrf"
	"github.com/gorilla/sessions"
	"go.mindeco.de/http/auth"
	"go.mindeco.de/http/render"
	"go.mindeco.de/logging"

	"github.com/ssb-ngi-pointer/go-ssb-room/admindb"
	"github.com/ssb-ngi-pointer/go-ssb-room/internal/repo"
	"github.com/ssb-ngi-pointer/go-ssb-room/roomstate"
	"github.com/ssb-ngi-pointer/go-ssb-room/web"
	"github.com/ssb-ngi-pointer/go-ssb-room/web/handlers/admin"
	roomsAuth "github.com/ssb-ngi-pointer/go-ssb-room/web/handlers/auth"
	"github.com/ssb-ngi-pointer/go-ssb-room/web/handlers/news"
	"github.com/ssb-ngi-pointer/go-ssb-room/web/i18n"
	"github.com/ssb-ngi-pointer/go-ssb-room/web/router"
)

// New initializes the whole web stack for rooms, with all the sub-modules and routing.
func New(
	logger logging.Interface,
	repo repo.Interface,
	roomState *roomstate.Manager,
	as admindb.AuthWithSSBService,
	fs admindb.AuthFallbackService,
	al admindb.AllowListService,
) (http.Handler, error) {
	m := router.CompleteApp()

	locHelper, err := i18n.New(repo)
	if err != nil {
		return nil, err
	}

	var a *auth.Handler

	r, err := render.New(web.Templates,
		render.SetLogger(logger),
		render.BaseTemplates("/base.tmpl"),
		render.AddTemplates(concatTemplates(
			[]string{
				"/landing/index.tmpl",
				"/landing/about.tmpl",
				"/error.tmpl",
			},
			news.HTMLTemplates,
			roomsAuth.HTMLTemplates,
			admin.HTMLTemplates,
		)...),
		render.FuncMap(web.TemplateFuncs(m)),
		// TODO: move these to the i18n helper pkg
		render.InjectTemplateFunc("i18npl", func(r *http.Request) interface{} {
			loc := i18n.LocalizerFromRequest(locHelper, r)
			return loc.LocalizePlurals
		}),
		render.InjectTemplateFunc("i18n", func(r *http.Request) interface{} {
			loc := i18n.LocalizerFromRequest(locHelper, r)
			return loc.LocalizeSimple
		}),
		render.InjectTemplateFunc("is_logged_in", func(r *http.Request) interface{} {
			no := func() *admindb.User { return nil }

			v, err := a.AuthenticateRequest(r)
			if err != nil {
				return no
			}

			uid, ok := v.(int64)
			if !ok {
				// TODO: hook up logging
				fmt.Fprintf(os.Stderr, "warning: not the expected ID type: %T\n", v)
				return no
			}

			user, err := fs.GetByID(r.Context(), uid)
			if err != nil {
				return no
			}

			yes := func() *admindb.User { return user }
			return yes
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("web Handler: failed to create renderer: %w", err)
	}

	cookieCodec, err := web.LoadOrCreateCookieSecrets(repo)
	if err != nil {
		return nil, err
	}

	store := &sessions.CookieStore{
		Codecs: cookieCodec,
		Options: &sessions.Options{
			Path:   "/",
			MaxAge: 2 * 60 * 60, // two hours in seconds  // TODO: configure
		},
	}

	// TODO: this is just the error handler for http/auth, not render
	authErrH := func(rw http.ResponseWriter, req *http.Request, err error, code int) {
		var ih = i18n.LocalizerFromRequest(locHelper, req)

		// default, unlocalized message
		msg := err.Error()

		// localize some specific error messages
		var (
			aa admindb.ErrAlreadyAdded
		)
		switch {
		case err == auth.ErrBadLogin:
			msg = ih.LocalizeSimple("AuthErrorBadLogin")

		case errors.Is(err, admindb.ErrNotFound):
			msg = ih.LocalizeSimple("ErrorNotFound")

		case errors.As(err, &aa):
			msg = ih.LocalizeSimple("ErrorAlreadyAdded")
		}

		r.HTML("/error.tmpl", func(rw http.ResponseWriter, req *http.Request) (interface{}, error) {
			return errorTemplateData{
				Err: msg,
				// TODO: localize?
				Status:     http.StatusText(code),
				StatusCode: code,
			}, nil
		}).ServeHTTP(rw, req)
	}

	notAuthorizedH := r.HTML("/error.tmpl", func(rw http.ResponseWriter, req *http.Request) (interface{}, error) {
		statusCode := http.StatusUnauthorized
		rw.WriteHeader(statusCode)
		return errorTemplateData{
			statusCode,
			"Unauthorized",
			"you are not authorized to access the requested site",
		}, nil
	})

	a, err = auth.NewHandler(fs,
		auth.SetStore(store),
		auth.SetErrorHandler(authErrH),
		auth.SetNotAuthorizedHandler(notAuthorizedH),
		auth.SetLifetime(2*time.Hour), // TODO: configure
	)
	if err != nil {
		return nil, fmt.Errorf("web Handler: failed to init fallback auth system: %w", err)
	}

	// Cross Site Request Forgery prevention middleware
	csrfKey, err := web.LoadOrCreateCSRFSecret(repo)
	if err != nil {
		return nil, err
	}

	CSRF := csrf.Protect(csrfKey,
		csrf.ErrorHandler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			err := csrf.FailureReason(req)
			// TODO: localize error?
			r.Error(w, req, http.StatusForbidden, err)
		})),
	)

	// this router is a bit of a qurik
	// TODO: explain problem between gorilla/mux named routers and authentication
	mainMux := &http.ServeMux{}

	// hookup handlers to the router
	news.Handler(m, r)
	roomsAuth.Handler(m, r, a)

	adminHandler := a.Authenticate(admin.Handler(r, roomState, al))
	mainMux.Handle("/admin/", adminHandler)

	m.Get(router.CompleteIndex).Handler(r.StaticHTML("/landing/index.tmpl"))
	m.Get(router.CompleteAbout).Handler(r.StaticHTML("/landing/about.tmpl"))

	m.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(web.Assets)))

	m.NotFoundHandler = r.HTML("/error.tmpl", func(rw http.ResponseWriter, req *http.Request) (interface{}, error) {
		rw.WriteHeader(http.StatusNotFound)
		msg := i18n.LocalizerFromRequest(locHelper, req).LocalizeSimple("PageNotFound")
		return errorTemplateData{http.StatusNotFound, "Not Found", msg}, nil
	})

	mainMux.Handle("/", m)

	// apply middleware
	var finalHandler http.Handler = mainMux
	finalHandler = logging.InjectHandler(logger)(finalHandler)
	finalHandler = CSRF(finalHandler)

	if web.Production {
		return finalHandler, nil
	}

	return r.GetReloader()(finalHandler), nil
}

// utils

type errorTemplateData struct {
	StatusCode int
	Status     string
	Err        string
}

func concatTemplates(lst ...[]string) []string {
	var catted []string

	for _, tpls := range lst {
		for _, t := range tpls {
			catted = append(catted, t)
		}

	}
	return catted
}
