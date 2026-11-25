package unitprofile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestR1TCPProfileBindsUIDAndSelectors(t *testing.T) {
	const uid = "DEMO-CR13"
	path := filepath.Join("..", "..", "testdata", "tcp-sener", "profile.json")
	p, err := LoadTCP(path, uid)
	if err != nil || p.IgnitionParam != "ign" || p.ConnectionParam != "online" {
		t.Fatalf("fixture: %v", err)
	}
	if _, err := LoadTCP(path, "OTHER"); err == nil {
		t.Fatal("accepted other UID")
	}
	for _, invalid := range []string{"", "online", "digital:in0", "digital:in33", "digital:in01", "digital:out1", "posinfo"} {
		p.IgnitionParam = invalid
		raw, err := json.Marshal(p)
		if err != nil {
			t.Fatal(err)
		}
		candidate := filepath.Join(t.TempDir(), "profile.json")
		if err := os.WriteFile(candidate, raw, 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadTCP(candidate, uid); err == nil {
			t.Fatalf("accepted selector %q", invalid)
		}
	}
}
