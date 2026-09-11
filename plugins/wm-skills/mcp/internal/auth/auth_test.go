package auth

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func write(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "auth.json")
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestLoadValid(t *testing.T) {
	p := write(t, `{"token":"abc123","id":12,"name":"Giuseppe","email":"g@webmapp.it"}`)
	c, err := Load(p)
	if err != nil {
		t.Fatalf("errore inatteso: %v", err)
	}
	if c.Token != "abc123" || c.ID != 12 || c.Name != "Giuseppe" {
		t.Fatalf("credenziali lette male: %+v", c)
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "assente.json"))
	if !errors.Is(err, ErrMissing) {
		t.Fatalf("atteso ErrMissing, ottenuto %v", err)
	}
}

func TestLoadMalformed(t *testing.T) {
	p := write(t, `{non json`)
	_, err := Load(p)
	if !errors.Is(err, ErrMalformed) {
		t.Fatalf("atteso ErrMalformed, ottenuto %v", err)
	}
}

func TestLoadEmptyToken(t *testing.T) {
	p := write(t, `{"token":"","id":12}`)
	_, err := Load(p)
	if !errors.Is(err, ErrNoToken) {
		t.Fatalf("atteso ErrNoToken, ottenuto %v", err)
	}
}
