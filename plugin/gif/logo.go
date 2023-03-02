package gif

import (
	"image"
	"strconv"
	"strings"

	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/process"
	"github.com/FloatTech/imgfactory"
)

func (cc *context) prepareLogos(s ...string) error {
	for i, v := range s {
		_, err := strconv.Atoi(v)
		if err != nil {
			err = file.DownloadTo("https://gchat.qpic.cn/gchatpic_new//--"+strings.ToUpper(v)+"/0", cc.usrdir+strconv.Itoa(i)+".gif")
		} else {
			err = file.DownloadTo("http://q4.qlogo.cn/g?b=qq&nk="+v+"&s=640", cc.usrdir+strconv.Itoa(i)+".gif")
		}
		if err != nil {
			return err
		}
		process.SleepAbout1sTo2s()
	}
	return nil
}

func (cc *context) getLogo(w int, h int) (*image.NRGBA, error) {
	frame, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], w, h)
	if err != nil {
		return nil, err
	}
	return frame.Circle(0).Image(), nil
}

func (cc *context) getLogo2(w int, h int) (*image.NRGBA, error) {
	frame, err := imgfactory.LoadFirstFrame(cc.headimgsdir[1], w, h)
	if err != nil {
		return nil, err
	}
	return frame.Circle(0).Image(), nil
}
