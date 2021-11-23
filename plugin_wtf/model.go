package wtf

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
)

/* JS path getter for https://wtf.hiigara.net/ranking
a = document.getElementById("testList").getElementsByTagName("a")
s = ""
for(i=0; i<a.length; i++) {
    s += "\"" + a[i].innerText + "\":\"" + a[i].href + "\",\n";
}
*/

const apiprefix = "https://wtf.hiigara.net/api/run/"

type Wtf struct {
	name string
	path string
}

var table = [...]*Wtf{
	&Wtf{"你的意义是什么?", "mRIFuS"},
	&Wtf{"【ABO】性別和信息素", "KXyy9"},
	&Wtf{"测测cp", "ZoGXQd"},
	&Wtf{"xxx和xxx的關係是？", "L4HfA"},
	&Wtf{"在JOJO世界，你的替身会是什么？", "lj0a8o"},
	&Wtf{"稱號產生器", "titlegen"},
	&Wtf{"成分报告", "2PCeo1"},
	&Wtf{"測驗你跟你的朋友是攻/受", "LkQXO3"},
	&Wtf{"测试两人的关系？", "uwjQQt"},
	&Wtf{"【Fate系列】當你成為了從者 2.0", "LHStH2"},
	&Wtf{"想不到自己未來要做什麼工作嗎?", "D1agGa"},
	&Wtf{"(σﾟ∀ﾟ)σ名字產生器", "LNxXq7"},
	&Wtf{"人設生產器", "LBtPu5"},
	&Wtf{"測驗你在ABO世界的訊息素", "SwmdU"},
	&Wtf{"爱是什么", "llpBEY"},
	&Wtf{"測測你和哪位名人相似？", "RHQeXu"},
	&Wtf{"S/M测试", "Ga47oZ"},
	&Wtf{"测测你是谁", "aV1AEi"},
	&Wtf{"取個綽號吧", "LTkyUy"},
	&Wtf{"什麼都不是", "vyrSCb"},
	&Wtf{"今天中午吃什麼", "LdS4K6"},
	&Wtf{"測試你的中二稱號", "LwUmQ6"},
	&Wtf{"神奇海螺", "Lon1h7"},
	&Wtf{"ABO測試", "H1Tgd"},
	&Wtf{"女主角姓名產生器", "MsQBTd"},
	&Wtf{"您是什么人", "49PwSd"},
	&Wtf{"如果你成为了干员", "ok5e7n"},
	&Wtf{"abo人设生成~", "Di8enA"},
	&Wtf{"✡你的命運✡塔羅占卜🔮", "ohCzID"},
	&Wtf{"小說大綱生產器", "Lnstjz"},
	&Wtf{"他会喜欢你吗？", "pezX3a"},
	&Wtf{"抽签！你明年的今天会干什么", "IF31kS"},
	&Wtf{"如果你是受，會是哪種受呢？", "Dr6zpF"},
	&Wtf{"cp文梗", "vEO2KD"},
	&Wtf{"您是什么人？", "TQ5qyl"},
	&Wtf{"你成為......的機率", "g0uoBL"},
	&Wtf{"ABO性別與信息素", "KFPju"},
	&Wtf{"異國名稱產生器(國家、人名、星球...)", "OBpu4"},
	&Wtf{"對方到底喜不喜歡你", "JSLoZC"},
	&Wtf{"【脑叶公司】测一测你在脑叶公司的经历", "uPBhjC"},
	&Wtf{"当你成为魔法少女", "7ZiGcJ"},
	&Wtf{"你是yyds吗?", "SpBnCa"},
	&Wtf{"○○喜歡你嗎？", "S6Uceo"},
	&Wtf{"测测你的sm属性", "dOtcO5"},
	&Wtf{"你/妳究竟是攻還是受呢?", "RXALH"},
	&Wtf{"神秘藏书阁", "tDRyET"},
	&Wtf{"中午吃什么？", "L0Wsis"},
	&Wtf{"十年后，你cp的结局是", "VUwnXQ"},
	&Wtf{"高维宇宙与常数的你", "6Zql97"},
	&Wtf{"色色的東東", "o2eg74"},
	&Wtf{"文章標題產生器", "Ky25WO"},
	&Wtf{"你的成績怎麼樣", "6kZv69"},
	&Wtf{"智能SM偵測器ヾ(*ΦωΦ)ツ", "9pY6HQ"},
	&Wtf{"你的使用注意事項", "La4Gir"},
	&Wtf{"戀愛指數", "Jsgz0"},
	&Wtf{"测试你今晚拉的屎", "N8dbcL"},
	&Wtf{"成為情侶的機率ᶫᵒᵛᵉᵧₒᵤ♥", "eDURch"},
	&Wtf{"他對你...", "CJxHMf"},
	&Wtf{"你的明日方舟人际关系", "u5z4Mw"},
	&Wtf{"日本姓氏產生器", "JJ5Ctb"},
	&Wtf{"當你轉生到了異世界，你將成為...", "FTpwK"},
	&Wtf{"魔幻世界大穿越2.0", "wUATOq"},
	&Wtf{"未來男朋友", "F3dSV"},
	&Wtf{"ABO與信息素", "KFOGA"},
	&Wtf{"你必將就這樣一事無成啊アホ", "RWw9oX"},
	&Wtf{"用習慣舉手的方式測試你的戀愛運!<3", "wv5bzA"},
	&Wtf{"攻受", "RaKmY"},
	&Wtf{"你和你喜歡的人的微h寵溺段子XD", "LdQqGz"},
	&Wtf{"我的藝名", "LBaTx"},
	&Wtf{"你是什麼神？", "LqZORE"},
	&Wtf{"你的起源是什麼？", "HXWwC"},
	&Wtf{"測你喜歡什麼", "Sue5g2"},
	&Wtf{"看看朋友的秘密", "PgKb8r"},
	&Wtf{"你在動漫裡的名字", "Lz82V7"},
	&Wtf{"小說男角名字產生器", "LyGDRr"},
	&Wtf{"測試短文", "S48yA"},
	&Wtf{"我們兩人在一起的機率......", "LBZbgE"},
	&Wtf{"創造小故事", "Kjy3AS"},
	&Wtf{"你的另外一個名字", "LuyYQA"},
	&Wtf{"與你最匹配的攻君屬性 ！？", "I7pxy"},
	&Wtf{"英文全名生產器(女)", "HcYbq"},
	&Wtf{"BL文章生產器", "LBZMO"},
	&Wtf{"輕小說書名產生器", "NFucA"},
	&Wtf{"長相評分", "2cQSDP"},
	&Wtf{"日本名字產生器（女孩子）", "JRiKv"},
	&Wtf{"中二技能名產生器", "Ky1BA"},
	&Wtf{"抽籤", "XqxfuH"},
	&Wtf{"你的蘿莉控程度全國排名", "IIWh9k"},
}

func NewWtf(index int) *Wtf {
	if index >= 0 && index < len(table) {
		return table[index]
	}
	return nil
}

type result struct {
	Text string `json:"text"`
	// Path string `json:"path"`
	Ok  bool   `json:"ok"`
	Msg string `json:"msg"`
}

func (w *Wtf) Predict(name string) (string, error) {
	u := apiprefix + w.path + "/" + url.PathEscape(name)
	resp, err := http.Get(u)
	if err != nil {
		return "", err
	}
	r, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return "", err
	}
	re := new(result)
	err = json.Unmarshal(r, re)
	if err != nil {
		return "", err
	}
	if re.Ok {
		return "> " + w.name + "\n" + re.Text, nil
	}
	return "", errors.New(re.Msg)
}
