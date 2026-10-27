package ssh

import "testing"

func TestStartPTYCommandRequiresConnection(t *testing.T) {
	c, err := NewClient(&Config{Host: "127.0.0.1", Port: 22, User: "u", Password: "p"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.StartPTYCommand("docker exec -i -t x /bin/sh", 80, 24); err == nil {
		t.Fatal("expected not-connected error")
	}
}
