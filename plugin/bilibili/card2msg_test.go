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

func TestDynamicDetail(t *testing.T) {
	t.Log("cType = 1")
	t.Log(dynamicDetail("642279068898689029"))

	t.Log("cType = 2")
	t.Log(dynamicDetail("642470680290394121"))

	t.Log("cType = 2048")
	t.Log(dynamicDetail("642277677329285174"))

	t.Log("cType = 4")
	t.Log(dynamicDetail("642154347357011968"))

	t.Log("cType = 8")
	t.Log(dynamicDetail("675892999274627104"))

	t.Log("cType = 4308")
	t.Log(dynamicDetail("668598718656675844"))

	t.Log("cType = 64")
	t.Log(dynamicDetail("675966082178088963"))

	t.Log("cType = 256")
	t.Log(dynamicDetail("599253048535707632"))

	t.Log("cType = 4,投票类型")
	t.Log(dynamicDetail("677231070435868704"))
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

func TestLiveRoomInfo(t *testing.T) {
	card, err := bz.GetLiveRoomInfo("83171")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(liveCard2msg(card))
}
