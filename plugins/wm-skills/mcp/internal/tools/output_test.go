package tools

import (
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Le prove seguenti verificano la forma della risposta, non solo il
// contenuto: un tool di lettura (attraverso jsonOrText, usata da get_story e
// da ogni altro tool di lettura) e un tool di scrittura (attraverso
// dataResult, usata da update_story/create_story e dagli altri tool di
// scrittura). Il difetto che correggono è l'annidamento: prima di questa
// correzione il JSON di Orchestrator finiva dentro il campo "text" di una
// struttura, poi l'intera struttura veniva a sua volta serializzata,
// producendo una stringa piena di escape annidati.

func TestJSONOrTextDecodesValidJSONInsteadOfNestingIt(t *testing.T) {
	// Come get_story: il corpo restituito da Orchestrator è già JSON.
	raw := []byte(`{"id":8420,"name":"Aggiorna Ticket"}`)
	got := jsonOrText(raw)

	m, ok := got.(map[string]any)
	if !ok {
		t.Fatalf("un corpo JSON valido deve tornare come dato deserializzato, non come stringa: %#v", got)
	}
	if m["id"] != float64(8420) || m["name"] != "Aggiorna Ticket" {
		t.Fatalf("i campi del ticket devono restare leggibili direttamente: %#v", m)
	}
}

func TestJSONOrTextFallsBackToTextWhenNotJSON(t *testing.T) {
	got := jsonOrText([]byte("non è json"))
	out, ok := got.(textOutput)
	if !ok || out.Text != "non è json" {
		t.Fatalf("un corpo non JSON deve restare testo semplice: %#v", got)
	}
}

func TestDataResultKeepsMessageAndDataDistinct(t *testing.T) {
	// Come update_story dopo la scrittura: prima della correzione il
	// messaggio e il JSON finivano incollati in un'unica stringa
	// ("Scritto.\n{...}"); ora devono restare parti separate della risposta.
	res, out, err := dataResult("Scritto.", []byte(`{"id":8420,"status":"progress"}`))
	if err != nil {
		t.Fatalf("errore inatteso: %v", err)
	}
	if out != nil {
		t.Fatalf("con un risultato costruito a mano il valore Out deve restare nil, non essere rimarshalato sopra: %#v", out)
	}
	if res == nil {
		t.Fatal("ci si aspetta un CallToolResult costruito esplicitamente")
	}

	if len(res.Content) != 1 {
		t.Fatalf("il messaggio discorsivo deve essere l'unico blocco di testo: %#v", res.Content)
	}
	text, ok := res.Content[0].(*mcp.TextContent)
	if !ok || text.Text != "Scritto." {
		t.Fatalf("il messaggio deve restare distinto dal dato, non contenere JSON al suo interno: %#v", res.Content[0])
	}

	data, ok := res.StructuredContent.(map[string]any)
	if !ok {
		t.Fatalf("il dato deve essere strutturato, non una stringa JSON annidata: %#v", res.StructuredContent)
	}
	if data["status"] != "progress" {
		t.Fatalf("il dato deve restare quello restituito da Orchestrator: %#v", data)
	}
}

func TestDataResultFallsBackToTextWhenNotJSON(t *testing.T) {
	res, out, err := dataResult("Creato.", []byte("non è json"))
	if err != nil {
		t.Fatalf("errore inatteso: %v", err)
	}
	if res != nil {
		t.Fatalf("nel caso anomalo non JSON non si costruisce un risultato su misura: %#v", res)
	}
	text, ok := out.(textOutput)
	if !ok {
		t.Fatalf("il fallback deve restare testo semplice: %#v", out)
	}
	if text.Text != "Creato.\nnon è json" {
		t.Fatalf("nel fallback messaggio e corpo restano uniti, per non perdere nulla: %q", text.Text)
	}
}
