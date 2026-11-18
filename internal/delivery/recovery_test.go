package delivery

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"wialon-sener-pilot/internal/application"
	"wialon-sener-pilot/internal/domain"
)

func TestR6LocateAndRepairInvalidPosition(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "pilot.db")
	open := func() *Store {
		s, err := Open(path, 42, "XAXX010101000")
		if err != nil {
			t.Fatal(err)
		}
		return s
	}
	s := open()
	from, to := time.Unix(100, 0), time.Unix(200, 0)
	record := domain.Input{Key: "fixture", Build: func(string) ([]byte, error) { return nil, errors.New("missing ignition") }}
	if err := s.SaveWindow(ctx, from, to, []domain.Input{record}); err != nil {
		t.Fatal(err)
	}
	s.Close()
	s = open()
	defer s.Close()
	var gotFrom, gotTo int64
	var uuid, reason string
	if err := s.db.QueryRow("SELECT window_from,window_to,id,detail FROM reports WHERE source_key=? AND status='invalid' AND body IS NULL", "fixture").Scan(&gotFrom, &gotTo, &uuid, &reason); err != nil {
		t.Fatal(err)
	}
	if gotFrom != from.Unix() || gotTo != to.Unix() || reason != "missing ignition" {
		t.Fatal("missing origin/validation details")
	}
	record.Build = func(id string) ([]byte, error) { return []byte(fmt.Sprintf(`{"id":%q}`, id)), nil }
	if err := s.SaveWindow(ctx, from, to, []domain.Input{record}); err != nil {
		t.Fatal(err)
	}
	ready, err := s.Claim(ctx, time.Now())
	if err != nil || ready == nil || ready.ID != uuid {
		t.Fatal("repair must preserve the transmission identity", err)
	}
}

func TestR5CredentialFailureHoldsQueue(t *testing.T) {
	ctx := context.Background()
	s, err := Open(filepath.Join(t.TempDir(), "pilot.db"), 42, "XAXX010101000")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	input := func(key string) domain.Input {
		return domain.Input{Key: key, Build: func(string) ([]byte, error) { return []byte("{}"), nil }}
	}
	if err = s.SaveWindow(ctx, time.Unix(0, 0), time.Unix(1, 0), []domain.Input{input("a"), input("b")}); err != nil {
		t.Fatal(err)
	}
	var original string
	if err = s.db.QueryRow("SELECT id FROM reports WHERE source_key='a'").Scan(&original); err != nil {
		t.Fatal(err)
	}
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls++; w.WriteHeader(http.StatusUnauthorized) }))
	defer server.Close()
	sender := Sender{URL: server.URL, Token: "fake", Client: server.Client()}
	if err = (application.Pilot{Outbox: s, Sender: sender, Profile: domain.Profile{RFC: "XAXX010101000"}}).DeliverPending(ctx); err != nil {
		t.Fatal(err)
	}
	if err = (application.Pilot{Outbox: s, Sender: sender, Profile: domain.Profile{RFC: "XAXX010101000"}}).DeliverPending(ctx); err != nil {
		t.Fatal(err)
	}
	counts, err := s.Counts(ctx)
	if calls != 1 || err != nil || counts[domain.Pending] != 2 || counts[domain.Rejected] != 0 {
		t.Fatalf("calls=%d counts=%v err=%v", calls, counts, err)
	}
	after, err := s.Claim(ctx, time.Now().Add(time.Hour))
	if err != nil || after == nil || after.ID != original {
		t.Fatal("first delivery identity changed", err)
	}
}
