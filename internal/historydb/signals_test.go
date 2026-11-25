package historydb

import (
	"context"
	"database/sql"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"wialon-sener-pilot/internal/domain"
)

func TestR1R2LegacyReaderThenMigrationPreservesUnknown(t *testing.T) {
	const uid = "DEMO-CR13"
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "hist.db")
	old, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = old.Exec(`CREATE TABLE receiver_unit(id INTEGER PRIMARY KEY, uid TEXT UNIQUE, alias TEXT);
CREATE TABLE gps_fixes(uid TEXT NOT NULL, fingerprint BLOB NOT NULL, gps_sec INTEGER NOT NULL, latitude REAL NOT NULL, longitude REAL NOT NULL, speed INTEGER NOT NULL, course INTEGER NOT NULL, PRIMARY KEY(uid,fingerprint));
INSERT INTO receiver_unit VALUES(1,'DEMO-CR13','CR-13');
INSERT INTO gps_fixes VALUES('DEMO-CR13',zeroblob(32),10,0,0,0,0);`)
	if err != nil {
		t.Fatal(err)
	}
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}
	reader, err := OpenReader(path, uid)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	latest, err := reader.Latest(ctx, uid)
	if err != nil || latest == nil || latest.Ignition != nil || latest.Connected != nil || latest.IgnitionSource != "" {
		t.Fatalf("legacy latest=%+v err=%v", latest, err)
	}
	rows, err := reader.Between(ctx, uid, time.Unix(0, 0), time.Unix(20, 0))
	if err != nil || len(rows) != 1 || rows[0].ConnectionSource != "" {
		t.Fatalf("legacy history=%+v err=%v", rows, err)
	}
	writer, err := OpenWriter(path, uid)
	if err != nil {
		t.Fatal(err)
	}
	off := false
	fresh := domain.GPSFix{UID: uid, GPSTime: time.Unix(11, 0), Fingerprint: [32]byte{1}, Ignition: &off, IgnitionSource: "ign", ConnectionSource: "online"}
	if err := writer.Save(ctx, fresh); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := OpenReader(path, uid)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	for _, s := range []*Store{reader, reopened} {
		rows, err = s.Between(ctx, uid, time.Unix(0, 0), time.Unix(20, 0))
		if err != nil || len(rows) != 2 {
			t.Fatalf("after migration history=%+v err=%v", rows, err)
		}
		if rows[0].Ignition != nil || rows[0].Connected != nil || rows[0].ConnectionSource != "" || rows[1].Ignition == nil || *rows[1].Ignition || rows[1].Connected != nil || rows[1].IgnitionSource != "ign" || rows[1].ConnectionSource != "online" {
			t.Fatalf("after migration rows=%+v", rows)
		}
	}
}

func TestR1R3ReplayCompletesOnlyCompatibleMeasurements(t *testing.T) {
	const uid = "DEMO-CR13"
	ctx := context.Background()
	store, err := OpenWriter(filepath.Join(t.TempDir(), "replay.db"), uid)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	base := domain.GPSFix{UID: uid, GPSTime: time.Unix(10, 0), Fingerprint: [32]byte{1}}
	off, on := false, true
	save := func(ign, online *bool, ignitionSource, connectionSource string) {
		t.Helper()
		fix := base
		fix.Ignition, fix.Connected = ign, online
		fix.IgnitionSource, fix.ConnectionSource = ignitionSource, connectionSource
		if err := store.Save(ctx, fix); err != nil {
			t.Fatal(err)
		}
	}
	check := func(ign, online *bool, ignitionSource, connectionSource string) {
		t.Helper()
		rows, err := store.Between(ctx, uid, time.Unix(0, 0), time.Unix(20, 0))
		if err != nil || len(rows) != 1 {
			t.Fatalf("rows=%d err=%v", len(rows), err)
		}
		got := rows[0]
		if !reflect.DeepEqual(got.Ignition, ign) || !reflect.DeepEqual(got.Connected, online) || got.IgnitionSource != ignitionSource || got.ConnectionSource != connectionSource {
			t.Fatalf("stored signals %v/%v from %q/%q, want %v/%v from %q/%q", got.Ignition, got.Connected, got.IgnitionSource, got.ConnectionSource, ign, online, ignitionSource, connectionSource)
		}
	}
	save(nil, nil, "", "") // Position-only capture, same layout as a migrated row.
	save(nil, nil, "other", "online")
	check(nil, nil, "", "")
	save(&off, nil, "ign", "old_link") // A replay without a measured connection does not supply that field.
	check(&off, nil, "ign", "")
	save(&on, &on, "other", "ign") // Never reuse known ignition as connection.
	check(&off, nil, "ign", "")
	save(nil, nil, "", "")
	check(&off, nil, "ign", "")
	save(&on, &off, "ign", "online")
	check(&off, &off, "ign", "online")
	save(nil, &on, "ign", "online")
	check(&off, &off, "ign", "online")
}

func TestR1ReplayPreservesKnownConnectionWhileFillingIgnition(t *testing.T) {
	const uid = "DEMO-CR13"
	ctx := context.Background()
	s, err := OpenWriter(filepath.Join(t.TempDir(), "history.db"), uid)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	off, on := false, true
	original := domain.GPSFix{UID: uid, GPSTime: time.Unix(25, 0).UTC(), Fingerprint: [32]byte{2}, Connected: &off, ConnectionSource: "online", IgnitionSource: "old_ign"}
	if err := s.Save(ctx, original); err != nil {
		t.Fatal(err)
	}
	replay := original
	replay.Ignition = &on
	replay.IgnitionSource = "online"
	replay.ConnectionSource = "ign"
	if err := s.Save(ctx, replay); err != nil {
		t.Fatal(err)
	}
	fix, err := s.Latest(ctx, uid)
	if err != nil || !reflect.DeepEqual(fix, &original) {
		t.Fatalf("after swapped mapping: %+v err=%v", fix, err)
	}
	replay.IgnitionSource = "ign"
	replay.ConnectionSource = "online"
	if err := s.Save(ctx, replay); err != nil {
		t.Fatal(err)
	}
	fix, err = s.Latest(ctx, uid)
	if err != nil || !reflect.DeepEqual(fix, &replay) {
		t.Fatalf("after corrected ignition mapping: %+v err=%v", fix, err)
	}
}

func TestR1ReplayCanCorrectOnlyTheUnmeasuredSelector(t *testing.T) {
	const uid = "DEMO-CR13"
	ctx := context.Background()
	s, err := OpenWriter(filepath.Join(t.TempDir(), "history.db"), uid)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	off, on := false, true
	original := domain.GPSFix{UID: uid, GPSTime: time.Unix(25, 0), Fingerprint: [32]byte{2}, Ignition: &off, IgnitionSource: "ign", ConnectionSource: "old_link"}
	if err := s.Save(ctx, original); err != nil {
		t.Fatal(err)
	}
	replay := original
	replay.Connected = &on
	replay.ConnectionSource = "online"
	if err := s.Save(ctx, replay); err != nil {
		t.Fatal(err)
	}
	fix, err := s.Latest(ctx, uid)
	if err != nil || fix == nil || !reflect.DeepEqual(fix.Ignition, &off) || !reflect.DeepEqual(fix.Connected, &on) || fix.IgnitionSource != "ign" || fix.ConnectionSource != "online" {
		t.Fatalf("after corrected connection mapping: %+v err=%v", fix, err)
	}
}
