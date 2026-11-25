package domain

import (
	"bytes"
	"encoding/json"
	"strconv"
)

// ObservationSummary describes one window of position messages without identifying its unit.
type ObservationSummary struct {
	Total           int            `json:"total"`
	Encendido       int            `json:"encendido"`
	Apagado         int            `json:"apagado"`
	Indeterminado   int            `json:"indeterminado"`
	Motivos         map[string]int `json:"motivos"`
	CriterioMotivos string         `json:"criterio_motivos"`
	ConexionGPS     string         `json:"conexion_gps"`
}

// SummarizeObservation counts each position once and compares signals from that message alone.
func SummarizeObservation(positions []Position) ObservationSummary {
	result := ObservationSummary{
		Total:           len(positions),
		Motivos:         make(map[string]int),
		CriterioMotivos: "Una posición puede sumar un motivo por la entrada y otro por las RPM. Una discrepancia o un mensaje ilegible suman un solo motivo.",
		ConexionGPS:     "pendiente_de_validacion",
	}
	for _, position := range positions {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(position.Message, &fields); err != nil || fields == nil {
			result.Indeterminado++
			result.Motivos["mensaje_invalido"]++
			continue
		}
		ignition, inputReason := observeIgnition(fields["f"], fields["i"])
		running, rpmReason := observeRPM(fields["p"])
		switch {
		case inputReason != "" || rpmReason != "":
			result.Indeterminado++
			if inputReason != "" {
				result.Motivos[inputReason]++
			}
			if rpmReason != "" {
				result.Motivos[rpmReason]++
			}
		case ignition != running:
			result.Indeterminado++
			result.Motivos["discrepancia"]++
		case ignition:
			result.Encendido++
		default:
			result.Apagado++
		}
	}
	return result
}

func missingObservationValue(raw json.RawMessage) bool {
	return len(raw) == 0 || string(raw) == "null"
}

func observeIgnition(rawFlags, rawInputs json.RawMessage) (bool, string) {
	if missingObservationValue(rawFlags) {
		return false, "entrada_ausente"
	}
	var flags, inputs uint32
	if err := json.Unmarshal(rawFlags, &flags); err != nil {
		return false, "entrada_invalida"
	}
	if flags&0x02 == 0 || missingObservationValue(rawInputs) {
		return false, "entrada_ausente"
	}
	if err := json.Unmarshal(rawInputs, &inputs); err != nil {
		return false, "entrada_invalida"
	}
	active, _ := (message{Flags: flags, Inputs: &inputs}).signal("digital:in1")
	return active, ""
}

func observeRPM(rawParams json.RawMessage) (bool, string) {
	if missingObservationValue(rawParams) {
		return false, "rpm_ausentes"
	}
	var params map[string]json.RawMessage
	if err := json.Unmarshal(rawParams, &params); err != nil {
		return false, "rpm_invalidas"
	}
	raw := params["can_rpm"]
	if missingObservationValue(raw) {
		return false, "rpm_ausentes"
	}
	value, err := strconv.ParseFloat(string(raw), 64)
	if err != nil || value < 0 {
		return false, "rpm_invalidas"
	}
	// RPM = can_rpm / 10. Comparing signs before dividing also works for
	// subnormal values whose scaled result would round to zero.
	if value > 0 {
		return true, ""
	}
	// ParseFloat can also round a nonzero JSON literal to signed zero.
	mantissa := raw
	if exponent := bytes.IndexAny(raw, "eE"); exponent >= 0 {
		mantissa = raw[:exponent]
	}
	if !bytes.ContainsAny(mantissa, "123456789") {
		return false, ""
	}
	if mantissa[0] == '-' {
		return false, "rpm_invalidas"
	}
	return true, ""
}
