package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"wialon-sener-pilot/internal/delivery"
	"wialon-sener-pilot/internal/ingest"
)

func TestR7SingleUnitEndToEnd(t *testing.T) {
	received := 0
	const ts = 1761307200
	wialon := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch service := r.FormValue("svc"); service {
		case "token/login":
			fmt.Fprint(w, `{"eid":"fixture-session","user":{"id":1}}`)
		case "messages/load_interval":
			var params struct {
				ItemID int64 `json:"itemId"`
			}
			if err := json.Unmarshal([]byte(r.FormValue("params")), &params); err != nil || params.ItemID != 7 {
				t.Error("wrong unit", err)
			}
			fmt.Fprint(w, `{"count":1,"messages":[{"t":1761307200,"f":3,"tp":"ud","pos":{"x":-100,"y":20,"s":0,"c":90,"sc":9},"p":{"ign":1,"online":1}}]}`)
		case "core/logout", "messages/unload":
			fmt.Fprint(w, `{}`)
		default:
			t.Errorf("unexpected Wialon service %q", service)
			http.Error(w, "unknown service", 400)
		}
	}))
	defer wialon.Close()
	sener := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		file, _, err := r.FormFile("gps")
		if err != nil {
			t.Error(err)
			return
		}
		defer file.Close()
		var body struct {
			Data []struct {
				Kind string `json:"tipo_reporte"`
			}
		}
		if err := json.NewDecoder(file).Decode(&body); err != nil || len(body.Data) != 1 || body.Data[0].Kind != "posicion" {
			t.Error("unexpected SENER payload", err)
			return
		}
		received++
		fmt.Fprint(w, `{"error":0}`)
	}))
	defer sener.Close()
	profile, err := loadProfile("../../internal/ingest/profile.example.json")
	if err != nil {
		t.Fatal(err)
	}
	store, err := delivery.Open(filepath.Join(t.TempDir(), "pilot.db"), 7, profile.RFC)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	source := ingest.Source{BaseURL: wialon.URL, Token: "read-only-fixture", Client: wialon.Client()}
	sender := delivery.Sender{URL: sener.URL, Token: "fixture", Client: sener.Client()}
	for i := 0; i < 2; i++ {
		if err := runWindow(context.Background(), store, source, sender, 7, profile, time.Unix(ts-1, 0), time.Unix(ts+1, 0)); err != nil {
			t.Fatal(err)
		}
	}
	if received != 1 {
		t.Fatalf("delivered %d times", received)
	}
}
