// Package ingest reads Wialon GPS messages and maps positions to SENER's schema.
package ingest

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Source struct {
	BaseURL string
	Token   string
	Client  *http.Client
}

// Position retains the original message so validation failures can be reconciled.
type Position struct {
	Key     string
	UnitID  int64
	Message json.RawMessage
}

type session struct {
	endpoint string
	client   http.Client
	id       string
}

func (s Source) start(ctx context.Context) (*session, error) {
	origin, err := url.Parse(s.BaseURL)
	if err != nil || origin.Hostname() == "" || origin.User != nil || origin.RawQuery != "" || origin.Fragment != "" || strings.Trim(origin.Path, "/") != "" {
		return nil, errors.New("WIALON_BASE_URL debe ser el origen HTTPS del centro de datos")
	}
	ip := net.ParseIP(origin.Hostname())
	if origin.Scheme != "https" && !(origin.Scheme == "http" && ip.IsLoopback()) {
		return nil, errors.New("Wialon requiere HTTPS; HTTP solo se admite en loopback para pruebas")
	}
	if strings.TrimSpace(s.Token) == "" {
		return nil, errors.New("falta WIALON_TOKEN")
	}
	origin.Path = "/wialon/ajax.html"
	client := http.Client{Timeout: 30 * time.Second}
	if s.Client != nil {
		client = *s.Client
	}
	if client.Timeout == 0 {
		client.Timeout = 30 * time.Second
	}
	// A redirect must never forward form credentials to another host.
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	sess := &session{endpoint: origin.String(), client: client}
	var login struct {
		Session string `json:"eid"`
	}
	if err := sess.call(ctx, "token/login", map[string]any{"token": s.Token, "fl": 1}, &login); err != nil {
		return nil, err
	}
	if login.Session == "" {
		return nil, errors.New("token/login no devolvió una sesión")
	}
	sess.id = login.Session
	return sess, nil
}

func (s *session) call(ctx context.Context, svc string, params any, out any) error {
	payload, err := json.Marshal(params)
	if err != nil {
		return err
	}
	form := url.Values{"svc": {svc}, "params": {string(payload)}}
	if s.id != "" {
		form.Set("sid", s.id)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("%s: petición inválida", svc)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := s.client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("%s: no se pudo completar la conexión Wialon", svc)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: Wialon HTTP %d", svc, resp.StatusCode)
	}
	const maxReply = 32 << 20
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxReply+1))
	if err != nil || len(body) > maxReply {
		return fmt.Errorf("%s: página Wialon incompleta o demasiado grande", svc)
	}
	body = bytes.TrimSpace(body)
	if len(body) == 0 || bytes.Equal(body, []byte("null")) {
		return fmt.Errorf("%s: respuesta Wialon vacía", svc)
	}
	if body[0] == '{' {
		var status struct {
			Error int `json:"error"`
		}
		if json.Unmarshal(body, &status) != nil {
			return fmt.Errorf("%s: objeto de respuesta inválido", svc)
		}
		if status.Error != 0 {
			return fmt.Errorf("%s: Wialon error %d", svc, status.Error)
		}
	}
	if out != nil && json.Unmarshal(body, out) != nil {
		return fmt.Errorf("%s: formato de respuesta Wialon inválido", svc)
	}
	return nil
}

// Positions loads one unit's inclusive, whole-second Wialon interval.
func (s Source) Positions(ctx context.Context, unitID int64, from, to time.Time) ([]Position, error) {
	if unitID <= 0 || from.Unix() < 0 || to.Before(from) || to.Unix() > math.MaxUint32 {
		return nil, errors.New("unidad o intervalo UTC inválido")
	}
	sess, err := s.start(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		// Drop this dedicated loader session even if the caller canceled its read.
		closeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 3*time.Second)
		defer cancel()
		_ = sess.call(closeCtx, "core/logout", struct{}{}, nil)
	}()

	const pageSize = 250
	var first struct {
		Count    *int              `json:"count"`
		Messages []json.RawMessage `json:"messages"`
	}
	// Match data messages (high byte 0) with position bit 0x01 set.
	params := map[string]any{"itemId": unitID, "timeFrom": from.Unix(), "timeTo": to.Unix(), "flags": 1, "flagsMask": 0xff01, "loadCount": pageSize}
	if err := sess.call(ctx, "messages/load_interval", params, &first); err != nil {
		return nil, err
	}
	if first.Count == nil || *first.Count < 0 || first.Messages == nil || len(first.Messages) > *first.Count {
		return nil, errors.New("messages/load_interval: conteo o página inválidos")
	}
	var positions []Position
	page := first.Messages
	for {
		for _, raw := range page {
			pos, err := newPosition(unitID, raw)
			if err != nil {
				return nil, err
			}
			positions = append(positions, pos)
		}
		start := len(positions)
		if start == *first.Count {
			return positions, nil
		}
		end := start + min(pageSize, *first.Count-start)
		page = nil
		if err := sess.call(ctx, "messages/get_messages", map[string]int{"indexFrom": start, "indexTo": end}, &page); err != nil {
			return nil, err
		}
		if len(page) == 0 || len(page) > end-start {
			return nil, errors.New("messages/get_messages: paginación incompleta")
		}
	}
}

func newPosition(unitID int64, raw json.RawMessage) (Position, error) {
	// Sort object keys recursively without rounding device identifiers or counters.
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var data map[string]any
	if err := decoder.Decode(&data); err != nil || data == nil {
		return Position{}, errors.New("mensaje Wialon inválido")
	}
	// Loader metadata is not a part of the device reading.
	delete(data, "sensors")
	canonical, err := json.Marshal(data)
	if err != nil {
		return Position{}, errors.New("contenido de mensaje Wialon inválido")
	}
	digest := sha256.Sum256(canonical)
	key := strconv.FormatInt(unitID, 10) + ":" + fmt.Sprintf("%x", digest)
	return Position{Key: key, UnitID: unitID, Message: raw}, nil
}
