package plugin_gif

import (
	"bufio"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

type context struct {
	usrdir      string
	headimgsdir []string
}

func dlchan(name string, c *chan *string) {
	target := datapath + `materials/` + name
	_, err := os.Stat(target)
	if err != nil {
		download(`https://codechina.csdn.net/u011570312/imagematerials/-/raw/main/`+name, target)
	} else {
		logrus.Debugln("[gif] dl", name, "exists")
	}
	*c <- &target
}

func dlblock(name string) string {
	target := datapath + `materials/` + name
	_, err := os.Stat(target)
	if err != nil {
		download(`https://codechina.csdn.net/u011570312/imagematerials/-/raw/main/`+name, target)
	}
	return target
}

func dlrange(prefix string, suffix string, end int) *[]chan *string {
	c := make([]chan *string, end)
	for i := range c {
		c[i] = make(chan *string)
		go dlchan(prefix+strconv.Itoa(i)+suffix, &c[i])
	}
	return &c
}

// 新的上下文
func newContext(user int64) *context {
	c := new(context)
	c.usrdir = datapath + "users/" + strconv.FormatInt(user, 10) + `/`
	os.MkdirAll(c.usrdir, 0755)
	c.headimgsdir = make([]string, 2)
	c.headimgsdir[0] = c.usrdir + "0.gif"
	c.headimgsdir[1] = c.usrdir + "1.gif"
	return c
}

// 下载图片
func download(url, dlpath string) error {
	// 创建目录
	var List = strings.Split(dlpath, `/`)
	err := os.MkdirAll(strings.TrimSuffix(dlpath, List[len(List)-1]), 0755)
	if err != nil {
		logrus.Errorln("[gif] mkdir err:", err)
		return err
	}
	res, err := http.Get(url)
	if err != nil {
		logrus.Errorln("[gif] http get err:", err)
		return err
	}
	// 获得get请求响应的reader对象
	reader := bufio.NewReaderSize(res.Body, 32*1024)
	// 创建文件
	file, err := os.Create(dlpath)
	if err != nil {
		logrus.Errorln("[gif] create file err:", err)
		return err
	}
	// 获得文件的writer对象
	writer := bufio.NewWriter(file)
	written, err := io.Copy(writer, reader)
	if err != nil {
		logrus.Errorln("[gif] copy err:", err)
		return err
	}
	res.Body.Close()
	logrus.Debugln("[gif] dl len:", written)
	return nil
}
