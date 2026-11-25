package domain

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestR2ObservationClassifiesSameMessage(t *testing.T) {
	for _, tc := range []struct {
		name, raw string
		on, off   int
		reason    string
	}{
		{"active and positive rpm", `{"f":3,"i":3,"p":{"can_rpm":8500}}`, 1, 0, ""},
		{"inactive and measured zero", `{"f":3,"i":0,"p":{"can_rpm":0},"pos":{"s":90}}`, 0, 1, ""},
		{"other input set", `{"f":3,"i":2,"p":{"can_rpm":0}}`, 0, 1, ""},
		{"upper boundary and fractional rpm", `{"f":4294967295,"i":4294967295,"p":{"can_rpm":0.05}}`, 1, 0, ""},
		{"no undocumented device sentinel", `{"f":3,"i":1,"p":{"can_rpm":65535}}`, 1, 0, ""},
		{"active with zero rpm", `{"f":2,"i":1,"p":{"can_rpm":0}}`, 0, 0, "discrepancia"},
		{"inactive with positive rpm", `{"f":2,"i":0,"p":{"can_rpm":10}}`, 0, 0, "discrepancia"},
		{"missing flags", `{"i":0,"p":{"can_rpm":0}}`, 0, 0, "entrada_ausente"},
		{"flag not set", `{"f":1,"i":0,"p":{"can_rpm":0}}`, 0, 0, "entrada_ausente"},
		{"mask omitted", `{"f":3,"p":{"can_rpm":0}}`, 0, 0, "entrada_ausente"},
		{"null mask", `{"f":3,"i":null,"p":{"can_rpm":0}}`, 0, 0, "entrada_ausente"},
		{"flags too large", `{"f":4294967298,"i":0,"p":{"can_rpm":0}}`, 0, 0, "entrada_invalida"},
		{"fractional mask", `{"f":3,"i":0.5,"p":{"can_rpm":0}}`, 0, 0, "entrada_invalida"},
		{"negative mask", `{"f":3,"i":-1,"p":{"can_rpm":0}}`, 0, 0, "entrada_invalida"},
		{"quoted mask", `{"f":3,"i":"0","p":{"can_rpm":0}}`, 0, 0, "entrada_invalida"},
		{"missing params", `{"f":3,"i":0}`, 0, 0, "rpm_ausentes"},
		{"missing rpm", `{"f":3,"i":0,"p":{"other":0}}`, 0, 0, "rpm_ausentes"},
		{"null rpm", `{"f":3,"i":0,"p":{"can_rpm":null}}`, 0, 0, "rpm_ausentes"},
		{"invalid params", `{"f":3,"i":0,"p":[]}`, 0, 0, "rpm_invalidas"},
		{"quoted rpm", `{"f":3,"i":0,"p":{"can_rpm":"0"}}`, 0, 0, "rpm_invalidas"},
		{"negative rpm", `{"f":3,"i":0,"p":{"can_rpm":-1}}`, 0, 0, "rpm_invalidas"},
		{"numeric overflow", `{"f":3,"i":0,"p":{"can_rpm":1e400}}`, 0, 0, "rpm_invalidas"},
		{"positive near underflow", `{"f":3,"i":1,"p":{"can_rpm":5e-324}}`, 1, 0, ""},
		{"negative near underflow", `{"f":3,"i":0,"p":{"can_rpm":-1e-400}}`, 0, 0, "rpm_invalidas"},
		{"unreadable JSON", `{"f":3`, 0, 0, "mensaje_invalido"},
		{"nonobject", `null`, 0, 0, "mensaje_invalido"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := SummarizeObservation([]Position{{Message: json.RawMessage(tc.raw)}})
			want := map[string]int{}
			if tc.reason != "" {
				want[tc.reason] = 1
			}
			if got.Total != 1 || got.Encendido != tc.on || got.Apagado != tc.off || got.Indeterminado != 1-tc.on-tc.off || !reflect.DeepEqual(got.Motivos, want) {
				t.Fatalf("got %+v", got)
			}
		})
	}
}

func TestR3ObservationCountsPositionsAndExplainsReasons(t *testing.T) {
	messages := []string{
		`{"f":3,"i":1,"p":{"can_rpm":200}}`,
		`{"f":3,"i":1,"p":{"can_rpm":200}}`,
		`{"f":3,"i":0,"p":{"can_rpm":0}}`,
		`{"f":3,"i":1,"p":{"can_rpm":0}}`,
		`{"f":3,"i":1,"pos":{"x":123.456}}`,
		`{"p":{"can_rpm":100}}`,
		`{"p":{},"f":2}`,
	}
	var rows []Position
	for _, raw := range messages {
		rows = append(rows, Position{UnitID: 56789012, Key: "FAKE-UNIT", Message: json.RawMessage(raw)})
	}
	got := SummarizeObservation(rows)
	if got.Total != 7 || got.Encendido != 2 || got.Apagado != 1 || got.Indeterminado != 4 || got.Total != got.Encendido+got.Apagado+got.Indeterminado {
		t.Fatal(got)
	}
	if !reflect.DeepEqual(got.Motivos, map[string]int{"discrepancia": 1, "entrada_ausente": 2, "rpm_ausentes": 2}) {
		t.Fatal(got.Motivos)
	}
	if got.ConexionGPS != "pendiente_de_validacion" || got.CriterioMotivos == "" {
		t.Fatal(got)
	}
	raw, err := json.Marshal(got)
	if err != nil || strings.Contains(string(raw), "FAKE-UNIT") || strings.Contains(string(raw), "56789012") || strings.Contains(string(raw), "123.456") {
		t.Fatalf("output=%s err=%v", raw, err)
	}
}

func TestR3EmptyObservationKeepsStableJSON(t *testing.T) {
	raw, err := json.Marshal(SummarizeObservation(nil))
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if got["total"] != 0.0 || got["encendido"] != 0.0 || got["apagado"] != 0.0 || got["indeterminado"] != 0.0 || got["conexion_gps"] != "pendiente_de_validacion" || !reflect.DeepEqual(got["motivos"], map[string]any{}) || got["criterio_motivos"] == nil {
		t.Fatal(string(raw))
	}
}
