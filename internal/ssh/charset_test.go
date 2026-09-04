package ssh

import (
	"bytes"
	"testing"
)

func TestEncodeFromUTF8GBK(t *testing.T) {
	got, err := EncodeFromUTF8("gbk", "你好")
	if err != nil {
		t.Fatal(err)
	}
	want := []byte{0xc4, 0xe3, 0xba, 0xc3}
	if !bytes.Equal(got, want) {
		t.Fatalf("got %x, want %x", got, want)
	}
}

func TestEncodeFromUTF8DefaultIsUTF8(t *testing.T) {
	got, err := EncodeFromUTF8("", "你好")
	if err != nil {
		t.Fatal(err)
	}
	want := []byte("你好")
	if !bytes.Equal(got, want) {
		t.Fatalf("got %x, want %x", got, want)
	}
}

func TestNormalizeCharsetAliases(t *testing.T) {
	if got := NormalizeCharset("GB2312"); got != "gb2312" {
		t.Fatalf("got %q", got)
	}
	if got := NormalizeCharset("unknown"); got != "utf-8" {
		t.Fatalf("got %q", got)
	}
}
