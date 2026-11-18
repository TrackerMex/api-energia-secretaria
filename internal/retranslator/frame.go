// Package retranslator decodes the binary Wialon Retranslator 1.0 protocol.
package retranslator

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"time"

	"wialon-sener-pilot/internal/domain"
)

// MaxFrameBytes is the local receiver's limit including the four-byte prefix.
const MaxFrameBytes = 1 << 20

var errInvalidFrame = errors.New("trama Wialon Retranslator inválida")

func readFrame(r io.Reader) ([]byte, error) {
	var prefix [4]byte
	if _, err := io.ReadFull(r, prefix[:]); err != nil {
		return nil, err
	}
	size := binary.LittleEndian.Uint32(prefix[:])
	// The smallest body has a one-character UID, terminator, time and mask.
	if size < 10 || size > MaxFrameBytes-4 {
		return nil, errInvalidFrame
	}
	frame := make([]byte, 4+int(size))
	copy(frame, prefix[:])
	if _, err := io.ReadFull(r, frame[4:]); err != nil {
		return nil, err
	}
	return frame, nil
}

func decodeFrame(frame []byte) (domain.GPSFix, bool, error) {
	var fix domain.GPSFix
	if len(frame) < 14 || len(frame) > MaxFrameBytes || int(binary.LittleEndian.Uint32(frame)) != len(frame)-4 {
		return fix, false, errInvalidFrame
	}
	uid, metadata, found := bytes.Cut(frame[4:], []byte{0})
	if !found || len(uid) == 0 || len(metadata) < 8 {
		return fix, false, errInvalidFrame
	}
	for _, char := range uid {
		if char < 0x20 || char > 0x7e {
			return fix, false, errInvalidFrame
		}
	}
	fix.UID = string(uid)
	fix.GPSTime = time.Unix(int64(binary.BigEndian.Uint32(metadata)), 0).UTC()
	blocks := metadata[8:]
	var hasPosition bool
	for len(blocks) > 0 {
		if len(blocks) < 6 {
			return fix, false, errInvalidFrame
		}
		kind := binary.BigEndian.Uint16(blocks)
		size := binary.BigEndian.Uint32(blocks[2:])
		blocks = blocks[6:]
		if uint64(size) > uint64(len(blocks)) {
			return fix, false, errInvalidFrame
		}
		payload := blocks[:size]
		blocks = blocks[size:]
		if kind != 0x0bbb {
			continue
		}
		if len(payload) < 4 {
			return fix, false, errInvalidFrame
		}
		// First two bytes are the stealth attribute and the type of value.
		valueType := payload[1]
		name, value, terminated := bytes.Cut(payload[2:], []byte{0})
		if !terminated || len(name) == 0 {
			return fix, false, errInvalidFrame
		}
		if string(name) != "posinfo" {
			if !validSensorValue(valueType, value) {
				return fix, false, errInvalidFrame
			}
			continue
		}
		if hasPosition || valueType != 2 || len(value) != 29 {
			return fix, false, errInvalidFrame
		}
		fix.Longitude = math.Float64frombits(binary.LittleEndian.Uint64(value))
		fix.Latitude = math.Float64frombits(binary.LittleEndian.Uint64(value[8:]))
		altitude := math.Float64frombits(binary.LittleEndian.Uint64(value[16:]))
		fix.Speed = binary.BigEndian.Uint16(value[24:])
		fix.Course = binary.BigEndian.Uint16(value[26:])
		if math.IsNaN(fix.Longitude) || math.Abs(fix.Longitude) > 180 ||
			math.IsNaN(fix.Latitude) || math.Abs(fix.Latitude) > 90 ||
			math.IsNaN(altitude) || math.IsInf(altitude, 0) || fix.Course > 359 {
			return fix, false, errInvalidFrame
		}
		hasPosition = true
	}
	if hasPosition {
		fix.Fingerprint = sha256.Sum256(frame)
	}
	return fix, hasPosition, nil
}

func validSensorValue(kind byte, value []byte) bool {
	switch kind {
	case 1, 2: // text and binary extension values have variable length
		return true
	case 3:
		return len(value) == 4
	case 4, 5:
		return len(value) == 8
	default:
		return false
	}
}
