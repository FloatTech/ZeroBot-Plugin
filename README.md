<div align="center">
  <img src=".github/yaya.jpg" width = "150" height = "150" alt="OneBot-YaYa"><br>
  <h1>ZeroBot-Plugin</h1>
  ZeroBot-Plugin 是 ZeroBot 的 实用插件合集<br><br>

  <img src="http://sayuri.fumiama.top/cmoe?name=ZeroBot-Plugin&theme=r34" />

[![YAYA](https://img.shields.io/badge/OneBot-YaYa-green.svg?style=social&logo=appveyor)](https://github.com/Yiwen-Chan/OneBot-YaYa)
[![GOCQ](https://img.shields.io/badge/OneBot-MiraiGo-green.svg?style=social&logo=appveyor)](https://github.com/Mrs4s/go-cqhttp)
[![OICQ](https://img.shields.io/badge/OneBot-OICQ-green.svg?style=social&logo=appveyor)](https://github.com/takayama-lily/node-onebot)
[![MIRAI](https://img.shields.io/badge/OneBot-Mirai-green.svg?style=social&logo=appveyor)](https://github.com/yyuueexxiinngg/onebot-kotlin)

[![Go Report Card](https://goreportcard.com/badge/github.com/FloatTech/ZeroBot-Plugin?style=flat-square&logo=go)](https://goreportcard.com/report/github.com/github.com/FloatTech/ZeroBot-Plugin)
[![Badge](https://img.shields.io/badge/onebot-v11-black?logo=data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAHAAAABwCAMAAADxPgR5AAAAGXRFWHRTb2Z0d2FyZQBBZG9iZSBJbWFnZVJlYWR5ccllPAAAAAxQTFRF////29vbr6+vAAAAk1hCcwAAAAR0Uk5T////AEAqqfQAAAKcSURBVHja7NrbctswDATQXfD//zlpO7FlmwAWIOnOtNaTM5JwDMa8E+PNFz7g3waJ24fviyDPgfhz8fHP39cBcBL9KoJbQUxjA2iYqHL3FAnvzhL4GtVNUcoSZe6eSHizBcK5LL7dBr2AUZlev1ARRHCljzRALIEog6H3U6bCIyqIZdAT0eBuJYaGiJaHSjmkYIZd+qSGWAQnIaz2OArVnX6vrItQvbhZJtVGB5qX9wKqCMkb9W7aexfCO/rwQRBzsDIsYx4AOz0nhAtWu7bqkEQBO0Pr+Ftjt5fFCUEbm0Sbgdu8WSgJ5NgH2iu46R/o1UcBXJsFusWF/QUaz3RwJMEgngfaGGdSxJkE/Yg4lOBryBiMwvAhZrVMUUvwqU7F05b5WLaUIN4M4hRocQQRnEedgsn7TZB3UCpRrIJwQfqvGwsg18EnI2uSVNC8t+0QmMXogvbPg/xk+Mnw/6kW/rraUlvqgmFreAA09xW5t0AFlHrQZ3CsgvZm0FbHNKyBmheBKIF2cCA8A600aHPmFtRB1XvMsJAiza7LpPog0UJwccKdzw8rdf8MyN2ePYF896LC5hTzdZqxb6VNXInaupARLDNBWgI8spq4T0Qb5H4vWfPmHo8OyB1ito+AysNNz0oglj1U955sjUN9d41LnrX2D/u7eRwxyOaOpfyevCWbTgDEoilsOnu7zsKhjRCsnD/QzhdkYLBLXjiK4f3UWmcx2M7PO21CKVTH84638NTplt6JIQH0ZwCNuiWAfvuLhdrcOYPVO9eW3A67l7hZtgaY9GZo9AFc6cryjoeFBIWeU+npnk/nLE0OxCHL1eQsc1IciehjpJv5mqCsjeopaH6r15/MrxNnVhu7tmcslay2gO2Z1QfcfX0JMACG41/u0RrI9QAAAABJRU5ErkJggg==)](https://github.com/howmanybots/onebot)
[![Badge](https://img.shields.io/badge/zerobot-v1.4.1-black?style=flat-square&logo=go)](https://github.com/wdvxdr1123/ZeroBot)
[![License](https://img.shields.io/github/license/FloatTech/ZeroBot-Plugin.svg?style=flat-square&logo=gnu)](https://raw.githubusercontent.com/FloatTech/ZeroBot-Plugin/master/LICENSE)
[![qq group](https://img.shields.io/badge/group-1048452984-red?style=flat-square&logo=tencent-qq)](https://jq.qq.com/?_wv=1027&k=QMb7x1mM)

本项目符合 [OneBot](https://github.com/howmanybots/onebot) 标准，可基于以下项目与机器人框架/平台进行交互
| 项目地址 | 平台 | 核心作者 |
| --- | --- | --- |
| [Mrs4s/go-cqhttp](https://github.com/Mrs4s/go-cqhttp) | [MiraiGo](https://github.com/Mrs4s/MiraiGo) | Mrs4s |
| [yyuueexxiinngg/cqhttp-mirai](https://github.com/yyuueexxiinngg/cqhttp-mirai) | [Mirai](https://github.com/mamoe/mirai) | yyuueexxiinngg |
| [takayama-lily/onebot](https://github.com/takayama-lily/onebot) | [OICQ](https://github.com/takayama-lily/oicq) | takayama |

</div>

> 如果您不知道什么是 [OneBot](https://github.com/howmanybots/onebot) 或不希望运行多个程序，还可以直接前往 [gocqzbp](https://github.com/FloatTech/gocqzbp) 的 [Release](https://github.com/FloatTech/gocqzbp/releases) 页面下载单一可执行文件或前往 [Packages](https://github.com/FloatTech/gocqzbp/pkgs/container/gocqzbp) 页面使用`docker`，运行后按提示登录即可。

## 命令行参数
> `[]`代表是可选参数
```bash
zerobot [-h] [-t token] [-u url] [-n nickname] [-p prefix] [-d|w] [-g 监听地址:端口] [qq1 qq2 qq3 ...] [&]
```
- **-h**: 显示帮助
- **-t token**: 设置`AccessToken`，默认为空
- **-u url**: 设置`Url`，默认为`ws://127.0.0.1:6700`
- **-n nickname**: 设置默认昵称，默认为`椛椛`
- **-p prefix**: 设置命令前缀，默认为`/`
- **-d|w**: 开启 debug | warning 级别及以上日志输出
- **-g 监听地址:端口**: 在 http://监听地址:端口 上开启 [webgui](https://github.com/FloatTech/bot-manager)
- **qqs**: superusers 的 qq 号
- **&**: 驻留在后台，必须放在最后，仅`Linux`下有效

## 功能
> 在编译时，以下功能除插件控制外，均可通过注释`main.go`中的相应`import`而物理禁用，减小插件体积。
> 通过插件控制，还可动态管理某个功能在某个群的打开/关闭。
> 插件的优先级为`import`的先后顺序
- **web管理** `import _ "github.com/FloatTech/zbputils/control/web"`
    - 开启后可执行文件大约增加 5M ，默认注释不开启。如需开启请自行编辑`main.go`取消注释
    - 需要配合 [webgui](https://github.com/FloatTech/bot-manager) 使用
- **动态加载插件** `import _ github.com/FloatTech/ZeroBot-Plugin-Dynamic/dyloader`
    - 本功能需要`cgo`，故已分离出主线。详见[ZeroBot-Plugin-Dynamic](https://github.com/FloatTech/ZeroBot-Plugin-Dynamic)
- **插件控制**
    - [x] /启用 xxx (在发送的群/用户启用xxx)
    - [x] /禁用 xxx (在发送的群/用户禁用xxx)
    - [x] /全局启用 xxx
    - [x] /全局禁用 xxx
    - [x] /还原 xxx (在发送的群/用户还原xxx的开启状态到初始状态)
    - 注：当全局未配置或与默认相同时，状态取决于单独配置，后备为默认配置；当全局与默认不同时，状态取决于全局配置，单独配置失效。
    - [x] /禁止 service qq1 qq2... (禁止 qqs 使用服务 service)
    - [x] /允许 service qq1 qq2... (重新允许 qqs 使用服务 service)
    - [x] /封禁 qq1 qq2... (禁止 qqs 使用全部服务)
    - [x] /解封 qq1 qq2... (允许 qqs 使用全部服务)
    - [x] /用法 xxx
    - [x] /服务列表
    - [x] /服务详情
    - [x] @Bot 插件冲突检测 (会在本群发送一条消息并在约 1s 后撤回以检测其它同类 bot 中已启用的插件并禁用)
- **聊天** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_chat"`
    - [x] [BOT名字]
    - [x] [戳一戳BOT]
    - [x] 空调开
    - [x] 空调关
    - [x] 群温度
    - [x] 设置温度[正整数]
- **词典匹配回复** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_thesaurus"`
    - [x] @Bot 关键词
- **ATRI** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_atri"`
    - [x] 具体指令看 /用法 atri
    - 注：本插件基于 [ATRI](https://github.com/Kyomotoi/ATRI) ，为 Golang 移植版
- **群管** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_manager"`
    - [x] 禁言[@xxx][分钟]
    - [x] 解除禁言[@xxx]
    - [x] 我要自闭 | 禅定 x [分钟 | 小时 | 天]
    - [x] 开启全员禁言
    - [x] 解除全员禁言
    - [x] 升为管理[@xxx]
    - [x] 取消管理[@xxx]
    - [x] 修改名片[@xxx][xxx]
    - [x] 修改头衔[@xxx][xxx]
    - [x] 申请头衔[xxx]
    - [x] 踢出群聊[@xxx]
    - [x] 退出群聊[群号]@Bot
    - [x] *入群欢迎
    - [x] *退群通知
    - [x] 设置欢迎语[欢迎~]
    - [x] 在[MM]月[dd]日的[hh]点[mm]分时(用[url])提醒大家[xxx]
    - [x] 在[MM]月[每周 | 周几]的[hh]点[mm]分时(用[url])提醒大家[xxx]
    - [x] 取消在[MM]月[dd]日的[hh]点[mm]分的提醒
    - [x] 取消在[MM]月[每周 | 周几]的[hh]点[mm]分的提醒
    - [x] 在"cron"时(用[url])提醒大家[xxx]
    - [x] 取消在"cron"的提醒
    - [x] 列出所有提醒
    - [x] 翻牌
    - [x] [开启 | 关闭]入群验证
    - [x] [开启 | 关闭]gist加群自动审批
    - [ ] 同意好友请求
    - [ ] 撤回[@xxx] [xxx]
    - [ ] 警告[@xxx]
    - [x] run[xxx]
    - 注：使用gist加群自动审批，请在群介绍添加以下说明，同时开启`需要回答问题并由管理员审核`：加群请在github新建一个gist，其文件名为本群群号的字符串的md5(小写)，内容为一行，是当前unix时间戳(10分钟内有效)。然后请将您的用户名和gist哈希(小写)按照username/gisthash的格式填写到回答即可。
- **GitHub仓库搜索** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_github"`
    - [x] >github [xxx]
    - [x] >github -p [xxx]
- **在线代码运行** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_runcode"`
    - [x] > runcode [language] help
    - [x] > runcode [language] [code block]
- **点歌** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_music"`
    - [x] 点歌[xxx]
    - [x] 网易点歌[xxx]
    - [x] 酷我点歌[xxx]
    - [x] 酷狗点歌[xxx]
- **shindan** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_shindan"`
    - [x] 今天是什么少女[@xxx]
    - [x] 异世界转生[@xxx]
    - [x] 卖萌[@xxx]
    - [x] 抽老婆[@xxx]
- **AIWife** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_aiwife"`
    - [x] waifu | 随机waifu(从[100000个AI生成的waifu](https://www.thiswaifudoesnotexist.net/)中随机一位)
- **gif** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_gif"`
    - [x] 爬[@xxx]
    - [x] 摸[@xxx]
    - [x] 搓[@xxx]
    - 注：更多指令见项目 --> https://github.com/FloatTech/ZeroBot-Plugin-Gif
- **base16384加解密** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_b14"`
    - [x] 加密xxx
    - [x] 解密xxx
    - [x] 用yyy加密xxx
    - [x] 用yyy解密xxx
- **摸鱼** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_moyu"`
    - [x] /启用 moyu
    - [x] /禁用 moyu
- **摸鱼人日历** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_moyu_calendar"`
    - [x] /启用 moyucalendar
    - [x] /禁用 moyucalendar
- **涩图** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_setutime"`
    - [x] 来份[涩图/二次元/风景/车万]
    - [x] 添加[涩图/二次元/风景/车万][P站图片ID]
    - [x] 删除[涩图/二次元/风景/车万][P站图片ID]
    - [x] > setu status
- **本地涩图** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_nativesetu"`
    - [x] 本地[xxx]
    - [x] 刷新本地[xxx]
    - [x] 设置本地setu绝对路径[xxx]
    - [x] 刷新所有本地setu
    - [x] 所有本地setu分类
    - 注：刷新文件夹较慢，请耐心等待刷新完成，会提示“成功”。
- **nsfw图片识别** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_nsfw"`
    - [x] nsfw打分[图片]
    - [x] 当图片属于非 neutral 类别时自动发送评价(默认禁用，启用输入 /启用 nsfwauto)
- **lolicon** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_lolicon"`
    - [x] 来份萝莉
- **搜图** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_saucenao"`
    - [x] 以图搜图 | 搜索图片 | 以图识图[图片]
    - [x] 搜图[P站图片ID]
- **搜番** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_tracemoe"`
    - [x] 搜番 | 搜索番剧[图片]
- **随机图片与AI点评** `import _ github.com/FloatTech/ZeroBot-Plugin/plugin_acgimage`
    - [x] 随机图片(评级大于6的图将私发)
    - [x] 直接随机(无r18检测，务必小心，仅管理可用)
    - [x] 设置随机图片网址[url]
    - [x] 太涩了(撤回最近发的图)
    - [x] 评价图片(发送一张图片让bot评分)
- **DeepDanbooru二次元图标签识别** `import _ github.com/FloatTech/ZeroBot-Plugin/plugin_danbooru`
    - [x] 鉴赏图片[图片]
- **叔叔的AI二次元图片放大** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_scale"`
    - [x] 放大图片[图片]
- **每日运势** `import _ github.com/FloatTech/ZeroBot-Plugin/plugin_fortune`
    - [x] 运势 | 抽签
    - [x] 设置底图[车万 DC4 爱因斯坦 星空列车 樱云之恋 富婆妹 李清歌 公主连结 原神 明日方舟 碧蓝航线 碧蓝幻想 战双 阴阳师 赛马娘]
- **睡眠管理** `import _ github.com/FloatTech/ZeroBot-Plugin/plugin_sleep_manage`
    - [x] 早安 | 晚安
- **浅草寺求签** `import _ github.com/FloatTech/ZeroBot-Plugin/plugin_omikuji`
    - [x] 求签 | 占卜
    - [x] 解签
- **漂流瓶** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_drift_bottle"`
    - [x] (在群xxx)丢漂流瓶(到频道xxx) [消息]
    - [x] (从频道xxx)捡漂流瓶
    - [x] @BOT 创建频道 xxx
    - [x] 跳入(频道)海中
    - [x] 注：不显式限制时，私聊发送可在所有群抽到，群聊发送仅可在本群抽到，默认频道为 global
- **猜单词** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_wordle"`
    - [x] 个人猜单词
    - [x] 团队猜单词
- **bilibili** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_bilibili"`
    - [x] >vup info [名字 | uid]
    - [x] >user info [名字 | uid]
    - [x] /开启粉丝日报
- **嘉然** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_diana"`
    - [x] 小作文
    - [x] 发大病
    - [x] 教你一篇小作文[作文]
    - [x] [回复]查重
- **鬼东西** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_wtf"`
    - [x] 鬼东西列表
    - [x] 查询鬼东西[序号][@xxx]
    - 注：由于需要科学，默认注释。
- **AIfalse** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_ai_false"`
    - [x] 查询计算机当前活跃度: [检查身体 | 自检 | 启动自检 | 系统状态]
    - [x] 清理缓存 (仅适用于 gocq 且需要 bot 的运行目录和 gocq 相同)
    - [ ] 简易语音
    - [ ] 爬图合成 [@xxx]
- **抽wife** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_nativewife"`
    - [x] 抽wife[@xxx]
    - [x] 添加wife[名字][图片]
    - [x] 删除wife[名字]
    - [x] [让 | 不让]所有人均可添加wife
    - 注：不同群添加后不会重叠
- **minecraft** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_minecraft"`
    - [x] /mcstart xxx
    - [x] /mcstop xxx
    - [x] /mclist servername
    - 注：此功能实现依赖[MCSManager](https://github.com/Suwings/MCSManager)项目对服务器的管理api，mc服务器如果没有在该管理平台部署此功能无效
- **炉石** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_hs"`
    - [x] 搜卡[xxxx]
    - [x] [卡组代码xxx]
    - 注：更多搜卡指令参数：https://hs.fbigame.com/misc/searchhelp
- **人工智能回复** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_ai_reply"`
    - [x] @Bot 任意文本(任意一句话回复)
    - [x] 设置回复模式[青云客 | 小爱]
- **关键字搜图** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_image_finder"`
    - [x] 来张 [xxx]
- **拼音首字母释义工具** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_nbnhhsh"`
    - [x] ?? [缩写]
- **选择困难症帮手** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_choose"`
    - [x] 选择[选择项1]还是[选项2]还是[更多选项]
- **投胎** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_reborn"`
    - [x] reborn
    - 注：本插件来源于[tgbot](https://github.com/YukariChiba/tgbot/blob/main/modules/Reborn.py)
- **翻译** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_translation"`
    - [x] >TL 你好
- **vtb语录** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_vtb_quotation"`
    - [x] vtb语录
    - [x] 随机vtb
    - [x] 更新vtb
- **书评** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_book_review"`
    - [x] 书评[xxx]
    - [x] 随机书评
- **coser** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_coser" `
    - [x] coser
- **小说** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_novel" `
    - [x] 小说[xxx]
- **沙雕app插件** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_shadiao"`
    - [x] 哄我
    - [x] 渣我
    - [x] 来碗绿茶
    - [x] 发个朋友圈
    - [x] 来碗毒鸡汤
    - [x] 讲个段子
- **笑话** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_funny"`
    - [x] 讲个笑话[@xxx] | 讲个笑话[qq号]
- **抽象话** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_chouxianghua"`
    - [x] 抽象翻译[xxx]
- **合成emoji** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_emojimix"`
    - [x] [emoji][emoji]
- **绝绝子** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_juejuezi"`
    - [x] 喝奶茶绝绝子 | 绝绝子吃饭
- **藏头诗** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_cangtoushi"`
    - [x] 藏头诗[xxx]
    - [x] 藏尾诗[xxx]
- **cp短打** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_cpstory"`
    - [x] 组cp[@xxx][@xxx]
    - [x] 磕cp大老师 雪乃
- **签到得分** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_score"`
    - [x] 签到
    - [x] 获得签到背景[@xxx] | 获得签到背景
- **骂人** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_curse"`
    - [x] 骂我
    - [x] 大力骂我
- **b站推送** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_bilibili_push"`
    - [x] 添加订阅[uid]
    - [x] 取消订阅[uid]
    - [x] 取消动态订阅[uid]
    - [x] 取消直播订阅[uid]
    - [x] 推送列表
- **网易云音乐热评** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_wangyiyun"`
    - [x] 来份网易云热评
- **b站视频链接解析** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_bilibili_parse"`
    - [x] https://www.bilibili.com/video/BV1xx411c7BF | https://www.bilibili.com/video/av1605 | https://b23.tv/I8uzWCA | https://www.bilibili.com/video/bv1xx411c7BF
- **煎蛋网无聊图** `import _ "github.com/FloatTech/ZeroBot-Plugin/plugin_jandan"`
    - [x] 来份屌图
    - [x] 更新屌图
- **TODO...**

## 使用方法

### 使用稳定版/测试版 (推荐)

可以前往[Release](https://github.com/FloatTech/ZeroBot-Plugin/releases)页面下载对应系统版本可执行文件，编译时开启了全部插件。您还可以选择 [gocqzbp](https://github.com/FloatTech/gocqzbp) 的 [Release](https://github.com/FloatTech/gocqzbp/releases) 或 [Package](https://github.com/FloatTech/gocqzbp/pkgs/container/gocqzbp)，它是 [Mrs4s/go-cqhttp](https://github.com/Mrs4s/go-cqhttp) 与本插件的合体。

### 本地直接运行

1. 下载安装最新 [Go](https://studygolang.com/dl) 环境
2. 下载本项目[压缩包](https://github.com/FloatTech/ZeroBot-Plugin/archive/master.zip)，本地解压
3. 编辑 main.go 文件，内容按需修改
4. 运行 OneBot 框架
5. `Windows`下双击 run.bat 文件，`Linux`下使用 run.sh 运行本插件

### 编译运行

#### 利用 Actions 在线编译

1. 点击右上角 Fork 本项目，并转跳到自己 Fork 的仓库
2. 点击仓库上方的 Actions 按钮，确认使用 Actions
3. 编辑 main.go 文件，内容按需修改
4. 前往 Release 页面发布一个 Release，`tag`形如`v1.2.3`，以触发稳定版编译流程
5. 点击 Actions 按钮，等待编译完成，回到 Release 页面下载编译好的文件
6. 运行 OneBot 框架，并同时运行本插件
7. 啾咪~

#### 本地编译/交叉编译

1. 下载安装最新 [Go](https://studygolang.com/dl) 环境
2. clone 并进入本项目，下载所需包

```bash
git clone --depth=1 https://github.com/FloatTech/ZeroBot-Plugin.git
cd ZeroBot-Plugin
go version
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GO111MODULE=auto
go mod tidy
```

3. 编辑 main.go 文件，内容按需修改
4. 按照平台输入命令编译，下面举了一些例子

```bash
# 本机平台
go build -ldflags "-s -w" -o zerobot -trimpath
# x64 Linux 平台 如各种云服务器
GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o zerobot -trimpath
# x64 Windows 平台 如大多数家用电脑
GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o zerobot.exe -trimpath
# armv6 Linux 平台 如树莓派 zero W
GOOS=linux GOARCH=arm GOARM=6 CGO_ENABLED=0 go build -ldflags "-s -w" -o zerobot -trimpath
# （由于引入了github.com/fumiama/sqlite3，本项不再可用）mips Linux 平台 如 路由器 wndr4300
GOOS=linux GOARCH=mips GOMIPS=softfloat CGO_ENABLED=0 go build -ldflags "-s -w" -o zerobot -trimpath
```

5. 运行 OneBot 框架，并同时运行本插件

## 特别感谢

- [ZeroBot](https://github.com/wdvxdr1123/ZeroBot)
- [ATRI](https://github.com/Kyomotoi/ATRI)

## License

[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2FFloatTech%2FZeroBot-Plugin.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2FFloatTech%2FZeroBot-Plugin?ref=badge_large)
