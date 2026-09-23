package tools

import (
	"context"
	"fmt"
	"net/url"

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

// listTasksPath costruisce il percorso di list_tasks codificando
// l'ordinamento con net/url invece di concatenarlo grezzo nella query string.
func listTasksPath(sort string) string {
	path := "/api/tasks"
	if sort != "" {
		q := url.Values{}
		q.Set("sort", sort)
		path += "?" + q.Encode()
	}
	return path
}

func registerTasks(server *mcp.Server, deps Deps) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_tasks",
		Description: "Elenca i task dell'utente autenticato: quelli sui preventivi che possiede o che ha creato.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in listTasksInput) (*mcp.CallToolResult, any, error) {
		path := listTasksPath(in.Sort)
		raw, err := deps.Client.Do(ctx, "GET", path, nil)
		if err != nil {
			return nil, nil, err
		}
		return nil, jsonOrText(raw), nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_task",
		Description: "Legge il dettaglio di un task.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in getTaskInput) (*mcp.CallToolResult, any, error) {
		raw, err := deps.Client.Do(ctx, "GET", fmt.Sprintf("/api/tasks/%d", in.TaskID), nil)
		if err != nil {
			return nil, nil, err
		}
		return nil, jsonOrText(raw), nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_task",
		Description: "Crea un task su un preventivo esistente. Senza confirm mostra i campi senza creare nulla.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in taskFieldsInput) (*mcp.CallToolResult, any, error) {
		if in.QuoteID == 0 {
			return nil, nil, fmt.Errorf("quote_id è obbligatorio per creare un task")
		}
		fields := map[string]any{}
		fields["quote_id"] = in.QuoteID
		if in.Title != "" {
			fields["title"] = in.Title
		}
		if in.Notes != "" {
			fields["notes"] = in.Notes
		}
		if !in.Confirm {
			return plainText(preview.NewResource(fields)), nil, nil
		}
		raw, err := deps.Client.Do(ctx, "POST", "/api/tasks", fields)
		if err != nil {
			return nil, nil, err
		}
		return dataResult("Creato.", raw)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "update_task",
		Description: "Modifica stato o note di un task. Senza confirm mostra la differenza senza scrivere. " +
			"Lo stato è modificabile solo da chi ha creato il task: se non sei tu, l'intera richiesta viene rifiutata e nemmeno le note vengono salvate.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in taskFieldsInput) (*mcp.CallToolResult, any, error) {
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
			return nil, nil, err
		}
		if !in.Confirm {
			return plainText(preview.Diff(decodeMap(currentRaw), fields)), nil, nil
		}
		raw, err := deps.Client.Do(ctx, "PATCH", path, fields)
		if err != nil {
			return nil, nil, err
		}
		return dataResult("Scritto.", raw)
	})
}
