// Package compose costruisce la description di un ticket quando si aggiunge
// informazione: un blocco in testa e brevi annotazioni dopo frasi esatte. Il
// testo esistente resta identico carattere per carattere, e non passa mai dal
// modello: l'agente fornisce solo ciò che aggiunge.
package compose

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
)

// Annotation è un testo da inserire subito dopo una frase esatta della
// description attuale. Occurrence sceglie quale presenza usare quando la frase
// compare più volte (1 = la prima); zero vuol dire «deve essere unica».
type Annotation struct {
	After      string `json:"after" jsonschema:"frase esatta, come compare nell'HTML della description attuale, dopo cui inserire text"`
	Text       string `json:"text" jsonschema:"HTML da inserire subito dopo la frase, preceduto da uno spazio"`
	Occurrence int    `json:"occurrence,omitempty" jsonschema:"quale presenza della frase usare se compare più volte (1 = la prima)"`
}

// Placed è un'annotazione inserita, con il blocco del risultato in cui è
// finita: serve all'anteprima per mostrare al dev dove cade ogni etichetta.
type Placed struct {
	Annotation Annotation
	Context    string
}

// Result è l'esito della composizione.
type Result struct {
	Text      string
	Changed   bool
	Prepended bool
	Placed    []Placed
	Skipped   []Annotation
}

type insertion struct {
	pos int
	ann Annotation
}

// Apply mette prepend in testa a current e inserisce ogni annotazione dopo la
// sua frase. Le frasi si cercano solo in current, mai in prepend. Se una
// frase manca, è ambigua, cade dentro un tag o coincide con il punto di
// un'altra annotazione, non compone nulla e restituisce l'errore: meglio
// fermarsi che scrivere un'etichetta nel posto sbagliato.
func Apply(current, prepend string, annotations []Annotation) (Result, error) {
	var res Result
	var ins []insertion
	used := map[int]bool{}

	// Se current inizia già con prepend (riapplicazione dopo una scrittura
	// riuscita), le frasi si cercano solo nella parte dopo il prepend: mai
	// dentro il prepend già scritto, anche se contiene la frase.
	searchIn := current
	searchOffset := 0
	if prepend != "" && strings.HasPrefix(current, prepend) {
		searchIn = current[len(prepend):]
		searchOffset = len(prepend)
	}

	for i, a := range annotations {
		n := i + 1
		if a.After == "" || a.Text == "" {
			return Result{}, fmt.Errorf("annotazione %d: after e text sono obbligatori", n)
		}
		positions := indexAll(searchIn, a.After)
		if len(positions) == 0 {
			return Result{}, fmt.Errorf("annotazione %d: la frase %q non compare nella description attuale", n, a.After)
		}
		occ := a.Occurrence
		if occ == 0 {
			if len(positions) > 1 {
				return Result{}, fmt.Errorf("annotazione %d: la frase %q compare %s: indica occurrence (1 = la prima)", n, a.After, timesLabel(len(positions)))
			}
			occ = 1
		}
		if occ < 1 || occ > len(positions) {
			return Result{}, fmt.Errorf("annotazione %d: occurrence %d non esiste, la frase %q compare %s", n, occ, a.After, timesLabel(len(positions)))
		}
		pos := searchOffset + positions[occ-1] + len(a.After)
		if insideTag(current, pos) {
			return Result{}, fmt.Errorf("annotazione %d: il punto dopo %q cade dentro un tag HTML", n, a.After)
		}
		if used[pos] {
			return Result{}, fmt.Errorf("annotazione %d: un'altra annotazione va nello stesso punto, dopo %q", n, a.After)
		}
		used[pos] = true
		if alreadyPresent(current, pos, a.Text) {
			res.Skipped = append(res.Skipped, a)
			continue
		}
		ins = append(ins, insertion{pos: pos, ann: a})
	}

	// Inserendo dal fondo, le posizioni calcolate sul testo originale restano
	// valide per tutte le annotazioni che precedono.
	sort.Slice(ins, func(i, j int) bool { return ins[i].pos > ins[j].pos })
	text := current
	for _, in := range ins {
		text = text[:in.pos] + " " + in.ann.Text + text[in.pos:]
	}

	res.Prepended = prepend != "" && !strings.HasPrefix(current, prepend)
	offset := 0
	if res.Prepended {
		text = prepend + text
		offset = len(prepend)
	}

	// Posizioni nel risultato, in ordine di testo: ogni inserimento precedente
	// sposta i successivi della propria lunghezza.
	sort.Slice(ins, func(i, j int) bool { return ins[i].pos < ins[j].pos })
	shift := offset
	for _, in := range ins {
		start := in.pos + shift
		end := start + 1 + len(in.ann.Text)
		res.Placed = append(res.Placed, Placed{Annotation: in.ann, Context: blockAround(text, start, end)})
		shift += 1 + len(in.ann.Text)
	}

	res.Text = text
	res.Changed = res.Prepended || len(ins) > 0
	return res, nil
}

// timesLabel scrive «1 volta» al singolare e «N volte» altrimenti, per i
// messaggi d'errore che contano quante volte una frase compare.
func timesLabel(n int) string {
	if n == 1 {
		return "1 volta"
	}
	return fmt.Sprintf("%d volte", n)
}

// indexAll restituisce l'inizio di ogni presenza di sub in s, anche
// sovrapposta.
func indexAll(s, sub string) []int {
	var out []int
	for from := 0; ; {
		i := strings.Index(s[from:], sub)
		if i < 0 {
			return out
		}
		out = append(out, from+i)
		from += i + 1
	}
}

// insideTag dice se la posizione pos di s cade fra un '<' e il suo '>'.
func insideTag(s string, pos int) bool {
	return strings.LastIndex(s[:pos], "<") > strings.LastIndex(s[:pos], ">")
}

// alreadyPresent dice se " "+text compare già subito dopo pos in s, come
// inserimento completo: dopo " "+text deve seguire la fine della stringa o un
// tag ('<'), non un carattere qualunque. Altrimenti un testo che comincia
// solo come text (es. "✅" seguito da "✅ Risolto...") verrebbe scambiato per
// presente.
func alreadyPresent(s string, pos int, text string) bool {
	marker := " " + text
	if !strings.HasPrefix(s[pos:], marker) {
		return false
	}
	rest := s[pos+len(marker):]
	return rest == "" || strings.HasPrefix(rest, "<")
}

var (
	blockOpen  = regexp.MustCompile(`(?i)<(li|p|h[1-6]|td|th)[\s>]`)
	blockClose = regexp.MustCompile(`(?i)</(li|p|h[1-6]|td|th)>`)
)

// blockAround restituisce il blocco HTML che contiene l'intervallo
// [start, end) di s: dall'ultima apertura di li/p/h1-h6/td/th prima di start
// alla prima chiusura dopo end. Senza blocchi, 120 byte per lato, spostati
// fuori da eventuali tag.
func blockAround(s string, start, end int) string {
	from := -1
	for _, loc := range blockOpen.FindAllStringIndex(s[:start], -1) {
		from = loc[0]
	}
	if from < 0 {
		from = max(0, start-120)
		from = nextRuneBoundary(s, from)
		if insideTag(s, from) {
			from += strings.Index(s[from:], ">") + 1
		}
	}
	to := len(s)
	if loc := blockClose.FindStringIndex(s[end:]); loc != nil {
		to = end + loc[1]
	} else if end+120 < len(s) {
		to = end + 120
		to = nextRuneBoundary(s, to)
		if insideTag(s, to) {
			to += strings.Index(s[to:], ">") + 1
		}
	}
	return s[from:to]
}

// nextRuneBoundary sposta i in avanti fino al successivo inizio di rune di s,
// per non tagliare un carattere multi-byte a metà.
func nextRuneBoundary(s string, i int) int {
	for i < len(s) && !utf8.RuneStart(s[i]) {
		i++
	}
	return i
}
