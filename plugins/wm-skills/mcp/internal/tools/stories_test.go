package tools

import (
	"strings"
	"testing"

	"github.com/webmappsrl/claude-marketplace/plugins/wm-skills/mcp/internal/enums"
	"github.com/webmappsrl/claude-marketplace/plugins/wm-skills/mcp/internal/preview"
)

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

// m1 — dopo una scrittura riuscita la risposta è una sintesi breve dei campi
// scritti, non l'intera anteprima già vista prima di confermare.
func TestWriteSummaryListsFieldsAndDescriptionLength(t *testing.T) {
	got := writeSummary(
		map[string]any{"description": "vecchia"},
		map[string]any{"status": "done", "description": "nuova description più lunga"},
	)
	if !strings.HasPrefix(got, "Scritto.") {
		t.Fatalf("deve aprirsi con «Scritto.»: %q", got)
	}
	if !strings.Contains(got, "status") || !strings.Contains(got, "description") {
		t.Fatalf("mancano i campi scritti: %q", got)
	}
	if !strings.Contains(got, "description: 7 → 27 caratteri") {
		t.Fatalf("manca il conteggio caratteri della description: %q", got)
	}
}

func TestWriteSummaryWithoutDescriptionHasNoCharacterCount(t *testing.T) {
	got := writeSummary(map[string]any{}, map[string]any{"status": "done"})
	if strings.Contains(got, "caratteri") {
		t.Fatalf("senza description non deve comparire il conteggio caratteri: %q", got)
	}
}

// m5 — quando customer_request cambia, sotto il suo blocco va l'avviso su
// cosa fa Orchestrator con quel testo.
func TestWithCustomerRequestNoticeAddsLineUnderBlock(t *testing.T) {
	diff := preview.Diff(map[string]any{}, map[string]any{"customer_request": "Risolto."}, "customer_request")
	got := withCustomerRequestNotice(diff, map[string]any{"customer_request": "Risolto."})
	if !strings.Contains(got, "Orchestrator mette questo testo in testa alla richiesta, con autore e data, e notifica il cliente.") {
		t.Fatalf("manca l'avviso: %s", got)
	}
}

func TestWithCustomerRequestNoticeLeavesOtherDiffsUntouched(t *testing.T) {
	diff := preview.Diff(map[string]any{}, map[string]any{"status": "done"})
	got := withCustomerRequestNotice(diff, map[string]any{"status": "done"})
	if got != diff {
		t.Fatalf("senza customer_request il diff non va toccato:\n got:  %q\n want: %q", got, diff)
	}
}
