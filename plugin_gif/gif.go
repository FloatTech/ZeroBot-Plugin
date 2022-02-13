package gif

import (
	"image"
	"sync"

	"github.com/FloatTech/zbputils/img"
	"github.com/FloatTech/zbputils/img/writer"
)

// A摸
func (cc *context) A摸() (string, error) {
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

// A搓
func (cc *context) A搓() (string, error) {
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

// A敲
func (cc *context) A敲() (string, error) {
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

// A吃
func (cc *context) A吃() (string, error) {
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

// A蹭
func (cc *context) A蹭() (string, error) {
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

// A啃
func (cc *context) A啃() (string, error) {
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

// A拍
func (cc *context) A拍() (string, error) {
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

// A冲
func (cc *context) A冲() (string, error) {
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

// A丢
func (cc *context) A丢() (string, error) {
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
