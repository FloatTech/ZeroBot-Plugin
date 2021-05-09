<div align="center">
  <img src=".github/yaya.jpg" width = "150" height = "150" alt="OneBot-YaYa"><br>
  <h2>ZeroBot-Plugin</h2>
  ZeroBot-Plugin 是 ZeroBot 的 实用插件合集<br><br>

[![YAYA](https://img.shields.io/badge/OneBot-YaYa-green.svg?style=social&logo=appveyor)](https://github.com/Yiwen-Chan/OneBot-YaYa)
[![GOCQ](https://img.shields.io/badge/OneBot-MiraiGo-green.svg?style=social&logo=appveyor)](https://github.com/Mrs4s/go-cqhttp)
[![OICQ](https://img.shields.io/badge/OneBot-OICQ-green.svg?style=social&logo=appveyor)](https://github.com/takayama-lily/node-onebot)
[![MIRAI](https://img.shields.io/badge/OneBot-Mirai-green.svg?style=social&logo=appveyor)](https://github.com/yyuueexxiinngg/onebot-kotlin)

[![Go Report Card](https://goreportcard.com/badge/github.com/Yiwen-Chan/ZeroBot-Plugin?style=flat-square&logo=go)](https://goreportcard.com/report/github.com/github.com/Yiwen-Chan/ZeroBot-Plugin)
[![Badge](https://img.shields.io/badge/onebot-v11-black?logo=data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAHAAAABwCAMAAADxPgR5AAAAGXRFWHRTb2Z0d2FyZQBBZG9iZSBJbWFnZVJlYWR5ccllPAAAAAxQTFRF////29vbr6+vAAAAk1hCcwAAAAR0Uk5T////AEAqqfQAAAKcSURBVHja7NrbctswDATQXfD//zlpO7FlmwAWIOnOtNaTM5JwDMa8E+PNFz7g3waJ24fviyDPgfhz8fHP39cBcBL9KoJbQUxjA2iYqHL3FAnvzhL4GtVNUcoSZe6eSHizBcK5LL7dBr2AUZlev1ARRHCljzRALIEog6H3U6bCIyqIZdAT0eBuJYaGiJaHSjmkYIZd+qSGWAQnIaz2OArVnX6vrItQvbhZJtVGB5qX9wKqCMkb9W7aexfCO/rwQRBzsDIsYx4AOz0nhAtWu7bqkEQBO0Pr+Ftjt5fFCUEbm0Sbgdu8WSgJ5NgH2iu46R/o1UcBXJsFusWF/QUaz3RwJMEgngfaGGdSxJkE/Yg4lOBryBiMwvAhZrVMUUvwqU7F05b5WLaUIN4M4hRocQQRnEedgsn7TZB3UCpRrIJwQfqvGwsg18EnI2uSVNC8t+0QmMXogvbPg/xk+Mnw/6kW/rraUlvqgmFreAA09xW5t0AFlHrQZ3CsgvZm0FbHNKyBmheBKIF2cCA8A600aHPmFtRB1XvMsJAiza7LpPog0UJwccKdzw8rdf8MyN2ePYF896LC5hTzdZqxb6VNXInaupARLDNBWgI8spq4T0Qb5H4vWfPmHo8OyB1ito+AysNNz0oglj1U955sjUN9d41LnrX2D/u7eRwxyOaOpfyevCWbTgDEoilsOnu7zsKhjRCsnD/QzhdkYLBLXjiK4f3UWmcx2M7PO21CKVTH84638NTplt6JIQH0ZwCNuiWAfvuLhdrcOYPVO9eW3A67l7hZtgaY9GZo9AFc6cryjoeFBIWeU+npnk/nLE0OxCHL1eQsc1IciehjpJv5mqCsjeopaH6r15/MrxNnVhu7tmcslay2gO2Z1QfcfX0JMACG41/u0RrI9QAAAABJRU5ErkJggg==)](https://github.com/howmanybots/onebot)
[![Badge](https://img.shields.io/badge/zerobot-v1.1.2-black?style=flat-square&logo=go)](https://github.com/wdvxdr1123/ZeroBot)
[![License](https://img.shields.io/github/license/Yiwen-Chan/OneBot-YaYa.svg?style=flat-square&logo=gnu)](https://raw.githubusercontent.com/Yiwen-Chan/ZeroBot-Plugin/master/LICENSE)
[![qq group](https://img.shields.io/badge/group-1048452984-red?style=flat-square&logo=tencent-qq)](https://jq.qq.com/?_wv=1027&k=QMb7x1mM)

</div>


### 功能
- 聊天 `import _ "github.com/Yiwen-Chan/ZeroBot-Plugin/chat"`
    - [x] [BOT名字]
    - [x] [戳一戳BOT]
    - [x] 空调开
    - [x] 空调关
    - [x] 群温度
    - [x] 设置温度[正整数]
- 椛椛 `import _ "github.com/Yiwen-Chan/ZeroBot-Plugin/huahua"`
    - [x] 具体指令看代码
- ATRI `import _ "github.com/Yiwen-Chan/ZeroBot-Plugin/atri"`
    - [x] 具体指令看代码
    - 注：本插件基于 [ATRI](https://github.com/Kyomotoi/ATRI) ，为 Golang 移植版
- 群管 `import _ "github.com/Yiwen-Chan/ZeroBot-Plugin/manager"`
    - [x] 禁言[@xxx][分钟]
    - [x] 解除禁言[@xxx]
    - [x] 我要自闭 [分钟]
    - [x] 开启全员禁言
    - [x] 解除全员禁言
    - [x] 升为管理[@xxx]
    - [x] 取消管理[@xxx]
    - [x] 修改名片[@xxx][xxx]
    - [x] 修改头衔[@xxx][xxx]
    - [x] 申请头衔[xxx]
    - [x] 踢出群聊[@xxx]
    - [x] 退出群聊[群号]
    - [x] *入群欢迎
    - [x] *退群通知
    - [x] 在[月份]月[日期]日的[小时]点[分钟]分时(用[url])提醒大家[消息]
    - [x] 在[月份]月[每周or周几]的[小时]点[分钟]分时(用[url])提醒大家[消息]
    - [x] 取消在[月份]月[日期]日的[小时]点[分钟]分的提醒
    - [x] 取消在[月份]月[每周or周几]的[小时]点[分钟]分的提醒
    - [ ] 同意入群请求
    - [ ] 同意好友请求
    - [ ] 撤回[@xxx] [xxx]
    - [ ] 警告[@xxx]
    - [x] run[xxx]
- 涩图 `import _ "github.com/Yiwen-Chan/ZeroBot-Plugin/setutime"`
    - [x] 搜索图片[P站图片ID]
    - [x] 搜索图片[图片]
    - [x] 来份[涩图/二次元/风景/车万]
    - [x] 添加[涩图/二次元/风景/车万][P站图片ID]
    - [x] 删除[涩图/二次元/风景/车万][P站图片ID]
    - [x] >setu status
    - [x] >setu xml
    - [x] >setu pic
- 点歌 `import _ "github.com/Yiwen-Chan/ZeroBot-Plugin/music"`
    - [x] 点歌[xxx]
    - [x] 网易点歌[xxx]
    - [x] 酷我点歌[xxx]
    - [x] 酷狗点歌[xxx]
- shindan `import _ "github.com/Yiwen-Chan/ZeroBot-Plugin/shindan"`
    - [x] 今天是什么少女[@xxx] 
    - [x] 异世界转生[@xxx] 
    - [x] 卖萌[@xxx] 
- GitHub仓库搜索 `import _ "github.com/Yiwen-Chan/ZeroBot-Plugin/github"`
    - [x] >github [xxx] 
    - [x] >github -p [xxx] 
- 在线代码运行 `import _ "github.com/Yiwen-Chan/ZeroBot-Plugin/runcode"`
    - [x] >runcode help
    - [x] >runcode [on/off]
    - [x] >runcode [language] [code block] 
- TODO...

### 使用方法

本项目符合 [OneBot](https://github.com/howmanybots/onebot) 标准，可基于以下项目与机器人框架/平台进行交互
| 项目地址 | 平台 | 核心作者 | 备注 |
| --- | --- | --- | --- |
| [Yiwen-Chan/OneBot-YaYa](https://github.com/Yiwen-Chan/OneBot-YaYa) | [先驱](https://www.xianqubot.com/) | kanri |  |
| [richardchien/coolq-http-api](https://github.com/richardchien/coolq-http-api) | CKYU | richardchien | 可在 Mirai 平台使用 [mirai-native](https://github.com/iTXTech/mirai-native) 加载 |
| [Mrs4s/go-cqhttp](https://github.com/Mrs4s/go-cqhttp) | [MiraiGo](https://github.com/Mrs4s/MiraiGo) | Mrs4s |  |
| [yyuueexxiinngg/cqhttp-mirai](https://github.com/yyuueexxiinngg/cqhttp-mirai) | [Mirai](https://github.com/mamoe/mirai) | yyuueexxiinngg |  |
| [takayama-lily/onebot](https://github.com/takayama-lily/onebot) | [OICQ](https://github.com/takayama-lily/oicq) | takayama |  |

#### 本地运行
1. 下载安装 [Go](https://studygolang.com/dl/golang/go1.16.2.windows-amd64.msi) 环境
2. 下载安装 [TDM-GCC](https://github.com/jmeubank/tdm-gcc/releases) 或 MinGW，并添加到环境变量
3. [clone](https://github.com/Yiwen-Chan/ZeroBot-Plugin/archive/master.zip) 本项目，本地解压
4. 编辑 main.go 文件，内容按需修改
5. 双击 build.bat 文件 或 直接双击 run.bat 文件
6. 运行 OneBot 框架，并同时运行本插件

#### 利用 Actions 在线编译 (推荐)
1. 点击右上角 Fork 本项目，并转跳到自己 Fork 的仓库
2. 点击仓库上方的 Actions 按钮，确认使用 Actions
3. 编辑 main.go 文件，内容按需修改，提交修改后 Actions 自动执行
4. 点击 Actions 按钮，等待编译完成，在 Actions 里下载编译好的文件
5. 运行 OneBot 框架，并同时运行本插件
6. 啾咪~

### 特别感谢
- [ZeroBot](https://github.com/wdvxdr1123/ZeroBot)
- [ATRI](https://github.com/Kyomotoi/ATRI)


## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2FYiwen-Chan%2FZeroBot-Plugin.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2FYiwen-Chan%2FZeroBot-Plugin?ref=badge_large)
