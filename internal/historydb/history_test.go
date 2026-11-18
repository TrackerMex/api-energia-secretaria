package historydb

import (
	"context"
	"path/filepath"
	"sync"
	"testing"
	"time"
	"wialon-sener-pilot/internal/application"
	"wialon-sener-pilot/internal/domain"
)

func TestR3PersistAndDeduplicate(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "r.db")
	open := func() *Store {
		s, e := OpenWriter(path, "U")
		if e != nil {
			t.Fatal(e)
		}
		return s
	}
	s := open()
	app := application.Receiver{UID: "U", Archive: s}
	fix := domain.GPSFix{UID: "U", GPSTime: time.Unix(50, 0).UTC(), Latitude: 10, Longitude: 20, Speed: 7, Course: 1, Fingerprint: [32]byte{1}}
	other := fix
	other.Fingerprint[0] = 2
	other.Speed = 8
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if e := app.Accept(ctx, fix); e != nil {
				t.Error(e)
			}
		}()
	}
	wg.Wait()
	if e := app.Accept(ctx, other); e != nil {
		t.Fatal(e)
	}
	s.Close()
	s = open()
	defer s.Close()
	app.Archive = s
	if e := app.Accept(ctx, fix); e != nil {
		t.Fatal(e)
	}
	rows, e := app.Between(ctx, time.Unix(49, 0), time.Unix(51, 0))
	if e != nil || len(rows) != 2 || rows[0] != fix || rows[1] != other {
		t.Fatalf("rows=%+v err=%v", rows, e)
	}
	if wrong, e := OpenWriter(path, "X"); e == nil {
		wrong.Close()
		t.Fatal("changed database identity")
	}
}
func TestR4GPSOrderAndHalfOpenRanges(t *testing.T) {
	ctx := context.Background()
	s, e := OpenWriter(filepath.Join(t.TempDir(), "r.db"), "U")
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	app := application.Receiver{UID: "U", Archive: s}
	for _, sec := range []int64{100, 80, 90} {
		if e := app.Accept(ctx, domain.GPSFix{UID: "U", GPSTime: time.Unix(sec, 0).UTC(), Fingerprint: [32]byte{byte(sec)}}); e != nil {
			t.Fatal(e)
		}
	}
	latest, e := app.Latest(ctx)
	if e != nil || latest == nil || latest.GPSTime.Unix() != 100 {
		t.Fatal(latest, e)
	}
	rows, e := app.Between(ctx, time.Unix(80, 1), time.Unix(100, 0))
	if e != nil || len(rows) != 1 || rows[0].GPSTime.Unix() != 90 {
		t.Fatal(rows, e)
	}
	if rows, e := app.Between(ctx, time.Unix(100, 1), time.Unix(101, 0)); e != nil || len(rows) != 0 {
		t.Fatal(rows, e)
	}
}
