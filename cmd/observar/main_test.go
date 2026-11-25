package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

const (
	firstSecond = "1970-01-01T00:01:40Z"
	lastSecond  = "1970-01-01T00:01:45Z"
	unit        = "73450219"
	token       = "sintetico-no-mostrar"
	session     = "sesion-ficticia-no-mostrar"
	physical    = "DEMO-IMEI-PRIVATE"
	longitude   = "-101.678902"
)

func environment(t *testing.T, origin string) func(string) string {
	t.Helper()
	return func(name string) string {
		switch name {
		case "WIALON_UNIT_ID":
			return unit
		case "WIALON_BASE_URL":
			return origin
		case "WIALON_TOKEN":
			return token
		default:
			t.Errorf("lectura inesperada de variable %q", name)
			return ""
		}
	}
}

func noSensitive(t *testing.T, text string) {
	t.Helper()
	for _, private := range []string{unit, token, session, physical, longitude, "pos\"", "imei"} {
		if strings.Contains(text, private) {
			t.Fatalf("la salida reveló un dato privado del ensayo")
		}
	}
}

func position(second, flags int, signals string) string {
	return fmt.Sprintf(`{"t":%d,"f":%d,"tp":"ud","pos":{"x":%s,"y":20.1,"s":0,"c":0},"imei":"%s"%s}`, second, flags, longitude, physical, signals)
}

type fixture struct {
	*httptest.Server
	mu     sync.Mutex
	called []string
}

func (f *fixture) services() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.called)
}
func wialon(t *testing.T, first, page string, last int) *fixture {
	t.Helper()
	fake := &fixture{}
	fake.Server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		svc := r.FormValue("svc")
		fake.mu.Lock()
		fake.called = append(fake.called, svc)
		fake.mu.Unlock()
		if r.Method != "POST" || r.URL.Path != "/wialon/ajax.html" || r.URL.RawQuery != "" {
			t.Errorf("petición inesperada %s %s", r.Method, r.URL.Path)
		}
		if svc != "token/login" && r.FormValue("sid") != session {
			t.Error("falta sesión dedicada")
		}
		switch svc {
		case "token/login":
			var p struct {
				Token string `json:"token"`
			}
			if json.Unmarshal([]byte(r.FormValue("params")), &p) != nil || p.Token != token {
				t.Error("credencial de simulación incorrecta")
			}
			fmt.Fprintf(w, `{"eid":%q}`, session)
		case "messages/load_interval":
			var p struct {
				Item  int `json:"itemId"`
				From  int `json:"timeFrom"`
				To    int `json:"timeTo"`
				Flags int `json:"flags"`
				Mask  int `json:"flagsMask"`
			}
			if json.Unmarshal([]byte(r.FormValue("params")), &p) != nil || p.Item != 73450219 || p.From != 100 || p.To != last || p.Flags != 1 || p.Mask != 0xff01 {
				t.Error("filtros de unidad, posición o ventana inclusiva incorrectos")
			}
			io.WriteString(w, first)
		case "messages/get_messages":
			io.WriteString(w, page)
		case "core/logout":
			io.WriteString(w, `{"error":0}`)
		default:
			t.Errorf("operación inesperada %q", svc)
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	t.Cleanup(fake.Close)
	return fake
}

func TestR4ObservationReadOnly(t *testing.T) {
	first := `{"count":4,"messages":[` + position(100, 3, `,"i":1,"p":{"can_rpm":7300}`) + "," + position(101, 3, `,"i":0,"p":{"can_rpm":0}`) + "]}"
	page := "[" + position(102, 3, `,"i":0,"p":{"can_rpm":50}`) + "," + position(105, 1, "") + "]"
	fake := wialon(t, first, page, 105)
	var out bytes.Buffer
	err := run(context.Background(), []string{"-from", firstSecond, "-to", lastSecond}, environment(t, fake.URL), fake.Client(), &out)
	if err != nil {
		t.Fatal(err)
	}
	noSensitive(t, out.String())
	var got map[string]any
	if err = json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	reasons := map[string]any{"discrepancia": float64(1), "entrada_ausente": float64(1), "rpm_ausentes": float64(1)}
	if len(got) != 7 || got["total"] != float64(4) || got["encendido"] != float64(1) || got["apagado"] != float64(1) || got["indeterminado"] != float64(2) || got["conexion_gps"] != "pendiente_de_validacion" || !reflect.DeepEqual(got["motivos"], reasons) {
		t.Fatalf("resumen inesperado: %v", got)
	}
	if rule, ok := got["criterio_motivos"].(string); !ok || rule == "" {
		t.Error("falta criterio de conteo de motivos")
	}
	want := []string{"token/login", "messages/load_interval", "messages/get_messages", "core/logout"}
	if !slices.Equal(fake.services(), want) {
		t.Fatal(fake.services())
	}
}
func TestR1InclusiveSecondAndFailedPage(t *testing.T) {
	for _, tc := range []struct {
		name, first, page string
		fails             bool
	}{
		{"vacía", `{"count":0,"messages":[]}`, "", false},
		{"página fallida", `{"count":2,"messages":[` + position(100, 3, `,"i":1,"p":{"can_rpm":7300}`) + "]}", `{"error":7,"detail":"` + physical + " " + token + `"}`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fake := wialon(t, tc.first, tc.page, 100)
			var out bytes.Buffer
			err := run(context.Background(), []string{"-from", firstSecond, "-to", firstSecond}, environment(t, fake.URL), fake.Client(), &out)
			services := fake.services()
			if len(services) < 3 || services[len(services)-1] != "core/logout" {
				t.Fatal("sesión no cerrada", services)
			}
			if tc.fails {
				if err == nil || out.Len() != 0 {
					t.Fatal("se esperaba error sin resumen parcial")
				}
				noSensitive(t, err.Error())
				return
			}
			var summary map[string]any
			if err != nil || json.Unmarshal(out.Bytes(), &summary) != nil {
				t.Fatalf("ventana vacía: %v", err)
			}
			for _, k := range []string{"total", "encendido", "apagado", "indeterminado"} {
				if summary[k] != float64(0) {
					t.Fatal(summary)
				}
			}
			if reasons, ok := summary["motivos"].(map[string]any); !ok || len(reasons) != 0 || summary["conexion_gps"] != "pendiente_de_validacion" {
				t.Fatal(summary)
			}
		})
	}
}
func TestR3PrivateFailuresAndRedirect(t *testing.T) {
	var received atomic.Int32
	second := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { received.Add(1) }))
	defer second.Close()
	var firstHits atomic.Int32
	first := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		firstHits.Add(1)
		w.Header().Set("Location", second.URL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}))
	defer first.Close()
	var out bytes.Buffer
	args := []string{"-from", firstSecond, "-to", lastSecond}
	err := run(context.Background(), args, environment(t, first.URL), first.Client(), &out)
	if err == nil || out.Len() != 0 || received.Load() != 0 {
		t.Fatal("se siguió una redirección o se devolvió un resumen")
	}
	noSensitive(t, err.Error())
	for _, bad := range [][]string{{"-" + token}, {"-from", physical, "-to", lastSecond}, {"-from", lastSecond, "-to", firstSecond}, {"-from", "1970-01-01T00:01:40.5Z", "-to", lastSecond}, {"-from", "1970-01-01T01:01:40+01:00", "-to", lastSecond}, {"-from", firstSecond}} {
		if err = run(context.Background(), bad, environment(t, first.URL), first.Client(), &out); err == nil || out.Len() != 0 {
			t.Fatal("configuración inválida aceptada")
		}
		noSensitive(t, err.Error())
	}
	if firstHits.Load() != 1 {
		t.Fatal("configuración inválida abrió sesión")
	}
	for _, field := range []string{"WIALON_UNIT_ID", "WIALON_TOKEN"} {
		cfg := environment(t, first.URL)
		incomplete := func(key string) string {
			if key == field {
				return ""
			}
			return cfg(key)
		}
		if err = run(context.Background(), args, incomplete, first.Client(), &out); err == nil || firstHits.Load() != 1 {
			t.Fatal("configuración incompleta abrió sesión", field)
		}
	}
	// Accidental plaintext origin must fail before starting a token session.
	var plainHits atomic.Int32
	plain := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { plainHits.Add(1) }))
	defer plain.Close()
	for _, origin := range []string{plain.URL, first.URL + "/untrusted/" + token, "https://" + token + "@127.0.0.1"} {
		if err = run(context.Background(), args, environment(t, origin), plain.Client(), &out); err == nil || plainHits.Load() != 0 {
			t.Fatal("se usó un origen no válido")
		}
		noSensitive(t, err.Error())
	}
}
