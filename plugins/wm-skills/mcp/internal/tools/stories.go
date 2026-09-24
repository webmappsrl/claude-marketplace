package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/webmappsrl/claude-marketplace/plugins/wm-skills/mcp/internal/compose"
	"github.com/webmappsrl/claude-marketplace/plugins/wm-skills/mcp/internal/enums"
	"github.com/webmappsrl/claude-marketplace/plugins/wm-skills/mcp/internal/preview"
)

// storyHTMLFields sono i campi di una story resi da un editor visuale su
// Orchestrator: solo questi, in anteprima, vanno mostrati come testo
// leggibile invece che grezzi.
var storyHTMLFields = []string{"description", "customer_request"}

type getStoryInput struct {
	StoryID int `json:"story_id" jsonschema:"identificatore numerico del ticket, la parte dopo oc:"`
}

type storyFieldsInput struct {
	StoryID         int                  `json:"story_id,omitempty" jsonschema:"identificatore del ticket da modificare"`
	Name            string               `json:"name,omitempty" jsonschema:"titolo del ticket"`
	Description     string               `json:"description,omitempty" jsonschema:"note di sviluppo, in HTML: il campo è reso da un editor visuale, non interpreta Markdown"`
	CustomerRequest string               `json:"customer_request,omitempty" jsonschema:"richiesta del cliente; scrivendola il cliente riceve una notifica"`
	Type            string               `json:"type,omitempty" jsonschema:"tipo del ticket, fra i valori ammessi dall'API"`
	Status          string               `json:"status,omitempty" jsonschema:"stato del ticket, fra i valori ammessi dall'API"`
	UserID          int                  `json:"user_id,omitempty" jsonschema:"utente assegnatario"`
	CreatorID       int                  `json:"creator_id,omitempty" jsonschema:"utente creatore"`
	EstimatedHours  float64              `json:"estimated_hours,omitempty" jsonschema:"stima in ore"`
	Tags            []int                `json:"tags,omitempty" jsonschema:"elenco completo degli identificatori di tag: sostituisce quelli presenti, per aggiungerne uno solo usa attach_story_to_tag"`
	Confirm         bool                 `json:"confirm,omitempty" jsonschema:"false o assente mostra l'anteprima senza scrivere; true esegue la scrittura"`
	Prepend         string               `json:"prepend,omitempty" jsonschema:"HTML da mettere in testa alla description attuale, che resta identica"`
	Annotations     []compose.Annotation `json:"annotations,omitempty" jsonschema:"testi da inserire dopo frasi esatte della description attuale"`
}

type textOutput struct {
	Text string `json:"text"`
}

func registerStories(server *mcp.Server, deps Deps) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_story",
		Description: "Legge un ticket Orchestrator dal suo identificatore numerico e restituisce tutti i suoi campi.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in getStoryInput) (*mcp.CallToolResult, any, error) {
		raw, err := deps.Client.Do(ctx, "GET", fmt.Sprintf("/api/stories/%d", in.StoryID), nil)
		if err != nil {
			return nil, nil, err
		}
		return nil, jsonOrText(raw), nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "update_story",
		Description: "Modifica un ticket Orchestrator. Senza confirm mostra, campo per campo e per intero, " +
			"cosa cambierebbe, senza scrivere nulla. Per AGGIUNGERE alla description usa prepend (in testa) " +
			"e annotations (dopo frasi esatte): il testo esistente resta identico. description invece " +
			"SOSTITUISCE tutto il campo: usala solo per riformattare o invalidare. " +
			"Scrivere customer_request invia una notifica al cliente: non va mai autorizzato in automatico.",
		InputSchema: storyFieldsInputSchema(deps.Enums, true),
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in storyFieldsInput) (*mcp.CallToolResult, any, error) {
		appending := in.Prepend != "" || len(in.Annotations) > 0
		if appending && in.Description != "" {
			return nil, nil, fmt.Errorf("description sostituisce tutto il campo, prepend e annotations aggiungono: usa l'una o gli altri, non insieme")
		}

		fields := nonEmptyFields(in)
		for _, k := range []string{"story_id", "confirm", "prepend", "annotations"} {
			delete(fields, k)
		}

		if err := validateStoryFields(deps.Enums, fields); err != nil {
			return nil, nil, err
		}

		path := fmt.Sprintf("/api/stories/%d", in.StoryID)
		currentRaw, err := deps.Client.Do(ctx, "GET", path, nil)
		if err != nil {
			return nil, nil, err
		}
		var current map[string]any
		if err := json.Unmarshal(currentRaw, &current); err != nil {
			return nil, nil, fmt.Errorf("ticket illeggibile: %w", err)
		}

		// La composizione si rifà a ogni chiamata, anche a quella con confirm:
		// parte sempre dalla description letta adesso, non da quella vista
		// nell'anteprima.
		var composed compose.Result
		if appending {
			currentText, _ := current["description"].(string)
			composed, err = compose.Apply(currentText, in.Prepend, in.Annotations)
			if err != nil {
				return nil, nil, err
			}
			if composed.Changed {
				fields["description"] = composed.Text
			}
			if len(fields) == 0 {
				return plainText("Nessuna scrittura: il contenuto richiesto è già presente nella description."), nil, nil
			}
		}

		diff := preview.Diff(current, fields, storyHTMLFields...)
		diff = withCustomerRequestNotice(diff, fields)
		summary := previewNotice(fields) + appendSummary(appending, composed) + diff
		if !in.Confirm {
			return plainText(summary), nil, nil
		}

		updated, err := deps.Client.Do(ctx, "PATCH", path, fields)
		if err != nil {
			return nil, nil, err
		}
		return dataResult(writeSummary(current, fields), updated)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_story",
		Description: "Crea un ticket Orchestrator. Senza confirm mostra i campi che verrebbero inviati senza creare nulla.",
		InputSchema: storyFieldsInputSchema(deps.Enums, false),
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in storyFieldsInput) (*mcp.CallToolResult, any, error) {
		fields := nonEmptyFields(in)
		delete(fields, "story_id")
		delete(fields, "confirm")
		delete(fields, "prepend")
		delete(fields, "annotations")

		if fields["name"] == nil {
			return nil, nil, fmt.Errorf("name è obbligatorio per creare un ticket")
		}
		if err := validateStoryFields(deps.Enums, fields); err != nil {
			return nil, nil, err
		}

		if !in.Confirm {
			return plainText(preview.NewResource(fields, storyHTMLFields...)), nil, nil
		}

		raw, err := deps.Client.Do(ctx, "POST", "/api/stories", fields)
		if err != nil {
			return nil, nil, err
		}
		return dataResult("Creato.", raw)
	})
}

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

// appendSummary elenca cosa l'aggiunta mette nella description, prima della
// differenza completa: il blocco in testa e, per ogni annotazione, il blocco
// in cui finisce reso come testo, così il dev vede se l'etichetta è al posto
// giusto.
func appendSummary(appending bool, r compose.Result) string {
	if !appending {
		return ""
	}
	var b strings.Builder
	b.WriteString("Aggiunta alla description\n")
	if r.Prepended {
		b.WriteString("  in testa: sì (il testo esistente resta identico)\n")
	} else {
		b.WriteString("  in testa: niente (vuoto o già presente)\n")
	}
	for i, p := range r.Placed {
		fmt.Fprintf(&b, "  annotazione %d, dopo %q:\n    %s\n", i+1, p.Annotation.After,
			strings.ReplaceAll(preview.Readable(p.Context), "\n", "\n    "))
	}
	for _, s := range r.Skipped {
		fmt.Fprintf(&b, "  già presente, non riaggiunta: %q dopo %q\n", s.Text, s.After)
	}
	return b.String() + "\n"
}

// writeSummary è la risposta data dopo una scrittura riuscita: un elenco
// breve dei campi scritti, non tutta l'anteprima già vista prima di
// confermare. Per description aggiunge quanti caratteri aveva prima e ha
// dopo, il solo dato che su un campo lungo vale la pena riportare.
func writeSummary(current, fields map[string]any) string {
	names := make([]string, 0, len(fields))
	for k := range fields {
		names = append(names, k)
	}
	sort.Strings(names)

	var b strings.Builder
	b.WriteString("Scritto.\n")
	for _, n := range names {
		b.WriteString("  " + n + "\n")
	}
	if after, ok := fields["description"].(string); ok {
		before, _ := current["description"].(string)
		fmt.Fprintf(&b, "  description: %d → %d caratteri\n", utf8.RuneCountInString(before), utf8.RuneCountInString(after))
	}
	return strings.TrimRight(b.String(), "\n")
}

// withCustomerRequestNotice aggiunge, subito sotto il blocco di
// customer_request nell'anteprima, l'avviso su cosa fa Orchestrator con quel
// testo: non tocca il diff se il campo non cambia.
func withCustomerRequestNotice(diff string, fields map[string]any) string {
	if _, changed := fields["customer_request"]; !changed {
		return diff
	}
	marker := "customer_request\n"
	idx := strings.Index(diff, marker)
	if idx < 0 {
		return diff
	}
	end := strings.Index(diff[idx:], "\n\n")
	if end < 0 {
		return diff
	}
	insertAt := idx + end
	notice := "\n  Orchestrator mette questo testo in testa alla richiesta, con autore e data, e notifica il cliente."
	return diff[:insertAt] + notice + diff[insertAt:]
}

type emptyInput struct{}

func registerMe(server *mcp.Server, deps Deps) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "me",
		Description: "Restituisce l'utente Orchestrator corrispondente alle credenziali in uso.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
		raw, err := deps.Client.Do(ctx, "GET", "/api/me", nil)
		if err != nil {
			return nil, nil, err
		}
		return nil, jsonOrText(raw), nil
	})
}

// storyFieldsInputSchema costruisce lo schema di ingresso di create_story e
// update_story a runtime, per rendere i valori inventati per type e status
// non esprimibili invece che solo respinti dopo l'invio. Quando gli elenchi
// non sono disponibili (deps.Enums nil o vuoto), il campo resta una stringa
// libera: la convalida interna in validateStoryFields resta comunque attiva
// come rete di sicurezza; con forUpdate aggiunge i parametri dell'aggiunta.
func storyFieldsInputSchema(e *enums.Values, forUpdate bool) map[string]any {
	typeSchema := map[string]any{
		"type":        "string",
		"description": "tipo del ticket, fra i valori ammessi dall'API",
	}
	statusSchema := map[string]any{
		"type":        "string",
		"description": "stato del ticket, fra i valori ammessi dall'API",
	}
	if e != nil && len(e.StoryType) > 0 {
		typeSchema["enum"] = e.StoryType
		typeSchema["description"] = fmt.Sprintf("tipo del ticket — valori ammessi: %s", strings.Join(e.StoryType, ", "))
	}
	if e != nil && len(e.StoryStatus) > 0 {
		statusSchema["enum"] = e.StoryStatus
		statusSchema["description"] = fmt.Sprintf("stato del ticket — valori ammessi: %s", strings.Join(e.StoryStatus, ", "))
	}

	properties := map[string]any{
		"story_id": map[string]any{
			"type":        "integer",
			"description": "identificatore del ticket da modificare",
		},
		"name": map[string]any{
			"type":        "string",
			"description": "titolo del ticket",
		},
		"description": map[string]any{
			"type":        "string",
			"description": "note di sviluppo, in HTML: il campo è reso da un editor visuale, non interpreta Markdown",
		},
		"customer_request": map[string]any{
			"type":        "string",
			"description": "richiesta del cliente; scrivendola il cliente riceve una notifica",
		},
		"type":   typeSchema,
		"status": statusSchema,
		"user_id": map[string]any{
			"type":        "integer",
			"description": "utente assegnatario",
		},
		"creator_id": map[string]any{
			"type":        "integer",
			"description": "utente creatore",
		},
		"estimated_hours": map[string]any{
			"type":        "number",
			"description": "stima in ore",
		},
		"tags": map[string]any{
			"type":        "array",
			"items":       map[string]any{"type": "integer"},
			"description": "elenco completo degli identificatori di tag: sostituisce quelli presenti, per aggiungerne uno solo usa attach_story_to_tag",
		},
		"confirm": map[string]any{
			"type":        "boolean",
			"description": "false o assente mostra l'anteprima senza scrivere; true esegue la scrittura",
		},
	}
	if forUpdate {
		properties["prepend"] = map[string]any{
			"type":        "string",
			"description": "HTML da mettere in testa alla description attuale, che resta identica. Non insieme a description.",
		}
		properties["annotations"] = map[string]any{
			"type":        "array",
			"description": "testi da inserire subito dopo frasi esatte della description attuale (es. «✅ Risolto» accanto a un bloccante). Non insieme a description.",
			"items": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"after":      map[string]any{"type": "string", "description": "frase esatta, come compare nell'HTML della description attuale"},
					"text":       map[string]any{"type": "string", "description": "HTML da inserire dopo la frase, preceduto da uno spazio"},
					"occurrence": map[string]any{"type": "integer", "description": "quale presenza usare se la frase compare più volte (1 = la prima)"},
				},
				"required":             []string{"after", "text"},
				"additionalProperties": false,
			},
		}
	}
	return map[string]any{
		"type":                 "object",
		"properties":           properties,
		"additionalProperties": false,
	}
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
