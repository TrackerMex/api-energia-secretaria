package main

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"wialon-sener-pilot/internal/delivery"
	"wialon-sener-pilot/internal/domain"
	"wialon-sener-pilot/internal/historydb"
)

const uid = "DEMO-CR13"
const profilePath = "../../testdata/tcp-sener/profile.json"

// Independent wire vector: posinfo followed by int32 sensors ign=0 and online=1.
const wireFixture = "6200000044454d4f2d435231330068fb6a4000000003" + "0bbb000000270102706f73696e666f0000000000000059c000000000000034400000000000209c40002a005a09" + "0bbb0000000a010369676e0000000000" + "0bbb0000000d01036f6e6c696e650000000001"

var start = time.Date(2025, 10, 24, 12, 0, 0, 0, time.UTC)

type fixture struct{ history, outbox string }

func newFixture(t *testing.T) fixture {
	dir := t.TempDir()
	return fixture{filepath.Join(dir, "positions.db"), filepath.Join(dir, "queue.db")}
}
func (f fixture) args(endpoint string, from, to time.Time) []string {
	return []string{"-uid", uid, "-profile", profilePath, "-history", f.history, "-outbox", f.outbox, "-from", from.UTC().Format(time.RFC3339Nano), "-to", to.UTC().Format(time.RFC3339Nano), "-endpoint", endpoint}
}
func (f fixture) process(t *testing.T, server *httptest.Server, from, to time.Time) {
	t.Helper()
	if err := run(context.Background(), f.args(server.URL, from, to), io.Discard); err != nil {
		t.Fatal(err)
	}
}
func (f fixture) save(t *testing.T, sec int, hash byte, signals bool) {
	t.Helper()
	db, err := historydb.OpenWriter(f.history, uid)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	fix := domain.GPSFix{UID: uid, GPSTime: start.Add(time.Duration(sec) * time.Second), Latitude: 20, Longitude: -100, Speed: 42, Course: 90, Fingerprint: [32]byte{hash}}
	if signals {
		off, on := false, true
		fix.Ignition = &off
		fix.Connected = &on
		fix.IgnitionSource = "ign"
		fix.ConnectionSource = "online"
	}
	if err = db.Save(context.Background(), fix); err != nil {
		t.Fatal(err)
	}
}

type record struct{ id, status, detail string }

func (f fixture) records(t *testing.T) []record {
	t.Helper()
	db, e := sql.Open("sqlite", f.outbox)
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	rows, e := db.Query("SELECT id,status,detail FROM reports ORDER BY rowid")
	if e != nil {
		t.Fatal(e)
	}
	defer rows.Close()
	var records []record
	for rows.Next() {
		var r record
		if e = rows.Scan(&r.id, &r.status, &r.detail); e != nil {
			t.Fatal(e)
		}
		records = append(records, r)
	}
	if e = rows.Err(); e != nil {
		t.Fatal(e)
	}
	return records
}

type response struct {
	code int
	body string
}
type post struct {
	id        string
	timestamp string
}

func simulator(t *testing.T, replies []response) (*httptest.Server, func() []post) {
	t.Helper()
	var posts []post
	var mu sync.Mutex
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.Header.Get("Authorization") != "Bearer local-demo" {
			t.Error("unexpected method or authorization")
		}
		if e := r.ParseMultipartForm(1 << 20); e != nil {
			t.Error(e)
			return
		}
		defer r.MultipartForm.RemoveAll()
		f, h, e := r.FormFile("gps")
		if e != nil {
			t.Error(e)
			return
		}
		defer f.Close()
		if h.Filename != "XAXX010101000.json" || h.Header.Get("Content-Type") != "application/json" || len(r.MultipartForm.File) != 1 || len(r.MultipartForm.Value) != 0 {
			t.Error("unexpected multipart", h)
		}
		var body struct {
			RFC  string `json:"rfc"`
			Name string `json:"razon_social"`
			Data []struct {
				ID        string             `json:"id_transmision"`
				Timestamp string             `json:"timestamp_gps"`
				Type      string             `json:"tipo_reporte"`
				Vehicle   domain.Vehicle     `json:"vehiculo"`
				Location  map[string]float64 `json:"geolocalizacion"`
				Status    struct {
					Speed    float64 `json:"velocidad_kmh"`
					Ignition *bool   `json:"motor_encendido"`
				} `json:"estado_vehiculo"`
				Device struct {
					State string `json:"estado_conexion"`
					IMEI  string `json:"imei"`
				} `json:"estado_dispositivo_gps"`
			} `json:"data"`
		}
		if e = json.NewDecoder(f).Decode(&body); e != nil || len(body.Data) != 1 {
			t.Errorf("incomplete report: %v", e)
			return
		}
		d := body.Data[0]
		if body.RFC != "XAXX010101000" || body.Name != "DEMO" || len(d.ID) != 36 || d.Type != "posicion" || d.Vehicle.Unit != "CR-13" || d.Location["latitud"] != 20 || d.Location["longitud"] != -100 || d.Location["direccion_grados"] != 90 || d.Status.Speed != 42 || d.Status.Ignition == nil || *d.Status.Ignition || d.Device.State != "conectado" || d.Device.IMEI != "000000000000000" {
			t.Errorf("unexpected report: %+v", body)
		}
		mu.Lock()
		idx := len(posts)
		posts = append(posts, post{d.ID, d.Timestamp})
		mu.Unlock()
		reply := response{200, `{"error":0}`}
		if idx < len(replies) {
			reply = replies[idx]
		}
		if reply.code == 429 {
			w.Header().Set("Retry-After", "60")
		}
		if reply.code == 307 {
			w.Header().Set("Location", "/redirected")
		}
		w.WriteHeader(reply.code)
		io.WriteString(w, reply.body)
	})), func() []post { mu.Lock(); defer mu.Unlock(); return slices.Clone(posts) }
}

func TestR2MissingAndLegacySignals(t *testing.T) {
	server, posts := simulator(t, nil)
	defer server.Close()
	for _, legacy := range []bool{false, true} {
		f := newFixture(t)
		if legacy {
			db, e := sql.Open("sqlite", f.history)
			if e != nil {
				t.Fatal(e)
			}
			_, e = db.Exec(`CREATE TABLE receiver_unit(id INTEGER PRIMARY KEY,uid TEXT,alias TEXT);INSERT INTO receiver_unit VALUES(1,'DEMO-CR13','CR-13');
    CREATE TABLE gps_fixes(uid TEXT,fingerprint BLOB,gps_sec INTEGER,latitude REAL,longitude REAL,speed INTEGER,course INTEGER, PRIMARY KEY(uid,fingerprint));
    INSERT INTO gps_fixes VALUES('DEMO-CR13',zeroblob(32),?,20,-100,42,90)`, start.Unix())
			db.Close()
			if e != nil {
				t.Fatal(e)
			}
		} else {
			f.save(t, 0, 1, false)
		}
		f.process(t, server, start, start.Add(time.Second))
		rows := f.records(t)
		if len(posts()) != 0 || len(rows) != 1 || rows[0].status != domain.Invalid || rows[0].detail == "" {
			t.Fatal(posts(), rows)
		}
		hist, e := historydb.OpenReader(f.history, uid)
		if e != nil {
			t.Fatal(e)
		}
		items, e := hist.Between(context.Background(), uid, start, start.Add(time.Second))
		hist.Close()
		if e != nil || len(items) != 1 {
			t.Fatal(items, e)
		}
	}
}
func TestR2ChangedSelectorKeepsInvalidUntilCorrected(t *testing.T) {
	f := newFixture(t)
	f.save(t, 0, 1, true)
	server, posts := simulator(t, nil)
	defer server.Close()
	original, err := os.ReadFile(profilePath)
	if err != nil {
		t.Fatal(err)
	}
	changed := strings.Replace(string(original), `"ign"`, `"another_input"`, 1)
	path := filepath.Join(t.TempDir(), "changed.json")
	if err = os.WriteFile(path, []byte(changed), 0600); err != nil {
		t.Fatal(err)
	}
	args := f.args(server.URL, start, start.Add(time.Second))
	args[3] = path
	if err = run(context.Background(), args, io.Discard); err != nil {
		t.Fatal(err)
	}
	rows := f.records(t)
	if len(rows) != 1 || rows[0].status != domain.Invalid || !strings.Contains(rows[0].detail, "another_input") || len(posts()) != 0 {
		t.Fatal(rows, posts())
	}
	changed = strings.Replace(string(original), `"wialon_unit_id": 7,`, "", 1)
	if err = os.WriteFile(path, []byte(changed), 0600); err != nil {
		t.Fatal(err)
	}
	if err = run(context.Background(), args, io.Discard); err != nil {
		t.Fatal(err)
	}
	if got := f.records(t); len(got) != 1 || got[0].id != rows[0].id || got[0].status != domain.Accepted || len(posts()) != 1 {
		t.Fatal(got, posts())
	}
}
func TestR3ReprocessAndLateArrival(t *testing.T) {
	f := newFixture(t)
	server, posts := simulator(t, nil)
	defer server.Close()
	f.save(t, 0, 1, true)
	f.save(t, 0, 2, true)
	f.save(t, 1, 3, true)
	f.process(t, server, start, start.Add(time.Second))
	first := f.records(t)
	f.save(t, -1, 4, true)
	f.process(t, server, start.Add(-time.Second), start.Add(time.Second))
	f.save(t, 150, 5, true)
	f.process(t, server, start.Add(150*time.Second), start.Add(151*time.Second))
	rows := f.records(t)
	if len(first) != 2 || len(rows) != 4 || len(posts()) != 4 || rows[0] != first[0] || rows[1] != first[1] || first[0].id == first[1].id || (posts())[2].timestamp != start.Add(-time.Second).Format(time.RFC3339) {
		t.Fatal(rows, posts())
	}
}
func TestR4DeliveryOutcomes(t *testing.T) {
	for _, tc := range []struct {
		reply response
		want  string
	}{{response{429, ""}, domain.Pending}, {response{422, `{"error":1}`}, domain.Rejected}, {response{200, `{}`}, domain.Uncertain}, {response{307, ""}, domain.Uncertain}} {
		t.Run(fmt.Sprint(tc.reply.code), func(t *testing.T) {
			f := newFixture(t)
			f.save(t, 0, 1, true)
			server, posts := simulator(t, []response{tc.reply})
			defer server.Close()
			f.process(t, server, start, start.Add(time.Second))
			rows := f.records(t)
			if len(rows) != 1 || rows[0].status != tc.want || rows[0].detail == "" {
				t.Fatal(rows)
			}
			f.process(t, server, start, start.Add(time.Second))
			if len(posts()) != 1 {
				t.Fatal("premature or unsafe resend", posts())
			}
			if tc.want == domain.Pending {
				// Advance the stored deadline rather than sleeping for a minute.
				db, e := sql.Open("sqlite", f.outbox)
				if e != nil {
					t.Fatal(e)
				}
				_, e = db.Exec("UPDATE pilot SET cooldown=0; UPDATE reports SET retry_at=0")
				db.Close()
				if e != nil {
					t.Fatal(e)
				}
				f.process(t, server, start.Add(20*time.Second), start.Add(21*time.Second))
				if len(posts()) != 1 {
					t.Fatal("sent a pending report outside requested window")
				}
				f.process(t, server, start, start.Add(time.Second))
				if len(posts()) != 2 || (posts())[0].id != (posts())[1].id || f.records(t)[0].status != domain.Accepted {
					t.Fatal(posts(), f.records(t))
				}
			}
		})
	}
}
func TestR2RejectExternalEndpointAndForeignQueue(t *testing.T) {
	f := newFixture(t)
	f.save(t, 0, 1, true)
	for _, endpoint := range []string{"http://example.invalid/", "http://0.0.0.0:80", "http://localhost.evil.invalid", "http://user:password@127.0.0.1:80", "ftp://127.0.0.1/"} {
		if err := run(context.Background(), f.args(endpoint, start, start.Add(time.Second)), io.Discard); err == nil {
			t.Fatal(endpoint)
		}
	}
	remote, e := delivery.Open(f.outbox, 7, "XAXX010101000")
	if e != nil {
		t.Fatal(e)
	}
	remote.Close()
	server, posts := simulator(t, nil)
	defer server.Close()
	if err := run(context.Background(), f.args(server.URL, start, start.Add(time.Second)), io.Discard); err == nil || len(posts()) != 0 {
		t.Fatal("adopted Remote API outbox")
	}
	// Pre-integration queues lack a source marker and retain their Remote API identity.
	legacy, err := sql.Open("sqlite", f.outbox)
	if err != nil {
		t.Fatal(err)
	}
	_, err = legacy.Exec("DROP TABLE outbox_scope")
	legacy.Close()
	if err != nil {
		t.Fatal(err)
	}
	if err = run(context.Background(), f.args(server.URL, start, start.Add(time.Second)), io.Discard); err == nil || len(posts()) != 0 {
		t.Fatal("adopted a legacy remote outbox")
	}
	remote, err = delivery.Open(f.outbox, 7, "XAXX010101000")
	if err != nil {
		t.Fatal(err)
	}
	remote.Close()
	f.outbox = filepath.Join(t.TempDir(), "tcp.db")
	f.process(t, server, start, start.Add(time.Second))
	if wrong, err := delivery.OpenTCP(f.outbox, "XAXX010101000", "DEMO-OTHER"); err == nil {
		wrong.Close()
		t.Fatal("allowed another retranslator UID")
	}
	if remote, e = delivery.Open(f.outbox, 7, "XAXX010101000"); e == nil {
		remote.Close()
		t.Fatal("remote API opened a local simulator queue")
	}
}

func TestR5TCPToSENERLocal(t *testing.T) {
	f := newFixture(t)
	dir := t.TempDir()
	build := exec.Command("go", "build", "-o", dir+string(filepath.Separator), "../receptor", ".")
	build.Env = append(os.Environ(), "GOPROXY=off")
	if out, e := build.CombinedOutput(); e != nil {
		t.Fatalf("build: %v %s", e, out)
	}
	command := func(name string, args ...string) *exec.Cmd {
		if runtime.GOOS == "windows" {
			name += ".exe"
		}
		cmd := exec.Command(filepath.Join(dir, name), args...)
		cmd.Env = append(os.Environ(), "WIALON_TOKEN=", "SENER_DEV_TOKEN=")
		return cmd
	}
	wire, e := hex.DecodeString(wireFixture)
	if e != nil {
		t.Fatal(e)
	}
	server, posts := simulator(t, nil)
	defer server.Close()
	for cycle := 0; cycle < 2; cycle++ {
		child := command("receptor", "listen", "-listen", "127.0.0.1:0", "-db", f.history, "-uid", uid, "-profile", profilePath)
		r, e := child.StderrPipe()
		if e != nil {
			t.Fatal(e)
		}
		if e = child.Start(); e != nil {
			t.Fatal(e)
		}
		func() {
			defer func() { child.Process.Kill(); child.Wait() }()
			lines := make(chan string, 1)
			go func() { sc := bufio.NewScanner(r); sc.Scan(); lines <- sc.Text() }()
			var address string
			select {
			case line := <-lines:
				var label string
				fmt.Sscan(line, &label, &address)
				if label != "listening" {
					t.Fatal(line)
				}
			case <-time.After(10 * time.Second):
				t.Fatal("listener startup timeout")
			}
			conn, e := net.DialTimeout("tcp", address, 3*time.Second)
			if e != nil {
				t.Fatal(e)
			}
			defer conn.Close()
			conn.SetDeadline(time.Now().Add(3 * time.Second))
			if _, e = conn.Write(wire); e != nil {
				t.Fatal(e)
			}
			if e = conn.(*net.TCPConn).CloseWrite(); e != nil {
				t.Fatal(e)
			}
			if _, e = io.Copy(io.Discard, conn); e != nil {
				t.Fatal(e)
			}
		}()
		out, e := command("tcp-sener", f.args(server.URL, start, start.Add(time.Second))...).CombinedOutput()
		if e != nil || !bytes.Contains(out, []byte("accepted=1")) {
			t.Fatalf("command: %v %s", e, out)
		}
	}
	history, e := historydb.OpenReader(f.history, uid)
	if e != nil {
		t.Fatal(e)
	}
	defer history.Close()
	positions, e := history.Between(context.Background(), uid, start, start.Add(time.Second))
	if e != nil {
		t.Fatal(e)
	}
	rows := f.records(t)
	if len(positions) != 1 || len(rows) != 1 || rows[0].status != domain.Accepted || len(posts()) != 1 || (posts())[0].id != rows[0].id || (posts())[0].timestamp != start.Format(time.RFC3339) {
		t.Fatal(positions, rows, posts())
	}
}
