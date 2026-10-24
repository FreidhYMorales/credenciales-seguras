// Package testdata embeds the 20 test pairs and the oracle fixtures.
package testdata

import (
	"embed"
	"encoding/json"
)

// FS holds testdata/pares.json.
//
//go:embed pares.json
var FS embed.FS

// Pair is one row of testdata/pares.json: a password and an email with the
// verdict each one must produce.
type Pair struct {
	ID               int     `json:"id"`
	Email            string  `json:"email"`
	Password         string  `json:"password"`
	EmailEsperado    bool    `json:"email_esperado"`
	PasswordEsperado bool    `json:"password_esperado"`
	ReglaFallida     *string `json:"regla_fallida"`
	SeedEmail        *int64  `json:"seed_email"`
	SeedPassword     *int64  `json:"seed_password"`
}

// Pairs decodes the embedded file.
func Pairs() ([]Pair, error) {
	raw, err := FS.ReadFile("pares.json")
	if err != nil {
		return nil, err
	}
	var out []Pair
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}
