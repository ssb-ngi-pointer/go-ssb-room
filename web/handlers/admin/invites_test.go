package admin

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ssb-ngi-pointer/go-ssb-room/roomdb"
	"github.com/ssb-ngi-pointer/go-ssb-room/web/router"
	"github.com/ssb-ngi-pointer/go-ssb-room/web/webassert"
)

func TestInvitesOverview(t *testing.T) {
	ts := newSession(t)
	a := assert.New(t)

	testUser := roomdb.Member{ID: 23}

	lst := []roomdb.Invite{
		{ID: 1, CreatedBy: testUser},
		{ID: 2, CreatedBy: testUser},
		{ID: 3, CreatedBy: testUser},
	}
	ts.InvitesDB.ListReturns(lst, nil)

	invitesOverviewURL := ts.URLTo(router.AdminInvitesOverview)

	html, resp := ts.Client.GetHTML(invitesOverviewURL)
	a.Equal(http.StatusOK, resp.Code, "wrong HTTP status code")

	webassert.Localized(t, html, []webassert.LocalizedElement{
		{"#welcome", "AdminInvitesWelcome"},
		{"title", "AdminInvitesTitle"},
		{"#invite-list-count", "AdminInvitesCountPlural"},
	})

	// devided by two because there is one for wide and one for slim/mobile
	trSelector := "#the-table-rows tr"
	a.EqualValues(3, html.Find(trSelector).Length()/2, "wrong number of entries on the table (plural)")

	lst = []roomdb.Invite{
		{ID: 666, CreatedBy: testUser},
	}
	ts.InvitesDB.ListReturns(lst, nil)

	html, resp = ts.Client.GetHTML(invitesOverviewURL)
	a.Equal(http.StatusOK, resp.Code, "wrong HTTP status code")

	webassert.Localized(t, html, []webassert.LocalizedElement{
		{"#welcome", "AdminInvitesWelcome"},
		{"title", "AdminInvitesTitle"},
		{"#invite-list-count", "AdminInvitesCountSingular"},
	})

	elems := html.Find(trSelector)
	a.EqualValues(1, elems.Length()/2, "wrong number of entries on the table (signular)")

	// check for link to remove confirm link
	link, yes := elems.Find("a").Attr("href")
	a.True(yes, "a-tag has href attribute")
	wantURL := ts.URLTo(router.AdminInvitesRevokeConfirm, "id", 666)
	a.Equal(wantURL.String(), link)

	testInviteButtonDisabled := func(shouldBeDisabled bool) {
		html, resp = ts.Client.GetHTML(invitesOverviewURL)
		a.Equal(http.StatusOK, resp.Code, "wrong HTTP status code")
		inviteButton := html.Find("#create-invite button")
		_, disabled := inviteButton.Attr("disabled")
		a.EqualValues(shouldBeDisabled, disabled, "invite button should be disabled")
	}

	// member, mod, admin should all be able to invite in ModeCommunity
	ts.ConfigDB.GetPrivacyModeReturns(roomdb.ModeCommunity, nil)
	ts.User = roomdb.Member{
		ID:   1234,
		Role: roomdb.RoleAdmin,
	}
	testInviteButtonDisabled(false)
	ts.User = roomdb.Member{
		ID:   7331,
		Role: roomdb.RoleModerator,
	}
	testInviteButtonDisabled(false)
	ts.User = roomdb.Member{
		ID:   9001,
		Role: roomdb.RoleMember,
	}
	testInviteButtonDisabled(false)

	// mod and admin should be able to invite, member should not
	ts.ConfigDB.GetPrivacyModeReturns(roomdb.ModeRestricted, nil)
	ts.User = roomdb.Member{
		ID:   1234,
		Role: roomdb.RoleAdmin,
	}
	testInviteButtonDisabled(false)
	ts.User = roomdb.Member{
		ID:   7331,
		Role: roomdb.RoleModerator,
	}
	testInviteButtonDisabled(false)
	ts.User = roomdb.Member{
		ID:   9001,
		Role: roomdb.RoleMember,
	}
	testInviteButtonDisabled(true)
}

func TestInvitesCreateForm(t *testing.T) {
	ts := newSession(t)
	a := assert.New(t)

	overviewURL := ts.URLTo(router.AdminInvitesOverview)

	html, resp := ts.Client.GetHTML(overviewURL)
	a.Equal(http.StatusOK, resp.Code, "wrong HTTP status code")

	webassert.Localized(t, html, []webassert.LocalizedElement{
		{"#welcome", "AdminInvitesWelcome"},
		{"title", "AdminInvitesTitle"},
	})

	formSelection := html.Find("form#create-invite")
	a.EqualValues(1, formSelection.Length())

	method, ok := formSelection.Attr("method")
	a.True(ok, "form has method set")
	a.Equal("POST", method)

	action, ok := formSelection.Attr("action")
	a.True(ok, "form has action set")

	addURL := ts.URLTo(router.AdminInvitesCreate)
	a.Equal(addURL.String(), action)
}

func TestInvitesCreate(t *testing.T) {
	ts := newSession(t)
	a := assert.New(t)
	r := require.New(t)

	urlRemove := ts.URLTo(router.AdminInvitesCreate)

	testInvite := "your-fake-test-invite"
	ts.InvitesDB.CreateReturns(testInvite, nil)

	rec := ts.Client.PostForm(urlRemove, url.Values{})
	a.Equal(http.StatusOK, rec.Code)

	r.Equal(1, ts.InvitesDB.CreateCallCount(), "expected one invites.Create call")
	_, userID := ts.InvitesDB.CreateArgsForCall(0)
	a.EqualValues(ts.User.ID, userID)

	doc, err := goquery.NewDocumentFromReader(rec.Body)
	require.NoError(t, err, "failed to parse response")

	webassert.Localized(t, doc, []webassert.LocalizedElement{
		{"title", "AdminInviteCreatedTitle"},
		{"#welcome", "AdminInviteCreatedTitle" + "AdminInviteCreatedInstruct"},
	})

	wantURL := ts.URLTo(router.CompleteInviteFacade, "token", testInvite)

	shownLink := doc.Find("#invite-facade-link").Text()
	a.Equal(wantURL.String(), shownLink)
}
