package shadiao

/*
func init() {
		engine.OnFullMatch("骂我").SetBlock(true).SetPriority(prio).Handle(func(ctx *zero.Ctx) {
			if !limit.Load(ctx.Event.GroupID).Acquire() {
				return
			}
			data, err := web.ReqWith(zuanURL, "GET", zuanReferer, ua)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(helper.BytesToString(data)))
		})
		engine.OnRegex(`^骂他.*?(\d+)`, zero.OnlyGroup).SetBlock(true).SetPriority(40).
			Handle(func(ctx *zero.Ctx) {
				if !limit.Load(ctx.Event.GroupID).Acquire() {
					return
				}
				data, err := web.ReqWith(zuanURL, "GET", "", ua)
				if err != nil {
					ctx.SendChain(message.Text("ERROR:", err))
					return
				}
				ctx.SendChain(message.At(math.Str2Int64(ctx.State["regex_matched"].([]string)[1])), message.Text(helper.BytesToString(data)))
			})
}
*/
