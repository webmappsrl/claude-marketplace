package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestIrreversiblePreviewNamesTheTarget(t *testing.T) {
	// Chi autorizza deve vedere che cosa sta colpendo, non solo un numero.
	text := irreversibleNotice("delete_quote", 418, "Preventivo Acme 2026")
	if !strings.Contains(text, "Preventivo Acme 2026") || !strings.Contains(text, "418") {
		t.Fatalf("l'avviso deve nominare il preventivo: %q", text)
	}
}

// TestCrmToolNamesMatchesRegisteredTools registra davvero il gruppo crm su un
// server e confronta l'elenco dei tool effettivamente esposti con
// crmToolNames(): se un tool viene aggiunto, rinominato o tolto dal gruppo
// crm senza aggiornare l'elenco, questo test fallisce — altrimenti
// crmToolNames() sarebbe solo un elenco che afferma se stesso.
func TestCrmToolNamesMatchesRegisteredTools(t *testing.T) {
	_, srv := newRecordingServer(t)
	session := testSession(t, map[string]bool{"crm": true}, srv.URL)

	res, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools fallita: %v", err)
	}

	got := map[string]bool{}
	for _, tool := range res.Tools {
		got[tool.Name] = true
	}

	want := crmToolNames()
	if len(want) != len(got) {
		t.Fatalf("crmToolNames() elenca %d tool, il gruppo crm ne registra %d: %v vs %v", len(want), len(got), want, got)
	}
	for _, name := range want {
		if !got[name] {
			t.Fatalf("crmToolNames() elenca %q ma il gruppo crm non lo registra", name)
		}
	}
}

// TestIrreversibleToolsWarnInPreview verifica che isIrreversible sia
// effettivamente collegato al testo mostrato in anteprima: i tool marcati
// come non annullabili devono mostrare l'avviso, gli altri no.
func TestIrreversibleToolsWarnInPreview(t *testing.T) {
	_, srv := newRecordingServer(t)
	session := testSession(t, map[string]bool{"crm": true}, srv.URL)

	res := callTool(t, session, "delete_quote", map[string]any{"quote_id": 418})
	if !previewContainsWarning(res) {
		t.Fatalf("delete_quote è marcato isIrreversible ma l'anteprima non contiene l'avviso: %+v", res)
	}

	res = callTool(t, session, "create_quote_pdf_link", map[string]any{"quote_id": 418})
	if !previewContainsWarning(res) {
		t.Fatalf("create_quote_pdf_link è marcato isIrreversible ma l'anteprima non contiene l'avviso: %+v", res)
	}

	res = callTool(t, session, "update_customer", map[string]any{"customer_id": 1, "name": "X"})
	if previewContainsWarning(res) {
		t.Fatalf("update_customer non è irreversibile e non deve mostrare l'avviso: %+v", res)
	}

	if isIrreversible("update_customer") {
		t.Fatal("una modifica correggibile non va marcata come non annullabile")
	}
}

func previewContainsWarning(res *mcp.CallToolResult) bool {
	for _, c := range res.Content {
		if tc, ok := c.(*mcp.TextContent); ok && strings.Contains(tc.Text, "non si può annullare") {
			return true
		}
	}
	return false
}

func TestRegisterWithoutCRMDoesNotPanic(t *testing.T) {
	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "v0"}, nil)
	Register(server, Deps{}, map[string]bool{"stories": true, "me": true})
}

func TestListCustomersPathEncodesAmpersandAndBrackets(t *testing.T) {
	// Le ragioni sociali contengono "&" ("Rossi & Figli"): senza codifica
	// diventa un separatore di parametri, cambiando la richiesta inviata.
	got := listCustomersPath("Rossi & Figli", "active")
	want := "/api/customers?search=Rossi+%26+Figli&status=active"
	if got != want {
		t.Fatalf("percorso non codificato correttamente:\n got:  %s\n want: %s", got, want)
	}
}

func TestListCustomersPathWithoutFiltersHasNoQuery(t *testing.T) {
	if got := listCustomersPath("", ""); got != "/api/customers" {
		t.Fatalf("senza filtri non deve comparire la query string: %s", got)
	}
}

func TestListQuotesPathEncodesStatus(t *testing.T) {
	got := listQuotesPath(42, "in corso & attesa")
	want := "/api/quotes?customer_id=42&status=in+corso+%26+attesa"
	if got != want {
		t.Fatalf("percorso non codificato correttamente:\n got:  %s\n want: %s", got, want)
	}
}

func TestGetQuotePathEncodesInclude(t *testing.T) {
	got := getQuotePath(7, "products,customer & history")
	want := "/api/quotes/7?include=products%2Ccustomer+%26+history"
	if got != want {
		t.Fatalf("percorso non codificato correttamente:\n got:  %s\n want: %s", got, want)
	}
}
