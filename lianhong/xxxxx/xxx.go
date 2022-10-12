/* Package novelai NovelAI作画
package novelai

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	rei "github.com/fumiama/ReiBot"
	tgba "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	nvai "github.com/FloatTech/AnimeAPI/novelai"
	"github.com/FloatTech/floatbox/binary"
	"github.com/FloatTech/floatbox/file"
	ctrl "github.com/FloatTech/zbpctrl"

	"github.com/FloatTech/ReiBot-Plugin/utils/ctxext"
)

var nv = map[int64]*nvai.NovalAI{}

func init() {
	en := rei.Register("novelai", &ctrl.Options[*rei.Ctx]{
		DisableOnDefault: false,
		Help: "novelai\n" +
			"- novelai作图 (seed=123) tag1 tag2...\n" +
			"- novelai查tag 文件哈希\n" +
			"- 设置(仅供我使用的|仅供群-1234使用的) novelai key [key]",
		PrivateDataFolder: "novelai",
	}).ApplySingle(ctxext.DefaultSingle)
	ims = newims(en.DataFolder() + "images.db")
	imgdir := en.DataFolder() + "imgs"
	err := os.MkdirAll(imgdir, 0755)
	if err != nil {
		panic(err)
	}
	keyfile := en.DataFolder() + "key.txt"
	if file.IsExist(keyfile) {
		key, err := os.ReadFile(keyfile)
		if err != nil {
			panic(err)
		}
		n := nvai.NewNovalAI(binary.BytesToString(key), nvai.NewDefaultPayload())
		err = n.Login()
		if err != nil {
			panic(err)
		}
		err = ims.Insert("k", &keystorage{Key: binary.BytesToString(key)})
		if err != nil {
			panic(err)
		}
		nv[0] = n
		_ = os.Remove(keyfile)
	} else {
		k := &keystorage{}
		p := nvai.NewDefaultPayload()
		_ = ims.FindFor("k", k, "WHERE onlyme=false", func() error {
			n := nvai.NewNovalAI(k.Key, p)
			err = n.Login()
			if err == nil && n.Tok != "" {
				nv[0] = n
				return io.EOF
			}
			return nil
		})
	}
	en.OnMessagePrefix("novelai作图").Limit(ctxext.LimitByGroup).SetBlock(true).
		Handle(func(ctx *rei.Ctx) {
			k := &keystorage{}
			mu.RLock()
			_ = ims.Find("k", k, "WHERE (sender="+strconv.FormatInt(ctx.Message.From.ID, 10)+" or sender="+strconv.FormatInt(ctx.Message.Chat.ID, 10)+") and onlyme=true")
			n, ok := nv[k.Sender]
			mu.RUnlock()
			if !ok {
				n = nvai.NewNovalAI(k.Key, nvai.NewDefaultPayload())
				err = n.Login()
				if err != nil {
					_, _ = ctx.SendPlainMessage(false, "ERROR: ", err)
					return
				}
				mu.Lock()
				nv[k.Sender] = n
				mu.Unlock()
			}
			if n.Tok == "" {
				_, _ = ctx.SendPlainMessage(false, "请私聊发送 设置(仅供我使用的|仅供群-1234使用的) novelai key [key] 以启用 novelai 作图 (方括号不需要输入)")
				return
			}
			t := strings.TrimSpace(ctx.State["args"].(string))
			if strings.HasPrefix(t, "seed=") {
				i := 5
				for ; i < len(t); i++ {
					if t[i] < '0' || t[i] > '9' {
						break
					}
				}
				s := t[5:i]
				t = t[i:]
				if s != "" {
					p := nvai.NewDefaultPayload()
					p.Parameters.Seed, err = strconv.Atoi(s)
					if err != nil {
						_, _ = ctx.SendPlainMessage(false, "ERROR: ", err)
						return
					}
					nn := nvai.NewNovalAI("", p)
					nn.Tok = n.Tok
					n = nn
				}
			}
			seed, tags, img, err := n.Draw(strings.TrimSpace(t))
			if err != nil {
				_, _ = ctx.SendPlainMessage(false, "ERROR: ", err)
				return
			}
			seedtext := strconv.Itoa(seed)
			fn := tags + " " + seedtext
			id := idof(fn)
			err = os.WriteFile(fmt.Sprintf("%s/%016x.png", imgdir, uint64(id)), img, 0755)
			if err != nil {
				_, _ = ctx.SendPlainMessage(false, "ERROR: ", err)
				return
			}
			mu.Lock()
			err = ims.Insert("s", &imgstorage{
				ID:   id,
				Seed: int32(seed),
				Tags: tags,
			})
			mu.Unlock()
			pho := &tgba.PhotoConfig{
				BaseFile: tgba.BaseFile{
					BaseChat: tgba.BaseChat{
						ChatID:           ctx.Message.Chat.ID,
						ReplyToMessageID: ctx.Message.MessageID,
					},
					File: tgba.FileBytes{Bytes: img},
				},
				Caption: "seed: " + seedtext + "\ntags: " + tags,
				CaptionEntities: []tgba.MessageEntity{
					{Type: "bold", Offset: 0, Length: 5},
					{Type: "bold", Offset: 5 + 1 + len(seedtext) + 1, Length: 5},
				},
			}
			if err == nil {
				pho.ReplyMarkup = tgba.NewInlineKeyboardMarkup(
					tgba.NewInlineKeyboardRow(
						tgba.NewInlineKeyboardButtonData(
							"发送原图",
							"nvaiorg"+fmt.Sprintf("%016x", uint64(id)),
						),
						tgba.NewInlineKeyboardButtonData(
							"移除该图",
							"nvaidel"+fmt.Sprintf("%016x", uint64(id)),
						),
					),
				)
			}
			_, _ = ctx.Caller.Send(pho)
		})
	en.OnMessageRegex(`^novelai查tag([0-9a-f]{16})$`).SetBlock(true).
		Handle(func(ctx *rei.Ctx) {
			fn := ctx.State["regex_matched"].([]string)[1]
			id, err := strconv.ParseUint(fn, 16, 64)
			if err != nil {
				_, _ = ctx.SendPlainMessage(false, "ERROR: ", err)
				return
			}
			ids := strconv.FormatInt(int64(id), 10)
			s := &imgstorage{}
			mu.RLock()
			err = ims.Find("s", s, "WHERE id="+ids)
			mu.RUnlock()
			if err != nil {
				_, _ = ctx.SendPlainMessage(false, "ERROR: ", err)
				return
			}
			seedtext := strconv.Itoa(int(s.Seed))
			_, _ = ctx.SendMessage(false, "seed: "+seedtext+"\ntags: "+s.Tags,
				tgba.MessageEntity{Type: "bold", Offset: 0, Length: 5},
				tgba.MessageEntity{Type: "bold", Offset: 5 + 1 + len(seedtext) + 1, Length: 5},
			)
		})
	en.OnCallbackQueryRegex(`^nvaiorg([0-9a-f]{16})$`).SetBlock(true).
		Handle(func(ctx *rei.Ctx) {
			fn := ctx.State["regex_matched"].([]string)[1]
			imgp := fmt.Sprintf("%s/%s.png", imgdir, fn)
			_, err = ctx.Caller.Send(&tgba.DocumentConfig{
				BaseFile: tgba.BaseFile{
					BaseChat: tgba.BaseChat{
						ChatID:           ctx.Message.Chat.ID,
						ReplyToMessageID: ctx.Message.MessageID,
					},
				File: tgba.FilePath(imgp),
				},
			})
			if err != nil {
				_, _ = ctx.Caller.Send(tgba.NewCallbackWithAlert(ctx.Value.(*tgba.CallbackQuery).ID, "ERROR: "+err.Error()))
				return
			}
			if len(ctx.Message.ReplyMarkup.InlineKeyboard) > 0 && len(ctx.Message.ReplyMarkup.InlineKeyboard[0]) > 1 {
				ctx.Message.ReplyMarkup.InlineKeyboard[0] = ctx.Message.ReplyMarkup.InlineKeyboard[0][1:]
				_, _ = ctx.Caller.Send(tgba.NewEditMessageReplyMarkup(ctx.Message.Chat.ID, ctx.Message.MessageID, *ctx.Message.ReplyMarkup))
			}
			_, _ = ctx.Caller.Send(tgba.NewCallbackWithAlert(ctx.Value.(*tgba.CallbackQuery).ID, "已发送"))
		})
	en.OnCallbackQueryRegex(`^nvaidel([0-9a-f]{16})$`).SetBlock(true).
		Handle(func(ctx *rei.Ctx) {
			if !rei.AdminPermission(ctx) && ctx.Message.ReplyToMessage.From.ID != ctx.Value.(*tgba.CallbackQuery).From.ID {
				_, _ = ctx.Caller.Send(tgba.NewCallbackWithAlert(ctx.Value.(*tgba.CallbackQuery).ID, "ERROR: 只有管理员或作图发起者才可移除图片"))
				return
			}
			fn := ctx.State["regex_matched"].([]string)[1]
			id, err := strconv.ParseUint(fn, 16, 64)
			if err != nil {
				_, _ = ctx.Caller.Send(tgba.NewCallbackWithAlert(ctx.Value.(*tgba.CallbackQuery).ID, "ERROR: "+err.Error()))
				return
			}
			imgp := fmt.Sprintf("%s/%s.png", imgdir, fn)
			ids := strconv.FormatInt(int64(id), 10)
			err = os.Remove(imgp)
			if err != nil {
				_, _ = ctx.Caller.Send(tgba.NewCallbackWithAlert(ctx.Value.(*tgba.CallbackQuery).ID, "ERROR: "+err.Error()))
				return
			}
			mu.Lock()
			_ = ims.Del("s", "WHERE id="+ids)
			mu.Unlock()
			_, _ = ctx.Caller.Send(tgba.NewDeleteMessage(ctx.Message.Chat.ID, ctx.Message.MessageID))
			_, _ = ctx.Caller.Send(tgba.NewCallbackWithAlert(ctx.Value.(*tgba.CallbackQuery).ID, "成功"))
		})
	en.OnMessageRegex(`^设置(仅供我使用的|仅供群-?\d+使用的)?\s?novelai\s?key\s?([0-9A-Za-z_\-]{64})$`, rei.OnlyPrivate).SetBlock(true).
		Handle(func(ctx *rei.Ctx) {
			opt := ctx.State["regex_matched"].([]string)[1]
			onlyme := opt == "仅供我使用的"
			onlychat := opt != "仅供我使用的" && opt != ""
			if !onlyme && !onlychat && !rei.SuperUserPermission(ctx) {
				_, _ = ctx.SendPlainMessage(false, "ERROR: 只有主人可以设置全局key")
				return
			}
			id := ctx.Message.From.ID
			if onlychat {
				id, err = strconv.ParseInt(opt[3*3:len(opt)-3*3], 10, 64)
				if err != nil {
					_, _ = ctx.SendPlainMessage(false, "ERROR: ", err)
					return
				}
				if !rei.SuperUserPermission(ctx) {
					memb, err := ctx.Caller.GetChatAdministrators(tgba.ChatAdministratorsConfig{
						ChatConfig: tgba.ChatConfig{
							ChatID: id,
						},
					})
					if err != nil {
						_, _ = ctx.SendPlainMessage(false, "ERROR: ", err)
						return
					}
					found := false
					for _, mb := range memb {
						if mb.User.ID == ctx.Message.From.ID {
							found = true
							break
						}
					}
					if !found {
						_, _ = ctx.SendPlainMessage(false, "ERROR: 只有群管理员可以设置本群key")
						return
					}
				}
			}
			onlyme = onlychat
			key := ctx.State["regex_matched"].([]string)[2]
			err := os.WriteFile(keyfile, binary.StringToBytes(key), 0644)
			if err != nil {
				_, _ = ctx.SendPlainMessage(false, "ERROR: ", err)
				return
			}
			nnv := nvai.NewNovalAI(key, nvai.NewDefaultPayload())
			err = nnv.Login()
			if err != nil {
				_, _ = ctx.SendPlainMessage(false, "ERROR: ", err)
				return
			}
			mu.Lock()
			err = ims.Insert("k", &keystorage{
				Sender: id,
				OnlyMe: onlyme,
				Key:    key,
			})
			if err == nil {
				nv[id] = nnv
			}
			mu.Unlock()
			if err != nil {
				_, _ = ctx.SendPlainMessage(false, "ERROR: ", err)
				return
			}
			_, _ = ctx.SendPlainMessage(false, "成功!")
		})
	en.OnMessageRegex(`^移除(仅供我使用的|仅供群-?\d+使用的)?\s?novelai\s?key$`, rei.OnlyPrivate).SetBlock(true).
		Handle(func(ctx *rei.Ctx) {
			opt := ctx.State["regex_matched"].([]string)[1]
			onlyme := opt == "仅供我使用的"
			onlychat := opt != "仅供我使用的" && opt != ""
			if !onlyme && !onlychat && !rei.SuperUserPermission(ctx) {
				_, _ = ctx.SendPlainMessage(false, "ERROR: 只有主人可以设置全局key")
				return
			}
			id := ctx.Message.From.ID
			if onlychat {
				id, err = strconv.ParseInt(opt[3*3:len(opt)-3*3], 10, 64)
				if err != nil {
					_, _ = ctx.SendPlainMessage(false, "ERROR: ", err)
					return
				}
				if !rei.SuperUserPermission(ctx) {
					memb, err := ctx.Caller.GetChatAdministrators(tgba.ChatAdministratorsConfig{
						ChatConfig: tgba.ChatConfig{
							ChatID: id,
						},
					})
					if err != nil {
						_, _ = ctx.SendPlainMessage(false, "ERROR: ", err)
						return
					}
					found := false
					for _, mb := range memb {
						if mb.User.ID == ctx.Message.From.ID {
							found = true
							break
						}
					}
					if !found {
						_, _ = ctx.SendPlainMessage(false, "ERROR: 只有群管理员可以设置本群key")
						return
					}
				}
			}
			onlyme = onlychat
			mu.Lock()
			err = ims.Del("k", "WHERE sender="+strconv.FormatInt(id, 10)+" and onlyme="+fmt.Sprint(onlyme))
			mu.Unlock()
			if err != nil {
				_, _ = ctx.SendPlainMessage(false, "ERROR: ", err)
				return
			}
			_, _ = ctx.SendPlainMessage(false, "成功!")
		})
}
*/