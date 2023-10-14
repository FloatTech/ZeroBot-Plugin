package bilibili

import (
	"testing"

	bz "github.com/FloatTech/AnimeAPI/bilibili"
)

func TestArticleInfo(t *testing.T) {
	card, err := bz.GetArticleInfo("17279244")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(articleCard2msg(card, "17279244"))

}

func TestMemberCard(t *testing.T) {
	card, err := bz.GetMemberCard(2)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v\n", card)
}

func TestVideoInfo(t *testing.T) {
	card, err := bz.GetVideoInfo("10007")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(videoCard2msg(card))
	card, err = bz.GetVideoInfo("BV1xx411c7mD")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(videoCard2msg(card))
	card, err = bz.GetVideoInfo("bv1xx411c7mD")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(videoCard2msg(card))
	card, err = bz.GetVideoInfo("BV1mF411j7iU")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(videoCard2msg(card))
}

func TestGetVideoSummary(t *testing.T) {
	card, err := bz.GetVideoInfo("BV1mF411j7iU")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(getVideoSummary(card))
}

func TestLiveRoomInfo(t *testing.T) {
	card, err := bz.GetLiveRoomInfo("83171")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(liveCard2msg(card))
}
