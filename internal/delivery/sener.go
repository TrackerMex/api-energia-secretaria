// Package delivery persists and transmits single-unit SENER development reports.
package delivery

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"regexp"
	"strconv"
	"time"

	"wialon-sener-pilot/internal/domain"
)

const DevelopmentURL = "https://dev-api.energia.gob.mx/api/v1/telemetria/reporte/"

type Sender struct {
	URL    string
	Token  string
	Client *http.Client
}

func (s Sender) Send(ctx context.Context, rfc string, body []byte) domain.Outcome {
	var buf bytes.Buffer
	form := multipart.NewWriter(&buf)
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", mime.FormatMediaType("form-data", map[string]string{"name": "gps", "filename": rfc + ".json"}))
	header.Set("Content-Type", "application/json")
	part, err := form.CreatePart(header)
	if err != nil {
		return domain.Outcome{State: domain.Rejected, Detail: "multipart_header"}
	}
	if _, err = part.Write(body); err != nil {
		return domain.Outcome{State: domain.Rejected, Detail: "multipart_body"}
	}
	if err = form.Close(); err != nil {
		return domain.Outcome{State: domain.Rejected, Detail: "multipart_close"}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.URL, &buf)
	if err != nil {
		return domain.Outcome{State: domain.Rejected, Detail: "endpoint"}
	}
	req.Header.Set("Content-Type", form.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.Token)
	client := http.Client{}
	if s.Client != nil {
		client = *s.Client
	}
	if client.Timeout == 0 {
		client.Timeout = 30 * time.Second
	}
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	resp, err := client.Do(req)
	if err != nil {
		return domain.Outcome{State: domain.Uncertain, Detail: "transport"}
	}
	defer resp.Body.Close()
	switch {
	case resp.StatusCode == http.StatusTooManyRequests:
		return domain.Outcome{State: domain.Pending, Detail: "HTTP 429", RetryAt: retryTime(time.Now(), resp.Header.Get("Retry-After"))}
	case resp.StatusCode == 401 || resp.StatusCode == 403 || resp.StatusCode == 404 || resp.StatusCode == 405:
		return domain.Outcome{State: domain.Pending, Detail: fmt.Sprintf("HTTP %d", resp.StatusCode), RetryAt: time.Now().Add(5 * time.Minute)}
	case resp.StatusCode >= 400 && resp.StatusCode < 500 && resp.StatusCode != http.StatusRequestTimeout:
		detail := fmt.Sprintf("HTTP %d", resp.StatusCode)
		var result apiResult
		if result.read(resp.Body) == nil {
			detail += result.code(s.Token)
		}
		return domain.Outcome{State: domain.Rejected, Detail: detail}
	case resp.StatusCode != http.StatusOK:
		return domain.Outcome{State: domain.Uncertain, Detail: fmt.Sprintf("HTTP %d", resp.StatusCode)}
	}
	var result apiResult
	if result.read(resp.Body) != nil || result.Error == nil {
		return domain.Outcome{State: domain.Uncertain, Detail: "invalid_response"}
	}
	switch *result.Error {
	case 0:
		return domain.Outcome{State: domain.Accepted, Detail: "HTTP 200 error=0"}
	case 1:
		return domain.Outcome{State: domain.Rejected, Detail: "HTTP 200 error=1" + result.code(s.Token)}
	default:
		return domain.Outcome{State: domain.Uncertain, Detail: "unknown_error_flag"}
	}
}

type apiResult struct {
	Error *int   `json:"error"`
	Code  string `json:"code_error"`
}

var safeCode = regexp.MustCompile("^[A-Za-z0-9_-]{1,32}$")

func (r apiResult) code(token string) string {
	if r.Code != token && safeCode.MatchString(r.Code) {
		return " code=" + r.Code
	}
	return ""
}
func (r *apiResult) read(body io.Reader) error {
	const limit = 1 << 20
	raw, err := io.ReadAll(io.LimitReader(body, limit+1))
	if err != nil {
		return err
	}
	if len(raw) > limit {
		return fmt.Errorf("SENER response exceeds limit")
	}
	return json.Unmarshal(raw, r)
}

func retryTime(now time.Time, value string) time.Time {
	if seconds, err := strconv.ParseInt(value, 10, 32); err == nil && seconds > 0 {
		return now.Add(time.Duration(seconds) * time.Second)
	}
	if date, err := http.ParseTime(value); err == nil && date.After(now) {
		return date
	}
	return now.Add(time.Minute)
}
