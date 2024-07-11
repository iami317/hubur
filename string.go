package hubur

import (
	"crypto/md5"
	cryptoSha256 "crypto/sha256"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// string 是否是 prefix 中的任意一个开头的
// 判断 string 是否是一个 prefix 开头的，使用 strings.HasPrefix 就可以了
func StringHasPrefix(s string, prefix []string) bool {
	for _, x := range prefix {
		if strings.HasPrefix(s, x) {
			return true
		}
	}
	return false
}

func ChunkString(s string, chunkSize int) []string {
	var chunks []string
	runes := []rune(s)

	if len(runes) == 0 {
		return []string{s}
	}

	for i := 0; i < len(runes); i += chunkSize {
		nn := i + chunkSize
		if nn > len(runes) {
			nn = len(runes)
		}
		chunks = append(chunks, string(runes[i:nn]))
	}
	return chunks
}

func Sha256(s []byte) string {
	c := cryptoSha256.New()
	_, _ = c.Write(s)
	ret := c.Sum(nil)
	return fmt.Sprintf("%x", ret)
}

func Sha256String(s string) string {
	return Sha256([]byte(s))
}

func MD5String(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}

func EscapeInvalidUTF8Byte(s []byte) string {
	// 将非法的 utf8 序列中的字符转换为 `\x` 的模式
	// 注意，这个操作返回的结果和原始字符串是非等价的
	// 详见测试 TestEscapeInvalidUTF8Byte
	ret := make([]rune, 0, len(s)+20)
	start := 0
	for {
		r, size := utf8.DecodeRune(s[start:])
		if r == utf8.RuneError {
			// 说明是空的
			if size == 0 {
				break
			} else {
				// 不是 rune
				ret = append(ret, []rune(fmt.Sprintf("\\x%02x", s[start]))...)
			}
		} else {
			// 不是换行之类的控制字符
			if unicode.IsControl(r) && !unicode.IsSpace(r) {
				ret = append(ret, []rune(fmt.Sprintf("\\x%02x", r))...)
			} else {
				// 正常字符
				ret = append(ret, r)
			}
		}
		start += size
	}
	return string(ret)
}
