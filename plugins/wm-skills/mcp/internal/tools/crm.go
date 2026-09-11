package tools

import (
	"context"
	"fmt"
	"net/url"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Il gruppo crm espone lettura e scrittura. Le scritture seguono la regola
// generale dell'anteprima con confirm; due di esse hanno effetti che non si
// annullano e richiedono in più una frase esatta ricavata dai dati veri.
func crmToolNames() []string {
	return []string{
		"list_customers", "get_customer", "create_customer", "update_customer",
		"list_quotes", "get_quote", "create_quote", "update_quote", "delete_quote",
		"attach_product_to_quote", "detach_product_from_quote",
		"create_quote_pdf_link",
		"list_products", "list_recurring_products",
	}
}

// isIrreversible marca i tool il cui effetto non si annulla: l'eliminazione di
// un preventivo, e il collegamento pubblico, che resta valido fino a 90 giorni
// e non può essere revocato prima della scadenza.
//
// Il server non può impedire queste operazioni all'agente: qualunque parametro
// ricavabile dai dati è soddisfacibile da chi quei dati li ha appena letti.
// L'unica decisione umana è l'autorizzazione richiesta dal programma, e questi
// tool non vanno mai messi fra quelli approvati in automatico. Ciò che il
// server può fare è rendere quella decisione informata.
func isIrreversible(tool string) bool {
	return tool == "delete_quote" || tool == "create_quote_pdf_link"
}

// irreversibleNotice compone l'avviso mostrato in anteprima, nominando il
// preventivo colpito invece del solo identificatore.
func irreversibleNotice(tool string, quoteID int, title string) string {
	return fmt.Sprintf("⚠️  %s non si può annullare — preventivo %d: %q", tool, quoteID, title)
}

type listCustomersInput struct {
	Search string `json:"search,omitempty" jsonschema:"ricerca sul nome del cliente"`
	Status string `json:"status,omitempty" jsonschema:"filtro sullo stato del cliente"`
}

type getCustomerInput struct {
	CustomerID int `json:"customer_id" jsonschema:"identificatore del cliente"`
}

type listQuotesInput struct {
	CustomerID int    `json:"customer_id,omitempty" jsonschema:"filtra i preventivi di un cliente"`
	Status     string `json:"status,omitempty" jsonschema:"filtro sullo stato del preventivo"`
}

type getQuoteInput struct {
	QuoteID int    `json:"quote_id" jsonschema:"identificatore del preventivo"`
	Include string `json:"include,omitempty" jsonschema:"relazioni da espandere nella risposta"`
}

// listCustomersPath, listQuotesPath e getQuotePath codificano i filtri con
// net/url invece di concatenarli grezzi nella query string: le ragioni
// sociali dei clienti contengono caratteri come "&" ("Rossi & Figli"), che una
// concatenazione grezza spezzerebbe in un parametro diverso da quello voluto.
func listCustomersPath(search, status string) string {
	path := "/api/customers"
	q := url.Values{}
	if search != "" {
		q.Set("search", search)
	}
	if status != "" {
		q.Set("status", status)
	}
	if len(q) > 0 {
		path += "?" + q.Encode()
	}
	return path
}

func listQuotesPath(customerID int, status string) string {
	path := "/api/quotes"
	q := url.Values{}
	if customerID != 0 {
		q.Set("customer_id", fmt.Sprintf("%d", customerID))
	}
	if status != "" {
		q.Set("status", status)
	}
	if len(q) > 0 {
		path += "?" + q.Encode()
	}
	return path
}

func getQuotePath(quoteID int, include string) string {
	path := fmt.Sprintf("/api/quotes/%d", quoteID)
	if include != "" {
		q := url.Values{}
		q.Set("include", include)
		path += "?" + q.Encode()
	}
	return path
}

func registerCRM(server *mcp.Server, deps Deps) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_customers",
		Description: "Elenca i clienti, con filtri facoltativi su nome e stato.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in listCustomersInput) (*mcp.CallToolResult, any, error) {
		path := listCustomersPath(in.Search, in.Status)
		raw, err := deps.Client.Do(ctx, "GET", path, nil)
		if err != nil {
			return nil, nil, err
		}
		return nil, jsonOrText(raw), nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_customer",
		Description: "Legge un cliente.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in getCustomerInput) (*mcp.CallToolResult, any, error) {
		raw, err := deps.Client.Do(ctx, "GET", fmt.Sprintf("/api/customers/%d", in.CustomerID), nil)
		if err != nil {
			return nil, nil, err
		}
		return nil, jsonOrText(raw), nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_quotes",
		Description: "Elenca i preventivi, con filtri facoltativi su cliente e stato.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in listQuotesInput) (*mcp.CallToolResult, any, error) {
		path := listQuotesPath(in.CustomerID, in.Status)
		raw, err := deps.Client.Do(ctx, "GET", path, nil)
		if err != nil {
			return nil, nil, err
		}
		return nil, jsonOrText(raw), nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_quote",
		Description: "Legge un preventivo, con la possibilità di espandere le relazioni.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in getQuoteInput) (*mcp.CallToolResult, any, error) {
		path := getQuotePath(in.QuoteID, in.Include)
		raw, err := deps.Client.Do(ctx, "GET", path, nil)
		if err != nil {
			return nil, nil, err
		}
		return nil, jsonOrText(raw), nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_products",
		Description: "Elenca i prodotti ordinari a catalogo. Gli identificatori di questo elenco valgono solo per prodotti non ricorrenti: per un prodotto ricorrente usa list_recurring_products, sono due cataloghi distinti con identificatori in spazi diversi.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
		raw, err := deps.Client.Do(ctx, "GET", "/api/products", nil)
		if err != nil {
			return nil, nil, err
		}
		return nil, jsonOrText(raw), nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_recurring_products",
		Description: "Elenca i prodotti ricorrenti a catalogo. Cataloghi distinto da list_products: l'identificatore di un prodotto ricorrente va preso qui, mai da list_products, per attach_product_to_quote/detach_product_from_quote con recurring: true.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
		raw, err := deps.Client.Do(ctx, "GET", "/api/recurring-products", nil)
		if err != nil {
			return nil, nil, err
		}
		return nil, jsonOrText(raw), nil
	})

	registerCRMWrites(server, deps)
}
