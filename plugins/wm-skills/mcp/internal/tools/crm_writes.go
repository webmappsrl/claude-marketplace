package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/webmappsrl/claude-marketplace/plugins/wm-skills/mcp/internal/preview"
)

type customerFieldsInput struct {
	CustomerID    int      `json:"customer_id,omitempty" jsonschema:"identificatore del cliente da modificare"`
	Name          string   `json:"name,omitempty" jsonschema:"nome del referente"`
	CompanyName   string   `json:"company_name,omitempty" jsonschema:"ragione sociale"`
	Vat           string   `json:"vat,omitempty" jsonschema:"partita IVA"`
	Address       string   `json:"address,omitempty" jsonschema:"indirizzo"`
	Phone         string   `json:"phone,omitempty" jsonschema:"telefono"`
	Status        string   `json:"status,omitempty" jsonschema:"stato del cliente"`
	Notes         string   `json:"notes,omitempty" jsonschema:"note interne"`
	ContactEmails []string `json:"contact_emails,omitempty" jsonschema:"elenco completo delle email di contatto: sostituisce quelle presenti"`
	Confirm       bool     `json:"confirm,omitempty" jsonschema:"false o assente mostra l'anteprima senza scrivere; true esegue la scrittura"`
}

type quoteFieldsInput struct {
	QuoteID        int     `json:"quote_id,omitempty" jsonschema:"identificatore del preventivo da modificare"`
	Title          string  `json:"title,omitempty" jsonschema:"titolo del preventivo, obbligatorio in creazione"`
	CustomerID     int     `json:"customer_id,omitempty" jsonschema:"cliente intestatario, obbligatorio in creazione"`
	Status         string  `json:"status,omitempty" jsonschema:"stato del preventivo; a preventivo chiuso le modifiche vengono rifiutate"`
	Priority       int     `json:"priority,omitempty" jsonschema:"priorità"`
	Discount       float64 `json:"discount,omitempty" jsonschema:"sconto"`
	GoogleDriveURL string  `json:"google_drive_url,omitempty" jsonschema:"collegamento alla cartella Drive"`
	Notes          string  `json:"notes,omitempty" jsonschema:"note interne"`
	Confirm        bool    `json:"confirm,omitempty" jsonschema:"false o assente mostra l'anteprima senza scrivere; true esegue la scrittura"`
}

type deleteQuoteInput struct {
	QuoteID    int    `json:"quote_id" jsonschema:"identificatore del preventivo da eliminare"`
	Confirm    bool   `json:"confirm,omitempty" jsonschema:"false o assente mostra cosa verrebbe eliminato; true esegue"`
	QuoteTitle string `json:"quote_title,omitempty" jsonschema:"titolo del preventivo, a scopo informativo: compare nella richiesta di autorizzazione così chi approva vede quale preventivo sta eliminando invece del solo numero"`
}

type quoteProductInput struct {
	QuoteID   int  `json:"quote_id" jsonschema:"identificatore del preventivo"`
	ProductID int  `json:"product_id" jsonschema:"identificatore del prodotto: da list_products se recurring è false/assente, da list_recurring_products se recurring è true — sono due cataloghi distinti, un identificatore dell'uno non è valido nell'altro"`
	Quantity  int  `json:"quantity,omitempty" jsonschema:"quantità, obbligatoria quando si associa"`
	Recurring bool `json:"recurring,omitempty" jsonschema:"true se si tratta di un prodotto ricorrente: in tal caso product_id va preso da list_recurring_products, non da list_products"`
	Confirm   bool `json:"confirm,omitempty" jsonschema:"false o assente mostra cosa verrebbe fatto; true esegue"`
}

type pdfLinkInput struct {
	QuoteID       int    `json:"quote_id" jsonschema:"identificatore del preventivo"`
	Lang          string `json:"lang,omitempty" jsonschema:"lingua del documento"`
	ExpiresInDays int    `json:"expires_in_days,omitempty" jsonschema:"durata del collegamento in giorni, massimo 90"`
	Confirm       bool   `json:"confirm,omitempty" jsonschema:"false o assente mostra cosa verrebbe generato; true esegue"`
	QuoteTitle    string `json:"quote_title,omitempty" jsonschema:"titolo del preventivo, a scopo informativo: compare nella richiesta di autorizzazione così chi approva vede per quale preventivo sta generando un collegamento pubblico"`
}

func registerCRMWrites(server *mcp.Server, deps Deps) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_customer",
		Description: "Crea un cliente. Senza confirm mostra i campi senza creare nulla.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in customerFieldsInput) (*mcp.CallToolResult, any, error) {
		fields := fieldsOf(in, "customer_id", "confirm")
		if !in.Confirm {
			return nil, textOutput{Text: preview.NewResource(fields)}, nil
		}
		raw, err := deps.Client.Do(ctx, "POST", "/api/customers", fields)
		if err != nil {
			return nil, nil, err
		}
		return dataResult("Creato.", raw)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "update_customer",
		Description: "Modifica un cliente. Senza confirm mostra la differenza senza scrivere. " +
			"Attenzione: contact_emails sostituisce l'elenco completo delle email di contatto.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in customerFieldsInput) (*mcp.CallToolResult, any, error) {
		fields := fieldsOf(in, "customer_id", "confirm")
		path := fmt.Sprintf("/api/customers/%d", in.CustomerID)

		currentRaw, err := deps.Client.Do(ctx, "GET", path, nil)
		if err != nil {
			return nil, nil, err
		}
		if !in.Confirm {
			return nil, textOutput{Text: preview.Diff(decodeMap(currentRaw), fields)}, nil
		}
		raw, err := deps.Client.Do(ctx, "PATCH", path, fields)
		if err != nil {
			return nil, nil, err
		}
		return dataResult("Scritto.", raw)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_quote",
		Description: "Crea un preventivo. Richiede titolo e cliente. Senza confirm mostra i campi senza creare nulla.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in quoteFieldsInput) (*mcp.CallToolResult, any, error) {
		if in.Title == "" || in.CustomerID == 0 {
			return nil, nil, fmt.Errorf("title e customer_id sono obbligatori per creare un preventivo")
		}
		fields := fieldsOf(in, "quote_id", "confirm")
		if !in.Confirm {
			return nil, textOutput{Text: preview.NewResource(fields)}, nil
		}
		raw, err := deps.Client.Do(ctx, "POST", "/api/quotes", fields)
		if err != nil {
			return nil, nil, err
		}
		return dataResult("Creato.", raw)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "update_quote",
		Description: "Modifica un preventivo. Senza confirm mostra la differenza senza scrivere. " +
			"A preventivo chiuso Orchestrator rifiuta ogni modifica.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in quoteFieldsInput) (*mcp.CallToolResult, any, error) {
		fields := fieldsOf(in, "quote_id", "confirm")
		path := fmt.Sprintf("/api/quotes/%d", in.QuoteID)

		currentRaw, err := deps.Client.Do(ctx, "GET", path, nil)
		if err != nil {
			return nil, nil, err
		}
		if !in.Confirm {
			return nil, textOutput{Text: preview.Diff(decodeMap(currentRaw), fields)}, nil
		}
		raw, err := deps.Client.Do(ctx, "PATCH", path, fields)
		if err != nil {
			return nil, nil, err
		}
		return dataResult("Scritto.", raw)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "delete_quote",
		Description: "Elimina un preventivo. Operazione non annullabile: non va mai autorizzata in automatico. " +
			"Riporta in quote_title il titolo letto dall'anteprima, così chi approva vede quale preventivo sta eliminando.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in deleteQuoteInput) (*mcp.CallToolResult, any, error) {
		path := fmt.Sprintf("/api/quotes/%d", in.QuoteID)
		currentRaw, err := deps.Client.Do(ctx, "GET", path, nil)
		if err != nil {
			return nil, nil, err
		}
		current := decodeMap(currentRaw)
		title, _ := current["title"].(string)

		if !in.Confirm {
			notice := ""
			if isIrreversible("delete_quote") {
				notice = irreversibleNotice("delete_quote", in.QuoteID, title)
			}
			return nil, textOutput{Text: fmt.Sprintf(
				"ANTEPRIMA — nulla è stato eliminato\n%s\n\nPer eseguire richiama con confirm: true e quote_title: %q.",
				notice, title)}, nil
		}
		raw, err := deps.Client.Do(ctx, "DELETE", path, nil)
		if err != nil {
			return nil, nil, err
		}
		return dataResult("Eliminato.", raw)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "attach_product_to_quote",
		Description: "Associa un prodotto a un preventivo con una quantità. Usa recurring: true per i prodotti ricorrenti: " +
			"in quel caso product_id va preso da list_recurring_products, non da list_products — sono due cataloghi distinti con identificatori in spazi diversi.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in quoteProductInput) (*mcp.CallToolResult, any, error) {
		if in.Quantity <= 0 {
			return nil, nil, fmt.Errorf("quantity è obbligatoria e deve essere maggiore di zero")
		}
		path := quoteProductPath(in.QuoteID, in.ProductID, in.Recurring)
		if !in.Confirm {
			return nil, textOutput{Text: fmt.Sprintf(
				"ANTEPRIMA — nulla è stato scritto\n  al preventivo %d verrebbe associato il prodotto %d in quantità %d\n\nPer applicare, richiama con confirm: true.",
				in.QuoteID, in.ProductID, in.Quantity)}, nil
		}
		raw, err := deps.Client.Do(ctx, "POST", path, map[string]any{"quantity": in.Quantity})
		if err != nil {
			return nil, nil, err
		}
		return dataResult("Associato.", raw)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "detach_product_from_quote",
		Description: "Toglie un prodotto da un preventivo. Usa recurring: true per i prodotti ricorrenti: " +
			"in quel caso product_id va preso da list_recurring_products, non da list_products — sono due cataloghi distinti con identificatori in spazi diversi.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in quoteProductInput) (*mcp.CallToolResult, any, error) {
		path := quoteProductPath(in.QuoteID, in.ProductID, in.Recurring)
		if !in.Confirm {
			return nil, textOutput{Text: fmt.Sprintf(
				"ANTEPRIMA — nulla è stato scritto\n  dal preventivo %d verrebbe tolto il prodotto %d\n\nPer applicare, richiama con confirm: true.",
				in.QuoteID, in.ProductID)}, nil
		}
		raw, err := deps.Client.Do(ctx, "DELETE", path, nil)
		if err != nil {
			return nil, nil, err
		}
		return dataResult("Rimosso.", raw)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "create_quote_pdf_link",
		Description: "Genera un collegamento pubblico al PDF del preventivo, pensato per essere inviato al cliente. " +
			"Il collegamento non richiede autenticazione, dura fino a 90 giorni e non può essere revocato prima della scadenza: " +
			"non va mai autorizzato in automatico. Riporta in quote_title il titolo letto dall'anteprima.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in pdfLinkInput) (*mcp.CallToolResult, any, error) {
		path := fmt.Sprintf("/api/quotes/%d", in.QuoteID)
		currentRaw, err := deps.Client.Do(ctx, "GET", path, nil)
		if err != nil {
			return nil, nil, err
		}
		current := decodeMap(currentRaw)
		title, _ := current["title"].(string)

		days := in.ExpiresInDays
		if days == 0 {
			days = 30
		}
		if days > 90 {
			return nil, nil, fmt.Errorf("expires_in_days non può superare 90")
		}

		if !in.Confirm {
			notice := ""
			if isIrreversible("create_quote_pdf_link") {
				notice = irreversibleNotice("create_quote_pdf_link", in.QuoteID, title) + "\n"
			}
			return nil, textOutput{Text: fmt.Sprintf(
				"ANTEPRIMA — nessun collegamento è stato generato\n%s  durata: %d giorni\n\n"+
					"Per generarlo richiama con confirm: true e quote_title: %q.",
				notice, days, title)}, nil
		}

		body := map[string]any{"expires_in_days": days}
		if in.Lang != "" {
			body["lang"] = in.Lang
		}
		raw, err := deps.Client.Do(ctx, "POST", fmt.Sprintf("/api/quotes/%d/pdf-link", in.QuoteID), body)
		if err != nil {
			return nil, nil, err
		}
		return dataResult("Collegamento generato.", raw)
	})
}

func quoteProductPath(quoteID, productID int, recurring bool) string {
	if recurring {
		return fmt.Sprintf("/api/quotes/%d/recurring-products/%d", quoteID, productID)
	}
	return fmt.Sprintf("/api/quotes/%d/products/%d", quoteID, productID)
}
