package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"io"
	"net"
	"testing"
	"time"

	"wialon-sener-pilot/internal/application"
	"wialon-sener-pilot/internal/domain"
	"wialon-sener-pilot/internal/historydb"
	"wialon-sener-pilot/internal/retranslator"
	"wialon-sener-pilot/internal/unitprofile"
)

func TestR1R2ReplayRepairsPreviouslyInvalidReport(t *testing.T) {
	f := newFixture(t)
	wire, err := hex.DecodeString(wireFixture)
	if err != nil {
		t.Fatal(err)
	}
	hash := sha256.Sum256(wire)
	// The previous receptor stored the fingerprint and location, even when the
	// original frame contained sensor blocks it did not yet decode.
	old, err := sql.Open("sqlite", f.history)
	if err != nil {
		t.Fatal(err)
	}
	_, err = old.Exec(`CREATE TABLE receiver_unit(id INTEGER PRIMARY KEY,uid TEXT,alias TEXT);
 INSERT INTO receiver_unit VALUES(1,'DEMO-CR13','CR-13');
 CREATE TABLE gps_fixes(uid TEXT,fingerprint BLOB,gps_sec INTEGER,latitude REAL,longitude REAL,speed INTEGER,course INTEGER, PRIMARY KEY(uid,fingerprint));
 INSERT INTO gps_fixes VALUES(?,?,?,20,-100,42,90)`, uid, hash[:], start.Unix())
	old.Close()
	if err != nil {
		t.Fatal(err)
	}
	sim, posts := simulator(t, nil)
	defer sim.Close()
	end := start.Add(time.Second)
	f.process(t, sim, start, end)
	before := f.records(t)
	if len(before) != 1 || before[0].id == "" || before[0].status != domain.Invalid || len(posts()) != 0 {
		t.Fatalf("before replay: reports=%v posts=%v", before, posts())
	}

	profile, err := unitprofile.LoadTCP(profilePath, uid)
	if err != nil {
		t.Fatal(err)
	}
	store, err := historydb.OpenWriter(f.history, uid)
	if err != nil {
		t.Fatal(err)
	}
	receiver := application.Receiver{UID: uid, Archive: store}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		store.Close()
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- retranslator.ServeWithProfile(ctx, listener, receiver.Accept, profile) }()
	func() {
		defer func() {
			cancel()
			if err := <-done; err != nil {
				t.Error(err)
			}
			store.Close()
		}()
		conn, err := net.DialTimeout("tcp", listener.Addr().String(), 3*time.Second)
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()
		conn.SetDeadline(time.Now().Add(3 * time.Second))
		if _, err = conn.Write(wire); err != nil {
			t.Fatal(err)
		}
		if err = conn.(*net.TCPConn).CloseWrite(); err != nil {
			t.Fatal(err)
		}
		if _, err = io.Copy(io.Discard, conn); err != nil {
			t.Fatal(err)
		}
	}()

	f.process(t, sim, start, end)
	f.process(t, sim, start, end)
	after := f.records(t)
	received := posts()
	if len(after) != 1 || after[0].id != before[0].id || after[0].status != domain.Accepted || len(received) != 1 || received[0].id != before[0].id {
		t.Fatalf("after replay: reports=%v posts=%v", after, received)
	}
	history, err := historydb.OpenReader(f.history, uid)
	if err != nil {
		t.Fatal(err)
	}
	defer history.Close()
	positions, err := history.Between(context.Background(), uid, start, end)
	if err != nil || len(positions) != 1 || positions[0].Fingerprint != hash || positions[0].Ignition == nil || *positions[0].Ignition || positions[0].Connected == nil || !*positions[0].Connected {
		t.Fatalf("history after restart: %v, err=%v", positions, err)
	}
}
