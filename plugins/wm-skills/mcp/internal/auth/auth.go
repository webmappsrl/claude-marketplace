// Package auth legge le credenziali Orchestrator dal file condiviso con le skill.
// Non esegue mai l'accesso e non chiede mai la password: se il file manca o il
// segno di riconoscimento è scaduto, restituisce un errore che spiega come rimediare.
package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var (
	ErrMissing   = errors.New("file delle credenziali assente")
	ErrMalformed = errors.New("file delle credenziali non leggibile")
	ErrNoToken   = errors.New("file delle credenziali senza token")
)

// Credentials rispecchia orchestrator-auth.json.
type Credentials struct {
	Token string `json:"token"`
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// DefaultPath restituisce il percorso usato anche dalle skill.
func DefaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".config/webmapp/orchestrator-auth.json"
	}
	return filepath.Join(home, ".config", "webmapp", "orchestrator-auth.json")
}

// Load rilegge le credenziali dal disco. Va chiamata a ogni richiesta, così un
// nuovo accesso fatto altrove ha effetto senza riavviare il server.
func Load(path string) (Credentials, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Credentials{}, fmt.Errorf("%w: %s — esegui l'accesso a Orchestrator come descritto nella skill wm-plan", ErrMissing, path)
		}
		return Credentials{}, fmt.Errorf("%w: %s: %v", ErrMalformed, path, err)
	}

	var c Credentials
	if err := json.Unmarshal(raw, &c); err != nil {
		return Credentials{}, fmt.Errorf("%w: %s non contiene JSON valido", ErrMalformed, path)
	}
	if c.Token == "" {
		return Credentials{}, fmt.Errorf("%w: %s — rifai l'accesso a Orchestrator", ErrNoToken, path)
	}
	return c, nil
}
