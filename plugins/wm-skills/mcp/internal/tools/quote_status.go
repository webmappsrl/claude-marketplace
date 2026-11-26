package tools

import (
	"fmt"
	"strings"

	"github.com/webmappsrl/claude-marketplace/plugins/wm-skills/mcp/internal/enums"
)

// Lo stato di un preventivo ha valori fissi, letti da QuoteStatus.php come gli
// stati dei ticket. Senza l'elenco nello schema l'LLM indovina il valore
// ("On_Hold" invece di "on hold"): GET /api/quotes risponde 422 a un valore
// sconosciuto (oc:8723), create_quote e update_quote lo rifiutano.

// quoteStatusSchema descrive il campo status; con gli elenchi disponibili
// aggiunge enum e valori ammessi, altrimenti resta una stringa libera.
func quoteStatusSchema(e *enums.Values, description string) map[string]any {
	schema := map[string]any{
		"type":        "string",
		"description": description,
	}
	if e != nil && len(e.QuoteStatus) > 0 {
		schema["enum"] = e.QuoteStatus
		schema["description"] = fmt.Sprintf("%s — valori ammessi: %s", description, strings.Join(e.QuoteStatus, ", "))
	}
	return schema
}

func listQuotesInputSchema(e *enums.Values) map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"customer_id": map[string]any{
				"type":        "integer",
				"description": "filtra i preventivi di un cliente",
			},
			"status": quoteStatusSchema(e, "filtro sullo stato del preventivo"),
		},
		"additionalProperties": false,
	}
}

func quoteFieldsInputSchema(e *enums.Values) map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"quote_id": map[string]any{
				"type":        "integer",
				"description": "identificatore del preventivo da modificare",
			},
			"title": map[string]any{
				"type":        "string",
				"description": "titolo del preventivo, obbligatorio in creazione",
			},
			"customer_id": map[string]any{
				"type":        "integer",
				"description": "cliente intestatario, obbligatorio in creazione",
			},
			"status": quoteStatusSchema(e, "stato del preventivo; a preventivo chiuso le modifiche vengono rifiutate"),
			"priority": map[string]any{
				"type":        "integer",
				"description": "priorità",
			},
			"discount": map[string]any{
				"type":        "number",
				"description": "sconto",
			},
			"google_drive_url": map[string]any{
				"type":        "string",
				"description": "collegamento alla cartella Drive",
			},
			"notes": map[string]any{
				"type":        "string",
				"description": "note interne",
			},
			"confirm": map[string]any{
				"type":        "boolean",
				"description": "false o assente mostra l'anteprima senza scrivere; true esegue la scrittura",
			},
		},
		"additionalProperties": false,
	}
}

// validateQuoteStatus rifiuta un valore fuori elenco prima della chiamata. È
// una rete di sicurezza: la difesa principale è l'elenco nello schema.
func validateQuoteStatus(e *enums.Values, status string) error {
	if status == "" || e == nil || len(e.QuoteStatus) == 0 {
		return nil
	}
	for _, allowed := range e.QuoteStatus {
		if allowed == status {
			return nil
		}
	}
	return fmt.Errorf("status: il valore %q non è ammesso — valori possibili: %s",
		status, strings.Join(e.QuoteStatus, ", "))
}
