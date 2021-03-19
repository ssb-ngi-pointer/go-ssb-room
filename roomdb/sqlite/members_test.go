package sqlite

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ssb-ngi-pointer/go-ssb-room/internal/repo"
	"github.com/ssb-ngi-pointer/go-ssb-room/roomdb"
	"github.com/ssb-ngi-pointer/go-ssb-room/roomdb/sqlite/models"
	"github.com/stretchr/testify/require"
	refs "go.mindeco.de/ssb-refs"
)

func TestMembers(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()

	testRepo := filepath.Join("testrun", t.Name())
	os.RemoveAll(testRepo)

	tr := repo.New(testRepo)

	db, err := Open(tr)
	require.NoError(t, err)

	// broken feed (unknown algo)
	tf := refs.FeedRef{ID: bytes.Repeat([]byte("fooo"), 8), Algo: "nope"}
	_, err = db.Members.Add(ctx, "dont-add-me", tf, roomdb.RoleMember)
	r.Error(err)

	// looks ok at least
	okFeed := refs.FeedRef{ID: bytes.Repeat([]byte("acab"), 8), Algo: refs.RefAlgoFeedSSB1}
	mid, err := db.Members.Add(ctx, "should-add-me", okFeed, roomdb.RoleMember)
	r.NoError(err)

	sqlDB := db.Members.db
	count, err := models.Members().Count(ctx, sqlDB)
	r.NoError(err)
	r.EqualValues(count, 1)

	lst, err := db.Members.List(ctx)
	r.NoError(err)
	r.Len(lst, 1)

	_, yes := db.Members.GetByFeed(ctx, okFeed)
	r.NoError(yes)

	okMember, err := db.Members.GetByFeed(ctx, okFeed)
	r.NoError(err)
	r.Equal(okMember.ID, mid)
	r.Equal(okMember.Nickname, "should-add-me")
	r.Equal(okMember.Role, roomdb.RoleMember)
	r.True(okMember.PubKey.Equal(&okFeed))

	_, yes = db.Members.GetByFeed(ctx, tf)
	r.Error(yes)

	err = db.Members.RemoveFeed(ctx, okFeed)
	r.NoError(err)

	count, err = models.Members().Count(ctx, sqlDB)
	r.NoError(err)
	r.EqualValues(count, 0)

	lst, err = db.Members.List(ctx)
	r.NoError(err)
	r.Len(lst, 0)

	_, yes = db.Members.GetByFeed(ctx, okFeed)
	r.Error(yes)

	r.NoError(db.Close())
}

func TestMembersUnique(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()

	testRepo := filepath.Join("testrun", t.Name())
	os.RemoveAll(testRepo)

	tr := repo.New(testRepo)

	db, err := Open(tr)
	require.NoError(t, err)

	feedA := refs.FeedRef{ID: bytes.Repeat([]byte("1312"), 8), Algo: refs.RefAlgoFeedSSB1}
	_, err = db.Members.Add(ctx, "add-me-first", feedA, roomdb.RoleMember)
	r.NoError(err)

	_, err = db.Members.Add(ctx, "dont-add-me-twice", feedA, roomdb.RoleMember)
	r.Error(err)

	lst, err := db.Members.List(ctx)
	r.NoError(err)
	r.Len(lst, 1)

	r.NoError(db.Close())
}

func TestMembersByID(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()

	testRepo := filepath.Join("testrun", t.Name())
	os.RemoveAll(testRepo)

	tr := repo.New(testRepo)

	db, err := Open(tr)
	require.NoError(t, err)

	feedA := refs.FeedRef{ID: bytes.Repeat([]byte("1312"), 8), Algo: refs.RefAlgoFeedSSB1}
	_, err = db.Members.Add(ctx, "add-me", feedA, roomdb.RoleMember)
	r.NoError(err)

	lst, err := db.Members.List(ctx)
	r.NoError(err)
	r.Len(lst, 1)

	_, yes := db.Members.GetByID(ctx, lst[0].ID)
	r.NoError(yes)

	_, yes = db.Members.GetByID(ctx, 666)
	r.Error(yes)

	err = db.Members.RemoveID(ctx, 666)
	r.Error(err)
	r.EqualError(err, roomdb.ErrNotFound.Error())

	err = db.Members.RemoveID(ctx, lst[0].ID)
	r.NoError(err)

	_, yes = db.Members.GetByID(ctx, lst[0].ID)
	r.Error(yes)
}