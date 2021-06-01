package admin

import (
	"bytes"
	"context"
	"net/http"
	"testing"

	"github.com/ssb-ngi-pointer/go-ssb-room/v2/roomdb"
	"github.com/ssb-ngi-pointer/go-ssb-room/v2/web/router"
	"github.com/ssb-ngi-pointer/go-ssb-room/v2/web/webassert"
	"github.com/stretchr/testify/assert"
	refs "go.mindeco.de/ssb-refs"
)

func TestDashboardSimple(t *testing.T) {
	ts := newSession(t)
	a := assert.New(t)

	testRef := refs.FeedRef{Algo: "ed25519", ID: bytes.Repeat([]byte{0}, 32)}
	ts.RoomState.AddEndpoint(testRef, nil) // 1 online
	ts.MembersDB.CountReturns(4, nil)      // 4 members
	ts.InvitesDB.CountReturns(3, nil)      // 3 invites
	ts.DeniedKeysDB.CountReturns(2, nil)   // 2 banned

	dashURL := ts.URLTo(router.AdminDashboard)

	html, resp := ts.Client.GetHTML(dashURL)
	a.Equal(http.StatusOK, resp.Code, "wrong HTTP status code")

	a.Equal("1", html.Find("#online-count").Text())
	a.Equal("4", html.Find("#member-count").Text())
	a.Equal("3", html.Find("#invite-count").Text())
	a.Equal("2", html.Find("#denied-count").Text())

	webassert.Localized(t, html, []webassert.LocalizedElement{
		{"title", "AdminDashboardTitle"},
	})
}

// make sure the dashboard renders when someone is connected that is not a member
func TestDashboardWithVisitors(t *testing.T) {
	ts := newSession(t)
	a := assert.New(t)

	visitorRef := refs.FeedRef{Algo: "ed25519", ID: bytes.Repeat([]byte{0}, 32)}
	memberRef := refs.FeedRef{Algo: "ed25519", ID: bytes.Repeat([]byte{1}, 32)}
	ts.RoomState.AddEndpoint(visitorRef, nil)
	ts.RoomState.AddEndpoint(memberRef, nil)

	ts.MembersDB.CountReturns(1, nil)
	// return a member for the member but not for the visitor
	ts.MembersDB.GetByFeedStub = func(ctx context.Context, r refs.FeedRef) (roomdb.Member, error) {
		if r.Equal(&memberRef) {
			return roomdb.Member{ID: 23, Role: roomdb.RoleMember, PubKey: r}, nil
		}
		return roomdb.Member{}, roomdb.ErrNotFound
	}

	dashURL := ts.URLTo(router.AdminDashboard)

	html, resp := ts.Client.GetHTML(dashURL)
	a.Equal(http.StatusOK, resp.Code, "wrong HTTP status code")

	a.Equal("2", html.Find("#online-count").Text())
	a.Equal("1", html.Find("#member-count").Text())

	memberList := html.Find("#connected-list a")
	a.Equal(2, memberList.Length())

	htmlVisitor := memberList.Eq(0)
	a.Equal(visitorRef.Ref(), htmlVisitor.Text())
	gotLink, has := htmlVisitor.Attr("href")
	a.False(has, "visitor should not have a link to a details page: %v", gotLink)

	htmlMember := memberList.Eq(1)
	a.Equal(memberRef.Ref(), htmlMember.Text())
	gotLink, has = htmlMember.Attr("href")
	a.True(has, "member should  have a link to a details page")
	wantLink := ts.URLTo(router.AdminMemberDetails, "id", 23)
	a.Equal(wantLink.String(), gotLink)
}
