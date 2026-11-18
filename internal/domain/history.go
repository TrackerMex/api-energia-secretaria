package domain

import "time"

// GPSFix is the historical position of one controller at the recorded GPS time.
type GPSFix struct {
	UID                 string
	GPSTime             time.Time
	Latitude, Longitude float64
	Speed, Course       uint16
	Fingerprint         [32]byte
}
