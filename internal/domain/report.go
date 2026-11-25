package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Vehicle struct {
	Unit   string `json:"id_unidad"`
	Permit string `json:"numero_permiso"`
	Plate  string `json:"placa"`
	VIN    string `json:"vin"`
	Cargo  string `json:"tipo_carga"`
	Type   string `json:"tipo_vehiculo"`
}

// Profile contains the verified unit metadata and the device-specific signal mapping.
type Profile struct {
	RetranslatorUID string  `json:"retranslator_uid,omitempty"`
	UnitID          int64   `json:"wialon_unit_id"`
	RFC             string  `json:"rfc"`
	Name            string  `json:"razon_social"`
	Vehicle         Vehicle `json:"vehiculo"`
	IMEI            string  `json:"imei"`
	IgnitionParam   string  `json:"ignition_param"`
	ConnectionParam string  `json:"connection_param"`
}

var rfcFormat = regexp.MustCompile(`^[A-ZÑ&]{3,4}[0-9]{6}[A-Z0-9]{3}$`)
var vinFormat = regexp.MustCompile(`^[A-HJ-NPR-Z0-9]{17}$`)
var imeiFormat = regexp.MustCompile(`^[0-9]{15}$`)
var uuidFormat = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func (p Profile) validate() error {
	if !rfcFormat.MatchString(p.RFC) || strings.TrimSpace(p.Name) == "" {
		return errors.New("RFC o razón social inválidos")
	}
	if strings.TrimSpace(p.Vehicle.Unit) == "" || strings.TrimSpace(p.Vehicle.Permit) == "" || strings.TrimSpace(p.Vehicle.Plate) == "" {
		return errors.New("faltan identificador, número de permiso o placa de la unidad")
	}
	if !vinFormat.MatchString(p.Vehicle.VIN) || !imeiFormat.MatchString(p.IMEI) {
		return errors.New("VIN o IMEI de formato inválido")
	}
	switch p.Vehicle.Cargo {
	case "D", "P", "M", "T", "GLP":
	default:
		return errors.New("tipo_carga fuera del catálogo SENER")
	}
	switch p.Vehicle.Type {
	case "A", "T", "S":
	default:
		return errors.New("tipo_vehiculo fuera del catálogo SENER")
	}
	if strings.TrimSpace(p.IgnitionParam) == "" || strings.TrimSpace(p.ConnectionParam) == "" || p.IgnitionParam == p.ConnectionParam {
		return errors.New("la ficha requiere parámetros distintos de ignición y conexión")
	}
	return nil
}

type coordinates struct {
	Longitude *float64 `json:"x"`
	Latitude  *float64 `json:"y"`
	Direction *int     `json:"c"`
	Speed     *float64 `json:"s"`
}

type message struct {
	Flags     uint32                     `json:"f"`
	Inputs    *uint32                    `json:"i"`
	Timestamp int64                      `json:"t"`
	Position  *coordinates               `json:"pos"`
	Params    map[string]json.RawMessage `json:"p"`
}

// binarySignal accepts a recorded, two-state measurement. A zero value can be
// valid, whereas a missing, null or unsupported measurement must be reconciled.
func binarySignal(raw json.RawMessage) (bool, bool) {
	var value any
	if len(raw) == 0 || json.Unmarshal(raw, &value) != nil {
		return false, false
	}
	switch v := value.(type) {
	case bool:
		return v, true
	case float64:
		switch v {
		case 0:
			return false, true
		case 1:
			return true, true
		}
	case string:
		switch v {
		case "0", "false":
			return false, true
		case "1", "true":
			return true, true
		}
	}
	return false, false
}

func (m message) signal(selector string) (bool, bool) {
	index, digital := strings.CutPrefix(selector, "digital:in")
	if !digital {
		return binarySignal(m.Params[selector])
	}
	n, err := strconv.Atoi(index)
	if err != nil || n < 1 || n > 32 || m.Flags&0x02 == 0 || m.Inputs == nil {
		return false, false
	}
	return (*m.Inputs & (uint32(1) << (n - 1))) != 0, true
}

type geo struct {
	Latitude  float64 `json:"latitud"`
	Longitude float64 `json:"longitud"`
	Direction int     `json:"direccion_grados"`
}

type vehicleStatus struct {
	Speed    float64 `json:"velocidad_kmh"`
	Ignition bool    `json:"motor_encendido"`
}

type deviceStatus struct {
	State string `json:"estado_conexion"`
	IMEI  string `json:"imei"`
}

type transmission struct {
	ID      string        `json:"id_transmision"`
	GPS     string        `json:"timestamp_gps"`
	Kind    string        `json:"tipo_reporte"`
	Vehicle Vehicle       `json:"vehiculo"`
	Geo     geo           `json:"geolocalizacion"`
	Status  vehicleStatus `json:"estado_vehiculo"`
	Device  deviceStatus  `json:"estado_dispositivo_gps"`
	Sensors struct{}      `json:"sensores_adicionales"`
}

type envelope struct {
	RFC  string         `json:"rfc"`
	Name string         `json:"razon_social"`
	Data []transmission `json:"data"`
}

// Report encodes the position as a one-element SENER report. Its caller
// supplies the durably assigned transmission ID so retries keep the same ID.
func Report(p Position, profile Profile, transmissionID string) ([]byte, error) {
	if p.UnitID <= 0 || profile.UnitID != p.UnitID {
		return nil, errors.New("la ficha no corresponde al ID de la unidad recibida")
	}
	if err := profile.validate(); err != nil {
		return nil, err
	}
	var m message
	if err := json.Unmarshal(p.Message, &m); err != nil {
		return nil, errors.New("mensaje o fecha GPS inválidos")
	}
	return encodeReading(profile, transmissionID, m.Timestamp, m.Position, m.reading(profile.IgnitionParam), m.reading(profile.ConnectionParam))
}

func (m message) reading(selector string) *bool {
	value, ok := m.signal(selector)
	if !ok {
		return nil
	}
	return &value
}

// TCPReport uses each sensor value only if it was captured with this selector.
func TCPReport(fix GPSFix, profile Profile, transmissionID string) ([]byte, error) {
	if err := profile.ValidateTCPMapping(fix.UID); err != nil {
		return nil, err
	}
	if err := profile.validate(); err != nil {
		return nil, err
	}
	ignition, connected := fix.Ignition, fix.Connected
	if fix.IgnitionSource != profile.IgnitionParam {
		ignition = nil
	}
	if fix.ConnectionSource != profile.ConnectionParam {
		connected = nil
	}
	direction, speed := int(fix.Course), float64(fix.Speed)
	coord := &coordinates{Longitude: &fix.Longitude, Latitude: &fix.Latitude, Direction: &direction, Speed: &speed}
	return encodeReading(profile, transmissionID, fix.GPSTime.Unix(), coord, ignition, connected)
}

func encodeReading(profile Profile, transmissionID string, timestamp int64, coord *coordinates, ignition, connected *bool) ([]byte, error) {
	if !uuidFormat.MatchString(transmissionID) {
		return nil, errors.New("id_transmision debe ser un UUID")
	}
	if timestamp <= 0 || timestamp > 4294967295 {
		return nil, errors.New("mensaje o fecha GPS inválidos")
	}
	if coord == nil || coord.Latitude == nil || coord.Longitude == nil || *coord.Latitude < -90 || *coord.Latitude > 90 || *coord.Longitude < -180 || *coord.Longitude > 180 {
		return nil, errors.New("coordenadas GPS ausentes o fuera de rango")
	}
	if coord.Direction == nil || *coord.Direction < 0 || *coord.Direction > 359 {
		return nil, errors.New("rumbo GPS ausente o fuera de rango")
	}
	if coord.Speed == nil || *coord.Speed < 0 {
		return nil, errors.New("velocidad GPS ausente o negativa")
	}
	if ignition == nil {
		return nil, fmt.Errorf("lectura de ignición desconocida en %s", profile.IgnitionParam)
	}
	if connected == nil {
		return nil, fmt.Errorf("lectura de conexión desconocida en %s", profile.ConnectionParam)
	}
	state := "desconectado"
	if *connected {
		state = "conectado"
	}
	row := transmission{
		ID: transmissionID, GPS: time.Unix(timestamp, 0).UTC().Format(time.RFC3339), Kind: "posicion",
		Vehicle: profile.Vehicle, Geo: geo{*coord.Latitude, *coord.Longitude, *coord.Direction},
		Status: vehicleStatus{*coord.Speed, *ignition}, Device: deviceStatus{state, profile.IMEI},
	}
	return json.Marshal(envelope{profile.RFC, profile.Name, []transmission{row}})
}
