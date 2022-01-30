package chat

import (
	"encoding/json"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/FloatTech/zbputils/file"
	"github.com/FloatTech/zbputils/process"

	"github.com/FloatTech/ZeroBot-Plugin/order"
)

type kimo = map[string]*[]string

func initChatList(postinit func()) {
	go func() {
		defer order.DoneOnExit()()
		process.SleepAbout1sTo2s()
		_ = os.MkdirAll(dbpath, 0755)
		data, err := file.GetLazyData(dbfile, true, true)
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal(data, &kimomap)
		if err != nil {
			panic(err)
		}
		for k := range kimomap {
			chatList = append(chatList, k)
		}
		logrus.Infoln("[chat]加载", len(chatList), "条kimoi")
		postinit()
	}()
}
