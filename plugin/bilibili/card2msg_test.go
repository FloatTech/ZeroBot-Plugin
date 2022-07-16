package bilibili

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/FloatTech/zbputils/binary"
	"github.com/FloatTech/zbputils/web"
	"github.com/tidwall/gjson"
)

func TestArticleInfo(t *testing.T) {
	card, err := getArticleInfo("17279244")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(articleCard2msg(card, "17279244"))

}

func TestSpaceHistory(t *testing.T) {
	data, err := web.GetData(fmt.Sprintf(SpaceHistoryURL, "667526012", "642279068898689029"))
	if err != nil {
		t.Fatal(err)
	}
	var desc Desc
	_ = json.Unmarshal([]byte(gjson.ParseBytes(data).Get("data.cards.0.desc").Raw), &desc)
	t.Logf("desc:%+v\n", desc)
	var card Card
	_ = json.Unmarshal([]byte(gjson.ParseBytes(data).Get("data.cards.0.card").Str), &card)
	t.Logf("card:%+v\n", card)
}

func TestCard2msg(t *testing.T) {
	data, err := web.GetData(fmt.Sprintf(SpaceHistoryURL, "667526012", "642279068898689029"))
	if err != nil {
		t.Fatal(err)
	}
	var dynamicCard DynamicCard
	_ = json.Unmarshal([]byte(gjson.ParseBytes(data).Get("data.cards.0").Raw), &dynamicCard)
	t.Logf("dynCard:%+v\n", dynamicCard)
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
	var card MemberCard
	data, err := web.GetData(fmt.Sprintf(MemberCardURL, 2))
	if err != nil {
		return
	}
	str := gjson.ParseBytes(data).Get("card").String()
	err = json.Unmarshal(binary.StringToBytes(str), &card)
	if err != nil {
		return
	}
	t.Logf("%+v\n", card)
}

func TestVideoInfo(t *testing.T) {
	card, err := getVideoInfo("10007")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(videoCard2msg(card))
	card, err = getVideoInfo("BV1xx411c7mD")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(videoCard2msg(card))
	card, err = getVideoInfo("bv1xx411c7mD")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(videoCard2msg(card))
	card, err = getVideoInfo("BV1mF411j7iU")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(videoCard2msg(card))
}

func TestLiveRoomInfo(t *testing.T) {
	card, err := getLiveRoomInfo("83171")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(liveCard2msg(card))
}
