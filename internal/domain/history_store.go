package domain

import (
	"context"
	"time"
)

const ReceiverUnit = "CR-13"

type PositionArchive interface {
	Save(context.Context, GPSFix) error
	Latest(context.Context, string) (*GPSFix, error)
	Between(context.Context, string, time.Time, time.Time) ([]GPSFix, error)
}
