package ssh

import (
	"strings"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
)

func NormalizeCharset(name string) string {
	switch strings.ToLower(strings.TrimSpace(strings.ReplaceAll(name, "_", "-"))) {
	case "", "utf8", "utf-8":
		return "utf-8"
	case "gbk", "cp936":
		return "gbk"
	case "gb2312":
		return "gb2312"
	case "gb18030":
		return "gb18030"
	case "big5", "big5-hkscs":
		return "big5"
	default:
		return "utf-8"
	}
}

func EncodeFromUTF8(charset, text string) ([]byte, error) {
	src := []byte(text)
	switch NormalizeCharset(charset) {
	case "utf-8":
		return src, nil
	case "gbk", "gb2312":
		return simplifiedchinese.GBK.NewEncoder().Bytes(src)
	case "gb18030":
		return simplifiedchinese.GB18030.NewEncoder().Bytes(src)
	case "big5":
		return traditionalchinese.Big5.NewEncoder().Bytes(src)
	default:
		return src, nil
	}
}
