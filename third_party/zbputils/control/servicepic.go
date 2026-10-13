package control

import (
	"image"

	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/rendercard"
	ctrl "github.com/FloatTech/zbpctrl"
	zero "github.com/wdvxdr1123/ZeroBot"

	"github.com/FloatTech/zbputils/img/text"
)

const (
	kanbanpath = "data/Control/"
)

var (
	// serverlistlogo 缓存
	serverlistlogo image.Image
	// 字体 GlowSans 数据
	glowsd []byte
	// 字体 Impact 数据
	impactd []byte
	// 字体 Torus 数据
	torussd []byte
	// 字体 Yuruka 数据
	yurukasd []byte
)

func init() {
	_, err := file.GetLazyData(kanbanpath+"kanban.png", Md5File, false)
	if err != nil {
		panic(err)
	}
	_, err = file.GetLazyData(kanbanpath+"zbpuwu.png", Md5File, false)
	if err != nil {
		panic(err)
	}
	glowsd, err = file.GetLazyData(text.GlowSansFontFile, Md5File, true)
	if err != nil {
		panic(err)
	}
	impactd, err = file.GetLazyData(text.ImpactFontFile, Md5File, true)
	if err != nil {
		panic(err)
	}
	torussd, err = file.GetLazyData(text.TorusFontFile, Md5File, true)
	if err != nil {
		panic(err)
	}
	yurukasd, err = file.GetLazyData(text.YurukaFontFile, Md5File, true)
	if err != nil {
		panic(err)
	}
}

func renderservepicof(gid int64) (img image.Image, err error) {
	pluginlist := make([]*rendercard.PluginInfo, len(priomap))
	ForEachByPrio(func(i int, manager *ctrl.Control[*zero.Ctx]) bool {
		pluginlist[i] = &rendercard.PluginInfo{
			Name:   manager.Service,
			Brief:  manager.Options.Brief,
			Status: manager.IsEnabledIn(gid),
		}
		return true
	})
	if serverlistlogo == nil {
		serverlistlogo, err = rendercard.RenderServerListLogo(yurukasd)
		if err != nil {
			return
		}
	}
	img, err = rendercard.RenderServerPic(pluginlist, torussd, glowsd, kanbanpath+"zbpuwu.png", serverlistlogo)
	return
}
