package tools

import (
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// collectionKey è la chiave sotto cui finisce un corpo JSON che non è un
// oggetto. È volutamente generica e uguale per tutti i tool: una chiave per
// risorsa ("tags", "tasks", …) costringerebbe a ricordarsene una nuova ad ogni
// endpoint aggiunto, ed è proprio la dimenticanza che ha prodotto oc:8545.
const collectionKey = "items"

// jsonOrText restituisce il corpo di Orchestrator già deserializzato quando è
// JSON valido, così l'SDK lo serializza una sola volta invece di incollarlo,
// con tutti i suoi escape annidati, dentro il campo "text" di una struttura.
// Se il corpo non è JSON valido (caso anomalo), ricade sul solo testo.
//
// Quando il JSON non è un oggetto lo incapsula sotto collectionKey (oc:8545).
// Il protocollo MCP impone che structuredContent sia un record: gli endpoint
// di collezione di Orchestrator (GET /api/tags, /api/tasks, /api/customers…)
// rispondono invece con un array top-level, e restituirlo nudo fa rifiutare
// l'intera risposta dal client, prima ancora che il modello la veda. Il caso
// va trattato qui e non nel singolo handler perché questa funzione è l'unico
// punto da cui passano tutti i tool di lettura: correggerne uno alla volta
// lascerebbe il difetto pronto a ricomparire sul prossimo endpoint di lista.
func jsonOrText(raw []byte) any {
	var v any
	if err := json.Unmarshal(raw, &v); err == nil {
		if _, isObject := v.(map[string]any); isObject {
			return v
		}
		return map[string]any{collectionKey: v}
	}
	return textOutput{Text: string(raw)}
}

// dataResult tiene distinti il messaggio discorsivo ("Creato.", "Scritto.", …)
// e il dato JSON restituito da Orchestrator, invece di incollarli in un'unica
// stringa piena di escape annidati: il primo resta un blocco di testo, il
// secondo diventa contenuto strutturato. Se il corpo non è JSON valido (caso
// anomalo), ricade su un unico blocco di testo che li unisce comunque, così
// non si perde nulla.
func dataResult(message string, raw []byte) (*mcp.CallToolResult, any, error) {
	var data any
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, textOutput{Text: message + "\n" + string(raw)}, nil
	}
	return &mcp.CallToolResult{
		Content:           []mcp.Content{&mcp.TextContent{Text: message}},
		StructuredContent: data,
	}, nil, nil
}
