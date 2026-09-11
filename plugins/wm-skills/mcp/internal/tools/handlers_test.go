package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/webmappsrl/claude-marketplace/plugins/wm-skills/mcp/internal/client"
)

// recordingServer è il server finto di Orchestrator usato per verificare quali
// richieste HTTP un tool emette davvero, invece di fidarsi della sola lettura
// del codice: conta separatamente le richieste di scrittura (POST/PATCH/
// DELETE) da quelle di lettura (GET), e registra l'ultima richiesta di
// scrittura per ispezionarne il corpo.
type recordingServer struct {
	mu            sync.Mutex
	reads         int
	writes        int
	lastWriteBody map[string]any
	lastWriteMeth string
	lastWritePath string
}

func newRecordingServer(t *testing.T) (*recordingServer, *httptest.Server) {
	t.Helper()
	rs := &recordingServer{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rs.mu.Lock()
		defer rs.mu.Unlock()
		switch r.Method {
		case http.MethodGet:
			rs.reads++
		case http.MethodPost, http.MethodPatch, http.MethodDelete:
			rs.writes++
			rs.lastWriteMeth = r.Method
			rs.lastWritePath = r.URL.Path
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			rs.lastWriteBody = body
		}
		w.Header().Set("Content-Type", "application/json")
		// Risposta generica valida per qualunque GET/POST/PATCH/DELETE usata
		// nei test: un oggetto con pochi campi noti basta, i tool non
		// richiedono altro per calcolare anteprima/diff in questi percorsi.
		_, _ = w.Write([]byte(`{"id":1,"name":"Titolo attuale","status":"todo"}`))
	}))
	t.Cleanup(srv.Close)
	return rs, srv
}

// testSession avvia un server MCP reale coi gruppi indicati, collegato a un
// client MCP in-memory: permette di invocare gli handler dei tool come farebbe
// un agente reale, invece di chiamare le funzioni Go interne direttamente.
func testSession(t *testing.T, groups map[string]bool, orchestratorURL string) *mcp.ClientSession {
	t.Helper()

	authPath := filepath.Join(t.TempDir(), "auth.json")
	if err := os.WriteFile(authPath, []byte(`{"token":"t0k3n","id":1}`), 0o600); err != nil {
		t.Fatal(err)
	}
	c := client.NewWithAuthPath(orchestratorURL, authPath)

	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "v0"}, nil)
	Register(server, Deps{Client: c}, groups)

	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	ctx := context.Background()
	if _, err := server.Connect(ctx, serverTransport, nil); err != nil {
		t.Fatalf("connessione server fallita: %v", err)
	}

	mcpClient := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "v0"}, nil)
	session, err := mcpClient.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("connessione client fallita: %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })
	return session
}

func callTool(t *testing.T, session *mcp.ClientSession, name string, args map[string]any) *mcp.CallToolResult {
	t.Helper()
	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("CallTool(%s) fallita a livello di protocollo: %v", name, err)
	}
	return res
}

func TestUpdateStoryWithoutConfirmEmitsNoWriteRequest(t *testing.T) {
	rs, srv := newRecordingServer(t)
	session := testSession(t, map[string]bool{"stories": true, "me": true}, srv.URL)

	res := callTool(t, session, "update_story", map[string]any{
		"story_id": 8503,
		"name":     "Nuovo titolo",
	})
	if res.IsError {
		t.Fatalf("update_story senza confirm non deve fallire: %+v", res)
	}

	rs.mu.Lock()
	defer rs.mu.Unlock()
	if rs.writes != 0 {
		t.Fatalf("update_story senza confirm non deve emettere nessuna richiesta di scrittura, ricevute %d", rs.writes)
	}
	if rs.reads == 0 {
		t.Fatalf("update_story deve comunque leggere lo stato attuale per calcolare l'anteprima")
	}
}

func TestUpdateStoryWithConfirmEmitsPatchWithExpectedBody(t *testing.T) {
	rs, srv := newRecordingServer(t)
	session := testSession(t, map[string]bool{"stories": true, "me": true}, srv.URL)

	res := callTool(t, session, "update_story", map[string]any{
		"story_id": 8503,
		"name":     "Nuovo titolo",
		"confirm":  true,
	})
	if res.IsError {
		t.Fatalf("update_story con confirm non deve fallire: %+v", res)
	}

	rs.mu.Lock()
	defer rs.mu.Unlock()
	if rs.writes != 1 {
		t.Fatalf("atteso esattamente 1 richiesta di scrittura, ricevute %d", rs.writes)
	}
	if rs.lastWriteMeth != http.MethodPatch {
		t.Fatalf("atteso PATCH, ricevuto %s", rs.lastWriteMeth)
	}
	if rs.lastWritePath != "/api/stories/8503" {
		t.Fatalf("percorso sbagliato: %s", rs.lastWritePath)
	}
	if rs.lastWriteBody["name"] != "Nuovo titolo" {
		t.Fatalf("corpo della PATCH sbagliato: %+v", rs.lastWriteBody)
	}
}

func TestAttachStoryToTagWithoutConfirmEmitsNoWriteRequest(t *testing.T) {
	rs, srv := newRecordingServer(t)
	session := testSession(t, map[string]bool{"stories": true, "me": true, "tags": true}, srv.URL)

	res := callTool(t, session, "attach_story_to_tag", map[string]any{
		"tag_id":   42,
		"story_id": 8503,
	})
	if res.IsError {
		t.Fatalf("attach_story_to_tag senza confirm non deve fallire: %+v", res)
	}

	rs.mu.Lock()
	defer rs.mu.Unlock()
	if rs.writes != 0 {
		t.Fatalf("attach_story_to_tag senza confirm non deve emettere nessuna richiesta di scrittura, ricevute %d", rs.writes)
	}
}

func TestAttachStoryToTagWithConfirmEmitsPost(t *testing.T) {
	rs, srv := newRecordingServer(t)
	session := testSession(t, map[string]bool{"stories": true, "me": true, "tags": true}, srv.URL)

	res := callTool(t, session, "attach_story_to_tag", map[string]any{
		"tag_id":   42,
		"story_id": 8503,
		"confirm":  true,
	})
	if res.IsError {
		t.Fatalf("attach_story_to_tag con confirm non deve fallire: %+v", res)
	}

	rs.mu.Lock()
	defer rs.mu.Unlock()
	if rs.writes != 1 {
		t.Fatalf("atteso esattamente 1 richiesta di scrittura, ricevute %d", rs.writes)
	}
	if rs.lastWriteMeth != http.MethodPost {
		t.Fatalf("atteso POST, ricevuto %s", rs.lastWriteMeth)
	}
	if rs.lastWritePath != "/api/tags/42/stories/8503" {
		t.Fatalf("percorso sbagliato: %s", rs.lastWritePath)
	}
}
