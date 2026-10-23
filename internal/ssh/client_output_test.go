package ssh

import "testing"

func TestCombinedOutputRequiresConnection(t *testing.T) {
	c, err := NewClient(&Config{Host: "127.0.0.1", Port: 22, User: "u", Password: "p"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.CombinedOutput("true"); err == nil {
		t.Fatal("expected not-connected error")
	}
}
