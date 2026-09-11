package tools

import (
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// jsonOrText restituisce il corpo di Orchestrator già deserializzato quando è
// JSON valido, così l'SDK lo serializza una sola volta invece di incollarlo,
// con tutti i suoi escape annidati, dentro il campo "text" di una struttura.
// Se il corpo non è JSON valido (caso anomalo), ricade sul solo testo.
func jsonOrText(raw []byte) any {
	var v any
	if err := json.Unmarshal(raw, &v); err == nil {
		return v
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
