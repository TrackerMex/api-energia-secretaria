package retranslator

import (
	"bytes"
	"encoding/binary"
	"math"

	"wialon-sener-pilot/internal/domain"
)

type signalReading struct {
	name  string
	bit   uint32
	seen  bool
	value *bool
}

func signalSource(selector string) signalReading {
	if mask, ok := domain.DigitalInputMask(selector); ok {
		return signalReading{name: "avl_inputs", bit: mask}
	}
	return signalReading{name: selector}
}

func (r *signalReading) observe(name string, kind byte, data []byte, flags uint32) {
	if name != r.name {
		return
	}
	if r.seen {
		r.value = nil
		return
	}
	r.seen = true
	if r.bit != 0 {
		if flags&2 != 0 && kind == 3 && len(data) == 4 {
			on := binary.BigEndian.Uint32(data)&r.bit != 0
			r.value = &on
		}
		return
	}
	r.value = twoState(kind, data)
}

func twoState(kind byte, data []byte) *bool {
	var value float64
	switch kind {
	case 1:
		switch string(bytes.TrimSuffix(data, []byte{0})) {
		case "true", "1":
			value = 1
		case "false", "0":
			value = 0
		default:
			return nil
		}
	case 3:
		value = float64(binary.BigEndian.Uint32(data))
	case 4:
		value = math.Float64frombits(binary.LittleEndian.Uint64(data))
	case 5:
		value = float64(binary.BigEndian.Uint64(data))
	default:
		return nil
	}
	if value != 0 && value != 1 {
		return nil
	}
	on := value == 1
	return &on
}
