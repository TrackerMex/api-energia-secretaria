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
	mapped, err := hasSignalColumns(context.Background(), tx)
	if err != nil {
		return err
	}
	if !mapped {
		_, err = tx.Exec(`ALTER TABLE gps_fixes ADD COLUMN ignition INTEGER CHECK(ignition IN (0,1));
ALTER TABLE gps_fixes ADD COLUMN connected INTEGER CHECK(connected IN (0,1));
ALTER TABLE gps_fixes ADD COLUMN ignition_source TEXT NOT NULL DEFAULT '';
ALTER TABLE gps_fixes ADD COLUMN connection_source TEXT NOT NULL DEFAULT '';`)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

// Inspect the schema in the same snapshot as the read or migration.
func hasSignalColumns(ctx context.Context, tx *sql.Tx) (bool, error) {
	var n int
	err := tx.QueryRowContext(ctx, `SELECT count(*) FROM pragma_table_info('gps_fixes')
WHERE name IN ('ignition','connected','ignition_source','connection_source')`).Scan(&n)
	if err != nil {
		return false, err
	}
	if n != 0 && n != 4 {
		return false, errors.New("esquema incompleto de señales del historial")
	}
	return n == 4, nil
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
	// Replaying identical bytes can reveal a previously unmapped signal. Keep
	// each known measurement and its original selector together.
	_, err := s.db.ExecContext(ctx, `INSERT INTO gps_fixes AS stored
 (uid,fingerprint,gps_sec,latitude,longitude,speed,course,ignition,connected,ignition_source,connection_source)
 VALUES(?,?,?,?,?,?,?,?,?,?,?)
 ON CONFLICT(uid,fingerprint) DO UPDATE SET
 (ignition,ignition_source,connected,connection_source) = (
   SELECT
     CASE WHEN add_ignition THEN excluded.ignition ELSE stored.ignition END,
     CASE WHEN add_ignition THEN excluded.ignition_source ELSE stored.ignition_source END,
     CASE WHEN add_connection THEN excluded.connected ELSE stored.connected END,
     CASE WHEN add_connection THEN excluded.connection_source ELSE stored.connection_source END
   FROM (SELECT
     stored.ignition IS NULL AND excluded.ignition IS NOT NULL AND excluded.ignition_source <> ''
       AND (stored.connected IS NULL OR excluded.ignition_source <> stored.connection_source) AS add_ignition,
     stored.connected IS NULL AND excluded.connected IS NOT NULL AND excluded.connection_source <> ''
       AND (stored.ignition IS NULL OR excluded.connection_source <> stored.ignition_source) AS add_connection)
 )
 WHERE (stored.ignition IS NULL AND excluded.ignition IS NOT NULL)
    OR (stored.connected IS NULL AND excluded.connected IS NOT NULL)`, fix.UID, fix.Fingerprint[:], fix.GPSTime.Unix(), fix.Latitude, fix.Longitude, fix.Speed, fix.Course, fix.Ignition, fix.Connected, fix.IgnitionSource, fix.ConnectionSource)
	return err
}

const columns = "uid,fingerprint,gps_sec,latitude,longitude,speed,course"

func scanFix(row interface{ Scan(...any) error }) (domain.GPSFix, error) {
	var fix domain.GPSFix
	var hash []byte
	var seconds int64
	err := row.Scan(&fix.UID, &hash, &seconds, &fix.Latitude, &fix.Longitude, &fix.Speed, &fix.Course, &fix.Ignition, &fix.Connected, &fix.IgnitionSource, &fix.ConnectionSource)
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
	fixes, err := s.readFixes(ctx, "WHERE uid=? ORDER BY gps_sec DESC,fingerprint DESC LIMIT 1", uid)
	if err != nil || len(fixes) == 0 {
		return nil, err
	}
	return &fixes[0], nil
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
	return s.readFixes(ctx, "WHERE uid=? AND gps_sec>=? AND gps_sec<? ORDER BY gps_sec,fingerprint", uid, ceilSecond(from), ceilSecond(to))
}

func (s *Store) readFixes(ctx context.Context, clause string, args ...any) ([]domain.GPSFix, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	mapped, err := hasSignalColumns(ctx, tx)
	if err != nil {
		return nil, err
	}
	signals := "NULL,NULL,'',''"
	if mapped {
		signals = "ignition,connected,ignition_source,connection_source"
	}
	rows, err := tx.QueryContext(ctx, "SELECT "+columns+","+signals+" FROM gps_fixes "+clause, args...)
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
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return fixes, tx.Commit()
}
