package tools

import (
	"strings"
	"testing"

	"github.com/webmappsrl/claude-marketplace/plugins/wm-skills/mcp/internal/enums"
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
