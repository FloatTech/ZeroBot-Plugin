package gif

import (
	"errors"
	"image"
	"image/color"
	"sync"

	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/gg"
	"github.com/FloatTech/imgfactory"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/img/text"
)

// mo 摸
func mo(cc *context, value ...string) (string, error) {
	_ = value
	var wg sync.WaitGroup
	var errwg error
	var m sync.Mutex
	// name := cc.usrdir + "摸.gif"
	c := dlrange("mo", 5, &wg, func(e error) {
		m.Lock()
		errwg = e
		m.Unlock()
	})
	tou, err := cc.getLogo(0, 0)
	if err != nil {
		return "", err
	}
	wg.Wait()
	if errwg != nil {
		return "", errwg
	}
	imgs, err := loadFirstFrames(c, 5)
	if err != nil {
		return "", err
	}
	mo := []*image.NRGBA{
		imgs[0].InsertBottom(tou, 80, 80, 32, 32).Image(),
		imgs[1].InsertBottom(tou, 70, 90, 42, 22).Image(),
		imgs[2].InsertBottom(tou, 75, 85, 37, 27).Image(),
		imgs[3].InsertBottom(tou, 85, 75, 27, 37).Image(),
		imgs[4].InsertBottom(tou, 90, 70, 22, 42).Image(),
	}
	g := imgfactory.MergeGif(1, mo)
	return imgfactory.GIF2Base64(g)
}

// cuo 搓
func cuo(cc *context, value ...string) (string, error) {
	_ = value
	var wg sync.WaitGroup
	var errwg error
	var m sync.Mutex
	// name := cc.usrdir + "搓.gif"
	c := dlrange("cuo", 5, &wg, func(e error) {
		m.Lock()
		errwg = e
		m.Unlock()
	})
	tou, err := cc.getLogo(110, 110)
	if err != nil {
		return "", err
	}
	m1 := imgfactory.Rotate(tou, 72, 0, 0)
	m2 := imgfactory.Rotate(tou, 144, 0, 0)
	m3 := imgfactory.Rotate(tou, 216, 0, 0)
	m4 := imgfactory.Rotate(tou, 288, 0, 0)
	wg.Wait()
	if errwg != nil {
		return "", errwg
	}
	imgs, err := loadFirstFrames(c, 5)
	if err != nil {
		return "", err
	}
	cuo := []*image.NRGBA{
		imgs[0].InsertBottomC(tou, 0, 0, 75, 130).Image(),
		imgs[1].InsertBottomC(m1.Image(), 0, 0, 75, 130).Image(),
		imgs[2].InsertBottomC(m2.Image(), 0, 0, 75, 130).Image(),
		imgs[3].InsertBottomC(m3.Image(), 0, 0, 75, 130).Image(),
		imgs[4].InsertBottomC(m4.Image(), 0, 0, 75, 130).Image(),
	}
	g := imgfactory.MergeGif(5, cuo)
	return imgfactory.GIF2Base64(g)
}

// qiao 敲
func qiao(cc *context, value ...string) (string, error) {
	_ = value
	var wg sync.WaitGroup
	var errwg error
	var m sync.Mutex
	// name := cc.usrdir + "敲.gif"
	c := dlrange("qiao", 2, &wg, func(e error) {
		m.Lock()
		errwg = e
		m.Unlock()
	})
	tou, err := cc.getLogo(40, 40)
	if err != nil {
		return "", err
	}
	wg.Wait()
	if errwg != nil {
		return "", errwg
	}
	imgs, err := loadFirstFrames(c, 2)
	if err != nil {
		return "", err
	}
	qiao := []*image.NRGBA{
		imgs[0].InsertUp(tou, 40, 33, 57, 52).Image(),
		imgs[1].InsertUp(tou, 38, 36, 58, 50).Image(),
	}
	g := imgfactory.MergeGif(1, qiao)
	return imgfactory.GIF2Base64(g)
}

// chi 吃
func chi(cc *context, value ...string) (string, error) {
	_ = value
	var wg sync.WaitGroup
	var errwg error
	var m sync.Mutex
	// name := cc.usrdir + "吃.gif"
	c := dlrange("chi", 3, &wg, func(e error) {
		m.Lock()
		errwg = e
		m.Unlock()
	})
	tou, err := cc.getLogo(32, 32)
	if err != nil {
		return "", err
	}
	wg.Wait()
	if errwg != nil {
		return "", errwg
	}
	imgs, err := loadFirstFrames(c, 3)
	if err != nil {
		return "", err
	}
	chi := []*image.NRGBA{
		imgs[0].InsertBottom(tou, 0, 0, 1, 38).Image(),
		imgs[1].InsertBottom(tou, 0, 0, 1, 38).Image(),
		imgs[2].InsertBottom(tou, 0, 0, 1, 38).Image(),
	}
	g := imgfactory.MergeGif(1, chi)
	return imgfactory.GIF2Base64(g)
}

// ceng 蹭
func ceng(cc *context, value ...string) (string, error) {
	_ = value
	var wg sync.WaitGroup
	var errwg error
	var m sync.Mutex
	// name := cc.usrdir + "蹭.gif"
	c := dlrange("ceng", 6, &wg, func(e error) {
		m.Lock()
		errwg = e
		m.Unlock()
	})
	tou, err := cc.getLogo(100, 100)
	if err != nil {
		return "", err
	}
	tou2, err := cc.getLogo2(100, 100)
	if err != nil {
		return "", err
	}
	wg.Wait()
	if errwg != nil {
		return "", errwg
	}
	imgs, err := loadFirstFrames(c, 6)
	if err != nil {
		return "", err
	}
	ceng := []*image.NRGBA{
		imgs[0].InsertUp(tou, 75, 77, 40, 88).InsertUp(tou2, 77, 103, 102, 81).Image(),
		imgs[1].InsertUp(tou, 75, 77, 46, 100).InsertUp(imgfactory.Rotate(tou2, 10, 62, 127).Image(), 0, 0, 92, 40).Image(),
		imgs[2].InsertUp(tou, 75, 77, 67, 99).InsertUp(tou2, 76, 117, 90, 8).Image(),
		imgs[3].InsertUp(tou, 75, 77, 52, 83).InsertUp(imgfactory.Rotate(tou2, -40, 94, 94).Image(), 0, 0, 53, -20).Image(),
		imgs[4].InsertUp(tou, 75, 77, 56, 110).InsertUp(imgfactory.Rotate(tou2, -66, 132, 80).Image(), 0, 0, 78, 40).Image(),
		imgs[5].InsertUp(tou, 75, 77, 62, 102).InsertUp(tou2, 71, 100, 110, 94).Image(),
	}
	g := imgfactory.MergeGif(8, ceng)
	return imgfactory.GIF2Base64(g)
}

// ken 啃
func ken(cc *context, value ...string) (string, error) {
	_ = value
	var wg sync.WaitGroup
	var errwg error
	var m sync.Mutex
	// name := cc.usrdir + "啃.gif"
	c := dlrange("ken", 16, &wg, func(e error) {
		m.Lock()
		errwg = e
		m.Unlock()
	})
	tou, err := cc.getLogo(100, 100)
	if err != nil {
		return "", err
	}
	wg.Wait()
	if errwg != nil {
		return "", errwg
	}
	imgs, err := loadFirstFrames(c, 16)
	if err != nil {
		return "", err
	}
	ken := []*image.NRGBA{
		imgs[0].InsertBottom(tou, 90, 90, 105, 150).Image(),
		imgs[1].InsertBottom(tou, 90, 83, 96, 172).Image(),
		imgs[2].InsertBottom(tou, 90, 90, 106, 148).Image(),
		imgs[3].InsertBottom(tou, 88, 88, 97, 167).Image(),
		imgs[4].InsertBottom(tou, 90, 85, 89, 179).Image(),
		imgs[5].InsertBottom(tou, 90, 90, 106, 151).Image(),
		imgs[6].Image(),
		imgs[7].Image(),
		imgs[8].Image(),
		imgs[9].Image(),
		imgs[10].Image(),
		imgs[11].Image(),
		imgs[12].Image(),
		imgs[13].Image(),
		imgs[14].Image(),
		imgs[15].Image(),
	}
	g := imgfactory.MergeGif(7, ken)
	return imgfactory.GIF2Base64(g)
}

// pai 拍
func pai(cc *context, value ...string) (string, error) {
	_ = value
	var wg sync.WaitGroup
	var errwg error
	var m sync.Mutex
	// name := cc.usrdir + "拍.gif"
	c := dlrange("pai", 2, &wg, func(e error) {
		m.Lock()
		errwg = e
		m.Unlock()
	})
	tou, err := cc.getLogo(30, 30)
	if err != nil {
		return "", err
	}
	wg.Wait()
	if errwg != nil {
		return "", errwg
	}
	imgs, err := loadFirstFrames(c, 2)
	if err != nil {
		return "", err
	}
	pai := []*image.NRGBA{
		imgs[0].InsertUp(tou, 0, 0, 1, 47).Image(),
		imgs[1].InsertUp(tou, 0, 0, 1, 67).Image(),
	}
	g := imgfactory.MergeGif(1, pai)
	return imgfactory.GIF2Base64(g)
}

// xqe 冲
func xqe(cc *context, value ...string) (string, error) {
	_ = value
	var wg sync.WaitGroup
	var errwg error
	var m sync.Mutex
	// name := cc.usrdir + "冲.gif"
	c := dlrange("xqe", 2, &wg, func(e error) {
		m.Lock()
		errwg = e
		m.Unlock()
	})
	tou, err := cc.getLogo(0, 0)
	if err != nil {
		return "", err
	}
	wg.Wait()
	if errwg != nil {
		return "", errwg
	}
	imgs, err := loadFirstFrames(c, 2)
	if err != nil {
		return "", err
	}
	chong := []*image.NRGBA{
		imgs[0].InsertUp(tou, 30, 30, 15, 53).Image(),
		imgs[1].InsertUp(tou, 30, 30, 40, 53).Image(),
	}
	g := imgfactory.MergeGif(1, chong)
	return imgfactory.GIF2Base64(g)
}

// diu 丢
func diu(cc *context, value ...string) (string, error) {
	_ = value
	var wg sync.WaitGroup
	var errwg error
	var m sync.Mutex
	// name := cc.usrdir + "丢.gif"
	c := dlrange("diu", 8, &wg, func(e error) {
		m.Lock()
		errwg = e
		m.Unlock()
	})
	tou, err := cc.getLogo(0, 0)
	if err != nil {
		return "", err
	}
	wg.Wait()
	if errwg != nil {
		return "", errwg
	}
	imgs, err := loadFirstFrames(c, 8)
	if err != nil {
		return "", err
	}
	diu := []*image.NRGBA{
		imgs[0].InsertUp(tou, 32, 32, 108, 36).Image(),
		imgs[1].InsertUp(tou, 32, 32, 122, 36).Image(),
		imgs[2].Image(),
		imgs[3].InsertUp(tou, 123, 123, 19, 129).Image(),
		imgs[4].InsertUp(tou, 185, 185, -50, 200).InsertUp(tou, 33, 33, 289, 70).Image(),
		imgs[5].InsertUp(tou, 32, 32, 280, 73).Image(),
		imgs[6].InsertUp(tou, 35, 35, 259, 31).Image(),
		imgs[7].InsertUp(tou, 175, 175, -50, 220).Image(),
	}
	g := imgfactory.MergeGif(7, diu)
	return imgfactory.GIF2Base64(g)
}

// kiss 亲
func kiss(cc *context, value ...string) (string, error) {
	_ = value
	var wg sync.WaitGroup
	var err error
	var m sync.Mutex
	piclen := 13
	// name := cc.usrdir + "Kiss.gif"
	c := dlrange("kiss", piclen, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
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
			InsertUp(tou2, 40, 40, selfLocs[i][0], selfLocs[i][1]).Image()
	}
	g := imgfactory.MergeGif(7, kiss)
	return imgfactory.GIF2Base64(g)
}

// garbage 垃圾 垃圾桶
func garbage(cc *context, value ...string) (string, error) {
	_ = value
	var wg sync.WaitGroup
	var err error
	var m sync.Mutex
	piclen := 25
	// name := cc.usrdir + "Garbage.gif"
	c := dlrange("garbage", piclen, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 79, 79)
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
		garbage[i] = imgs[i].InsertBottom(im.Image(), 0, 0, locs[i][0], locs[i][1]).Image()
	}
	g := imgfactory.MergeGif(7, garbage)
	return imgfactory.GIF2Base64(g)
}

// thump 捶
func thump(cc *context, value ...string) (string, error) {
	_ = value
	var wg sync.WaitGroup
	var err error
	var m sync.Mutex
	piclen := 4
	// name := cc.usrdir + "Thump.gif"
	c := dlrange("thump", piclen, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 0, 0)
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
		thump[i] = imgs[i].InsertBottom(im.Image(), locs[i][2], locs[i][3], locs[i][0], locs[i][1]).Image()
	}
	g := imgfactory.MergeGif(7, thump)
	return imgfactory.GIF2Base64(g)
}

// jiujiu 啾啾
func jiujiu(cc *context, value ...string) (string, error) {
	_ = value
	var wg sync.WaitGroup
	var err error
	var m sync.Mutex
	piclen := 8
	// name := cc.usrdir + "Jiujiu.gif"
	c := dlrange("jiujiu", piclen, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 75, 51)
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, piclen)
	if err != nil {
		return "", err
	}
	jiujiu := make([]*image.NRGBA, piclen)
	for i := 0; i < piclen; i++ {
		jiujiu[i] = imgs[i].InsertBottom(im.Image(), 0, 0, 0, 0).Image()
	}
	g := imgfactory.MergeGif(7, jiujiu)
	return imgfactory.GIF2Base64(g)
}

// knock 2敲
func knock(cc *context, value ...string) (string, error) {
	_ = value
	var wg sync.WaitGroup
	var err error
	var m sync.Mutex
	piclen := 8
	// name := cc.usrdir + "Knock.gif"
	c := dlrange("knock", piclen, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 0, 0)
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
		knock[i] = imgs[i].InsertBottom(im.Image(), locs[i][2], locs[i][3], locs[i][0], locs[i][1]).Image()
	}
	g := imgfactory.MergeGif(7, knock)
	return imgfactory.GIF2Base64(g)
}

// 听音乐 listenMusic
func listenMusic(cc *context, value ...string) (string, error) {
	_ = value
	var wg sync.WaitGroup
	var err error
	var m sync.Mutex
	piclen := 1
	// name := cc.usrdir + "ListenMusic.gif"
	c := dlrange("listen_music", piclen, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
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
		listenmusic[i] = imgs[0].InsertBottomC(imgfactory.Rotate(face, float64(-i*10), 215, 215).Image(), 0, 0, 207, 207).Image()
	}
	g := imgfactory.MergeGif(7, listenmusic)
	return imgfactory.GIF2Base64(g)
}

// loveYou 永远爱你
func loveYou(cc *context, value ...string) (string, error) {
	_ = value
	var wg sync.WaitGroup
	var err error
	var m sync.Mutex
	piclen := 2
	// name := cc.usrdir + "LoveYou.gif"
	c := dlrange("love_you", piclen, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 0, 0)
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
		loveyou[i] = imgs[i].InsertBottom(im.Image(), locs[i][2], locs[i][3], locs[i][0], locs[i][1]).Image()
	}
	g := imgfactory.MergeGif(7, loveyou)
	return imgfactory.GIF2Base64(g)
}

// pat 2拍
func pat(cc *context, value ...string) (string, error) {
	_ = value
	var wg sync.WaitGroup
	var err error
	var m sync.Mutex
	piclen := 10
	// name := cc.usrdir + "Pat.gif"
	c := dlrange("pat", piclen, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 0, 0)
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
			p[i] = imgs[i].InsertBottom(im.Image(), locs[1][2], locs[1][3], locs[1][0], locs[1][1]).Image()
		} else {
			p[i] = imgs[i].InsertBottom(im.Image(), locs[0][2], locs[0][3], locs[0][0], locs[0][1]).Image()
		}
	}
	seq := []int{0, 1, 2, 3, 1, 2, 3, 0, 1, 2, 3, 0, 0, 1, 2, 3, 0, 0, 0, 0, 4, 5, 5, 5, 6, 7, 8, 9}
	pat := make([]*image.NRGBA, len(seq))
	for i := 0; i < len(pat); i++ {
		pat[i] = p[seq[i]]
	}
	g := imgfactory.MergeGif(7, pat)
	return imgfactory.GIF2Base64(g)
}

// jackUp 顶
func jackUp(cc *context, value ...string) (string, error) {
	_ = value
	var wg sync.WaitGroup
	var err error
	var m sync.Mutex
	piclen := 23
	// name := cc.usrdir + "JackUp.gif"
	c := dlrange("play", piclen, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 0, 0)
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
			p[i] = imgs[i].InsertBottom(im.Image(), locs[i][2], locs[i][3], locs[i][0], locs[i][1]).Image()
		} else {
			p[i] = imgs[i].Image()
		}
	}
	play := make([]*image.NRGBA, 0, 16)
	play = append(play, p[0:12]...)
	play = append(play, p[0:12]...)
	play = append(play, p[0:8]...)
	play = append(play, p[12:18]...)
	play = append(play, p[18:23]...)
	g := imgfactory.MergeGif(7, play)
	return imgfactory.GIF2Base64(g)
}

// pound 捣
func pound(cc *context, value ...string) (string, error) {
	_ = value
	var wg sync.WaitGroup
	var err error
	var m sync.Mutex
	piclen := 8
	// name := cc.usrdir + "Pound.gif"
	c := dlrange("pound", piclen, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 0, 0)
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
		pound[i] = imgs[i].InsertBottom(im.Image(), locs[i][2], locs[i][3], locs[i][0], locs[i][1]).Image()
	}
	g := imgfactory.MergeGif(7, pound)
	return imgfactory.GIF2Base64(g)
}

// punch 打拳
func punch(cc *context, value ...string) (string, error) {
	_ = value
	var wg sync.WaitGroup
	var err error
	var m sync.Mutex
	piclen := 13
	// name := cc.usrdir + "Punch.gif"
	c := dlrange("punch", piclen, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 260, 260)
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
		punch[i] = imgs[i].InsertBottom(im.Image(), 0, 0, locs[i][0], locs[i][1]-15).Image()
	}
	g := imgfactory.MergeGif(7, punch)
	return imgfactory.GIF2Base64(g)
}

// roll 滚
func roll(cc *context, value ...string) (string, error) {
	_ = value
	var wg sync.WaitGroup
	var err error
	var m sync.Mutex
	piclen := 8
	// name := cc.usrdir + "roll.gif"
	c := dlrange("roll", piclen, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 210, 210)
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
		roll[i] = imgs[i].InsertBottomC(imgfactory.Rotate(im.Image(), float64(locs[i][2]), 0, 0).Image(), 0, 0, locs[i][0]+105, locs[i][1]+105).Image()
	}
	g := imgfactory.MergeGif(7, roll)
	return imgfactory.GIF2Base64(g)
}

// suck 吸 嗦
func suck(cc *context, value ...string) (string, error) {
	_ = value
	var wg sync.WaitGroup
	var err error
	var m sync.Mutex
	piclen := 12
	// name := cc.usrdir + "Suck.gif"
	c := dlrange("suck", piclen, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 0, 0)
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
		suck[i] = imgs[i].InsertBottom(im.Image(), locs[i][2], locs[i][3], locs[i][0], locs[i][1]).Image()
	}
	g := imgfactory.MergeGif(7, suck)
	return imgfactory.GIF2Base64(g)
}

// hammer 锤
func hammer(cc *context, value ...string) (string, error) {
	_ = value
	var wg sync.WaitGroup
	var err error
	var m sync.Mutex
	piclen := 7
	// name := cc.usrdir + "Hammer.gif"
	c := dlrange("hammer", piclen, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 0, 0)
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
		hammer[i] = imgs[i].InsertBottom(im.Image(), locs[i][2], locs[i][3], locs[i][0], locs[i][1]).Image()
	}
	g := imgfactory.MergeGif(7, hammer)
	return imgfactory.GIF2Base64(g)
}

// tightly 紧贴 紧紧贴着
func tightly(cc *context, value ...string) (string, error) {
	_ = value
	var wg sync.WaitGroup
	var err error
	var m sync.Mutex
	piclen := 20
	// name := cc.usrdir + "Tightly.gif"
	c := dlrange("tightly", piclen, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 0, 0)
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
		tightly[i] = imgs[i].InsertBottom(im.Image(), locs[i][2], locs[i][3], locs[i][0], locs[i][1]).Image()
	}
	g := imgfactory.MergeGif(7, tightly)
	return imgfactory.GIF2Base64(g)
}

// turn 转
func turn(cc *context, value ...string) (string, error) {
	_ = value
	// name := cc.usrdir + "Turn.gif"
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
		turn[i] = imgfactory.Size(canvas.Image(), 0, 0).InsertUpC(imgfactory.Rotate(face, float64(10*i), 250, 250).Image(), 0, 0, 125, 125).Image()
	}
	g := imgfactory.MergeGif(7, turn)
	return imgfactory.GIF2Base64(g)
}

// taiguan 抬棺
func taiguan(cc *context, value ...string) (string, error) {
	_ = value
	var wg sync.WaitGroup
	var errwg error
	var m sync.Mutex
	// name := cc.usrdir + "taiguan.gif"
	c := dlrange("taiguan", 20, &wg, func(e error) {
		m.Lock()
		errwg = e
		m.Unlock()
	})
	tou, err := cc.getLogo(0, 0)
	if err != nil {
		return "", err
	}
	wg.Wait()
	if errwg != nil {
		return "", errwg
	}
	imgs, err := loadFirstFrames(c, 20)
	if err != nil {
		return "", err
	}
	taiguan := []*image.NRGBA{
		imgs[0].InsertUp(tou, 85, 85, 180, 65).Image(),
		imgs[1].InsertUp(tou, 85, 85, 180, 65).Image(),
		imgs[2].InsertUp(tou, 85, 85, 180, 65).Image(),
		imgs[3].InsertUp(tou, 85, 85, 180, 65).Image(),
		imgs[4].InsertUp(tou, 85, 85, 177, 65).Image(),
		imgs[5].InsertUp(tou, 85, 85, 175, 65).Image(),
		imgs[6].InsertUp(tou, 85, 85, 173, 65).Image(),
		imgs[7].InsertUp(tou, 85, 85, 171, 65).Image(),
		imgs[8].InsertUp(tou, 85, 85, 170, 65).Image(),
		imgs[9].InsertUp(tou, 85, 85, 170, 65).Image(),
		imgs[10].InsertUp(tou, 85, 85, 170, 65).Image(),
		imgs[11].InsertUp(tou, 85, 85, 170, 65).Image(),
		imgs[12].InsertUp(tou, 85, 85, 170, 65).Image(),
		imgs[13].InsertUp(tou, 85, 85, 170, 65).Image(),
		imgs[14].InsertUp(tou, 85, 85, 170, 65).Image(),
		imgs[15].InsertUp(tou, 85, 85, 170, 65).Image(),
		imgs[16].InsertUp(tou, 85, 85, 170, 65).Image(),
		imgs[17].InsertUp(tou, 85, 85, 170, 65).Image(),
		imgs[18].InsertUp(tou, 85, 85, 170, 65).Image(),
		imgs[19].InsertUp(tou, 85, 85, 175, 65).Image(),
	}
	g := imgfactory.MergeGif(7, taiguan)
	return imgfactory.GIF2Base64(g)
}

// zou 揍
func zou(cc *context, value ...string) (string, error) {
	_ = value
	var wg sync.WaitGroup
	var errwg error
	var m sync.Mutex
	// name := cc.usrdir + "zou.gif"
	c := dlrange("zou", 3, &wg, func(e error) {
		m.Lock()
		errwg = e
		m.Unlock()
	})
	tou, err := cc.getLogo(100, 100)
	if err != nil {
		return "", err
	}
	tou2, err := cc.getLogo2(100, 100)
	if err != nil {
		return "", err
	}
	wg.Wait()
	if errwg != nil {
		return "", errwg
	}
	imgs, err := loadFirstFrames(c, 3)
	if err != nil {
		return "", err
	}
	zou := []*image.NRGBA{
		imgs[0].InsertUp(tou, 40, 40, 98, 138).InsertUp(tou2, 55, 55, 100, 45).Image(),
		imgs[1].InsertUp(tou, 40, 40, 98, 138).InsertUp(tou2, 55, 55, 101, 45).Image(),
		imgs[2].InsertUp(tou, 40, 40, 89, 140).InsertUp(tou2, 55, 55, 99, 40).Image(),
	}
	g := imgfactory.MergeGif(8, zou)
	return imgfactory.GIF2Base64(g)
}

// ci 吞
func ci(cc *context, value ...string) (string, error) {
	_ = value
	var wg sync.WaitGroup
	var errwg error
	var m sync.Mutex
	// name := cc.usrdir + "ci.gif"
	c := dlrange("ci", 26, &wg, func(e error) {
		m.Lock()
		errwg = e
		m.Unlock()
	})
	tou, err := cc.getLogo(100, 100)
	if err != nil {
		return "", err
	}
	wg.Wait()
	if errwg != nil {
		return "", errwg
	}
	imgs, err := loadFirstFrames(c, 26)
	if err != nil {
		return "", err
	}
	ci := []*image.NRGBA{
		imgs[0].InsertBottom(tou, 25, 25, 25, 57).Image(),
		imgs[1].InsertBottom(tou, 25, 25, 27, 58).Image(),
		imgs[2].InsertBottom(tou, 25, 25, 28, 57).Image(),
		imgs[3].InsertBottom(tou, 25, 25, 30, 57).Image(),
		imgs[4].InsertBottom(tou, 25, 25, 30, 58).Image(),
		imgs[5].InsertBottom(tou, 25, 25, 30, 59).Image(),
		imgs[6].Image(),
		imgs[7].Image(),
		imgs[8].Image(),
		imgs[9].Image(),
		imgs[10].Image(),
		imgs[11].Image(),
		imgs[12].Image(),
		imgs[13].Image(),
		imgs[14].Image(),
		imgs[15].Image(),
		imgs[16].Image(),
		imgs[17].Image(),
		imgs[18].Image(),
		imgs[19].Image(),
		imgs[20].Image(),
		imgs[21].Image(),
		imgs[22].Image(),
		imgs[23].Image(),
		imgs[24].Image(),
		imgs[25].Image(),
	}
	g := imgfactory.MergeGif(7, ci)
	return imgfactory.GIF2Base64(g)
}

// worship 膜拜
func worship(cc *context, value ...string) (string, error) {
	_ = value
	var wg sync.WaitGroup
	var err error
	var m sync.Mutex
	// name := cc.usrdir + "worship.gif"
	c := dlrange("worship", 9, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	face, err := gg.LoadImage(cc.headimgsdir[0])
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 9)
	if err != nil {
		return "", err
	}
	worship := []*image.NRGBA{
		imgs[0].InsertBottom(face, 140, 140, 0, 0).Image(),
		imgs[1].InsertBottom(face, 140, 140, 0, 0).Image(),
		imgs[2].InsertBottom(face, 140, 140, 0, 0).Image(),
		imgs[3].InsertBottom(face, 140, 140, 0, 0).Image(),
		imgs[4].InsertBottom(face, 140, 140, 0, 0).Image(),
		imgs[5].InsertBottom(face, 140, 140, 0, 0).Image(),
		imgs[6].InsertBottom(face, 140, 140, 0, 0).Image(),
		imgs[7].InsertBottom(face, 140, 140, 0, 0).Image(),
		imgs[8].InsertBottom(face, 140, 140, 0, 0).Image(),
	}
	g := imgfactory.MergeGif(7, worship)
	return imgfactory.GIF2Base64(g)
}

// 2ceng 2蹭
func ceng2(cc *context, value ...string) (string, error) {
	_ = value
	var wg sync.WaitGroup
	var errwg error
	var m sync.Mutex
	// name := cc.usrdir + "ceng2.gif"
	c := dlrange("ceng2", 4, &wg, func(e error) {
		m.Lock()
		errwg = e
		m.Unlock()
	})
	tou, err := cc.getLogo(100, 100)
	if err != nil {
		return "", err
	}
	wg.Wait()
	if errwg != nil {
		return "", errwg
	}
	imgs, err := loadFirstFrames(c, 4)
	if err != nil {
		return "", err
	}
	ceng2 := []*image.NRGBA{
		imgs[0].InsertBottom(tou, 175, 175, 78, 263).Image(),
		imgs[1].InsertBottom(tou, 175, 175, 78, 263).Image(),
		imgs[2].InsertBottom(tou, 175, 175, 78, 263).Image(),
		imgs[3].InsertBottom(tou, 175, 175, 78, 263).Image(),
	}
	g := imgfactory.MergeGif(7, ceng2)
	return imgfactory.GIF2Base64(g)
}

// dun 炖
func dun(cc *context, value ...string) (string, error) {
	_ = value
	var wg sync.WaitGroup
	var errwg error
	var m sync.Mutex
	// name := cc.usrdir + "dun.gif"
	c := dlrange("dun", 5, &wg, func(e error) {
		m.Lock()
		errwg = e
		m.Unlock()
	})
	tou, err := cc.getLogo(100, 100)
	if err != nil {
		return "", err
	}
	wg.Wait()
	if errwg != nil {
		return "", errwg
	}
	imgs, err := loadFirstFrames(c, 5)
	if err != nil {
		return "", err
	}
	dun := []*image.NRGBA{
		imgs[0].InsertBottom(tou, 80, 80, 85, 45).Image(),
		imgs[1].InsertBottom(tou, 80, 80, 85, 45).Image(),
		imgs[2].InsertBottom(tou, 80, 80, 85, 45).Image(),
		imgs[3].InsertBottom(tou, 80, 80, 85, 45).Image(),
		imgs[4].InsertBottom(tou, 80, 80, 85, 45).Image(),
	}
	g := imgfactory.MergeGif(7, dun)
	return imgfactory.GIF2Base64(g)
}

// push 滚高清重置版 过渡
func push(cc *context, value ...string) (string, error) {
	_ = value
	var wg sync.WaitGroup
	var err error
	var m sync.Mutex
	piclen := 16
	// name := cc.usrdir + "push.gif"
	c := dlrange("push", piclen, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	tou, err := cc.getLogo(0, 0)
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, piclen)
	if err != nil {
		return "", err
	}
	push := make([]*image.NRGBA, piclen)
	for i := 0; i < piclen; i++ {
		push[i] = imgs[i].InsertUpC(imgfactory.Rotate(tou, float64(-22*i), 280, 280).Image(), 0, 0, 523, 291).Image()
	}
	g := imgfactory.MergeGif(7, push)
	return imgfactory.GIF2Base64(g)
}

// peng 砰
func peng(cc *context, value ...string) (string, error) {
	_ = value
	var wg sync.WaitGroup
	var errwg error
	var m sync.Mutex
	// name := cc.usrdir + "peng.gif"
	c := dlrange("peng", 25, &wg, func(e error) {
		m.Lock()
		errwg = e
		m.Unlock()
	})
	tou, err := cc.getLogo(100, 100)
	if err != nil {
		return "", err
	}
	m1 := imgfactory.Rotate(tou, 1, 80, 80)
	m2 := imgfactory.Rotate(tou, 30, 80, 80)
	m3 := imgfactory.Rotate(tou, 45, 85, 85)
	m4 := imgfactory.Rotate(tou, 90, 80, 80)
	wg.Wait()
	if errwg != nil {
		return "", errwg
	}
	imgs, err := loadFirstFrames(c, 25)
	if err != nil {
		return "", err
	}
	peng := []*image.NRGBA{
		imgs[0].Image(),
		imgs[1].Image(),
		imgs[2].Image(),
		imgs[3].Image(),
		imgs[4].Image(),
		imgs[5].Image(),
		imgs[6].Image(),
		imgs[7].InsertUp(m1.Image(), 0, 0, 205, 80).Image(),
		imgs[8].InsertUp(m1.Image(), 0, 0, 205, 80).Image(),
		imgs[9].InsertUp(m1.Image(), 0, 0, 205, 80).Image(),
		imgs[10].InsertUp(m1.Image(), 0, 0, 205, 80).Image(),
		imgs[11].InsertUp(m1.Image(), 0, 0, 205, 80).Image(),
		imgs[12].InsertUp(m1.Image(), 0, 0, 205, 80).Image(),
		imgs[13].InsertUp(m1.Image(), 0, 0, 205, 80).Image(),
		imgs[14].InsertUp(m1.Image(), 0, 0, 205, 80).Image(),
		imgs[15].InsertUp(m1.Image(), 0, 0, 205, 80).Image(),
		imgs[16].InsertUp(m1.Image(), 0, 0, 200, 80).Image(),
		imgs[17].InsertUp(m2.Image(), 0, 0, 169, 65).Image(),
		imgs[18].InsertUp(m2.Image(), 0, 0, 160, 69).Image(),
		imgs[19].InsertUp(m3.Image(), 0, 0, 113, 90).Image(),
		imgs[20].InsertUp(m4.Image(), 0, 0, 89, 159).Image(),
		imgs[21].InsertUp(m4.Image(), 0, 0, 89, 159).Image(),
		imgs[22].InsertUp(m4.Image(), 0, 0, 86, 160).Image(),
		imgs[23].InsertUp(m4.Image(), 0, 0, 89, 159).Image(),
		imgs[24].InsertUp(m4.Image(), 0, 0, 86, 160).Image(),
	}
	g := imgfactory.MergeGif(8, peng)
	return imgfactory.GIF2Base64(g)
}

// klee 可莉吃
func klee(cc *context, value ...string) (string, error) {
	_ = value
	var wg sync.WaitGroup
	var err error
	var m sync.Mutex
	piclen := 31
	// name := cc.usrdir + "klee.gif"
	c := dlrange("klee", piclen, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 82, 83)
	if err != nil {
		return "", err
	}
	locs := [][]int{{0, 174}, {0, 174}, {0, 174}, {0, 174}, {0, 174}, {12, 160}, {19, 152}, {23, 148}, {26, 145}, {32, 140}, {37, 136}, {42, 131}, {49, 127}, {70, 126}, {88, 128}, {-30, 210}, {-19, 207}, {-14, 200}, {-10, 188}, {-7, 179}, {-3, 170}, {-3, 175}, {-1, 174}, {0, 174}, {0, 174}, {0, 174}, {0, 174}, {0, 174}, {0, 174}, {0, 174}, {0, 174}}
	imgs, err := loadFirstFrames(c, piclen)
	if err != nil {
		return "", err
	}
	klee := make([]*image.NRGBA, piclen)
	for i := 0; i < piclen; i++ {
		klee[i] = imgs[i].InsertBottom(im.Image(), 0, 0, locs[i][0], locs[i][1]).Image()
	}
	g := imgfactory.MergeGif(7, klee)
	return imgfactory.GIF2Base64(g)
}

// hutaoken 胡桃啃
func hutaoken(cc *context, value ...string) (string, error) {
	_ = value
	var wg sync.WaitGroup
	var errwg error
	var m sync.Mutex
	// name := cc.usrdir + "hutaoken.gif"
	c := dlrange("hutaoken", 2, &wg, func(e error) {
		m.Lock()
		errwg = e
		m.Unlock()
	})
	tou, err := cc.getLogo(55, 55)
	if err != nil {
		return "", err
	}
	wg.Wait()
	if errwg != nil {
		return "", errwg
	}
	imgs, err := loadFirstFrames(c, 2)
	if err != nil {
		return "", err
	}
	hutaoken := []*image.NRGBA{
		imgs[0].InsertBottom(tou, 98, 101, 108, 234).Image(),
		imgs[1].InsertBottom(tou, 96, 100, 108, 237).Image(),
	}
	g := imgfactory.MergeGif(8, hutaoken)
	return imgfactory.GIF2Base64(g)
}

// lick 2舔
func lick(cc *context, value ...string) (string, error) {
	_ = value
	var wg sync.WaitGroup
	var errwg error
	var m sync.Mutex
	// name := cc.usrdir + "lick.gif"
	c := dlrange("lick", 2, &wg, func(e error) {
		m.Lock()
		errwg = e
		m.Unlock()
	})
	tou, err := cc.getLogo(100, 100)
	if err != nil {
		return "", err
	}
	wg.Wait()
	if errwg != nil {
		return "", errwg
	}
	imgs, err := loadFirstFrames(c, 2)
	if err != nil {
		return "", err
	}
	lick := []*image.NRGBA{
		imgs[0].InsertUp(tou, 44, 44, 10, 138).Image(),
		imgs[1].InsertUp(tou, 44, 44, 10, 138).Image(),
	}
	g := imgfactory.MergeGif(8, lick)
	return imgfactory.GIF2Base64(g)
}

// tiqiu 踢球
func tiqiu(cc *context, value ...string) (string, error) {
	_ = value
	var wg sync.WaitGroup
	var err error
	var m sync.Mutex
	piclen := 15
	// name := cc.usrdir + "tiqiu.gif"
	c := dlrange("tiqiu", piclen, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	tou, err := cc.getLogo(78, 78)
	if err != nil {
		return "", err
	}
	locs := [][]int{{58, 137}, {57, 118}, {56, 100}, {53, 114}, {51, 127}, {49, 140}, {48, 113}, {48, 86}, {48, 58}, {49, 98}, {51, 137}, {52, 177}, {53, 170}, {56, 182}, {59, 154}}
	imgs, err := loadFirstFrames(c, piclen)
	if err != nil {
		return "", err
	}
	tiqiu := make([]*image.NRGBA, piclen)
	for i := 0; i < piclen; i++ {
		tiqiu[i] = imgs[i].InsertUpC(imgfactory.Rotate(tou, float64(-24*i), 0, 0).Image(), 0, 0, locs[i][0]+38, locs[i][1]+38).Image()
	}
	g := imgfactory.MergeGif(7, tiqiu)
	return imgfactory.GIF2Base64(g)
}

// cai 踩
func cai(cc *context, value ...string) (string, error) {
	_ = value
	var wg sync.WaitGroup
	var errwg error
	var m sync.Mutex
	// name := cc.usrdir + "cai.gif"
	c := dlrange("cai", 5, &wg, func(e error) {
		m.Lock()
		errwg = e
		m.Unlock()
	})
	tou, err := cc.getLogo(0, 0)
	if err != nil {
		return "", err
	}
	wg.Wait()
	if errwg != nil {
		return "", errwg
	}
	imgs, err := loadFirstFrames(c, 5)
	if err != nil {
		return "", err
	}
	m1 := imgfactory.Rotate(tou, -20, 130, 80)
	cai := []*image.NRGBA{
		imgs[0].InsertBottom(m1.Image(), 123, 105, 39, 188).Image(),
		imgs[1].InsertBottom(m1.Image(), 123, 105, 39, 188).Image(),
		imgs[2].InsertBottom(tou, 90, 71, 50, 209).Image(),
		imgs[3].InsertBottom(tou, 85, 76, 52, 203).Image(),
		imgs[4].InsertBottom(tou, 88, 82, 49, 198).Image(),
	}
	g := imgfactory.MergeGif(7, cai)
	return imgfactory.GIF2Base64(g)
}

// whir 2转
func whirl(cc *context, value ...string) (string, error) {
	_ = value
	var wg sync.WaitGroup
	var err error
	var m sync.Mutex
	piclen := 15
	// name := cc.usrdir + "whirl.gif"
	c := dlrange("whirl", piclen, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	tou, err := cc.getLogo(0, 0)
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, piclen)
	if err != nil {
		return "", err
	}
	whirl := make([]*image.NRGBA, piclen)
	for i := 0; i < piclen; i++ {
		whirl[i] = imgs[i].InsertUpC(imgfactory.Rotate(tou, float64(-24*i), 145, 145).Image(), 0, 0, 115, 89).Image()
	}
	g := imgfactory.MergeGif(7, whirl)
	return imgfactory.GIF2Base64(g)
}

// always 一直
func alwaysDoGif(cc *context, value ...string) (string, error) {
	_ = value
	var err error
	var face []*imgfactory.Factory
	// name := cc.usrdir + "AlwaysDo.gif"
	face, err = imgfactory.LoadAllTrueFrames(cc.headimgsdir[0], 500, 500)
	if err != nil {
		// 载入失败尝试载入第一帧
		face = nil
		first, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 500, 500)
		if err != nil {
			return "", err
		}
		face = append(face, imgfactory.NewFactory(first.Image()))
	}
	canvas := gg.NewContext(500, 600)
	canvas.SetColor(color.Black)
	data, err := file.GetLazyData(text.BoldFontFile, control.Md5File, true)
	if err != nil {
		return "", err
	}
	err = canvas.ParseFontFace(data, 40)
	if err != nil {
		return "", err
	}
	length := len(face)
	if length > 50 {
		length = 50
	}
	arg := "要我一直"
	l, _ := canvas.MeasureString(arg)
	if l > 500 {
		return "", errors.New("文字消息太长了")
	}
	turn := make([]*image.NRGBA, length)
	for i, f := range face {
		canvas := gg.NewContext(500, 600)
		canvas.DrawImage(f.Image(), 0, 0)
		canvas.SetColor(color.Black)
		_ = canvas.ParseFontFace(data, 40)
		canvas.DrawString(arg, 280-l, 560)
		canvas.DrawImage(imgfactory.Size(f.Image(), 90, 90).Image(), 280, 505)
		canvas.DrawString("吗", 370, 560)
		turn[i] = imgfactory.Size(canvas.Image(), 0, 0).Image()
	}
	g := imgfactory.MergeGif(8, turn)
	return imgfactory.GIF2Base64(g)
}
