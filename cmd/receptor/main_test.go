package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestR1LocalOnly(t *testing.T) {
	for _, tc := range []struct{ addr, uid string }{{":0", "U"}, {"0.0.0.0:0", "U"}, {"[::]:0", "U"}, {"192.0.2.1:0", "U"}, {"127.0.0.1:0", " "}} {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		db := filepath.Join(t.TempDir(), "r.db")
		e := run(ctx, []string{"listen", "-listen", tc.addr, "-db", db, "-uid", tc.uid}, io.Discard, io.Discard)
		cancel()
		if e == nil {
			t.Fatalf("accepted %q %q", tc.addr, tc.uid)
		}
		if _, e = os.Stat(db); !os.IsNotExist(e) {
			t.Fatal("created database before parameter validation")
		}
	}
}

// Literal test vector, independent of the TCP parser. lon=-100, lat=20, altitude=1800, speed=42, course=90.
const fixtureHex = "3f000000" + "44454d4f2d4352313300" + "68fb6a4000000001" + "0bbb00000027" + "0102706f73696e666f00" + "00000000000059c000000000000034400000000000209c40" + "002a005a09"

func TestR6TCPHistoryLocal(t *testing.T) {
	for _, n := range []string{"WIALON_TOKEN", "SENER_DEV_TOKEN"} {
		t.Setenv(n, "")
	}
	db := filepath.Join(t.TempDir(), "r.db")
	uid := "DEMO-CR13"
	start := func() (string, func()) {
		ctx, cancel := context.WithCancel(context.Background())
		r, w := io.Pipe()
		ready := make(chan string, 1)
		done := make(chan error, 1)
		go func() {
			var label, addr string
			fmt.Fscan(r, &label, &addr)
			ready <- addr
			io.Copy(io.Discard, r)
		}()
		go func() {
			done <- run(ctx, []string{"listen", "-uid", uid, "-db", db, "-listen", "127.0.0.1:0"}, io.Discard, w)
			w.Close()
		}()
		t.Cleanup(cancel)
		select {
		case addr := <-ready:
			if addr == "" {
				cancel()
				t.Fatalf("listen: %v", <-done)
			}
			return addr, func() {
				cancel()
				select {
				case e := <-done:
					if e != nil {
						t.Fatal(e)
					}
				case <-time.After(5 * time.Second):
					t.Fatal("listener did not stop")
				}
				r.Close()
			}
		case <-time.After(5 * time.Second):
			cancel()
			r.Close()
			t.Fatal("listener did not announce itself")
		}
		return "", nil
	}
	send := func(addr string, frame []byte) {
		c, e := net.DialTimeout("tcp", addr, 3*time.Second)
		if e != nil {
			t.Fatal(e)
		}
		defer c.Close()
		c.SetDeadline(time.Now().Add(3 * time.Second))
		if _, e = c.Write(frame); e != nil {
			t.Fatal(e)
		}
		if err := c.(*net.TCPConn).CloseWrite(); err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(c)
		if err != nil || len(data) != 0 {
			t.Fatalf("respuesta inesperada: %x, err=%v", data, err)
		}
	}
	type pos struct {
		Unit   string  `json:"unit"`
		UID    string  `json:"uid"`
		GPS    string  `json:"gps_utc"`
		Lat    float64 `json:"latitude"`
		Lon    float64 `json:"longitude"`
		Speed  uint16  `json:"speed"`
		Course uint16  `json:"course"`
	}
	query := func(mode string, extra ...string) []pos {
		var out bytes.Buffer
		args := append([]string{mode, "-uid", uid, "-db", db}, extra...)
		if e := run(context.Background(), args, &out, io.Discard); e != nil {
			t.Fatal(e)
		}
		var p []pos
		if e := json.Unmarshal(out.Bytes(), &p); e != nil {
			t.Fatal(e)
		}
		return p
	}
	packet, e := hex.DecodeString(fixtureHex)
	if e != nil {
		t.Fatal(e)
	}
	addr, stop := start()
	wrong := bytes.Replace(packet, []byte(uid), []byte("DEMO-NOPE"), 1)
	send(addr, wrong)
	if n := len(query("latest")); n != 0 {
		t.Fatalf("saved %d unauthorized positions", n)
	}
	send(addr, packet)
	want := pos{"CR-13", uid, time.Unix(1761307200, 0).UTC().Format(time.RFC3339), 20, -100, 42, 90}
	check := func(p []pos) {
		if len(p) != 1 || p[0] != want {
			t.Fatalf("got %+v want %+v", p, want)
		}
	}
	check(query("latest"))
	// Next GPS second, flags zero and no data blocks, followed by an incomplete frame.
	noFix, err := hex.DecodeString("1200000044454d4f2d435231330068fb6a4100000000")
	if err != nil {
		t.Fatal(err)
	}
	send(addr, append(noFix, packet[:8]...))
	check(query("latest"))
	stop()
	check(query("latest"))
	addr, stop = start()
	send(addr, packet)
	check(query("history", "-from", want.GPS, "-to", "2025-10-25T00:00:00Z"))
	stop()
	if rows := query("history", "-from", "2025-10-23T00:00:00Z", "-to", want.GPS); len(rows) != 0 {
		t.Fatal("interval must exclude upper bound")
	}
	var out bytes.Buffer
	if err := run(context.Background(), []string{"latest", "-uid", "DEMO-NOPE", "-db", db}, &out, io.Discard); err == nil || out.Len() != 0 {
		t.Fatal("different configured UID exposed data")
	}
}
