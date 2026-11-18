package ingest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"wialon-sener-pilot/internal/domain"
)

const sample = `{"t":100,"tp":"ud","f":3,"pos":{"x":-100,"y":20,"c":90,"s":0},"p":{"ign":1,"online":1}}`
const profileJSON = `{"wialon_unit_id":7,"rfc":"XAXX010101000","razon_social":"DEMO","vehiculo":{"id_unidad":"DEMO-01","numero_permiso":"FICTICIO","placa":"DEMO001","vin":"DEMX0000000000001","tipo_carga":"M","tipo_vehiculo":"A"},"imei":"000000000000000","ignition_param":"ign","connection_param":"online"}`

func demoProfile(t *testing.T) domain.Profile {
	t.Helper()
	var profile domain.Profile
	if err := json.Unmarshal([]byte(profileJSON), &profile); err != nil {
		t.Fatal(err)
	}
	return profile
}

func TestR1FetchAllPositionsForOneUnit(t *testing.T) {
	var services []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/wialon/ajax.html" || r.URL.RawQuery != "" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL)
		}
		svc := r.FormValue("svc")
		services = append(services, svc)
		if svc != "token/login" && r.FormValue("sid") != "fake-session" {
			t.Error("missing session")
		}
		var params map[string]any
		if err := json.Unmarshal([]byte(r.FormValue("params")), &params); err != nil {
			t.Error(err)
		}
		switch svc {
		case "token/login":
			if params["token"] != "fake" {
				t.Error("missing token")
			}
			w.Write([]byte(`{"eid":"fake-session"}`))
		case "messages/load_interval":
			if params["itemId"] != float64(7) || params["timeFrom"] != float64(100) || params["timeTo"] != float64(110) {
				t.Errorf("incorrect filters %v", params)
			}
			w.Write([]byte(`{"count":3,"messages":[` + sample + `]}`))
		case "messages/get_messages":
			if params["indexFrom"] != float64(1) || params["indexTo"] != float64(3) {
				t.Errorf("incorrect page %v", params)
			}
			reordered := `{"p":{"online":1,"ign":1},"pos":{"s":0,"c":90,"y":20,"x":-100},"f":3,"tp":"ud","t":100}`
			w.Write([]byte("[" + reordered + "," + strings.Replace(sample, `"ign":1`, `"ign":0`, 1) + "]"))
		case "core/logout":
			w.Write([]byte(`{"error":0}`))
		default:
			t.Errorf("unexpected service %s", svc)
		}
	}))
	defer server.Close()
	source := Source{BaseURL: server.URL, Token: "fake", Client: server.Client()}
	rows, err := source.Positions(context.Background(), 7, time.Unix(100, 0), time.Unix(110, 0))
	if err != nil || len(rows) != 3 {
		t.Fatalf("rows=%v err=%v", rows, err)
	}
	if rows[0].Key == "" || rows[0].Key != rows[1].Key || rows[0].Key == rows[2].Key {
		t.Error("unstable or incomplete message identity")
	}
	if strings.Join(services, ",") != "token/login,messages/load_interval,messages/get_messages,core/logout" {
		t.Errorf("requests %v", services)
	}
}

func TestR1IncompletePageAndErrors(t *testing.T) {
	for _, reply := range []string{`[]`, `{"error":7,"detail":"do-not-log"}`, `null`} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.FormValue("svc") {
			case "token/login":
				w.Write([]byte(`{"eid":"session"}`))
			case "messages/load_interval":
				w.Write([]byte(`{"count":2,"messages":[]}`))
			case "messages/get_messages":
				w.Write([]byte(reply))
			case "core/logout":
				w.Write([]byte(`{"error":0}`))
			}
		}))
		src := Source{BaseURL: server.URL, Token: "fake", Client: server.Client()}
		rows, err := src.Positions(context.Background(), 7, time.Unix(100, 0), time.Unix(110, 0))
		server.Close()
		if rows != nil || err == nil || strings.Contains(err.Error(), "do-not-log") {
			t.Fatalf("malformed page not rejected privately: %v", err)
		}
	}
}

func TestR2BuildSENERPosition(t *testing.T) {
	p := domain.Position{UnitID: 7, Message: json.RawMessage(sample)}
	const uuid = "00000000-0000-4000-8000-000000000001"
	b, err := domain.Report(p, demoProfile(t), uuid)
	if err != nil {
		t.Fatal(err)
	}
	var want, got map[string]any
	expected := `{"rfc":"XAXX010101000","razon_social":"DEMO","data":[{"id_transmision":"00000000-0000-4000-8000-000000000001","timestamp_gps":"1970-01-01T00:01:40Z","tipo_reporte":"posicion","vehiculo":{"id_unidad":"DEMO-01","numero_permiso":"FICTICIO","placa":"DEMO001","vin":"DEMX0000000000001","tipo_carga":"M","tipo_vehiculo":"A"},"geolocalizacion":{"latitud":20,"longitud":-100,"direccion_grados":90},"estado_vehiculo":{"velocidad_kmh":0,"motor_encendido":true},"estado_dispositivo_gps":{"imei":"000000000000000","estado_conexion":"conectado"},"sensores_adicionales":{}}]}`
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(expected), &want); err != nil {
		t.Fatal(err)
	}
	sortedGot, _ := json.Marshal(got)
	sortedWant, _ := json.Marshal(want)
	if string(sortedGot) != string(sortedWant) {
		t.Fatalf("got %s, want %s", sortedGot, sortedWant)
	}
}

func TestR3RejectMissingAndUnknownRequiredValues(t *testing.T) {
	p := demoProfile(t)
	for _, invalid := range []string{
		strings.Replace(sample, `"ign":1`, `"ign":2`, 1),
		strings.Replace(sample, `"online":1`, `"online":null`, 1),
		strings.Replace(sample, `"s":0`, `"other":0`, 1),
		strings.Replace(sample, `"y":20`, `"y":100`, 1),
		strings.Replace(sample, `"t":100`, `"other":100`, 1),
	} {
		if _, err := domain.Report(domain.Position{UnitID: 7, Message: json.RawMessage(invalid)}, p, "00000000-0000-4000-8000-000000000001"); err == nil {
			t.Fatalf("accepted invalid message: %s", invalid)
		}
	}
	for _, profileText := range []string{
		strings.Replace(profileJSON, `"M"`, `"Magna"`, 1),
		strings.Replace(profileJSON, `"wialon_unit_id":7`, `"wialon_unit_id":8`, 1),
		strings.Replace(profileJSON, "DEMX0000000000001", "VIN-CORTO", 1),
	} {
		var bad domain.Profile
		if err := json.Unmarshal([]byte(profileText), &bad); err != nil {
			t.Fatal(err)
		}
		if _, err := domain.Report(domain.Position{UnitID: 7, Message: json.RawMessage(sample)}, bad, "00000000-0000-4000-8000-000000000001"); err == nil {
			t.Fatal("accepted invalid profile")
		}
	}
}
