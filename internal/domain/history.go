package domain

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

// GPSFix is the historical position of one controller at the recorded GPS time.
type GPSFix struct {
	UID                 string
	GPSTime             time.Time
	Latitude, Longitude float64
	Speed, Course       uint16
	Fingerprint         [32]byte
	// Nil is unknown. Source names preserve the capture-time meaning of each value.
	Ignition, Connected              *bool
	IgnitionSource, ConnectionSource string
}

// DigitalInputMask returns the bit for a canonical Wialon digital input selector.
func DigitalInputMask(selector string) (uint32, bool) {
	index, found := strings.CutPrefix(selector, "digital:in")
	if !found {
		return 0, false
	}
	n, err := strconv.Atoi(index)
	if err != nil || n < 1 || n > 32 || strconv.Itoa(n) != index {
		return 0, false
	}
	return uint32(1) << (n - 1), true
}

// ValidateTCPMapping binds the recorded signal sources to the receiver's unit.
func (p Profile) ValidateTCPMapping(uid string) error {
	if uid == "" || p.RetranslatorUID != uid || p.Vehicle.Unit != ReceiverUnit {
		return errors.New("el UID o la unidad de la ficha TCP difiere del receptor")
	}
	if p.IgnitionParam == p.ConnectionParam {
		return errors.New("se requieren fuentes distintas de ignición y conexión")
	}
	for _, selector := range []string{p.IgnitionParam, p.ConnectionParam} {
		if selector == "" || strings.TrimSpace(selector) != selector || strings.ContainsRune(selector, 0) || selector == "posinfo" || selector == "avl_inputs" {
			return errors.New("nombre de señal TCP inválido")
		}
		if strings.Contains(selector, ":") {
			if _, ok := DigitalInputMask(selector); !ok {
				return errors.New("selector TCP desconocido; se admite digital:in1 a digital:in32")
			}
		}
	}
	return nil
}
