package job

import (
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/FloatTech/floatbox/binary"
	"github.com/FloatTech/floatbox/process"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type (
	// Type 任务类型,1=指令别名,2=定时任务,3=你问我答
	Type uint8
	// FullMatchType 指令别名类型, jobType=1使用的参数, 1=无状态消息, 2=主人消息
	FullMatchType uint8
	// QuestionType 问题类型, jobType=3使用的参数, 1=单独问, 2=所有人问
	QuestionType uint8
	// AnswerType 回答类型, jobType=3使用的参数, 1=文本消息, 2=注入消息
	AnswerType uint8
)

const (
	// FullMatchJob 指令别名
	FullMatchJob Type = iota + 1
	// CronJob 定时任务
	CronJob
	// RegexpJob 你问我答
	RegexpJob
)

const (
	// NoStateMsg 无状态消息
	NoStateMsg FullMatchType = iota + 1
	// SuperMsg 主人消息
	SuperMsg
)

const (
	// OneQuestion 单独问
	OneQuestion QuestionType = iota + 1
	// AllQuestion 所有人问
	AllQuestion
)

const (
	// TextMsg 文本消息
	TextMsg AnswerType = iota + 1
	// InjectMsg 注入消息
	InjectMsg
)

// Job 添加任务的入参
//
//	@Description	添加任务的入参
type Job struct {
	ID            string        `json:"id"`            // 任务id
	SelfID        int64         `json:"selfId"`        // 机器人id
	JobType       Type          `json:"jobType"`       // 任务类型,1-指令别名,2-定时任务,3-你问我答
	Matcher       string        `json:"matcher"`       // 当jobType=1时 为指令别名,当jobType=2时 为cron表达式,当jobType=3时 为正则表达式
	Handler       string        `json:"handler"`       // 执行内容
	FullMatchType FullMatchType `json:"fullMatchType"` // 指令别名类型, jobType=1使用的参数, 1=无状态消息, 2=主人消息
	QuestionType  QuestionType  `json:"questionType"`  // 问题类型, jobType=3使用的参数, 1=单独问, 2=所有人问
	AnswerType    AnswerType    `json:"answerType"`    // 回答类型, jobType=3使用的参数, 1=文本消息, 2=注入消息
	GroupID       int64         `json:"groupId"`       // 群聊id, jobType=2,3使用的参数, jobType=2且私聊, group_id=0
	UserID        int64         `json:"userId"`        // 用户id, jobType=2,3使用的参数, 当jobType=3, QuestionType=2,userId=0
}

// List 任务列表
func List() (jobList []Job, err error) {
	jobList = make([]Job, 0, 16)
	zero.RangeBot(func(id int64, ctx *zero.Ctx) bool {
		c := &cmd{}
		ids := strconv.FormatInt(id, 36)
		_ = db.FindFor(ids, c, "", func() error {
			var j Job
			var e zero.Event
			j.SelfID = id
			j.ID = strconv.FormatInt(c.ID, 10)
			if len(c.Cron) >= 3 {
				switch c.Cron[:3] {
				case "sm:":
					j.JobType = FullMatchJob
					j.FullMatchType = SuperMsg
					j.Matcher = c.Cron[3:]
					err = json.Unmarshal(binary.StringToBytes(c.Cmd), &e)
					if err != nil {
						return err
					}
					j.Handler = e.RawMessage
				case "fm:":
					j.JobType = FullMatchJob
					j.FullMatchType = NoStateMsg
					j.Matcher = c.Cron[3:]
					j.Handler = c.Cmd
				case "rm:":
					j.JobType = RegexpJob
					j.QuestionType = AllQuestion
					j.AnswerType = TextMsg
					j.Handler = message.UnescapeCQCodeText(c.Cmd)
					cutList := strings.SplitN(c.Cron, ":", 3)
					if len(cutList) == 3 {
						j.GroupID, err = strconv.ParseInt(cutList[len(cutList)-2], 36, 64)
						if err != nil {
							return err
						}
						j.Matcher = cutList[len(cutList)-1]
					}
				case "rp:":
					j.JobType = RegexpJob
					j.QuestionType = OneQuestion
					j.AnswerType = TextMsg
					j.Handler = message.UnescapeCQCodeText(c.Cmd)
					cutList := strings.SplitN(c.Cron, ":", 4)
					if len(cutList) == 4 {
						j.UserID, err = strconv.ParseInt(cutList[len(cutList)-3], 36, 64)
						if err != nil {
							return err
						}
						j.GroupID, err = strconv.ParseInt(cutList[len(cutList)-2], 36, 64)
						if err != nil {
							return err
						}
						j.Matcher = cutList[len(cutList)-1]
					}
				case "im:":
					j.JobType = RegexpJob
					j.QuestionType = AllQuestion
					j.AnswerType = InjectMsg
					j.Handler = message.UnescapeCQCodeText(c.Cmd)
					cutList := strings.SplitN(c.Cron, ":", 3)
					if len(cutList) == 3 {
						j.GroupID, err = strconv.ParseInt(cutList[len(cutList)-2], 36, 64)
						if err != nil {
							return err
						}
						j.Matcher = cutList[len(cutList)-1]
					}
				case "ip:":
					j.JobType = RegexpJob
					j.QuestionType = OneQuestion
					j.AnswerType = InjectMsg
					j.Handler = message.UnescapeCQCodeText(c.Cmd)
					cutList := strings.SplitN(c.Cron, ":", 4)
					if len(cutList) == 4 {
						j.UserID, err = strconv.ParseInt(cutList[len(cutList)-3], 36, 64)
						if err != nil {
							return err
						}
						j.GroupID, err = strconv.ParseInt(cutList[len(cutList)-2], 36, 64)
						if err != nil {
							return err
						}
						j.Matcher = cutList[len(cutList)-1]
					}
				default:
					j.JobType = CronJob
					j.Matcher = c.Cron
					err = json.Unmarshal(binary.StringToBytes(c.Cmd), &e)
					if err != nil {
						return err
					}
					j.Handler = e.RawMessage
					j.GroupID = e.GroupID
					j.UserID = e.UserID
				}
			}
			jobList = append(jobList, j)
			return nil
		})
		// 不能打断循环
		if err != nil {
			logrus.Errorln("jobList: ", err)
		}
		return true
	})
	return
}

// Add 添加任务
func Add(j *Job) (err error) {
	var (
		c        cmd
		bot      *zero.Ctx
		b        []byte
		compiled *regexp.Regexp
	)
	bot = zero.GetBot(j.SelfID)
	bot.Event = &zero.Event{
		SelfID: j.SelfID,
	}
	switch j.JobType {
	case FullMatchJob:
		if j.FullMatchType == 1 {
			c.Cron = "fm:" + j.Matcher
			c.Cmd = binary.BytesToString(json.RawMessage("\"" + j.Handler + "\""))
		} else {
			c.Cron = "sm:" + j.Matcher
			var e zero.Event
			if len(zero.BotConfig.SuperUsers) > 0 {
				e.UserID = zero.BotConfig.SuperUsers[0]
				e.Sender = &zero.User{
					ID: zero.BotConfig.SuperUsers[0],
				}
				e.RawMessage = j.Handler
				e.NativeMessage = json.RawMessage("\"" + j.Handler + "\"")
			}
			b, err = json.Marshal(&e)
			if err != nil {
				return
			}
			c.Cmd = binary.BytesToString(b)
		}
		c.ID = idof(c.Cron, c.Cmd)
		err = registercmd(j.SelfID, &c)
		if err != nil {
			return
		}
	case CronJob:
		var e zero.Event
		e.UserID = j.UserID
		e.Sender = &zero.User{
			ID: j.UserID,
		}
		e.SelfID = j.SelfID
		e.RawMessage = j.Handler
		e.NativeMessage = json.RawMessage("\"" + j.Handler + "\"")
		e.GroupID = j.GroupID
		e.PostType = "message"
		if e.GroupID > 0 {
			e.MessageType = "group"
		} else {
			e.MessageType = "private"
			e.TargetID = j.SelfID
		}
		b, err = json.Marshal(&e)
		if err != nil {
			return
		}
		c.Cmd = binary.BytesToString(b)
		c.Cron = j.Matcher
		c.ID = idof(c.Cron, c.Cmd)
		err = addcmd(bot, &c)
		if err != nil {
			return
		}
	case RegexpJob:
		all := false
		isInject := false
		gid := j.GroupID
		uid := j.UserID
		switch {
		case j.QuestionType == AllQuestion && j.AnswerType == InjectMsg:
			all = true
			isInject = true
			c.Cron = "im:" + strconv.FormatInt(gid, 36) + ":" + j.Matcher
		case j.QuestionType == AllQuestion && j.AnswerType == TextMsg:
			all = true
			isInject = false
			c.Cron = "rm:" + strconv.FormatInt(gid, 36) + ":" + j.Matcher
		case j.QuestionType == OneQuestion && j.AnswerType == InjectMsg:
			all = false
			isInject = true
			c.Cron = "ip:" + strconv.FormatInt(uid, 36) + ":" + strconv.FormatInt(gid, 36) + ":" + j.Matcher
		case j.QuestionType == OneQuestion && j.AnswerType == TextMsg:
			all = false
			isInject = false
			c.Cron = "rp:" + strconv.FormatInt(uid, 36) + ":" + strconv.FormatInt(gid, 36) + ":" + j.Matcher
		default:
		}
		c.Cmd = message.EscapeCQCodeText(j.Handler)
		c.ID = idof(c.Cron, c.Cmd)
		pattern := j.Matcher
		template := message.EscapeCQCodeText(j.Handler)
		if global.group[gid] == nil {
			global.group[gid] = &regexGroup{
				Private: make(map[int64][]inst),
			}
		}
		if global.group[gid].Private == nil {
			global.group[gid].Private = make(map[int64][]inst)
		}
		compiled, err = regexp.Compile(transformPattern(pattern))
		if err != nil {
			return
		}
		regexInst := inst{
			regex:    compiled,
			Pattern:  pattern,
			Template: template,
			IsInject: isInject,
		}
		rg := global.group[gid]
		if all {
			if isInject {
				err = saveInjectRegex(gid, 0, strconv.FormatInt(j.SelfID, 36), pattern, template)
			} else {
				err = saveRegex(gid, 0, strconv.FormatInt(j.SelfID, 36), pattern, template)
			}
			if err == nil {
				rg.All = append(rg.All, regexInst)
			}
		} else {
			if isInject {
				err = saveInjectRegex(gid, uid, strconv.FormatInt(j.SelfID, 36), pattern, template)
			} else {
				err = saveRegex(gid, uid, strconv.FormatInt(j.SelfID, 36), pattern, template)
			}
			if err == nil {
				rg.Private[uid] = append(rg.Private[uid], regexInst)
			}
		}
		if err != nil {
			return
		}
	default:
		err = errors.New("不存在的任务类型")
		return
	}
	return
}

// DeleteReq 删除任务的入参
//
//	@Description	删除任务的入参
type DeleteReq struct {
	IDList []string `json:"idList" form:"idList"` // 任务id
	SelfID int64    `json:"selfId" form:"selfId"` // 机器人qq
}

// Delete 删除任务
func Delete(req *DeleteReq) (err error) {
	var (
		c cmd
	)
	mu.Lock()
	defer mu.Unlock()
	bots := strconv.FormatInt(req.SelfID, 36)
	var delcmd []string
	err = db.FindFor(bots, &c, "WHERE id in ( "+strings.Join(req.IDList, ",")+" )", func() error {
		switch {
		case len(c.Cron) >= 3 && (c.Cron[:3] == "fm:" || c.Cron[:3] == "sm:"):
			m, ok := matchers[c.ID]
			if ok {
				m.Delete()
				delete(matchers, c.ID)
			}
		case len(c.Cron) >= 3 && (c.Cron[:3] == "ip:" || c.Cron[:3] == "rp:" || c.Cron[:3] == "rm:" || c.Cron[:3] == "im:"):
			var (
				all     bool
				gid     int64
				uid     int64
				pattern string
			)
			if len(c.Cron) >= 3 && (c.Cron[:3] == "ip:" || c.Cron[:3] == "rp:") {
				cutList := strings.SplitN(c.Cron, ":", 4)
				if len(cutList) == 4 {
					uid, err = strconv.ParseInt(cutList[len(cutList)-3], 36, 64)
					if err != nil {
						return err
					}
					gid, err = strconv.ParseInt(cutList[len(cutList)-2], 36, 64)
					if err != nil {
						return err
					}
					pattern = cutList[len(cutList)-1]
				}
				all = false
			} else {
				cutList := strings.SplitN(c.Cron, ":", 3)
				if len(cutList) == 3 {
					gid, err = strconv.ParseInt(cutList[len(cutList)-2], 36, 64)
					if err != nil {
						return err
					}
				}
				all = true
				pattern = cutList[len(cutList)-1]
			}
			escapedpattern := message.UnescapeCQCodeText(pattern)
			if pattern == escapedpattern {
				escapedpattern = ""
			}
			rg := global.group[gid]
			if rg == nil {
				return nil
			}
			var deleteInst func(insts []inst) ([]inst, error)
			if escapedpattern == "" {
				deleteInst = func(insts []inst) ([]inst, error) {
					for i := range insts {
						if insts[i].Pattern == pattern {
							insts[i] = insts[len(insts)-1]
							insts = insts[:len(insts)-1]
							return insts, nil
						}
					}
					return insts, errors.New("没有找到对应的问答词条")
				}
			} else {
				deleteInst = func(insts []inst) ([]inst, error) {
					for i := range insts {
						if insts[i].Pattern == pattern || insts[i].Pattern == escapedpattern {
							insts[i] = insts[len(insts)-1]
							insts = insts[:len(insts)-1]
							return insts, nil
						}
					}
					return insts, errors.New("没有找到对应的问答词条")
				}
			}
			if all {
				rg.All, err = deleteInst(rg.All)
			} else {
				rg.Private[uid], err = deleteInst(rg.Private[uid])
			}
			if err != nil {
				return err
			}
		default:
			eid, ok := entries[c.ID]
			if ok {
				process.CronTab.Remove(eid)
				delete(entries, c.ID)
			}
		}
		delcmd = append(delcmd, "id="+strconv.FormatInt(c.ID, 10))
		return nil
	})
	if err != nil {
		return
	}
	if len(delcmd) > 0 {
		err = db.Del(bots, "WHERE "+strings.Join(delcmd, " or "))
		if err != nil {
			return
		}
	}
	return
}
