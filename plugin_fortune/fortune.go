// Package fortune 每日运势
package fortune

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"image/jpeg"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/fogleman/gg"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"

	"github.com/FloatTech/ZeroBot-Plugin/control"
	"github.com/FloatTech/ZeroBot-Plugin/utils/file"
	"github.com/FloatTech/ZeroBot-Plugin/utils/math"
)

var (
	// 底图缓存位置
	base = "data/fortune/"
	// 素材下载网站
	site = "https://pan.dihe.moe/fortune/"
	// 底图类型列表：车万 DC4 爱因斯坦 星空列车 樱云之恋 富婆妹 李清歌
	// 				公主连结 原神 明日方舟 碧蓝航线 碧蓝幻想 战双 阴阳师
	table = [...]string{"车万", "DC4", "爱因斯坦", "星空列车", "樱云之恋", "富婆妹", "李清歌", "公主连结", "原神", "明日方舟", "碧蓝航线", "碧蓝幻想", "战双", "阴阳师"}
	// 映射底图与 index
	index = make(map[string]uint8)
	// 下载锁
	dlmu sync.Mutex
)

func init() {
	for i, s := range table {
		index[s] = uint8(i)
	}
	err := os.MkdirAll(base, 0755)
	if err != nil {
		panic(err)
	}
	// 插件主体
	en := control.Register("fortune", &control.Options{
		DisableOnDefault: false,
		Help: "每日运势: \n" +
			"- 运势|抽签\n" +
			"- 设置底图[车万 DC4 爱因斯坦 星空列车 樱云之恋 富婆妹 李清歌 公主连结 原神 明日方舟 碧蓝航线 碧蓝幻想 战双 阴阳师]",
	})
	en.OnRegex(`^设置底图(.*)`).SetBlock(true).SecondPriority().
		Handle(func(ctx *zero.Ctx) {
			gid := ctx.Event.GroupID
			if gid <= 0 {
				// 个人用户设为负数
				gid = -ctx.Event.UserID
			}
			i, ok := index[ctx.State["regex_matched"].([]string)[1]]
			if ok {
				c, ok := control.Lookup("fortune")
				if ok {
					err = c.SetData(gid, int64(i)&0xff)
					if err != nil {
						ctx.SendChain(message.Text("设置失败:", err))
						return
					}
					ctx.SendChain(message.Text("设置成功~"))
					return
				}
				ctx.SendChain(message.Text("设置失败: 找不到插件"))
				return
			}
			ctx.SendChain(message.Text("没有这个底图哦～"))
		})
	en.OnFullMatchGroup([]string{"运势", "抽签"}).SetBlock(true).SecondPriority().
		Handle(func(ctx *zero.Ctx) {
			// 检查签文文件是否存在
			mikuji := base + "运势签文.json"
			if file.IsNotExist(mikuji) {
				dlmu.Lock()
				if file.IsNotExist(mikuji) {
					ctx.SendChain(message.Text("正在下载签文文件，请稍后..."))
					err := file.DownloadTo(site+"运势签文.json", mikuji, false)
					if err != nil {
						ctx.SendChain(message.Text("ERROR: ", err))
						return
					}
					ctx.SendChain(message.Text("下载签文文件完毕"))
				}
				dlmu.Unlock()
			}
			// 检查字体文件是否存在
			ttf := base + "sakura.ttf"
			if file.IsNotExist(ttf) {
				dlmu.Lock()
				if file.IsNotExist(ttf) {
					ctx.SendChain(message.Text("正在下载字体文件，请稍后..."))
					err := file.DownloadTo(site+"sakura.ttf", ttf, false)
					if err != nil {
						ctx.SendChain(message.Text("ERROR: ", err))
						return
					}
					ctx.SendChain(message.Text("下载字体文件完毕"))
				}
				dlmu.Unlock()
			}
			// 获取该群背景类型，默认车万
			kind := "车万"
			gid := ctx.Event.GroupID
			if gid <= 0 {
				// 个人用户设为负数
				gid = -ctx.Event.UserID
			}
			logrus.Debugln("[fortune]gid:", ctx.Event.GroupID, "uid:", ctx.Event.UserID)
			c, ok := control.Lookup("fortune")
			if ok {
				v := uint8(c.GetData(gid) & 0xff)
				if int(v) < len(table) {
					kind = table[v]
				}
			}
			// 检查背景图片是否存在
			folder := base + kind
			if file.IsNotExist(folder) {
				dlmu.Lock()
				if file.IsNotExist(folder) {
					ctx.SendChain(message.Text("正在下载背景图片，请稍后..."))
					zipfile := kind + ".zip"
					zipcache := base + zipfile
					err := file.DownloadTo(site+zipfile, zipcache, false)
					if err != nil {
						ctx.SendChain(message.Text("ERROR: ", err))
						return
					}
					ctx.SendChain(message.Text("下载背景图片完毕"))
					err = unpack(zipcache, folder+"/")
					if err != nil {
						ctx.SendChain(message.Text("ERROR: ", err))
						return
					}
					ctx.SendChain(message.Text("解压背景图片完毕"))
					// 释放空间
					os.Remove(zipcache)
				}
				dlmu.Unlock()
			}
			// 生成种子
			t, _ := strconv.ParseInt(time.Now().Format("20060102"), 10, 64)
			seed := ctx.Event.UserID + t
			// 随机获取背景
			background, err := randimage(folder+"/", seed)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			// 随机获取签文
			title, text, err := randtext(mikuji, seed)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			// 绘制背景
			d, err := draw(background, title, text)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			// 发送图片
			ctx.SendChain(message.Image("base64://" + helper.BytesToString(d)))
		})
}

// @function unpack 解压资源包
// @param tgt 压缩文件位置
// @param dest 解压位置
// @return 错误信息
func unpack(tgt, dest string) error {
	// 路径目录不存在则创建目录
	if file.IsNotExist(dest) {
		if err := os.MkdirAll(dest, 0755); err != nil {
			panic(err)
		}
	}
	reader, err := zip.OpenReader(tgt)
	if err != nil {
		return err
	}
	defer reader.Close()
	// 遍历解压到文件
	for _, file := range reader.File {
		// 打开解压文件
		rc, err := file.Open()
		if err != nil {
			return err
		}
		// 打开目标文件
		w, err := os.OpenFile(dest+file.Name, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
		if err != nil {
			rc.Close()
			return err
		}
		// 复制到文件
		_, err = io.Copy(w, rc)
		rc.Close()
		w.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// @function randimage 随机选取文件夹下的文件
// @param path 文件夹路径
// @param seed 随机数种子
// @return 文件路径 & 错误信息
func randimage(path string, seed int64) (string, error) {
	rd, err := ioutil.ReadDir(path)
	if err != nil {
		return "", err
	}
	rand.Seed(seed)
	return path + rd[rand.Intn(len(rd))].Name(), nil
}

// @function randtext 随机选取签文
// @param file 文件路径
// @param seed 随机数种子
// @return 签名 & 签文 & 错误信息
func randtext(file string, seed int64) (string, string, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return "", "", err
	}
	temp := []map[string]string{}
	if err := json.Unmarshal(data, &temp); err != nil {
		return "", "", err
	}
	rand.Seed(seed)
	r := rand.Intn(len(temp))
	return temp[r]["title"], temp[r]["content"], nil
}

// @function draw 绘制运势图
// @param background 背景图片路径
// @param seed 随机数种子
// @param title 签名
// @param text 签文
// @return 错误信息
func draw(background, title, text string) ([]byte, error) {
	// 加载背景
	back, err := gg.LoadImage(background)
	if err != nil {
		return nil, err
	}
	canvas := gg.NewContext(back.Bounds().Size().Y, back.Bounds().Size().X)
	canvas.DrawImage(back, 0, 0)
	// 写标题
	canvas.SetRGB(1, 1, 1)
	if err := canvas.LoadFontFace(base+"sakura.ttf", 45); err != nil {
		return nil, err
	}
	sw, _ := canvas.MeasureString(title)
	canvas.DrawString(title, 140-sw/2, 112)
	// 写正文
	canvas.SetRGB(0, 0, 0)
	if err := canvas.LoadFontFace(base+"sakura.ttf", 23); err != nil {
		return nil, err
	}
	tw, th := canvas.MeasureString("测")
	tw, th = tw+10, th+10
	r := []rune(text)
	xsum := rowsnum(len(r), 9)
	switch xsum {
	default:
		for i, o := range r {
			xnow := rowsnum(i+1, 9)
			ysum := math.Min(len(r)-(xnow-1)*9, 9)
			ynow := i%9 + 1
			canvas.DrawString(string(o), -offest(xsum, xnow, tw)+115, offest(ysum, ynow, th)+320.0)
		}
	case 2:
		div := rowsnum(len(r), 2)
		for i, o := range r {
			xnow := rowsnum(i+1, div)
			ysum := math.Min(len(r)-(xnow-1)*div, div)
			ynow := i%div + 1
			switch xnow {
			case 1:
				canvas.DrawString(string(o), -offest(xsum, xnow, tw)+115, offest(9, ynow, th)+320.0)
			case 2:
				canvas.DrawString(string(o), -offest(xsum, xnow, tw)+115, offest(9, ynow+(9-ysum), th)+320.0)
			}
		}
	}
	// 转成 base64
	buffer := new(bytes.Buffer)
	encoder := base64.NewEncoder(base64.StdEncoding, buffer)
	var opt jpeg.Options
	opt.Quality = 70
	err = jpeg.Encode(encoder, canvas.Image(), &opt)
	if err != nil {
		return nil, err
	}
	encoder.Close()
	return buffer.Bytes(), nil
}

func offest(total, now int, distance float64) float64 {
	if total%2 == 0 {
		return (float64(now-total/2) - 1) * distance
	}
	return (float64(now-total/2) - 1.5) * distance
}

func rowsnum(total, div int) int {
	temp := total / div
	if total%div != 0 {
		temp++
	}
	return temp
}
