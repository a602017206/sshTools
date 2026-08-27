package ssh

import (
	"os"
	"testing"
)

func TestParseOctalFileMode(t *testing.T) {
	tests := []struct {
		in      string
		want    os.FileMode
		wantErr bool
	}{
		{in: "644", want: 0o644},
		{in: "0755", want: 0o755},
		{in: "777", want: 0o777},
		{in: "88", wantErr: true},
		{in: "12", wantErr: true},
		{in: "", wantErr: true},
	}

	for _, tt := range tests {
		got, err := parseOctalFileMode(tt.in)
		if tt.wantErr {
			if err == nil {
				t.Fatalf("parseOctalFileMode(%q) 期望失败", tt.in)
			}
			continue
		}
		if err != nil {
			t.Fatalf("parseOctalFileMode(%q) 失败: %v", tt.in, err)
		}
		if got != tt.want {
			t.Fatalf("parseOctalFileMode(%q) = %o, 期望 %o", tt.in, got, tt.want)
		}
	}
}
