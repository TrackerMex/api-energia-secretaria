// Package historydb persists the fixes of a single configured retranslator UID.
package historydb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"wialon-sener-pilot/internal/domain"
)

type Store struct {
	db  *sql.DB
	uid string
}

func OpenWriter(path, uid string) (*Store, error) {
	if strings.TrimSpace(path) == "" || strings.TrimSpace(uid) == "" {
		return nil, errors.New("faltan la ruta de SQLite o el UID")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
	if err == nil {
		if err = file.Close(); err != nil {
			return nil, err
		}
	} else if !errors.Is(err, os.ErrExist) {
		return nil, err
	}
	return open(path, uid, true)
}

func OpenReader(path, uid string) (*Store, error) {
	if strings.TrimSpace(path) == "" || strings.TrimSpace(uid) == "" {
		return nil, errors.New("faltan la ruta de SQLite o el UID")
	}
	return open(path, uid, false)
}

func open(path, uid string, write bool) (*Store, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	location := url.URL{Scheme: "file", Path: "/" + strings.TrimPrefix(filepath.ToSlash(absolute), "/")}
	params := location.Query()
	if write {
		params.Set("mode", "rw")
		params.Set("_txlock", "immediate")
	} else {
		params.Set("mode", "ro")
	}
	params.Set("_pragma", "busy_timeout(5000)")
	location.RawQuery = params.Encode()
	db, err := sql.Open("sqlite", location.String())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	store := &Store{db: db, uid: uid}
	if write {
		err = store.initialize()
	} else {
		err = store.verifyIdentity(db.QueryRow("SELECT uid, alias FROM receiver_unit WHERE id=1"))
	}
	if err != nil {
		db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) initialize() error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var foreignTables int
	err = tx.QueryRow("SELECT count(*) FROM sqlite_schema WHERE type='table' AND name NOT LIKE 'sqlite_%' AND name NOT IN ('receiver_unit','gps_fixes')").Scan(&foreignTables)
	if err != nil {
		return err
	}
	if foreignTables != 0 {
		return errors.New("el archivo contiene tablas ajenas al historial del receptor")
	}
	_, err = tx.Exec(`CREATE TABLE IF NOT EXISTS receiver_unit (
  id INTEGER PRIMARY KEY CHECK(id=1), uid TEXT NOT NULL UNIQUE, alias TEXT NOT NULL
 );
 CREATE TABLE IF NOT EXISTS gps_fixes (
  uid TEXT NOT NULL, fingerprint BLOB NOT NULL, gps_sec INTEGER NOT NULL,
  latitude REAL NOT NULL, longitude REAL NOT NULL, speed INTEGER NOT NULL, course INTEGER NOT NULL,
  PRIMARY KEY(uid, fingerprint)
 );
 CREATE INDEX IF NOT EXISTS gps_time ON gps_fixes(uid,gps_sec,fingerprint);`)
	if err != nil {
		return err
	}
	if _, err = tx.Exec("INSERT INTO receiver_unit(id,uid,alias) VALUES(1,?,?) ON CONFLICT(id) DO NOTHING", s.uid, domain.ReceiverUnit); err != nil {
		return err
	}
	if err = s.verifyIdentity(tx.QueryRow("SELECT uid, alias FROM receiver_unit WHERE id=1")); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) verifyIdentity(row *sql.Row) error {
	var uid, alias string
	if err := row.Scan(&uid, &alias); err != nil {
		return fmt.Errorf("leer identidad del historial: %w", err)
	}
	if uid != s.uid || alias != domain.ReceiverUnit {
		return errors.New("el UID o alias configurado difiere de la identidad de este historial")
	}
	return nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) Save(ctx context.Context, fix domain.GPSFix) error {
	if fix.UID != s.uid {
		return errors.New("el UID de la posición difiere del configurado")
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO gps_fixes(uid,fingerprint,gps_sec,latitude,longitude,speed,course)
 VALUES(?,?,?,?,?,?,?) ON CONFLICT(uid,fingerprint) DO NOTHING`, fix.UID, fix.Fingerprint[:], fix.GPSTime.Unix(), fix.Latitude, fix.Longitude, fix.Speed, fix.Course)
	return err
}

const columns = "uid,fingerprint,gps_sec,latitude,longitude,speed,course"

func scanFix(row interface{ Scan(...any) error }) (domain.GPSFix, error) {
	var fix domain.GPSFix
	var hash []byte
	var seconds int64
	err := row.Scan(&fix.UID, &hash, &seconds, &fix.Latitude, &fix.Longitude, &fix.Speed, &fix.Course)
	if err != nil {
		return fix, err
	}
	if len(hash) != len(fix.Fingerprint) {
		return fix, errors.New("huella inválida en el historial")
	}
	copy(fix.Fingerprint[:], hash)
	fix.GPSTime = time.Unix(seconds, 0).UTC()
	return fix, nil
}

func (s *Store) Latest(ctx context.Context, uid string) (*domain.GPSFix, error) {
	if uid != s.uid {
		return nil, errors.New("el UID consultado difiere del configurado")
	}
	row := s.db.QueryRowContext(ctx, "SELECT "+columns+" FROM gps_fixes WHERE uid=? ORDER BY gps_sec DESC,fingerprint DESC LIMIT 1", uid)
	fix, err := scanFix(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &fix, nil
}

// Wialon timestamps are whole seconds, so both fractional endpoints round up
// for a half-open comparison against the stored integer.
func ceilSecond(t time.Time) int64 {
	if t.Nanosecond() != 0 {
		return t.Unix() + 1
	}
	return t.Unix()
}

func (s *Store) Between(ctx context.Context, uid string, from, to time.Time) ([]domain.GPSFix, error) {
	if uid != s.uid {
		return nil, errors.New("el UID consultado difiere del configurado")
	}
	rows, err := s.db.QueryContext(ctx, "SELECT "+columns+" FROM gps_fixes WHERE uid=? AND gps_sec>=? AND gps_sec<? ORDER BY gps_sec,fingerprint", uid, ceilSecond(from), ceilSecond(to))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	fixes := make([]domain.GPSFix, 0)
	for rows.Next() {
		fix, err := scanFix(rows)
		if err != nil {
			return nil, err
		}
		fixes = append(fixes, fix)
	}
	return fixes, rows.Err()
}
