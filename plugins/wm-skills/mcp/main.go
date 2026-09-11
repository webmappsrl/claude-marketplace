// Comando orchestrator-mcp: espone l'API di Orchestrator come tool MCP.
// Comunica sui canali standard del processo, senza aprire porte in ascolto.
package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/webmappsrl/claude-marketplace/plugins/wm-skills/mcp/internal/auth"
	"github.com/webmappsrl/claude-marketplace/plugins/wm-skills/mcp/internal/client"
	"github.com/webmappsrl/claude-marketplace/plugins/wm-skills/mcp/internal/enums"
	"github.com/webmappsrl/claude-marketplace/plugins/wm-skills/mcp/internal/tools"
	"github.com/webmappsrl/claude-marketplace/plugins/wm-skills/mcp/internal/version"
)

func main() {
	// I messaggi diagnostici vanno sul canale degli errori: quello di uscita
	// è riservato al dialogo con Claude Code.
	log.SetOutput(os.Stderr)
	log.SetPrefix("orchestrator-mcp: ")

	// L'indirizzo arriva solo dalla riga di comando: senza argomento è la
	// produzione. Il plugin distribuito non passa nulla; il server di sviluppo
	// dichiarato nel repo claude-marketplace passa l'istanza locale.
	baseURL := flag.String("base-url", "", "indirizzo di Orchestrator; vuoto significa produzione")
	// Anche il file delle credenziali è specifico dell'ambiente: la produzione
	// e un'istanza locale non condividono lo stesso segno di riconoscimento,
	// quindi non possono condividere lo stesso file.
	authFile := flag.String("auth-file", auth.DefaultPath(), "percorso del file delle credenziali Orchestrator")
	flag.Parse()

	ctx := context.Background()
	c := client.NewWithAuthPath(*baseURL, *authFile)

	// Gli elenchi di valori ammessi si leggono dagli enum PHP, non da una
	// specifica OpenAPI: il server non ne scarica né ne legge nessuna, e non
	// dipende dalla sua disponibilità per avviarsi. Se gli enum non si
	// riescono a leggere, il server parte lo stesso ma lo dice.
	allowed, err := enums.LoadOrFetch(ctx, enums.CachePath())
	if err != nil {
		log.Printf("attenzione: elenchi dei valori ammessi non disponibili (%v): type e status non saranno verificati prima dell'invio", err)
	}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "orchestrator",
		Version: version.Version,
	}, nil)

	groups := tools.ActiveGroups(os.Getenv("ORCHESTRATOR_MCP_GROUPS"))
	tools.Register(server, tools.Deps{Client: c, Enums: allowed}, groups)

	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		log.Fatalf("il server si è fermato: %v", err)
	}
}
