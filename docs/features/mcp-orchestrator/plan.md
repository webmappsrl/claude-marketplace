# Server MCP Orchestrator — Piano di implementazione

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Dotare il plugin `wm-skills` di un server MCP in Go che espone l'API di Orchestrator come tool tipizzati, e riscrivere le tre skill perché lo usino al posto dei comandi `curl`.

**Architecture:** Programma Go che comunica sui canali standard del processo (stdio), avviato da Claude Code tramite `.mcp.json` del plugin. Non contiene logica di lavoro: legge le credenziali dal file già in uso, gira le chiamate all'API, e restituisce risposte e errori in forma leggibile. Nomi e descrizioni dei tool sono dichiarati a mano; campi, tipi e valori ammessi si leggono dalla specifica OpenAPI pubblica.

**Tech Stack:** Go 1.24+, `github.com/modelcontextprotocol/go-sdk` v1.7.0, solo libreria standard per il resto.

**Spec:** `docs/features/mcp-orchestrator/overview.md`

## Global Constraints

- Modulo Go: `github.com/webmappsrl/claude-marketplace/plugins/wm-skills/mcp`
- SDK MCP: `github.com/modelcontextprotocol/go-sdk` versione `v1.7.0` esatta
- Nessuna dipendenza esterna oltre all'SDK MCP
- Compilazione: solo `darwin/arm64`, binario in `plugins/wm-skills/bin/orchestrator-mcp`
- Credenziali: `~/.config/webmapp/orchestrator-auth.json`, campi `token`, `id`, `name`, `email` — rilette a ogni chiamata, mai tenute in memoria fra una chiamata e l'altra
- **Indirizzo base: argomento `--base-url`, valore predefinito `https://orchestrator.maphub.it`.** Il binario non legge nessuna variabile d'ambiente per l'indirizzo, così non può ereditarne una impostata per altri scopi. Il `.mcp.json` del plugin non passa l'argomento: il plugin distribuito parla solo con la produzione. Solo il repo `claude-marketplace` dichiara un secondo server `orchestrator-dev` che passa `--base-url` verso l'istanza locale
- Il server non esegue mai l'accesso e non chiede mai la password
- Tool di scrittura: parametro `confirm bool`, valore predefinito `false` (senza conferma mostra la differenza e non scrive)
- Tool di lettura: nessuna conferma
- Prove: mai contro la produzione. Istanza locale via Docker, raggiunta tramite il server `orchestrator-dev` dichiarato in questo repo
- Identificatori di codice, nomi dei tool e messaggi del protocollo in inglese; commenti e documentazione in italiano
- Ramo, revisione del diff e commit sono governati da `wm-skills:wm-plan` (`execution: branch`, `execution: review-gate`): questo piano non li ridefinisce e i suoi compiti non toccano git

## Struttura dei file

```
plugins/wm-skills/
  .mcp.json                          dichiarazione del server per il plugin
  bin/orchestrator-mcp               binario compilato (versionato)
  mcp/
    go.mod  go.sum
    main.go                          composizione e avvio
    build.sh                         compilazione
    internal/
      auth/auth.go       auth_test.go        lettura credenziali
      client/client.go   client_test.go      chiamate HTTP e traduzione errori
      spec/spec.go       spec_test.go        specifica OpenAPI: scarico, copia locale, estrazione
      preview/preview.go preview_test.go     differenza fra stato attuale e valori richiesti
      tools/registry.go  registry_test.go    registro e gruppi attivi
      tools/stories.go   stories_test.go
      tools/tags.go      tags_test.go
      tools/tasks.go     tasks_test.go
      tools/crm.go       crm_test.go
      testdata/                              risposte registrate di Orchestrator
```

---

### Task 1: Impostazione del progetto Go

**Files:**
- Create: `plugins/wm-skills/mcp/go.mod`, `plugins/wm-skills/mcp/build.sh`, `plugins/wm-skills/mcp/internal/version/version.go`, `plugins/wm-skills/mcp/internal/version/version_test.go`
- Modify: `.gitignore` (se presente)

**Interfaces:**
- Consumes: niente
- Produces: modulo Go compilabile; `version.Version` stringa

- [ ] **Step 1: Verificare la presenza di Go e installarlo se manca**

Go non è installato sulla macchina di sviluppo. Serve solo a chi compila, non a chi usa il binario.

```bash
go version || brew install go
go version   # atteso: go1.24 o superiore
```

- [ ] **Step 2: Creare il modulo**

```bash
mkdir -p plugins/wm-skills/mcp/internal/version
cd plugins/wm-skills/mcp
go mod init github.com/webmappsrl/claude-marketplace/plugins/wm-skills/mcp
go get github.com/modelcontextprotocol/go-sdk@v1.7.0
```

- [ ] **Step 3: Scrivere la prova che fallisce**

`internal/version/version_test.go`:

```go
package version

import "testing"

func TestVersionIsSet(t *testing.T) {
	if Version == "" {
		t.Fatal("Version non deve essere vuota")
	}
}
```

- [ ] **Step 4: Eseguire la prova e verificare che fallisca**

Run: `cd plugins/wm-skills/mcp && go test ./internal/version/`
Expected: FAIL — `undefined: Version`

- [ ] **Step 5: Implementare**

`internal/version/version.go`:

```go
// Package version espone la versione del server, allineata a quella del plugin.
package version

// Version è la versione del plugin wm-skills di cui questo server fa parte.
// Va aggiornata a ogni rilascio, insieme a plugin.json e a wm-plan/SKILL.md.
const Version = "1.3.0"
```

- [ ] **Step 6: Eseguire la prova e verificare che passi**

Run: `cd plugins/wm-skills/mcp && go test ./internal/version/`
Expected: PASS

- [ ] **Step 7: Scrivere lo script di compilazione**

`build.sh`:

```bash
#!/usr/bin/env bash
# Compila il server MCP per darwin/arm64 e lo colloca in plugins/wm-skills/bin/.
set -euo pipefail

cd "$(dirname "$0")"
OUT="../bin/orchestrator-mcp"

mkdir -p ../bin
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o "$OUT" .

echo "Compilato: $(cd .. && pwd)/bin/orchestrator-mcp"
ls -lh "$OUT"
```

```bash
chmod +x plugins/wm-skills/mcp/build.sh
```


---

### Task 2: Lettura delle credenziali

**Files:**
- Create: `plugins/wm-skills/mcp/internal/auth/auth.go`, `plugins/wm-skills/mcp/internal/auth/auth_test.go`

**Interfaces:**
- Consumes: niente
- Produces:
  - `type Credentials struct { Token string; ID int; Name string; Email string }`
  - `func Load(path string) (Credentials, error)`
  - `func DefaultPath() string` — `~/.config/webmapp/orchestrator-auth.json`
  - `var ErrMissing error`, `var ErrMalformed error`, `var ErrNoToken error`

- [ ] **Step 1: Scrivere le prove che falliscono**

`internal/auth/auth_test.go`:

```go
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
```

- [ ] **Step 2: Eseguire e verificare il fallimento**

Run: `cd plugins/wm-skills/mcp && go test ./internal/auth/`
Expected: FAIL — `undefined: Load`

- [ ] **Step 3: Implementare**

`internal/auth/auth.go`:

```go
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
```

- [ ] **Step 4: Eseguire e verificare che passi**

Run: `cd plugins/wm-skills/mcp && go test ./internal/auth/ -v`
Expected: PASS, quattro prove


---

### Task 3: Chiamate HTTP e traduzione degli errori

**Files:**
- Create: `plugins/wm-skills/mcp/internal/client/client.go`, `plugins/wm-skills/mcp/internal/client/client_test.go`

**Interfaces:**
- Consumes: `auth.Credentials`, `auth.Load`, `auth.DefaultPath`
- Produces:
  - `type Client struct { BaseURL string; AuthPath string; HTTP *http.Client }`
  - `func New(baseURL string) *Client` — con stringa vuota usa `https://orchestrator.maphub.it`
  - `func (c *Client) Do(ctx context.Context, method, path string, body any) (json.RawMessage, error)`
  - `type APIError struct { Status int; Message string; Fields map[string][]string }` con `Error() string`

- [ ] **Step 1: Scrivere le prove che falliscono**

`internal/client/client_test.go`:

```go
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
```

Aggiungere in fondo al file di prova la funzione di comodo:

```go
func asAPIError(err error, target **APIError) bool {
	e, ok := err.(*APIError)
	if ok {
		*target = e
	}
	return ok
}
```

- [ ] **Step 2: Eseguire e verificare il fallimento**

Run: `cd plugins/wm-skills/mcp && go test ./internal/client/`
Expected: FAIL — `undefined: Client`

- [ ] **Step 3: Implementare**

`internal/client/client.go`:

```go
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
	"os"
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

// New costruisce il client sull'indirizzo indicato. Con stringa vuota usa la
// produzione: l'indirizzo non arriva mai dall'ambiente, così non può essere
// ereditato per sbaglio da una variabile impostata per altri scopi.
func New(base string) *Client {
	if base == "" {
		base = defaultBaseURL
	}
	return &Client{
		BaseURL:  strings.TrimRight(base, "/"),
		AuthPath: auth.DefaultPath(),
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
		return nil, fmt.Errorf("Orchestrator non raggiungibile su %s: %w — controlla la rete. Se devi procedere a mano, le istruzioni di ripiego sono in shared/orchestrator-fallback.md dentro il plugin", c.BaseURL, err)
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
```

- [ ] **Step 4: Eseguire e verificare che passi**

Run: `cd plugins/wm-skills/mcp && go test ./internal/client/ -v`
Expected: PASS, quattro prove


---

### Task 4: Specifica OpenAPI — scarico, copia locale, estrazione

**Files:**
- Create: `plugins/wm-skills/mcp/internal/spec/spec.go`, `plugins/wm-skills/mcp/internal/spec/spec_test.go`, `plugins/wm-skills/mcp/internal/spec/testdata/mini-openapi.json`

**Interfaces:**
- Consumes: niente
- Produces:
  - `type Spec struct { … }` con `func (s *Spec) EnumValues(schemaName, field string) []string`
  - `func Parse(raw []byte) (*Spec, error)`
  - `func LoadOrFetch(ctx context.Context, baseURL, cachePath string) (*Spec, error)` — prova a scaricare, in caso di errore usa la copia locale
  - `func CachePath() string` — `~/.cache/webmapp/orchestrator-openapi.json`

- [ ] **Step 1: Preparare il campione di specifica**

`internal/spec/testdata/mini-openapi.json`:

```json
{
  "openapi": "3.1.0",
  "info": { "title": "orchestrator", "version": "0.0.1" },
  "components": {
    "schemas": {
      "StoryApiRequest": {
        "type": "object",
        "properties": {
          "name": { "type": "string" },
          "type": { "type": "string", "enum": ["Bug", "Feature", "Help desk", "Scrum"] },
          "status": { "type": "string", "enum": ["new", "progress", "done"] }
        },
        "required": ["name"]
      }
    }
  },
  "paths": {}
}
```

- [ ] **Step 2: Scrivere le prove che falliscono**

`internal/spec/spec_test.go`:

```go
package spec

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func mini(t *testing.T) []byte {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "mini-openapi.json"))
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestEnumValues(t *testing.T) {
	s, err := Parse(mini(t))
	if err != nil {
		t.Fatal(err)
	}
	got := s.EnumValues("StoryApiRequest", "type")
	want := []string{"Bug", "Feature", "Help desk", "Scrum"}
	if len(got) != len(want) {
		t.Fatalf("attesi %d valori, ottenuti %v", len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("valore %d: atteso %q, ottenuto %q", i, want[i], got[i])
		}
	}
}

func TestEnumValuesUnknownField(t *testing.T) {
	s, _ := Parse(mini(t))
	if got := s.EnumValues("StoryApiRequest", "inesistente"); got != nil {
		t.Fatalf("atteso nil per un campo sconosciuto, ottenuto %v", got)
	}
}

func TestLoadOrFetchWritesCache(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/docs/api.json" {
			t.Errorf("percorso sbagliato: %s", r.URL.Path)
		}
		w.Write(mini(t))
	}))
	defer srv.Close()

	cache := filepath.Join(t.TempDir(), "openapi.json")
	s, err := LoadOrFetch(context.Background(), srv.URL, cache)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.EnumValues("StoryApiRequest", "type")) != 4 {
		t.Fatal("specifica scaricata ma non interpretata")
	}
	if _, err := os.Stat(cache); err != nil {
		t.Fatalf("la copia locale non è stata scritta: %v", err)
	}
}

func TestLoadOrFetchFallsBackToCache(t *testing.T) {
	cache := filepath.Join(t.TempDir(), "openapi.json")
	if err := os.WriteFile(cache, mini(t), 0o644); err != nil {
		t.Fatal(err)
	}
	// Indirizzo non raggiungibile: deve usare la copia locale senza fallire.
	s, err := LoadOrFetch(context.Background(), "http://127.0.0.1:1", cache)
	if err != nil {
		t.Fatalf("doveva ripiegare sulla copia locale: %v", err)
	}
	if len(s.EnumValues("StoryApiRequest", "type")) != 4 {
		t.Fatal("copia locale non interpretata")
	}
}
```

- [ ] **Step 3: Eseguire e verificare il fallimento**

Run: `cd plugins/wm-skills/mcp && go test ./internal/spec/`
Expected: FAIL — `undefined: Parse`

- [ ] **Step 4: Implementare**

`internal/spec/spec.go`:

```go
// Package spec legge la specifica OpenAPI pubblica di Orchestrator e ne estrae
// i valori ammessi dei campi. È la parte che tiene i tool allineati all'API
// senza che nessuno debba aggiornarli a mano.
package spec

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Spec struct {
	Components struct {
		Schemas map[string]struct {
			Properties map[string]struct {
				Type string   `json:"type"`
				Enum []string `json:"enum"`
			} `json:"properties"`
			Required []string `json:"required"`
		} `json:"schemas"`
	} `json:"components"`
}

func Parse(raw []byte) (*Spec, error) {
	var s Spec
	if err := json.Unmarshal(raw, &s); err != nil {
		return nil, fmt.Errorf("specifica OpenAPI illeggibile: %w", err)
	}
	return &s, nil
}

// EnumValues restituisce i valori ammessi per un campo, o nil se il campo non
// esiste o non ha un elenco chiuso.
func (s *Spec) EnumValues(schemaName, field string) []string {
	schema, ok := s.Components.Schemas[schemaName]
	if !ok {
		return nil
	}
	prop, ok := schema.Properties[field]
	if !ok || len(prop.Enum) == 0 {
		return nil
	}
	return prop.Enum
}

func CachePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "orchestrator-openapi.json"
	}
	return filepath.Join(home, ".cache", "webmapp", "orchestrator-openapi.json")
}

// LoadOrFetch prova a scaricare la specifica aggiornata; se il servizio non è
// raggiungibile usa l'ultima copia salvata, così il server parte anche offline.
func LoadOrFetch(ctx context.Context, baseURL, cachePath string) (*Spec, error) {
	raw, err := fetch(ctx, baseURL)
	if err == nil {
		if mkErr := os.MkdirAll(filepath.Dir(cachePath), 0o755); mkErr == nil {
			_ = os.WriteFile(cachePath, raw, 0o644)
		}
		return Parse(raw)
	}

	cached, readErr := os.ReadFile(cachePath)
	if readErr != nil {
		return nil, fmt.Errorf("specifica non scaricabile (%v) e nessuna copia locale in %s", err, cachePath)
	}
	return Parse(cached)
}

func fetch(ctx context.Context, baseURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/docs/api.json", nil)
	if err != nil {
		return nil, err
	}
	c := &http.Client{Timeout: 10 * time.Second}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("specifica non disponibile: stato %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}
```

- [ ] **Step 5: Eseguire e verificare che passi**

Run: `cd plugins/wm-skills/mcp && go test ./internal/spec/ -v`
Expected: PASS, quattro prove


---

### Task 4-bis: Elenchi dei valori ammessi dagli enum PHP

**Files:**
- Create: `plugins/wm-skills/mcp/internal/enums/enums.go`, `plugins/wm-skills/mcp/internal/enums/enums_test.go`, `plugins/wm-skills/mcp/internal/enums/testdata/StoryType.php`

**Interfaces:**
- Consumes: niente
- Produces:
  - `type Values struct { StoryType []string; StoryStatus []string }`
  - `func ParsePHPEnum(raw []byte) []string`
  - `func LoadOrFetch(ctx context.Context, cachePath string) (*Values, error)`
  - `func CachePath() string` — `~/.cache/webmapp/orchestrator-enums.json`

**Perché questo compito esiste.** La specifica OpenAPI **non espone** gli elenchi di valori ammessi: verificato sulla specifica vera, `type` e `status` di `StoryApiRequest` risultano senza tipo e senza `enum`, perché Scramble non deduce `Rule::enum()`. Leggerli da lì darebbe una convalida che non convalida nulla. Si leggono quindi dai file sorgente degli enum, che sono pubblici e stabili.

- [ ] **Step 1: Preparare il campione**

`internal/enums/testdata/StoryType.php`, copia fedele del file reale:

```php
<?php

namespace App\Enums;

enum StoryType: string
{
    case Bug = 'Bug';
    case Feature = 'Feature';
    case Helpdesk = 'Help desk';
    case Scrum = 'Scrum';

    public static function values(): array
    {
        return array_column(self::cases(), 'value');
    }
}
```

- [ ] **Step 2: Scrivere le prove che falliscono**

`internal/enums/enums_test.go`:

```go
package enums

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParsePHPEnumReadsValuesNotCaseNames(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "StoryType.php"))
	if err != nil {
		t.Fatal(err)
	}
	got := ParsePHPEnum(raw)
	want := []string{"Bug", "Feature", "Help desk", "Scrum"}

	if len(got) != len(want) {
		t.Fatalf("attesi %v, ottenuti %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("valore %d: atteso %q, ottenuto %q", i, want[i], got[i])
		}
	}
}

func TestParsePHPEnumIgnoresMethodBody(t *testing.T) {
	// Il corpo del metodo values() contiene la parola "cases" e non deve
	// produrre valori spuri.
	raw := []byte("enum X: string {\n case A = 'a';\n public static function values(): array { return array_column(self::cases(), 'value'); }\n}")
	if got := ParsePHPEnum(raw); len(got) != 1 || got[0] != "a" {
		t.Fatalf("atteso un solo valore \"a\", ottenuto %v", got)
	}
}

func TestParsePHPEnumOnGarbage(t *testing.T) {
	if got := ParsePHPEnum([]byte("non è PHP")); got != nil {
		t.Fatalf("atteso nil su contenuto non riconoscibile, ottenuto %v", got)
	}
}
```

- [ ] **Step 3: Eseguire e verificare il fallimento**

Run: `cd plugins/wm-skills/mcp && go test ./internal/enums/`
Expected: FAIL — `undefined: ParsePHPEnum`

- [ ] **Step 4: Implementare**

`internal/enums/enums.go`:

```go
// Package enums legge gli elenchi di valori ammessi dai file sorgente degli
// enum PHP di Orchestrator. Non si usa la specifica OpenAPI perché non li
// espone: per type e status dichiara campi senza tipo e senza elenco.
package enums

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

const (
	storyTypeURL   = "https://raw.githubusercontent.com/webmappsrl/orchestrator/main/app/Enums/StoryType.php"
	storyStatusURL = "https://raw.githubusercontent.com/webmappsrl/orchestrator/main/app/Enums/StoryStatus.php"
)

// Values raccoglie gli elenchi letti.
type Values struct {
	StoryType   []string `json:"story_type"`
	StoryStatus []string `json:"story_status"`
}

// caseLine intercetta le righe "case Nome = 'valore';": si prende il valore
// fra apici, non il nome del case, perché i due possono differire —
// "case Helpdesk = 'Help desk'" ne è l'esempio.
var caseLine = regexp.MustCompile(`(?m)^\s*case\s+\w+\s*=\s*'([^']*)'\s*;`)

func ParsePHPEnum(raw []byte) []string {
	matches := caseLine.FindAllSubmatch(raw, -1)
	if len(matches) == 0 {
		return nil
	}
	values := make([]string, 0, len(matches))
	for _, m := range matches {
		values = append(values, string(m[1]))
	}
	return values
}

func CachePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "orchestrator-enums.json"
	}
	return filepath.Join(home, ".cache", "webmapp", "orchestrator-enums.json")
}

// LoadOrFetch scarica i due enum; se la rete non è disponibile usa l'ultima
// copia salvata. Restituisce errore solo se non ha né l'una né l'altra.
func LoadOrFetch(ctx context.Context, cachePath string) (*Values, error) {
	v := &Values{}
	if raw, err := fetch(ctx, storyTypeURL); err == nil {
		v.StoryType = ParsePHPEnum(raw)
	}
	if raw, err := fetch(ctx, storyStatusURL); err == nil {
		v.StoryStatus = ParsePHPEnum(raw)
	}

	if len(v.StoryType) > 0 && len(v.StoryStatus) > 0 {
		if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err == nil {
			if encoded, err := json.Marshal(v); err == nil {
				_ = os.WriteFile(cachePath, encoded, 0o644)
			}
		}
		return v, nil
	}

	cached, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, fmt.Errorf("elenchi non scaricabili e nessuna copia locale in %s", cachePath)
	}
	var fromCache Values
	if err := json.Unmarshal(cached, &fromCache); err != nil {
		return nil, fmt.Errorf("copia locale degli elenchi illeggibile: %w", err)
	}
	return &fromCache, nil
}

func fetch(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("stato %d su %s", resp.StatusCode, url)
	}
	return io.ReadAll(resp.Body)
}
```

- [ ] **Step 5: Eseguire e verificare che passi**

Run: `cd plugins/wm-skills/mcp && go test ./internal/enums/ -v`
Expected: PASS, tre prove

- [ ] **Step 6: Verificare sul file reale**

```bash
curl -s https://raw.githubusercontent.com/webmappsrl/orchestrator/main/app/Enums/StoryStatus.php | head -20
```

Controllare che le righe dei case abbiano la forma prevista dall'espressione regolare. Se la forma è diversa — per esempio valori fra virgolette doppie — aggiornare l'espressione e aggiungere un campione in `testdata/`.

---

### Task 5: Differenza fra stato attuale e valori richiesti

**Files:**
- Create: `plugins/wm-skills/mcp/internal/preview/preview.go`, `plugins/wm-skills/mcp/internal/preview/preview_test.go`

**Interfaces:**
- Consumes: niente
- Produces:
  - `func Diff(current map[string]any, requested map[string]any) string` — testo leggibile della differenza
  - `func NewResource(requested map[string]any) string` — testo per una creazione, dove non c'è stato precedente

- [ ] **Step 1: Scrivere le prove che falliscono**

`internal/preview/preview_test.go`:

```go
package preview

import (
	"strings"
	"testing"
)

func TestDiffShowsChangedFields(t *testing.T) {
	current := map[string]any{"status": "assigned", "name": "Titolo", "user_id": nil}
	requested := map[string]any{"status": "progress", "user_id": float64(12)}

	got := Diff(current, requested)

	if !strings.Contains(got, "status") || !strings.Contains(got, "assigned") || !strings.Contains(got, "progress") {
		t.Fatalf("la differenza deve mostrare vecchio e nuovo valore:\n%s", got)
	}
	if !strings.Contains(got, "user_id") {
		t.Fatalf("manca un campo modificato:\n%s", got)
	}
}

func TestDiffMarksUnchangedFields(t *testing.T) {
	current := map[string]any{"status": "progress", "name": "Titolo"}
	requested := map[string]any{"status": "progress"}

	got := Diff(current, requested)

	if !strings.Contains(strings.ToLower(got), "nessuna modifica") {
		t.Fatalf("un valore identico non è una modifica:\n%s", got)
	}
}

func TestDiffShowsEmptyPreviousValue(t *testing.T) {
	current := map[string]any{"estimated_hours": nil}
	requested := map[string]any{"estimated_hours": float64(3.5)}

	got := Diff(current, requested)

	if !strings.Contains(got, "(vuoto)") {
		t.Fatalf("un valore assente va reso esplicito:\n%s", got)
	}
}

func TestNewResourceListsAllFields(t *testing.T) {
	got := NewResource(map[string]any{"name": "Nuovo ticket", "type": "Feature"})

	if !strings.Contains(got, "name") || !strings.Contains(got, "Nuovo ticket") {
		t.Fatalf("la creazione deve elencare i campi inviati:\n%s", got)
	}
}
```

- [ ] **Step 2: Eseguire e verificare il fallimento**

Run: `cd plugins/wm-skills/mcp && go test ./internal/preview/`
Expected: FAIL — `undefined: Diff`

- [ ] **Step 3: Implementare**

`internal/preview/preview.go`:

```go
// Package preview costruisce il testo che il dev legge prima di autorizzare una
// scrittura. La differenza è calcolata sui dati veri letti da Orchestrator, non
// ricostruita a memoria.
package preview

import (
	"fmt"
	"sort"
	"strings"
)

// Diff descrive cosa cambierebbe applicando requested allo stato current.
func Diff(current, requested map[string]any) string {
	names := make([]string, 0, len(requested))
	for k := range requested {
		names = append(names, k)
	}
	sort.Strings(names)

	var changed []string
	for _, n := range names {
		before := format(current[n])
		after := format(requested[n])
		if before == after {
			continue
		}
		changed = append(changed, fmt.Sprintf("  %-18s %s  →  %s", n, before, after))
	}

	if len(changed) == 0 {
		return "Nessuna modifica: i valori richiesti coincidono con quelli attuali."
	}

	var b strings.Builder
	b.WriteString("ANTEPRIMA — nulla è stato scritto\n")
	b.WriteString(strings.Join(changed, "\n"))
	b.WriteString("\n\nPer applicare, richiama lo stesso tool con confirm: true.")
	return b.String()
}

// NewResource descrive una creazione, dove non esiste uno stato precedente.
func NewResource(requested map[string]any) string {
	names := make([]string, 0, len(requested))
	for k := range requested {
		names = append(names, k)
	}
	sort.Strings(names)

	var b strings.Builder
	b.WriteString("ANTEPRIMA — nulla è stato creato\n")
	for _, n := range names {
		fmt.Fprintf(&b, "  %-18s %s\n", n, format(requested[n]))
	}
	b.WriteString("\nPer creare, richiama lo stesso tool con confirm: true.")
	return b.String()
}

func format(v any) string {
	switch t := v.(type) {
	case nil:
		return "(vuoto)"
	case string:
		if t == "" {
			return "(vuoto)"
		}
		if len(t) > 120 {
			return fmt.Sprintf("%q… (%d caratteri)", t[:120], len(t))
		}
		return fmt.Sprintf("%q", t)
	case float64:
		if t == float64(int64(t)) {
			return fmt.Sprintf("%d", int64(t))
		}
		return fmt.Sprintf("%g", t)
	default:
		return fmt.Sprintf("%v", v)
	}
}
```

- [ ] **Step 4: Eseguire e verificare che passi**

Run: `cd plugins/wm-skills/mcp && go test ./internal/preview/ -v`
Expected: PASS, quattro prove


---

### Task 6: Registro dei tool e gruppi attivi

**Files:**
- Create: `plugins/wm-skills/mcp/internal/tools/registry.go`, `plugins/wm-skills/mcp/internal/tools/registry_test.go`

**Interfaces:**
- Consumes: `client.Client`, `spec.Spec`
- Produces:
  - `type Deps struct { Client *client.Client; Spec *spec.Spec }`
  - `func ActiveGroups(env string) map[string]bool` — interpreta `ORCHESTRATOR_MCP_GROUPS`; `stories` e `me` sempre attivi
  - `func Register(server *mcp.Server, deps Deps, groups map[string]bool)`

- [ ] **Step 1: Scrivere le prove che falliscono**

`internal/tools/registry_test.go`:

```go
package tools

import "testing"

func TestActiveGroupsAlwaysIncludesCore(t *testing.T) {
	g := ActiveGroups("")
	if !g["stories"] || !g["me"] {
		t.Fatalf("stories e me devono essere sempre attivi: %v", g)
	}
	if g["crm"] {
		t.Fatalf("crm non deve essere attivo senza richiesta: %v", g)
	}
}

func TestActiveGroupsParsesList(t *testing.T) {
	g := ActiveGroups("tags, crm")
	if !g["tags"] || !g["crm"] {
		t.Fatalf("gruppi richiesti non attivati: %v", g)
	}
	if !g["stories"] {
		t.Fatalf("il nucleo resta attivo anche con una lista esplicita: %v", g)
	}
}

func TestActiveGroupsIgnoresUnknown(t *testing.T) {
	g := ActiveGroups("inesistente")
	if g["inesistente"] {
		t.Fatalf("un gruppo sconosciuto non va attivato: %v", g)
	}
}
```

- [ ] **Step 2: Eseguire e verificare il fallimento**

Run: `cd plugins/wm-skills/mcp && go test ./internal/tools/`
Expected: FAIL — `undefined: ActiveGroups`

- [ ] **Step 3: Implementare**

`internal/tools/registry.go`:

```go
// Package tools dichiara i tool esposti dal server. Nomi, descrizioni e
// appartenenza ai gruppi sono scritti a mano; i valori ammessi dei campi
// arrivano dalla specifica OpenAPI.
package tools

import (
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/webmappsrl/claude-marketplace/plugins/wm-skills/mcp/internal/client"
	"github.com/webmappsrl/claude-marketplace/plugins/wm-skills/mcp/internal/spec"
)

// Deps raccoglie ciò di cui i tool hanno bisogno.
type Deps struct {
	Client *client.Client
	Spec   *spec.Spec
	Enums  *enums.Values
}

var known = []string{"stories", "me", "tags", "tasks", "crm"}

// ActiveGroups interpreta la variabile ORCHESTRATOR_MCP_GROUPS. I gruppi
// stories e me sono sempre attivi: chi apre wm-plan sta quasi sempre
// lavorando su un ticket.
func ActiveGroups(env string) map[string]bool {
	active := map[string]bool{"stories": true, "me": true}
	for _, raw := range strings.Split(env, ",") {
		name := strings.TrimSpace(raw)
		for _, k := range known {
			if name == k {
				active[k] = true
			}
		}
	}
	return active
}

// Register aggiunge al server i tool dei gruppi attivi.
func Register(server *mcp.Server, deps Deps, groups map[string]bool) {
	if groups["stories"] {
		registerStories(server, deps)
	}
	if groups["me"] {
		registerMe(server, deps)
	}
	if groups["tags"] {
		registerTags(server, deps)
	}
	if groups["tasks"] {
		registerTasks(server, deps)
	}
	if groups["crm"] {
		registerCRM(server, deps)
	}
}
```

- [ ] **Step 4: Eseguire e verificare che passi**

Le funzioni `registerX` non esistono ancora: crea un file temporaneo `internal/tools/stubs.go` con definizioni vuote, che i compiti successivi sostituiranno.

```go
package tools

import "github.com/modelcontextprotocol/go-sdk/mcp"

func registerStories(*mcp.Server, Deps) {}
func registerMe(*mcp.Server, Deps)      {}
func registerTags(*mcp.Server, Deps)    {}
func registerTasks(*mcp.Server, Deps)   {}
func registerCRM(*mcp.Server, Deps)     {}
```

Run: `cd plugins/wm-skills/mcp && go test ./internal/tools/ -v`
Expected: PASS, tre prove


---

### Task 7: Tool sulle storie

**Files:**
- Create: `plugins/wm-skills/mcp/internal/tools/stories.go`, `plugins/wm-skills/mcp/internal/tools/stories_test.go`
- Modify: `plugins/wm-skills/mcp/internal/tools/stubs.go` (rimuovere `registerStories` e `registerMe`)

**Interfaces:**
- Consumes: `Deps`, `client.Do`, `spec.EnumValues`, `preview.Diff`, `preview.NewResource`
- Produces: tool `get_story`, `create_story`, `update_story`, `me`; funzione interna `validateStoryFields(s *spec.Spec, fields map[string]any) error`

- [ ] **Step 1: Scrivere le prove che falliscono**

`internal/tools/stories_test.go`:

```go
package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/webmappsrl/claude-marketplace/plugins/wm-skills/mcp/internal/spec"
)

func specForTest(t *testing.T) *spec.Spec {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "spec", "testdata", "mini-openapi.json"))
	if err != nil {
		t.Fatal(err)
	}
	s, err := spec.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func enumsForTest() *enums.Values {
	return &enums.Values{
		StoryType:   []string{"Bug", "Feature", "Help desk", "Scrum"},
		StoryStatus: []string{"new", "progress", "done"},
	}
}

func TestValidateRejectsUnknownType(t *testing.T) {
	err := validateStoryFields(enumsForTest(), map[string]any{"type": "Task"})
	if err == nil {
		t.Fatal("un tipo fuori elenco deve essere rifiutato prima di partire")
	}
	if !strings.Contains(err.Error(), "Feature") {
		t.Fatalf("l'errore deve elencare i valori ammessi: %v", err)
	}
}

func TestValidateAcceptsKnownType(t *testing.T) {
	if err := validateStoryFields(enumsForTest(), map[string]any{"type": "Feature"}); err != nil {
		t.Fatalf("un tipo valido non va rifiutato: %v", err)
	}
}

func TestValidateIgnoresFieldsWithoutEnum(t *testing.T) {
	if err := validateStoryFields(enumsForTest(), map[string]any{"name": "qualunque"}); err != nil {
		t.Fatalf("un campo senza elenco chiuso non va convalidato: %v", err)
	}
}

func TestValidateSkipsWhenListsUnavailable(t *testing.T) {
	// Se gli elenchi non si sono potuti leggere, non si inventa una convalida:
	// si lascia decidere a Orchestrator, che risponde comunque con un rifiuto.
	if err := validateStoryFields(nil, map[string]any{"type": "Task"}); err != nil {
		t.Fatalf("senza elenchi la convalida va saltata, non fatta a caso: %v", err)
	}
}
```

- [ ] **Step 2: Eseguire e verificare il fallimento**

Run: `cd plugins/wm-skills/mcp && go test ./internal/tools/ -run TestValidate`
Expected: FAIL — `undefined: validateStoryFields`

- [ ] **Step 3: Implementare**

Rimuovere `registerStories` e `registerMe` da `stubs.go`, poi creare `internal/tools/stories.go`:

```go
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/webmappsrl/claude-marketplace/plugins/wm-skills/mcp/internal/preview"
	"github.com/webmappsrl/claude-marketplace/plugins/wm-skills/mcp/internal/spec"
)

type getStoryInput struct {
	StoryID int `json:"story_id" jsonschema:"identificatore numerico del ticket, la parte dopo oc:"`
}

type storyFieldsInput struct {
	StoryID         int      `json:"story_id,omitempty" jsonschema:"identificatore del ticket da modificare"`
	Name            string   `json:"name,omitempty" jsonschema:"titolo del ticket"`
	Description     string   `json:"description,omitempty" jsonschema:"note di sviluppo, in HTML: il campo è reso da un editor visuale, non interpreta Markdown"`
	CustomerRequest string   `json:"customer_request,omitempty" jsonschema:"richiesta del cliente; scrivendola il cliente riceve una notifica"`
	Type            string   `json:"type,omitempty" jsonschema:"tipo del ticket, fra i valori ammessi dall'API"`
	Status          string   `json:"status,omitempty" jsonschema:"stato del ticket, fra i valori ammessi dall'API"`
	UserID          int      `json:"user_id,omitempty" jsonschema:"utente assegnatario"`
	CreatorID       int      `json:"creator_id,omitempty" jsonschema:"utente creatore"`
	EstimatedHours  float64  `json:"estimated_hours,omitempty" jsonschema:"stima in ore"`
	Tags            []int    `json:"tags,omitempty" jsonschema:"elenco completo degli identificatori di tag: sostituisce quelli presenti, per aggiungerne uno solo usa attach_story_to_tag"`
	Confirm         bool     `json:"confirm,omitempty" jsonschema:"false o assente mostra l'anteprima senza scrivere; true esegue la scrittura"`
}

type textOutput struct {
	Text string `json:"text"`
}

func registerStories(server *mcp.Server, deps Deps) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_story",
		Description: "Legge un ticket Orchestrator dal suo identificatore numerico e restituisce tutti i suoi campi.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in getStoryInput) (*mcp.CallToolResult, textOutput, error) {
		raw, err := deps.Client.Do(ctx, "GET", fmt.Sprintf("/api/stories/%d", in.StoryID), nil)
		if err != nil {
			return nil, textOutput{}, err
		}
		return nil, textOutput{Text: string(raw)}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_story",
		Description: "Modifica un ticket Orchestrator. Senza confirm mostra la differenza rispetto allo stato attuale senza scrivere nulla.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in storyFieldsInput) (*mcp.CallToolResult, textOutput, error) {
		fields := nonEmptyFields(in)
		delete(fields, "story_id")
		delete(fields, "confirm")

		if err := validateStoryFields(deps.Enums, fields); err != nil {
			return nil, textOutput{}, err
		}

		path := fmt.Sprintf("/api/stories/%d", in.StoryID)
		currentRaw, err := deps.Client.Do(ctx, "GET", path, nil)
		if err != nil {
			return nil, textOutput{}, err
		}
		var current map[string]any
		if err := json.Unmarshal(currentRaw, &current); err != nil {
			return nil, textOutput{}, fmt.Errorf("ticket illeggibile: %w", err)
		}

		if !in.Confirm {
			return nil, textOutput{Text: preview.Diff(current, fields)}, nil
		}

		updated, err := deps.Client.Do(ctx, "PATCH", path, fields)
		if err != nil {
			return nil, textOutput{}, err
		}
		var after map[string]any
		_ = json.Unmarshal(updated, &after)
		return nil, textOutput{Text: "Scritto.\n" + preview.Diff(current, fields)}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_story",
		Description: "Crea un ticket Orchestrator. Senza confirm mostra i campi che verrebbero inviati senza creare nulla.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in storyFieldsInput) (*mcp.CallToolResult, textOutput, error) {
		fields := nonEmptyFields(in)
		delete(fields, "story_id")
		delete(fields, "confirm")

		if fields["name"] == nil {
			return nil, textOutput{}, fmt.Errorf("name è obbligatorio per creare un ticket")
		}
		if err := validateStoryFields(deps.Enums, fields); err != nil {
			return nil, textOutput{}, err
		}

		if !in.Confirm {
			return nil, textOutput{Text: preview.NewResource(fields)}, nil
		}

		raw, err := deps.Client.Do(ctx, "POST", "/api/stories", fields)
		if err != nil {
			return nil, textOutput{}, err
		}
		return nil, textOutput{Text: "Creato.\n" + string(raw)}, nil
	})
}

type emptyInput struct{}

func registerMe(server *mcp.Server, deps Deps) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "me",
		Description: "Restituisce l'utente Orchestrator corrispondente alle credenziali in uso.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, textOutput, error) {
		raw, err := deps.Client.Do(ctx, "GET", "/api/me", nil)
		if err != nil {
			return nil, textOutput{}, err
		}
		return nil, textOutput{Text: string(raw)}, nil
	})
}

// validateStoryFields rifiuta i valori fuori dagli elenchi letti dagli enum
// PHP, prima che la chiamata parta. È una rete di sicurezza: la difesa
// principale è l'elenco scritto nello schema del tool.
func validateStoryFields(e *enums.Values, fields map[string]any) error {
	if e == nil {
		return nil
	}
	for _, field := range []string{"type", "status"} {
		value, present := fields[field]
		if !present {
			continue
		}
		text, ok := value.(string)
		if !ok || text == "" {
			continue
		}
		allowed := e.StoryType
		if field == "status" {
			allowed = e.StoryStatus
		}
		if len(allowed) == 0 {
			continue
		}
		found := false
		for _, a := range allowed {
			if a == text {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("%s: il valore %q non è ammesso — valori possibili: %s",
				field, text, strings.Join(allowed, ", "))
		}
	}
	return nil
}

// nonEmptyFields converte l'ingresso in una mappa contenente solo i campi
// effettivamente valorizzati, così una modifica non azzera ciò che non tocca.
func nonEmptyFields(in storyFieldsInput) map[string]any {
	raw, _ := json.Marshal(in)
	fields := map[string]any{}
	_ = json.Unmarshal(raw, &fields)
	return fields
}
```

- [ ] **Step 4: Mettere gli elenchi nello schema del tool**

Questa è la parte che rende un valore inventato **non esprimibile**, invece di fermarlo dopo averlo scritto. La convalida del passo precedente resta come rete, ma la difesa vera è che l'elenco compaia nella definizione del tool, che viene letta prima di chiamarlo.

Verificare come l'SDK permette di fornire uno schema di ingresso costruito a runtime, invece di dedurlo dalla struttura Go:

```bash
cd plugins/wm-skills/mcp
go doc github.com/modelcontextprotocol/go-sdk/mcp.Tool
go doc github.com/modelcontextprotocol/go-sdk/mcp.AddTool
```

- **Se `mcp.Tool` accetta uno schema di ingresso esplicito:** costruirlo per `create_story` e `update_story` partendo da quello dedotto, e sostituire la definizione di `type` e `status` con un elenco chiuso costruito da `deps.Enums`. Quando gli elenchi non sono disponibili, lasciare lo schema dedotto.
- **Se l'SDK non lo permette:** non inventare un aggiramento. Riferire al dev che con questa libreria la garanzia sullo schema non è ottenibile, e che resta la sola convalida interna — cioè un errore immediato e chiaro invece di un rifiuto del backend. È un esito accettabile, ma va dichiarato, non scoperto dopo.

In entrambi i casi la descrizione dei due campi elenca i valori ammessi, così sono visibili anche a chi legge soltanto il tool.

- [ ] **Step 5: Eseguire e verificare che passi**

Run: `cd plugins/wm-skills/mcp && go test ./internal/tools/ -v && go build ./...`
Expected: PASS e compilazione senza errori

---

### Task 8: Tool sui tag

**Files:**
- Create: `plugins/wm-skills/mcp/internal/tools/tags.go`, `plugins/wm-skills/mcp/internal/tools/tags_test.go`
- Modify: `plugins/wm-skills/mcp/internal/tools/stubs.go` (rimuovere `registerTags`)

**Interfaces:**
- Consumes: `Deps`, `preview.NewResource`
- Produces: tool `list_tags`, `get_tag`, `create_tag`, `update_tag`, `attach_story_to_tag`, `detach_story_from_tag`

- [ ] **Step 1: Scrivere la prova che fallisce**

`internal/tools/tags_test.go`:

```go
package tools

import (
	"strings"
	"testing"
)

func TestAttachPathIsBuiltCorrectly(t *testing.T) {
	got := attachPath(42, 8503)
	if got != "/api/tags/42/stories/8503" {
		t.Fatalf("percorso sbagliato: %s", got)
	}
}

func TestTagDescriptionMentionsMarkdown(t *testing.T) {
	// La descrizione del tag è resa da un editor Markdown, a differenza della
	// description di una story che è HTML: il tool deve dirlo, altrimenti chi
	// lo usa applica la convenzione sbagliata.
	if !strings.Contains(strings.ToLower(tagDescriptionHint), "markdown") {
		t.Fatalf("la descrizione del campo deve citare Markdown: %s", tagDescriptionHint)
	}
}
```

- [ ] **Step 2: Eseguire e verificare il fallimento**

Run: `cd plugins/wm-skills/mcp && go test ./internal/tools/ -run TestAttach`
Expected: FAIL — `undefined: attachPath`

- [ ] **Step 3: Implementare**

Rimuovere `registerTags` da `stubs.go`, poi creare `internal/tools/tags.go`:

```go
package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/webmappsrl/claude-marketplace/plugins/wm-skills/mcp/internal/preview"
)

const tagDescriptionHint = "descrizione del tag, in Markdown: a differenza della description di una story, questo campo è reso da un editor Markdown"

type listTagsInput struct {
	Name string `json:"name,omitempty" jsonschema:"filtro sul nome del tag"`
}

type getTagInput struct {
	TagID int `json:"tag_id" jsonschema:"identificatore del tag"`
}

type tagFieldsInput struct {
	TagID       int    `json:"tag_id,omitempty" jsonschema:"identificatore del tag da modificare"`
	Name        string `json:"name,omitempty" jsonschema:"nome del tag"`
	Description string `json:"description,omitempty" jsonschema:"descrizione del tag, in Markdown"`
	Confirm     bool   `json:"confirm,omitempty" jsonschema:"false o assente mostra l'anteprima senza scrivere; true esegue la scrittura"`
}

type tagStoryInput struct {
	TagID   int  `json:"tag_id" jsonschema:"identificatore del tag"`
	StoryID int  `json:"story_id" jsonschema:"identificatore del ticket"`
	Confirm bool `json:"confirm,omitempty" jsonschema:"false o assente mostra cosa verrebbe fatto; true esegue"`
}

func attachPath(tagID, storyID int) string {
	return fmt.Sprintf("/api/tags/%d/stories/%d", tagID, storyID)
}

func registerTags(server *mcp.Server, deps Deps) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_tags",
		Description: "Elenca i tag Orchestrator, con filtro facoltativo sul nome.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in listTagsInput) (*mcp.CallToolResult, textOutput, error) {
		path := "/api/tags"
		if in.Name != "" {
			path += "?name=" + in.Name
		}
		raw, err := deps.Client.Do(ctx, "GET", path, nil)
		if err != nil {
			return nil, textOutput{}, err
		}
		return nil, textOutput{Text: string(raw)}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_tag",
		Description: "Legge un tag con i ticket a esso associati.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in getTagInput) (*mcp.CallToolResult, textOutput, error) {
		raw, err := deps.Client.Do(ctx, "GET", fmt.Sprintf("/api/tags/%d", in.TagID), nil)
		if err != nil {
			return nil, textOutput{}, err
		}
		return nil, textOutput{Text: string(raw)}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_tag",
		Description: "Crea un tag Orchestrator. Senza confirm mostra i campi senza creare nulla. " + tagDescriptionHint + ".",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in tagFieldsInput) (*mcp.CallToolResult, textOutput, error) {
		fields := map[string]any{"name": in.Name}
		if in.Description != "" {
			fields["description"] = in.Description
		}
		if in.Name == "" {
			return nil, textOutput{}, fmt.Errorf("name è obbligatorio per creare un tag")
		}
		if !in.Confirm {
			return nil, textOutput{Text: preview.NewResource(fields)}, nil
		}
		raw, err := deps.Client.Do(ctx, "POST", "/api/tags", fields)
		if err != nil {
			return nil, textOutput{}, err
		}
		return nil, textOutput{Text: "Creato.\n" + string(raw)}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_tag",
		Description: "Modifica un tag Orchestrator. Senza confirm mostra la differenza senza scrivere. " + tagDescriptionHint + ".",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in tagFieldsInput) (*mcp.CallToolResult, textOutput, error) {
		fields := map[string]any{}
		if in.Name != "" {
			fields["name"] = in.Name
		}
		if in.Description != "" {
			fields["description"] = in.Description
		}
		path := fmt.Sprintf("/api/tags/%d", in.TagID)

		currentRaw, err := deps.Client.Do(ctx, "GET", path, nil)
		if err != nil {
			return nil, textOutput{}, err
		}
		current := decodeMap(currentRaw)

		if !in.Confirm {
			return nil, textOutput{Text: preview.Diff(current, fields)}, nil
		}
		raw, err := deps.Client.Do(ctx, "PATCH", path, fields)
		if err != nil {
			return nil, textOutput{}, err
		}
		return nil, textOutput{Text: "Scritto.\n" + string(raw)}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "attach_story_to_tag",
		Description: "Associa un ticket a un tag senza toccare gli altri tag del ticket. " +
			"Da preferire sempre alla modifica del campo tags, che sostituisce l'elenco completo e può cancellare tag già presenti.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in tagStoryInput) (*mcp.CallToolResult, textOutput, error) {
		if !in.Confirm {
			return nil, textOutput{Text: fmt.Sprintf("ANTEPRIMA — nulla è stato scritto\n  il ticket %d verrebbe associato al tag %d\n\nPer applicare, richiama con confirm: true.", in.StoryID, in.TagID)}, nil
		}
		raw, err := deps.Client.Do(ctx, "POST", attachPath(in.TagID, in.StoryID), nil)
		if err != nil {
			return nil, textOutput{}, err
		}
		return nil, textOutput{Text: "Associato.\n" + string(raw)}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "detach_story_from_tag",
		Description: "Toglie l'associazione fra un ticket e un tag, lasciando intatti gli altri tag del ticket.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in tagStoryInput) (*mcp.CallToolResult, textOutput, error) {
		if !in.Confirm {
			return nil, textOutput{Text: fmt.Sprintf("ANTEPRIMA — nulla è stato scritto\n  il ticket %d verrebbe tolto dal tag %d\n\nPer applicare, richiama con confirm: true.", in.StoryID, in.TagID)}, nil
		}
		raw, err := deps.Client.Do(ctx, "DELETE", attachPath(in.TagID, in.StoryID), nil)
		if err != nil {
			return nil, textOutput{}, err
		}
		return nil, textOutput{Text: "Rimosso.\n" + string(raw)}, nil
	})
}
```

Aggiungere in `internal/tools/registry.go` la funzione di comodo condivisa:

```go
func decodeMap(raw []byte) map[string]any {
	m := map[string]any{}
	_ = json.Unmarshal(raw, &m)
	return m
}
```

(con l'importazione di `encoding/json`)

- [ ] **Step 4: Eseguire e verificare che passi**

Run: `cd plugins/wm-skills/mcp && go test ./internal/tools/ -v && go build ./...`
Expected: PASS e compilazione senza errori


---

### Task 9: Tool su task e area commerciale (lettura e scrittura)

**Files:**
- Create: `plugins/wm-skills/mcp/internal/tools/tasks.go`, `plugins/wm-skills/mcp/internal/tools/crm.go`, `plugins/wm-skills/mcp/internal/tools/crm_writes.go`, `plugins/wm-skills/mcp/internal/tools/crm_test.go`
- Modify: `plugins/wm-skills/mcp/internal/tools/registry.go` (aggiunta di `fieldsOf`), `plugins/wm-skills/mcp/internal/tools/stubs.go` (eliminarlo: non restano stub)

**Interfaces:**
- Consumes: `Deps`, `preview.Diff`, `preview.NewResource`, `decodeMap`
- Produces: tool `list_tasks`, `get_task`, `create_task`, `update_task`, `list_customers`, `get_customer`, `create_customer`, `update_customer`, `list_quotes`, `get_quote`, `create_quote`, `update_quote`, `delete_quote`, `attach_product_to_quote`, `detach_product_from_quote`, `create_quote_pdf_link`, `list_products`; funzioni `isIrreversible`, `irreversibleNotice`, `fieldsOf`, `quoteProductPath`

- [ ] **Step 1: Scrivere la prova che fallisce**

`internal/tools/crm_test.go`:

```go
package tools

import (
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestIrreversibleToolsAreMarked(t *testing.T) {
	// Eliminare un preventivo e generare un collegamento pubblico non si
	// annullano. Il server non può impedirlo all'agente — nessun parametro
	// derivabile dai dati è una barriera — ma può marcarli, perché la
	// descrizione e l'anteprima lo dicano a chi autorizza.
	if !isIrreversible("delete_quote") || !isIrreversible("create_quote_pdf_link") {
		t.Fatal("le operazioni non annullabili vanno marcate come tali")
	}
	if isIrreversible("update_customer") {
		t.Fatal("una modifica correggibile non va marcata come non annullabile")
	}
}

func TestIrreversiblePreviewNamesTheTarget(t *testing.T) {
	// Chi autorizza deve vedere che cosa sta colpendo, non solo un numero.
	text := irreversibleNotice("delete_quote", 418, "Preventivo Acme 2026")
	if !strings.Contains(text, "Preventivo Acme 2026") || !strings.Contains(text, "418") {
		t.Fatalf("l'avviso deve nominare il preventivo: %q", text)
	}
}

func TestRegisterWithoutCRMDoesNotPanic(t *testing.T) {
	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "v0"}, nil)
	Register(server, Deps{}, map[string]bool{"stories": true, "me": true})
}
```

- [ ] **Step 2: Eseguire e verificare il fallimento**

Run: `cd plugins/wm-skills/mcp && go test ./internal/tools/ -run TestCRM`
Expected: FAIL — `undefined: crmToolNames`

- [ ] **Step 3: Implementare**

Eliminare `stubs.go`. Creare `internal/tools/tasks.go`:

```go
package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/webmappsrl/claude-marketplace/plugins/wm-skills/mcp/internal/preview"
)

type listTasksInput struct {
	Sort string `json:"sort,omitempty" jsonschema:"ordinamento, per esempio created_at oppure -created_at per l'ordine inverso"`
}

type getTaskInput struct {
	TaskID int `json:"task_id" jsonschema:"identificatore del task"`
}

type taskFieldsInput struct {
	TaskID  int    `json:"task_id,omitempty" jsonschema:"identificatore del task da modificare"`
	QuoteID int    `json:"quote_id,omitempty" jsonschema:"preventivo a cui il task appartiene, obbligatorio in creazione"`
	Title   string `json:"title,omitempty" jsonschema:"titolo del task"`
	Status  string `json:"status,omitempty" jsonschema:"stato del task; modificabile solo da chi lo ha creato"`
	Notes   string `json:"notes,omitempty" jsonschema:"note sul task"`
	Confirm bool   `json:"confirm,omitempty" jsonschema:"false o assente mostra l'anteprima senza scrivere; true esegue la scrittura"`
}

func registerTasks(server *mcp.Server, deps Deps) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_tasks",
		Description: "Elenca i task dell'utente autenticato: quelli sui preventivi che possiede o che ha creato.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in listTasksInput) (*mcp.CallToolResult, textOutput, error) {
		path := "/api/tasks"
		if in.Sort != "" {
			path += "?sort=" + in.Sort
		}
		raw, err := deps.Client.Do(ctx, "GET", path, nil)
		if err != nil {
			return nil, textOutput{}, err
		}
		return nil, textOutput{Text: string(raw)}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_task",
		Description: "Legge il dettaglio di un task.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in getTaskInput) (*mcp.CallToolResult, textOutput, error) {
		raw, err := deps.Client.Do(ctx, "GET", fmt.Sprintf("/api/tasks/%d", in.TaskID), nil)
		if err != nil {
			return nil, textOutput{}, err
		}
		return nil, textOutput{Text: string(raw)}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_task",
		Description: "Crea un task su un preventivo esistente. Senza confirm mostra i campi senza creare nulla.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in taskFieldsInput) (*mcp.CallToolResult, textOutput, error) {
		fields := map[string]any{}
		if in.QuoteID != 0 {
			fields["quote_id"] = in.QuoteID
		}
		if in.Title != "" {
			fields["title"] = in.Title
		}
		if in.Notes != "" {
			fields["notes"] = in.Notes
		}
		if !in.Confirm {
			return nil, textOutput{Text: preview.NewResource(fields)}, nil
		}
		raw, err := deps.Client.Do(ctx, "POST", "/api/tasks", fields)
		if err != nil {
			return nil, textOutput{}, err
		}
		return nil, textOutput{Text: "Creato.\n" + string(raw)}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "update_task",
		Description: "Modifica stato o note di un task. Senza confirm mostra la differenza senza scrivere. " +
			"Lo stato è modificabile solo da chi ha creato il task: se non sei tu, l'intera richiesta viene rifiutata e nemmeno le note vengono salvate.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in taskFieldsInput) (*mcp.CallToolResult, textOutput, error) {
		fields := map[string]any{}
		if in.Status != "" {
			fields["status"] = in.Status
		}
		if in.Notes != "" {
			fields["notes"] = in.Notes
		}
		path := fmt.Sprintf("/api/tasks/%d", in.TaskID)

		currentRaw, err := deps.Client.Do(ctx, "GET", path, nil)
		if err != nil {
			return nil, textOutput{}, err
		}
		if !in.Confirm {
			return nil, textOutput{Text: preview.Diff(decodeMap(currentRaw), fields)}, nil
		}
		raw, err := deps.Client.Do(ctx, "PATCH", path, fields)
		if err != nil {
			return nil, textOutput{}, err
		}
		return nil, textOutput{Text: "Scritto.\n" + string(raw)}, nil
	})
}
```

Creare `internal/tools/crm.go`:

```go
package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Il gruppo crm espone lettura e scrittura. Le scritture seguono la regola
// generale dell'anteprima con confirm; due di esse hanno effetti che non si
// annullano e richiedono in più una frase esatta ricavata dai dati veri.
func crmToolNames() []string {
	return []string{
		"list_customers", "get_customer", "create_customer", "update_customer",
		"list_quotes", "get_quote", "create_quote", "update_quote", "delete_quote",
		"attach_product_to_quote", "detach_product_from_quote",
		"create_quote_pdf_link",
		"list_products",
	}
}

// isIrreversible marca i tool il cui effetto non si annulla: l'eliminazione di
// un preventivo, e il collegamento pubblico, che resta valido fino a 90 giorni
// e non può essere revocato prima della scadenza.
//
// Il server non può impedire queste operazioni all'agente: qualunque parametro
// ricavabile dai dati è soddisfacibile da chi quei dati li ha appena letti.
// L'unica decisione umana è l'autorizzazione richiesta dal programma, e questi
// tool non vanno mai messi fra quelli approvati in automatico. Ciò che il
// server può fare è rendere quella decisione informata.
func isIrreversible(tool string) bool {
	return tool == "delete_quote" || tool == "create_quote_pdf_link"
}

// irreversibleNotice compone l'avviso mostrato in anteprima, nominando il
// preventivo colpito invece del solo identificatore.
func irreversibleNotice(tool string, quoteID int, title string) string {
	return fmt.Sprintf("⚠️  %s non si può annullare — preventivo %d: %q", tool, quoteID, title)
}

type listCustomersInput struct {
	Search string `json:"search,omitempty" jsonschema:"ricerca sul nome del cliente"`
	Status string `json:"status,omitempty" jsonschema:"filtro sullo stato del cliente"`
}

type getCustomerInput struct {
	CustomerID int `json:"customer_id" jsonschema:"identificatore del cliente"`
}

type listQuotesInput struct {
	CustomerID int    `json:"customer_id,omitempty" jsonschema:"filtra i preventivi di un cliente"`
	Status     string `json:"status,omitempty" jsonschema:"filtro sullo stato del preventivo"`
}

type getQuoteInput struct {
	QuoteID int    `json:"quote_id" jsonschema:"identificatore del preventivo"`
	Include string `json:"include,omitempty" jsonschema:"relazioni da espandere nella risposta"`
}

func registerCRM(server *mcp.Server, deps Deps) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_customers",
		Description: "Elenca i clienti, con filtri facoltativi su nome e stato.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in listCustomersInput) (*mcp.CallToolResult, textOutput, error) {
		path := "/api/customers"
		sep := "?"
		if in.Search != "" {
			path += sep + "search=" + in.Search
			sep = "&"
		}
		if in.Status != "" {
			path += sep + "status=" + in.Status
		}
		raw, err := deps.Client.Do(ctx, "GET", path, nil)
		if err != nil {
			return nil, textOutput{}, err
		}
		return nil, textOutput{Text: string(raw)}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_customer",
		Description: "Legge un cliente.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in getCustomerInput) (*mcp.CallToolResult, textOutput, error) {
		raw, err := deps.Client.Do(ctx, "GET", fmt.Sprintf("/api/customers/%d", in.CustomerID), nil)
		if err != nil {
			return nil, textOutput{}, err
		}
		return nil, textOutput{Text: string(raw)}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_quotes",
		Description: "Elenca i preventivi, con filtri facoltativi su cliente e stato.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in listQuotesInput) (*mcp.CallToolResult, textOutput, error) {
		path := "/api/quotes"
		sep := "?"
		if in.CustomerID != 0 {
			path += fmt.Sprintf("%scustomer_id=%d", sep, in.CustomerID)
			sep = "&"
		}
		if in.Status != "" {
			path += sep + "status=" + in.Status
		}
		raw, err := deps.Client.Do(ctx, "GET", path, nil)
		if err != nil {
			return nil, textOutput{}, err
		}
		return nil, textOutput{Text: string(raw)}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_quote",
		Description: "Legge un preventivo, con la possibilità di espandere le relazioni.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in getQuoteInput) (*mcp.CallToolResult, textOutput, error) {
		path := fmt.Sprintf("/api/quotes/%d", in.QuoteID)
		if in.Include != "" {
			path += "?include=" + in.Include
		}
		raw, err := deps.Client.Do(ctx, "GET", path, nil)
		if err != nil {
			return nil, textOutput{}, err
		}
		return nil, textOutput{Text: string(raw)}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_products",
		Description: "Elenca i prodotti a catalogo.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, textOutput, error) {
		raw, err := deps.Client.Do(ctx, "GET", "/api/products", nil)
		if err != nil {
			return nil, textOutput{}, err
		}
		return nil, textOutput{Text: string(raw)}, nil
	})

	registerCRMWrites(server, deps)
}
```

Creare `internal/tools/crm_writes.go` con le scritture:

```go
package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/webmappsrl/claude-marketplace/plugins/wm-skills/mcp/internal/preview"
)

type customerFieldsInput struct {
	CustomerID    int      `json:"customer_id,omitempty" jsonschema:"identificatore del cliente da modificare"`
	Name          string   `json:"name,omitempty" jsonschema:"nome del referente"`
	CompanyName   string   `json:"company_name,omitempty" jsonschema:"ragione sociale"`
	Vat           string   `json:"vat,omitempty" jsonschema:"partita IVA"`
	Address       string   `json:"address,omitempty" jsonschema:"indirizzo"`
	Phone         string   `json:"phone,omitempty" jsonschema:"telefono"`
	Status        string   `json:"status,omitempty" jsonschema:"stato del cliente"`
	Notes         string   `json:"notes,omitempty" jsonschema:"note interne"`
	ContactEmails []string `json:"contact_emails,omitempty" jsonschema:"elenco completo delle email di contatto: sostituisce quelle presenti"`
	Confirm       bool     `json:"confirm,omitempty" jsonschema:"false o assente mostra l'anteprima senza scrivere; true esegue la scrittura"`
}

type quoteFieldsInput struct {
	QuoteID        int     `json:"quote_id,omitempty" jsonschema:"identificatore del preventivo da modificare"`
	Title          string  `json:"title,omitempty" jsonschema:"titolo del preventivo, obbligatorio in creazione"`
	CustomerID     int     `json:"customer_id,omitempty" jsonschema:"cliente intestatario, obbligatorio in creazione"`
	Status         string  `json:"status,omitempty" jsonschema:"stato del preventivo; a preventivo chiuso le modifiche vengono rifiutate"`
	Priority       int     `json:"priority,omitempty" jsonschema:"priorità"`
	Discount       float64 `json:"discount,omitempty" jsonschema:"sconto"`
	GoogleDriveURL string  `json:"google_drive_url,omitempty" jsonschema:"collegamento alla cartella Drive"`
	Notes          string  `json:"notes,omitempty" jsonschema:"note interne"`
	Confirm        bool    `json:"confirm,omitempty" jsonschema:"false o assente mostra l'anteprima senza scrivere; true esegue la scrittura"`
}

type deleteQuoteInput struct {
	QuoteID       int    `json:"quote_id" jsonschema:"identificatore del preventivo da eliminare"`
	Confirm       bool   `json:"confirm,omitempty" jsonschema:"false o assente mostra cosa verrebbe eliminato; true esegue"`
	QuoteTitle string `json:"quote_title,omitempty" jsonschema:"titolo del preventivo, a scopo informativo: compare nella richiesta di autorizzazione così chi approva vede quale preventivo sta eliminando invece del solo numero"`
}

type quoteProductInput struct {
	QuoteID   int  `json:"quote_id" jsonschema:"identificatore del preventivo"`
	ProductID int  `json:"product_id" jsonschema:"identificatore del prodotto"`
	Quantity  int  `json:"quantity,omitempty" jsonschema:"quantità, obbligatoria quando si associa"`
	Recurring bool `json:"recurring,omitempty" jsonschema:"true se si tratta di un prodotto ricorrente"`
	Confirm   bool `json:"confirm,omitempty" jsonschema:"false o assente mostra cosa verrebbe fatto; true esegue"`
}

type pdfLinkInput struct {
	QuoteID       int    `json:"quote_id" jsonschema:"identificatore del preventivo"`
	Lang          string `json:"lang,omitempty" jsonschema:"lingua del documento"`
	ExpiresInDays int    `json:"expires_in_days,omitempty" jsonschema:"durata del collegamento in giorni, massimo 90"`
	Confirm       bool   `json:"confirm,omitempty" jsonschema:"false o assente mostra cosa verrebbe generato; true esegue"`
	QuoteTitle string `json:"quote_title,omitempty" jsonschema:"titolo del preventivo, a scopo informativo: compare nella richiesta di autorizzazione così chi approva vede per quale preventivo sta generando un collegamento pubblico"`
}

func registerCRMWrites(server *mcp.Server, deps Deps) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_customer",
		Description: "Crea un cliente. Senza confirm mostra i campi senza creare nulla.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in customerFieldsInput) (*mcp.CallToolResult, textOutput, error) {
		fields := fieldsOf(in, "customer_id", "confirm")
		if !in.Confirm {
			return nil, textOutput{Text: preview.NewResource(fields)}, nil
		}
		raw, err := deps.Client.Do(ctx, "POST", "/api/customers", fields)
		if err != nil {
			return nil, textOutput{}, err
		}
		return nil, textOutput{Text: "Creato.\n" + string(raw)}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "update_customer",
		Description: "Modifica un cliente. Senza confirm mostra la differenza senza scrivere. " +
			"Attenzione: contact_emails sostituisce l'elenco completo delle email di contatto.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in customerFieldsInput) (*mcp.CallToolResult, textOutput, error) {
		fields := fieldsOf(in, "customer_id", "confirm")
		path := fmt.Sprintf("/api/customers/%d", in.CustomerID)

		currentRaw, err := deps.Client.Do(ctx, "GET", path, nil)
		if err != nil {
			return nil, textOutput{}, err
		}
		if !in.Confirm {
			return nil, textOutput{Text: preview.Diff(decodeMap(currentRaw), fields)}, nil
		}
		raw, err := deps.Client.Do(ctx, "PATCH", path, fields)
		if err != nil {
			return nil, textOutput{}, err
		}
		return nil, textOutput{Text: "Scritto.\n" + string(raw)}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_quote",
		Description: "Crea un preventivo. Richiede titolo e cliente. Senza confirm mostra i campi senza creare nulla.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in quoteFieldsInput) (*mcp.CallToolResult, textOutput, error) {
		if in.Title == "" || in.CustomerID == 0 {
			return nil, textOutput{}, fmt.Errorf("title e customer_id sono obbligatori per creare un preventivo")
		}
		fields := fieldsOf(in, "quote_id", "confirm")
		if !in.Confirm {
			return nil, textOutput{Text: preview.NewResource(fields)}, nil
		}
		raw, err := deps.Client.Do(ctx, "POST", "/api/quotes", fields)
		if err != nil {
			return nil, textOutput{}, err
		}
		return nil, textOutput{Text: "Creato.\n" + string(raw)}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "update_quote",
		Description: "Modifica un preventivo. Senza confirm mostra la differenza senza scrivere. " +
			"A preventivo chiuso Orchestrator rifiuta ogni modifica.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in quoteFieldsInput) (*mcp.CallToolResult, textOutput, error) {
		fields := fieldsOf(in, "quote_id", "confirm")
		path := fmt.Sprintf("/api/quotes/%d", in.QuoteID)

		currentRaw, err := deps.Client.Do(ctx, "GET", path, nil)
		if err != nil {
			return nil, textOutput{}, err
		}
		if !in.Confirm {
			return nil, textOutput{Text: preview.Diff(decodeMap(currentRaw), fields)}, nil
		}
		raw, err := deps.Client.Do(ctx, "PATCH", path, fields)
		if err != nil {
			return nil, textOutput{}, err
		}
		return nil, textOutput{Text: "Scritto.\n" + string(raw)}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "delete_quote",
		Description: "Elimina un preventivo. Operazione non annullabile: non va mai autorizzata in automatico. " +
			"Riporta in quote_title il titolo letto dall'anteprima, così chi approva vede quale preventivo sta eliminando.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in deleteQuoteInput) (*mcp.CallToolResult, textOutput, error) {
		path := fmt.Sprintf("/api/quotes/%d", in.QuoteID)
		currentRaw, err := deps.Client.Do(ctx, "GET", path, nil)
		if err != nil {
			return nil, textOutput{}, err
		}
		current := decodeMap(currentRaw)
		title, _ := current["title"].(string)

		if !in.Confirm {
			return nil, textOutput{Text: fmt.Sprintf(
				"ANTEPRIMA — nulla è stato eliminato\n%s\n\nPer eseguire richiama con confirm: true e quote_title: %q.",
				irreversibleNotice("delete_quote", in.QuoteID, title), title)}, nil
		}
		raw, err := deps.Client.Do(ctx, "DELETE", path, nil)
		if err != nil {
			return nil, textOutput{}, err
		}
		return nil, textOutput{Text: "Eliminato.\n" + string(raw)}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "attach_product_to_quote",
		Description: "Associa un prodotto a un preventivo con una quantità. Usa recurring: true per i prodotti ricorrenti.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in quoteProductInput) (*mcp.CallToolResult, textOutput, error) {
		if in.Quantity <= 0 {
			return nil, textOutput{}, fmt.Errorf("quantity è obbligatoria e deve essere maggiore di zero")
		}
		path := quoteProductPath(in.QuoteID, in.ProductID, in.Recurring)
		if !in.Confirm {
			return nil, textOutput{Text: fmt.Sprintf(
				"ANTEPRIMA — nulla è stato scritto\n  al preventivo %d verrebbe associato il prodotto %d in quantità %d\n\nPer applicare, richiama con confirm: true.",
				in.QuoteID, in.ProductID, in.Quantity)}, nil
		}
		raw, err := deps.Client.Do(ctx, "POST", path, map[string]any{"quantity": in.Quantity})
		if err != nil {
			return nil, textOutput{}, err
		}
		return nil, textOutput{Text: "Associato.\n" + string(raw)}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "detach_product_from_quote",
		Description: "Toglie un prodotto da un preventivo. Usa recurring: true per i prodotti ricorrenti.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in quoteProductInput) (*mcp.CallToolResult, textOutput, error) {
		path := quoteProductPath(in.QuoteID, in.ProductID, in.Recurring)
		if !in.Confirm {
			return nil, textOutput{Text: fmt.Sprintf(
				"ANTEPRIMA — nulla è stato scritto\n  dal preventivo %d verrebbe tolto il prodotto %d\n\nPer applicare, richiama con confirm: true.",
				in.QuoteID, in.ProductID)}, nil
		}
		raw, err := deps.Client.Do(ctx, "DELETE", path, nil)
		if err != nil {
			return nil, textOutput{}, err
		}
		return nil, textOutput{Text: "Rimosso.\n" + string(raw)}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "create_quote_pdf_link",
		Description: "Genera un collegamento pubblico al PDF del preventivo, pensato per essere inviato al cliente. " +
			"Il collegamento non richiede autenticazione, dura fino a 90 giorni e non può essere revocato prima della scadenza: " +
			"non va mai autorizzato in automatico. Riporta in quote_title il titolo letto dall'anteprima.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in pdfLinkInput) (*mcp.CallToolResult, textOutput, error) {
		path := fmt.Sprintf("/api/quotes/%d", in.QuoteID)
		currentRaw, err := deps.Client.Do(ctx, "GET", path, nil)
		if err != nil {
			return nil, textOutput{}, err
		}
		current := decodeMap(currentRaw)
		title, _ := current["title"].(string)

		days := in.ExpiresInDays
		if days == 0 {
			days = 30
		}
		if days > 90 {
			return nil, textOutput{}, fmt.Errorf("expires_in_days non può superare 90")
		}

		if !in.Confirm {
			return nil, textOutput{Text: fmt.Sprintf(
				"ANTEPRIMA — nessun collegamento è stato generato\n  preventivo %d: %q\n  durata: %d giorni\n\n"+
					"Il collegamento sarà pubblico e non revocabile prima della scadenza.\nPer generarlo richiama con confirm: true e quote_title: %q.",
				in.QuoteID, title, days, title)}, nil
		}

		body := map[string]any{"expires_in_days": days}
		if in.Lang != "" {
			body["lang"] = in.Lang
		}
		raw, err := deps.Client.Do(ctx, "POST", fmt.Sprintf("/api/quotes/%d/pdf-link", in.QuoteID), body)
		if err != nil {
			return nil, textOutput{}, err
		}
		return nil, textOutput{Text: "Collegamento generato.\n" + string(raw)}, nil
	})
}

func quoteProductPath(quoteID, productID int, recurring bool) string {
	if recurring {
		return fmt.Sprintf("/api/quotes/%d/recurring-products/%d", quoteID, productID)
	}
	return fmt.Sprintf("/api/quotes/%d/products/%d", quoteID, productID)
}
```

Aggiungere in `internal/tools/registry.go` la funzione condivisa che estrae i campi valorizzati da una qualsiasi struttura di ingresso, escludendo quelli indicati:

```go
func fieldsOf(in any, exclude ...string) map[string]any {
	raw, _ := json.Marshal(in)
	fields := map[string]any{}
	_ = json.Unmarshal(raw, &fields)
	for _, name := range exclude {
		delete(fields, name)
	}
	return fields
}
```

- [ ] **Step 4: Eseguire e verificare che passi**

Run: `cd plugins/wm-skills/mcp && go test ./... -v && go build ./...`
Expected: PASS su tutti i pacchetti


---

### Task 10: Avvio del server e dichiarazione nel plugin

**Files:**
- Create: `plugins/wm-skills/mcp/main.go`, `plugins/wm-skills/.mcp.json`
- Modify: `plugins/wm-skills/.claude-plugin/plugin.json` (versione)

**Interfaces:**
- Consumes: `client.New`, `spec.LoadOrFetch`, `spec.CachePath`, `tools.ActiveGroups`, `tools.Register`, `version.Version`
- Produces: binario eseguibile

- [ ] **Step 1: Scrivere il punto di ingresso**

`main.go`:

```go
// Comando orchestrator-mcp: espone l'API di Orchestrator come tool MCP.
// Comunica sui canali standard del processo, senza aprire porte in ascolto.
package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/webmappsrl/claude-marketplace/plugins/wm-skills/mcp/internal/client"
	"github.com/webmappsrl/claude-marketplace/plugins/wm-skills/mcp/internal/enums"
	"github.com/webmappsrl/claude-marketplace/plugins/wm-skills/mcp/internal/spec"
	"github.com/webmappsrl/claude-marketplace/plugins/wm-skills/mcp/internal/tools"
	"github.com/webmappsrl/claude-marketplace/plugins/wm-skills/mcp/internal/version"
)

func main() {
	// I messaggi diagnostici vanno sul canale degli errori: quello di uscita
	// è riservato al dialogo con Claude Code.
	log.SetOutput(os.Stderr)
	log.SetPrefix("orchestrator-mcp: ")

	// L'indirizzo arriva solo dalla riga di comando: senza argomento è la
	// produzione. Il plugin distribuito non passa nulla; il server di sviluppo
	// dichiarato nel repo claude-marketplace passa l'istanza locale.
	baseURL := flag.String("base-url", "", "indirizzo di Orchestrator; vuoto significa produzione")
	flag.Parse()

	ctx := context.Background()
	c := client.New(*baseURL)

	s, err := spec.LoadOrFetch(ctx, c.BaseURL, spec.CachePath())
	if err != nil {
		log.Fatalf("impossibile ottenere la specifica dell'API: %v", err)
	}

	// Gli elenchi di valori ammessi non stanno nella specifica: si leggono
	// dagli enum PHP. Se non si riesce, il server parte lo stesso ma lo dice.
	allowed, err := enums.LoadOrFetch(ctx, enums.CachePath())
	if err != nil {
		log.Printf("attenzione: elenchi dei valori ammessi non disponibili (%v): type e status non saranno verificati prima dell'invio", err)
	}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "orchestrator",
		Version: version.Version,
	}, nil)

	groups := tools.ActiveGroups(os.Getenv("ORCHESTRATOR_MCP_GROUPS"))
	tools.Register(server, tools.Deps{Client: c, Spec: s, Enums: allowed}, groups)

	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		log.Fatalf("il server si è fermato: %v", err)
	}
}
```

- [ ] **Step 2: Compilare**

Run: `cd plugins/wm-skills/mcp && ./build.sh`
Expected: `Compilato: …/plugins/wm-skills/bin/orchestrator-mcp`, dimensione intorno a 10 MB

- [ ] **Step 3: Verificare che il programma risponda al protocollo**

```bash
cd plugins/wm-skills
printf '%s\n' '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"prova","version":"1"}}}' \
  | ./bin/orchestrator-mcp
```

Expected: una risposta JSON contenente `"serverInfo"` con nome `orchestrator`. Il comando resta in attesa: interrompilo con Ctrl-C.

- [ ] **Step 4: Dichiarare il server nel plugin**

`plugins/wm-skills/.mcp.json` — il plugin distribuito, **senza argomento di indirizzo**, quindi fisso sulla produzione:

```json
{
  "mcpServers": {
    "orchestrator": {
      "command": "${CLAUDE_PLUGIN_ROOT}/bin/orchestrator-mcp",
      "env": {
        "ORCHESTRATOR_MCP_GROUPS": "tags"
      }
    }
  }
}
```

I gruppi predefiniti includono `tags` perché `wm-tag` e `wm-plan` lo usano. Chi lavora sull'area commerciale aggiunge `tasks,crm` nella propria configurazione.

Creare inoltre `.mcp.json` nella radice di **questo repo** — non nel plugin, quindi mai distribuito — con il server di sviluppo puntato all'istanza locale:

```json
{
  "mcpServers": {
    "orchestrator-dev": {
      "command": "./plugins/wm-skills/bin/orchestrator-mcp",
      "args": ["--base-url", "http://localhost:8000"],
      "env": {
        "ORCHESTRATOR_MCP_GROUPS": "tags,tasks,crm"
      }
    }
  }
}
```

Sostituire la porta con quella dell'istanza locale. I due server hanno nomi diversi, quindi anche i tool li hanno: nella richiesta di autorizzazione si legge `orchestrator-dev · update_story` oppure `orchestrator · update_story`, e i due ambienti non sono confondibili.


---

### Task 11: CANCELLO — collaudo manuale prima di toccare le skill

**Files:** nessuno

**Interfaces:**
- Consumes: il binario compilato
- Produces: approvazione esplicita del dev

> **Questo compito non produce codice. Nessun compito successivo può iniziare senza l'approvazione esplicita del dev.**

- [ ] **Step 1: Avviare Orchestrator in locale**

Nel repo `orchestrator`, avviare l'istanza locale con Docker. Annotare l'indirizzo su cui risponde.

- [ ] **Step 2: Verificare che il server di sviluppo punti all'istanza locale**

Controllare che l'argomento `--base-url` nel `.mcp.json` di questo repo corrisponda all'indirizzo dell'istanza appena avviata, poi ricaricare i plugin. Il `.mcp.json` del plugin **non va toccato**: resta fisso sulla produzione.

Da qui in avanti usare sempre i tool di `orchestrator-dev`. Se in una richiesta di autorizzazione compare `orchestrator` senza il suffisso, si sta colpendo la produzione: rifiutare.

- [ ] **Step 3: Prova di lettura**

Chiedere a Claude di leggere un ticket esistente sull'istanza locale. Verificare che i campi corrispondano a quelli visibili nell'interfaccia.

- [ ] **Step 4: Prova di anteprima**

Chiedere una modifica senza conferma. Verificare che: non venga scritto nulla, la differenza mostri solo i campi che cambiano, i valori precedenti siano quelli veri.

- [ ] **Step 5: Prova di scrittura**

Confermare la modifica. Verificare sull'interfaccia che il valore sia cambiato.

- [ ] **Step 6: Prova di rifiuto**

Chiedere una modifica con un tipo inesistente, per esempio `Task`. Verificare che l'errore arrivi **prima** della chiamata e che elenchi i valori ammessi.

- [ ] **Step 7: Prova di creazione**

Creare un ticket di prova con anteprima e conferma.

- [ ] **Step 8: Registrare le risposte vere per le prove future**

È il momento giusto: si hanno dati veri sotto mano. Salvare le risposte dell'istanza locale come campioni, così le prove verificano il formato reale e non quello immaginato.

```bash
cd plugins/wm-skills/mcp
mkdir -p internal/client/testdata
TOKEN=$(jq -r '.token' ~/.config/webmapp/orchestrator-auth.json)
LOCAL_URL="<indirizzo dell'istanza locale>"

curl -s -H "Authorization: Bearer $TOKEN" -H "Accept: application/json" \
  "$LOCAL_URL/api/stories/<id di prova>" > internal/client/testdata/story.json

curl -s -H "Authorization: Bearer $TOKEN" -H "Accept: application/json" \
  "$LOCAL_URL/api/tags" > internal/client/testdata/tags.json

# Risposta di rifiuto: tipo inesistente
curl -s -X POST -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -H "Accept: application/json" -d '{"name":"prova","type":"Task"}' \
  "$LOCAL_URL/api/stories" > internal/client/testdata/rejected.json
```

Controllare i campioni e **rimuovere eventuali dati personali** prima di registrarli nel repo.

- [ ] **Step 9: Aggiungere la prova che usa i campioni**

In `internal/client/client_test.go`:

```go
func TestTranslateOnRecordedRejection(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "rejected.json"))
	if err != nil {
		t.Skip("campione non ancora registrato")
	}
	err = translate(422, raw)
	var apiErr *APIError
	if !asAPIError(err, &apiErr) {
		t.Fatalf("atteso *APIError, ottenuto %T", err)
	}
	if len(apiErr.Fields) == 0 {
		t.Fatalf("il rifiuto reale deve produrre l'elenco dei campi: %s", apiErr.Error())
	}
}
```

Run: `cd plugins/wm-skills/mcp && go test ./internal/client/ -v`
Expected: PASS


- [ ] **Step 10: Nessuna configurazione da ripristinare**

Il server di sviluppo vive nel `.mcp.json` di questo repo e non viene distribuito; quello del plugin non è mai stato modificato. Non c'è niente da riportare indietro — che è il motivo per cui i due sono separati.

- [ ] **Step 11: Approvazione**

Il dev dichiara esplicitamente se il comportamento è quello atteso. In caso contrario, si torna ai compiti precedenti prima di proseguire.

---

### Task 11-bis: Rendere evidente, in fase di autorizzazione, cosa esce verso il cliente

**Files:**
- Modify: `plugins/wm-skills/mcp/internal/tools/stories.go`, `plugins/wm-skills/mcp/internal/tools/stories_test.go`

**Interfaces:**
- Consumes: `storyFieldsInput`
- Produces: `func requiresSeenPreview(fields map[string]any) bool`

`customer_request` fa scattare `addResponse()` lato Orchestrator, che **notifica il cliente**: è l'unica scrittura delle skill con un effetto verso l'esterno non annullabile.

**Nessun meccanismo interno al server può impedirlo all'agente**: qualunque parametro derivabile dai dati è auto-soddisfacibile da chi quei dati li ha appena letti. L'unico punto in cui esiste una decisione umana è la richiesta di autorizzazione del programma. Questo compito quindi non costruisce una barriera — sarebbe finta — ma fa in modo che, quando quella richiesta arriva, sia evidente che sta per partire una comunicazione al cliente e se ne veda il testo.

- [ ] **Step 1: Scrivere le prove che falliscono**

Aggiungere in `internal/tools/stories_test.go`:

```go
func TestPreviewWarnsAboutCustomerNotification(t *testing.T) {
	text := previewNotice(map[string]any{"customer_request": "Buongiorno, abbiamo risolto."})
	if !strings.Contains(strings.ToLower(text), "cliente") {
		t.Fatalf("l'anteprima deve dire che parte una notifica al cliente: %q", text)
	}
}

func TestNoNoticeForInternalFields(t *testing.T) {
	if previewNotice(map[string]any{"status": "progress"}) != "" {
		t.Fatal("una scrittura interna non deve generare avvisi")
	}
}
```

- [ ] **Step 2: Eseguire e verificare il fallimento**

Run: `cd plugins/wm-skills/mcp && go test ./internal/tools/ -run SeenPreview`
Expected: FAIL — `undefined: requiresSeenPreview`

- [ ] **Step 3: Implementare**

Aggiungere a `internal/tools/stories.go`:

```go
// previewNotice restituisce l'avviso da anteporre all'anteprima quando la
// scrittura ha effetti fuori dal team. Serve a rendere informata la decisione
// umana, non a impedire la chiamata: nessun controllo interno al server
// potrebbe farlo.
func previewNotice(fields map[string]any) string {
	if text, ok := fields["customer_request"].(string); ok && text != "" {
		return "⚠️  Scrivendo customer_request parte una notifica al cliente. Una email inviata non si annulla.\n"
	}
	return ""
}
```

Nel gestore di `update_story`, anteporre l'avviso all'anteprima:

```go
		if !in.Confirm {
			return nil, textOutput{Text: previewNotice(fields) + preview.Diff(current, fields)}, nil
		}
```

Aggiungere infine alla descrizione del tool `update_story`:

```
Scrivere customer_request invia una notifica al cliente: non va mai autorizzato in automatico.
```

- [ ] **Step 4: Eseguire e verificare che passi**

Run: `cd plugins/wm-skills/mcp && go test ./internal/tools/ -v && go build ./...`
Expected: PASS


---

### Task 12: Riscrittura di wm-plan sui tool

**Files:**
- Create: `plugins/wm-skills/shared/orchestrator-fallback.md`
- Modify: `plugins/wm-skills/skills/wm-plan/SKILL.md`

**Interfaces:**
- Consumes: i tool registrati nei compiti 7 e 8
- Produces: skill che non contiene più comandi `curl` verso Orchestrator

- [ ] **Step 1: Salvare le istruzioni attuali come ripiego**

Prima di toccare la skill, spostare l'intera sezione `## Orchestrator API` di `wm-plan/SKILL.md` — così com'è, senza riscriverla — in `plugins/wm-skills/shared/orchestrator-fallback.md`, premettendo:

```markdown
# Ripiego: Orchestrator senza il server MCP

Queste istruzioni valgono **solo** quando i tool del server `orchestrator` non rispondono.
Nel funzionamento normale non vanno lette né seguite: usa i tool.
```

Il file vive a livello di plugin, quindi è raggiungibile allo stesso percorso da tutte e tre le skill, e non viene caricato in nessuna sessione finché qualcuno non lo apre.

- [ ] **Step 2: Sostituire la sezione sull'API**

Sostituire l'intera sezione `## Orchestrator API` (dalla riga `## Orchestrator API` fino a `## Fase: ticket` esclusa) con:

```markdown
## Orchestrator

Le operazioni su Orchestrator si fanno con i tool del server `orchestrator`, distribuito con questo plugin. Non costruire chiamate HTTP a mano.

| Operazione | Tool |
|---|---|
| Leggere un ticket | `get_story` |
| Creare un ticket | `create_story` |
| Modificare un ticket | `update_story` |
| Utente corrente | `me` |
| Tag: elenco, lettura, creazione, modifica | `list_tags`, `get_tag`, `create_tag`, `update_tag` |
| Associare o togliere un ticket da un tag | `attach_story_to_tag`, `detach_story_from_tag` |

**Regola sulle scritture.** I tool di scrittura accettano `confirm`. Chiamali **sempre prima senza `confirm`**: restituiscono la differenza rispetto allo stato attuale senza scrivere nulla. Mostra quella differenza al dev, attendi un'approvazione esplicita, e solo allora richiama lo stesso tool con `confirm: true`. Non costruire tu la tabella dell'anteprima: quella del tool è calcolata sui dati veri.

**Tipi e stati.** Non scrivere valori a memoria: il tool rifiuta i valori fuori elenco e ti dice quali sono ammessi.

**Formato dei campi.** `description` e `customer_request` di un ticket sono resi da un editor visuale: vanno scritti in HTML, non in Markdown. La `description` di un tag è invece in Markdown.

**Associazione a un tag.** Usa `attach_story_to_tag`, mai il campo `tags` di `update_story`: quel campo sostituisce l'elenco completo e cancella i tag già presenti.

**Credenziali.** Il server legge `~/.config/webmapp/orchestrator-auth.json`. Se un tool segnala credenziali assenti o scadute, guida il dev a rifare l'accesso; il server non lo fa da sé e non chiede mai la password.

**Se i tool non sono disponibili** (server non avviato o in errore), segnalalo al dev, poi leggi `${CLAUDE_PLUGIN_ROOT}/shared/orchestrator-fallback.md` e segui quelle istruzioni per questa sessione. Non ricostruire le chiamate a memoria: quel file è l'unica forma ammessa di ripiego.
```

- [ ] **Step 2: Sostituire le chiamate nelle fasi**

Per ciascuno dei punti in cui la skill esegue un comando `curl` — `ticket: caso-a`, `ticket: progress`, `ticket: caso-b`, `caso-a-split-execution`, `estimation: scrittura su Orchestrator`, `update-context: orchestrator`, e la modalità tag — sostituire il blocco di comando con l'indicazione del tool corrispondente e dei suoi parametri, mantenendo invariata la prosa che descrive quando e perché farlo.

Esempio, in `ticket: progress`:

```markdown
Se l'utente risponde sì, chiama `update_story` con `story_id`, `status: "progress"` e `user_id` preso da `me`, prima senza `confirm` per mostrare la differenza, poi con `confirm: true` dopo l'approvazione.

Se il tool restituisce un errore, avvisa l'utente con "⚠️ Impossibile aggiornare lo status del ticket — procedo comunque con il workflow." e continua.
```

- [ ] **Step 3: Verificare che non restino comandi HTTP**

Run: `grep -n "curl" plugins/wm-skills/skills/wm-plan/SKILL.md`
Expected: nessun risultato riferito a Orchestrator (restano ammesse le chiamate a GitHub per la specifica o per il controllo della versione)

- [ ] **Step 4: Convalidare il plugin**

Run: `claude plugin validate .`
Expected: `Validation passed`


---

### Task 13: Riscrittura di wm-tag e wm-review-ticket

**Files:**
- Modify: `plugins/wm-skills/skills/wm-tag/SKILL.md`, `plugins/wm-skills/skills/wm-review-ticket/SKILL.md`

**Interfaces:**
- Consumes: i tool registrati nei compiti 7 e 8
- Produces: due skill senza comandi HTTP verso Orchestrator

- [ ] **Step 1: Sostituire le chiamate in wm-tag**

Sostituire i due comandi `curl` di creazione tag con `create_tag` (prima senza `confirm` per l'anteprima, poi con `confirm: true`), e sostituire il rimando in prosa alle regole di autenticazione di `wm-plan` con:

```markdown
Le operazioni su Orchestrator si fanno con i tool del server `orchestrator`: vedi `wm-skills:wm-plan` → `## Orchestrator`. La regola dell'anteprima prima della scrittura vale identica qui.
```

Sostituire inoltre l'associazione dei ticket al tag con `attach_story_to_tag`, invece della modifica del campo `tags`.

- [ ] **Step 2: Sostituire le chiamate in wm-review-ticket**

Sostituire il comando di lettura del ticket con `get_story`, quello di aggiornamento con `update_story`, e rimuovere il blocco di accesso, che il server non richiede più.

- [ ] **Step 3: Verificare**

Run: `grep -n "curl\|orchestrator-auth" plugins/wm-skills/skills/wm-tag/SKILL.md plugins/wm-skills/skills/wm-review-ticket/SKILL.md`
Expected: nessun riferimento residuo a chiamate HTTP verso Orchestrator

- [ ] **Step 4: Convalidare il plugin**

Run: `claude plugin validate .`
Expected: `Validation passed`


---

### Task 14: Rilascio e documentazione

**Files:**
- Modify: `plugins/wm-skills/.claude-plugin/plugin.json`, `plugins/wm-skills/skills/wm-plan/SKILL.md`, `plugins/wm-skills/mcp/internal/version/version.go`, `CLAUDE.md`, `docs/wm-plan-diagram/index.html`
- Create: `docs/features/mcp-orchestrator/notes.md`

**Interfaces:**
- Consumes: tutto il lavoro precedente
- Produces: versione 1.3.0 rilasciata

- [ ] **Step 1: Allineare la versione nei tre punti**

Portare a `1.3.0`: il campo `version` di `plugin.json`, la riga `**Versione installata:** v1.3.0` e la riga successiva in `wm-plan/SKILL.md`, la costante `Version` in `internal/version/version.go`.

- [ ] **Step 2: Ricompilare il binario con la versione aggiornata**

Run: `cd plugins/wm-skills/mcp && ./build.sh`

- [ ] **Step 3: Aggiornare la lista di controllo del rilascio in CLAUDE.md**

Nella sezione `## Versioning del plugin wm-skills`, inserire dopo il punto 2:

```markdown
2-bis. Aggiorna la costante `Version` in `plugins/wm-skills/mcp/internal/version/version.go` e ricompila il binario con `plugins/wm-skills/mcp/build.sh` — il binario compilato è versionato nel repo e va aggiornato a ogni rilascio.
```

- [ ] **Step 4: Aggiungere la voce nelle funzionalità disponibili**

In `## Feature disponibili`, in cima alla tabella:

```markdown
| Server MCP per Orchestrator | — (nessun ticket) | `plugins/wm-skills/mcp/`, `plugins/wm-skills/.mcp.json`, le tre `SKILL.md` | Server MCP in Go che espone l'API di Orchestrator come tool tipizzati: valori ammessi letti dalla specifica OpenAPI, anteprima obbligatoria prima delle scritture tramite il parametro `confirm`, gruppi di tool attivabili. Le skill non costruiscono più chiamate HTTP a mano |
```

- [ ] **Step 5: Aggiungere le decisioni architetturali**

In cima a `## Decisioni architetturali`:

```markdown
### Server MCP per Orchestrator
- **Go compilato invece di un ambiente da installare**: il binario viaggia nel plugin, quindi nessuno deve installare nulla; Go produce un eseguibile di ~10 MB senza dipendenze, contro i ~50-100 MB di un JavaScript compilato che si porta dentro il proprio ambiente di esecuzione
- **Comunicazione sui canali standard del processo, nessuna porta**: esclude per costruzione ogni conflitto con i container Docker del team
- **Tool dichiarati a mano, valori ammessi letti dalla specifica**: la specifica OpenAPI è dedotta dal codice e ha descrizioni di qualità disomogenea, inadatte a diventare descrizioni di tool; si separa ciò che invecchia (campi e valori) da ciò che va curato (nomi e descrizioni)
- **Parametro `confirm` invece di un codice di conferma a scadenza**: valutato un meccanismo a due chiamate con codice, scartato perché richiedeva uno stato da conservare e un tempo di scadenza arbitrario; il parametro con valore predefinito prudente ottiene lo stesso risultato senza stato
- **Gruppo commerciale in lettura e scrittura**: tutte le scritture seguono la regola dell'anteprima con `confirm`. Per quelle non annullabili — eliminazione di un preventivo, collegamento PDF pubblico, `customer_request` che notifica il cliente — il server non prova a costruire una barriera: **nessun parametro ricavabile dai dati può fermare l'agente che li ha appena letti**, e una conferma auto-soddisfacibile darebbe una falsa sicurezza. Il server si limita a rendere informata la decisione umana, portando il titolo del preventivo e l'avviso dentro la richiesta di autorizzazione; la regola operativa, scritta dove si configura il plugin, è che quei tool non vanno mai fra quelli approvati in automatico
- **`attach_story_to_tag` al posto della modifica del campo `tags`**: quel campo sostituisce l'elenco completo e può cancellare tag già presenti — difetto presente nelle skill prima di questo lavoro
- **Server locale invece che ospitato su Orchestrator**: rimandata l'ipotesi di esporre l'MCP dal Laravel, che risolverebbe la proprietà del codice ma richiede lavoro su un servizio in produzione; se il server locale dimostra il suo valore, la migrazione è un secondo tempo e i tool restano gli stessi
```

- [ ] **Step 6: Scrivere le note**

Creare `docs/features/mcp-orchestrator/notes.md` con le sezioni Deviazioni dal piano, Bug trovati, Decisioni, Follow-up. Registrare almeno: l'esito del collaudo del compito 11, e il fatto che `GET /stories` resta da aprire come ticket separato su `orchestrator`.

- [ ] **Step 7: Rigenerare il diagramma**

Aggiornare `docs/wm-plan-diagram/index.html` (contenuto, non struttura) e ripubblicare l'Artifact sullo stesso indirizzo, come prescritto in `CLAUDE.md` → `## Diagramma di flusso wm-plan`.

- [ ] **Step 8: Convalidare**

Run: `claude plugin validate .`
Expected: `Validation passed`

Run: `cd plugins/wm-skills/mcp && go test ./... && go build ./...`
Expected: tutte le prove passano

- [ ] **Step 9: Consegnare a wm-plan**

Il lavoro tecnico finisce qui. Revisione del diff, approvazione, ramo, commit, richiesta di incorporazione ed etichetta `v1.3.0` sono governati da `wm-skills:wm-plan` → `execution: review-gate` e dalla lista di controllo del rilascio in `CLAUDE.md`. Nessun comando git va eseguito da questo piano.
