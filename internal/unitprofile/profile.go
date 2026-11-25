// Package unitprofile loads the receiver's explicitly bound single-unit profile.
package unitprofile

import (
	"encoding/json"
	"errors"
	"io"
	"os"

	"wialon-sener-pilot/internal/domain"
)

func LoadTCP(path, uid string) (domain.Profile, error) {
	var p domain.Profile
	file, err := os.Open(path)
	if err != nil {
		return p, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&p); err != nil {
		return p, errors.New("JSON de la ficha TCP inválido")
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return p, errors.New("contenido adicional en la ficha TCP")
	}
	return p, p.ValidateTCPMapping(uid)
}
