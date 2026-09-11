// Package tools dichiara i tool esposti dal server. Nomi, descrizioni e
// appartenenza ai gruppi sono scritti a mano; i valori ammessi dei campi
// arrivano dagli enum PHP di Orchestrator (vedi internal/enums), non da una
// specifica OpenAPI.
package tools

import (
	"encoding/json"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/webmappsrl/claude-marketplace/plugins/wm-skills/mcp/internal/client"
	"github.com/webmappsrl/claude-marketplace/plugins/wm-skills/mcp/internal/enums"
)

// Deps raccoglie ciò di cui i tool hanno bisogno.
type Deps struct {
	Client *client.Client
	Enums  *enums.Values
}

var known = []string{"stories", "me", "tags", "tasks", "crm"}

// ActiveGroups interpreta la variabile ORCHESTRATOR_MCP_GROUPS. I gruppi
// stories e me sono sempre attivi: chi apre wm-plan sta quasi sempre
// lavorando su un ticket.
func ActiveGroups(env string) map[string]bool {
	active := map[string]bool{"stories": true, "me": true}
	for _, raw := range strings.Split(env, ",") {
		name := strings.TrimSpace(raw)
		for _, k := range known {
			if name == k {
				active[k] = true
			}
		}
	}
	return active
}

// Register aggiunge al server i tool dei gruppi attivi.
func Register(server *mcp.Server, deps Deps, groups map[string]bool) {
	if groups["stories"] {
		registerStories(server, deps)
	}
	if groups["me"] {
		registerMe(server, deps)
	}
	if groups["tags"] {
		registerTags(server, deps)
	}
	if groups["tasks"] {
		registerTasks(server, deps)
	}
	if groups["crm"] {
		registerCRM(server, deps)
	}
}

// decodeMap converte una risposta JSON grezza in mappa, ignorando errori di
// decodifica: usata per confrontare lo stato attuale di una risorsa con i
// campi richiesti in scrittura.
func decodeMap(raw []byte) map[string]any {
	m := map[string]any{}
	_ = json.Unmarshal(raw, &m)
	return m
}

// fieldsOf estrae dalla struttura di ingresso i soli campi valorizzati (grazie
// a "omitempty" nel tag json), escludendo quelli indicati — tipicamente
// l'identificatore della risorsa e il parametro confirm.
func fieldsOf(in any, exclude ...string) map[string]any {
	raw, _ := json.Marshal(in)
	fields := map[string]any{}
	_ = json.Unmarshal(raw, &fields)
	for _, name := range exclude {
		delete(fields, name)
	}
	return fields
}
