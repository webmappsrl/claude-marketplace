package tools

import (
	"context"
	"fmt"
	"net/url"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/webmappsrl/claude-marketplace/plugins/wm-skills/mcp/internal/preview"
)

const tagDescriptionHint = "descrizione del tag, in Markdown: a differenza della description di una story, questo campo è reso da un editor Markdown"

type listTagsInput struct {
	Search string `json:"search,omitempty" jsonschema:"ricerca sul nome del tag"`
}

type getTagInput struct {
	TagID int `json:"tag_id" jsonschema:"identificatore del tag"`
}

type tagFieldsInput struct {
	TagID       int    `json:"tag_id,omitempty" jsonschema:"identificatore del tag da modificare"`
	Name        string `json:"name,omitempty" jsonschema:"nome del tag"`
	Description string `json:"description,omitempty" jsonschema:"descrizione del tag, in Markdown"`
	Confirm     bool   `json:"confirm,omitempty" jsonschema:"false o assente mostra l'anteprima senza scrivere; true esegue la scrittura"`
}

type tagStoryInput struct {
	TagID   int  `json:"tag_id" jsonschema:"identificatore del tag"`
	StoryID int  `json:"story_id" jsonschema:"identificatore del ticket"`
	Confirm bool `json:"confirm,omitempty" jsonschema:"false o assente mostra cosa verrebbe fatto; true esegue"`
}

func attachPath(tagID, storyID int) string {
	return fmt.Sprintf("/api/tags/%d/stories/%d", tagID, storyID)
}

// listTagsPath costruisce il percorso di list_tags codificando il filtro di
// ricerca con net/url: i nomi dei tag di questo team contengono parentesi
// quadre ("[RDO][CLIENTE][ANNO]N") e una concatenazione di stringhe grezza le
// manderebbe non codificate, cambiando silenziosamente la richiesta.
//
// Il parametro è "search", non "name": verificato contro l'istanza locale —
// "?name=" viene ignorato dall'API e restituisce tutti i tag, "?search="
// applica davvero il filtro.
func listTagsPath(search string) string {
	path := "/api/tags"
	if search != "" {
		q := url.Values{}
		q.Set("search", search)
		path += "?" + q.Encode()
	}
	return path
}

func registerTags(server *mcp.Server, deps Deps) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_tags",
		Description: "Elenca i tag Orchestrator, con ricerca facoltativa sul nome.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in listTagsInput) (*mcp.CallToolResult, any, error) {
		path := listTagsPath(in.Search)
		raw, err := deps.Client.Do(ctx, "GET", path, nil)
		if err != nil {
			return nil, nil, err
		}
		return nil, jsonOrText(raw), nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_tag",
		Description: "Legge un tag con i ticket a esso associati.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in getTagInput) (*mcp.CallToolResult, any, error) {
		raw, err := deps.Client.Do(ctx, "GET", fmt.Sprintf("/api/tags/%d", in.TagID), nil)
		if err != nil {
			return nil, nil, err
		}
		return nil, jsonOrText(raw), nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_tag",
		Description: "Crea un tag Orchestrator. Senza confirm mostra i campi senza creare nulla. " + tagDescriptionHint + ".",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in tagFieldsInput) (*mcp.CallToolResult, any, error) {
		fields := map[string]any{"name": in.Name}
		if in.Description != "" {
			fields["description"] = in.Description
		}
		if in.Name == "" {
			return nil, nil, fmt.Errorf("name è obbligatorio per creare un tag")
		}
		if !in.Confirm {
			return plainText(preview.NewResource(fields)), nil, nil
		}
		raw, err := deps.Client.Do(ctx, "POST", "/api/tags", fields)
		if err != nil {
			return nil, nil, err
		}
		return dataResult("Creato.", raw)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_tag",
		Description: "Modifica un tag Orchestrator. Senza confirm mostra la differenza senza scrivere. " + tagDescriptionHint + ".",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in tagFieldsInput) (*mcp.CallToolResult, any, error) {
		fields := map[string]any{}
		if in.Name != "" {
			fields["name"] = in.Name
		}
		if in.Description != "" {
			fields["description"] = in.Description
		}
		path := fmt.Sprintf("/api/tags/%d", in.TagID)

		currentRaw, err := deps.Client.Do(ctx, "GET", path, nil)
		if err != nil {
			return nil, nil, err
		}
		current := decodeMap(currentRaw)

		if !in.Confirm {
			return plainText(preview.Diff(current, fields)), nil, nil
		}
		raw, err := deps.Client.Do(ctx, "PATCH", path, fields)
		if err != nil {
			return nil, nil, err
		}
		return dataResult("Scritto.", raw)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "attach_story_to_tag",
		Description: "Associa un ticket a un tag senza toccare gli altri tag del ticket. " +
			"Da preferire sempre alla modifica del campo tags, che sostituisce l'elenco completo e può cancellare tag già presenti.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in tagStoryInput) (*mcp.CallToolResult, any, error) {
		if !in.Confirm {
			return plainText(fmt.Sprintf("ANTEPRIMA — nulla è stato scritto\n  il ticket %d verrebbe associato al tag %d\n\nPer applicare, richiama con confirm: true.", in.StoryID, in.TagID)), nil, nil
		}
		raw, err := deps.Client.Do(ctx, "POST", attachPath(in.TagID, in.StoryID), nil)
		if err != nil {
			return nil, nil, err
		}
		return dataResult("Associato.", raw)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "detach_story_from_tag",
		Description: "Toglie l'associazione fra un ticket e un tag, lasciando intatti gli altri tag del ticket.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in tagStoryInput) (*mcp.CallToolResult, any, error) {
		if !in.Confirm {
			return plainText(fmt.Sprintf("ANTEPRIMA — nulla è stato scritto\n  il ticket %d verrebbe tolto dal tag %d\n\nPer applicare, richiama con confirm: true.", in.StoryID, in.TagID)), nil, nil
		}
		raw, err := deps.Client.Do(ctx, "DELETE", attachPath(in.TagID, in.StoryID), nil)
		if err != nil {
			return nil, nil, err
		}
		return dataResult("Rimosso.", raw)
	})
}
