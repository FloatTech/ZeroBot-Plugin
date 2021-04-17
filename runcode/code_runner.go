package program

import (
	"encoding/json"
	"fmt"
	zero "github.com/wdvxdr1123/ZeroBot"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func init()  {
	runAllow := true
	runTemplates := map[string]string{
		"Py2": "print 'Hello World!'",
		"Ruby": "puts \"Hello World!\";",
		"PHP": "<?php\n\techo 'Hello World!';\n?>",
		"Go": "package main\n\nimport \"fmt\"\n\nfunc main() {\n   fmt.Println(\"Hello, World!\")\n}",
		"C": "#include <stdio.h>\n\nint main()\n{\n   printf(\"Hello, World! \n\");\n   return 0;\n}",
		"C++": "#include <iostream>\nusing namespace std;\n\nint main()\n{\n   cout << \"Hello World\";\n   return 0;\n}",
		"Java": "public class HelloWorld {\n    public static void main(String []args) {\n       System.out.println(\"Hello World!\");\n    }\n}",
		"Rust": "fn main() {\n    println!(\"Hello World!\");\n}",
		"C#": "using System;\nnamespace HelloWorldApplication\n{\n   class HelloWorld\n   {\n      static void Main(string[] args)\n      {\n         Console.WriteLine(\"Hello World!\");\n      }\n   }\n}",
		"Perl": "print \"Hello, World!\n\";",
		"Python": "print(\"Hello, World!\")",
		"Swift": "var myString = \"Hello, World!\"\nprint(myString)",
		"Lua": "var myString = \"Hello, World!\"\nprint(myString)",
	}
	runTypes := map[string][2]string{
		"Py2": {"0","py"},
		"Ruby": {"1","rb"},
		"PHP": {"3","php"},
		"Go": {"6","go"},
		"C": {"7","c"},
		"C++": {"7","cpp"},
		"Java": {"8","java"},
		"Rust": {"9","rs"},
		"C#": {"10","cs"},
		"Perl": {"14","pl"},
		"Python": {"15","py3"},
		"Swift": {"16","swift"},
		"Lua": {"17","lua"},
	}

	zero.OnCommand("runList").Handle(func(ctx *zero.Ctx) {
		ctx.Send(`[使用说明]
Run 语种<<<
代码块
>>>
[支持语种]
Go || Python || Java || C/C++ || C# || Lua
Rust || PHP || Perl || Ruby || Swift || Py2
PS: 使用(runTemplate 语种)查看该语种模板`)
	})

	zero.OnCommand("runOpen").Handle(func(ctx *zero.Ctx) {
		if ctx.Event.UserID == 213864964{
			runAllow = true
			ctx.Send(fmt.Sprintf(
				"[CQ:at,qq=%d]在线运行代码功能已启用",
				ctx.Event.UserID,
			))
		}
	})

	zero.OnCommand("runClose").Handle(func(ctx *zero.Ctx) {
		if ctx.Event.UserID == 213864964{
			runAllow = false
			ctx.Send(fmt.Sprintf(
				"[CQ:at,qq=%d]在线运行代码功能已禁用",
				ctx.Event.UserID,
			))
		}
	})

	zero.OnRegex("^runTemplate (.+?)$").Handle(func(ctx *zero.Ctx) {
		getType := ctx.State["regex_matched"].([]string)[1]
		if runTemplate,exist:=runTemplates[getType];exist{
			ctx.Send(fmt.Sprintf(
				"[CQ:at,qq=%d]%s template<<<\n%s\n>>>",
				ctx.Event.UserID,
				getType,
				runTemplate,
				))
		}else {
			ctx.Send(fmt.Sprintf(
				"[CQ:at,qq=%d]没有找到%s语言的模板",
				ctx.Event.UserID,
				getType,
			))
		}

	})

	zero.OnRegex("(?is:Run (.+?)<<<(.+?)>>>)").Handle(func(ctx *zero.Ctx) {
		if runAllow==false{
			ctx.Send(fmt.Sprintf(
				"[CQ:at,qq=%d]在线运行代码功能已被禁用",
				ctx.Event.UserID,
				))
			return
		}
		getType := ctx.State["regex_matched"].([]string)[1]
		if runType,exist:=runTypes[getType];exist{
			println("正在尝试执行",getType,"代码块")
			getCode := ctx.State["regex_matched"].([]string)[2]
			getCode = strings.Replace(getCode,"&#91;","[",-1)
			getCode = strings.Replace(getCode,"&#93;","]",-1)

			res := runCode(getCode,runType)
			if res["errors"] == "\n\n"{
				ctx.Send(fmt.Sprintf(
					"[CQ:at,qq=%d]本次%s语言代码执行结果如下:\n%s",
					ctx.Event.UserID,
					getType,
					res["output"][:len(res["output"])-1],
				))
			}else {
				ctx.Send(fmt.Sprintf(
					"[CQ:at,qq=%d]本次%s语言代码执行失败:%s",
					ctx.Event.UserID,
					getType,
					res["errors"],
					))
			}
		}else {
			ctx.Send(fmt.Sprintf(
				"[CQ:at,qq=%d][%s]语言不是受支持的编程语种呢~",
				ctx.Event.UserID,
				getType,
				))
		}
	})
}

func runCode(code string,runType [2]string) map[string]string {
	//对菜鸟api发送数据并返回结果
	result := map[string]string{}
	api := "https://tool.runoob.com/compile2.php"
	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded; charset=UTF-8",
		"Referer": "https://c.runoob.com/",
		"User-Agent": "Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:87.0) Gecko/20100101 Firefox/87.0",
	}
	data := map[string]string{
		"code": code,
		"token": "4381fe197827ec87cbac9552f14ec62a",
		"stdin": "",
		"language": runType[0],
		"fileext": runType[1],
	}
	json.Unmarshal(netPost(api,data,headers),&result)
	return result
}

func netPost(api string,data map[string]string,headers map[string]string) []byte {
	//发送POST请求获取返回数据
	client := &http.Client{
		Timeout: time.Duration(6 * time.Second),
	}

	param := url.Values{}
	for key,value := range data{
		param.Set(key,value)
	}

	request,_ := http.NewRequest("POST",api,strings.NewReader(param.Encode()))
	for key,value := range headers{
		request.Header.Add(key,value)
	}
	res,_ := client.Do(request)
	defer res.Body.Close()
	result,_ := ioutil.ReadAll(res.Body)
	return result
}
