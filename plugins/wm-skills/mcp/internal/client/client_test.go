package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func clientForTest(t *testing.T, h http.Handler) *Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	authPath := filepath.Join(t.TempDir(), "auth.json")
	if err := os.WriteFile(authPath, []byte(`{"token":"t0k3n","id":12}`), 0o600); err != nil {
		t.Fatal(err)
	}
	return &Client{BaseURL: srv.URL, AuthPath: authPath, HTTP: srv.Client()}
}

func TestDoSendsBearerToken(t *testing.T) {
	var got string
	c := clientForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("Authorization")
		w.Write([]byte(`{"id":1}`))
	}))

	if _, err := c.Do(context.Background(), "GET", "/api/stories/1", nil); err != nil {
		t.Fatalf("errore inatteso: %v", err)
	}
	if got != "Bearer t0k3n" {
		t.Fatalf("intestazione sbagliata: %q", got)
	}
}

func TestDoTranslatesValidationError(t *testing.T) {
	c := clientForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(422)
		w.Write([]byte(`{"message":"The given data was invalid.","errors":{"type":["Il valore selezionato per type non è valido."]}}`))
	}))

	_, err := c.Do(context.Background(), "POST", "/api/stories", map[string]any{"type": "Task"})
	if err == nil {
		t.Fatal("atteso un errore")
	}
	var apiErr *APIError
	if !asAPIError(err, &apiErr) {
		t.Fatalf("atteso *APIError, ottenuto %T", err)
	}
	if apiErr.Status != 422 {
		t.Fatalf("stato sbagliato: %d", apiErr.Status)
	}
	if !strings.Contains(apiErr.Error(), "type") {
		t.Fatalf("il messaggio deve nominare il campo rifiutato: %s", apiErr.Error())
	}
}

func TestDoTranslatesUnauthorized(t *testing.T) {
	c := clientForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		w.Write([]byte(`{"message":"Unauthenticated."}`))
	}))

	_, err := c.Do(context.Background(), "GET", "/api/stories/1", nil)
	if err == nil || !strings.Contains(err.Error(), "accesso") {
		t.Fatalf("il messaggio deve suggerire di rifare l'accesso: %v", err)
	}
}

func TestDoReturnsRawBody(t *testing.T) {
	c := clientForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"id":8503,"name":"Titolo"}`))
	}))

	raw, err := c.Do(context.Background(), "GET", "/api/stories/8503", nil)
	if err != nil {
		t.Fatalf("errore inatteso: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if got["name"] != "Titolo" {
		t.Fatalf("corpo sbagliato: %v", got)
	}
}

func asAPIError(err error, target **APIError) bool {
	e, ok := err.(*APIError)
	if ok {
		*target = e
	}
	return ok
}

func TestNewWithAuthPathUsesGivenPath(t *testing.T) {
	authPath := filepath.Join(t.TempDir(), "auth-dev.json")
	if err := os.WriteFile(authPath, []byte(`{"token":"d3v-t0k3n","id":1}`), 0o600); err != nil {
		t.Fatal(err)
	}

	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("Authorization")
		w.Write([]byte(`{"id":1}`))
	}))
	t.Cleanup(srv.Close)

	c := NewWithAuthPath(srv.URL, authPath)
	if c.AuthPath != authPath {
		t.Fatalf("il client non usa il percorso passato esplicitamente: %q", c.AuthPath)
	}

	c.HTTP = srv.Client()
	if _, err := c.Do(context.Background(), "GET", "/api/stories/1", nil); err != nil {
		t.Fatalf("errore inatteso: %v", err)
	}
	if got != "Bearer d3v-t0k3n" {
		t.Fatalf("il client non ha letto le credenziali dal file passato esplicitamente: %q", got)
	}
}

func TestNewWithAuthPathNamesMissingFile(t *testing.T) {
	authPath := filepath.Join(t.TempDir(), "assente.json")
	c := NewWithAuthPath("http://example.invalid", authPath)

	_, err := c.Do(context.Background(), "GET", "/api/stories/1", nil)
	if err == nil {
		t.Fatal("un file credenziali assente deve produrre un errore")
	}
	if !strings.Contains(err.Error(), authPath) {
		t.Fatalf("l'errore deve nominare il percorso cercato (%q): %v", authPath, err)
	}
}
