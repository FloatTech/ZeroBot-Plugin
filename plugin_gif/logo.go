package gif

import (
	"strconv"
	"strings"

	"github.com/FloatTech/zbputils/file"
)

func (cc *context) prepareLogos(s ...string) {
	for i, v := range s {
		_, err := strconv.Atoi(v)
		if err != nil {
			_ = file.DownloadTo("https://gchat.qpic.cn/gchatpic_new//--"+strings.ToUpper(v)+"/0", cc.usrdir+strconv.Itoa(i)+".gif", true)
		} else {
			_ = file.DownloadTo("http://q4.qlogo.cn/g?b=qq&nk="+v+"&s=640", cc.usrdir+strconv.Itoa(i)+".gif", true)
		}
	}
}
