// Package shadiao 来源于 https://shadiao.app/# 的接口
package shadiao

import (
	control "github.com/FloatTech/zbputils/control"

	"github.com/FloatTech/zbputils/control/order"
)

const (
	chpURL          = "https://chp.shadiao.app/api.php"
	duURL           = "https://du.shadiao.app/api.php"
	pyqURL          = "https://pyq.shadiao.app/api.php"
	yduanziURL      = "http://www.yduanzi.com/duanzi/getduanzi"
	chayiURL        = "https://api.lovelive.tools/api/SweetNothings/Web/0"
	ganhaiURL       = "https://api.lovelive.tools/api/SweetNothings/Web/1"
	ua              = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36"
	chpReferer      = "https://chp.shadiao.app/"
	duReferer       = "https://du.shadiao.app/"
	pyqReferer      = "https://pyq.shadiao.app/"
	yduanziReferer  = "http://www.yduanzi.com/?utm_source=shadiao.app"
	loveliveReferer = "https://lovelive.tools/"
)

var (
	engine = control.Register("shadiao", order.AcquirePrio(), &control.Options{
		DisableOnDefault: false,
		Help: "沙雕app\n" +
			"- 哄我\n- 渣我\n- 来碗绿茶\n- 发个朋友圈\n- 来碗毒鸡汤\n- 讲个段子",
	})
)
