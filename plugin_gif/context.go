package gif

import (
	"os"
	"strconv"

	"github.com/FloatTech/zbputils/file"
	"github.com/sirupsen/logrus"
)

type context struct {
	usrdir      string
	headimgsdir []string
}

func dlchan(name string, c *chan *string) {
	target := datapath + `materials/` + name
	if file.IsNotExist(target) {
		_ = file.DownloadTo(`https://codechina.csdn.net/u011570312/imagematerials/-/raw/main/`+name, target, true)
	} else {
		logrus.Debugln("[gif] dl", name, "exists")
	}
	*c <- &target
}

func dlblock(name string) string {
	target := datapath + `materials/` + name
	if file.IsNotExist(target) {
		_ = file.DownloadTo(`https://codechina.csdn.net/u011570312/imagematerials/-/raw/main/`+name, target, true)
	} else {
		logrus.Debugln("[gif] dl", name, "exists")
	}
	return target
}

func dlrange(prefix string, end int) *[]chan *string {
	if file.IsNotExist(datapath + `materials/` + prefix) {
		_ = os.MkdirAll(datapath+`materials/`+prefix, 0755)
	}
	c := make([]chan *string, end)
	for i := range c {
		c[i] = make(chan *string)
		go dlchan(prefix+"/"+strconv.Itoa(i)+".png", &c[i])
	}
	return &c
}

// 新的上下文
func newContext(user int64) *context {
	c := new(context)
	c.usrdir = datapath + "users/" + strconv.FormatInt(user, 10) + `/`
	_ = os.MkdirAll(c.usrdir, 0755)
	c.headimgsdir = make([]string, 2)
	c.headimgsdir[0] = c.usrdir + "0.gif"
	c.headimgsdir[1] = c.usrdir + "1.gif"
	return c
}
