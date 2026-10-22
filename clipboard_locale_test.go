package main

import "testing"

func TestUTF8LocaleOverride(t *testing.T) {
	cases := []struct {
		name      string
		env       map[string]string
		wantKey   string
		wantValue string
		wantOK    bool
	}{
		{name: "Finder 启动无 locale", env: map[string]string{}, wantKey: "LC_CTYPE", wantValue: "UTF-8", wantOK: true},
		{name: "LANG 已是 UTF-8", env: map[string]string{"LANG": "zh_CN.UTF-8"}},
		{name: "LC_CTYPE 小写 utf8", env: map[string]string{"LC_CTYPE": "en_US.utf8"}},
		{name: "LANG 为 C", env: map[string]string{"LANG": "C"}, wantKey: "LC_CTYPE", wantValue: "UTF-8", wantOK: true},
		{name: "LC_ALL 覆盖为 C", env: map[string]string{"LC_ALL": "C", "LANG": "en_US.UTF-8"}, wantKey: "LC_ALL", wantValue: "en_US.UTF-8", wantOK: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			key, value, ok := utf8LocaleOverride(func(k string) string { return tc.env[k] })
			if key != tc.wantKey || value != tc.wantValue || ok != tc.wantOK {
				t.Fatalf("got (%q, %q, %v), want (%q, %q, %v)", key, value, ok, tc.wantKey, tc.wantValue, tc.wantOK)
			}
		})
	}
}
