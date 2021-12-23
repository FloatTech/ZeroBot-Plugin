// Package shadiao 基于 https://shadiao.app/# 的j接口
package shadiao

import (
	"github.com/FloatTech/ZeroBot-Plugin/control"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
	"time"
)

const (
	chpURL          = "https://chp.shadiao.app/api.php"
	duURL           = "https://du.shadiao.app/api.php"
	pyqURL          = "https://pyq.shadiao.app/api.php"
	yduanziURL      = "http://www.yduanzi.com/duanzi/getduanzi"
	chayiURL        = "https://api.lovelive.tools/api/SweetNothings/Web/0"
	ganhaiURL       = "https://api.lovelive.tools/api/SweetNothings/Web/1"
	zuanURL         = "https://zuanbot.com/api.php?level=min&lang=zh_cn"
	ua              = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36"
	chpReferer      = "https://chp.shadiao.app/"
	duReferer       = "https://du.shadiao.app/"
	pyqReferer      = "https://pyq.shadiao.app/"
	yduanziReferer  = "http://www.yduanzi.com/?utm_source=shadiao.app"
	loveliveReferer = "https://lovelive.tools/"
	zuanReferer     = "https://zuanbot.com/"
)

var (
	engine = control.Register("curse", &control.Options{
		DisableOnDefault: false,
		Help: "沙雕app\n" +
			"- 骂他[@xxx]|骂他[qq号]\n- 骂我\n- 哄我\n- 渣我\n- 来碗绿茶\n- 发个朋友圈\n- 来碗毒鸡汤\n- 讲个段子",
	})
	limit = rate.NewManager(time.Minute, 60)
)
