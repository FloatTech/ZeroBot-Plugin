package plugin_gif

import (
	"strconv"
	"strings"

	"github.com/FloatTech/zbputils/file"
)

func (c *context) prepareLogos(s ...string) {
	for i, v := range s {
		_, err := strconv.Atoi(v)
		if err != nil {
			file.DownloadTo("https://gchat.qpic.cn/gchatpic_new//--"+strings.ToUpper(v)+"/0", c.usrdir+strconv.Itoa(i)+".gif", true)
		} else {
			file.DownloadTo("http://q4.qlogo.cn/g?b=qq&nk="+v+"&s=640", c.usrdir+strconv.Itoa(i)+".gif", true)
		}
	}
}
