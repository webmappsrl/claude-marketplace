package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/webmappsrl/claude-marketplace/plugins/wm-skills/mcp/internal/enums"
	"github.com/webmappsrl/claude-marketplace/plugins/wm-skills/mcp/internal/preview"
)

type getStoryInput struct {
	StoryID int `json:"story_id" jsonschema:"identificatore numerico del ticket, la parte dopo oc:"`
}

type storyFieldsInput struct {
	StoryID         int     `json:"story_id,omitempty" jsonschema:"identificatore del ticket da modificare"`
	Name            string  `json:"name,omitempty" jsonschema:"titolo del ticket"`
	Description     string  `json:"description,omitempty" jsonschema:"note di sviluppo, in HTML: il campo è reso da un editor visuale, non interpreta Markdown"`
	CustomerRequest string  `json:"customer_request,omitempty" jsonschema:"richiesta del cliente; scrivendola il cliente riceve una notifica"`
	Type            string  `json:"type,omitempty" jsonschema:"tipo del ticket, fra i valori ammessi dall'API"`
	Status          string  `json:"status,omitempty" jsonschema:"stato del ticket, fra i valori ammessi dall'API"`
	UserID          int     `json:"user_id,omitempty" jsonschema:"utente assegnatario"`
	CreatorID       int     `json:"creator_id,omitempty" jsonschema:"utente creatore"`
	EstimatedHours  float64 `json:"estimated_hours,omitempty" jsonschema:"stima in ore"`
	Tags            []int   `json:"tags,omitempty" jsonschema:"elenco completo degli identificatori di tag: sostituisce quelli presenti, per aggiungerne uno solo usa attach_story_to_tag"`
	Confirm         bool    `json:"confirm,omitempty" jsonschema:"false o assente mostra l'anteprima senza scrivere; true esegue la scrittura"`
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
		Name:        "update_story",
		Description: "Modifica un ticket Orchestrator. Senza confirm mostra la differenza rispetto allo stato attuale senza scrivere nulla. Scrivere customer_request invia una notifica al cliente: non va mai autorizzato in automatico.",
		InputSchema: storyFieldsInputSchema(deps.Enums),
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in storyFieldsInput) (*mcp.CallToolResult, any, error) {
		fields := nonEmptyFields(in)
		delete(fields, "story_id")
		delete(fields, "confirm")

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

		if !in.Confirm {
			return nil, textOutput{Text: previewNotice(fields) + preview.Diff(current, fields)}, nil
		}

		updated, err := deps.Client.Do(ctx, "PATCH", path, fields)
		if err != nil {
			return nil, nil, err
		}
		return dataResult("Scritto.\n"+preview.Diff(current, fields), updated)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_story",
		Description: "Crea un ticket Orchestrator. Senza confirm mostra i campi che verrebbero inviati senza creare nulla.",
		InputSchema: storyFieldsInputSchema(deps.Enums),
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in storyFieldsInput) (*mcp.CallToolResult, any, error) {
		fields := nonEmptyFields(in)
		delete(fields, "story_id")
		delete(fields, "confirm")

		if fields["name"] == nil {
			return nil, nil, fmt.Errorf("name è obbligatorio per creare un ticket")
		}
		if err := validateStoryFields(deps.Enums, fields); err != nil {
			return nil, nil, err
		}

		if !in.Confirm {
			return nil, textOutput{Text: preview.NewResource(fields)}, nil
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
// come rete di sicurezza.
func storyFieldsInputSchema(e *enums.Values) map[string]any {
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

	return map[string]any{
		"type": "object",
		"properties": map[string]any{
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
		},
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
