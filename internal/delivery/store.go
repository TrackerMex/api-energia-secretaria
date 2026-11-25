package delivery

import (
	"context"
	"crypto/rand"
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
	db       *sql.DB
	localTCP bool
}

func Open(path string, unitID int64, rfc string) (*Store, error) {
	return open(path, unitID, rfc, "")
}

// OpenTCP binds a simulator queue to a retranslator UID. A legacy unmarked
// outbox belongs to the Remote API flow and cannot become a simulator queue.
func OpenTCP(path, rfc, uid string) (*Store, error) {
	if uid == "" {
		return nil, errors.New("missing retranslator UID")
	}
	return open(path, 0, rfc, uid)
}

func open(path string, unitID int64, rfc, uid string) (*Store, error) {
	if (uid == "" && unitID <= 0) || rfc == "" {
		return nil, errors.New("missing database identity")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	if err = file.Close(); err != nil {
		return nil, err
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	dsn := url.URL{Scheme: "file", Path: "/" + strings.TrimPrefix(filepath.ToSlash(absolute), "/"), RawQuery: "_pragma=busy_timeout(10000)&_txlock=immediate&mode=rw"}
	db, err := sql.Open("sqlite", dsn.String())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	ok := false
	defer func() {
		if !ok {
			db.Close()
		}
	}()
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	// Bind and verify identity atomically, including the first concurrent open.
	var tables, scopeTables, otherTables int
	err = tx.QueryRow(`SELECT COUNT(*),COALESCE(SUM(name='outbox_scope'),0),
 COALESCE(SUM(name NOT IN ('pilot','reports','outbox_scope')),0) FROM sqlite_schema WHERE type='table' AND name NOT GLOB 'sqlite_*'`).Scan(&tables, &scopeTables, &otherTables)
	if err != nil {
		return nil, err
	}
	if otherTables != 0 || (uid != "" && tables != 0 && scopeTables == 0) {
		return nil, errors.New("el archivo no es una cola de esta fuente")
	}
	_, err = tx.Exec(`CREATE TABLE IF NOT EXISTS outbox_scope(instance INTEGER PRIMARY KEY CHECK(instance=1),tcp_uid TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS pilot (
 instance INTEGER PRIMARY KEY CHECK(instance=1), unit_id INTEGER NOT NULL, rfc TEXT NOT NULL,
 cursor INTEGER, cooldown INTEGER NOT NULL DEFAULT 0
);
CREATE TABLE IF NOT EXISTS reports (
 source_key TEXT UNIQUE NOT NULL, id TEXT PRIMARY KEY, body BLOB, status TEXT NOT NULL,
 window_from INTEGER NOT NULL, window_to INTEGER NOT NULL, detail TEXT NOT NULL, retry_at INTEGER NOT NULL DEFAULT 0, updated INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS by_status ON reports(status,retry_at);`)
	if err != nil {
		return nil, err
	}
	if _, err = tx.Exec("INSERT OR IGNORE INTO outbox_scope(instance,tcp_uid) VALUES(1,?)", uid); err != nil {
		return nil, err
	}
	var ownerUID string
	if err = tx.QueryRow("SELECT tcp_uid FROM outbox_scope WHERE instance=1").Scan(&ownerUID); err != nil {
		return nil, err
	}
	if ownerUID != uid {
		return nil, errors.New("la cola pertenece a otra fuente o UID")
	}
	if _, err = tx.Exec("INSERT OR IGNORE INTO pilot(instance,unit_id,rfc) VALUES(1,?,?)", unitID, rfc); err != nil {
		return nil, err
	}
	var unit int64
	var owner string
	if err = tx.QueryRow("SELECT unit_id,rfc FROM pilot WHERE instance=1").Scan(&unit, &owner); err != nil {
		return nil, err
	}
	if unit != unitID || owner != rfc {
		return nil, errors.New("database belongs to a different unit or RFC")
	}
	if uid != "" {
		if _, err = tx.Exec("CREATE TEMP TABLE tcp_window(source_key TEXT PRIMARY KEY)"); err != nil {
			return nil, err
		}
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	if err = os.Chmod(path, 0600); err != nil {
		return nil, err
	}
	ok = true
	return &Store{db: db, localTCP: uid != ""}, nil
}
func (s *Store) Close() error { return s.db.Close() }

func (s *Store) Cursor(ctx context.Context) (time.Time, error) {
	var cursor sql.NullInt64
	err := s.db.QueryRowContext(ctx, "SELECT cursor FROM pilot WHERE instance=1").Scan(&cursor)
	if err != nil || !cursor.Valid {
		return time.Time{}, err
	}
	return time.Unix(cursor.Int64, 0).UTC(), nil
}

func makeUUID() string {
	var b [16]byte
	rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// SaveWindow commits the entire batch and cursor together. Reading an older
// window again can repair a previously invalid mapping without resending it.
func (s *Store) SaveWindow(ctx context.Context, from, to time.Time, inputs []domain.Input) error {
	if from.Unix() < 0 || to.Before(from) {
		return errors.New("invalid message window")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if s.localTCP {
		// The connection-local set prevents a concurrent command using another
		// interval from changing which already-pending reports this one may send.
		if _, err = tx.ExecContext(ctx, "DELETE FROM tcp_window"); err != nil {
			return err
		}
	} else {
		var cursor sql.NullInt64
		if err = tx.QueryRowContext(ctx, "SELECT cursor FROM pilot WHERE instance=1").Scan(&cursor); err != nil {
			return err
		}
		if cursor.Valid && from.Unix() > cursor.Int64+1 {
			return errors.New("window leaves a gap after the stored cursor")
		}
	}
	for _, input := range inputs {
		if input.Key == "" || input.Build == nil {
			return errors.New("missing source identity or mapper")
		}
		if s.localTCP {
			if _, err = tx.ExecContext(ctx, "INSERT OR IGNORE INTO tcp_window(source_key) VALUES(?)", input.Key); err != nil {
				return err
			}
		}
		var id, state string
		err = tx.QueryRowContext(ctx, "SELECT id,status FROM reports WHERE source_key=?", input.Key).Scan(&id, &state)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			id = makeUUID()
		case err != nil:
			return err
		case state != domain.Invalid:
			continue
		}
		body, validationErr := input.Build(id)
		state, detail := domain.Pending, ""
		if validationErr != nil {
			// Validation messages from the local mapper name the missing fields.
			state, detail, body = domain.Invalid, validationErr.Error(), nil
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO reports(source_key,id,body,status,detail,updated,window_from,window_to) VALUES(?,?,?,?,?,?,?,?)
ON CONFLICT(source_key) DO UPDATE SET body=excluded.body,status=excluded.status,detail=excluded.detail,updated=excluded.updated WHERE reports.status='invalid'`, input.Key, id, body, state, detail, time.Now().Unix(), from.Unix(), to.Unix())
		if err != nil {
			return err
		}
	}
	if !s.localTCP {
		if _, err = tx.ExecContext(ctx, "UPDATE pilot SET cursor=MAX(COALESCE(cursor,?),?) WHERE instance=1", to.Unix(), to.Unix()); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// Mark the intention durably before the network call. A crash here leaves the
// outcome uncertain; automatic replay could deliver a duplicate.
func (s *Store) Claim(ctx context.Context, now time.Time) (*domain.Envelope, error) {
	var row domain.Envelope
	filter := ""
	if s.localTCP {
		filter = " AND source_key IN (SELECT source_key FROM tcp_window)"
	}
	err := s.db.QueryRowContext(ctx, `UPDATE reports SET status='uncertain',detail='delivery_started',updated=?
WHERE rowid=(SELECT rowid FROM reports WHERE status='pending' AND retry_at<=?
AND (SELECT cooldown FROM pilot WHERE instance=1)<=?`+filter+` ORDER BY rowid LIMIT 1)
RETURNING id,body`, now.Unix(), now.Unix(), now.Unix()).Scan(&row.ID, &row.Body)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *Store) Finish(ctx context.Context, id string, out domain.Outcome) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	retry := int64(0)
	if out.State == domain.Pending {
		retry = out.RetryAt.Unix()
		if _, err = tx.ExecContext(ctx, "UPDATE pilot SET cooldown=MAX(cooldown,?) WHERE instance=1", retry); err != nil {
			return err
		}
	}
	res, err := tx.ExecContext(ctx, "UPDATE reports SET status=?,detail=?,retry_at=?,updated=? WHERE id=? AND status='uncertain'", out.State, out.Detail, retry, time.Now().Unix(), id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return errors.New("delivery state has changed")
	}
	return tx.Commit()
}

func (s *Store) Counts(ctx context.Context) (map[string]int, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT status,COUNT(*) FROM reports GROUP BY status")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	counts := make(map[string]int)
	for rows.Next() {
		var state string
		var n int
		if err = rows.Scan(&state, &n); err != nil {
			return nil, err
		}
		counts[state] = n
	}
	return counts, rows.Err()
}
