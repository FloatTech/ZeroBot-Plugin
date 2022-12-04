package dress

import (
	"fmt"

	"github.com/FloatTech/floatbox/web"
	"github.com/tidwall/gjson"
)

const (
	dressURL       = "http://www.yoooooooooo.com/gitdress"
	male           = "dress"
	female         = "girldress"
	dressListURL   = dressURL + "/%v/album/list.json"
	dressDetailURL = dressURL + "/%v/album/%v/info.json"
	dressImageURL  = dressURL + "/%v/album/%v/%v-m.webp"
)

func dressList(sex string) (dressList []string, err error) {
	data, err := web.GetData(fmt.Sprintf(dressListURL, sex))
	if err != nil {
		return
	}
	gj := gjson.ParseBytes(data)
	dressList = make([]string, 0, int(gj.Get("@this.#").Int()))
	gj.Get("@this").ForEach(func(_, v gjson.Result) bool {
		dressList = append(dressList, v.String())
		return true
	})
	return
}

func detail(sex, name string) (count int, err error) {
	data, err := web.GetData(fmt.Sprintf(dressDetailURL, sex, name))
	if err != nil {
		return
	}
	count = int(gjson.ParseBytes(data).Get("@this.#").Int())
	return
}
