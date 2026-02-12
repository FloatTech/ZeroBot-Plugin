// Package crypter QQ表情加解密
package crypter

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/zstd"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	emojiZeroID = 297
	emojiOneID  = 424
)

func encodeQQEmoji(text string) message.Message {
	if text == "" {
		return message.Message{message.Text("请输入要加密的文本")}
	}

	data := []byte(text)
	best, header := data, "0"
	if br := tryCompress(func(w io.Writer) io.WriteCloser { return brotli.NewWriterLevel(w, brotli.BestCompression) }, data); len(br) > 0 && len(br) < len(best) {
		best, header = br, "10"
	}
	if zs := tryCompress(func(w io.Writer) io.WriteCloser {
		enc, _ := zstd.NewWriter(w, zstd.WithEncoderLevel(zstd.SpeedBestCompression))
		return enc
	}, data); len(zs) > 0 && len(zs) < len(best) {
		best, header = zs, "11"
	}

	var bin strings.Builder
	bin.WriteString(header)
	for _, b := range best {
		fmt.Fprintf(&bin, "%08b", b)
	}

	s := bin.String()
	msg := make(message.Message, 0, len(s))
	for _, bit := range s {
		if bit == '0' {
			msg = append(msg, message.Face(emojiZeroID))
		} else {
			msg = append(msg, message.Face(emojiOneID))
		}
	}
	return msg
}

func decodeQQEmoji(faceIDs []int) string {
	var bin strings.Builder
	for _, id := range faceIDs {
		if id == emojiZeroID {
			bin.WriteByte('0')
		} else if id == emojiOneID {
			bin.WriteByte('1')
		}
	}
	binary := bin.String()
	if len(binary) < 2 {
		return "QQ表情密文格式错误"
	}

	var header int
	switch {
	case binary[:2] == "11":
		header = 2
	case binary[:2] == "10":
		header = 2
	case binary[0] == '0':
		header = 1
	default:
		return "QQ表情密文格式错误"
	}

	dataBin := binary[header:]
	if len(dataBin)%8 != 0 {
		return fmt.Sprintf("QQ表情解密失败：数据长度不正确（%d位）", len(dataBin))
	}

	data := make([]byte, len(dataBin)/8)
	for i := range data {
		for j := 0; j < 8; j++ {
			if dataBin[i*8+j] == '1' {
				data[i] |= 1 << (7 - j)
			}
		}
	}

	var out []byte
	var err error
	switch binary[:header] {
	case "0":
		out = data
	case "10":
		r := brotli.NewReader(bytes.NewReader(data))
		out, err = io.ReadAll(r)
	case "11":
		var dec *zstd.Decoder
		dec, err = zstd.NewReader(bytes.NewReader(data))
		if err == nil {
			out, err = io.ReadAll(dec)
			dec.Close()
		}
	}
	if err != nil {
		return fmt.Sprintf("QQ表情解压失败: %v", err)
	}
	if !utf8.Valid(out) {
		return "QQ表情解密失败：结果不是有效文本"
	}
	return string(out)
}

func tryCompress(newWriter func(io.Writer) io.WriteCloser, data []byte) []byte {
	var buf bytes.Buffer
	w := newWriter(&buf)
	if _, err := w.Write(data); err != nil {
		return nil
	}
	if err := w.Close(); err != nil {
		return nil
	}
	return buf.Bytes()
}
