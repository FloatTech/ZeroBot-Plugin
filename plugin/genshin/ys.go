package genshin

import (
	"archive/zip"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/control/order"
	"github.com/FloatTech/zbputils/file"
	"github.com/golang/freetype"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	//genzip                        = "https://gitcode.net/qq_33747476/Images/-/raw/master/zero/Genshin.zip"
	genzip                          = "https://raw.githubusercontent.com/FloatTech/zbpdata/main/Genshin/Genshin.zip" //素材包
	Three, four, four2, five, five2 = []string{}, []string{}, []string{}, []string{}, []string{}                     //三 , 四, 五星的名字

	DP, gen, ZipN, Spath       = "./data/Genshin/", "_genshin.jpg", DP + "Genshin.zip", DP + "gacha/"    //路径
	five_bg, four_bg, three_bg = DP + "five_bg.jpg", DP + "four_bg.jpg", DP + "three_bg.jpg"             //背景图片名
	StarN3, StarN4, StarN5     = Spath + "ThreeStar.png", Spath + "FourStar.png", Spath + "FiveStar.png" //星级图标

	Toatl, IMGN, isPath, Lock      = 0, 0, false, false //累计抽奖次数, 连抽次数, 卡池总数, 是否遍历过路径
	DefaultMode, FiveMode, iPrefix = true, false, 0     //抽卡模式,文件前缀

	engine = control.Register("genshin", order.AcquirePrio(), &control.Options{
		DisableOnDefault: false,
		Help:             "genshin\n@bot\n- 十连\n- 十连抽\n- 切换卡池\n- 刷新卡池",
	})
)

func init() {

	engine.OnFullMatchGroup([]string{"刷新卡池"}, zero.OnlyToMe).SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			if IsExists(DP) {
				IMGN = 0
				LoadFolder()
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("刷新完成!\n数量:", IMGN))
				IMGN = 0
			} else {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("没有素材哦~\n请发送 十连 \n下载素材.."))
			}
		})

	engine.OnFullMatchGroup([]string{"切换卡池"}, zero.OnlyToMe).SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			if !DefaultMode {
				DefaultMode, FiveMode = true, false
				ctx.SendChain(message.Text("切换到普通卡池~"))
			} else {
				DefaultMode, FiveMode = false, true
				ctx.SendChain(message.Text("切换到五星卡池~"))
			}
		})
	//@bot
	engine.OnFullMatchGroup([]string{"十连", "十连抽", "来发十连", "来次十连", "来份十连"}, zero.OnlyToMe).SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			times := time.Now().Unix()
			iPrefix++
			if !isPath {
				if Lock {
					return
				}
				if !IsExists(DP) && !Lock {
					Lock = true
					ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("素材未下载...\n是否需要下载素材\n下载素材|取消下载"))
					ch := make(chan int, 1)
					defer close(ch)
					rule := func(ctx *zero.Ctx) bool {
						for _, v := range ctx.Event.Message {
							if v.Type == "text" {
								text := strings.ReplaceAll(v.Data["text"], " ", "")
								if text == "下载素材" {
									return true
								}
								if text == "取消下载" {
									ch <- 1
									return false
								}
							}
						}
						return false
					}
					next := zero.NewFutureEvent("message", 999, false, zero.CheckUser(ctx.Event.UserID), rule)
					recv, cancel := next.Repeat()
					select {
					case <-time.After(time.Second * 60):
						Lock = false
						cancel()
						return
					case <-recv:
						cancel()
						ctx.SendChain(message.Text("正在下载中~"))
						os.Mkdir(DP, 0755)
						err := file.DownloadTo(genzip, ZipN, true)
						Lock = false
						if err != nil {
							ctx.SendChain(message.Text("下载出错了~", err))
							return
						}
					case <-ch:
						cancel()
						ctx.SendChain(message.Text("取消了呢~"))
						Lock = false
						close(ch)
						return
					}
					err := unzip(ZipN, "./data/")
					if err != nil {
						ctx.SendChain(message.Text("解压出错了~", err))
					}
					ctx.SendChain(message.Text("下载完成~"))
					os.Remove(ZipN)
					os.Chmod(DP, 0755)
					return
				} else {
					LoadFolder()
				}
			}
			Str := strconv.Itoa(iPrefix)
			defer os.Remove(DP + Str + gen) //删除图片
			Add(Str)
			ctx.SendChain(message.At(ctx.Event.UserID), message.Text("十连成功~"),
				message.Image("file:///"+file.Pwd()+"/data/Genshin/"+Str+gen))
			fmt.Printf("OK %v 秒\n", time.Now().Unix()-times)
		})
}

func Add(sf string) {
	var (
		fourName, fiveName            = []string{}, []string{}             //抽到 四, 五星角色的名字
		ThreeArms, fourArms, fiveArms = []string{}, []string{}, []string{} //抽到 三 , 四, 五星武器的名字

		fourN, fiveN, Nums, bgs = 0, 0, 10, []string{}   //抽到 四, 五星角色的数量, 连抽次数, 背景图片名
		ThreeN2, fourN2, fiveN2 = 0, 0, 0                //抽到 三 , 四, 五星武器的数量
		Hero, StarName          = []string{}, []string{} //角色武器名, 储存星级图标

		Cicon, He_N = []string{}, []string{} //元素图标
	)

	rand.Seed(time.Now().UnixNano())
	if Toatl == 9 { //累计9次十连加入一个五星
		switch a := rand.Intn(2); a {
		case 0:
			fiveN++
			f := rand.Intn(len(five))
			fiveName = append(fiveName, five[f])
		case 1:
			fiveN2++
			f := rand.Intn(len(five2))
			fiveArms = append(fiveArms, five2[f])
		}
		Toatl, Nums = 0, 9
	}

	if DefaultMode { //默认模式
		for i := 0; i < Nums; i++ {
			a := rand.Intn(1000)
			if a >= 0 && a <= 750 { //抽卡几率 三星75% 四星18% 五星7%
				ThreeN2++
				ta := rand.Intn(len(Three)) //名字总个数
				ThreeArms = append(ThreeArms, Three[ta])
			} else if a > 750 && a <= 840 {
				fourN++
				tb := rand.Intn(len(four))
				fourName = append(fourName, four[tb]) //随机角色
			} else if a > 840 && a <= 930 {
				fourN2++
				tb := rand.Intn(len(four2))
				fourArms = append(fourArms, four2[tb]) //随机武器
			} else if a > 930 && a <= 965 {
				fiveN++
				tc := rand.Intn(len(five))
				fiveName = append(fiveName, five[tc])
			} else if a > 965 {
				fiveN2++
				tc := rand.Intn(len(five2))
				fiveArms = append(fiveArms, five2[tc])
			}
		}
		if fourN+fourN2 == 0 && ThreeN2 > 0 { //没有四星时自动加入
			ThreeN2--
			ThreeArms = ThreeArms[:len(ThreeArms)-1]
			switch a := rand.Intn(2); a {
			case 0:
				fourN++
				f := rand.Intn(len(four))
				fourName = append(fourName, four[f]) //随机角色
			case 1:
				fourN2++
				f := rand.Intn(len(four2))
				fourArms = append(fourArms, four2[f]) //随机武器
			}
		}
		Toatl++ //次数+1
	}

	if FiveMode { //5星模式
		for i := 0; i < Nums; i++ {
			a := rand.Intn(100)
			if a >= 0 && a <= 50 {
				fiveN++
				tc := rand.Intn(len(five))
				fiveName = append(fiveName, five[tc]) //随机角色
			} else if a > 50 {
				fiveN2++
				tc := rand.Intn(len(five2))
				fiveArms = append(fiveArms, five2[tc]) //随机武器
			}
		}
	}
	//正则取出图标名
	Addicon := func(Prefix string) string {

		regC, _ := regexp.Compile(`\/[a-z]+_`)
		regC2, _ := regexp.Compile(`\/`)

		if runtime.GOOS == "windows" {
			regC, _ = regexp.Compile(`\\[a-z]+_`)
			regC2, _ = regexp.Compile(`\\`)
		}
		Pr := regC.FindAllString(Prefix, 1) //取出字符
		str := strings.Join(Pr, "")         //连接字符
		Sp := strings.Split(str, "_")
		Str2 := strings.Join(Sp, "")
		Str3 := regC2.ReplaceAllString(Str2, "")

		switch Str3 {
		case "anemo":
			return DP + "anemo.png"
		case "cryo":
			return DP + "cryo.png"
		case "dendro":
			return DP + "dendro.png"
		case "electro":
			return DP + "electro.png"
		case "geo":
			return DP + "geo.png"
		case "hydro":
			return DP + "hydro.png"
		case "pyro":
			return DP + "pyro.png"
		case "bow":
			return DP + "bow.png"
		case "catalyst":
			return DP + "catalyst.png"
		case "claymore":
			return DP + "claymore.png"
		case "polearm":
			return DP + "polearm.png"
		default:
			return DP + "sword.png"
		}
	}

	He := func(StarNum int, id int, Star string, bg string) {
		for i := 0; i < StarNum; i++ {
			switch id {
			case 1:
				He_N = ThreeArms
			case 2:
				He_N = fourArms
			case 3:
				He_N = fourName
			case 4:
				He_N = fiveArms
			case 5:
				He_N = fiveName
			}
			bgs = append(bgs, bg) //加入颜色背景
			Hero = append(Hero, He_N[i])
			StarName = append(StarName, Star)       //加入星级图标
			Cicon = append(Cicon, Addicon(He_N[i])) //加入元素图标
		}
	}

	if fiveN > 0 { //按顺序加入
		He(fiveN, 5, StarN5, five_bg) //五星角色
	}
	if fourN > 0 {
		He(fourN, 3, StarN4, four_bg) //四星角色
	}
	if fiveN2 > 0 {
		He(fiveN2, 4, StarN5, five_bg) //五星武器
	}
	if fourN2 > 0 {
		He(fourN2, 2, StarN4, four_bg) //四星武器
	}
	if ThreeN2 > 0 {
		He(ThreeN2, 1, StarN3, three_bg) //三星武器
	}
	//	fmt.Printf("三星数量 = %v 四星数量 = %v 五星数量 = %v\n", ThreeN2, fourN+fourN2, fiveN+fiveN2)
	//	fmt.Printf("三星 = %v \n四星 = %v%v \n五星 = %v%v\n", ThreeArms, fourName, fourArms, fiveName, fiveArms)

	var (
		inp              = DP + "bg0.jpg" //背景导入图片路径
		c1, c2, c3 uint8 = 50, 50, 50     //背景颜色
		opt        jpeg.Options
		//字体rgb 205, 205, 205
	)

	rectangle := image.Rect(0, 0, 1920, 1080) // 图片宽度, 图片高度
	rgba := image.NewRGBA(rectangle)
	draw.Draw(rgba, rgba.Bounds(), image.NewUniform(color.RGBA{c1, c2, c3, 255}), image.Point{}, draw.Over)
	context := freetype.NewContext() // 创建一个新的上下文
	context.SetDPI(72)               //每英寸 dpi
	context.SetClip(rgba.Bounds())
	context.SetDst(rgba)

	img00, _ := os.Open(inp) // 打开背景图片
	defer img00.Close()
	img0, _ := jpeg.Decode(img00) //读取一个本地图像
	offset := image.Pt(0, 0)      //图片在背景上的位置
	draw.Draw(rgba, img0.Bounds().Add(offset), img0, image.Point{}, draw.Over)

	OpenW1, OpenH1 := 230, 0
	for i := 0; i < len(Hero); i++ {
		if i > 0 {
			OpenW1 += 146 //图片宽度
		}
		imgs, _ := os.Open(bgs[i]) //取出背景图片
		defer imgs.Close()
		img, _ := jpeg.Decode(imgs)
		offset := image.Pt(OpenW1, OpenH1)
		draw.Draw(rgba, img.Bounds().Add(offset), img, image.Point{}, draw.Over)

		imgs1, _ := os.Open(Hero[i]) //取出图片名
		defer imgs1.Close()
		img1, _ := png.Decode(imgs1)
		offset1 := image.Pt(OpenW1, OpenH1)
		draw.Draw(rgba, img1.Bounds().Add(offset1), img1, image.Point{}, draw.Over)

		imgs2, _ := os.Open(StarName[i]) //取出星级图标
		defer imgs2.Close()
		img2, _ := png.Decode(imgs2)
		offset2 := image.Pt(OpenW1, OpenH1)
		draw.Draw(rgba, img2.Bounds().Add(offset2), img2, image.Point{}, draw.Over)

		imgs3, _ := os.Open(Cicon[i]) //取出类型图标
		defer imgs3.Close()
		img3, _ := png.Decode(imgs3)
		offset3 := image.Pt(OpenW1, OpenH1)
		draw.Draw(rgba, img3.Bounds().Add(offset3), img3, image.Point{}, draw.Over)
	}
	imgs4, _ := os.Open(DP + "Reply.png") //"分享" 图标
	defer imgs4.Close()
	img4, _ := png.Decode(imgs4)
	offset4 := image.Pt(1270, 945) //宽, 高
	draw.Draw(rgba, img4.Bounds().Add(offset4), img4, image.Point{}, draw.Over)

	file, err := os.Create(DP + sf + gen) //输出图片

	if err != nil {
		fmt.Printf("输出图片错误: %v\n", err)
		return
	}
	defer file.Close()
	//png.Encode(file, rgba)  //输出png
	opt.Quality = 100
	jpeg.Encode(file, rgba, &opt) //输出jpg
}

//遍历文件夹
func LoadFolder() {
	isPath = true
	Three, four = []string{}, []string{} //重置变量
	four2, five, five2 = []string{}, []string{}, []string{}
	ReadPath(DP+"Three", "Three")
	ReadPath(DP+"four", "four")
	ReadPath(DP+"four2", "four2")
	ReadPath(DP+"five", "five")
	ReadPath(DP+"five2", "five2")
}

func ReadPath(lpath string, star string) {
	err := filepath.Walk(lpath, func(filename string, fi os.FileInfo, err error) error {
		if fi.IsDir() { // 忽略目录
			return nil
		}
		f := filename
		if star == "five" {
			five = append(five, "./"+f)
		} else if star == "five2" {
			five2 = append(five2, "./"+f)
		} else if star == "four" {
			four = append(four, "./"+f)
		} else if star == "four2" {
			four2 = append(four2, "./"+f)
		} else {
			Three = append(Three, "./"+f)
		}
		IMGN++
		return nil
	})
	if err != nil {
		fmt.Printf("读取文件错误%v\n", err)
	}
}

func IsExists(path string) bool { //判断文件是否存在
	_, err := os.Stat(path)
	if err != nil && !os.IsExist(err) {
		return false
	}
	return true
}

//解压缩
func unzip(zipFile string, destDir string) error {
	zipReader, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer zipReader.Close()
	for _, f := range zipReader.File {
		fpath := filepath.Join(destDir, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
		} else {
			if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return err
			}
			inFile, err := f.Open() //读取压缩文件
			if err != nil {
				return err
			}
			defer inFile.Close()
			outFile, err := os.OpenFile(fpath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode()) //创建的新文件
			if err != nil {
				return err
			}
			defer outFile.Close()
			_, _ = io.Copy(outFile, inFile)
		}
	}
	return err
}
