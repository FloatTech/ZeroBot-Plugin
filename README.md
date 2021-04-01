<img src="https://socialify.git.ci/Yiwen-Chan/ZeroBot-Plugin/image?forks=1&issues=1&language=1&owner=1&pulls=1&stargazers=1&theme=Light" alt="ZeroBot-Plugin" width="640" height="320" />

# ZeroBot-Plugin

[![Go Report Card](https://goreportcard.com/badge/github.com/Yiwen-Chan/ZeroBot-Plugin)](https://goreportcard.com/report/github.com/github.com/Yiwen-Chan/ZeroBot-Plugin)
![Badge](https://img.shields.io/badge/OneBot-v11-black)
![Badge](https://img.shields.io/badge/ZeroBot-v1.0.1-black)
[![License](https://img.shields.io/github/license/Yiwen-Chan/ZeroBot-Plugin.svg)](https://raw.githubusercontent.com/Yiwen-Chan/ZeroBot-Plugin/master/LICENSE)
[![反馈群](https://img.shields.io/badge/反馈群-1048452984-green.svg)](https://jq.qq.com/?_wv=1027&k=QMb7x1mM)


### 功能
- 群管
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
    - [ ] 同意入群请求
    - [ ] 同意好友请求
    - [ ] 撤回[@xxx] [xxx]
    - [ ] 警告[@xxx]
    - [x] run[xxx]
- 涩图
    - [x] 来份[涩图/二次元/风景]
    - [x] 添加[涩图/二次元/风景][P站图片ID]
    - [x] 删除[涩图/二次元/风景][P站图片ID]
    - [x] setu -s
    - [x] setu -x
    - [x] setu -p
- 点歌
    - [x] 点歌[xxx]
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

#### 本地编译
1. 下载安装 [Go](https://studygolang.com/dl/golang/go1.16.2.windows-amd64.msi) 环境
2. 下载安装 [TDM-GCC](https://github.com/jmeubank/tdm-gcc/releases)，并添加到环境变量
3. [clone](https://github.com/Yiwen-Chan/ZeroBot-Plugin/archive/master.zip) 本项目，本地解压
4. 编辑 main.go 文件，内容按需修改
5. 双击点击 build.bat 文件
6. 运行框架，并同时运行本插件

#### 利用 Actions 编译 (推荐)
1. 点击右上角 Fork 本项目，并转跳到自己 Fork 的仓库
2. 点击仓库上方的 Actions 按钮，确认使用 Actions
3. 编辑 main.go 文件，内容按需修改，返回仓库
4. 点击 Actions 按钮，等待编译完成，在 Actions 里下载编译好的文件
5. 运行框架，并同时运行本插件

