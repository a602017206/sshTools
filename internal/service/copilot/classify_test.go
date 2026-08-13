package copilot

import "testing"

func TestClassifyMarksUnqualifiedUpdate(t *testing.T) {
	got := Classify("sql", "UPDATE orders SET status=1")
	if !got.Destructive {
		t.Fatal("expected destructive")
	}
}

func TestClassifyMarksDropTable(t *testing.T) {
	got := Classify("sql", "DROP TABLE users")
	if !got.Destructive {
		t.Fatal("expected destructive")
	}
}

func TestClassifyMarksDeleteFrom(t *testing.T) {
	got := Classify("sql", "DELETE FROM t")
	if !got.Destructive {
		t.Fatal("expected destructive")
	}
}

func TestClassifyMarksTruncate(t *testing.T) {
	got := Classify("sql", "TRUNCATE TABLE logs")
	if !got.Destructive {
		t.Fatal("expected destructive")
	}
}

func TestClassifyMarksRmRf(t *testing.T) {
	got := Classify("shell", "rm -rf /")
	if !got.Destructive {
		t.Fatal("expected destructive")
	}
}

func TestClassifySelectIsSafe(t *testing.T) {
	got := Classify("sql", "SELECT 1")
	if got.Destructive {
		t.Fatal("SELECT 1 must not be destructive")
	}
}

func TestClassifyLsIsSafe(t *testing.T) {
	got := Classify("shell", "ls -la")
	if got.Destructive {
		t.Fatal("ls -la must not be destructive")
	}
}

func TestClassifyMarksDestructiveShellCommands(t *testing.T) {
	cmds := []string{
		"rm /tmp/x",
		"mkfs.ext4 /dev/sdb",
		"dd if=/dev/zero of=/dev/sda",
		"shutdown -h now",
		"reboot",
		"kill -9 1",
		"chmod 777 /tmp",
		">/dev/sda",
		"> /dev/sdb1",
	}
	for _, cmd := range cmds {
		got := Classify("shell", cmd)
		if !got.Destructive {
			t.Errorf("Classify(shell, %q) Destructive=false, want true", cmd)
		}
	}
}
