package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var (
	metriche = map[string]bool{"distance": true, "ascent": true, "descent": true,
		"duration_forward": true, "duration_backward": true, "ele_max": true, "ele_min": true,
		"ele_from": true, "ele_to": true}
	sempreAParte = map[string]bool{"feature_image": true, "image_gallery": true,
		"taxonomy.where": true, "taxonomy_where": true, "taxonomyWheres": true}
)

const maxEsempi = 2

// Gli spazi come li intende \s di Python sui testi Unicode.
const spazi = `\s\p{Z}\x{85}\x{1c}-\x{1f}`

var (
	reSpazi     = regexp.MustCompile(`[` + spazi + `]+`)
	reFraTag    = regexp.MustCompile(`>[` + spazi + `]+<`)
	reTagVuoto  = regexp.MustCompile(`<(\w+)([^<>]*?)[` + spazi + `]*></(\w+)>`)
	reNomeWhere = regexp.MustCompile(`[` + spazi + `-]+`)
)

// jsonInStringa: un valore JSON salvato come testo: "[]", "{...}", "false", "true", "null" (oc:8664).
func jsonInStringa(v any) bool {
	s, ok := v.(string)
	if !ok {
		return false
	}
	t := strings.TrimSpace(s)
	if t == "false" || t == "true" || t == "null" {
		return true
	}
	if !strings.HasPrefix(t, "[") && !strings.HasPrefix(t, "{") {
		return false
	}
	var x any
	if json.Unmarshal([]byte(s), &x) != nil {
		return false
	}
	switch x.(type) {
	case []any, map[string]any:
		return true
	}
	return false
}

// numero: il valore come numero, se è un numero o un testo che contiene solo un numero.
func numero(v any) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(x), 64)
		return f, err == nil
	}
	return 0, false
}

// svg: un SVG a meno di spazi, a capo, tipo di apici e di tag vuoti scritti chiusi (<g></g> = <g/>).
func svg(v string) string {
	v = reSpazi.ReplaceAllString(strings.ReplaceAll(v, `"`, "'"), " ")
	v = strings.TrimSpace(reFraTag.ReplaceAllString(v, "><"))
	return reTagVuoto.ReplaceAllStringFunc(v, func(t string) string {
		m := reTagVuoto.FindStringSubmatch(t)
		if m[1] != m[3] {
			return t
		}
		return "<" + m[1] + m[2] + "/>"
	})
}

func identifierDi(x any) any {
	return comeMappa(x)["identifier"]
}

// tipoGenerico: POI senza tipo su Geohub e con il solo tipo generico «poi» sullo shard.
func tipoGenerico(g, s map[string]any) bool {
	taxg := comeMappa(g["taxonomy"])
	haTipoG := false
	if tg, ok := taxg["poi_type"].(map[string]any); ok && (vero(tg["identifier"]) || vero(tg["name"])) {
		haTipoG = true
	}
	if vero(taxg["poi_types"]) {
		haTipoG = true
	}
	for _, x := range comeLista(g["taxonomyIdentifiers"]) {
		if strings.HasPrefix(strPy(x), "poi_type_") {
			haTipoG = true
		}
	}
	ts := comeMappa(s["taxonomy"])
	soloPoi := identifierDi(ts["poi_type"]) == "poi"
	for _, x := range comeLista(ts["poi_types"]) {
		if m, ok := x.(map[string]any); ok && m["identifier"] != "poi" {
			soloPoi = false
		}
	}
	var tipiS []string
	for _, x := range comeLista(s["taxonomyIdentifiers"]) {
		if strings.HasPrefix(strPy(x), "poi_type_") {
			tipiS = append(tipiS, strPy(x))
		}
	}
	soloGenerico := soloPoi || (len(tipiS) == 1 && tipiS[0] == "poi_type_poi")
	return !haTipoG && soloGenerico
}

// senzaTipoGenerico: copia del POI dello shard senza il tipo generico, per confrontare il resto.
func senzaTipoGenerico(s map[string]any) map[string]any {
	s = copia(s).(map[string]any)
	if tax, ok := s["taxonomy"].(map[string]any); ok {
		delete(tax, "poi_type")
		delete(tax, "poi_types")
	}
	if ids, ok := s["taxonomyIdentifiers"]; ok {
		resto := []any{}
		for _, x := range comeLista(ids) {
			if x != "poi_type_poi" {
				resto = append(resto, x)
			}
		}
		s["taxonomyIdentifiers"] = resto
	}
	return s
}

type perIdentifier struct {
	chiavi []string
	valori map[string]any
	id     map[string]any
}

// perIdentifierDi: identifier → elemento, se ogni elemento è un oggetto con identifier.
func perIdentifierDi(lista []any) *perIdentifier {
	if len(lista) == 0 {
		return nil
	}
	p := &perIdentifier{valori: map[string]any{}, id: map[string]any{}}
	for _, x := range lista {
		m, ok := x.(map[string]any)
		if !ok || !vero(m["identifier"]) {
			return nil
		}
		k := chiave(m["identifier"])
		if _, gia := p.valori[k]; !gia {
			p.chiavi = append(p.chiavi, k)
		}
		p.valori[k], p.id[k] = m, m["identifier"]
	}
	sort.Strings(p.chiavi)
	return p
}

// lunghezzaRune e taglia contano i caratteri, non i byte, come Python.
func taglia(s string, n int) (string, bool) {
	r := []rune(s)
	if len(r) <= n {
		return s, false
	}
	return string(r[:n]), true
}

func mostra(v any) any {
	switch x := v.(type) {
	case string:
		if strings.Contains(x, "<svg") {
			return fmt.Sprintf("<svg… (%d caratteri)>", len([]rune(x)))
		}
		if t, tagliato := taglia(x, 200); tagliato {
			return t + "…"
		}
		return x
	case map[string]any, []any:
		t, tagliato := taglia(dumpPy(x), 200)
		if tagliato {
			return t + "…"
		}
		return t
	}
	return v
}

// --- regole -----------------------------------------------------------------------------------

type Regole struct {
	versione        any
	codiceValore    map[string]string
	soloShard       map[string]bool
	nonUsate        map[string]bool
	soloShardAttesi map[string]bool
	contatori       map[string]int
}

func caricaRegole(percorso string) (*Regole, error) {
	grezzo, err := os.ReadFile(percorso)
	if err != nil {
		return nil, errore(2, "Non riesco a leggere le regole %s: %v", percorso, err)
	}
	var dati struct {
		Versione any `json:"versione"`
		Regole   []struct {
			Codice string   `json:"codice"`
			Tipo   string   `json:"tipo"`
			Chiavi []string `json:"chiavi"`
		} `json:"regole"`
	}
	if err := json.Unmarshal(grezzo, &dati); err != nil || dati.Regole == nil {
		return nil, errore(2, "Le regole %s non sono JSON valido o non hanno l'elenco `regole`.", percorso)
	}
	r := &Regole{versione: dati.Versione, codiceValore: map[string]string{}, soloShard: map[string]bool{},
		nonUsate: map[string]bool{}, soloShardAttesi: map[string]bool{}, contatori: map[string]int{}}
	for _, regola := range dati.Regole {
		for _, c := range regola.Chiavi {
			switch regola.Tipo {
			case "valore_diverso_atteso":
				r.codiceValore[c] = regola.Codice
			case "chiave_solo_shard":
				r.soloShard[c] = true
			case "chiave_non_usata":
				r.nonUsate[c] = true
			case "valore_solo_shard":
				r.soloShardAttesi[c] = true
			}
		}
	}
	return r, nil
}

// nonUsata: chiave che il frontend non legge: vale ovunque o solo per una risorsa («elastic:doc»).
func (r *Regole) nonUsata(risorsa, k string) bool {
	return r.nonUsate[k] || r.nonUsate[risorsa+":"+k]
}

func (r *Regole) conta(codice string, n int) { r.contatori[codice] += n }

// valoreDiversoAtteso: la regola vale sul percorso intero o sull'ultima chiave (senza indici).
func (r *Regole) valoreDiversoAtteso(campo string) string {
	parti := strings.Split(campo, ".")
	for _, c := range []string{campo, parti[len(parti)-1]} {
		if codice, ok := r.codiceValore[c]; ok {
			return codice
		}
	}
	return ""
}

// --- confronto --------------------------------------------------------------------------------

type Esempio struct {
	GeohubID any `json:"geohub_id"`
	ShardID  any `json:"shard_id"`
	Geohub   any `json:"geohub"`
	Shard    any `json:"shard"`
}

type Gruppo struct {
	Codice         string    `json:"codice"`
	Risorsa        string    `json:"risorsa"`
	Campo          string    `json:"campo"`
	Tipo           string    `json:"tipo"`
	Casi           int       `json:"casi"`
	Esempi         []Esempio `json:"esempi"`
	PossibileCausa *string   `json:"possibile_causa"`
	SuGeohub       int       `json:"su_geohub,omitempty"`
}

type Segnalata struct {
	Regola  string    `json:"regola"`
	Risorsa string    `json:"risorsa"`
	Campo   string    `json:"campo"`
	Casi    int       `json:"casi"`
	Esempi  []Esempio `json:"esempi"`
}

type SoloGeohub struct {
	Risorsa  string `json:"risorsa"`
	GeohubID any    `json:"geohub_id"`
	Nome     any    `json:"nome"`
}

type SoloShard struct {
	Risorsa  string `json:"risorsa"`
	GeohubID any    `json:"geohub_id"`
	ShardID  any    `json:"shard_id"`
	Nome     any    `json:"nome"`
}

type ChiaveDaUnaParte struct {
	Risorsa string `json:"risorsa"`
	Chiave  string `json:"chiave"`
	Parte   string `json:"parte"`
	Feature int    `json:"feature"`
}

type Ignorata struct {
	Regola string `json:"regola"`
	Casi   int    `json:"casi"`
}

type Diff struct {
	RegoleVersione   any                `json:"regole_versione"`
	Gruppi           []*Gruppo          `json:"gruppi"`
	SoloGeohub       []SoloGeohub       `json:"solo_geohub"`
	SoloShard        []SoloShard        `json:"solo_shard"`
	ChiaviDaUnaParte []ChiaveDaUnaParte `json:"chiavi_da_una_parte"`
	Segnalate        []*Segnalata       `json:"segnalate"`
	Ignorate         []Ignorata         `json:"ignorate"`
	NonScaricate     []NonScaricata     `json:"non_scaricate"`
}

type Confronto struct {
	regole     *Regole
	gruppi     map[[3]string]*Gruppo
	segnalate  map[[2]string]*Segnalata
	ordineSeg  [][2]string
	soloGeohub []SoloGeohub
	soloShard  []SoloShard
	chiavi     []ChiaveDaUnaParte
	presenti   map[[2]string]int // (risorsa, campo) → feature che su Geohub hanno un valore
}

func nuovoConfronto(r *Regole) *Confronto {
	return &Confronto{regole: r, gruppi: map[[3]string]*Gruppo{}, segnalate: map[[2]string]*Segnalata{},
		soloGeohub: []SoloGeohub{}, soloShard: []SoloShard{}, chiavi: []ChiaveDaUnaParte{},
		presenti: map[[2]string]int{}}
}

func esempio(gid, sid, g, s any) Esempio {
	return Esempio{intero(gid), intero(sid), mostra(g), mostra(s)}
}

// segnala: differenza attesa per regola, che però va mostrata: non arriva a wm-tag.
func (c *Confronto) segnala(regola, risorsa, campo string, gid, sid, g, s any) {
	k := [2]string{regola, risorsa}
	sg, ok := c.segnalate[k]
	if !ok {
		sg = &Segnalata{Regola: regola, Risorsa: risorsa, Campo: campo, Esempi: []Esempio{}}
		c.segnalate[k] = sg
		c.ordineSeg = append(c.ordineSeg, k)
	}
	sg.Casi++
	if len(sg.Esempi) < maxEsempi {
		sg.Esempi = append(sg.Esempi, esempio(gid, sid, g, s))
	}
}

func (c *Confronto) differenza(risorsa, campo, tipo string, gid, sid, g, s any) *Gruppo {
	k := [3]string{risorsa, campo, tipo}
	gr, ok := c.gruppi[k]
	if !ok {
		gr = &Gruppo{Codice: risorsa + ":" + campo + ":" + tipo, Risorsa: risorsa, Campo: campo,
			Tipo: tipo, Esempi: []Esempio{}}
		c.gruppi[k] = gr
	}
	gr.Casi++
	if len(gr.Esempi) < maxEsempi {
		gr.Esempi = append(gr.Esempi, esempio(gid, sid, g, s))
	}
	return gr
}

func chiaviOrdinate(g, s map[string]any) []string {
	tutte := map[string]bool{}
	for k := range g {
		tutte[k] = true
	}
	for k := range s {
		tutte[k] = true
	}
	l := make([]string, 0, len(tutte))
	for k := range tutte {
		l = append(l, k)
	}
	sort.Strings(l)
	return l
}

func ultimaChiave(campo string) string {
	parti := strings.Split(campo, ".")
	return parti[len(parti)-1]
}

// valori: confronto ricorsivo: presenza, tipo e valore. Le regole normalizzano, non escludono.
func (c *Confronto) valori(risorsa, campo string, g, s, gid, sid any, elastic bool) {
	if sempreAParte[campo] {
		return
	}
	ultima := strings.TrimRight(ultimaChiave(campo), "[]")
	vg, vs := vuoto(g), vuoto(s)
	if c.regole.nonUsata(risorsa, ultima) {
		if !(vg && vs) {
			c.regole.conta("chiavi_non_usate", 1)
		}
		return
	}
	if vg && !vs && c.regole.soloShardAttesi[ultima] {
		c.regole.conta("valori_solo_shard", 1)
		return
	}
	if vg && vs {
		return
	}
	if !vg {
		c.presenti[[2]string{risorsa, campo}]++
	}
	if ultimaChiave(campo) == "related_pois" {
		c.poiCollegati(risorsa, comeLista(g), comeLista(s), gid, sid)
		return
	}
	if vg != vs {
		if vs {
			c.differenza(risorsa, campo, "assente", gid, sid, g, s)
		} else if jsonInStringa(s) {
			// Un oggetto serializzato come testo al posto di niente, come "[]" (oc:8664).
			c.differenza(risorsa, campo, "tipo_diverso", gid, sid, g, s)
		}
		return
	}
	if codice := c.regole.valoreDiversoAtteso(campo); codice != "" {
		lg, okg := g.([]any)
		ls, oks := s.([]any)
		if okg && oks && len(lg) != len(ls) {
			c.differenza(risorsa, campo, "valore", gid, sid, float64(len(lg)), float64(len(ls)))
		} else {
			c.regole.conta(codice, 1)
		}
		return
	}
	mg, gDict := g.(map[string]any)
	ms, sDict := s.(map[string]any)
	if elastic && gDict != sDict {
		o, t := mg, s
		if !gDict {
			o, t = ms, g
		}
		if ts, ok := t.(string); ok && len(o) == 1 {
			for _, unico := range o {
				if unico == ts {
					c.regole.conta("elastic_nome_una_lingua", 1)
					return
				}
			}
		}
	}
	if ng, ok := numero(g); ok {
		if ns, ok2 := numero(s); ok2 && ng == ns {
			if !uguali(g, s) {
				c.regole.conta("numeri_come_testo", 1)
			}
			return
		}
	}
	sg, gStr := g.(string)
	ss, sStr := s.(string)
	if gStr && sStr && strings.Contains(sg, "<svg") && strings.Contains(ss, "<svg") {
		if svg(sg) == svg(ss) {
			c.regole.conta("svg_spazi", 1)
		} else {
			c.differenza(risorsa, campo, "valore", gid, sid, g, s)
		}
		return
	}
	lg, gList := g.([]any)
	ls, _ := s.([]any)
	switch {
	case tipo(g) != tipo(s):
		c.differenza(risorsa, campo, "tipo_diverso", gid, sid, g, s)
	case gList && perIdentifierDi(lg) != nil && perIdentifierDi(ls) != nil:
		// Liste di tassonomie: si abbinano per identifier, non per posizione.
		pg, ps := perIdentifierDi(lg), perIdentifierDi(ls)
		for _, k := range pg.chiavi {
			if _, ok := ps.valori[k]; !ok {
				c.differenza(risorsa, campo+"[]", "assente", gid, sid, pg.id[k], nil)
			}
		}
		for _, k := range pg.chiavi {
			if b, ok := ps.valori[k]; ok {
				c.valori(risorsa, campo+"[]", pg.valori[k], b, gid, sid, elastic)
			}
		}
	case gDict:
		for _, k := range chiaviOrdinate(mg, ms) {
			figlio := k
			if campo != "" {
				figlio = campo + "." + k
			}
			c.valori(risorsa, figlio, mg[k], ms[k], gid, sid, elastic)
		}
	case gList:
		if len(lg) != len(ls) {
			c.differenza(risorsa, campo+"[]", "valore", gid, sid, float64(len(lg)), float64(len(ls)))
		}
		for i := 0; i < len(lg) && i < len(ls); i++ {
			c.valori(risorsa, campo+"[]", lg[i], ls[i], gid, sid, elastic)
		}
	case gStr && strings.HasPrefix(sg, "http") && strings.HasPrefix(ss, "http"):
		c.regole.conta("prefissi_url", 1)
	case !uguali(g, s):
		c.differenza(risorsa, campo, "valore", gid, sid, g, s)
	}
}

// uguali confronta due valori letti da JSON.
func uguali(a, b any) bool {
	return tipo(a) == tipo(b) && dumpPy(a) == dumpPy(b)
}

// poiCollegati: POI collegati a una traccia: stessi POI, nello stesso ordine, e col nome. Il resto
// dei loro campi si confronta nel pois.geojson.
func (c *Confronto) poiCollegati(risorsa string, g, s []any, gid, sid any) {
	props := func(l []any) []map[string]any {
		var r []map[string]any
		for _, p := range l {
			m := comeMappa(p)
			if pr := comeMappa(m["properties"]); len(pr) > 0 {
				m = pr
			}
			r = append(r, m)
		}
		return r
	}
	pg, ps := props(g), props(s)
	ig, is := []any{}, []any{}
	for _, p := range pg {
		ig = append(ig, intero(p["id"]))
	}
	for _, p := range ps {
		is = append(is, intero(p["geohub_id"]))
	}
	if dumpPy(ig) != dumpPy(is) {
		c.differenza(risorsa, "related_pois", "valore", gid, sid, ig, is)
	}
	for i := 0; i < len(pg) && i < len(ps); i++ {
		if !vuoto(pg[i]["name"]) && vuoto(ps[i]["name"]) {
			c.differenza(risorsa, "related_pois[].name", "assente", gid, sid, pg[i]["name"], ps[i]["name"])
		}
	}
}

func lunghezza(v any) int {
	switch x := v.(type) {
	case []any:
		return len(x)
	case map[string]any:
		return len(x)
	case string:
		return len([]rune(x))
	}
	return 0
}

func (c *Confronto) immagini(risorsa string, g, s map[string]any, gid, sid any) {
	ha := func(p map[string]any) bool {
		if fi, ok := p["feature_image"].(map[string]any); ok {
			return vero(fi["url"])
		}
		return vero(p["feature_image"])
	}
	if ha(g) != ha(s) {
		c.differenza(risorsa, "feature_image", "conteggio_immagini", gid, sid, ha(g), ha(s))
	}
	ng, ns := lunghezza(g["image_gallery"]), lunghezza(s["image_gallery"])
	if ng != ns {
		c.differenza(risorsa, "image_gallery", "conteggio_immagini", gid, sid, float64(ng), float64(ns))
	}
}

// nomeWhere: nome di una where per il confronto: «Massa Carrara» e «Massa-Carrara» sono lo stesso.
func nomeWhere(n any) string {
	if s, ok := n.(string); ok {
		return strings.ToLower(strings.TrimSpace(reNomeWhere.ReplaceAllString(s, " ")))
	}
	return "\x00" + dumpPy(n)
}

// where: ogni where di Geohub deve comparire, per nome italiano, fra le where dello shard.
// `origine` è la feature da cui leggere le where OSMFeatures, se `s` non le ha (Elastic).
func (c *Confronto) where(risorsa string, nomiG []any, s map[string]any, gid, sid any, origine map[string]any, conOrigine bool) {
	mappa := func(x map[string]any) map[string]any {
		return comeMappa(x["taxonomy_where"])
	}
	tw := mappa(s)
	nomiS := map[string]bool{}
	for _, v := range tw {
		if m, ok := v.(map[string]any); ok {
			nomiS[nomeWhere(m["it"])] = true
		}
	}
	for _, w := range comeLista(s["taxonomyWheres"]) {
		if ws, ok := w.(string); ok {
			nomiS[nomeWhere(ws)] = true
		}
	}
	conOSM := tw
	if conOrigine {
		conOSM = mappa(origine)
	}
	calcolate := false
	for _, v := range conOSM {
		if m, ok := v.(map[string]any); ok && m["_source"] == "osmfeatures" {
			calcolate = true
		}
	}
	nomiGeohub := map[string]bool{}
	for _, n := range nomiG {
		nomiGeohub[nomeWhere(n)] = true
		if !vero(n) || nomiS[nomeWhere(n)] {
			continue
		}
		if calcolate {
			// Sullo shard le where vengono dalla geometria: una where scritta a mano su Geohub
			// che la geometria non tocca è attesa, ma si mostra (regola where_geometria).
			c.segnala("where_geometria", risorsa, "taxonomy_where", gid, sid, n, nil)
			continue
		}
		gr := c.differenza(risorsa, "taxonomy_where", "assente", gid, sid, n, nil)
		causa := "sincronizzazione delle taxonomy where non fatta o incompleta " +
			"(«Sincronizza Taxonomy Where su EC Features»)"
		gr.PossibileCausa = &causa
	}
	extra := 0
	for n := range nomiS {
		if !nomiGeohub[n] {
			extra++
		}
	}
	if extra > 0 {
		c.regole.conta("where_osmfeatures", extra)
	}
}

// chiaviPresenti: chiavi presenti da una parte sola, una volta per chiave. Restituisce quelle
// presenti solo sullo shard.
func (c *Confronto) chiaviPresenti(risorsa string, propsG, propsS []map[string]any) map[string]bool {
	insieme := func(props []map[string]any) map[string]bool {
		k := map[string]bool{}
		for _, p := range props {
			for chiave := range p {
				if !c.regole.nonUsata(risorsa, chiave) {
					k[chiave] = true
				}
			}
		}
		return k
	}
	valorizzate := func(props []map[string]any, k string) int {
		n := 0
		for _, p := range props {
			if !vuoto(p[k]) {
				n++
			}
		}
		return n
	}
	kg, ks := insieme(propsG), insieme(propsS)
	for _, k := range differenzaOrdinata(kg, ks) {
		c.chiavi = append(c.chiavi, ChiaveDaUnaParte{risorsa, k, "geohub", valorizzate(propsG, k)})
	}
	soloShard := map[string]bool{}
	for _, k := range differenzaOrdinata(ks, kg) {
		soloShard[k] = true
		if c.regole.soloShard[k] {
			c.regole.conta("chiavi_solo_shard", 1)
		} else if n := valorizzate(propsS, k); n == 0 {
			c.regole.conta("chiavi_vuote", 1) // campo presente ma vuoto su tutte le feature
		} else {
			c.chiavi = append(c.chiavi, ChiaveDaUnaParte{risorsa, k, "shard", n})
		}
	}
	return soloShard
}

func differenzaOrdinata(a, b map[string]bool) []string {
	var l []string
	for k := range a {
		if !b[k] {
			l = append(l, k)
		}
	}
	sort.Strings(l)
	return l
}

func nomiWhere(dirG string, ids any) []any {
	nomi := []any{}
	for _, i := range comeLista(ids) {
		d := comeMappa(leggiJSON(filepath.Join(dirG, "where", strPy(i)+".json")))
		n := d["name"]
		if m, ok := n.(map[string]any); ok {
			n = m["it"]
		}
		nomi = append(nomi, n)
	}
	return nomi
}

// abbinate: feature indicizzate per id normalizzato, con l'id com'era.
type abbinate struct {
	id      map[string]any
	feature map[string]map[string]any
}

func (a abbinate) chiavi() map[string]bool {
	k := map[string]bool{}
	for c := range a.feature {
		k[c] = true
	}
	return k
}

func abbina(features []map[string]any, campo string, conSenza bool) (abbinate, []map[string]any) {
	a := abbinate{id: map[string]any{}, feature: map[string]map[string]any{}}
	var senza []map[string]any
	for _, f := range features {
		v := intero(f[campo])
		if conSenza && v == nil {
			senza = append(senza, f)
			continue
		}
		a.id[chiave(v)], a.feature[chiave(v)] = v, f
	}
	return a, senza
}

func intersezioneOrdinata(a, b map[string]bool) []string {
	var l []string
	for k := range a {
		if b[k] {
			l = append(l, k)
		}
	}
	sort.Strings(l)
	return l
}

func proprietaDeiFile(file []string) []map[string]any {
	l := []map[string]any{}
	for _, t := range file {
		l = append(l, comeMappa(comeMappa(leggiJSON(t))["properties"]))
	}
	return l
}

func confronta(cartella, fileRegole string) (*Diff, *Manifest, error) {
	dg, ds := filepath.Join(cartella, "geohub"), filepath.Join(cartella, "shard")
	regole, err := caricaRegole(fileRegole)
	if err != nil {
		return nil, nil, err
	}
	c := nuovoConfronto(regole)
	manifest := &Manifest{NonScaricate: []NonScaricata{}}
	if grezzo, err := os.ReadFile(filepath.Join(cartella, "manifest.json")); err == nil {
		_ = json.Unmarshal(grezzo, manifest)
		if manifest.NonScaricate == nil {
			manifest.NonScaricate = []NonScaricata{}
		}
	}

	// config e icone: un solo oggetto per parte.
	for _, nome := range []string{"config.json", "icons.json"} {
		g, s := leggiJSON(filepath.Join(dg, nome)), leggiJSON(filepath.Join(ds, nome))
		if g != nil && s != nil {
			c.valori(nome, "", g, s, nil, nil, false)
		}
	}

	// POI e tracce: abbinati per id Geohub, confrontati campo per campo.
	proprietaPOI := func(percorso string) []map[string]any {
		l := []map[string]any{}
		for _, f := range featureDi(percorso) {
			l = append(l, comeMappa(f["properties"]))
		}
		return l
	}
	risorse := []struct {
		nome   string
		pg, ps []map[string]any
	}{
		{"pois.geojson", proprietaPOI(filepath.Join(dg, "pois.geojson")), proprietaPOI(filepath.Join(ds, "pois.geojson"))},
		{"tracks", proprietaDeiFile(fileTracce(dg)), proprietaDeiFile(fileTracce(ds))},
	}
	for _, r := range risorse {
		soloShard := c.chiaviPresenti(r.nome, r.pg, r.ps)
		g, _ := abbina(r.pg, "id", false)
		s, senza := abbina(r.ps, "geohub_id", true)
		kg, ks := g.chiavi(), s.chiavi()
		for _, k := range differenzaOrdinata(kg, ks) {
			c.soloGeohub = append(c.soloGeohub, SoloGeohub{r.nome, g.id[k], mostra(g.feature[k]["name"])})
		}
		for _, k := range differenzaOrdinata(ks, kg) {
			c.soloShard = append(c.soloShard, SoloShard{r.nome, s.id[k], intero(s.feature[k]["id"]),
				mostra(s.feature[k]["name"])})
		}
		for _, f := range senza {
			c.soloShard = append(c.soloShard, SoloShard{r.nome, nil, intero(f["id"]), mostra(f["name"])})
		}
		for _, k := range intersezioneOrdinata(kg, ks) {
			gid := g.id[k]
			a, b := g.feature[k], s.feature[k]
			sid := b["id"]
			if r.nome == "pois.geojson" && tipoGenerico(a, b) {
				// Su Geohub il POI non ha un tipo, lo shard gli assegna quello generico
				// «Punto di interesse» (wm-package EcPoi.php:276-280): atteso, da segnalare.
				c.segnala("poi_type_generico", r.nome, "taxonomy.poi_type", gid, sid, nil,
					comeMappa(comeMappa(b["taxonomy"])["poi_type"])["name"])
				b = senzaTipoGenerico(b)
			}
			for _, k := range chiaviOrdinate(a, b) {
				if !soloShard[k] {
					c.valori(r.nome, k, a[k], b[k], gid, sid, false)
				} else if jsonInStringa(b[k]) && !regole.soloShardAttesi[k] {
					// Chiave che su Geohub non c'è, ma sullo shard è un oggetto scritto come testo.
					c.differenza(r.nome, k, "tipo_diverso", gid, sid, a[k], b[k])
				}
			}
			c.immagini(r.nome, a, b, gid, sid)
			c.where(r.nome, nomiWhere(dg, comeMappa(a["taxonomy"])["where"]), b, gid, sid, nil, false)
		}
	}

	// Elastic: l'elenco delle tracce; sullo shard l'abbinamento passa dal file della traccia.
	eg, es := hits(leggiJSON(filepath.Join(dg, "elastic.json"))), hits(leggiJSON(filepath.Join(ds, "elastic.json")))
	fileS := map[string]map[string]any{}
	for _, p := range proprietaDeiFile(fileTracce(ds)) {
		fileS[chiave(p["id"])] = p
	}
	if len(eg) > 0 && len(es) == 0 {
		gr := c.differenza("elastic", "hits", "valore", nil, nil, float64(len(eg)), float64(0))
		gr.Codice = "elastic_vuoto"
		causa := "«Reindicizza Scout» non eseguito dopo l'import"
		gr.PossibileCausa = &causa
	} else {
		g, _ := abbina(eg, "id", false)
		s := abbinate{id: map[string]any{}, feature: map[string]map[string]any{}}
		for _, h := range es {
			gid := intero(fileS[chiave(h["id"])]["geohub_id"])
			if gid != nil { // traccia non scaricata: né diversa né uguale
				s.id[chiave(gid)], s.feature[chiave(gid)] = gid, h
			}
		}
		kg, ks := g.chiavi(), s.chiavi()
		for _, k := range differenzaOrdinata(kg, ks) {
			mancante := false
			for _, x := range manifest.NonScaricate {
				if strings.Contains(x.URL, "tracks/"+strPy(g.id[k])+".json") {
					mancante = true
				}
			}
			if !mancante {
				c.soloGeohub = append(c.soloGeohub, SoloGeohub{"elastic", g.id[k], mostra(g.feature[k]["name"])})
			}
		}
		for _, k := range intersezioneOrdinata(kg, ks) {
			gid := g.id[k]
			a, b := g.feature[k], s.feature[k]
			for _, campo := range chiaviOrdinate(a, b) {
				if regole.soloShard[campo] {
					continue
				}
				c.valori("elastic", campo, a[campo], b[campo], gid, b["id"], true)
			}
			nomi := []any{}
			for _, w := range comeLista(a["taxonomyWheres"]) {
				if _, ok := w.(string); ok {
					nomi = append(nomi, w)
				}
			}
			origine := fileS[chiave(b["id"])]
			if origine == nil {
				origine = map[string]any{}
			}
			c.where("elastic", nomi, b, gid, b["id"], origine, true)
		}
	}

	// Possibili cause dei passi saltati: indicazione, la differenza resta.
	zero := func(v any) bool {
		switch x := v.(type) {
		case nil:
			return true
		case float64:
			return x == 0
		case int64:
			return x == 0
		case bool:
			return !x
		}
		return false
	}
	for k, gr := range c.gruppi {
		risorsa, campo := k[0], k[1]
		if !metriche[campo] || len(gr.Esempi) == 0 {
			continue
		}
		tuttiZero := true
		for _, e := range gr.Esempi {
			tuttiZero = tuttiZero && zero(e.Shard)
		}
		if !tuttiZero {
			continue
		}
		nelFile := risorsa == "elastic"
		for _, e := range gr.Esempi {
			nelFile = nelFile && !zero(fileS[chiave(e.ShardID)][campo])
		}
		causa := "«Process Track Data» sulle tracce, poi «Aggiorna Tracks su AWS»"
		if nelFile {
			causa = "«Reindicizza Scout» non rieseguito dopo «Process Track Data»: nel file della traccia il valore c'è"
		}
		gr.PossibileCausa = &causa
	}

	// Campioni di immagini: gli originali devono coincidere.
	for _, cp := range manifest.CampioniImmagini {
		if cp.GeohubFile == "" || cp.ShardFile == "" {
			continue
		}
		bg, errG := os.ReadFile(filepath.Join(cartella, cp.GeohubFile))
		bs, errS := os.ReadFile(filepath.Join(cartella, cp.ShardFile))
		if errG == nil && errS == nil && sha256.Sum256(bg) != sha256.Sum256(bs) {
			c.differenza("immagini", cp.Tipo, "immagine_diversa", cp.GeohubID, cp.ShardID, cp.GeohubURL, cp.ShardURL)
		}
	}

	gruppi := make([]*Gruppo, 0, len(c.gruppi))
	for k, gr := range c.gruppi {
		if n, ok := c.presenti[[2]string{k[0], k[1]}]; ok && k[2] == "assente" {
			gr.SuGeohub = n
		}
		gruppi = append(gruppi, gr)
	}
	sort.SliceStable(gruppi, func(i, j int) bool {
		if gruppi[i].Casi != gruppi[j].Casi {
			return gruppi[i].Casi > gruppi[j].Casi
		}
		return gruppi[i].Codice < gruppi[j].Codice
	})
	segnalate := []*Segnalata{}
	for _, k := range c.ordineSeg {
		segnalate = append(segnalate, c.segnalate[k])
	}
	nomiRegole := make([]string, 0, len(regole.contatori))
	for k := range regole.contatori {
		nomiRegole = append(nomiRegole, k)
	}
	sort.Strings(nomiRegole)
	ignorate := []Ignorata{}
	for _, k := range nomiRegole {
		ignorate = append(ignorate, Ignorata{k, regole.contatori[k]})
	}
	diff := &Diff{RegoleVersione: regole.versione, Gruppi: gruppi, SoloGeohub: c.soloGeohub,
		SoloShard: c.soloShard, ChiaviDaUnaParte: c.chiavi, Segnalate: segnalate,
		Ignorate: ignorate, NonScaricate: manifest.NonScaricate}
	if err := scriviJSON(filepath.Join(cartella, "diff.json"), diff); err != nil {
		return nil, nil, errore(2, "Non riesco a scrivere diff.json in %s: %v", cartella, err)
	}
	if err := os.WriteFile(filepath.Join(cartella, "report.md"), []byte(report(diff, manifest)), 0o644); err != nil {
		return nil, nil, errore(2, "Non riesco a scrivere report.md in %s: %v", cartella, err)
	}
	return diff, manifest, nil
}
