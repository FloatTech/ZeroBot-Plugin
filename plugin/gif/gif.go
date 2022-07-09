package gif

import (
	"image"
	"image/color"
	"sync"

	"github.com/FloatTech/zbputils/img"
	"github.com/FloatTech/zbputils/img/writer"
	"github.com/fogleman/gg"
)

// Mo 摸
func (cc *context) Mo(value ...string) (string, error) {
	var wg sync.WaitGroup
	var err error
	var m sync.Mutex
	name := cc.usrdir + "摸.gif"
	c := dlrange("mo", 5, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	tou, err := cc.getLogo(0, 0)
	if err != nil {
		return "", err
	}
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 5)
	if err != nil {
		return "", err
	}
	mo := []*image.NRGBA{
		imgs[0].InsertBottom(tou, 80, 80, 32, 32).Im,
		imgs[1].InsertBottom(tou, 70, 90, 42, 22).Im,
		imgs[2].InsertBottom(tou, 75, 85, 37, 27).Im,
		imgs[3].InsertBottom(tou, 85, 75, 27, 37).Im,
		imgs[4].InsertBottom(tou, 90, 70, 22, 42).Im,
	}
	return "file:///" + name, writer.SaveGIF2Path(name, img.MergeGif(1, mo))
}

// Cuo 搓
func (cc *context) Cuo(value ...string) (string, error) {
	var wg sync.WaitGroup
	var err error
	var m sync.Mutex
	name := cc.usrdir + "搓.gif"
	c := dlrange("cuo", 5, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	tou, err := cc.getLogo(110, 110)
	if err != nil {
		return "", err
	}
	m1 := img.Rotate(tou, 72, 0, 0)
	m2 := img.Rotate(tou, 144, 0, 0)
	m3 := img.Rotate(tou, 216, 0, 0)
	m4 := img.Rotate(tou, 288, 0, 0)
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 5)
	if err != nil {
		return "", err
	}
	cuo := []*image.NRGBA{
		imgs[0].InsertBottomC(tou, 0, 0, 75, 130).Im,
		imgs[1].InsertBottomC(m1.Im, 0, 0, 75, 130).Im,
		imgs[2].InsertBottomC(m2.Im, 0, 0, 75, 130).Im,
		imgs[3].InsertBottomC(m3.Im, 0, 0, 75, 130).Im,
		imgs[4].InsertBottomC(m4.Im, 0, 0, 75, 130).Im,
	}
	return "file:///" + name, writer.SaveGIF2Path(name, img.MergeGif(5, cuo))
}

// Qiao 敲
func (cc *context) Qiao(value ...string) (string, error) {
	var wg sync.WaitGroup
	var err error
	var m sync.Mutex
	name := cc.usrdir + "敲.gif"
	c := dlrange("qiao", 2, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	tou, err := cc.getLogo(40, 40)
	if err != nil {
		return "", err
	}
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 2)
	if err != nil {
		return "", err
	}
	qiao := []*image.NRGBA{
		imgs[0].InsertUp(tou, 40, 33, 57, 52).Im,
		imgs[1].InsertUp(tou, 38, 36, 58, 50).Im,
	}
	return "file:///" + name, writer.SaveGIF2Path(name, img.MergeGif(1, qiao))
}

// Chi 吃
func (cc *context) Chi(value ...string) (string, error) {
	var wg sync.WaitGroup
	var err error
	var m sync.Mutex
	name := cc.usrdir + "吃.gif"
	c := dlrange("chi", 3, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	tou, err := cc.getLogo(32, 32)
	if err != nil {
		return "", err
	}
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 3)
	if err != nil {
		return "", err
	}
	chi := []*image.NRGBA{
		imgs[0].InsertBottom(tou, 0, 0, 1, 38).Im,
		imgs[1].InsertBottom(tou, 0, 0, 1, 38).Im,
		imgs[2].InsertBottom(tou, 0, 0, 1, 38).Im,
	}
	return "file:///" + name, writer.SaveGIF2Path(name, img.MergeGif(1, chi))
}

// Ceng 蹭
func (cc *context) Ceng(value ...string) (string, error) {
	var wg sync.WaitGroup
	var err error
	var m sync.Mutex
	name := cc.usrdir + "蹭.gif"
	c := dlrange("ceng", 6, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	tou, err := cc.getLogo(100, 100)
	if err != nil {
		return "", err
	}
	tou2, err := cc.getLogo2(100, 100)
	if err != nil {
		return "", err
	}
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 6)
	if err != nil {
		return "", err
	}
	ceng := []*image.NRGBA{
		imgs[0].InsertUp(tou, 75, 77, 40, 88).InsertUp(tou2, 77, 103, 102, 81).Im,
		imgs[1].InsertUp(tou, 75, 77, 46, 100).InsertUp(img.Rotate(tou2, 10, 62, 127).Im, 0, 0, 92, 40).Im,
		imgs[2].InsertUp(tou, 75, 77, 67, 99).InsertUp(tou2, 76, 117, 90, 8).Im,
		imgs[3].InsertUp(tou, 75, 77, 52, 83).InsertUp(img.Rotate(tou2, -40, 94, 94).Im, 0, 0, 53, -20).Im,
		imgs[4].InsertUp(tou, 75, 77, 56, 110).InsertUp(img.Rotate(tou2, -66, 132, 80).Im, 0, 0, 78, 40).Im,
		imgs[5].InsertUp(tou, 75, 77, 62, 102).InsertUp(tou2, 71, 100, 110, 94).Im,
	}
	return "file:///" + name, writer.SaveGIF2Path(name, img.MergeGif(8, ceng))
}

// Ken 啃
func (cc *context) Ken(value ...string) (string, error) {
	var wg sync.WaitGroup
	var err error
	var m sync.Mutex
	name := cc.usrdir + "啃.gif"
	c := dlrange("ken", 16, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	tou, err := cc.getLogo(100, 100)
	if err != nil {
		return "", err
	}
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 16)
	if err != nil {
		return "", err
	}
	ken := []*image.NRGBA{
		imgs[0].InsertBottom(tou, 90, 90, 105, 150).Im,
		imgs[1].InsertBottom(tou, 90, 83, 96, 172).Im,
		imgs[2].InsertBottom(tou, 90, 90, 106, 148).Im,
		imgs[3].InsertBottom(tou, 88, 88, 97, 167).Im,
		imgs[4].InsertBottom(tou, 90, 85, 89, 179).Im,
		imgs[5].InsertBottom(tou, 90, 90, 106, 151).Im,
		imgs[6].Im,
		imgs[7].Im,
		imgs[8].Im,
		imgs[9].Im,
		imgs[10].Im,
		imgs[11].Im,
		imgs[12].Im,
		imgs[13].Im,
		imgs[14].Im,
		imgs[15].Im,
	}
	return "file:///" + name, writer.SaveGIF2Path(name, img.MergeGif(7, ken))
}

// Pai 拍
func (cc *context) Pai(value ...string) (string, error) {
	var wg sync.WaitGroup
	var err error
	var m sync.Mutex
	name := cc.usrdir + "拍.gif"
	c := dlrange("pai", 2, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	tou, err := cc.getLogo(30, 30)
	if err != nil {
		return "", err
	}
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 2)
	if err != nil {
		return "", err
	}
	pai := []*image.NRGBA{
		imgs[0].InsertUp(tou, 0, 0, 1, 47).Im,
		imgs[1].InsertUp(tou, 0, 0, 1, 67).Im,
	}
	return "file:///" + name, writer.SaveGIF2Path(name, img.MergeGif(1, pai))
}

// Xqe 冲
func (cc *context) Xqe(value ...string) (string, error) {
	var wg sync.WaitGroup
	var err error
	var m sync.Mutex
	name := cc.usrdir + "冲.gif"
	c := dlrange("xqe", 2, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	tou, err := cc.getLogo(0, 0)
	if err != nil {
		return "", err
	}
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 2)
	if err != nil {
		return "", err
	}
	chong := []*image.NRGBA{
		imgs[0].InsertUp(tou, 30, 30, 15, 53).Im,
		imgs[1].InsertUp(tou, 30, 30, 40, 53).Im,
	}
	return "file:///" + name, writer.SaveGIF2Path(name, img.MergeGif(1, chong))
}

// Diu 丢
func (cc *context) Diu(value ...string) (string, error) {
	var wg sync.WaitGroup
	var err error
	var m sync.Mutex
	name := cc.usrdir + "丢.gif"
	c := dlrange("diu", 8, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	tou, err := cc.getLogo(0, 0)
	if err != nil {
		return "", err
	}
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 8)
	if err != nil {
		return "", err
	}
	diu := []*image.NRGBA{
		imgs[0].InsertUp(tou, 32, 32, 108, 36).Im,
		imgs[1].InsertUp(tou, 32, 32, 122, 36).Im,
		imgs[2].Im,
		imgs[3].InsertUp(tou, 123, 123, 19, 129).Im,
		imgs[4].InsertUp(tou, 185, 185, -50, 200).InsertUp(tou, 33, 33, 289, 70).Im,
		imgs[5].InsertUp(tou, 32, 32, 280, 73).Im,
		imgs[6].InsertUp(tou, 35, 35, 259, 31).Im,
		imgs[7].InsertUp(tou, 175, 175, -50, 220).Im,
	}
	return "file:///" + name, writer.SaveGIF2Path(name, img.MergeGif(7, diu))
}

// Kiss 亲
func (cc *context) Kiss(value ...string) (string, error) {
	var wg sync.WaitGroup
	var err error
	var m sync.Mutex
	piclen := 13
	name := cc.usrdir + "Kiss.gif"
	c := dlrange("kiss", piclen, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	wg.Wait()
	tou, err := cc.getLogo(0, 0)
	if err != nil {
		return "", err
	}
	tou2, err := cc.getLogo2(0, 0)
	if err != nil {
		return "", err
	}
	userLocs := [][]int{{58, 90}, {62, 95}, {42, 100}, {50, 100}, {56, 100}, {18, 120}, {28, 110}, {54, 100}, {46, 100}, {60, 100}, {35, 115}, {20, 120}, {40, 96}}
	selfLocs := [][]int{{92, 64}, {135, 40}, {84, 105}, {80, 110}, {155, 82}, {60, 96}, {50, 80}, {98, 55}, {35, 65}, {38, 100}, {70, 80}, {84, 65}, {75, 65}}
	imgs, err := loadFirstFrames(c, piclen)
	if err != nil {
		return "", err
	}
	kiss := make([]*image.NRGBA, piclen)
	for i := 0; i < piclen; i++ {
		kiss[i] = imgs[i].InsertUp(tou, 50, 50, userLocs[i][0], userLocs[i][1]).
			InsertUp(tou2, 40, 40, selfLocs[i][0], selfLocs[i][1]).Im
	}
	return "file:///" + name, writer.SaveGIF2Path(name, img.MergeGif(7, kiss))
}

// Garbage 垃圾 垃圾桶
func (cc *context) Garbage(value ...string) (string, error) {
	var wg sync.WaitGroup
	var err error
	var m sync.Mutex
	piclen := 25
	name := cc.usrdir + "Garbage.gif"
	c := dlrange("garbage", piclen, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	wg.Wait()
	im, err := img.LoadFirstFrame(cc.headimgsdir[0], 79, 79)
	if err != nil {
		return "", err
	}
	locs := [][]int{{39, 40}, {39, 40}, {39, 40}, {39, 30}, {39, 30}, {39, 32}, {39, 32}, {39, 32}, {39, 32}, {39, 32}, {39, 32}, {39, 32}, {39, 32}, {39, 32}, {39, 32}, {39, 30}, {39, 27}, {39, 32}, {37, 49}, {37, 64}, {37, 67}, {37, 67}, {39, 69}, {37, 70}, {37, 70}}
	imgs, err := loadFirstFrames(c, piclen)
	if err != nil {
		return "", err
	}
	garbage := make([]*image.NRGBA, piclen)
	for i := 0; i < piclen; i++ {
		garbage[i] = imgs[i].InsertBottom(im.Im, 0, 0, locs[i][0], locs[i][1]).Im
	}
	return "file:///" + name, writer.SaveGIF2Path(name, img.MergeGif(7, garbage))
}

// Thump 捶
func (cc *context) Thump(value ...string) (string, error) {
	var wg sync.WaitGroup
	var err error
	var m sync.Mutex
	piclen := 4
	name := cc.usrdir + "Thump.gif"
	c := dlrange("thump", piclen, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	wg.Wait()
	im, err := img.LoadFirstFrame(cc.headimgsdir[0], 0, 0)
	if err != nil {
		return "", err
	}
	locs := [][]int{{65, 128, 77, 72}, {67, 128, 73, 72}, {54, 139, 94, 61}, {57, 135, 86, 65}}
	imgs, err := loadFirstFrames(c, piclen)
	if err != nil {
		return "", err
	}
	thump := make([]*image.NRGBA, piclen)
	for i := 0; i < piclen; i++ {
		thump[i] = imgs[i].InsertBottom(im.Im, locs[i][2], locs[i][3], locs[i][0], locs[i][1]).Im
	}
	return "file:///" + name, writer.SaveGIF2Path(name, img.MergeGif(7, thump))
}

// Jiujiu 啾啾
func (cc *context) Jiujiu(value ...string) (string, error) {
	var wg sync.WaitGroup
	var err error
	var m sync.Mutex
	piclen := 8
	name := cc.usrdir + "Jiujiu.gif"
	c := dlrange("jiujiu", piclen, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	wg.Wait()
	im, err := img.LoadFirstFrame(cc.headimgsdir[0], 75, 51)
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, piclen)
	if err != nil {
		return "", err
	}
	jiujiu := make([]*image.NRGBA, piclen)
	for i := 0; i < piclen; i++ {
		jiujiu[i] = imgs[i].InsertBottom(im.Im, 0, 0, 0, 0).Im
	}
	return "file:///" + name, writer.SaveGIF2Path(name, img.MergeGif(7, jiujiu))
}

// Knock 2敲
func (cc *context) Knock(value ...string) (string, error) {
	var wg sync.WaitGroup
	var err error
	var m sync.Mutex
	piclen := 8
	name := cc.usrdir + "Knock.gif"
	c := dlrange("knock", piclen, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	wg.Wait()
	im, err := img.LoadFirstFrame(cc.headimgsdir[0], 0, 0)
	if err != nil {
		return "", err
	}
	locs := [][]int{{60, 308, 210, 195}, {60, 308, 210, 198}, {45, 330, 250, 172}, {58, 320, 218, 180}, {60, 310, 215, 193}, {40, 320, 250, 285}, {48, 308, 226, 192}, {51, 301, 223, 200}}
	imgs, err := loadFirstFrames(c, piclen)
	if err != nil {
		return "", err
	}
	knock := make([]*image.NRGBA, piclen)
	for i := 0; i < piclen; i++ {
		knock[i] = imgs[i].InsertBottom(im.Im, locs[i][2], locs[i][3], locs[i][0], locs[i][1]).Im
	}
	return "file:///" + name, writer.SaveGIF2Path(name, img.MergeGif(7, knock))
}

// 听音乐 ListenMusic
func (cc *context) ListenMusic(value ...string) (string, error) {
	var wg sync.WaitGroup
	var err error
	var m sync.Mutex
	piclen := 1
	name := cc.usrdir + "ListenMusic.gif"
	c := dlrange("listen_music", piclen, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	wg.Wait()
	face, err := cc.getLogo(0, 0)
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, piclen)
	if err != nil {
		return "", err
	}
	listenmusic := make([]*image.NRGBA, 36)
	for i := 0; i < 36; i++ {
		listenmusic[i] = imgs[0].InsertBottomC(img.Rotate(face, float64(-i*10), 215, 215).Im, 0, 0, 207, 207).Im
	}
	return "file:///" + name, writer.SaveGIF2Path(name, img.MergeGif(7, listenmusic))
}

// LoveYou 永远爱你
func (cc *context) LoveYou(value ...string) (string, error) {
	var wg sync.WaitGroup
	var err error
	var m sync.Mutex
	piclen := 2
	name := cc.usrdir + "LoveYou.gif"
	c := dlrange("love_you", piclen, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	wg.Wait()
	im, err := img.LoadFirstFrame(cc.headimgsdir[0], 0, 0)
	if err != nil {
		return "", err
	}
	locs := [][]int{{68, 65, 70, 70}, {63, 59, 80, 80}}
	imgs, err := loadFirstFrames(c, piclen)
	if err != nil {
		return "", err
	}
	loveyou := make([]*image.NRGBA, piclen)
	for i := 0; i < piclen; i++ {
		loveyou[i] = imgs[i].InsertBottom(im.Im, locs[i][2], locs[i][3], locs[i][0], locs[i][1]).Im
	}
	return "file:///" + name, writer.SaveGIF2Path(name, img.MergeGif(7, loveyou))
}

// Pat 2拍
func (cc *context) Pat(value ...string) (string, error) {
	var wg sync.WaitGroup
	var err error
	var m sync.Mutex
	piclen := 10
	name := cc.usrdir + "Pat.gif"
	c := dlrange("pat", piclen, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	wg.Wait()
	im, err := img.LoadFirstFrame(cc.headimgsdir[0], 0, 0)
	if err != nil {
		return "", err
	}
	locs := [][]int{{11, 73, 106, 100}, {8, 79, 112, 96}}
	imgs, err := loadFirstFrames(c, piclen)
	if err != nil {
		return "", err
	}
	p := make([]*image.NRGBA, piclen)
	for i := 0; i < piclen; i++ {
		if i == 2 {
			p[i] = imgs[i].InsertBottom(im.Im, locs[1][2], locs[1][3], locs[1][0], locs[1][1]).Im
		} else {
			p[i] = imgs[i].InsertBottom(im.Im, locs[0][2], locs[0][3], locs[0][0], locs[0][1]).Im
		}
	}
	seq := []int{0, 1, 2, 3, 1, 2, 3, 0, 1, 2, 3, 0, 0, 1, 2, 3, 0, 0, 0, 0, 4, 5, 5, 5, 6, 7, 8, 9}
	pat := make([]*image.NRGBA, len(seq))
	for i := 0; i < len(pat); i++ {
		pat[i] = p[seq[i]]
	}
	return "file:///" + name, writer.SaveGIF2Path(name, img.MergeGif(7, pat))
}

// JackUp 顶
func (cc *context) JackUp(value ...string) (string, error) {
	var wg sync.WaitGroup
	var err error
	var m sync.Mutex
	piclen := 23
	name := cc.usrdir + "JackUp.gif"
	c := dlrange("play", piclen, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	wg.Wait()
	im, err := img.LoadFirstFrame(cc.headimgsdir[0], 0, 0)
	if err != nil {
		return "", err
	}
	locs := [][]int{{180, 60, 100, 100}, {184, 75, 100, 100}, {183, 98, 100, 100}, {179, 118, 110, 100}, {156, 194, 150, 48}, {178, 136, 122, 69}, {175, 66, 122, 85}, {170, 42, 130, 96}, {175, 34, 118, 95}, {179, 35, 110, 93}, {180, 54, 102, 93}, {183, 58, 97, 92}, {174, 35, 120, 94}, {179, 35, 109, 93}, {181, 54, 101, 92}, {182, 59, 98, 92}, {183, 71, 90, 96}, {180, 131, 92, 101}}
	imgs, err := loadFirstFrames(c, piclen)
	if err != nil {
		return "", err
	}
	p := make([]*image.NRGBA, piclen)
	for i := 0; i < piclen; i++ {
		if i < len(locs) {
			p[i] = imgs[i].InsertBottom(im.Im, locs[i][2], locs[i][3], locs[i][0], locs[i][1]).Im
		} else {
			p[i] = imgs[i].Im
		}
	}
	play := make([]*image.NRGBA, 0, 16)
	play = append(play, p[0:12]...)
	play = append(play, p[0:12]...)
	play = append(play, p[0:8]...)
	play = append(play, p[12:18]...)
	play = append(play, p[18:23]...)
	return "file:///" + name, writer.SaveGIF2Path(name, img.MergeGif(7, play))
}

// Pound 捣
func (cc *context) Pound(value ...string) (string, error) {
	var wg sync.WaitGroup
	var err error
	var m sync.Mutex
	piclen := 8
	name := cc.usrdir + "Pound.gif"
	c := dlrange("pound", piclen, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	wg.Wait()
	im, err := img.LoadFirstFrame(cc.headimgsdir[0], 0, 0)
	if err != nil {
		return "", err
	}
	locs := [][]int{{135, 240, 138, 47}, {135, 240, 138, 47}, {150, 190, 105, 95}, {150, 190, 105, 95}, {148, 188, 106, 98}, {146, 196, 110, 88}, {145, 223, 112, 61}, {145, 223, 112, 61}}
	imgs, err := loadFirstFrames(c, piclen)
	if err != nil {
		return "", err
	}
	pound := make([]*image.NRGBA, piclen)
	for i := 0; i < piclen; i++ {
		pound[i] = imgs[i].InsertBottom(im.Im, locs[i][2], locs[i][3], locs[i][0], locs[i][1]).Im
	}
	return "file:///" + name, writer.SaveGIF2Path(name, img.MergeGif(7, pound))
}

// Punch 打拳
func (cc *context) Punch(value ...string) (string, error) {
	var wg sync.WaitGroup
	var err error
	var m sync.Mutex
	piclen := 13
	name := cc.usrdir + "Punch.gif"
	c := dlrange("punch", piclen, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	wg.Wait()
	im, err := img.LoadFirstFrame(cc.headimgsdir[0], 260, 260)
	if err != nil {
		return "", err
	}
	locs := [][]int{{-50, 20}, {-40, 10}, {-30, 0}, {-20, -10}, {-10, -10}, {0, 0}, {10, 10}, {20, 20}, {10, 10}, {0, 0}, {-10, -10}, {10, 0}, {-30, 10}}
	imgs, err := loadFirstFrames(c, piclen)
	if err != nil {
		return "", err
	}
	punch := make([]*image.NRGBA, piclen)
	for i := 0; i < piclen; i++ {
		punch[i] = imgs[i].InsertBottom(im.Im, 0, 0, locs[i][0], locs[i][1]-15).Im
	}
	return "file:///" + name, writer.SaveGIF2Path(name, img.MergeGif(7, punch))
}

// Roll 滚
func (cc *context) Roll(value ...string) (string, error) {
	var wg sync.WaitGroup
	var err error
	var m sync.Mutex
	piclen := 8
	name := cc.usrdir + "roll.gif"
	c := dlrange("roll", piclen, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	wg.Wait()
	im, err := img.LoadFirstFrame(cc.headimgsdir[0], 210, 210)
	if err != nil {
		return "", err
	}
	locs := [][]int{{87, 77, 0}, {96, 85, -45}, {92, 79, -90}, {92, 78, -135}, {92, 75, -180}, {92, 75, -225}, {93, 76, -270}, {90, 80, -315}}
	imgs, err := loadFirstFrames(c, piclen)
	if err != nil {
		return "", err
	}
	roll := make([]*image.NRGBA, piclen)
	for i := 0; i < piclen; i++ {
		roll[i] = imgs[i].InsertBottomC(img.Rotate(im.Im, float64(locs[i][2]), 0, 0).Im, 0, 0, locs[i][0]+105, locs[i][1]+105).Im
	}
	return "file:///" + name, writer.SaveGIF2Path(name, img.MergeGif(7, roll))
}

// Suck 吸 嗦
func (cc *context) Suck(value ...string) (string, error) {
	var wg sync.WaitGroup
	var err error
	var m sync.Mutex
	piclen := 12
	name := cc.usrdir + "Suck.gif"
	c := dlrange("suck", piclen, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	wg.Wait()
	im, err := img.LoadFirstFrame(cc.headimgsdir[0], 0, 0)
	if err != nil {
		return "", err
	}
	locs := [][]int{{82, 100, 130, 119}, {82, 94, 126, 125}, {82, 120, 128, 99}, {81, 164, 132, 55}, {79, 163, 132, 55}, {82, 140, 127, 79}, {83, 152, 125, 67}, {75, 157, 140, 62}, {72, 165, 144, 54}, {80, 132, 128, 87}, {81, 127, 127, 92}, {79, 111, 132, 108}}
	imgs, err := loadFirstFrames(c, piclen)
	if err != nil {
		return "", err
	}
	suck := make([]*image.NRGBA, piclen)
	for i := 0; i < piclen; i++ {
		suck[i] = imgs[i].InsertBottom(im.Im, locs[i][2], locs[i][3], locs[i][0], locs[i][1]).Im
	}
	return "file:///" + name, writer.SaveGIF2Path(name, img.MergeGif(7, suck))
}

// Hammer 锤
func (cc *context) Hammer(value ...string) (string, error) {
	var wg sync.WaitGroup
	var err error
	var m sync.Mutex
	piclen := 7
	name := cc.usrdir + "Hammer.gif"
	c := dlrange("hammer", piclen, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	wg.Wait()
	im, err := img.LoadFirstFrame(cc.headimgsdir[0], 0, 0)
	if err != nil {
		return "", err
	}
	locs := [][]int{{62, 143, 158, 113}, {52, 177, 173, 105}, {42, 192, 192, 92}, {46, 182, 184, 100}, {54, 169, 174, 110}, {69, 128, 144, 135}, {65, 130, 152, 124}}
	imgs, err := loadFirstFrames(c, piclen)
	if err != nil {
		return "", err
	}
	hammer := make([]*image.NRGBA, piclen)
	for i := 0; i < piclen; i++ {
		hammer[i] = imgs[i].InsertBottom(im.Im, locs[i][2], locs[i][3], locs[i][0], locs[i][1]).Im
	}
	return "file:///" + name, writer.SaveGIF2Path(name, img.MergeGif(7, hammer))
}

// Tightly 紧贴 紧紧贴着
func (cc *context) Tightly(value ...string) (string, error) {
	var wg sync.WaitGroup
	var err error
	var m sync.Mutex
	piclen := 20
	name := cc.usrdir + "Tightly.gif"
	c := dlrange("tightly", piclen, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	wg.Wait()
	im, err := img.LoadFirstFrame(cc.headimgsdir[0], 0, 0)
	if err != nil {
		return "", err
	}
	locs := [][]int{{39, 169, 267, 141}, {40, 167, 264, 143}, {38, 174, 270, 135}, {40, 167, 264, 143}, {38, 174, 270, 135}, {40, 167, 264, 143}, {38, 174, 270, 135}, {40, 167, 264, 143}, {38, 174, 270, 135}, {28, 176, 293, 134}, {5, 215, 333, 96}, {10, 210, 321, 102}, {3, 210, 330, 104}, {4, 210, 328, 102}, {4, 212, 328, 100}, {4, 212, 328, 100}, {4, 212, 328, 100}, {4, 212, 328, 100}, {4, 212, 328, 100}, {29, 195, 285, 120}}
	imgs, err := loadFirstFrames(c, piclen)
	if err != nil {
		return "", err
	}
	tightly := make([]*image.NRGBA, piclen)
	for i := 0; i < piclen; i++ {
		tightly[i] = imgs[i].InsertBottom(im.Im, locs[i][2], locs[i][3], locs[i][0], locs[i][1]).Im
	}
	return "file:///" + name, writer.SaveGIF2Path(name, img.MergeGif(7, tightly))
}

// Turn 转
func (cc *context) Turn(value ...string) (string, error) {
	name := cc.usrdir + "Turn.gif"
	face, err := cc.getLogo(0, 0)
	if err != nil {
		return "", err
	}
	canvas := gg.NewContext(250, 250)
	canvas.SetColor(color.White)
	canvas.DrawRectangle(0, 0, 250, 250)
	canvas.Fill()
	turn := make([]*image.NRGBA, 36)
	for i := 0; i < 36; i++ {
		turn[i] = img.Size(canvas.Image(), 0, 0).InsertUpC(img.Rotate(face, float64(10*i), 250, 250).Im, 0, 0, 125, 125).Im
	}
	return "file:///" + name, writer.SaveGIF2Path(name, img.MergeGif(7, turn))
}
