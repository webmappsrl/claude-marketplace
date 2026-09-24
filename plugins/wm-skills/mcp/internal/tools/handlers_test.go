package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
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
	return newRecordingServerWithBody(t, `{"id":1,"name":"Titolo attuale","status":"todo"}`)
}

// newRecordingServerWithBody risponde a ogni richiesta con body: serve ai test
// che devono leggere una description precisa dal ticket.
func newRecordingServerWithBody(t *testing.T, body string) (*recordingServer, *httptest.Server) {
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
			var b map[string]any
			_ = json.NewDecoder(r.Body).Decode(&b)
			rs.lastWriteBody = b
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return rs, srv
}

// resultText unisce il testo di tutti i blocchi del risultato.
func resultText(res *mcp.CallToolResult) string {
	var parts []string
	for _, c := range res.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			parts = append(parts, tc.Text)
		}
	}
	return strings.Join(parts, "\n")
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

const storyWithCycle = `{"id":8543,"status":"progress","description":"<h2>Primo ciclo</h2><ul><li>perde la quota</li></ul>"}`

func TestUpdateStoryAppendSendsComposedDescription(t *testing.T) {
	rs, srv := newRecordingServerWithBody(t, storyWithCycle)
	session := testSession(t, map[string]bool{"stories": true, "me": true}, srv.URL)

	res := callTool(t, session, "update_story", map[string]any{
		"story_id":    8543,
		"status":      "todo",
		"prepend":     "<h2>Secondo ciclo</h2>",
		"annotations": []map[string]any{{"after": "perde la quota", "text": "<strong>✅ Risolto</strong>"}},
		"confirm":     true,
	})
	if res.IsError {
		t.Fatalf("update_story in aggiunta non deve fallire: %s", resultText(res))
	}

	rs.mu.Lock()
	defer rs.mu.Unlock()
	if rs.writes != 1 || rs.lastWriteMeth != http.MethodPatch {
		t.Fatalf("attesa una sola PATCH, ricevute %d scritture (%s)", rs.writes, rs.lastWriteMeth)
	}
	want := "<h2>Secondo ciclo</h2><h2>Primo ciclo</h2><ul><li>perde la quota <strong>✅ Risolto</strong></li></ul>"
	if rs.lastWriteBody["description"] != want || rs.lastWriteBody["status"] != "todo" {
		t.Fatalf("corpo della PATCH sbagliato: %+v", rs.lastWriteBody)
	}
	if _, leaked := rs.lastWriteBody["prepend"]; leaked {
		t.Fatalf("prepend non deve arrivare a Orchestrator: %+v", rs.lastWriteBody)
	}
}

func TestUpdateStoryAppendPreviewShowsEverythingReadable(t *testing.T) {
	rs, srv := newRecordingServerWithBody(t, storyWithCycle)
	session := testSession(t, map[string]bool{"stories": true, "me": true}, srv.URL)

	res := callTool(t, session, "update_story", map[string]any{
		"story_id":    8543,
		"prepend":     "<h2>Secondo ciclo</h2>",
		"annotations": []map[string]any{{"after": "perde la quota", "text": "<strong>✅ Risolto</strong>"}},
	})
	text := resultText(res)
	if res.IsError {
		t.Fatalf("l'anteprima non deve fallire: %s", text)
	}
	for _, want := range []string{"SECONDO CICLO", "PRIMO CICLO", "- perde la quota ✅ Risolto", "annotazione 1", "prima:", "dopo:"} {
		if !strings.Contains(text, want) {
			t.Fatalf("all'anteprima manca %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "<h2>") {
		t.Fatalf("l'anteprima deve essere testo leggibile, non HTML:\n%s", text)
	}
	rs.mu.Lock()
	defer rs.mu.Unlock()
	if rs.writes != 0 {
		t.Fatalf("senza confirm nessuna scrittura, ricevute %d", rs.writes)
	}
}

func TestUpdateStoryRejectsDescriptionTogetherWithPrepend(t *testing.T) {
	rs, srv := newRecordingServerWithBody(t, storyWithCycle)
	session := testSession(t, map[string]bool{"stories": true, "me": true}, srv.URL)

	res := callTool(t, session, "update_story", map[string]any{
		"story_id":    8543,
		"description": "<p>tutto nuovo</p>",
		"prepend":     "<h2>x</h2>",
		"confirm":     true,
	})
	if !res.IsError || !strings.Contains(resultText(res), "non insieme") {
		t.Fatalf("description e prepend insieme vanno rifiutati: %s", resultText(res))
	}
	rs.mu.Lock()
	defer rs.mu.Unlock()
	if rs.writes != 0 {
		t.Fatalf("nessuna scrittura attesa, ricevute %d", rs.writes)
	}
}

func TestUpdateStoryAppendAlreadyPresentWritesNothing(t *testing.T) {
	body := `{"id":8543,"status":"todo","description":"<h2>Secondo ciclo</h2><h2>Primo ciclo</h2>"}`
	rs, srv := newRecordingServerWithBody(t, body)
	session := testSession(t, map[string]bool{"stories": true, "me": true}, srv.URL)

	res := callTool(t, session, "update_story", map[string]any{
		"story_id": 8543,
		"prepend":  "<h2>Secondo ciclo</h2>",
		"confirm":  true,
	})
	if res.IsError || !strings.Contains(resultText(res), "già presente") {
		t.Fatalf("un blocco già in testa non va riaggiunto: %s", resultText(res))
	}
	rs.mu.Lock()
	defer rs.mu.Unlock()
	if rs.writes != 0 {
		t.Fatalf("nessuna scrittura attesa, ricevute %d", rs.writes)
	}
}

func TestUpdateStoryAppendAlreadyPresentStillWritesOtherFields(t *testing.T) {
	body := `{"id":8543,"status":"progress","description":"<h2>Secondo ciclo</h2><h2>Primo ciclo</h2>"}`
	rs, srv := newRecordingServerWithBody(t, body)
	session := testSession(t, map[string]bool{"stories": true, "me": true}, srv.URL)

	res := callTool(t, session, "update_story", map[string]any{
		"story_id": 8543,
		"status":   "todo",
		"prepend":  "<h2>Secondo ciclo</h2>",
		"confirm":  true,
	})
	if res.IsError {
		t.Fatalf("non deve fallire: %s", resultText(res))
	}
	rs.mu.Lock()
	defer rs.mu.Unlock()
	if rs.writes != 1 || rs.lastWriteBody["status"] != "todo" {
		t.Fatalf("lo status va scritto comunque: %+v", rs.lastWriteBody)
	}
	if _, has := rs.lastWriteBody["description"]; has {
		t.Fatalf("la description non va riscritta se non c'è nulla da aggiungere: %+v", rs.lastWriteBody)
	}
}

func TestUpdateStoryAppendOnNullDescription(t *testing.T) {
	rs, srv := newRecordingServerWithBody(t, `{"id":8543,"status":"todo","description":null}`)
	session := testSession(t, map[string]bool{"stories": true, "me": true}, srv.URL)

	res := callTool(t, session, "update_story", map[string]any{
		"story_id": 8543,
		"prepend":  "<h2>Primo ciclo</h2>",
		"confirm":  true,
	})
	if res.IsError {
		t.Fatalf("non deve fallire: %s", resultText(res))
	}
	rs.mu.Lock()
	defer rs.mu.Unlock()
	if rs.lastWriteBody["description"] != "<h2>Primo ciclo</h2>" {
		t.Fatalf("su description nulla il blocco diventa tutto il testo: %+v", rs.lastWriteBody)
	}
}

// m6 — l'anteprima (senza confirm) di un tool di scrittura deve arrivare come
// testo semplice, con a capo veri, non incapsulata in un JSON {"text": …}
// con "\n" scritti a escape.
func TestUpdateStoryPreviewIsPlainTextNotEscapedJSON(t *testing.T) {
	_, srv := newRecordingServer(t)
	session := testSession(t, map[string]bool{"stories": true, "me": true}, srv.URL)

	res := callTool(t, session, "update_story", map[string]any{
		"story_id": 8503,
		"name":     "Nuovo titolo",
	})
	text := resultText(res)
	if strings.Contains(text, `\n`) {
		t.Fatalf("l'anteprima non deve contenere \\n scritti a escape, ma a capo veri: %q", text)
	}
	if strings.Contains(text, `{"text"`) {
		t.Fatalf("l'anteprima non deve essere incapsulata in un JSON: %q", text)
	}
	if !strings.Contains(text, "\n") {
		t.Fatalf("l'anteprima deve contenere a capo veri: %q", text)
	}
}

func TestCreateStoryPreviewIsPlainTextNotEscapedJSON(t *testing.T) {
	_, srv := newRecordingServer(t)
	session := testSession(t, map[string]bool{"stories": true, "me": true}, srv.URL)

	res := callTool(t, session, "create_story", map[string]any{
		"name": "Nuovo ticket",
	})
	text := resultText(res)
	if strings.Contains(text, `\n`) || strings.Contains(text, `{"text"`) {
		t.Fatalf("l'anteprima di create_story non deve essere JSON con \\n a escape: %q", text)
	}
}

func TestUpdateTagPreviewIsPlainTextNotEscapedJSON(t *testing.T) {
	_, srv := newRecordingServer(t)
	session := testSession(t, map[string]bool{"tags": true, "stories": true, "me": true}, srv.URL)

	res := callTool(t, session, "update_tag", map[string]any{
		"tag_id":      42,
		"description": "## Cosa\n\n- a\n- b",
	})
	text := resultText(res)
	if strings.Contains(text, `\n`) || strings.Contains(text, `{"text"`) {
		t.Fatalf("l'anteprima di update_tag non deve essere JSON con \\n a escape: %q", text)
	}
	if !strings.Contains(text, "- a\n") {
		t.Fatalf("gli a capo del Markdown devono restare veri: %q", text)
	}
}

func TestUpdateStoryRejectsUnknownParameter(t *testing.T) {
	rs, srv := newRecordingServer(t)
	session := testSession(t, map[string]bool{"stories": true, "me": true}, srv.URL)

	res := callTool(t, session, "update_story", map[string]any{
		"story_id": 8543,
		"prepen":   "<h2>refuso</h2>",
		"confirm":  true,
	})
	if !res.IsError {
		t.Fatalf("un parametro sconosciuto deve dare errore, non essere ignorato")
	}
	rs.mu.Lock()
	defer rs.mu.Unlock()
	if rs.writes != 0 {
		t.Fatalf("nessuna scrittura attesa, ricevute %d", rs.writes)
	}
}
