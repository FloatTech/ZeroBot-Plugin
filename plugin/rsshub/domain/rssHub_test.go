package domain

import (
	"context"
	"encoding/json"
	"testing"
)

func TestNewRssDomain(t *testing.T) {
	dm, err := newRssDomain("rsshub.db")
	if err != nil {
		t.Fatal(err)
		return
	}
	if dm == nil {
		t.Fatal("domain is nil")
	}
}

//var testRssHubChannelUrl = "https://rsshub.rssforever.com/bangumi/tv/calendar/today"

var dm, _ = newRssDomain("rsshub.db")

func TestSub(t *testing.T) {
	testCases := []struct {
		name     string
		feedLink string
		gid      int64
	}{
		{
			name:     "test1",
			feedLink: "/bangumi/tv/calendar/today",
			gid:      99,
		},
		{
			name:     "test2",
			feedLink: "/go-weekly",
			gid:      99,
		},
		{
			name:     "test3",
			feedLink: "/go-weekly",
			gid:      123,
		},
		{
			name:     "test3",
			feedLink: "/go-weekly",
			gid:      321,
		},
		{
			name:     "test3",
			feedLink: "/go-weekly",
			gid:      4123,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			channel, ifExisted, ifSub, err := dm.Subscribe(ctx, tc.gid, tc.feedLink)
			if err != nil {
				t.Fatal(err)
				return
			}
			t.Logf("[TEST] add sub res: %+v,%+v,%+v\n", channel, ifExisted, ifSub)
			res, ext, err := dm.storage.GetIfExistedSubscribe(ctx, tc.gid, tc.feedLink)
			if err != nil {
				t.Fatal(err)
				return
			}
			t.Logf("[TEST] if exist: %+v,%+v", res, ext)
			channels, err := dm.GetSubscribedChannelsByGroupID(ctx, 2)
			if err != nil {
				t.Fatal(err)
				return
			}
			t.Logf("[TEST] 2 channels: %+v", channels)
			// del
			//err = dm.Unsubscribe(ctx, tc.gid, tc.feedLink)
			//if err != nil {
			//	t.Fatal(err)
			//	return
			//}
			//res, ext, err = dm.storage.GetIfExistedSubscribe(ctx, tc.gid, tc.feedLink)
			//if err != nil {
			//	t.Fatal(err)
			//	return
			//}
			//t.Logf("[TEST] after del: %+v,%+v", res, ext)
			//if res != nil || ext {
			//	t.Fatal("delete failed")
			//}

		})
	}
}

func Test_SyncFeed(t *testing.T) {
	feed, err := dm.Sync(context.Background())
	if err != nil {
		t.Fatal(err)
		return
	}
	rs, _ := json.Marshal(feed)
	t.Logf("[Test] feed: %+v", string(rs))
}
