// Package niuniu 牛牛大作战
package niuniu

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"sync"
	"time"

	fcext "github.com/FloatTech/floatbox/ctxext"
	sql "github.com/FloatTech/sqlite"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type model struct {
	sql sql.Sqlite
	sync.RWMutex
}

type userInfo struct {
	UID       int64
	Length    float64
	UserCount int
	WeiGe     int // 伟哥
	Philter   int // 媚药
	Artifact  int // 击剑神器
	ShenJi    int // 击剑神稽
	Buff1     int // 暂定
	Buff2     int // 暂定
	Buff3     int // 暂定
	Buff4     int // 暂定
	Buff5     int // 暂定
}

type users []*userInfo

var (
	db    = &model{}
	getdb = fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		db.sql.DBPath = en.DataFolder() + "niuniu.db"
		err := db.sql.Open(time.Hour * 24)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return false
		}
		return true
	})
)

// useWeiGe 使用道具伟哥
func (u *userInfo) useWeiGe() (string, float64) {
	niuniu := u.Length
	reduce := math.Abs(hitGlue(niuniu))
	niuniu += reduce
	return randomChoice([]string{
		fmt.Sprintf("哈哈，你这一用道具，牛牛就像是被激发了潜能，增加了%.2fcm！看来今天是个大日子呢！", reduce),
		fmt.Sprintf("你这是用了什么神奇的道具？牛牛竟然增加了%.2fcm，简直是牛气冲天！", reduce),
		fmt.Sprintf("使用道具后，你的牛牛就像是开启了加速模式，一下增加了%.2fcm，这成长速度让人惊叹！", reduce),
	}), niuniu
}

// usePhilter 使用道具媚药
func (u *userInfo) usePhilter() (string, float64) {
	niuniu := u.Length
	reduce := math.Abs(hitGlue(niuniu))
	niuniu -= reduce
	return randomChoice([]string{
		fmt.Sprintf("你使用媚药,咿呀咿呀一下使当前长度发生了一些变化，当前长度%.2f", niuniu),
		fmt.Sprintf("看来你追求的是‘微观之美’，故意使用道具让牛牛凹进去了%.2fcm！", reduce),
		fmt.Sprintf("缩小奇迹’在你身上发生了，牛牛凹进去了%.2fcm，你的选择真是独特！", reduce),
	}), niuniu
}

// useArtifact 使用道具击剑神器
func (u *userInfo) useArtifact(adduserniuniu float64) (string, float64, float64) {
	myLength := u.Length
	difference := myLength - adduserniuniu
	var (
		change float64
	)
	if difference > 0 {
		change = hitGlue(myLength + adduserniuniu)
	} else {
		change = hitGlue((myLength + adduserniuniu) / 2)
	}
	myLength += change
	return randomChoice([]string{
		fmt.Sprintf("凭借神秘道具的力量，你让对方在你的长度面前俯首称臣！你的长度增加了%.2fcm，当前长度达到了%.2fcm", change, myLength),
		fmt.Sprintf("神器在手，天下我有！你使用道具后，长度猛增%.2fcm，现在的总长度是%.2fcm，无人能敌！", change, myLength),
		fmt.Sprintf("这就是道具的魔力！你轻松增加了%.2fcm，让对手望尘莫及，当前长度为%.2fcm！", change, myLength),
		fmt.Sprintf("道具一出，谁与争锋！你的长度因道具而增长%.2fcm，现在的长度是%.2fcm，霸气尽显！", change, myLength),
		fmt.Sprintf("使用道具的你，如同获得神助！你的长度增长了%.2fcm，达到%.2fcm的惊人长度，胜利自然到手！", change, myLength),
	}), myLength, adduserniuniu - change/1.3
}

// useShenJi 使用道具击剑神稽
func (u *userInfo) useShenJi(adduserniuniu float64) (string, float64, float64) {
	myLength := u.Length
	difference := myLength - adduserniuniu
	var (
		change float64
	)
	if difference > 0 {
		change = hitGlue(myLength + adduserniuniu)
	} else {
		change = hitGlue((myLength + adduserniuniu) / 2)
	}
	myLength -= change
	var r string
	if myLength > 0 {
		r = randomChoice([]string{
			fmt.Sprintf("哦吼！？看来你的牛牛因为使用了神秘道具而缩水了呢🤣🤣🤣！缩小了%.2fcm！", change),
			fmt.Sprintf("哈哈，看来这个道具有点儿调皮，让你的长度缩水了%.2fcm！现在你的长度是%.2fcm，下次可得小心使用哦！", change, myLength),
			fmt.Sprintf("使用道具后，你的牛牛似乎有点儿害羞，缩水了%.2fcm！现在的长度是%.2fcm，希望下次它能挺直腰板！", change, myLength),
			fmt.Sprintf("哎呀，这个道具的效果有点儿意外，你的长度减少了%.2fcm，现在只有%.2fcm了！下次选道具可得睁大眼睛！", change, myLength),
		})
	} else {
		r = randomChoice([]string{
			fmt.Sprintf("哦哟，小姐姐真是玩得一手好游戏，使用道具后数值又降低了%.2fcm，小巧得更显魅力！", change),
			fmt.Sprintf("看来小姐姐喜欢更加精致的风格，使用道具后，数值减少了%.2fcm，更加迷人了！", change),
			fmt.Sprintf("小姐姐的每一次变化都让人惊喜，使用道具后，数值减少了%.2fcm，更加优雅动人！", change),
			fmt.Sprintf("小姐姐这是在展示什么是真正的精致小巧，使用道具后，数值减少了%.2fcm，美得不可方物！", change),
		})
	}
	return r, myLength, adduserniuniu + 0.7*change
}

func (u *userInfo) processNiuNiuAction(t string, props string) (string, error) {
	var (
		messages string
		info     userInfo
		err      error
		f        float64
	)
	load, ok := prop.Load(t)
	info = *u
	if props != "" {
		if !contains(t, dajiaoProp) {
			return "", errors.New("道具不能混着用哦")
		}
		if err = u.createUserInfoByProps(props); err != nil {
			return "", err
		}
	}
	switch {
	case ok && load.Count > 1 && time.Since(load.TimeLimit) < time.Minute*8:
		messages, f = generateRandomStingTwo(u.Length)
		u.Length = f
		errMessage := fmt.Sprintf("你使用道具次数太快了，此次道具不会生效，等待%d再来吧", time.Minute*8-time.Since(load.TimeLimit))
		err = errors.New(errMessage)

	case u.WeiGe-info.WeiGe != 0:
		messages, f = u.useWeiGe()
		u.Length = f
		updateMap(t, true)

	case u.Philter-info.Philter != 0:
		messages, f = u.usePhilter()
		u.Length = f
		updateMap(t, true)

	default:
		messages, f = generateRandomStingTwo(u.Length)
		u.Length = f
	}
	return messages, err
}

func (u *userInfo) createUserInfoByProps(props string) error {
	var (
		err error
	)
	switch props {
	case "伟哥":
		if u.WeiGe > 0 {
			u.WeiGe--
		} else {
			err = errors.New("你还没有伟哥呢,不能使用")
		}
	case "媚药":
		if u.Philter > 0 {
			u.Philter--
		} else {
			err = errors.New("你还没有媚药呢,不能使用")
		}
	case "击剑神器":
		if u.Artifact > 0 {
			u.Artifact--
		} else {
			err = errors.New("你还没有击剑神器呢,不能使用")
		}
	case "击剑神稽":
		if u.ShenJi > 0 {
			u.ShenJi--
		} else {
			err = errors.New("你还没有击剑神稽呢,不能使用")
		}
	default:
		err = errors.New("道具不存在")
	}
	return err
}

// 接收值依次是 被jj用户的信息 记录gid和uid的字符串 道具名称
// 返回值依次是 要发送的消息 错误信息
func (u *userInfo) processJJuAction(adduserniuniu *userInfo, t string, props string) (string, error) {
	var (
		fencingResult string
		f             float64
		f1            float64
		info          userInfo
		err           error
	)
	v, ok := prop.Load(t)
	info = *u
	if props != "" {
		if !contains(t, jjProp) {
			return "", errors.New("道具不能混着用哦")
		}
		if err = u.createUserInfoByProps(props); err != nil {
			return "", err
		}
	}
	switch {
	case ok && v.Count > 1 && time.Since(v.TimeLimit) < time.Minute*8:
		fencingResult, f, f1 = fencing(u.Length, adduserniuniu.Length)
		u.Length = f
		adduserniuniu.Length = f1
		errMessage := fmt.Sprintf("你使用道具次数太快了，此次道具不会生效，等待%d再来吧", time.Minute*8-time.Since(v.TimeLimit))
		err = errors.New(errMessage)
	case u.ShenJi-info.ShenJi != 0:
		fencingResult, f, f1 = u.useShenJi(adduserniuniu.Length)
		u.Length = f
		adduserniuniu.Length = f1
		updateMap(t, true)
	case u.Artifact-info.Artifact != 0:
		fencingResult, f, f1 = u.useArtifact(adduserniuniu.Length)
		u.Length = f
		adduserniuniu.Length = f1
		updateMap(t, true)
	default:
		fencingResult, f, f1 = fencing(u.Length, adduserniuniu.Length)
		u.Length = f
		adduserniuniu.Length = f1
	}
	return fencingResult, err
}

func (u *userInfo) purchaseItem(n int) (int, error) {
	var (
		money int
		err   error
	)
	switch n {
	case 1:
		money = 300
		u.WeiGe += 5
	case 2:
		money = 300
		u.Philter += 5
	case 3:
		money = 500
		u.Artifact += 2
	case 4:
		money = 500
		u.ShenJi += 2
	default:
		err = errors.New("无效的选择")
	}
	return money, err
}

func (m users) positive() users {
	var m1 []*userInfo
	for _, i2 := range m {
		if i2.Length > 0 {
			m1 = append(m1, i2)
		}
	}
	return m1
}

func (m users) negative() users {
	var m1 []*userInfo
	for _, i2 := range m {
		if i2.Length <= 0 {
			m1 = append(m1, i2)
		}
	}
	return m1
}

func (m users) sort(isDesc bool) users {
	t := func(i, j int) bool {
		return m[i].Length < m[j].Length
	}
	if isDesc {
		t = func(i, j int) bool {
			return m[i].Length > m[j].Length
		}
	}
	sort.Slice(m, t)
	return m
}

func (m users) ranking(niuniu float64, uid int64) int {
	result := niuniu > 0
	for i, user := range m.sort(result) {
		if user.UID == uid {
			return i + 1
		}
	}
	return -1
}

func (db *model) randLength() float64 {
	return float64(rand.Intn(9)+1) + (float64(rand.Intn(100)) / 100)
}

func (db *model) createGIDTable(gid int64) error {
	db.Lock()
	defer db.Unlock()
	return db.sql.Create(strconv.FormatInt(gid, 10), &userInfo{})
}

// findNiuNiu 返回一个用户的牛牛信息
func (db *model) findNiuNiu(gid, uid int64) (userInfo, error) {
	db.RLock()
	defer db.RUnlock()
	u := userInfo{}
	err := db.sql.Find(strconv.FormatInt(gid, 10), &u, "where UID = "+strconv.FormatInt(uid, 10))
	return u, err
}

// insertNiuNiu 更新一个用户的牛牛信息
func (db *model) insertNiuNiu(u *userInfo, gid int64) error {
	db.Lock()
	defer db.Unlock()
	return db.sql.Insert(strconv.FormatInt(gid, 10), u)
}

func (db *model) deleteniuniu(gid, uid int64) error {
	db.Lock()
	defer db.Unlock()
	return db.sql.Del(strconv.FormatInt(gid, 10), "where UID = "+strconv.FormatInt(uid, 10))
}

func (db *model) readAllTable(gid int64) (users, error) {
	db.Lock()
	defer db.Unlock()
	a, err := sql.FindAll[userInfo](&db.sql, strconv.FormatInt(gid, 10), "where UserCount  = 0")
	return a, err
}
