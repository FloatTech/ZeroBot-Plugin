package wangyiyun

import
(
    
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
    control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/web"
	"github.com/FloatTech/ZeroBot-Plugin/order"
)

const(
    wangyiyunURL = "http://ovooa.com/API/wyrp/api.php?type=text"
    wangyiyunReferer = "http://ovooa.com/"
    	ua           = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36"
    )
    
func init() { 
    engine := control.Register("wangyiyun",order.PrioWangYiYun, &control.Options{
        
		DisableOnDefault: false,
		Help:           "wangyiyun \n- 来份网易云热评",
        
    })

    
	engine.OnFullMatch("来份网易云热评").SetBlock(true).
	Handle(func(ctx *zero.Ctx) {
		data, err := web.ReqWith(wangyiyunURL, "GET", wangyiyunReferer, ua)
		
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		
	})
}
