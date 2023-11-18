package chess

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image/color"
	"strconv"
	"strings"
	"time"

	"github.com/FloatTech/floatbox/binary"
	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/img/text"

	"github.com/RomiChan/syncx"
	"github.com/jinzhu/gorm"
	resvg "github.com/kanrichan/resvg-go"
	"github.com/notnil/chess"
	cimage "github.com/notnil/chess/image"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const eloDefault = 500

var (
	chessRoomMap syncx.Map[int64, *chessRoom]
	errNotExist  = errors.New("对局不存在, 发送「下棋」或「chess」可创建对局。")
)

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
func game(groupCode, senderUin int64, senderName string) (message.Message, error) {
	return createGame(false, groupCode, senderUin, senderName)
}

// blindfold 盲棋
func blindfold(groupCode, senderUin int64, senderName string) (message.Message, error) {
	return createGame(true, groupCode, senderUin, senderName)
}

// abort 中断对局
func abort(groupCode int64) (message.Message, error) {
	if room, ok := chessRoomMap.Load(groupCode); ok {
		return abortGame(*room, groupCode, "对局已被管理员中断, 游戏结束。")
	}
	return nil, errNotExist
}

// draw 和棋
func draw(groupCode, senderUin int64) (msg message.Message, err error) {
	msg = message.Message{message.At(senderUin)}
	// 检查对局是否存在
	room, ok := chessRoomMap.Load(groupCode)
	if !ok {
		return nil, errNotExist
	}
	// 检查消息发送者是否为对局中的玩家
	if senderUin != room.whitePlayer && senderUin != room.blackPlayer {
		return
	}
	// 处理和棋逻辑
	room.lastMoveTime = time.Now().Unix()
	if room.drawPlayer == 0 {
		room.drawPlayer = senderUin
		chessRoomMap.Store(groupCode, room)
		msg = append(msg, message.Text("请求和棋, 发送「和棋」或「draw」接受和棋。走棋视为拒绝和棋。"))
		return
	}
	if room.drawPlayer == senderUin {
		return
	}
	err = room.chessGame.Draw(chess.DrawOffer)
	if err != nil {
		return
	}
	chessString := getChessString(*room)
	eloString := ""
	if len(room.chessGame.Moves()) > 4 {
		// 若走子次数超过 4 认为是有效对局, 存入数据库
		dbService := newDBService()
		if err = dbService.createPGN(chessString, room.whitePlayer, room.blackPlayer, room.whiteName, room.blackName); err != nil {
			return
		}
		whiteScore, blackScore := 0.5, 0.5
		eloString, err = getELOString(*room, whiteScore, blackScore)
		if err != nil {
			return
		}
	}
	msg = append(msg, message.Text("接受和棋, 游戏结束。\n", eloString, chessString))
	chessRoomMap.Delete(groupCode)
	return
}

// resign 认输
func resign(groupCode, senderUin int64) (msg message.Message, err error) {
	msg = message.Message{message.At(senderUin)}
	// 检查对局是否存在
	room, ok := chessRoomMap.Load(groupCode)
	if !ok {
		return nil, errNotExist
	}
	// 检查是否是当前游戏玩家
	if senderUin != room.whitePlayer && senderUin != room.blackPlayer {
		return
	}
	// 如果对局未建立, 中断对局
	if room.whitePlayer == 0 || room.blackPlayer == 0 {
		chessRoomMap.Delete(groupCode)
		msg = append(msg, message.Text("对局结束"))
		return
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
		// 若走子次数超过 4 认为是有效对局, 存入数据库
		dbService := newDBService()
		if err = dbService.createPGN(chessString, room.whitePlayer, room.blackPlayer, room.whiteName, room.blackName); err != nil {
			return
		}
		whiteScore, blackScore := 1.0, 1.0
		if resignColor == chess.White {
			whiteScore = 0.0
		} else {
			blackScore = 0.0
		}
		eloString, err = getELOString(*room, whiteScore, blackScore)
		if err != nil {
			return
		}
	}
	msg = append(msg, message.Text("认输, 游戏结束。\n", eloString, chessString))
	if isAprilFoolsDay() {
		msg = append(msg, message.Text("对手认输, 游戏结束, 你胜利了。\n", eloString, chessString))
	}
	chessRoomMap.Delete(groupCode)
	return
}

// play 走棋
func play(groupCode, senderUin int64, moveStr string) (msg message.Message, err error) {
	msg = message.Message{message.At(senderUin)}
	// 检查对局是否存在
	room, ok := chessRoomMap.Load(groupCode)
	if !ok {
		return nil, errNotExist
	}
	// 不是对局中的玩家, 忽略消息
	if (senderUin != room.whitePlayer) && (senderUin != room.blackPlayer) && !isAprilFoolsDay() {
		return
	}
	// 对局未建立
	if (room.whitePlayer == 0) || (room.blackPlayer == 0) {
		msg = append(msg, message.Text("请等候其他玩家加入游戏。"))
		return
	}
	// 需要对手走棋
	if ((senderUin == room.whitePlayer) && (room.chessGame.Position().Turn() != chess.White)) || ((senderUin == room.blackPlayer) && (room.chessGame.Position().Turn() != chess.Black)) {
		msg = append(msg, message.Text("请等待对手走棋。"))
		return
	}
	room.lastMoveTime = time.Now().Unix()
	// 走棋
	if err = room.chessGame.MoveStr(moveStr); err != nil {
		// 指令错误时检查
		if !room.isBlindfold {
			// 未开启盲棋, 提示指令错误
			msg = append(msg, message.Text("移动「", moveStr, "」违规, 请检查, 格式请参考「代数记谱法」(Algebraic notation)。"))
			return
		}
		// 开启盲棋, 判断违例情况
		var currentPlayerColor chess.Color
		if senderUin == room.whitePlayer {
			currentPlayerColor = chess.White
		} else {
			currentPlayerColor = chess.Black
		}
		// 第一次违例, 提示
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
			msg = append(msg, message.Text("移动「", moveStr, "」违规, 再次违规会立即判负。"))
			return
		}
		// 出现多次违例, 判负
		room.chessGame.Resign(currentPlayerColor)
		chessString := getChessString(*room)
		msg = append(msg, message.Text("违规两次,游戏结束。\n", chessString))

		chessRoomMap.Delete(groupCode)
		return
	}
	// 走子之后, 视为拒绝和棋
	if room.drawPlayer != 0 {
		room.drawPlayer = 0
		chessRoomMap.Store(groupCode, room)
	}
	// 生成棋盘图片
	var boardImgEle message.MessageSegment
	if !room.isBlindfold {
		boardImgEle, err = getBoardElement(groupCode)
		if err != nil {
			return
		}
	}
	// 检查游戏是否结束
	if room.chessGame.Method() != chess.NoMethod {
		whiteScore, blackScore := 0.5, 0.5
		var msgBuilder strings.Builder
		msgBuilder.WriteString("游戏结束, ")
		switch room.chessGame.Method() {
		case chess.FivefoldRepetition:
			msgBuilder.WriteString("和棋, 因为五次重复走子。\n")
		case chess.SeventyFiveMoveRule:
			msgBuilder.WriteString("和棋, 因为七十五步规则。\n")
		case chess.InsufficientMaterial:
			msgBuilder.WriteString("和棋, 因为不可能将死。\n")
		case chess.Stalemate:
			msgBuilder.WriteString("和棋, 因为逼和（无子可动和棋）。\n")
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
			msgBuilder.WriteString("胜利, 因为将杀。\n")
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
			// 若走子次数超过 4 认为是有效对局, 存入数据库
			dbService := newDBService()
			if err = dbService.createPGN(chessString, room.whitePlayer, room.blackPlayer, room.whiteName, room.blackName); err != nil {
				return
			}
			// 仅有效对局才会计算等级分
			eloString, err = getELOString(*room, whiteScore, blackScore)
			if err != nil {
				return
			}
		}
		msgBuilder.WriteString(eloString)
		msgBuilder.WriteString(chessString)
		msg = append(msg, message.Text(msgBuilder.String()))
		if !room.isBlindfold {
			msg = append(msg, boardImgEle)
		}

		chessRoomMap.Delete(groupCode)
		return
	}
	// 提示玩家继续游戏
	var currentPlayer int64
	if room.chessGame.Position().Turn() == chess.White {
		currentPlayer = room.whitePlayer
	} else {
		currentPlayer = room.blackPlayer
	}
	msg = message.Message{message.At(currentPlayer), message.Text("对手已走子, 游戏继续。"), boardImgEle}
	return
}

// rate 获取等级分
func rate(senderUin int64, senderName string) (msg message.Message, err error) {
	rate := 0
	dbService := newDBService()
	rate, err = dbService.getELORateByUin(senderUin)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			err = errors.New("无法获取等级分信息。")
			return
		}
		err = errors.New("没有查找到等级分信息, 请至少进行一局对局。")
	}
	msg = append(msg, message.Text("玩家「", senderName, "」目前的等级分: ", rate))
	return
}

// cleanUserRate 清空用户等级分
func cleanUserRate(senderUin int64) (msg message.Message, err error) {
	dbService := newDBService()
	err = dbService.cleanELOByUin(senderUin)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			err = errors.New("无法清空等级分。")
			return
		}
		err = errors.New("没有查找到等级分信息, 请检查用户 uid 是否正确。")
	}
	msg = append(msg, message.Text("已清空用户「", senderUin, "」的等级分。"))
	return
}

// createGame 创建游戏
func createGame(isBlindfold bool, groupCode, senderUin int64, senderName string) (msg message.Message, err error) {
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
		text := "已创建新的对局, 发送「下棋」或「chess」可加入对局。"
		if isBlindfold {
			text = "已创建新的盲棋对局, 发送「盲棋」或「blind」可加入对局。"
		}
		msg = append(msg, message.Text(text))
		return
	}
	msg = message.Message{message.At(senderUin)}
	if room.blackPlayer != 0 {
		// 检测对局是否已存在超过 6 小时
		if (time.Now().Unix() - room.lastMoveTime) > 21600 {
			msg, err = abortGame(*room, groupCode, "对局已存在超过 6 小时, 游戏结束。")
			msg = append(msg, message.Text("\n\n已有对局已被中断, 如需创建新对局请重新发送指令。"))
			msg = append(msg, message.At(senderUin))
			return
		}
		// 对局在进行
		msg = append(msg, message.Text("对局已在进行中, 无法创建或加入对局, 当前对局玩家为: "))
		if room.whitePlayer != 0 {
			msg = append(msg, message.At(room.whitePlayer))
		}
		if room.blackPlayer != 0 {
			msg = append(msg, message.At(room.blackPlayer))
		}
		msg = append(msg, message.Text(", 群主或管理员发送「中断」或「abort」可中断对局(自动判和)。"))
		return
	}
	if senderUin == room.whitePlayer {
		msg = append(msg, message.Text("请等候其他玩家加入游戏。"))
		return
	}
	if room.isBlindfold && !isBlindfold {
		msg = append(msg, message.Text("已创建盲棋对局, 请加入或等待盲棋对局结束之后创建普通对局。"))
		return
	}
	if !room.isBlindfold && isBlindfold {
		msg = append(msg, message.Text("已创建普通对局, 请加入或等待普通对局结束之后创建盲棋对局。"))
		return
	}
	room.blackPlayer = senderUin
	room.blackName = senderName
	chessRoomMap.Store(groupCode, room)
	var boardImgEle message.MessageSegment
	if !room.isBlindfold {
		boardImgEle, err = getBoardElement(groupCode)
		if err != nil {
			return
		}
	}
	msg = append(msg, message.Text("黑棋已加入对局, 请白方下棋。"), message.At(room.whitePlayer))
	if !isBlindfold {
		msg = append(msg, boardImgEle)
	}
	return
}

// abortGame 中断游戏
func abortGame(room chessRoom, groupCode int64, hint string) (message.Message, error) {
	var msg message.Message
	err := room.chessGame.Draw(chess.DrawOffer)
	if err != nil {
		return nil, err
	}
	chessString := getChessString(room)
	if len(room.chessGame.Moves()) > 4 {
		dbService := newDBService()
		if err := dbService.createPGN(chessString, room.whitePlayer, room.blackPlayer, room.whiteName, room.blackName); err != nil {
			return nil, err
		}
	}

	chessRoomMap.Delete(groupCode)
	msg = append(msg, message.Text(hint))
	if room.whitePlayer != 0 {
		msg = append(msg, message.At(room.whitePlayer))
	}
	if room.blackPlayer != 0 {
		msg = append(msg, message.At(room.blackPlayer))
	}
	msg = append(msg, message.Text("\n\n"+chessString))
	return msg, nil
}

// getBoardElement 获取棋盘图片的消息内容
func getBoardElement(groupCode int64) (imgMsg message.MessageSegment, err error) {
	fontdata, err := file.GetLazyData(text.GNUUnifontFontFile, control.Md5File, true)
	if err != nil {
		return
	}
	room, ok := chessRoomMap.Load(groupCode)
	if !ok {
		return imgMsg, errNotExist
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
	buf := bytes.NewBuffer([]byte{})
	fenStr := room.chessGame.FEN()
	gameTurn := room.chessGame.Position().Turn()
	pos := &chess.Position{}
	if err = pos.UnmarshalText(binary.StringToBytes(fenStr)); err != nil {
		return
	}
	yellow := color.RGBA{255, 255, 0, 1}
	mark := cimage.MarkSquares(yellow, highlightSquare...)
	board := pos.Board()
	fromBlack := cimage.Perspective(gameTurn)
	err = cimage.SVG(buf, board, fromBlack, mark)
	if err != nil {
		return
	}

	worker, err := resvg.NewDefaultWorker(context.Background())
	if err != nil {
		return
	}
	defer worker.Close()

	tree, err := worker.NewTreeFromData(buf.Bytes(), &resvg.Options{
		Dpi:        96,
		FontFamily: "Unifont",
		FontSize:   24.0,
	})
	if err != nil {
		return
	}
	defer tree.Close()

	fontdb, err := worker.NewFontDBDefault()
	if err != nil {
		return
	}
	defer fontdb.Close()

	err = fontdb.LoadFontData(fontdata)
	if err != nil {
		return
	}

	err = tree.ConvertText(fontdb)
	if err != nil {
		return
	}

	pixmap, err := worker.NewPixmap(720, 720)
	if err != nil {
		return
	}
	defer pixmap.Close()

	err = tree.Render(resvg.TransformFromScale(2, 2), pixmap)
	if err != nil {
		return
	}

	out, err := pixmap.EncodePNG()
	if err != nil {
		return
	}

	imgMsg = message.ImageBytes(out)
	return imgMsg, nil
}

// getELOString 获得玩家等级分的文本内容
func getELOString(room chessRoom, whiteScore, blackScore float64) (string, error) {
	if room.whitePlayer == 0 || room.blackPlayer == 0 {
		return "", nil
	}
	var msgBuilder strings.Builder
	msgBuilder.WriteString("玩家等级分: \n")
	dbService := newDBService()
	if err := updateELORate(room.whitePlayer, room.blackPlayer, room.whiteName, room.blackName, whiteScore, blackScore, dbService); err != nil {
		return "", err
	}
	whiteRate, blackRate, err := getELORate(room.whitePlayer, room.blackPlayer, dbService)
	if err != nil {
		return "", err
	}
	msgBuilder.WriteString(room.whiteName)
	msgBuilder.WriteString(": ")
	msgBuilder.WriteString(strconv.Itoa(whiteRate))
	msgBuilder.WriteString("\n")
	msgBuilder.WriteString(room.blackName)
	msgBuilder.WriteString(": ")
	msgBuilder.WriteString(strconv.Itoa(blackRate))
	msgBuilder.WriteString("\n\n")
	return msgBuilder.String(), nil
}

// getRankingString 获取等级分排行榜的文本内容
func getRanking() (message.Message, error) {
	dbService := newDBService()
	eloList, err := dbService.getHighestRateList()
	if err != nil {
		return nil, err
	}
	var msgBuilder strings.Builder
	msgBuilder.WriteString("当前等级分排行榜: \n\n")
	for _, elo := range eloList {
		msgBuilder.WriteString(elo.Name)
		msgBuilder.WriteString(": ")
		msgBuilder.WriteString(strconv.Itoa(elo.Rate))
		msgBuilder.WriteString("\n")
	}
	return message.Message{message.Text(msgBuilder.String())}, nil
}

// updateELORate 更新 elo 等级分
// 当数据库中没有玩家的等级分信息时, 自动新建一条记录
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

// isAprilFoolsDay 判断当前时间是否为愚人节期间
func isAprilFoolsDay() bool {
	now := time.Now()
	return now.Month() == 4 && now.Day() == 1
}
