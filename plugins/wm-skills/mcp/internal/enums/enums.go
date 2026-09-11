// Package enums legge gli elenchi di valori ammessi dai file sorgente degli
// enum PHP di Orchestrator. Non si usa la specifica OpenAPI perché non li
// espone: per type e status dichiara campi senza tipo e senza elenco.
package enums

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

// Variabili e non costanti: le prove le sostituiscono con un server locale.
var (
	storyTypeURL   = "https://raw.githubusercontent.com/webmappsrl/orchestrator/main/app/Enums/StoryType.php"
	storyStatusURL = "https://raw.githubusercontent.com/webmappsrl/orchestrator/main/app/Enums/StoryStatus.php"
)

// Values raccoglie gli elenchi letti.
type Values struct {
	StoryType   []string `json:"story_type"`
	StoryStatus []string `json:"story_status"`
}

// caseDecl riconosce qualunque riga dichiarativa di un case, indipendentemente
// dal fatto che il valore sia riconoscibile o no: serve solo a contare quante
// dichiarazioni ci sono nel sorgente, per accorgersi di un'estrazione parziale.
var caseDecl = regexp.MustCompile(`(?m)^\s*case\s+\w+\s*=`)

// caseLine intercetta le righe "case Nome = 'valore';" o "case Nome = "valore";":
// si prende il valore fra apici, non il nome del case, perché i due possono
// differire — "case Helpdesk = 'Help desk'" ne è l'esempio. Accetta sia apici
// singoli sia doppi perché il sorgente PHP non è vincolato a uno dei due.
var caseLine = regexp.MustCompile(`(?m)^\s*case\s+\w+\s*=\s*(?:'([^']*)'|"([^"]*)")\s*;`)

// ParsePHPEnum estrae i valori letterali dei case di un enum PHP di backing
// string. Se il sorgente non contiene alcuna dichiarazione "case ... =",
// restituisce nil senza errore (contenuto non riconoscibile come enum). Se
// invece contiene dichiarazioni ma alcune non vengono riconosciute
// dall'espressione regolare (per esempio un valore non tra apici), è un
// fallimento rumoroso: un elenco incompleto spacciato per completo farebbe
// respingere come inesistente un valore che invece c'è.
func ParsePHPEnum(raw []byte) ([]string, error) {
	declCount := len(caseDecl.FindAll(raw, -1))
	if declCount == 0 {
		return nil, nil
	}

	matches := caseLine.FindAllSubmatch(raw, -1)
	if len(matches) != declCount {
		return nil, fmt.Errorf(
			"elenco enum incompleto: %d righe case su %d non riconosciute dall'espressione regolare",
			declCount-len(matches), declCount,
		)
	}

	values := make([]string, 0, len(matches))
	for _, m := range matches {
		if m[1] != nil {
			values = append(values, string(m[1]))
		} else {
			values = append(values, string(m[2]))
		}
	}
	return values, nil
}

func CachePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "orchestrator-enums.json"
	}
	return filepath.Join(home, ".cache", "webmapp", "orchestrator-enums.json")
}

// LoadOrFetch restituisce gli elenchi ammessi. Se una copia locale esiste ed
// è utilizzabile (almeno un campo valorizzato), viene restituita subito,
// senza attendere la rete: il server deve essere pronto immediatamente a ogni
// sessione. L'aggiornamento della copia locale avviene comunque, ma in un
// goroutine separato che non blocca il chiamante — il dato più fresco sarà
// disponibile dalla sessione successiva.
//
// Se non esiste alcuna copia locale utilizzabile, si scarica in modo
// sincrono (comportamento invariato): è l'unico caso in cui LoadOrFetch può
// bloccare l'avvio in attesa della rete.
func LoadOrFetch(ctx context.Context, cachePath string) (*Values, error) {
	if cached, err := readCache(cachePath); err == nil && (len(cached.StoryType) > 0 || len(cached.StoryStatus) > 0) {
		go refreshCacheInBackground(cachePath)
		return cached, nil
	}
	return fetchFresh(ctx, cachePath)
}

// refreshCacheInBackground riscarica gli enum e aggiorna la copia locale,
// senza che nessuno attenda il suo risultato. Usa un contesto proprio,
// indipendente da quello della chiamata che ha già ricevuto la risposta dalla
// cache: quel contesto potrebbe terminare (es. fine della richiesta MCP)
// prima che lo scarico finisca.
func refreshCacheInBackground(cachePath string) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_, _ = fetchFresh(ctx, cachePath)
}

// fetchFresh scarica i due enum; se uno dei due scarichi (o l'interpretazione
// del relativo sorgente) fallisce, usa per quel solo campo l'ultima copia
// locale salvata — il dato fresco ottenuto per l'altro campo non viene
// buttato via. La copia locale viene poi riscritta con la combinazione
// risultante. Restituisce errore solo se, alla fine, non resta alcun valore
// né per StoryType né per StoryStatus (niente di fresco e nessuna copia
// locale utilizzabile).
func fetchFresh(ctx context.Context, cachePath string) (*Values, error) {
	freshType, typeErr := fetchAndParse(ctx, storyTypeURL)
	freshStatus, statusErr := fetchAndParse(ctx, storyStatusURL)

	cached, cacheErr := readCache(cachePath)

	v := &Values{}
	if len(freshType) > 0 {
		v.StoryType = freshType
	} else if cached != nil {
		v.StoryType = cached.StoryType
	}
	if len(freshStatus) > 0 {
		v.StoryStatus = freshStatus
	} else if cached != nil {
		v.StoryStatus = cached.StoryStatus
	}

	if len(v.StoryType) == 0 && len(v.StoryStatus) == 0 {
		return nil, fmt.Errorf(
			"elenchi non scaricabili (type: %v; status: %v) e nessuna copia locale utilizzabile in %s (%v)",
			typeErr, statusErr, cachePath, cacheErr,
		)
	}

	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err == nil {
		if encoded, err := json.Marshal(v); err == nil {
			_ = os.WriteFile(cachePath, encoded, 0o644)
		}
	}
	return v, nil
}

// fetchAndParse scarica un enum e lo interpreta; qualunque errore (rete o
// interpretazione) è riportato al chiamante, che decide come ripiegare.
func fetchAndParse(ctx context.Context, url string) ([]string, error) {
	raw, err := fetch(ctx, url)
	if err != nil {
		return nil, err
	}
	values, err := ParsePHPEnum(raw)
	if err != nil {
		return nil, err
	}
	return values, nil
}

// readCache rilegge l'ultima copia locale salvata. Un file mancante,
// illeggibile o corrotto è un errore riportato al chiamante, mai un valore
// vuoto silenzioso.
func readCache(path string) (*Values, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var v Values
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

func fetch(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := (&http.Client{Timeout: 5 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("stato %d su %s", resp.StatusCode, url)
	}
	return io.ReadAll(resp.Body)
}
