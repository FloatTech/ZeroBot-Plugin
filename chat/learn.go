package chat

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type QA struct {
	Mutex sync.Mutex
	Data  map[int64]map[string]string
	Path  string
}

var (
	BotPath   = botPath()
	BotQQ     = "0"
	ImagePath = "data/chat/image/"

	Char      = map[int64]string{}                   // 群对应角色
	CharIndex = map[string]int64{"椛椛": 0, "ATRI": 1} // 角色对应编号

	// 角色问、有人问、我问 对象
	QACharPool  = &QA{Data: map[int64]map[string]string{}, Path: BotPath + "data/chat/char.json"}
	QAGroupPool = &QA{Data: map[int64]map[string]string{}, Path: BotPath + "data/chat/group.json"}
	QAUserPool  = &QA{Data: map[int64]map[string]string{}, Path: BotPath + "data/chat/user.json"}
)

func init() {
	QACharPool.load()
	QAGroupPool.load()
	QAUserPool.load()
	zero.OnRegex(`切换角色(.*)`, zero.AdminPermission).SetBlock(true).FirstPriority().
		Handle(func(ctx *zero.Ctx) {
			if _, ok := CharIndex[ctx.State["regex_matched"].([]string)[1]]; !ok {
				ctx.SendChain(message.Text("???"))
				return
			}
			Char[ctx.Event.GroupID] = ctx.State["regex_matched"].([]string)[1]
			ctx.SendChain(message.Text("已经切换了哦~"))
		})
	zero.OnRegex(`(.{1,2})问(.*)你答\s?\s?([\s\S]*)`, QAMatch(), QAPermission()).SetBlock(true).FirstPriority().
		Handle(func(ctx *zero.Ctx) {
			question := ctx.State["regex_matched"].([]string)[2]
			answer := ctx.State["regex_matched"].([]string)[3]
			// 根据匹配使用不同池子对象
			pool := ctx.State["qa_pool"].(*QA)
			user := ctx.State["qa_user"].(int64)
			// 重组艾特
			q := message.ParseMessageFromString(question)
			for i := range q {
				if q[i].Type == "at" && q[i].Data["qq"] == fmt.Sprint(ctx.Event.SelfID) {
					q[i].Data["qq"] = "BOT_QQ"
				}
			}
			question = q.CQString()

			// 对 CQ码 的 "[" 进行正则转义
			question = strings.ReplaceAll(question, "[", "\\[")
			question = strings.ReplaceAll(question, "]", "\\]")
			question = message.UnescapeCQCodeText(question)

			// 判断 question 是否符合正则表达式
			_, err := regexp.Compile("^" + question + "$")
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			// 保存图片，重组图片信息
			r := message.ParseMessageFromString(answer)
			for i := range r {
				if r[i].Type == "image" {
					if filename, err := down(r[i].Data["url"], BotPath+ImagePath); err == nil {
						r[i].Data["file"] = "file:///BOT_PATH/" + ImagePath + filename
						delete(r[i].Data, "url")
					} else { // 下载图片发生错误
						ctx.SendChain(message.Text("ERROR: ", err))
						return
					}
				}
			}
			answer = r.CQString()

			// 如果是BOT主人，则可以CQ码注入
			// 不存在CQ码也需要转义
			if !strings.Contains(answer, "CQ") || zero.SuperUserPermission(ctx) {
				answer = message.UnescapeCQCodeText(answer)
			}
			// 添加到池子
			pool.add(user, question, answer)
			ctx.SendChain(message.Text("好的我记住了~"))
		})
	zero.OnRegex(`删除(.{1,2})问(.*)`, QAMatch(), QAPermission()).SetBlock(true).FirstPriority().
		Handle(func(ctx *zero.Ctx) {
			question := ctx.State["regex_matched"].([]string)[2]
			// 根据匹配使用不同池子对象
			pool := ctx.State["qa_pool"].(*QA)
			user := ctx.State["qa_user"].(int64)
			if !strings.Contains(question, "CQ") || zero.AdminPermission(ctx) {
				question = message.UnescapeCQCodeText(question)
			}
			// 重组艾特
			q := message.ParseMessageFromString(question)
			for i := range q {
				if q[i].Type == "at" && q[i].Data["qq"] == fmt.Sprint(ctx.Event.SelfID) {
					q[i].Data["qq"] = "BOT_QQ"
				}
			}
			question = q.CQString()
			if answer := pool.del(user, question); answer != "" {
				ctx.SendChain(message.Text("我不会再回答[", answer, "]了"))
			} else {
				ctx.SendChain(message.Text("啊咧[", question, "]是什么？"))
			}
		})
	zero.OnRegex(`看看(.{1,2})问`, QAMatch()).SetBlock(true).FirstPriority().
		Handle(func(ctx *zero.Ctx) {
			title := ""
			r := []string{}
			switch ctx.State["regex_matched"].([]string)[1] {
			case "角色":
				char := "椛椛"
				if Char[ctx.Event.GroupID] != "" {
					char = Char[ctx.Event.GroupID]
				}
				title = "当前角色[" + char + "] 设置的关键词有：\n"
			case "有人":
				title = "本群设置的关键词有：\n"
			case "我":
				title = "你设置的关键词有：\n"
			}
			// 根据匹配使用不同池子对象
			pool := ctx.State["qa_pool"].(*QA)
			user := ctx.State["qa_user"].(int64)
			for question := range pool.Data[user] {
				r = append(r, question)
			}
			if len(r) == 0 {
				ctx.SendChain(message.Text("啊咧？我忘掉什么了吗"))
				return
			}
			ctx.SendChain(message.Text(
				title,
				strings.Join(r, " | "),
			))
		})
	zero.OnMessage().SetBlock(false).SetPriority(9999).
		Handle(func(ctx *zero.Ctx) {
			BotQQ = fmt.Sprint(ctx.Event.SelfID)
			m := message.UnescapeCQCodeText(ctx.Event.RawEvent.Get("message").Str)
			// 角色问
			if answer := QACharPool.get(CharIndex[Char[ctx.Event.GroupID]], m); answer != "" {
				ctx.Send(answer)
				return
			}
			// 有人问
			if answer := QAGroupPool.get(ctx.Event.GroupID, m); answer != "" {
				ctx.Send(answer)
				return
			}
			// 我问
			if answer := QAUserPool.get(ctx.Event.UserID, m); answer != "" {
				ctx.Send(answer)
				return
			}
		})
}

func botPath() string {
	dir, _ := os.Getwd()
	return dir + "/"
}

func (qa *QA) load() {
	path := qa.Path
	idx := strings.LastIndex(qa.Path, "/")
	if idx != -1 {
		path = path[:idx]
	}
	_, err := os.Stat(path)
	if err != nil && !os.IsExist(err) {
		os.MkdirAll(path, 0644)
	}
	f, _ := os.Open(qa.Path)
	defer f.Close()
	data, _ := ioutil.ReadAll(f)
	qa.Mutex.Lock()
	defer qa.Mutex.Unlock()
	if data == nil {
		data = []byte("0: \n  test: ok")
	}
	json.Unmarshal(data, &qa.Data)
}

func (qa *QA) save() {
	path := qa.Path
	idx := strings.LastIndex(qa.Path, "/")
	if idx != -1 {
		path = path[:idx]
	}
	_, err := os.Stat(path)
	if err != nil && !os.IsExist(err) {
		os.MkdirAll(path, 0644)
	}
	data, _ := json.MarshalIndent(&qa.Data, "", "\t")
	ioutil.WriteFile(qa.Path, data, 0644)
}

func (qa *QA) add(user int64, question, answer string) {
	qa.Mutex.Lock()
	defer qa.Mutex.Unlock()
	if qa.Data[user] == nil { // 防止未初始化
		qa.Data[user] = make(map[string]string)
	}
	qa.Data[user][question] = answer
	qa.save()
}

func (qa *QA) del(user int64, question string) (answer string) {
	qa.Mutex.Lock()
	defer qa.Mutex.Unlock()
	if answer, ok := qa.Data[user][question]; ok {
		delete(qa.Data[user], question)
		qa.save()
		return answer
	}
	return ""
}

func (qa *QA) get(user int64, msg string) (answer string) {
	for question, answer := range qa.Data[user] {
		// 替换BOT
		question = strings.ReplaceAll(question, "BOT_QQ", BotQQ)
		r := regexp.MustCompile("^" + question + "$")
		if r.MatchString(msg) {
			match := r.FindStringSubmatch(msg)
			// 正则替换参数
			for i, p := range match {
				if p == "[" || p == "]" {
					continue
				}
				answer = strings.ReplaceAll(answer, fmt.Sprintf("$%d", i), p)
			}
			// 替换图片路径
			answer = strings.ReplaceAll(answer, "BOT_PATH", BotPath)
			// 随机回复
			if strings.Contains(answer, "|") {
				s := strings.Split(answer, "|")
				return s[rand.Intn(len(s))]
			}
			return answer
		}
	}
	return ""
}

func down(url, path string) (filename string, err error) {
	client := &http.Client{}
	reqest, _ := http.NewRequest("GET", url, nil)
	reqest.Header.Set("User-Agent", "QQ/8.2.0.1296 CFNetwork/1126")
	reqest.Header.Set("Net-Type", "Wifi")
	resp, err := client.Do(reqest)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("code %d", resp.StatusCode)
	}
	data, _ := ioutil.ReadAll(resp.Body)
	// 获取文件MD5值
	m := md5.New()
	m.Write(data)
	filename = strings.ToUpper(hex.EncodeToString(m.Sum(nil)))
	// 判断文件类型
	switch resp.Header.Get("Content-Type") {
	case "image/jpeg":
		filename = filename + ".jpg"
	case "image/png":
		filename = filename + ".png"
	case "image/gif":
		filename = filename + ".gif"
	}
	// 保存文件
	_, err = os.Stat(path)
	if err != nil && !os.IsExist(err) {
		os.MkdirAll(path, 0644)
	}
	f, err := os.OpenFile(path+"/"+filename, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return "", err
	}
	defer f.Close()
	f.Write(data)
	return filename, nil
}

// QAMatch 返回对应的权限
func QAMatch() zero.Rule {
	return func(ctx *zero.Ctx) bool {
		switch ctx.State["regex_matched"].([]string)[1] {
		case "角色":
			ctx.State["qa_pool"] = QACharPool
			ctx.State["qa_user"] = CharIndex[Char[ctx.Event.GroupID]]
			return true
		case "有人":
			ctx.State["qa_pool"] = QAGroupPool
			ctx.State["qa_user"] = ctx.Event.GroupID
			return true
		case "我":
			ctx.State["qa_pool"] = QAUserPool
			ctx.State["qa_user"] = ctx.Event.UserID
			return true
		}
		return false
	}
}

// QAPermission 返回对应的权限
func QAPermission() zero.Rule {
	return func(ctx *zero.Ctx) bool {
		switch ctx.State["regex_matched"].([]string)[1] {
		case "角色":
			return zero.AdminPermission(ctx)
		case "有人":
			return zero.AdminPermission(ctx)
		case "我":
			return true
		}
		return false
	}
}
