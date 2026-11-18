// Package domain defines the position report and the ports used to capture and deliver it.
package domain

import (
	"context"
	"encoding/json"
	"time"
)

// Position identifies one unit reading and carries its source values for mapping.
type Position struct {
	Key     string
	UnitID  int64
	Message json.RawMessage
}

type PositionSource interface {
	Positions(context.Context, int64, time.Time, time.Time) ([]Position, error)
}

// Input builds the report after the outbox has assigned its persistent UUID.
type Input struct {
	Key   string
	Build func(id string) ([]byte, error)
}

type Envelope struct {
	ID   string
	Body []byte
}

const (
	Pending   = "pending"
	Accepted  = "accepted"
	Rejected  = "rejected"
	Uncertain = "uncertain"
	Invalid   = "invalid"
)

type Outcome struct {
	State   string
	Detail  string
	RetryAt time.Time
}

type ReportSender interface {
	Send(context.Context, string, []byte) Outcome
}

type Outbox interface {
	SaveWindow(context.Context, time.Time, time.Time, []Input) error
	Claim(context.Context, time.Time) (*Envelope, error)
	Finish(context.Context, string, Outcome) error
}
