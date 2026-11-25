package retranslator

import (
	"encoding/binary"
	"testing"
	"wialon-sener-pilot/internal/domain"
)

func syntheticBlock(name string, kind byte, v []byte) []byte {
	body := append([]byte{0, kind}, []byte(name)...)
	body = append(body, 0)
	body = append(body, v...)
	header := binary.BigEndian.AppendUint16(nil, 0x0bbb)
	header = binary.BigEndian.AppendUint32(header, uint32(len(body)))
	return append(header, body...)
}

func syntheticFrame(flags uint32, blocks ...[]byte) []byte {
	body := append([]byte("DEMO-CR13"), 0)
	body = binary.BigEndian.AppendUint32(body, 60)
	body = binary.BigEndian.AppendUint32(body, flags)
	body = append(body, syntheticBlock("posinfo", 2, make([]byte, 29))...)
	for _, b := range blocks {
		body = append(body, b...)
	}
	return append(binary.LittleEndian.AppendUint32(nil, uint32(len(body))), body...)
}

func TestR1MappedSignalsBelongToThisFrame(t *testing.T) {
	profile := domain.Profile{IgnitionParam: "contact", ConnectionParam: "link"}
	ign := func(n uint32) []byte { return syntheticBlock("contact", 3, binary.BigEndian.AppendUint32(nil, n)) }
	on := syntheticBlock("link", 1, []byte("true\x00"))
	for _, tc := range []struct {
		name                string
		blocks              [][]byte
		ignition, connected string
	}{
		{"measured zero", [][]byte{ign(0), on}, "false", "true"},
		{"missing", [][]byte{ign(1)}, "true", "unknown"},
		{"unsupported", [][]byte{ign(2), on}, "unknown", "true"},
		{"ambiguous", [][]byte{ign(0), ign(1), on}, "unknown", "true"},
		{"wrong name", [][]byte{syntheticBlock("ign", 3, []byte{0, 0, 0, 1})}, "unknown", "unknown"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fix, ok, err := decodeWithProfile(syntheticFrame(1, tc.blocks...), profile)
			if err != nil || !ok || state(fix.Ignition) != tc.ignition || state(fix.Connected) != tc.connected ||
				fix.IgnitionSource != "contact" || fix.ConnectionSource != "link" {
				t.Fatalf("position=%v signals=%s/%s sources=%s/%s err=%v", ok, state(fix.Ignition), state(fix.Connected), fix.IgnitionSource, fix.ConnectionSource, err)
			}
		})
	}
}

func TestR1DigitalBitRequiresMaskAndFlag(t *testing.T) {
	p := domain.Profile{IgnitionParam: "digital:in32", ConnectionParam: "digital:in1"}
	mask := syntheticBlock("avl_inputs", 3, []byte{0x80, 0, 0, 0})
	for _, tc := range []struct {
		name     string
		flags    uint32
		blocks   [][]byte
		ign, con string
	}{
		{"high and low bits", 3, [][]byte{mask}, "true", "false"},
		{"mask missing", 3, nil, "unknown", "unknown"},
		{"flag missing", 1, [][]byte{mask}, "unknown", "unknown"},
		{"wrong type", 3, [][]byte{syntheticBlock("avl_inputs", 2, []byte{1})}, "unknown", "unknown"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fix, ok, err := decodeWithProfile(syntheticFrame(tc.flags, tc.blocks...), p)
			if err != nil || !ok || state(fix.Ignition) != tc.ign || state(fix.Connected) != tc.con {
				t.Fatalf("position=%v signals=%s/%s err=%v", ok, state(fix.Ignition), state(fix.Connected), err)
			}
		})
	}
}

func state(v *bool) string {
	if v == nil {
		return "unknown"
	}
	if *v {
		return "true"
	}
	return "false"
}
