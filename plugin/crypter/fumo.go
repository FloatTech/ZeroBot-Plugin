// Package crypter Fumo语
package crypter

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
)

// Base64字符表
const base64Chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

// Fumo语字符表 - 使用各种fumo变体来表示base64字符
var fumoChars = []string{
	"fumo-", "Fumo-", "fUmo-", "fuMo-", "fumO-", "FUmo-", "FuMo-", "FumO-",
	"fUMo-", "fUmO-", "fuMO-", "FUMo-", "FUmO-", "fUMO-", "FUMO-", "fumo.",
	"Fumo.", "fUmo.", "fuMo.", "fumO.", "FUmo.", "FuMo.", "FumO.", "fUMo.",
	"fUmO.", "fuMO.", "FUMo.", "FUmO.", "fUMO.", "FUMO.", "fumo,", "Fumo,",
	"fUmo,", "fuMo,", "fumO,", "FUmo,", "FuMo,", "FumO,", "fUMo,", "fUmO,",
	"fuMO,", "FUMo,", "FuMO,", "fUMO,", "FUMO,", "fumo+", "Fumo+", "fUmo+",
	"fuMo+", "fumO+", "FUmo+", "FuMo+", "FumO+", "fUMo+", "fUmO+", "fuMO+",
	"FUMo+", "FUmO+", "fUMO+", "FUMO+", "fumo|", "Fumo|", "fUmo|", "fuMo|",
	"fumO|", "FUmo|", "FuMo|", "FumO|", "fUMo|", "fUmO|", "fuMO|", "fumo/",
	"Fumo/", "fUmo/",
}

// Base64 2 Fumo
// 创建编码映射表
var encodeMap = make(map[byte]string)

// 创建解码映射表
var decodeMap = make(map[string]byte)

func init() {
	for i := 0; i < 64 && i < len(fumoChars); i++ {
		base64Char := base64Chars[i]
		fumoChar := fumoChars[i]

		encodeMap[base64Char] = fumoChar
		decodeMap[fumoChar] = base64Char
	}
}

// 加密
func encryptFumo(text string) string {
	if text == "" {
		return "请输入要加密的文本"
	}
	textBytes := []byte(text)
	base64String := base64.StdEncoding.EncodeToString(textBytes)
	base64Body := strings.TrimRight(base64String, "=")
	paddingCount := len(base64String) - len(base64Body)
	var fumoBody strings.Builder
	for _, char := range base64Body {
		if fumoChar, exists := encodeMap[byte(char)]; exists {
			fumoBody.WriteString(fumoChar)
		} else {
			return fmt.Sprintf("Fumo加密失败: 未知字符 %c", char)
		}
	}
	result := fumoBody.String() + strings.Repeat("=", paddingCount)

	return result
}

// 解密
func decryptFumo(fumoText string) string {
	if fumoText == "" {
		return "请输入要解密的Fumo语密文"
	}
	fumoBody := strings.TrimRight(fumoText, "=")
	paddingCount := len(fumoText) - len(fumoBody)
	fumoPattern := regexp.MustCompile(`(\w+[-.,+|/])`)
	fumoWords := fumoPattern.FindAllString(fumoBody, -1)
	reconstructed := strings.Join(fumoWords, "")
	if reconstructed != fumoBody {
		return "Fumo解密失败: 包含无效的Fumo字符或格式错误"
	}
	var base64Body strings.Builder
	for _, fumoWord := range fumoWords {
		if base64Char, exists := decodeMap[fumoWord]; exists {
			base64Body.WriteByte(base64Char)
		} else {
			return fmt.Sprintf("Fumo解密失败: 包含无效的Fumo字符 %s", fumoWord)
		}
	}
	base64String := base64Body.String() + strings.Repeat("=", paddingCount)
	decodedBytes, err := base64.StdEncoding.DecodeString(base64String)
	if err != nil {
		return fmt.Sprintf("Fumo解密失败: Base64解码错误 %v", err)
	}
	originalText := string(decodedBytes)
	return originalText
}
