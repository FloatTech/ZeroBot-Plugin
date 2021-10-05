// Package fortune 每日运势
package fortune

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image/jpeg"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fogleman/gg"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/ZeroBot-Plugin/control"
)

var (
	// 底图缓存位置
	base = "data/fortune/"
	// int64 群号 string 底图类型
	// 底图类型列表：车万 DC4 爱因斯坦 星空列车 樱云之恋 富婆妹 李清歌
	// 				公主连结 原神 明日方舟 碧蓝航线 碧蓝幻想 战双 阴阳师
	table = map[int64]string{
		0:          "车万",
		1048452984: "爱因斯坦",
	}
)

func init() {
	// 插件主体
	control.Register("runcode", &control.Options{
		DisableOnDefault: false,
		Help: "每日运势: \n" +
			"- 运势",
	}).OnFullMatchGroup([]string{"运势", "抽签"}).SetBlock(true).SecondPriority().
		Handle(func(ctx *zero.Ctx) {
			// 检查签文文件是否存在
			if _, err := os.Stat(base + "运势签文.json"); err != nil && !os.IsExist(err) {
				ctx.SendChain(message.Text("正在下载签文文件，请稍后..."))
				_, err := download("https://pan.dihe.moe/fortune/运势签文.json", base)
				if err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
					return
				}
				ctx.SendChain(message.Text("下载签文文件完毕"))
			}
			// 检查字体文件是否存在
			if _, err := os.Stat(base + "sakura.ttf"); err != nil && !os.IsExist(err) {
				ctx.SendChain(message.Text("正在下载字体文件，请稍后..."))
				_, err := download("https://pan.dihe.moe/fortune/sakura.ttf", base)
				if err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
					return
				}
				ctx.SendChain(message.Text("下载字体文件完毕"))
			}
			// 获取该群背景类型
			var kind string
			if v, ok := table[ctx.Event.GroupID]; ok {
				kind = v
			} else {
				kind = table[0]
			}
			// 检查背景图片是否存在
			if _, err := os.Stat(base + kind); err != nil && !os.IsExist(err) {
				ctx.SendChain(message.Text("正在下载背景图片，请稍后..."))
				file, err := download("https://pan.dihe.moe/fortune/"+kind+".zip", base)
				if err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
					return
				}
				ctx.SendChain(message.Text("下载背景图片完毕"))
				err = unpack(file, base+kind+"/")
				if err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
					return
				}
				ctx.SendChain(message.Text("解压背景图片完毕"))
			}
			// 生成种子
			t, _ := strconv.ParseInt(time.Now().Format("20060102"), 10, 64)
			seed := ctx.Event.UserID + t
			// 随机获取背景
			background, err := randimage(base+kind+"/", seed)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			// 随机获取签文
			title, text, err := randtext(base+"运势签文.json", seed)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			// 绘制背景
			data, err := draw(background, title, text)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			// 发送图片
			ctx.SendChain(message.Image("base64://" + string(data)))
		})
}

// @function download 下载资源包
// @param link 下载链接
// @param dest 下载位置
// @return 文件路径 & 错误信息
func download(link, dest string) (string, error) {
	// 路径目录不存在则创建目录
	if _, err := os.Stat(dest); err != nil && !os.IsExist(err) {
		if err := os.MkdirAll(dest, 0755); err != nil {
			panic(err)
		}
	}
	client := &http.Client{}
	// 网络请求
	request, _ := http.NewRequest("GET", link, nil)
	request.Header.Set("Accept", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:6.0) Gecko/20100101 Firefox/6.0")
	resp, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	// 验证接收到的长度
	length, _ := strconv.Atoi(resp.Header.Get("Content-Length"))
	data, _ := ioutil.ReadAll(resp.Body)
	if length > len(data) {
		fmt.Println(len(data))
		fmt.Println(length)
		return "", errors.New("download not complete")
	}
	// 获取文件名
	temp := strings.Split(resp.Header.Get("Content-Disposition"), "\"")
	name, _ := url.QueryUnescape(temp[len(temp)-2])
	// 写入文件
	f, err := os.OpenFile(dest+name, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return "", err
	}
	defer f.Close()
	_, err = f.Write(data)
	if err != nil {
		return "", err
	}
	return dest + name, nil
}

// @function unpack 解压资源包
// @param tgt 压缩文件位置
// @param dest 解压位置
// @return 错误信息
func unpack(tgt, dest string) error {
	// 路径目录不存在则创建目录
	if _, err := os.Stat(dest); err != nil && !os.IsExist(err) {
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
	offest := func(total, now int, distance float64) float64 {
		if total%2 == 0 {
			return (float64(now-total/2) - 1) * distance
		}
		return (float64(now-total/2) - 1.5) * distance
	}
	rowsnum := func(total, div int) int {
		temp := total / div
		if total%div != 0 {
			temp++
		}
		return temp
	}
	min := func(a, b int) int {
		if a < b {
			return a
		}
		return b
	}
	tw, th := canvas.MeasureString("测")
	tw, th = tw+10, th+10
	r := []rune(text)
	xsum := rowsnum(len(r), 9)
	switch xsum {
	default:
		for i, o := range r {
			xnow := rowsnum(i+1, 9)
			ysum := min(len(r)-(xnow-1)*9, 9)
			ynow := i%9 + 1
			canvas.DrawString(string(o), -offest(xsum, xnow, tw)+115, offest(ysum, ynow, th)+320.0)
		}
	case 2:
		div := rowsnum(len(r), 2)
		for i, o := range r {
			xnow := rowsnum(i+1, div)
			ysum := min(len(r)-(xnow-1)*div, div)
			ynow := i%div + 1
			switch xnow {
			case 1:
				canvas.DrawString(string(o), -offest(xsum, xnow, tw)+115, offest(9, ynow, th)+320.0)
			case 2:
				canvas.DrawString(string(o), -offest(xsum, xnow, tw)+115, offest(9, ynow+(9-ysum), th)+320.0)
			}
		}
	}
	// 保存
	buffer := new(bytes.Buffer)
	encoder := base64.NewEncoder(base64.StdEncoding, buffer)
	var opt jpeg.Options
	opt.Quality = 70
	err = jpeg.Encode(encoder, canvas.Image(), &opt)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
