package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const envTSURL = "https://raw.githubusercontent.com/webmappsrl/wm-types/main/src/environment.ts"

// Da wm-core EnvironmentService: stesse espressioni, stesso ordine di controllo.
var (
	reWmpackages      = regexp.MustCompile(`^(\d+)\.([a-zA-Z0-9-]+)(?:\.mobile)?(?:\.[^.]+)+$`)
	vecchiSottodomini = map[string]bool{"app": true, "geohub": true, "mobile": true}
	shardVecchi       = map[string]bool{"geohub": true, "geohubdev": true}

	reVoce  = regexp.MustCompile(`(?ms)^\s{2}'?([\w.-]+)'?:\s*\{(.*?)^\s{2}\},?`)
	reCampo = regexp.MustCompile(`(\w+):\s*(?:'([^']*)'|(\d+))`)
)

// Ambiente è la tabella degli shard e la mappa redirects di wm-types.
type Ambiente struct {
	Shards          map[string]map[string]any `json:"shards"`
	Redirects       map[string]map[string]any `json:"redirects"`
	OrdineRedirects []string                  `json:"-"`
	Fonte           string                    `json:"-"`
	CopiatoIl       string                    `json:"copiato_il"`
}

// Risoluzione è ciò che `risolvi` stampa e che `scarica` usa per gli indirizzi.
type Risoluzione struct {
	Shard      string `json:"shard"`
	AppID      int64  `json:"app_id"`
	Origin     string `json:"origin"`
	AwsAPI     string `json:"awsApi"`
	ElasticAPI string `json:"elasticApi"`
	Layout     string `json:"layout"`
	Fonte      string `json:"fonte"`
}

// blocco: il corpo di `export const <nome>...= { ... };`, o "" se manca.
func blocco(testo, nome string) (string, bool) {
	re := regexp.MustCompile(`(?s)export const ` + nome + `\b[^=]*=\s*\{(.*?)\n\};`)
	m := re.FindStringSubmatch(testo)
	if m == nil {
		return "", false
	}
	return m[1], true
}

// voci: le voci `chiave: { campo: valore, ... },` di un blocco, nell'ordine del file.
func voci(corpo string) (map[string]map[string]any, []string) {
	tutte := map[string]map[string]any{}
	var ordine []string
	for _, m := range reVoce.FindAllStringSubmatch(corpo, -1) {
		campi := map[string]any{}
		for _, c := range reCampo.FindAllStringSubmatch(m[2], -1) {
			if c[3] != "" {
				n, _ := strconv.ParseInt(c[3], 10, 64)
				campi[c[1]] = float64(n)
			} else {
				campi[c[1]] = c[2]
			}
		}
		if _, gia := tutte[m[1]]; !gia {
			ordine = append(ordine, m[1])
		}
		tutte[m[1]] = campi
	}
	return tutte, ordine
}

func leggiEnvironment(testo string) (*Ambiente, error) {
	shards, ok1 := blocco(testo, "shards")
	redirects, ok2 := blocco(testo, "redirects")
	if !ok1 || !ok2 {
		return nil, errore(2, "environment.ts di wm-types ha una struttura che lo script non riconosce: "+
			"manca il blocco `export const shards` o `export const redirects`.")
	}
	tabella, _ := voci(shards)
	if len(tabella) == 0 {
		return nil, errore(2, "environment.ts di wm-types: nessuno shard letto dal blocco `shards`.")
	}
	r, ordine := voci(redirects)
	return &Ambiente{Shards: tabella, Redirects: r, OrdineRedirects: ordine}, nil
}

// cartellaSkill: i file della skill, accanto alla cartella bin/ in cui sta il binario.
func cartellaSkill() string {
	exe, err := os.Executable()
	if err != nil {
		return filepath.Join("skills", "wm-geohub-import-check")
	}
	if reale, err := filepath.EvalSymlinks(exe); err == nil {
		exe = reale
	}
	return filepath.Join(filepath.Dir(filepath.Dir(exe)), "skills", "wm-geohub-import-check")
}

// caricaEnvironment: da GitHub; se non risponde, dalla copia locale shards.json, dichiarandolo.
func caricaEnvironment() (*Ambiente, error) {
	indirizzo := os.Getenv("GIC_ENV_TS_URL")
	if indirizzo == "" {
		indirizzo = envTSURL
	}
	dati, err := leggiURL(indirizzo)
	if err != nil {
		copia := os.Getenv("GIC_SHARDS_FALLBACK")
		if copia == "" {
			copia = filepath.Join(cartellaSkill(), "shards.json")
		}
		grezzo, errCopia := os.ReadFile(copia)
		if errCopia != nil {
			return nil, errore(2, "%s non ha risposto (%v) e la copia locale %s non si legge: %v",
				indirizzo, err, copia, errCopia)
		}
		var amb Ambiente
		if errJSON := json.Unmarshal(grezzo, &amb); errJSON != nil {
			return nil, errore(2, "La copia locale %s non è JSON valido: %v", copia, errJSON)
		}
		for k := range amb.Redirects {
			amb.OrdineRedirects = append(amb.OrdineRedirects, k)
		}
		ordinaPerTesto(amb.OrdineRedirects)
		amb.Fonte = fmt.Sprintf("⚠️ copia locale del %s (%s): %s non ha risposto (%v)",
			amb.CopiatoIl, filepath.Base(copia), indirizzo, err)
		return &amb, nil
	}
	amb, err := leggiEnvironment(string(dati))
	if err != nil {
		return nil, err
	}
	amb.Fonte = indirizzo
	return amb, nil
}

func testoDi(m map[string]any, k string) string {
	s, _ := m[k].(string)
	return s
}

// risolvi: URL del frontend → shard, id dell'app e indirizzi, con la regola di wm-core.
func risolvi(indirizzo string, amb *Ambiente) (*Risoluzione, error) {
	grezzo := indirizzo
	if !strings.Contains(grezzo, "//") {
		grezzo = "https://" + grezzo
	}
	host := ""
	if u, err := url.Parse(grezzo); err == nil {
		host = strings.ToLower(u.Hostname())
	}
	var shard string
	var appID int64
	var redirect map[string]any
	for _, k := range amb.OrdineRedirects {
		if strings.Contains(host, k) {
			redirect = amb.Redirects[k]
			break
		}
	}
	m := reWmpackages.FindStringSubmatch(host)
	switch {
	case redirect != nil:
		shard = testoDi(redirect, "shardName")
		appID, _ = intero(redirect["appId"]).(int64)
	case m != nil && !vecchiSottodomini[m[2]]:
		shard = m[2]
		appID, _ = strconv.ParseInt(m[1], 10, 64)
	default:
		primo := strings.Split(host, ".")[0]
		n, err := strconv.ParseInt(primo, 10, 64)
		if primo == "" || strings.Trim(primo, "0123456789") != "" || err != nil {
			return nil, errore(2, "Dall'indirizzo %s non si ricava l'id dell'app.", indirizzo)
		}
		shard, appID = "geohub", n
	}
	s, ok := amb.Shards[shard]
	if !ok {
		return nil, errore(2, "Lo shard «%s» (da %s) non è nella tabella degli shard di wm-types (%s): "+
			"va aggiunto lì, o in shards.json se si usa la copia locale.", shard, indirizzo, amb.Fonte)
	}
	layout := "nuovo"
	if shardVecchi[shard] {
		layout = "vecchio"
	}
	return &Risoluzione{
		Shard: shard, AppID: appID,
		Origin:     strings.TrimRight(testoDi(s, "origin"), "/"),
		AwsAPI:     strings.TrimRight(testoDi(s, "awsApi"), "/"),
		ElasticAPI: strings.TrimRight(testoDi(s, "elasticApi"), "/"),
		Layout:     layout, Fonte: amb.Fonte,
	}, nil
}
