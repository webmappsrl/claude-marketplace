package tools

import (
	"strings"
	"testing"

	"github.com/webmappsrl/claude-marketplace/plugins/wm-skills/mcp/internal/enums"
)

func quoteEnums() *enums.Values {
	return &enums.Values{QuoteStatus: []string{"new", "presented", "on hold"}}
}

func statusProperty(t *testing.T, schema map[string]any) map[string]any {
	t.Helper()
	props, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("schema senza properties: %v", schema)
	}
	status, ok := props["status"].(map[string]any)
	if !ok {
		t.Fatalf("schema senza status: %v", props)
	}
	return status
}

func TestQuoteSchemasListAllowedStatuses(t *testing.T) {
	for name, schema := range map[string]map[string]any{
		"list_quotes":  listQuotesInputSchema(quoteEnums()),
		"quote_fields": quoteFieldsInputSchema(quoteEnums()),
	} {
		status := statusProperty(t, schema)
		values, ok := status["enum"].([]string)
		if !ok || len(values) != 3 || values[2] != "on hold" {
			t.Fatalf("%s: enum atteso con i valori letterali, ottenuto %v", name, status["enum"])
		}
		if !strings.Contains(status["description"].(string), "on hold") {
			t.Fatalf("%s: la descrizione deve elencare i valori ammessi: %v", name, status["description"])
		}
	}
}

func TestQuoteSchemasWithoutEnumsKeepFreeString(t *testing.T) {
	status := statusProperty(t, listQuotesInputSchema(nil))
	if _, has := status["enum"]; has {
		t.Fatalf("senza elenchi lo stato deve restare una stringa libera: %v", status)
	}
}

func TestQuoteFieldsSchemaCoversInputStruct(t *testing.T) {
	// Lo schema è scritto a mano: ogni campo di quoteFieldsInput deve
	// esserci, altrimenti additionalProperties lo rifiuterebbe.
	props := quoteFieldsInputSchema(nil)["properties"].(map[string]any)
	for _, field := range []string{"quote_id", "title", "customer_id", "status", "priority", "discount", "google_drive_url", "notes", "confirm"} {
		if _, ok := props[field]; !ok {
			t.Fatalf("campo %q mancante nello schema di create_quote/update_quote", field)
		}
	}
}

func TestValidateQuoteStatus(t *testing.T) {
	e := quoteEnums()
	if err := validateQuoteStatus(e, "on hold"); err != nil {
		t.Fatalf("valore ammesso rifiutato: %v", err)
	}
	if err := validateQuoteStatus(e, ""); err != nil {
		t.Fatalf("stato vuoto = nessun filtro, non un errore: %v", err)
	}
	err := validateQuoteStatus(e, "On_Hold")
	if err == nil || !strings.Contains(err.Error(), "on hold") {
		t.Fatalf("atteso errore con i valori possibili, ottenuto %v", err)
	}
	if err := validateQuoteStatus(nil, "qualunque"); err != nil {
		t.Fatalf("senza elenchi non si blocca nulla: %v", err)
	}
}
