package aipaint

import (
	"os"
	"strconv"
	"strings"

	"github.com/FloatTech/floatbox/file"
)

type context struct {
	usrdir      string
	headimgsdir []string
}

func newContext(user int64) *context {
	c := new(context)
	c.usrdir = datapath + "users/" + strconv.FormatInt(user, 10) + `/`
	_ = os.MkdirAll(c.usrdir, 0755)
	c.headimgsdir = make([]string, 2)
	c.headimgsdir[0] = c.usrdir + "0.gif"
	c.headimgsdir[1] = c.usrdir + "1.gif"
	return c
}

func (cc *context) prepareLogos(s ...string) error {
	for i, v := range s {
		_, err := strconv.Atoi(v)
		if err != nil {
			err = file.DownloadTo("https://gchat.qpic.cn/gchatpic_new//--"+strings.ToUpper(v)+"/0", cc.usrdir+strconv.Itoa(i)+".gif", true)
		} else {
			err = file.DownloadTo("http://q4.qlogo.cn/g?b=qq&nk="+v+"&s=640", cc.usrdir+strconv.Itoa(i)+".gif", true)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
