package ingest

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestR3VerifiedDigitalInput(t *testing.T) {
	for _, tc := range []struct {
		name, selector, flagsAndInput string
		on, valid                     bool
	}{
		{"first input on", "digital:in1", `"f":3,"i":5`, true, true},
		{"first input off", "digital:in1", `"f":3,"i":2`, false, true},
		{"known all off", "digital:in1", `"f":3,"i":0`, false, true},
		{"fourth input on", "digital:in4", `"f":3,"i":8`, true, true},
		{"no flags", "digital:in1", `"f":1,"i":1`, false, false},
		{"no inputs", "digital:in1", `"f":3`, false, false},
		{"out of range", "digital:in33", `"f":3,"i":1`, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			profile := demoProfile(t)
			profile.IgnitionParam = tc.selector
			reading := strings.Replace(sample, `"f":3`, tc.flagsAndInput, 1)
			data, err := Report(Position{UnitID: 7, Message: json.RawMessage(reading)}, profile, "00000000-0000-4000-8000-000000000001")
			if !tc.valid {
				if err == nil {
					t.Fatal("accepted unknown input")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			var report envelope
			if err := json.Unmarshal(data, &report); err != nil {
				t.Fatal(err)
			}
			if report.Data[0].Status.Ignition != tc.on {
				t.Errorf("ignition should be %t", tc.on)
			}
		})
	}
}
