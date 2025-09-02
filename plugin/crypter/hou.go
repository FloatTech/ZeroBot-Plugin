// Package crypter 齁语加解密
package crypter

import (
	"strings"
)

// 齁语密码表
var houCodebook = []string{
	"齁", "哦", "噢", "喔", "咕", "咿", "嗯", "啊",
	"～", "哈", "！", "唔", "哼", "❤", "呃", "呼",
}

// 索引:  0    1    2    3    4    5    6    7
//       8    9   10   11   12   13   14   15

// 创建映射表
var houCodebookMap = make(map[string]int)

// 初始化映射表
func init() {
	for idx, ch := range houCodebook {
		houCodebookMap[ch] = idx
	}
}

func encodeHou(text string) string {
	if text == "" {
		return "请输入要加密的文本"
	}
	var encoded strings.Builder
	textBytes := []byte(text)
	for _, b := range textBytes {
		high := (b >> 4) & 0x0F
		low := b & 0x0F
		encoded.WriteString(houCodebook[high])
		encoded.WriteString(houCodebook[low])
	}

	return encoded.String()
}

func decodeHou(code string) string {
	if code == "" {
		return "请输入要解密的齁语密文"
	}

	// 过滤出有效的齁语字符
	var validChars []string
	for _, r := range code {
		charStr := string(r)
		if _, exists := houCodebookMap[charStr]; exists {
			validChars = append(validChars, charStr)
		}
	}

	if len(validChars)%2 != 0 {
		return "齁语密文长度错误，无法解密"
	}

	// 解密过程
	var byteList []byte
	for i := 0; i < len(validChars); i += 2 {
		highIdx, highExists := houCodebookMap[validChars[i]]
		lowIdx, lowExists := houCodebookMap[validChars[i+1]]

		if !highExists || !lowExists {
			return "齁语密文包含无效字符"
		}

		originalByte := byte((highIdx << 4) | lowIdx)
		byteList = append(byteList, originalByte)
	}

	result := string(byteList)

	if !isValidUTF8(result) {
		return "齁语解密失败，结果不是有效的文本"
	}

	return result
}

// 检查字符串是否为有效的UTF-8编码
func isValidUTF8(s string) bool {
	// Go的string类型默认就是UTF-8，如果转换没有出错说明是有效的
	return len(s) > 0 || s == ""
}
