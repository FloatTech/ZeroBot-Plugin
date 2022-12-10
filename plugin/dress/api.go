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
	arr := gjson.ParseBytes(data).Get("@this").Array()
	dressList = make([]string, len(arr))
	for i, v := range arr {
		dressList[i] = v.String()
	}
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
