// SPDX-License-Identifier: MIT

package handlers

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/gorilla/mux"
	"go.mindeco.de/http/tester"
	"go.mindeco.de/logging/logtest"
	refs "go.mindeco.de/ssb-refs"

	"github.com/ssb-ngi-pointer/go-ssb-room/internal/network/mocked"
	"github.com/ssb-ngi-pointer/go-ssb-room/internal/repo"
	"github.com/ssb-ngi-pointer/go-ssb-room/roomdb"
	"github.com/ssb-ngi-pointer/go-ssb-room/roomdb/mockdb"
	"github.com/ssb-ngi-pointer/go-ssb-room/roomstate"
	"github.com/ssb-ngi-pointer/go-ssb-room/web/i18n"
	"github.com/ssb-ngi-pointer/go-ssb-room/web/router"
)

type testSession struct {
	Mux    *http.ServeMux
	Client *tester.Tester
	Router *mux.Router

	// mocked dbs
	AuthDB         *mockdb.FakeAuthWithSSBService
	AuthFallbackDB *mockdb.FakeAuthFallbackService
	AliasesDB      *mockdb.FakeAliasesService
	MembersDB      *mockdb.FakeMembersService
	InvitesDB      *mockdb.FakeInvitesService
	PinnedDB       *mockdb.FakePinnedNoticesService
	NoticeDB       *mockdb.FakeNoticesService

	RoomState *roomstate.Manager

	MockedEndpoints *mocked.FakeEndpoints

	NetworkInfo NetworkInfo
}

var testI18N = justTheKeys()

func setup(t *testing.T) *testSession {
	var ts testSession

	testRepoPath := filepath.Join("testrun", t.Name())
	os.RemoveAll(testRepoPath)
	testRepo := repo.New(testRepoPath)

	testOverride := filepath.Join(testRepo.GetPath("i18n"), "active.en.toml")
	os.MkdirAll(filepath.Dir(testOverride), 0700)
	err := ioutil.WriteFile(testOverride, testI18N, 0700)
	if err != nil {
		t.Fatal(err)
	}

	ts.AuthDB = new(mockdb.FakeAuthWithSSBService)
	ts.AuthFallbackDB = new(mockdb.FakeAuthFallbackService)
	ts.AliasesDB = new(mockdb.FakeAliasesService)
	ts.MembersDB = new(mockdb.FakeMembersService)
	ts.InvitesDB = new(mockdb.FakeInvitesService)
	ts.PinnedDB = new(mockdb.FakePinnedNoticesService)
	defaultNotice := &roomdb.Notice{
		Title:   "Default Notice Title",
		Content: "Default Notice Content",
	}
	ts.PinnedDB.GetReturns(defaultNotice, nil)
	ts.NoticeDB = new(mockdb.FakeNoticesService)

	ts.MockedEndpoints = new(mocked.FakeEndpoints)

	ts.NetworkInfo = NetworkInfo{
		Domain:     "localhost",
		PortMUXRPC: 8008,
		PortHTTPS:  443,

		RoomID: refs.FeedRef{
			ID:   bytes.Repeat([]byte("test"), 8),
			Algo: refs.RefAlgoFeedSSB1,
		},
	}

	log, _ := logtest.KitLogger("complete", t)
	ctx := context.TODO()
	ts.RoomState = roomstate.NewManager(ctx, log)

	ts.Router = router.CompleteApp()

	h, err := New(
		log,
		testRepo,
		ts.NetworkInfo,
		ts.RoomState,
		ts.MockedEndpoints,
		Databases{
			Aliases:       ts.AliasesDB,
			AuthFallback:  ts.AuthFallbackDB,
			Members:       ts.MembersDB,
			Invites:       ts.InvitesDB,
			Notices:       ts.NoticeDB,
			PinnedNotices: ts.PinnedDB,
		},
	)
	if err != nil {
		t.Fatal("setup: handler init failed:", err)
	}

	ts.Mux = http.NewServeMux()
	ts.Mux.Handle("/", h)
	ts.Client = tester.New(ts.Mux, t)

	return &ts
}

// auto generate from defaults a list of Label = "Label"
// must keep order of input intact
// (at least all the globals before starting with nested plurals)
// also replaces 'one' and 'other' in plurals
func justTheKeys() []byte {
	f, err := i18n.Defaults.Open("active.en.toml")
	if err != nil {
		panic(err)
	}
	justAMap := make(map[string]interface{})
	md, err := toml.DecodeReader(f, &justAMap)
	if err != nil {
		panic(err)
	}

	var buf = &bytes.Buffer{}

	// if we don't produce the same order as the input
	// (in go maps are ALWAYS random access when ranged over)
	// nested keys (such as plural form) will mess up the global level...
	for _, k := range md.Keys() {
		key := k.String()
		val, has := justAMap[key]
		if !has {
			// fmt.Println("i18n test warning:", key, "not unmarshaled")
			continue
		}

		switch tv := val.(type) {

		case string:
			fmt.Fprintf(buf, "%s = \"%s\"\n", key, key)

		case map[string]interface{}:
			// fmt.Println("i18n test warning: custom map for ", key)

			fmt.Fprintf(buf, "\n[%s]\n", key)
			// replace "one" and "other" keys
			// with  Label and LabelPlural
			tv["one"] = key + "Singular"
			tv["other"] = key + "Plural"
			toml.NewEncoder(buf).Encode(tv)
			fmt.Fprintln(buf)

		default:
			panic(fmt.Sprintf("unhandled toml structure under %s: %T\n", key, val))
		}
	}

	return buf.Bytes()
}
