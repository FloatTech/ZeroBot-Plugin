package ygotools

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fogleman/gg"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	control "github.com/FloatTech/zbputils/control"

	//ygo
	"unicode/utf8"

	"github.com/FloatTech/zbputils/file"
	"github.com/FloatTech/zbputils/img"
	"github.com/FloatTech/zbputils/img/text"
	"github.com/FloatTech/zbputils/img/writer"
	"github.com/tealeg/xlsx"
	"github.com/tidwall/gjson"
)

// 用户数据信息
type userdata struct {
	UID          int64     //0
	userName     string    //1
	Count        int       //2
	UpdatedAt    time.Time //3
	obtainpoints int       //4
	lostpoints   int       //5
}

const (
	backgroundURL = "https://iw233.cn/API/pc.php?type=json"
	referer       = "https://iw233.cn/main.html"
	ua            = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36"
	signinMax     = 1
	//积分基数
	initSCORE = 100
)

func init() {
	engine := control.Register("ygotools", &control.Options{
		DisableOnDefault:  false,
		PrivateDataFolder: "ygotools",
	})
	cachePath := engine.DataFolder() + "cache/"

	go func() {
		os.RemoveAll(cachePath)
		err := os.MkdirAll(cachePath, 0755)
		if err != nil {
			panic(err)
		}
		_, err = file.GetLazyData(text.BoldFontFile, false, true)
		if err != nil {
			panic(err)
		}
		_, err = file.GetLazyData(text.FontFile, false, true)
		if err != nil {
			panic(err)
		}
	}()

	engine.OnFullMatch("签到", zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		if ctx.Event.GroupID != 979031435 {
			return
		}
		uid := ctx.Event.UserID
		xfile, err := xlsx.FileToSlice(engine.DataFolder() + "积分.xlsx")
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		//遍历查询是否注册了
		user := userdata{UID: uid}
		rowIndexint := 0
		for rowIndex := range xfile[0] {
			if xfile[0][rowIndex][0] == strconv.FormatInt(int64(uid), 10) {
				rowIndexint = rowIndex
				user.userName = xfile[0][rowIndex][1]
				user.Count, _ = strconv.Atoi(xfile[0][rowIndex][2])
				user.UpdatedAt, _ = time.Parse("20060102", xfile[0][rowIndex][3])
				user.obtainpoints, _ = strconv.Atoi(xfile[0][rowIndex][4])
				user.lostpoints, _ = strconv.Atoi(xfile[0][rowIndex][5])
			}
		}
		if user.userName == "" {
			ctx.SendChain(message.Text("决斗者未注册！\n请输入“登记决斗者 xxx”进行登记(xxx为决斗者昵称)。"))
			return
		}
		now := time.Now()
		today := now.Format("20060102")
		if user.UpdatedAt.Format("20060102") != today {
			user.Count = 0
		}

		drawedFile := cachePath + strconv.FormatInt(uid, 10) + today + "signin.png"
		if user.Count >= 1 && user.UpdatedAt.Format("20060102") == today {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("今天你已经签到过了！"))
			if file.IsExist(drawedFile) {
				ctx.SendChain(message.Image("file:///" + file.BOTPATH + "/" + drawedFile))
			}
			return
		}

		picFile := cachePath + strconv.FormatInt(uid, 10) + today + ".png"
		initPic(picFile)

		add := 1
		user.Count += add
		user.obtainpoints += add
		scoreresult := (initSCORE + user.obtainpoints - user.lostpoints)

		elsxfile, err := xlsx.OpenFile(engine.DataFolder() + "积分.xlsx")
		if err == nil {
			Sheet := elsxfile.Sheets[0]
			row := Sheet.Row(rowIndexint)
			cell := row.Cells
			cell[2].Value = strconv.Itoa(user.Count)
			cell[3].Value = today
			cell[4].Value = strconv.Itoa(user.obtainpoints)

		} else {
			ctx.SendChain(message.Text("error:", err))
			return
		}
		err = elsxfile.Save(engine.DataFolder() + "积分.xlsx")
		if err != nil {
			ctx.SendChain(message.Text("error:", err))
			return
		}

		back, err := gg.LoadImage(picFile)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}

		// 避免图片过大，最大 1280*720
		back = img.Limit(back, 1280, 720)

		canvas := gg.NewContext(back.Bounds().Size().X, int(float64(back.Bounds().Size().Y)*1.6))
		canvas.SetRGB(1, 1, 1)
		canvas.Clear()
		canvas.DrawImage(back, 0, 0)

		monthWord := now.Format("01/02")
		if err = canvas.LoadFontFace(text.BoldFontFile, float64(back.Bounds().Size().X)*0.07); err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		canvas.SetRGB(0, 0, 0)
		canvas.DrawString(user.userName, float64(back.Bounds().Size().X)*0.02, float64(back.Bounds().Size().Y)*1.13)
		canvas.DrawString(monthWord, float64(back.Bounds().Size().X)*0.7, float64(back.Bounds().Size().Y)*1.13)
		if err = canvas.LoadFontFace(text.BoldFontFile, float64(back.Bounds().Size().X)*0.1); err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		canvas.DrawString("今天你", float64(back.Bounds().Size().X)*0.02, float64(back.Bounds().Size().Y)*1.3)
		canvas.DrawString("决斗了吗?", float64(back.Bounds().Size().X)*0.02, float64(back.Bounds().Size().Y)*1.5)

		canvas.DrawString(fmt.Sprintf(" 积分+%d", add), float64(back.Bounds().Size().X)*0.6, float64(back.Bounds().Size().Y)*1.3)
		if err = canvas.LoadFontFace(text.FontFile, float64(back.Bounds().Size().X)*0.04); err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		canvas.DrawString("当前积分:"+strconv.FormatInt(int64(scoreresult), 10), float64(back.Bounds().Size().X)*0.65, float64(back.Bounds().Size().Y)*1.4)
		canvas.DrawString("折算金额:"+strconv.FormatInt(int64(scoreresult/100), 10)+"元", float64(back.Bounds().Size().X)*0.65, float64(back.Bounds().Size().Y)*1.5)

		f, err := os.Create(drawedFile)
		if err != nil {
			fmt.Print("[score]", err)
			data, cl := writer.ToBytes(canvas.Image())
			if err != nil {
				panic(err)
			}
			ctx.SendChain(message.ImageBytes(data))
			cl()
			return
		}
		_, err = writer.WriteTo(canvas.Image(), f)
		_ = f.Close()
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		ctx.SendChain(message.Image("file:///" + file.BOTPATH + "/" + drawedFile))
	})
	engine.OnFullMatch("/积分", zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		if ctx.Event.GroupID != 979031435 {
			return
		}
		uid := ctx.Event.UserID
		elsxfile, err := xlsx.FileToSlice(engine.DataFolder() + "积分.xlsx")
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		//遍历查询是否注册了
		user := userdata{UID: uid}
		for rowIndex := range elsxfile[0] {
			if elsxfile[0][rowIndex][0] == strconv.FormatInt(int64(uid), 10) {
				user.userName = elsxfile[0][rowIndex][1]
				user.obtainpoints, _ = strconv.Atoi(elsxfile[0][rowIndex][4])
				user.lostpoints, _ = strconv.Atoi(elsxfile[0][rowIndex][5])
			}
		}
		if user.userName == "" {
			ctx.SendChain(message.Text("决斗者未注册！\n请输入“登记决斗者 xxx”进行登记(xxx为决斗者昵称)。"))
			return
		}
		now := time.Now()
		today := now.Format("20060102")
		picFile := cachePath + strconv.FormatInt(uid, 10) + today + ".png"
		initPic(picFile)
		back, err := gg.LoadImage(picFile)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}

		// 避免图片过大，最大 1280*720
		back = img.Limit(back, 1280, 720)

		canvas := gg.NewContext(back.Bounds().Size().X, int(float64(back.Bounds().Size().Y)*1.25))
		canvas.SetRGB(1, 1, 1)
		canvas.Clear()
		canvas.DrawImage(back, 0, 0)

		if err = canvas.LoadFontFace(text.BoldFontFile, float64(back.Bounds().Size().X)*0.07); err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		canvas.SetRGB(0, 0, 0)
		canvas.DrawString(user.userName, float64(back.Bounds().Size().X)*0.07, float64(back.Bounds().Size().Y)*1.13)

		scoreresult := (initSCORE + user.obtainpoints - user.lostpoints)
		if err = canvas.LoadFontFace(text.FontFile, float64(back.Bounds().Size().X)*0.04); err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		canvas.DrawString("当前积分:"+strconv.FormatInt(int64(scoreresult), 10), float64(back.Bounds().Size().X)*0.37, float64(back.Bounds().Size().Y)*1.1)
		canvas.DrawString("折算金额:"+strconv.FormatInt(int64(scoreresult/100), 10), float64(back.Bounds().Size().X)*0.67, float64(back.Bounds().Size().Y)*1.1)
		canvas.DrawString("初始积分:"+strconv.FormatInt(int64(initSCORE), 10), float64(back.Bounds().Size().X)*0.07, float64(back.Bounds().Size().Y)*1.23)
		canvas.DrawString("获得积分:"+strconv.FormatInt(int64(user.obtainpoints), 10), float64(back.Bounds().Size().X)*0.37, float64(back.Bounds().Size().Y)*1.23)
		canvas.DrawString("已用积分:"+strconv.FormatInt(int64(user.lostpoints), 10), float64(back.Bounds().Size().X)*0.67, float64(back.Bounds().Size().Y)*1.23)

		data, cl := writer.ToBytes(canvas.Image())
		ctx.SendChain(message.ImageBytes(data))
		cl()
	})

	zero.OnRegex(`^/记录 (.+)$`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).SetPriority(20).Handle(func(ctx *zero.Ctx) {
		//if ctx.Event.GroupID != 979031435 {
		//	return
		//}
		List := ctx.State["regex_matched"].([]string)[1]
		arr := strings.Fields(List)
		if len(arr) < 2 || len(arr) > 3 {
			ctx.SendChain(message.Text("指令存在错误，重新输入"))
			return
		}
		adduser := arr[0]
		addscore := arr[1]
		devuser := ""
		if len(arr) == 3 {
			devuser = arr[2]
		}
		xfile, err := xlsx.FileToSlice(engine.DataFolder() + "积分.xlsx")
		if err != nil {
			fmt.Println(err)
			return
		}
		var add_ID = 0
		var dev_ID = 0
		scoreint, _ := strconv.Atoi(addscore)

		for rowIndex := range xfile[0] {
			if xfile[0][rowIndex][1] == adduser {
				add_ID = rowIndex
			}
		}
		if devuser != "" {
			for rowIndex := range xfile[0] {
				if xfile[0][rowIndex][1] == devuser {
					dev_ID = rowIndex
				}
			}
		}

		file, err := xlsx.OpenFile(engine.DataFolder() + "积分.xlsx")
		if err != nil {
			panic(err)
		}
		Sheet := file.Sheets[0]
		if add_ID == 0 {
			row := Sheet.AddRow()
			cell := row.AddCell() //ID
			cell = row.AddCell()  //Name
			cell.Value = adduser
			cell = row.AddCell() //count
			cell.Value = "0"
			cell = row.AddCell() //date
			cell.Value = "0"
			if scoreint > 0 {
				cell = row.AddCell() //add
				cell.Value = strconv.Itoa(scoreint)
				cell = row.AddCell() //lost
				cell.Value = "0"
			} else {
				cell = row.AddCell() //add
				cell.Value = "0"
				cell = row.AddCell() //lost
				cell.Value = strconv.Itoa(-scoreint)
			}
		} else {
			if scoreint > 0 {
				jifen_data := Sheet.Rows[add_ID].Cells[4].Value
				jifen_int, _ := strconv.Atoi(jifen_data)
				last_data := jifen_int + scoreint
				Sheet.Rows[add_ID].Cells[4].Value = strconv.Itoa(last_data)
			} else {
				jifen_data := Sheet.Rows[add_ID].Cells[5].Value
				jifen_int, _ := strconv.Atoi(jifen_data)
				last_data := jifen_int - scoreint
				Sheet.Rows[add_ID].Cells[5].Value = strconv.Itoa(last_data)
			}
		}
		if devuser != "" {
			if dev_ID == 0 {
				row := Sheet.AddRow()
				cell := row.AddCell() //ID
				cell = row.AddCell()  //Name
				cell.Value = devuser
				cell = row.AddCell() //count
				cell.Value = "0"
				cell = row.AddCell() //date
				cell.Value = "0"
				cell = row.AddCell() //add
				cell.Value = "0"
				cell = row.AddCell() //lost
				cell.Value = strconv.Itoa(scoreint)
			} else if dev_ID > 0 {
				jifen_data := Sheet.Rows[dev_ID].Cells[5].Value
				jifen_int, _ := strconv.Atoi(jifen_data)
				last_data := jifen_int + scoreint
				Sheet.Rows[dev_ID].Cells[5].Value = strconv.Itoa(last_data)
			}
		}
		err = file.Save(engine.DataFolder() + "积分.xlsx")
		if err == nil {
			ctx.SendChain(message.Text("登记完成"))
		} else {
			ctx.SendChain(message.Text("error:", err))
			return
		}
	})

	engine.OnPrefix("登记决斗者", zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		if ctx.Event.GroupID != 979031435 {
			return
		}
		checkinName := ctx.State["args"].(string)
		if utf8.RuneCountInString(checkinName) > 8 {
			ctx.SendChain(message.Text("昵称字段仅允许8个字符以下"))
			return
		}
		uid := ctx.Event.UserID
		file, err := xlsx.FileToSlice(engine.DataFolder() + "积分.xlsx")
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		rowedIndex := 0
		hadsorce := false
		//遍历查询是否注册了
		for rowIndex := range file[0] {
			if file[0][rowIndex][0] == strconv.FormatInt(int64(uid), 10) {
				ctx.SendChain(message.Text("你已注册！\n如需更改请联系管理员"))
				return
			} else if file[0][rowIndex][1] == checkinName {
				rowedIndex = rowIndex
				hadsorce = true
			}
		}
		elsxfile, err := xlsx.OpenFile(engine.DataFolder() + "积分.xlsx")
		if err != nil {
			panic(err)
		}
		Sheet := elsxfile.Sheets[0]
		if hadsorce == false {
			row := Sheet.AddRow()
			cell := row.AddCell()
			cell.Value = strconv.Itoa(int(uid))
			cell = row.AddCell()
			cell.Value = checkinName
			cell = row.AddCell()
			cell.Value = "0"
			cell = row.AddCell()
			cell.Value = "0"
			cell = row.AddCell()
			cell.Value = "0"
			cell = row.AddCell()
			cell.Value = "0"
		} else {
			row := Sheet.Row(rowedIndex)
			cell := row.Cells
			cell[0].Value = strconv.FormatInt(int64(uid), 10)
		}
		err = elsxfile.Save(engine.DataFolder() + "积分.xlsx")
		if err == nil {
			ctx.SendChain(message.Text("注册成功！"))
		} else {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
	})
	engine.OnPrefix("获得签到背景", zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		if ctx.Event.GroupID != 979031435 {
			return
		}
		param := ctx.State["args"].(string)
		var uidStr string
		if len(ctx.Event.Message) > 1 && ctx.Event.Message[1].Type == "at" {
			uidStr = ctx.Event.Message[1].Data["qq"]
		} else if param == "" {
			uidStr = strconv.FormatInt(ctx.Event.UserID, 10)
		}
		picFile := cachePath + uidStr + time.Now().Format("20060102") + ".png"
		if file.IsNotExist(picFile) {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("请先签到！"))
			return
		}
		ctx.SendChain(message.Image("file:///" + file.BOTPATH + "/" + picFile))
	})

}

// ReqWith 使用自定义请求头获取数据
func ReqWith(url string, method string, referer string, ua string) (data []byte, err error) {
	client := &http.Client{}
	// 提交请求
	var request *http.Request
	request, err = http.NewRequest(method, url, nil)
	if err == nil {
		// 增加header选项
		request.Header.Add("Referer", referer)
		request.Header.Add("User-Agent", ua)
		var response *http.Response
		response, err = client.Do(request)
		if err == nil {
			data, err = io.ReadAll(response.Body)
			response.Body.Close()
		}
	}
	return
}
func initPic(picFile string) {
	if file.IsNotExist(picFile) {
		data, err := ReqWith(backgroundURL, "GET", referer, ua)
		if err != nil {
			fmt.Print("[score]", err)
		}
		picURL := gjson.Get(string(data), "pic").String()
		data, err = ReqWith(picURL, "GET", "", ua)
		if err != nil {
			fmt.Print("[score]", err)
		}
		err = os.WriteFile(picFile, data, 0666)
		if err != nil {
			fmt.Print("[score]", err)
		}
	}
}
