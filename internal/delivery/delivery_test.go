package delivery

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"wialon-sener-pilot/internal/domain"
)

func TestR4Multipart(t *testing.T) {
	payload := []byte("{\"razon_social\":\"Demostración\"}")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.Header.Get("Authorization") != "Bearer ficticio" || r.Header.Get("Accept") != "application/json" {
			t.Error("unexpected request headers or method")
		}
		reader, err := r.MultipartReader()
		if err != nil {
			t.Error(err)
			return
		}
		part, err := reader.NextPart()
		if err != nil {
			t.Error(err)
			return
		}
		body, err := io.ReadAll(part)
		if err != nil || part.FormName() != "gps" || part.FileName() != "XAXX010101000.json" || part.Header.Get("Content-Type") != "application/json" || string(body) != string(payload) {
			t.Error("invalid multipart attachment")
		}
		if _, err = reader.NextPart(); err != io.EOF {
			t.Error("expected a single attachment")
		}
		fmt.Fprint(w, `{"error":0}`)
	}))
	defer server.Close()
	got := (Sender{URL: server.URL, Token: "ficticio", Client: server.Client()}).Send(context.Background(), "XAXX010101000", payload)
	if got.State != domain.Accepted {
		t.Fatalf("outcome = %+v", got)
	}
}

func TestR5Classify(t *testing.T) {
	for _, tc := range []struct {
		status int
		body   string
		want   string
	}{
		{200, `{"error":0}`, domain.Accepted}, {200, `{}`, domain.Uncertain}, {200, `{"error":"0"}`, domain.Uncertain},
		{200, `{"error":1,"code_error":"INV","error_description":"secret-value"}`, domain.Rejected},
		{401, `{}`, domain.Pending}, {422, `{}`, domain.Rejected}, {429, `{}`, domain.Pending}, {500, `{}`, domain.Uncertain}, {200, `invalid`, domain.Uncertain},
	} {
		t.Run(fmt.Sprintf("%d-%s", tc.status, tc.body), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Retry-After", "120")
				w.WriteHeader(tc.status)
				fmt.Fprint(w, tc.body)
			}))
			defer server.Close()
			now := time.Now()
			got := (Sender{URL: server.URL, Token: "fake", Client: server.Client()}).Send(context.Background(), "XAXX010101000", []byte("{}"))
			if got.State != tc.want || strings.Contains(got.Detail, "secret-value") {
				t.Fatalf("outcome = %+v", got)
			}
			if tc.status == 200 && strings.Contains(tc.body, "INV") && !strings.Contains(got.Detail, "code=INV") {
				t.Fatal("missing application error code")
			}
			if tc.status == 429 && got.RetryAt.Before(now.Add(119*time.Second)) {
				t.Fatal("Retry-After ignored")
			}
		})
	}
}

func TestR6RestartAndCursor(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "private", "pilot.db")
	const rfc = "XAXX010101000"
	open := func() *Store {
		s, err := Open(path, 42, rfc)
		if err != nil {
			t.Fatal(err)
		}
		return s
	}
	s := open()
	from, to := time.Unix(10, 0), time.Unix(20, 0)
	makeInput := func(key string) domain.Input {
		return domain.Input{Key: key, Build: func(id string) ([]byte, error) { return json.Marshal(map[string]string{"uuid": id}) }}
	}
	if err := s.SaveWindow(ctx, from, to, []domain.Input{makeInput("good"), {Key: "", Build: func(string) ([]byte, error) { return nil, nil }}}); err == nil {
		t.Fatal("expected rollback")
	}
	if cursor, err := s.Cursor(ctx); err != nil || !cursor.IsZero() {
		t.Fatalf("uncommitted cursor: %v %v", cursor, err)
	}
	if err := s.SaveWindow(ctx, from, to, []domain.Input{makeInput("good"), makeInput("interrupted")}); err != nil {
		t.Fatal(err)
	}
	first, err := s.Claim(ctx, time.Now())
	if err != nil || first == nil {
		t.Fatal(first, err)
	}
	if ok := regexp.MustCompile("^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$").MatchString(first.ID); !ok {
		t.Fatal("invalid UUID")
	}
	if err = s.Finish(ctx, first.ID, domain.Outcome{State: domain.Pending, RetryAt: time.Now().Add(5 * time.Minute)}); err != nil {
		t.Fatal(err)
	}
	if retry, err := s.Claim(ctx, time.Now()); err != nil || retry != nil {
		t.Fatalf("expected cooldown: %v %v", retry, err)
	}
	later := time.Now().Add(time.Hour)
	retry, err := s.Claim(ctx, later)
	if err != nil || retry == nil || retry.ID != first.ID {
		t.Fatal("retry changed id", err)
	}
	if err = s.Finish(ctx, retry.ID, domain.Outcome{State: domain.Accepted}); err != nil {
		t.Fatal(err)
	}
	crash, err := s.Claim(ctx, later)
	if err != nil || crash == nil {
		t.Fatal("missing second record", err)
	}
	s.Close()
	s = open()
	defer s.Close()
	if err := s.SaveWindow(ctx, from, to, []domain.Input{makeInput("good"), makeInput("interrupted")}); err != nil {
		t.Fatal(err)
	}
	if again, err := s.Claim(ctx, later); err != nil || again != nil {
		t.Fatalf("resent completed/uncertain rows: %v %v", again, err)
	}
	if cursor, err := s.Cursor(ctx); err != nil || !cursor.Equal(to) {
		t.Fatal("missing checkpoint", cursor, err)
	}
	counts, err := s.Counts(ctx)
	if err != nil || counts[domain.Accepted] != 1 || counts[domain.Uncertain] != 1 {
		t.Fatal(counts, err)
	}
	if wrong, err := Open(path, 99, rfc); err == nil {
		wrong.Close()
		t.Fatal("reused a different unit's database")
	}
}
