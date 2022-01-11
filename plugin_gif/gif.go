package plugin_gif

import (
	"image"

	"github.com/FloatTech/zbputils/img"
)

// 摸
func (cc *context) mo() string {
	name := cc.usrdir + `摸.gif`
	c := dlrange(`mo/`, `.png`, 5)
	tou := img.LoadFirstFrame(cc.headimgsdir[0], 0, 0).Circle(0).Im
	mo := []*image.NRGBA{
		img.LoadFirstFrame(*<-(*c)[0], 0, 0).InsertBottom(tou, 80, 80, 32, 32).Im,
		img.LoadFirstFrame(*<-(*c)[1], 0, 0).InsertBottom(tou, 70, 90, 42, 22).Im,
		img.LoadFirstFrame(*<-(*c)[2], 0, 0).InsertBottom(tou, 75, 85, 37, 27).Im,
		img.LoadFirstFrame(*<-(*c)[3], 0, 0).InsertBottom(tou, 85, 75, 27, 37).Im,
		img.LoadFirstFrame(*<-(*c)[4], 0, 0).InsertBottom(tou, 90, 70, 22, 42).Im,
	}
	img.SaveGif(img.MergeGif(1, mo), name)
	return "file:///" + name
}

// 搓
func (cc *context) cuo() string {
	name := cc.usrdir + `搓.gif`
	c := dlrange(`cuo/`, `.png`, 5)
	tou := img.LoadFirstFrame(cc.headimgsdir[0], 110, 110).Circle(0).Im
	m1 := img.Rotate(tou, 72, 0, 0)
	m2 := img.Rotate(tou, 144, 0, 0)
	m3 := img.Rotate(tou, 216, 0, 0)
	m4 := img.Rotate(tou, 288, 0, 0)
	cuo := []*image.NRGBA{
		img.LoadFirstFrame(*<-(*c)[0], 0, 0).InsertBottomC(tou, 0, 0, 75, 130).Im,
		img.LoadFirstFrame(*<-(*c)[1], 0, 0).InsertBottomC(m1.Im, 0, 0, 75, 130).Im,
		img.LoadFirstFrame(*<-(*c)[2], 0, 0).InsertBottomC(m2.Im, 0, 0, 75, 130).Im,
		img.LoadFirstFrame(*<-(*c)[3], 0, 0).InsertBottomC(m3.Im, 0, 0, 75, 130).Im,
		img.LoadFirstFrame(*<-(*c)[4], 0, 0).InsertBottomC(m4.Im, 0, 0, 75, 130).Im,
	}
	img.SaveGif(img.MergeGif(5, cuo), name)
	return "file:///" + name
}

// 敲
func (cc *context) qiao() string {
	name := cc.usrdir + `敲.gif`
	c := dlrange(`qiao/`, `.png`, 2)
	tou := img.LoadFirstFrame(cc.headimgsdir[0], 40, 40).Circle(0).Im
	qiao := []*image.NRGBA{
		img.LoadFirstFrame(*<-(*c)[0], 0, 0).InsertUp(tou, 40, 33, 57, 52).Im,
		img.LoadFirstFrame(*<-(*c)[1], 0, 0).InsertUp(tou, 38, 36, 58, 50).Im,
	}
	img.SaveGif(img.MergeGif(1, qiao), name)
	return "file:///" + name
}

// 吃
func (cc *context) chi() string {
	name := cc.usrdir + `吃.gif`
	c := dlrange(`chi/`, `.png`, 3)
	tou := img.LoadFirstFrame(cc.headimgsdir[0], 32, 32).Im
	chi := []*image.NRGBA{
		img.LoadFirstFrame(*<-(*c)[0], 0, 0).InsertBottom(tou, 0, 0, 1, 38).Im,
		img.LoadFirstFrame(*<-(*c)[1], 0, 0).InsertBottom(tou, 0, 0, 1, 38).Im,
		img.LoadFirstFrame(*<-(*c)[2], 0, 0).InsertBottom(tou, 0, 0, 1, 38).Im,
	}
	img.SaveGif(img.MergeGif(1, chi), name)
	return "file:///" + name
}

// 蹭
func (cc *context) ceng() string {
	name := cc.usrdir + `蹭.gif`
	c := dlrange(`ceng/`, `.png`, 6)
	tou := img.LoadFirstFrame(cc.headimgsdir[0], 100, 100).Circle(0).Im
	tou2 := img.LoadFirstFrame(cc.headimgsdir[1], 100, 100).Circle(0).Im
	ceng := []*image.NRGBA{
		img.LoadFirstFrame(*<-(*c)[0], 0, 0).InsertUp(tou, 75, 77, 40, 88).InsertUp(tou2, 77, 103, 102, 81).Im,
		img.LoadFirstFrame(*<-(*c)[1], 0, 0).InsertUp(tou, 75, 77, 46, 100).InsertUp(img.Rotate(tou2, 10, 62, 127).Im, 0, 0, 92, 40).Im,
		img.LoadFirstFrame(*<-(*c)[2], 0, 0).InsertUp(tou, 75, 77, 67, 99).InsertUp(tou2, 76, 117, 90, 8).Im,
		img.LoadFirstFrame(*<-(*c)[3], 0, 0).InsertUp(tou, 75, 77, 52, 83).InsertUp(img.Rotate(tou2, -40, 94, 94).Im, 0, 0, 53, -20).Im,
		img.LoadFirstFrame(*<-(*c)[4], 0, 0).InsertUp(tou, 75, 77, 56, 110).InsertUp(img.Rotate(tou2, -66, 132, 80).Im, 0, 0, 78, 40).Im,
		img.LoadFirstFrame(*<-(*c)[5], 0, 0).InsertUp(tou, 75, 77, 62, 102).InsertUp(tou2, 71, 100, 110, 94).Im,
	}
	img.SaveGif(img.MergeGif(8, ceng), name)
	return "file:///" + name
}

// 啃
func (cc *context) ken() string {
	name := cc.usrdir + `啃.gif`
	c := dlrange(`ken/`, `.png`, 16)
	tou := img.LoadFirstFrame(cc.headimgsdir[0], 100, 100).Im
	ken := []*image.NRGBA{
		img.LoadFirstFrame(*<-(*c)[0], 0, 0).InsertBottom(tou, 90, 90, 105, 150).Im,
		img.LoadFirstFrame(*<-(*c)[1], 0, 0).InsertBottom(tou, 90, 83, 96, 172).Im,
		img.LoadFirstFrame(*<-(*c)[2], 0, 0).InsertBottom(tou, 90, 90, 106, 148).Im,
		img.LoadFirstFrame(*<-(*c)[3], 0, 0).InsertBottom(tou, 88, 88, 97, 167).Im,
		img.LoadFirstFrame(*<-(*c)[4], 0, 0).InsertBottom(tou, 90, 85, 89, 179).Im,
		img.LoadFirstFrame(*<-(*c)[5], 0, 0).InsertBottom(tou, 90, 90, 106, 151).Im,
		img.LoadFirstFrame(*<-(*c)[6], 0, 0).Im,
		img.LoadFirstFrame(*<-(*c)[7], 0, 0).Im,
		img.LoadFirstFrame(*<-(*c)[8], 0, 0).Im,
		img.LoadFirstFrame(*<-(*c)[9], 0, 0).Im,
		img.LoadFirstFrame(*<-(*c)[10], 0, 0).Im,
		img.LoadFirstFrame(*<-(*c)[11], 0, 0).Im,
		img.LoadFirstFrame(*<-(*c)[12], 0, 0).Im,
		img.LoadFirstFrame(*<-(*c)[13], 0, 0).Im,
		img.LoadFirstFrame(*<-(*c)[14], 0, 0).Im,
		img.LoadFirstFrame(*<-(*c)[15], 0, 0).Im,
	}
	img.SaveGif(img.MergeGif(7, ken), name)
	return "file:///" + name
}

// 拍
func (cc *context) pai() string {
	name := cc.usrdir + `拍.gif`
	c := dlrange(`pai/`, `.png`, 2)
	tou := img.LoadFirstFrame(cc.headimgsdir[0], 30, 30).Circle(0).Im
	pai := []*image.NRGBA{
		img.LoadFirstFrame(*<-(*c)[0], 0, 0).InsertUp(tou, 0, 0, 1, 47).Im,
		img.LoadFirstFrame(*<-(*c)[1], 0, 0).InsertUp(tou, 0, 0, 1, 67).Im,
	}
	img.SaveGif(img.MergeGif(1, pai), name)
	return "file:///" + name
}

// 冲
func (cc *context) chong() string {
	name := cc.usrdir + `冲.gif`
	c := dlrange(`xqe/`, `.png`, 2)
	tou := img.LoadFirstFrame(cc.headimgsdir[0], 0, 0).Circle(0).Im
	chong := []*image.NRGBA{
		img.LoadFirstFrame(*<-(*c)[0], 0, 0).InsertUp(tou, 30, 30, 15, 53).Im,
		img.LoadFirstFrame(*<-(*c)[1], 0, 0).InsertUp(tou, 30, 30, 40, 53).Im,
	}
	img.SaveGif(img.MergeGif(1, chong), name)
	return "file:///" + name
}

// 丢
func (cc *context) diu() string {
	name := cc.usrdir + `丢.gif`
	c := dlrange(`diu/`, `.png`, 8)
	tou := img.LoadFirstFrame(cc.headimgsdir[0], 0, 0).Circle(0).Im
	diu := []*image.NRGBA{
		img.LoadFirstFrame(*<-(*c)[0], 0, 0).InsertUp(tou, 32, 32, 108, 36).Im,
		img.LoadFirstFrame(*<-(*c)[1], 0, 0).InsertUp(tou, 32, 32, 122, 36).Im,
		img.LoadFirstFrame(*<-(*c)[2], 0, 0).Im,
		img.LoadFirstFrame(*<-(*c)[3], 0, 0).InsertUp(tou, 123, 123, 19, 129).Im,
		img.LoadFirstFrame(*<-(*c)[4], 0, 0).InsertUp(tou, 185, 185, -50, 200).InsertUp(tou, 33, 33, 289, 70).Im,
		img.LoadFirstFrame(*<-(*c)[5], 0, 0).InsertUp(tou, 32, 32, 280, 73).Im,
		img.LoadFirstFrame(*<-(*c)[6], 0, 0).InsertUp(tou, 35, 35, 259, 31).Im,
		img.LoadFirstFrame(*<-(*c)[7], 0, 0).InsertUp(tou, 175, 175, -50, 220).Im,
	}
	img.SaveGif(img.MergeGif(7, diu), name)
	return "file:///" + name
}
