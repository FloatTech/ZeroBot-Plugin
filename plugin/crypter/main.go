// Package crypter 奇怪语言加解密
package crypter

import (
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
)

func init() {
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "奇怪语言加解密",
		Help: "多种语言加解密插件\n" +
			"- 齁语加解密:\n" +
			"- 齁语加密 [文本] 或 h加密 [文本]\n" +
			"- 齁语解密 [密文] 或 h解密 [密文]\n\n" +
			"- Fumo语加解密:\n" +
			"- fumo加密 [文本]\n" +
			"- fumo解密 [密文]\n\n",
		PublicDataFolder: "Crypter",
	})

	// hou
	engine.OnRegex(`^(?:齁语加密|h加密)\s*(.+)$`).SetBlock(true).Handle(houEncryptHandler)
	engine.OnRegex(`^(?:齁语解密|h解密)\s*(.+)$`).SetBlock(true).Handle(houDecryptHandler)

	// Fumo
	engine.OnRegex(`^fumo加密\s*(.+)$`).SetBlock(true).Handle(fumoEncryptHandler)
	engine.OnRegex(`^fumo解密\s*(.+)$`).SetBlock(true).Handle(fumoDecryptHandler)
}
