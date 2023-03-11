package base

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/FloatTech/floatbox/binary"
	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/zbputils/img/text"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"archive/zip"
)

var (
	commandsOfPush = [...][]string{
		{"add", "--all"},
		{"commit", "-m", "\"Update\""},
		{"push", "-u", "origin", "master"},
	}
	commandsOfFetch = [...][]string{
		// {"remote", "add", "upstream", "git@github.com:FloatTech/ZeroBot-Plugin.git"},
		// {"remote", "-v"},
		{"fetch", "upstream", "master"},
		{"merge", "upstream/master"},
		{"push", "-u", "origin", "master"},
		{"pull", "--tags", "-r", "origin", "master"},
	}
	commandsOfZbp = [...][]string{
		{"-o", file.BOTPATH + "/go.mod", "https://raw.githubusercontent.com/FloatTech/ZeroBot-Plugin/master/go.mod"},
		{"-o", file.BOTPATH + "/go.sum", "https://raw.githubusercontent.com/FloatTech/ZeroBot-Plugin/master/go.sum"},
	}
)

func init() {
	// 备份
	zero.OnFullMatch("备份代码", zero.OnlyToMe, zero.SuperUserPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		err := fileZipTo(file.BOTPATH, file.BOTPATH+time.Now().Format("_2006_01_02_15_04_05")+".zip")
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		ctx.SendChain(message.Text("备份完成"))
	})
	// 更新zbp
	zero.OnFullMatch("上传代码", zero.OnlyToMe, zero.SuperUserPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		var msg []string
		var img []byte
		var err error
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		msg = append(msg, "开始上传GitHub")
		for _, command := range commandsOfPush {
			cmd := exec.Command("git", command...)
			msg = append(msg, "Command:", strings.Join(cmd.Args, " "))
			cmd.Dir = file.BOTPATH
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			err = cmd.Run()
			if err != nil {
				msg = append(msg, "StdErr:", err.Error(), "\n", stderr.String())
				// 输出图片
				img, err = text.RenderToBase64(strings.Join(msg, "\n"), text.BoldFontFile, 1280, 50)
				if err != nil {
					ctx.SendChain(message.Text("[ERROR]:", err))
					return
				}
				ctx.SendChain(message.Image("base64://" + binary.BytesToString(img)))
				return
			}
			msg = append(msg, "StdOut:", stdout.String())
		}
		// 输出图片
		img, err = text.RenderToBase64(strings.Join(msg, "\n"), text.BoldFontFile, 1280, 50)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		ctx.SendChain(message.Image("base64://" + binary.BytesToString(img)))
	})
	// 更新zbp
	zero.OnFullMatch("检查更新", zero.OnlyToMe, zero.SuperUserPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		var msg []string
		var img []byte
		var err error
		ctx.SendChain(message.Text("是否备份?"))
		recv, cancel := zero.NewFutureEvent("message", 999, false, zero.RegexRule(`^(是|否)$`), zero.SuperUserPermission).Repeat()
		for {
			select {
			case <-time.After(time.Second * 40): // 40s等待
				ctx.SendChain(message.Text("等待超时,自动备份"))
				err := fileZipTo(file.BOTPATH, file.BOTPATH+time.Now().Format("_2006_01_02_15_04_05")+".zip")
				if err != nil {
					ctx.SendChain(message.Text("[ERROR]:", err))
					return
				}
				msg = append(msg, "已经对旧版zbp压缩备份")
			case e := <-recv:
				nextcmd := e.Event.Message.String() // 获取下一个指令
				switch nextcmd {
				case "是":
					err := fileZipTo(file.BOTPATH, file.BOTPATH+time.Now().Format("_2006_01_02_15_04_05")+".zip")
					if err != nil {
						ctx.SendChain(message.Text("[ERROR]:", err))
						return
					}
					msg = append(msg, "已经对旧版zbp压缩备份")
				default:
					msg = append(msg, "已取消备份")
				}
			}
			if len(msg) != 0 {
				break
			}
		}
		cancel()
		msg = append(msg, "\n\n开始检查更新")
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		for _, command := range commandsOfFetch {
			cmd := exec.Command("git", command...)
			msg = append(msg, "Command:", strings.Join(cmd.Args, " "))
			cmd.Dir = file.BOTPATH
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			err = cmd.Run()
			if err != nil {
				msg = append(msg, "StdErr:", err.Error(), "\n", stderr.String())
				// 输出图片
				img, err = text.RenderToBase64(strings.Join(msg, "\n"), text.BoldFontFile, 1280, 50)
				if err != nil {
					ctx.SendChain(message.Text("[ERROR]:", err))
					return
				}
				ctx.SendChain(message.Image("base64://" + binary.BytesToString(img)))
				return
			}
			msg = append(msg, "StdOut:", stdout.String())
		}
		// 输出图片
		img, err = text.RenderToBase64(strings.Join(msg, "\n"), text.BoldFontFile, 1280, 50)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		ctx.SendChain(message.Image("base64://" + binary.BytesToString(img)))
	})
	// 更新zbp
	zero.OnFullMatch("更新zbp", zero.OnlyToMe, zero.SuperUserPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		var msg []string
		var img []byte
		var err error
		ctx.SendChain(message.Text("是否备份?"))
		recv, cancel := zero.NewFutureEvent("message", 999, false, zero.RegexRule(`^(是|否)$`), zero.SuperUserPermission).Repeat()
		for {
			select {
			case <-time.After(time.Second * 40): // 40s等待
				ctx.SendChain(message.Text("等待超时,自动备份"))
				err := fileZipTo(file.BOTPATH, file.BOTPATH+time.Now().Format("_2006_01_02_15_04_05")+".zip")
				if err != nil {
					ctx.SendChain(message.Text("[ERROR]:", err))
					return
				}
				msg = append(msg, "已经对旧版zbp压缩备份")
			case e := <-recv:
				nextcmd := e.Event.Message.String() // 获取下一个指令
				switch nextcmd {
				case "是":
					err := fileZipTo(file.BOTPATH, file.BOTPATH+time.Now().Format("_2006_01_02_15_04_05")+".zip")
					if err != nil {
						ctx.SendChain(message.Text("[ERROR]:", err))
						return
					}
					msg = append(msg, "已经对旧版zbp压缩备份")
				default:
					msg = append(msg, "已取消备份")
				}
			}
			if len(msg) != 0 {
				break
			}
		}
		cancel()
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		msg = append(msg, "\n\n开始更新go.mod")
		for _, command := range commandsOfZbp {
			cmd := exec.Command("curl", command...)
			msg = append(msg, "Command:", strings.Join(cmd.Args, " "))
			cmd.Dir = file.BOTPATH
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			err = cmd.Run()
			if err != nil {
				msg = append(msg, "StdErr:", err.Error(), "\n", stderr.String())
				// 输出图片
				img, err = text.RenderToBase64(strings.Join(msg, "\n"), text.BoldFontFile, 1280, 50)
				if err != nil {
					ctx.SendChain(message.Text("[ERROR]:", err))
					return
				}
				ctx.SendChain(message.Image("base64://" + binary.BytesToString(img)))
				return
			}
			msg = append(msg, "StdOut:", stdout.String())
		}
		// 输出图片
		img, err = text.RenderToBase64(strings.Join(msg, "\n"), text.BoldFontFile, 1280, 50)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		ctx.SendChain(message.Image("base64://" + binary.BytesToString(img)))
	})
}

// 打包成zip文件
func fileZipTo(src_dir string, zip_file_name string) error {
	// 创建：zip文件
	zipfile, err := os.Create(zip_file_name)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	// 打开：zip文件
	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	// 遍历路径信息
	filepath.Walk(src_dir, func(path string, info os.FileInfo, _ error) error {

		// 如果是源路径，提前进行下一个遍历
		if path == src_dir {
			return nil
		}

		// 获取：文件头信息
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = strings.TrimPrefix(path, src_dir+`/`)

		// 判断：文件是不是文件夹
		if info.IsDir() {
			header.Name += `/`
		} else {
			// 设置：zip的文件压缩算法
			header.Method = zip.Deflate
		}

		// 创建：压缩包头部信息
		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			io.Copy(writer, file)
		}
		return nil
	})
	return nil
}
