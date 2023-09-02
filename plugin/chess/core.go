package chess

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/color"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/FloatTech/floatbox/binary"
	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/gg"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/img/text"
	"github.com/RomiChan/syncx"
	"github.com/jinzhu/gorm"
	"github.com/notnil/chess"
	"github.com/notnil/chess/image"
	log "github.com/sirupsen/logrus"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const eloDefault = 500

var chessRoomMap syncx.Map[int64, *chessRoom]

type chessRoom struct {
	chessGame    *chess.Game
	whitePlayer  int64
	whiteName    string
	blackPlayer  int64
	blackName    string
	drawPlayer   int64
	lastMoveTime int64
	isBlindfold  bool
	whiteErr     bool // 违例记录（盲棋用）
	blackErr     bool
}

// game 下棋
func game(groupCode, senderUin int64, senderName string) message.Message {
	return createGame(false, groupCode, senderUin, senderName)
}

// blindfold 盲棋
func blindfold(groupCode, senderUin int64, senderName string) message.Message {
	return createGame(true, groupCode, senderUin, senderName)
}

// abort 中断对局
func abort(groupCode int64) message.Message {
	if room, ok := chessRoomMap.Load(groupCode); ok {
		return abortGame(*room, groupCode, "对局已被管理员中断，游戏结束。")
	}
	return simpleText("对局不存在，发送「下棋」或「chess」可创建对局。")
}

// draw 和棋
func draw(groupCode, senderUin int64) message.Message {
	// 检查对局是否存在
	room, ok := chessRoomMap.Load(groupCode)
	if !ok {
		return simpleText("对局不存在，发送「下棋」或「chess」可创建对局。")
	}
	// 检查消息发送者是否为对局中的玩家
	if senderUin != room.whitePlayer && senderUin != room.blackPlayer {
		return textWithAt(senderUin, "您不是对局中的玩家，无法请求和棋。")
	}
	// 处理和棋逻辑
	room.lastMoveTime = time.Now().Unix()
	if room.drawPlayer == 0 {
		room.drawPlayer = senderUin
		chessRoomMap.Store(groupCode, room)
		return textWithAt(senderUin, "请求和棋，发送「和棋」或「draw」接受和棋。走棋视为拒绝和棋。")
	}
	if room.drawPlayer == senderUin {
		return textWithAt(senderUin, "已发起和棋请求，请勿重复发送。")
	}
	err := room.chessGame.Draw(chess.DrawOffer)
	if err != nil {
		log.Debugln("[chess]", "Fail to draw a game.", err)
		return textWithAt(senderUin, fmt.Sprintln("程序发生了错误，和棋失败，请反馈开发者修复 bug。\nERROR:", err))
	}
	chessString := getChessString(*room)
	eloString := ""
	if len(room.chessGame.Moves()) > 4 {
		// 若走子次数超过 4 认为是有效对局，存入数据库
		dbService := newDBService()
		if err := dbService.createPGN(chessString, room.whitePlayer, room.blackPlayer, room.whiteName, room.blackName); err != nil {
			log.Debugln("[chess]", "Fail to create PGN.", err)
			return message.Message{message.Text("ERROR: ", err)}
		}
		whiteScore, blackScore := 0.5, 0.5
		elo, err := getELOString(*room, whiteScore, blackScore)
		if err != nil {
			log.Debugln("[chess]", "Fail to get eloString.", eloString, err)
			return message.Message{message.Text("ERROR: ", err)}
		}
		eloString = elo
	}
	replyMsg := textWithAt(senderUin, "接受和棋，游戏结束。\n"+eloString+chessString)
	if inkscapeExists() {
		if err := cleanTempFiles(groupCode); err != nil {
			log.Debugln("[chess]", "Fail to clean temp files", err)
			return message.Message{message.Text("ERROR: ", err)}
		}
	}
	chessRoomMap.Delete(groupCode)
	return replyMsg
}

// resign 认输
func resign(groupCode, senderUin int64) message.Message {
	// 检查对局是否存在
	room, ok := chessRoomMap.Load(groupCode)
	if !ok {
		return simpleText("对局不存在，发送「下棋」或「chess」可创建对局。")
	}
	// 检查是否是当前游戏玩家
	if senderUin != room.whitePlayer && senderUin != room.blackPlayer {
		return textWithAt(senderUin, "不是对局中的玩家，无法认输。")
	}
	// 如果对局未建立，中断对局
	if room.whitePlayer == 0 || room.blackPlayer == 0 {
		chessRoomMap.Delete(groupCode)
		return simpleText("对局已释放。")
	}
	// 计算认输方
	var resignColor chess.Color
	if senderUin == room.whitePlayer {
		resignColor = chess.White
	} else {
		resignColor = chess.Black
	}
	if isAprilFoolsDay() {
		if resignColor == chess.White {
			resignColor = chess.Black
		} else {
			resignColor = chess.White
		}
	}
	room.chessGame.Resign(resignColor)
	chessString := getChessString(*room)
	eloString := ""
	if len(room.chessGame.Moves()) > 4 {
		// 若走子次数超过 4 认为是有效对局，存入数据库
		dbService := newDBService()
		if err := dbService.createPGN(chessString, room.whitePlayer, room.blackPlayer, room.whiteName, room.blackName); err != nil {
			log.Debugln("[chess]", "Fail to create PGN.", err)
			return message.Message{message.Text("ERROR: ", err)}
		}
		whiteScore, blackScore := 1.0, 1.0
		if resignColor == chess.White {
			whiteScore = 0.0
		} else {
			blackScore = 0.0
		}
		elo, err := getELOString(*room, whiteScore, blackScore)
		if err != nil {
			log.Debugln("[chess]", "Fail to get eloString.", eloString, err)
			return message.Message{message.Text("ERROR: ", err)}
		}
		eloString = elo
	}
	replyMsg := textWithAt(senderUin, "认输，游戏结束。\n"+eloString+chessString)
	if isAprilFoolsDay() {
		replyMsg = textWithAt(senderUin, "对手认输，游戏结束，你胜利了。\n"+eloString+chessString)
	}
	// 删除临时文件
	if inkscapeExists() {
		if err := cleanTempFiles(groupCode); err != nil {
			log.Debugln("[chess]", "Fail to clean temp files", err)
			return message.Message{message.Text("ERROR: ", err)}
		}
	}
	chessRoomMap.Delete(groupCode)
	return replyMsg
}

// play 走棋
func play(senderUin int64, groupCode int64, moveStr string) message.Message {
	// 检查对局是否存在
	room, ok := chessRoomMap.Load(groupCode)
	if !ok {
		return nil
	}
	// 不是对局中的玩家，忽略消息
	if (senderUin != room.whitePlayer) && (senderUin != room.blackPlayer) && !isAprilFoolsDay() {
		return nil
	}
	// 对局未建立
	if (room.whitePlayer == 0) || (room.blackPlayer == 0) {
		return textWithAt(senderUin, "请等候其他玩家加入游戏。")
	}
	// 需要对手走棋
	if ((senderUin == room.whitePlayer) && (room.chessGame.Position().Turn() != chess.White)) || ((senderUin == room.blackPlayer) && (room.chessGame.Position().Turn() != chess.Black)) {
		return textWithAt(senderUin, "请等待对手走棋。")
	}
	room.lastMoveTime = time.Now().Unix()
	// 走棋
	if err := room.chessGame.MoveStr(moveStr); err != nil {
		// 指令错误时检查
		if !room.isBlindfold {
			// 未开启盲棋，提示指令错误
			return simpleText(fmt.Sprintf("移动「%s」违规，请检查，格式请参考「代数记谱法」(Algebraic notation)。", moveStr))
		}
		// 开启盲棋，判断违例情况
		var currentPlayerColor chess.Color
		if senderUin == room.whitePlayer {
			currentPlayerColor = chess.White
		} else {
			currentPlayerColor = chess.Black
		}
		// 第一次违例，提示
		_flag := false
		if (currentPlayerColor == chess.White) && !room.whiteErr {
			room.whiteErr = true
			chessRoomMap.Store(groupCode, room)
			_flag = true
		}
		if (currentPlayerColor == chess.Black) && !room.blackErr {
			room.blackErr = true
			chessRoomMap.Store(groupCode, room)
			_flag = true
		}
		if _flag {
			return simpleText(fmt.Sprintf("移动「%s」违例，再次违例会立即判负。", moveStr))
		}
		// 出现多次违例，判负
		room.chessGame.Resign(currentPlayerColor)
		chessString := getChessString(*room)
		replyMsg := textWithAt(senderUin, "违例两次，游戏结束。\n"+chessString)
		// 删除临时文件
		if inkscapeExists() {
			if err := cleanTempFiles(groupCode); err != nil {
				log.Debugln("[chess]", "Fail to clean temp files", err)
				return message.Message{message.Text("ERROR: ", err)}
			}
		}
		chessRoomMap.Delete(groupCode)
		return replyMsg
	}
	// 走子之后，视为拒绝和棋
	if room.drawPlayer != 0 {
		room.drawPlayer = 0
		chessRoomMap.Store(groupCode, room)
	}
	// 生成棋盘图片
	var boardImgEle message.MessageSegment
	if !room.isBlindfold {
		boardMsg, ok, errMsg := getBoardElement(groupCode)
		boardImgEle = boardMsg
		if !ok {
			return errorText(errMsg)
		}
	}
	// 检查游戏是否结束
	if room.chessGame.Method() != chess.NoMethod {
		whiteScore, blackScore := 0.5, 0.5
		var msgBuilder strings.Builder
		msgBuilder.WriteString("游戏结束，")
		switch room.chessGame.Method() {
		case chess.FivefoldRepetition:
			msgBuilder.WriteString("和棋，因为五次重复走子。\n")
		case chess.SeventyFiveMoveRule:
			msgBuilder.WriteString("和棋，因为七十五步规则。\n")
		case chess.InsufficientMaterial:
			msgBuilder.WriteString("和棋，因为不可能将死。\n")
		case chess.Stalemate:
			msgBuilder.WriteString("和棋，因为逼和（无子可动和棋）。\n")
		case chess.Checkmate:
			var winner string
			if room.chessGame.Position().Turn() == chess.White {
				whiteScore = 0.0
				blackScore = 1.0
				winner = "黑方"
			} else {
				whiteScore = 1.0
				blackScore = 0.0
				winner = "白方"
			}
			msgBuilder.WriteString(winner)
			msgBuilder.WriteString("胜利，因为将杀。\n")
		case chess.NoMethod:
		case chess.Resignation:
		case chess.DrawOffer:
		case chess.ThreefoldRepetition:
		case chess.FiftyMoveRule:
		default:
		}
		chessString := getChessString(*room)
		eloString := ""
		if len(room.chessGame.Moves()) > 4 {
			// 若走子次数超过 4 认为是有效对局，存入数据库
			dbService := newDBService()
			if err := dbService.createPGN(chessString, room.whitePlayer, room.blackPlayer, room.whiteName, room.blackName); err != nil {
				log.Debugln("[chess]", "Fail to create PGN.", err)
				return message.Message{message.Text("ERROR: ", err)}
			}
			// 仅有效对局才会计算等级分
			elo, err := getELOString(*room, whiteScore, blackScore)
			if err != nil {
				log.Debugln("[chess]", "Fail to get eloString.", eloString, err)
				return message.Message{message.Text("ERROR: ", err)}
			}
			eloString = elo
		}
		msgBuilder.WriteString(eloString)
		msgBuilder.WriteString(chessString)
		replyMsg := simpleText(msgBuilder.String())
		if !room.isBlindfold {
			replyMsg = append(replyMsg, boardImgEle)
		}
		if inkscapeExists() {
			if err := cleanTempFiles(groupCode); err != nil {
				log.Debugln("[chess]", "Fail to clean temp files", err)
				return message.Message{message.Text("ERROR: ", err)}
			}
		}
		chessRoomMap.Delete(groupCode)
		return replyMsg
	}
	// 提示玩家继续游戏
	var currentPlayer int64
	if room.chessGame.Position().Turn() == chess.White {
		currentPlayer = room.whitePlayer
	} else {
		currentPlayer = room.blackPlayer
	}
	return append(textWithAt(currentPlayer, "对手已走子，游戏继续。"), boardImgEle)
}

// ranking 排行榜
func ranking() message.Message {
	ranking, err := getRankingString()
	if err != nil {
		log.Debugln("[chess]", "Fail to get player ranking.", err)
		return simpleText(fmt.Sprintln("服务器错误，无法获取排行榜信息。请联系开发者修 bug。", err))
	}
	return simpleText(ranking)
}

// rate 获取等级分
func rate(senderUin int64, senderName string) message.Message {
	dbService := newDBService()
	rate, err := dbService.getELORateByUin(senderUin)
	if err == gorm.ErrRecordNotFound {
		return simpleText("没有查找到等级分信息。请至少进行一局对局。")
	}
	if err != nil {
		log.Debugln("[chess]", "Fail to get player rank.", err)
		return simpleText(fmt.Sprintln("服务器错误，无法获取等级分信息。请联系开发者修 bug。", err))
	}
	return simpleText(fmt.Sprintf("玩家「%s」目前的等级分：%d", senderName, rate))
}

// cleanUserRate 清空用户等级分
func cleanUserRate(senderUin int64) message.Message {
	dbService := newDBService()
	err := dbService.cleanELOByUin(senderUin)
	if err == gorm.ErrRecordNotFound {
		return simpleText("没有查找到等级分信息。请检查用户 uid 是否正确。")
	}
	if err != nil {
		log.Debugln("[chess]", "Fail to clean player rank.", err)
		return simpleText(fmt.Sprintln("服务器错误，无法清空等级分。请联系开发者修 bug。", err))
	}
	return simpleText(fmt.Sprintf("已清空用户「%d」的等级分。", senderUin))
}

// createGame 创建游戏
func createGame(isBlindfold bool, groupCode int64, senderUin int64, senderName string) message.Message {
	room, ok := chessRoomMap.Load(groupCode)
	if !ok {
		chessRoomMap.Store(groupCode, &chessRoom{
			chessGame:    chess.NewGame(),
			whitePlayer:  senderUin,
			whiteName:    senderName,
			blackPlayer:  0,
			blackName:    "",
			drawPlayer:   0,
			lastMoveTime: time.Now().Unix(),
			isBlindfold:  isBlindfold,
			whiteErr:     false,
			blackErr:     false,
		})
		if isBlindfold {
			return simpleText("已创建新的盲棋对局，发送「盲棋」或「blind」可加入对局。")
		}
		return simpleText("已创建新的对局，发送「下棋」或「chess」可加入对局。")
	}
	if room.blackPlayer != 0 {
		// 检测对局是否已存在超过 6 小时
		if (time.Now().Unix() - room.lastMoveTime) > 21600 {
			autoAbortMsg := abortGame(*room, groupCode, "对局已存在超过 6 小时，游戏结束。")
			autoAbortMsg = append(autoAbortMsg, message.Text("\n\n已有对局已被中断，如需创建新对局请重新发送指令。"))
			autoAbortMsg = append(autoAbortMsg, message.At(senderUin))
			return autoAbortMsg
		}
		// 对局在进行
		msg := textWithAt(senderUin, "对局已在进行中，无法创建或加入对局，当前对局玩家为：")
		if room.whitePlayer != 0 {
			msg = append(msg, message.At(room.whitePlayer))
		}
		if room.blackPlayer != 0 {
			msg = append(msg, message.At(room.blackPlayer))
		}
		msg = append(msg, message.Text("，群主或管理员发送「中断」或「abort」可中断对局（自动判和）。"))
		return msg
	}
	if senderUin == room.whitePlayer {
		return textWithAt(senderUin, "请等候其他玩家加入游戏。")
	}
	if room.isBlindfold && !isBlindfold {
		return simpleText("已创建盲棋对局，请加入或等待盲棋对局结束之后创建普通对局。")
	}
	if !room.isBlindfold && isBlindfold {
		return simpleText("已创建普通对局，请加入或等待普通对局结束之后创建盲棋对局。")
	}
	room.blackPlayer = senderUin
	room.blackName = senderName
	chessRoomMap.Store(groupCode, room)
	var boardImgEle message.MessageSegment
	if !room.isBlindfold {
		boardMsg, ok, errMsg := getBoardElement(groupCode)
		if !ok {
			return errorText(errMsg)
		}
		boardImgEle = boardMsg
	}
	if isBlindfold {
		return append(simpleText("黑棋已加入对局，请白方下棋。"), message.At(room.whitePlayer))
	}
	return append(simpleText("黑棋已加入对局，请白方下棋。"), message.At(room.whitePlayer), boardImgEle)
}

// abortGame 中断游戏
func abortGame(room chessRoom, groupCode int64, hint string) message.Message {
	err := room.chessGame.Draw(chess.DrawOffer)
	if err != nil {
		log.Debugln("[chess]", "Fail to draw a game.", err)
		return simpleText(fmt.Sprintln("程序发生了错误，和棋失败，请反馈开发者修复 bug。", err))
	}
	chessString := getChessString(room)
	if len(room.chessGame.Moves()) > 4 {
		dbService := newDBService()
		if err := dbService.createPGN(chessString, room.whitePlayer, room.blackPlayer, room.whiteName, room.blackName); err != nil {
			log.Debugln("[chess]", "Fail to create PGN.", err)
			return message.Message{message.Text("ERROR: ", err)}
		}
	}
	if inkscapeExists() {
		if err := cleanTempFiles(groupCode); err != nil {
			log.Debugln("[chess]", "Fail to clean temp files", err)
			return message.Message{message.Text("ERROR: ", err)}
		}
	}
	chessRoomMap.Delete(groupCode)
	msg := simpleText(hint)
	if room.whitePlayer != 0 {
		msg = append(msg, message.At(room.whitePlayer))
	}
	if room.blackPlayer != 0 {
		msg = append(msg, message.At(room.blackPlayer))
	}
	msg = append(msg, message.Text("\n\n"+chessString))
	return msg
}

// getBoardElement 获取棋盘图片的消息内容
func getBoardElement(groupCode int64) (message.MessageSegment, bool, string) {
	room, ok := chessRoomMap.Load(groupCode)
	if !ok {
		log.Debugln(fmt.Sprintf("No room for groupCode %d.", groupCode))
		return message.MessageSegment{}, false, "对局不存在"
	}
	// 未安装 inkscape 直接返回对局字符串
	// TODO: 使用原生 go 库渲染 svg
	if !inkscapeExists() {
		boardString := room.chessGame.Position().Board().Draw()
		boardImageB64, err := generateCharBoardImage(boardString)
		if err != nil {
			return message.MessageSegment{}, false, "生成棋盘图片时发生错误"
		}
		replyMsg := message.Image("base64://" + boardImageB64)
		return replyMsg, true, ""
	}
	// 获取高亮方块
	highlightSquare := make([]chess.Square, 0, 2)
	moves := room.chessGame.Moves()
	if len(moves) != 0 {
		lastMove := moves[len(moves)-1]
		highlightSquare = append(highlightSquare, lastMove.S1())
		highlightSquare = append(highlightSquare, lastMove.S2())
	}
	// 生成棋盘 svg 文件
	svgFilePath := path.Join(tempFileDir, fmt.Sprintf("%d.svg", groupCode))
	fenStr := room.chessGame.FEN()
	gameTurn := room.chessGame.Position().Turn()
	if err := generateBoardSVG(svgFilePath, fenStr, gameTurn, highlightSquare...); err != nil {
		log.Debugln("[chess]", "Unable to generate svg file.", err)
		return message.MessageSegment{}, false, "无法生成 svg 图片，请检查后台日志。"
	}
	// 调用 inkscape 将 svg 图片转化为 png 图片
	pngFilePath := path.Join(tempFileDir, fmt.Sprintf("%d.png", groupCode))
	if err := exec.Command("inkscape", "-w", "720", "-h", "720", svgFilePath, "-o", pngFilePath).Run(); err != nil {
		log.Debugln("[chess]", "Unable to convert to png.", err)
		return message.MessageSegment{}, false, "无法生成 png 图片，请检查 inkscape 安装情况及其依赖 libfuse。"
	}
	// 尝试读取 png 图片
	imgData, err := os.ReadFile(pngFilePath)
	if err != nil {
		log.Debugln("[chess]", fmt.Sprintf("Unable to read image file in %s.", pngFilePath), err)
		return message.MessageSegment{}, false, "无法读取 png 图片"
	}
	imgMsg := message.Image("base64://" + base64.StdEncoding.EncodeToString(imgData))
	return imgMsg, true, ""
}

// getELOString 获得玩家等级分的文本内容
func getELOString(room chessRoom, whiteScore, blackScore float64) (string, error) {
	if room.whitePlayer == 0 || room.blackPlayer == 0 {
		return "", nil
	}
	var msgBuilder strings.Builder
	msgBuilder.WriteString("玩家等级分：\n")
	dbService := newDBService()
	if err := updateELORate(room.whitePlayer, room.blackPlayer, room.whiteName, room.blackName, whiteScore, blackScore, dbService); err != nil {
		msgBuilder.WriteString("发生错误，无法更新等级分。")
		msgBuilder.WriteString(err.Error())
		return msgBuilder.String(), err
	}
	whiteRate, blackRate, err := getELORate(room.whitePlayer, room.blackPlayer, dbService)
	if err != nil {
		msgBuilder.WriteString("发生错误，无法获取等级分。")
		msgBuilder.WriteString(err.Error())
		return msgBuilder.String(), err
	}
	msgBuilder.WriteString(fmt.Sprintf("%s：%d\n%s：%d\n\n", room.whiteName, whiteRate, room.blackName, blackRate))
	return msgBuilder.String(), nil
}

// getRankingString 获取等级分排行榜的文本内容
func getRankingString() (string, error) {
	dbService := newDBService()
	eloList, err := dbService.getHighestRateList()
	if err != nil {
		return "", err
	}
	var msgBuilder strings.Builder
	msgBuilder.WriteString("当前等级分排行榜：\n\n")
	for _, elo := range eloList {
		msgBuilder.WriteString(fmt.Sprintf("%s: %d\n", elo.Name, elo.Rate))
	}
	return msgBuilder.String(), nil
}

func simpleText(msg string) message.Message {
	return []message.MessageSegment{message.Text(msg)}
}

func textWithAt(target int64, msg string) message.Message {
	if target == 0 {
		return simpleText("@全体成员 " + msg)
	}
	return []message.MessageSegment{message.At(target), message.Text(msg)}
}

func errorText(errMsg string) message.Message {
	return simpleText("发生错误，请联系开发者修 bug。\n错误信息：" + errMsg)
}

// updateELORate 更新 elo 等级分
// 当数据库中没有玩家的等级分信息时，自动新建一条记录
func updateELORate(whiteUin, blackUin int64, whiteName, blackName string, whiteScore, blackScore float64, dbService *chessDBService) error {
	whiteRate, err := dbService.getELORateByUin(whiteUin)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return err
		}
		// create white elo
		if err := dbService.createELO(whiteUin, whiteName, eloDefault); err != nil {
			return err
		}
		whiteRate = eloDefault
	}
	blackRate, err := dbService.getELORateByUin(blackUin)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return err
		}
		// create black elo
		if err := dbService.createELO(blackUin, blackName, eloDefault); err != nil {
			return err
		}
		blackRate = eloDefault
	}
	whiteRate, blackRate = calculateNewRate(whiteRate, blackRate, whiteScore, blackScore)
	// 更新白棋玩家的 ELO 等级分
	if err := dbService.updateELOByUin(whiteUin, whiteName, whiteRate); err != nil {
		return err
	}
	// 更新黑棋玩家的 ELO 等级分
	return dbService.updateELOByUin(blackUin, blackName, blackRate)
}

// cleanTempFiles 清理临时文件
func cleanTempFiles(groupCode int64) error {
	svgFilePath := path.Join(tempFileDir, fmt.Sprintf("%d.svg", groupCode))
	if err := os.Remove(svgFilePath); err != nil {
		return err
	}
	pngFilePath := path.Join(tempFileDir, fmt.Sprintf("%d.png", groupCode))
	return os.Remove(pngFilePath)
}

// generateCharBoardImage 生成文字版的棋盘
func generateCharBoardImage(boardString string) (string, error) {
	boardString = strings.Trim(boardString, "\n")
	const FontSize = 72
	h := FontSize*8 + 36
	w := FontSize*9 + 24
	dc := gg.NewContext(h, w)
	dc.SetRGB(1, 1, 1)
	dc.Clear()
	dc.SetRGB(0, 0, 0)
	fontdata, err := file.GetLazyData(text.GNUUnifontFontFile, control.Md5File, true)
	if err != nil {
		// TODO: err solve
		panic(err)
	}
	if err := dc.ParseFontFace(fontdata, FontSize); err != nil {
		return "", err
	}
	lines := strings.Split(boardString, "\n")
	if len(lines) != 9 {
		lines = make([]string, 9)
		lines[0] = "ERROR [500]"
		lines[1] = "程序内部错误"
		lines[2] = "棋盘字符串不合法"
		lines[3] = "请反馈开发者修复"
	}
	for i := 0; i < 9; i++ {
		dc.DrawString(lines[i], 18, float64(FontSize*(i+1)))
	}
	imgBuffer := bytes.NewBuffer([]byte{})
	if err := dc.EncodePNG(imgBuffer); err != nil {
		return "", err
	}
	imgData, err := io.ReadAll(imgBuffer)
	if err != nil {
		return "", err
	}
	imgB64 := base64.StdEncoding.EncodeToString(imgData)
	return imgB64, nil
}

// generateBoardSVG 生成棋盘 SVG 图片
func generateBoardSVG(svgFilePath, fenStr string, gameTurn chess.Color, sqs ...chess.Square) error {
	os.Remove(svgFilePath)
	f, err := os.Create(svgFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	pos := &chess.Position{}
	if err := pos.UnmarshalText(binary.StringToBytes(fenStr)); err != nil {
		return err
	}
	yellow := color.RGBA{255, 255, 0, 1}
	mark := image.MarkSquares(yellow, sqs...)
	board := pos.Board()
	fromBlack := image.Perspective(gameTurn)
	return image.SVG(f, board, fromBlack, mark)
}

// getChessString 获取 PGN 字符串
func getChessString(room chessRoom) string {
	game := room.chessGame
	dataString := fmt.Sprintf("[Date \"%s\"]\n", time.Now().Format("2006-01-02"))
	whiteString := fmt.Sprintf("[White \"%s\"]\n", room.whiteName)
	blackString := fmt.Sprintf("[Black \"%s\"]\n", room.blackName)
	chessString := game.String()

	return dataString + whiteString + blackString + chessString
}

// getELORate 获取玩家的 ELO 等级分
func getELORate(whiteUin, blackUin int64, dbService *chessDBService) (whiteRate int, blackRate int, err error) {
	whiteRate, err = dbService.getELORateByUin(whiteUin)
	if err != nil {
		return
	}
	blackRate, err = dbService.getELORateByUin(blackUin)
	if err != nil {
		return
	}
	return
}

// inkscapeExists 判断 inkscape 是否存在
func inkscapeExists() bool {
	_, err := exec.LookPath("inkscape")
	return err == nil
}

// isAprilFoolsDay 判断当前时间是否为愚人节期间
func isAprilFoolsDay() bool {
	now := time.Now()
	return now.Month() == 4 && now.Day() == 1
}
