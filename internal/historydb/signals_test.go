package historydb

import (
	"context"
	"database/sql"
	"path/filepath"
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
