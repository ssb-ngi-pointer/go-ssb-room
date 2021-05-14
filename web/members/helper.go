// SPDX-License-Identifier: MIT

// Package members implements helpers for accessing the currently logged in admin or moderator of an active request.
package members

import (
	"context"
	"fmt"
	"net/http"

	"go.mindeco.de/http/auth"
	"go.mindeco.de/http/render"

	"github.com/ssb-ngi-pointer/go-ssb-room/roomdb"
	weberrors "github.com/ssb-ngi-pointer/go-ssb-room/web/errors"
	authWithSSB "github.com/ssb-ngi-pointer/go-ssb-room/web/handlers/auth"
)

type roomMemberContextKeyType string

var roomMemberContextKey roomMemberContextKeyType = "ssb:room:httpcontext:member"

type Middleware func(next http.Handler) http.Handler

// AuthenticateFromContext calls the next http handler if there is a member stored in the context
// otherwise it will call r.Error
func AuthenticateFromContext(r *render.Renderer) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if FromContext(req.Context()) == nil {
				r.Error(w, req, http.StatusUnauthorized, weberrors.ErrNotAuthorized)
				return
			}
			next.ServeHTTP(w, req)
		})
	}
}

// FromContext returns the member or nil if not logged in
func FromContext(ctx context.Context) *roomdb.Member {
	v := ctx.Value(roomMemberContextKey)

	m, ok := v.(*roomdb.Member)
	if !ok {
		return nil
	}

	return m
}

// ContextInjecter returns middleware for injecting a member into the context of the request.
// Retreive it using FromContext(ctx)
func ContextInjecter(mdb roomdb.MembersService, withPassword *auth.Handler, withSSB *authWithSSB.WithSSBHandler) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			var (
				member *roomdb.Member

				errWithPassword, errWithSSB error
			)

			v, errWithPassword := withPassword.AuthenticateRequest(req)
			if errWithPassword == nil {
				mid, ok := v.(int64)
				if !ok {
					next.ServeHTTP(w, req)
					return
				}

				m, err := mdb.GetByID(req.Context(), mid)
				if err != nil {
					next.ServeHTTP(w, req)
					return
				}
				member = &m
			}

			m, errWithSSB := withSSB.AuthenticateRequest(req)
			if errWithSSB == nil {
				member = m
			}

			// if both methods failed, don't update the context
			if errWithPassword != nil && errWithSSB != nil {
				next.ServeHTTP(w, req)
				return
			}

			ctx := context.WithValue(req.Context(), roomMemberContextKey, member)
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	}
}

// TemplateHelpers returns functions to be used with the go.mindeco.de/http/render package.
// Each helper has to return a function twice because the first is evaluated with the request before it gets passed onto html/template's FuncMap.
//
//  {{ is_logged_in }} returns true or false depending on if the user is logged in
//
//  {{ member_has_role "string" }} returns a boolean which confrms wether the member has a certain role (RoleMemeber, RoleAdmin, etc)
//
//  {{ member_is_admin }} is a shortcut for {{ member_has_role "RoleAdmin" }}
//
//  {{ member_is_elevated }} is a shortcut for {{ or member_has_role "RoleAdmin" member_has_role "RoleModerator"}}
//
//  {{ member_can "action" }} returns true if a member can execute a certain action. Actions are "invite" and "remove-denied-key". See allowedActions to add more.
func TemplateHelpers(roomCfg roomdb.RoomConfig) []render.Option {

	return []render.Option{
		render.InjectTemplateFunc("is_logged_in", func(r *http.Request) interface{} {
			no := func() *roomdb.Member { return nil }

			member := FromContext(r.Context())
			if member == nil {
				return no
			}

			yes := func() *roomdb.Member { return member }
			return yes
		}),

		render.InjectTemplateFunc("member_has_role", func(r *http.Request) interface{} {
			no := func(_ string) bool { return false }

			member := FromContext(r.Context())
			if member == nil {
				return no
			}

			return func(has string) bool {
				var r roomdb.Role
				if err := r.UnmarshalText([]byte(has)); err != nil {
					return false
				}
				return member.Role == r
			}
		}),

		render.InjectTemplateFunc("member_is_admin", func(r *http.Request) interface{} {
			no := func() bool { return false }

			member := FromContext(r.Context())
			if member == nil {
				return no
			}

			return func() bool {
				return member.Role == roomdb.RoleAdmin
			}
		}),

		// shorthand for is admin || mod (used for editing notices, managing users, managing aliases)
		render.InjectTemplateFunc("member_is_elevated", func(r *http.Request) interface{} {
			no := func() bool { return false }

			member := FromContext(r.Context())
			if member == nil {
				return no
			}

			return func() bool {
				return member.Role == roomdb.RoleAdmin || member.Role == roomdb.RoleModerator
			}
		}),

		render.InjectTemplateFunc("member_can", func(r *http.Request) interface{} {
			// evaluate member and privacy mode first to reduce some churn for multiple calls to this helper
			// works fine since they are not changing during one request
			member := FromContext(r.Context())
			if member == nil {
				return func(_ string) bool { return false }
			}

			pm, err := roomCfg.GetPrivacyMode(r.Context())
			if err != nil {
				return func(_ string) (bool, error) { return false, err }
			}

			// now return the template func which closes over pm and the member
			return func(what string) (bool, error) {
				actionCheck, has := allowedActionsMap[what]
				if !has {
					return false, fmt.Errorf("unrecognized action: %s", what)
				}

				return actionCheck(pm, member.Role), nil
			}
		}),
	}
}

// AllowedFunc returns true if a member role is allowed to do a thing under the passed mode
type AllowedFunc func(mode roomdb.PrivacyMode, role roomdb.Role) bool

// AllowedActions exposes check function by name. It exists to protected against changes of the map
func AllowedActions(name string) (AllowedFunc, bool) {
	fn, has := allowedActionsMap[name]
	return fn, has
}

// member actions
const (
	ActionInviteMember     = "invite"
	ActionChangeDeniedKeys = "change-denied-keys"
	ActionRemoveMember     = "remove-member"
	ActionChangeNotice     = "change-notice"
)

var allowedActionsMap = map[string]AllowedFunc{
	ActionInviteMember: func(pm roomdb.PrivacyMode, role roomdb.Role) bool {
		switch pm {
		case roomdb.ModeOpen:
			return true
		case roomdb.ModeCommunity:
			return role > roomdb.RoleUnknown && role <= roomdb.RoleAdmin
		case roomdb.ModeRestricted:
			return role == roomdb.RoleAdmin || role == roomdb.RoleModerator
		default:
			return false
		}
	},

	ActionChangeDeniedKeys: func(pm roomdb.PrivacyMode, role roomdb.Role) bool {
		switch pm {
		case roomdb.ModeCommunity:
			return true
		case roomdb.ModeOpen:
			fallthrough
		case roomdb.ModeRestricted:
			return role == roomdb.RoleAdmin || role == roomdb.RoleModerator
		default:
			return false
		}
	},

	ActionRemoveMember: func(_ roomdb.PrivacyMode, role roomdb.Role) bool {
		return role == roomdb.RoleAdmin || role == roomdb.RoleModerator
	},

	ActionChangeNotice: func(pm roomdb.PrivacyMode, role roomdb.Role) bool {
		switch pm {
		case roomdb.ModeCommunity:
			return true
		case roomdb.ModeOpen:
			fallthrough
		case roomdb.ModeRestricted:
			return role == roomdb.RoleAdmin || role == roomdb.RoleModerator
		default:
			return false
		}
	},
}

// CheckAllowed retreives the member from the passed context and lookups the current privacy mode from the passed cfg to determain if the action is okay or not.
// If it's not it returns an error. For convenience it also returns the member if the action is okay.
func CheckAllowed(ctx context.Context, cfg roomdb.RoomConfig, action string) (*roomdb.Member, error) {
	member := FromContext(ctx)
	if member == nil {
		return nil, weberrors.ErrNotAuthorized
	}

	pm, err := cfg.GetPrivacyMode(ctx)
	if err != nil {
		return nil, err
	}

	allowed, ok := AllowedActions(action)
	if !ok {
		return nil, fmt.Errorf("unknown action: %s: %w", action, weberrors.ErrNotAuthorized)
	}

	if !allowed(pm, member.Role) {
		return nil, weberrors.ErrNotAuthorized
	}

	return member, nil
}
