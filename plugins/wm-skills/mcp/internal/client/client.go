// Package client incapsula le chiamate HTTP all'API di Orchestrator e traduce
// gli errori in messaggi utilizzabili invece che in risposte grezze.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/webmappsrl/claude-marketplace/plugins/wm-skills/mcp/internal/auth"
)

const defaultBaseURL = "https://orchestrator.maphub.it"

type Client struct {
	BaseURL  string
	AuthPath string
	HTTP     *http.Client
}

// New costruisce il client sull'indirizzo indicato, con il file delle
// credenziali predefinito (`auth.DefaultPath()`). Con stringa vuota per
// l'indirizzo usa la produzione: l'indirizzo non arriva mai dall'ambiente,
// così non può essere ereditato per sbaglio da una variabile impostata per
// altri scopi.
func New(base string) *Client {
	return NewWithAuthPath(base, auth.DefaultPath())
}

// NewWithAuthPath costruisce il client sull'indirizzo e sul file delle
// credenziali indicati. Serve a tenere gli ambienti separati anche sulle
// credenziali, non solo sull'indirizzo: due server puntati a due istanze
// diverse di Orchestrator devono poter leggere due file diversi, altrimenti
// il segno di riconoscimento di uno dei due ambienti non è valido nell'altro.
func NewWithAuthPath(base, authPath string) *Client {
	if base == "" {
		base = defaultBaseURL
	}
	if authPath == "" {
		authPath = auth.DefaultPath()
	}
	return &Client{
		BaseURL:  strings.TrimRight(base, "/"),
		AuthPath: authPath,
		HTTP:     &http.Client{Timeout: 30 * time.Second},
	}
}

// APIError rappresenta una risposta di errore di Orchestrator già interpretata.
type APIError struct {
	Status  int
	Message string
	Fields  map[string][]string
}

func (e *APIError) Error() string {
	if len(e.Fields) == 0 {
		return fmt.Sprintf("Orchestrator ha risposto %d: %s", e.Status, e.Message)
	}
	names := make([]string, 0, len(e.Fields))
	for k := range e.Fields {
		names = append(names, k)
	}
	sort.Strings(names)

	var b strings.Builder
	fmt.Fprintf(&b, "Orchestrator ha rifiutato i dati (%d):", e.Status)
	for _, n := range names {
		fmt.Fprintf(&b, "\n  %s: %s", n, strings.Join(e.Fields[n], " "))
	}
	return b.String()
}

func (c *Client) Do(ctx context.Context, method, path string, body any) (json.RawMessage, error) {
	creds, err := auth.Load(c.AuthPath)
	if err != nil {
		return nil, err
	}

	var payload io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("impossibile costruire il messaggio da inviare: %w", err)
		}
		payload = bytes.NewReader(encoded)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, payload)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+creds.Token)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Orchestrator non raggiungibile su %s: %w — controlla la rete", c.BaseURL, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("risposta di Orchestrator illeggibile: %w", err)
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return raw, nil
	}
	return nil, translate(resp.StatusCode, raw)
}

func translate(status int, raw []byte) error {
	var parsed struct {
		Message string              `json:"message"`
		Errors  map[string][]string `json:"errors"`
	}
	_ = json.Unmarshal(raw, &parsed)

	switch status {
	case 401:
		return &APIError{Status: status, Message: "segno di riconoscimento assente o scaduto — rifai l'accesso a Orchestrator"}
	case 403:
		msg := parsed.Message
		if msg == "" {
			msg = "operazione non permessa al tuo utente"
		}
		return &APIError{Status: status, Message: msg}
	case 404:
		return &APIError{Status: status, Message: "risorsa inesistente: " + firstNonEmpty(parsed.Message, "controlla l'identificatore")}
	case 422:
		return &APIError{Status: status, Message: parsed.Message, Fields: parsed.Errors}
	default:
		return &APIError{Status: status, Message: firstNonEmpty(parsed.Message, string(raw))}
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
