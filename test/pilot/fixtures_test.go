package pilot

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

func readFixture(t *testing.T, name string, target any) {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		t.Fatal(err)
	}
}

// The files are reference examples for development, not live telemetry.
func TestSingleUnitFixtures(t *testing.T) {
	var source struct {
		Count    int
		Messages []struct {
			Time int64 `json:"t"`
		}
	}
	var report struct {
		Data []struct {
			Time    time.Time `json:"timestamp_gps"`
			Vehicle struct {
				Unit string `json:"id_unidad"`
			} `json:"vehiculo"`
		}
	}
	readFixture(t, "wialon.json", &source)
	readFixture(t, "sener.json", &report)
	if source.Count != 1 || len(source.Messages) != 1 || len(report.Data) != 1 {
		t.Fatal("expected one Wialon message and one SENER report")
	}
	transmission := report.Data[0]
	if transmission.Vehicle.Unit != "DEMO-01" || source.Messages[0].Time <= 0 || !transmission.Time.Equal(time.Unix(source.Messages[0].Time, 0)) {
		t.Fatal("fixtures must describe the same demo unit scenario and GPS time")
	}
}
