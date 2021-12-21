package chat

import (
	"encoding/json"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/FloatTech/ZeroBot-Plugin/utils/file"
	"github.com/FloatTech/ZeroBot-Plugin/utils/process"
)

type kimo = map[string]*[]string

func initChatList(postinit func()) {
	go func() {
		process.SleepAbout1sTo2s()
		_ = os.MkdirAll(dbpath, 0755)
		data, err := file.GetLazyData(dbfile, true, true)
		if err != nil {
			panic(err)
		}
		json.Unmarshal(data, &kimomap)
		for k := range kimomap {
			chatList = append(chatList, k)
		}
		logrus.Infoln("[chat]加载", len(chatList), "条kimoi")
		postinit()
	}()
}
