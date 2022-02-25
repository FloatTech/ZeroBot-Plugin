package gif

import (
	"os"
	"strconv"
	"sync"

	"github.com/FloatTech/zbputils/file"
	"github.com/FloatTech/zbputils/img"
	"github.com/sirupsen/logrus"
)

type context struct {
	usrdir      string
	headimgsdir []string
}

func dlchan(name string, s *string, wg *sync.WaitGroup, exit func(error)) {
	defer wg.Done()
	target := datapath + `materials/` + name
	var err error
	if file.IsNotExist(target) {
		err = file.DownloadTo(`https://codechina.csdn.net/u011570312/imagematerials/-/raw/main/`+name, target, true)
		if err != nil {
			exit(err)
			return
		}
		logrus.Debugln("[gif] dl", name, "to", target, "succeeded")
	} else {
		logrus.Debugln("[gif] dl", name, "exists at", target)
	}
	*s = target
}

func dlblock(name string) (string, error) {
	target := datapath + `materials/` + name
	if file.IsNotExist(target) {
		err := file.DownloadTo(`https://codechina.csdn.net/u011570312/imagematerials/-/raw/main/`+name, target, true)
		if err != nil {
			return "", err
		}
		logrus.Debugln("[gif] dl", name, "to", target, "succeeded")
	} else {
		logrus.Debugln("[gif] dl", name, "exists at", target)
	}
	return target, nil
}

func dlrange(prefix string, end int, wg *sync.WaitGroup, exit func(error)) []string {
	if file.IsNotExist(datapath + `materials/` + prefix) {
		err := os.MkdirAll(datapath+`materials/`+prefix, 0755)
		if err != nil {
			exit(err)
			return nil
		}
	}
	c := make([]string, end)
	for i := range c {
		wg.Add(1)
		go dlchan(prefix+"/"+strconv.Itoa(i)+".png", &c[i], wg, exit)
	}
	return c
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

func loadFirstFrames(paths []string, size int) (imgs []*img.ImgFactory, err error) {
	imgs = make([]*img.ImgFactory, size)
	for i := range imgs {
		imgs[i], err = img.LoadFirstFrame(paths[i], 0, 0)
		if err != nil {
			return nil, err
		}
	}
	return imgs, nil
}
