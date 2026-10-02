package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	timeout   = 30 * time.Second
	parallelo = 8
	tentativi = 3 // il primo più due nuovi tentativi
)

var tipiImmagine = []string{"traccia_evidenza", "traccia_galleria", "poi_evidenza", "poi_galleria"}

// La verifica dei certificati resta sempre attiva: nessuna opzione la spegne.
var client = &http.Client{Timeout: timeout}

// erroreHTTP distingue una risposta del server (con il suo codice) da un errore di rete.
type erroreHTTP struct{ codice int }

func (e *erroreHTTP) Error() string { return fmt.Sprintf("HTTP %d", e.codice) }

// leggiURL: GET di un URL http(s) o file://.
func leggiURL(indirizzo string) ([]byte, error) {
	if strings.HasPrefix(indirizzo, "file://") {
		u, err := url.Parse(indirizzo)
		if err != nil {
			return nil, err
		}
		return os.ReadFile(u.Path)
	}
	risposta, err := client.Get(indirizzo)
	if err != nil {
		return nil, err
	}
	defer risposta.Body.Close()
	if risposta.StatusCode >= 400 {
		return nil, &erroreHTTP{risposta.StatusCode}
	}
	return io.ReadAll(risposta.Body)
}

// indirizzi dei file di un'app, secondo il layout dello shard (vecchio o nuovo).
func indirizzi(r *Risoluzione) [][2]string {
	a, i := r.AwsAPI, r.AppID
	var base [][2]string
	if r.Layout == "vecchio" {
		base = [][2]string{
			{"config.json", fmt.Sprintf("%s/conf/%d.json", a, i)},
			{"icons.json", fmt.Sprintf("%s/icons/%d.json", a, i)},
			{"pois.geojson", fmt.Sprintf("%s/pois/%d.geojson", a, i)},
		}
	} else {
		base = [][2]string{
			{"config.json", fmt.Sprintf("%s/%d/config.json", a, i)},
			{"icons.json", fmt.Sprintf("%s/%d/icons.json", a, i)},
			{"pois.geojson", fmt.Sprintf("%s/%d/pois.geojson", a, i)},
		}
	}
	return append(base, [2]string{"elastic.json", fmt.Sprintf("%s/?app=geohub_app_%d", r.ElasticAPI, i)})
}

// NonScaricata è una risorsa che non si è potuta leggere, con il motivo.
type NonScaricata struct {
	URL    string `json:"url"`
	Motivo string `json:"motivo"`
}

// Scaricatore: GET con tentativi; ogni fallimento finisce in nonScaricate, mai in un errore.
type Scaricatore struct {
	mu           sync.Mutex
	nonScaricate []NonScaricata
}

func (d *Scaricatore) prendi(indirizzo string) []byte {
	ultimo := ""
	for range tentativi {
		dati, err := leggiURL(indirizzo)
		if err == nil {
			return dati
		}
		if e, ok := err.(*erroreHTTP); ok {
			ultimo = e.Error()
			if e.codice == 404 {
				break
			}
		} else {
			ultimo = fmt.Sprintf("%T: %v", err, err)
		}
	}
	d.mu.Lock()
	d.nonScaricate = append(d.nonScaricate, NonScaricata{indirizzo, ultimo})
	d.mu.Unlock()
	return nil
}

func (d *Scaricatore) salva(indirizzo, percorso string) bool {
	dati := d.prendi(indirizzo)
	if dati == nil {
		return false
	}
	if os.MkdirAll(filepath.Dir(percorso), 0o755) != nil {
		return false
	}
	return os.WriteFile(percorso, dati, 0o644) == nil
}

// salvaTutti scarica le coppie (url, percorso) con al più `parallelo` richieste insieme.
func (d *Scaricatore) salvaTutti(coppie [][2]string) {
	var wg sync.WaitGroup
	posti := make(chan struct{}, parallelo)
	for _, c := range coppie {
		wg.Add(1)
		posti <- struct{}{}
		go func(c [2]string) {
			defer wg.Done()
			defer func() { <-posti }()
			d.salva(c[0], c[1])
		}(c)
	}
	wg.Wait()
}

// hits: le tracce di una risposta Elastic: lista sullo shard, {hits: [...]} o lista su Geohub.
func hits(elastic any) []map[string]any {
	var h any = elastic
	if m, ok := elastic.(map[string]any); ok {
		h = m["hits"]
	}
	if m, ok := h.(map[string]any); ok {
		h = m["hits"]
	}
	var tracce []map[string]any
	for _, x := range comeLista(h) {
		m := comeMappa(x)
		if src, ok := m["_source"]; ok {
			m = comeMappa(src)
		}
		tracce = append(tracce, m)
	}
	return tracce
}

// controllaApp: l'app dello shard deve avere properties.geohub_id uguale all'id Geohub.
func controllaApp(d *Scaricatore, rs *Risoluzione, geohubID int64, out string) error {
	percorso := filepath.Join(out, "shard", "app-all.json")
	if !d.salva(rs.Origin+"/api/v2/app/all", percorso) {
		return errore(2, "Non riesco a leggere %s/api/v2/app/all per controllare che l'app %d dello "+
			"shard venga dall'app Geohub %d.", rs.Origin, rs.AppID, geohubID)
	}
	var trovato any
	for _, a := range comeLista(leggiJSON(percorso)) {
		app := comeMappa(a)
		if id, ok := app["id"].(float64); ok && int64(id) == rs.AppID {
			trovato = comeMappa(app["properties"])["geohub_id"]
			break
		}
	}
	if strPy(trovato) != fmt.Sprint(geohubID) {
		return errore(3, "L'app %d dello shard %s ha geohub_id %s, non %d: le due app non sono la "+
			"stessa, il confronto non parte.", rs.AppID, rs.Shard, strPy(trovato), geohubID)
	}
	return nil
}

// proprieta: le properties di una feature GeoJSON, o la feature stessa se non le ha.
func proprieta(f map[string]any) map[string]any {
	if p, ok := f["properties"]; ok {
		return comeMappa(p)
	}
	return f
}

func primaImmagine(feature map[string]any, tipo string) any {
	p := proprieta(feature)
	if strings.HasSuffix(tipo, "evidenza") {
		if fi, ok := p["feature_image"].(map[string]any); ok {
			return fi["url"]
		}
		return nil
	}
	gal := comeLista(p["image_gallery"])
	if len(gal) > 0 {
		if g, ok := gal[0].(map[string]any); ok {
			return g["url"]
		}
	}
	return nil
}

// Campione è un'immagine originale da confrontare per contenuto, una per tipo.
type Campione struct {
	Tipo       string `json:"tipo"`
	GeohubID   any    `json:"geohub_id"`
	ShardID    any    `json:"shard_id"`
	GeohubURL  any    `json:"geohub_url"`
	ShardURL   any    `json:"shard_url"`
	GeohubFile string `json:"geohub_file,omitempty"`
	ShardFile  string `json:"shard_file,omitempty"`
}

type featurePerID struct {
	id      any
	feature map[string]any
}

func featureDi(percorso string) []map[string]any {
	var fs []map[string]any
	for _, f := range comeLista(comeMappa(leggiJSON(percorso))["features"]) {
		fs = append(fs, comeMappa(f))
	}
	return fs
}

func fileTracce(cartella string) []string {
	file, _ := filepath.Glob(filepath.Join(cartella, "tracks", "*.json"))
	sort.Strings(file)
	return file
}

// campioniImmagini: il primo originale per tipo su Geohub, con quello della feature
// corrispondente sullo shard.
func campioniImmagini(out string) []Campione {
	g, s := filepath.Join(out, "geohub"), filepath.Join(out, "shard")
	gpoi := map[string]featurePerID{}
	for _, f := range featureDi(filepath.Join(g, "pois.geojson")) {
		id := comeMappa(f["properties"])["id"]
		gpoi[strPy(id)] = featurePerID{id, f}
	}
	spoi := map[string]map[string]any{}
	for _, f := range featureDi(filepath.Join(s, "pois.geojson")) {
		spoi[strPy(comeMappa(f["properties"])["geohub_id"])] = f
	}
	gtr := map[string]featurePerID{}
	for _, t := range fileTracce(g) {
		id := strings.TrimSuffix(filepath.Base(t), ".json")
		gtr[id] = featurePerID{id, comeMappa(leggiJSON(t))}
	}
	str := map[string]map[string]any{}
	for _, t := range fileTracce(s) {
		f := comeMappa(leggiJSON(t))
		str[strPy(comeMappa(f["properties"])["geohub_id"])] = f
	}
	campioni := []Campione{}
	for _, tipo := range tipiImmagine {
		sorgente, altra := gpoi, spoi
		if strings.HasPrefix(tipo, "traccia") {
			sorgente, altra = gtr, str
		}
		chiavi := make([]string, 0, len(sorgente))
		for k := range sorgente {
			chiavi = append(chiavi, k)
		}
		sort.Strings(chiavi)
		for _, k := range chiavi {
			gu := primaImmagine(sorgente[k].feature, tipo)
			su := primaImmagine(altra[k], tipo)
			if vero(gu) && vero(su) {
				campioni = append(campioni, Campione{Tipo: tipo, GeohubID: sorgente[k].id,
					ShardID: proprieta(altra[k])["id"], GeohubURL: gu, ShardURL: su})
				break
			}
		}
	}
	return campioni
}

// Manifest descrive uno scaricamento: cosa, quando, cosa è mancato.
type Manifest struct {
	Geohub           *Risoluzione   `json:"geohub"`
	Shard            *Risoluzione   `json:"shard"`
	ScaricatoIl      string         `json:"scaricato_il"`
	NonScaricate     []NonScaricata `json:"non_scaricate"`
	CampioniImmagini []Campione     `json:"campioni_immagini"`
}

func scarica(urlGeohub, urlShard, out string, immagini bool) (*Manifest, error) {
	amb, err := caricaEnvironment()
	if err != nil {
		return nil, err
	}
	rg, err := risolvi(urlGeohub, amb)
	if err != nil {
		return nil, err
	}
	rs, err := risolvi(urlShard, amb)
	if err != nil {
		return nil, err
	}
	d := &Scaricatore{nonScaricate: []NonScaricata{}}
	if err := controllaApp(d, rs, rg.AppID, out); err != nil {
		return nil, err
	}
	for _, parte := range []struct {
		nome string
		r    *Risoluzione
	}{{"geohub", rg}, {"shard", rs}} {
		var coppie [][2]string
		for _, c := range indirizzi(parte.r) {
			coppie = append(coppie, [2]string{c[1], filepath.Join(out, parte.nome, c[0])})
		}
		d.salvaTutti(coppie)
		// Le tracce dell'app sono quelle di Elastic: sullo shard la cartella tracks/ è condivisa.
		coppie = nil
		for _, t := range hits(leggiJSON(filepath.Join(out, parte.nome, "elastic.json"))) {
			if id := t["id"]; vero(id) {
				coppie = append(coppie, [2]string{
					fmt.Sprintf("%s/tracks/%s.json", parte.r.AwsAPI, strPy(id)),
					filepath.Join(out, parte.nome, "tracks", strPy(id)+".json")})
			}
		}
		d.salvaTutti(coppie)
	}
	scaricaWhere(d, rg, out)
	campioni := []Campione{}
	if immagini {
		campioni = campioniImmagini(out)
	}
	for i := range campioni {
		c := &campioni[i]
		c.GeohubFile = "immagini/" + c.Tipo + "-geohub.bin"
		c.ShardFile = "immagini/" + c.Tipo + "-shard.bin"
		gu, _ := c.GeohubURL.(string)
		su, _ := c.ShardURL.(string)
		d.salva(gu, filepath.Join(out, c.GeohubFile))
		d.salva(su, filepath.Join(out, c.ShardFile))
	}
	m := &Manifest{Geohub: rg, Shard: rs,
		ScaricatoIl:  time.Now().UTC().Format("2006-01-02T15:04:05.000000-07:00"),
		NonScaricate: d.nonScaricate, CampioniImmagini: campioni}
	if err := scriviJSON(filepath.Join(out, "manifest.json"), m); err != nil {
		return nil, errore(2, "Non riesco a scrivere il manifest in %s: %v", out, err)
	}
	return m, nil
}

// scaricaWhere: i nomi delle where Geohub citate da tracce e POI, per confrontarle per nome.
func scaricaWhere(d *Scaricatore, rg *Risoluzione, out string) {
	var feature []map[string]any
	for _, t := range fileTracce(filepath.Join(out, "geohub")) {
		feature = append(feature, comeMappa(leggiJSON(t)))
	}
	feature = append(feature, featureDi(filepath.Join(out, "geohub", "pois.geojson"))...)
	where := map[string]bool{}
	for _, f := range feature {
		for _, w := range comeLista(comeMappa(comeMappa(f["properties"])["taxonomy"])["where"]) {
			where[strPy(w)] = true
		}
	}
	var wg sync.WaitGroup
	posti := make(chan struct{}, parallelo)
	for w := range where {
		wg.Add(1)
		posti <- struct{}{}
		go func(w string) {
			defer wg.Done()
			defer func() { <-posti }()
			indirizzo := fmt.Sprintf("%s/api/taxonomy/where/%s", rg.Origin, w)
			dati := d.prendi(indirizzo)
			if dati == nil {
				return
			}
			var v map[string]any
			if json.Unmarshal(dati, &v) != nil {
				d.mu.Lock()
				d.nonScaricate = append(d.nonScaricate, NonScaricata{indirizzo, "risposta non JSON"})
				d.mu.Unlock()
				return
			}
			p := filepath.Join(out, "geohub", "where", w+".json")
			if os.MkdirAll(filepath.Dir(p), 0o755) == nil {
				_ = os.WriteFile(p, []byte(dumpPy(map[string]any{"id": v["id"], "name": v["name"]})), 0o644)
			}
		}(w)
	}
	wg.Wait()
}
