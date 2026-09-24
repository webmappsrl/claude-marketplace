# Aggiungere informazione alla description senza riscriverla — piano di implementazione

> **Per chi esegue:** SUB-SKILL RICHIESTA: usa superpowers:subagent-driven-development
> (consigliata) o superpowers:executing-plans per eseguire il piano task per task. I passi usano
> le caselle (`- [ ]`) per tenere traccia.

**Obiettivo:** `update_story` aggiunge informazione alla `description` di un ticket (blocco in
testa ed etichette dopo frasi esatte) componendo il testo nel codice, e l'anteprima di ogni tool
di scrittura mostra i campi per intero, resi come testo leggibile.

**Architettura:** un pacchetto nuovo `internal/compose` contiene la funzione pura che compone la
`description` (nessun I/O, tutta testabile). `internal/preview` smette di troncare, rende i campi
HTML come testo e avvisa quando si perde più di metà del testo. `update_story` sceglie fra
sostituzione (`description`) e aggiunta (`prepend`/`annotations`), chiama `compose.Apply` e manda
una sola PATCH. Le skill passano all'aggiunta e smettono di ricopiare la `description`.

**Tecnologie:** Go 1.27.1, `github.com/modelcontextprotocol/go-sdk` v1.7.0, libreria standard
(`html`, `regexp`, `strings`, `sort`). Nessuna dipendenza nuova.

**Spec:** [overview.md](overview.md)

## Vincoli globali

- Tutto il codice Go sta in `plugins/wm-skills/mcp/`; i test si lanciano da lì con `go test ./...`.
- Nessuna dipendenza nuova in `go.mod`.
- Nessuna scrittura su `https://orchestrator.maphub.it`: i test usano `httptest`; una prova a mano
  si fa solo su `http://localhost:8099`.
- Commenti, messaggi d'errore del tool e testi delle skill in italiano; termini tecnici in inglese.
- Il binario `plugins/wm-skills/bin/orchestrator-mcp` non si modifica a mano: si rigenera con
  `plugins/wm-skills/mcp/build.sh` (solo `darwin/arm64`).
- Prima di dichiarare pronto: `claude plugin validate .` e `./.github/scripts/verifica-diagramma.sh`
  dalla root del repo.
- **Nessun `git add`, `git commit`, `git push` né creazione di branch da parte di Claude.** I passi
  «Commit» di questo piano sono istruzioni per il dev, da eseguire solo dopo la sua conferma
  esplicita; lo scope è `nota-in-testa-description`.

## Punti da non lasciar scoperti

1. **`description` nulla nel ticket** (JSON `null`): va trattata come testo vuoto, e il `prepend`
   diventa tutta la `description`. Test in Task 1 (`TestApplyOnEmptyDescription`) e Task 3.
2. **Frase `after` che contiene caratteri accentati o emoji**: le posizioni sono in byte, il
   risultato deve restare UTF-8 valido. Test in Task 1 (`TestApplyKeepsUTF8`).
3. **Annotazione il cui `after` finisce esattamente su `>`** (per esempio `</strong>`): il punto di
   inserimento è fuori dal tag e va accettato. Test in Task 1 (`TestApplyAcceptsAfterEndingWithTag`).
4. **Entità HTML dentro un titolo** (`&amp;` in un `<h2>`): la resa leggibile deve decodificarle
   prima di mettere in maiuscolo. Test in Task 2 (`TestReadableDecodesEntitiesInHeadings`).
5. **Riprova dopo un timeout con solo lo status già scritto**: se blocco ed etichette sono già
   presenti ma lo status è diverso, si scrive solo lo status. Test in Task 3
   (`TestUpdateStoryAppendAlreadyPresentStillWritesOtherFields`).

---

### Task 1: funzione di composizione `compose.Apply`

**File:**
- Crea: `plugins/wm-skills/mcp/internal/compose/compose.go`
- Crea: `plugins/wm-skills/mcp/internal/compose/compose_test.go`

**Interfacce:**
- Consuma: niente.
- Produce:
  ```go
  type Annotation struct {
      After      string `json:"after"`
      Text       string `json:"text"`
      Occurrence int    `json:"occurrence,omitempty"`
  }
  type Placed struct {
      Annotation Annotation
      Context    string // HTML del blocco (li, p, h1-h6, td, th) in cui finisce Text, nel risultato
  }
  type Result struct {
      Text      string       // description composta
      Changed   bool         // true se c'è qualcosa di nuovo da scrivere
      Prepended bool         // true se prepend è stato aggiunto (false se vuoto o già in testa)
      Placed    []Placed     // annotazioni inserite, nell'ordine in cui compaiono nel testo
      Skipped   []Annotation // annotazioni già presenti subito dopo la loro frase
  }
  func Apply(current, prepend string, annotations []Annotation) (Result, error)
  ```

- [ ] **Passo 1: scrivi i test che falliscono**

`plugins/wm-skills/mcp/internal/compose/compose_test.go`:

```go
package compose

import (
	"strings"
	"testing"
	"unicode/utf8"
)

const twoCycles = `<h2>Primo ciclo — da fare</h2><ul><li>Bloccante 1: la traccia perde la quota</li><li>Bloccante 2: manca il test</li></ul>`

func TestApplyPrependsAndAnnotates(t *testing.T) {
	res, err := Apply(twoCycles, "<h2>Secondo ciclo — da fare</h2>", []Annotation{
		{After: "la traccia perde la quota", Text: "<strong>✅ Risolto (23/09)</strong>"},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := `<h2>Secondo ciclo — da fare</h2><h2>Primo ciclo — da fare</h2><ul><li>Bloccante 1: la traccia perde la quota <strong>✅ Risolto (23/09)</strong></li><li>Bloccante 2: manca il test</li></ul>`
	if res.Text != want {
		t.Fatalf("testo composto sbagliato:\n got %s\nwant %s", res.Text, want)
	}
	if !res.Changed || !res.Prepended || len(res.Placed) != 1 {
		t.Fatalf("esito sbagliato: %+v", res)
	}
	if res.Placed[0].Context != "<li>Bloccante 1: la traccia perde la quota <strong>✅ Risolto (23/09)</strong></li>" {
		t.Fatalf("contesto sbagliato: %q", res.Placed[0].Context)
	}
}

// Togliendo dal risultato il blocco in testa e ogni " "+Text inserito si deve
// riottenere la description di partenza carattere per carattere.
func TestApplyLeavesOriginalIntact(t *testing.T) {
	anns := []Annotation{
		{After: "la traccia perde la quota", Text: "<strong>✅</strong>"},
		{After: "manca il test", Text: "<strong>⚠️ in parte</strong>"},
	}
	prepend := "<h2>Secondo</h2>"
	res, err := Apply(twoCycles, prepend, anns)
	if err != nil {
		t.Fatal(err)
	}
	back := strings.TrimPrefix(res.Text, prepend)
	for _, a := range anns {
		back = strings.Replace(back, " "+a.Text, "", 1)
	}
	if back != twoCycles {
		t.Fatalf("il testo originale è cambiato:\n got %s\nwant %s", back, twoCycles)
	}
}

func TestApplyRejectsMissingPhrase(t *testing.T) {
	_, err := Apply(twoCycles, "", []Annotation{{After: "frase che non c'è", Text: "x"}})
	if err == nil || !strings.Contains(err.Error(), "non compare") {
		t.Fatalf("atteso errore «non compare», ottenuto %v", err)
	}
}

func TestApplyRejectsAmbiguousPhraseWithoutOccurrence(t *testing.T) {
	desc := "<p>manca il test</p><p>manca il test</p>"
	_, err := Apply(desc, "", []Annotation{{After: "manca il test", Text: "x"}})
	if err == nil || !strings.Contains(err.Error(), "compare 2 volte") {
		t.Fatalf("atteso errore «compare 2 volte», ottenuto %v", err)
	}
}

func TestApplyUsesOccurrence(t *testing.T) {
	desc := "<p>manca il test</p><p>manca il test</p>"
	res, err := Apply(desc, "", []Annotation{{After: "manca il test", Text: "✅", Occurrence: 2}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Text != "<p>manca il test</p><p>manca il test ✅</p>" {
		t.Fatalf("occurrence ignorata: %s", res.Text)
	}
}

func TestApplyRejectsOccurrenceOutOfRange(t *testing.T) {
	_, err := Apply("<p>a</p>", "", []Annotation{{After: "a", Text: "x", Occurrence: 3}})
	if err == nil || !strings.Contains(err.Error(), "occurrence 3") {
		t.Fatalf("atteso errore su occurrence, ottenuto %v", err)
	}
}

func TestApplyRejectsInsertionInsideTag(t *testing.T) {
	desc := `<a href="https://github.com/x/pull/1">PR</a>`
	_, err := Apply(desc, "", []Annotation{{After: "github.com/x", Text: "x"}})
	if err == nil || !strings.Contains(err.Error(), "dentro un tag") {
		t.Fatalf("atteso errore «dentro un tag», ottenuto %v", err)
	}
}

func TestApplyAcceptsAfterEndingWithTag(t *testing.T) {
	res, err := Apply("<li><strong>Bloccante</strong> resto</li>", "", []Annotation{{After: "<strong>Bloccante</strong>", Text: "✅"}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Text != "<li><strong>Bloccante</strong> ✅ resto</li>" {
		t.Fatalf("inserimento sbagliato: %s", res.Text)
	}
}

func TestApplyRejectsTwoAnnotationsAtSamePoint(t *testing.T) {
	_, err := Apply("<p>abc</p>", "", []Annotation{{After: "abc", Text: "1"}, {After: "abc", Text: "2"}})
	if err == nil || !strings.Contains(err.Error(), "stesso punto") {
		t.Fatalf("atteso errore «stesso punto», ottenuto %v", err)
	}
}

func TestApplyRejectsEmptyAnnotation(t *testing.T) {
	_, err := Apply("<p>abc</p>", "", []Annotation{{After: "abc", Text: ""}})
	if err == nil || !strings.Contains(err.Error(), "obbligatori") {
		t.Fatalf("atteso errore su campi obbligatori, ottenuto %v", err)
	}
}

func TestApplySkipsWhatIsAlreadyThere(t *testing.T) {
	prepend := "<h2>Secondo</h2>"
	first, err := Apply(twoCycles, prepend, []Annotation{{After: "manca il test", Text: "✅"}})
	if err != nil {
		t.Fatal(err)
	}
	again, err := Apply(first.Text, prepend, []Annotation{{After: "manca il test", Text: "✅"}})
	if err != nil {
		t.Fatal(err)
	}
	if again.Changed || again.Prepended || len(again.Skipped) != 1 || again.Text != first.Text {
		t.Fatalf("una seconda applicazione non deve aggiungere nulla: %+v", again)
	}
}

func TestApplyOnEmptyDescription(t *testing.T) {
	res, err := Apply("", "<h2>Primo ciclo</h2>", nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.Text != "<h2>Primo ciclo</h2>" || !res.Changed {
		t.Fatalf("su description vuota il blocco deve diventare tutto il testo: %+v", res)
	}
}

func TestApplyKeepsUTF8(t *testing.T) {
	res, err := Apply("<p>perché è così</p>", "", []Annotation{{After: "perché è", Text: "⚠️"}})
	if err != nil {
		t.Fatal(err)
	}
	if !utf8.ValidString(res.Text) || res.Text != "<p>perché è ⚠️ così</p>" {
		t.Fatalf("inserimento con caratteri multi-byte sbagliato: %q", res.Text)
	}
}
```

- [ ] **Passo 2: verifica che falliscano**

Esegui (da `plugins/wm-skills/mcp`): `go test ./internal/compose/...`
Atteso: FAIL, `undefined: Apply` / `undefined: Annotation`.

- [ ] **Passo 3: implementa**

`plugins/wm-skills/mcp/internal/compose/compose.go`:

```go
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

	for i, a := range annotations {
		n := i + 1
		if a.After == "" || a.Text == "" {
			return Result{}, fmt.Errorf("annotazione %d: after e text sono obbligatori", n)
		}
		positions := indexAll(current, a.After)
		if len(positions) == 0 {
			return Result{}, fmt.Errorf("annotazione %d: la frase %q non compare nella description attuale", n, a.After)
		}
		occ := a.Occurrence
		if occ == 0 {
			if len(positions) > 1 {
				return Result{}, fmt.Errorf("annotazione %d: la frase %q compare %d volte: indica occurrence (1 = la prima)", n, a.After, len(positions))
			}
			occ = 1
		}
		if occ < 1 || occ > len(positions) {
			return Result{}, fmt.Errorf("annotazione %d: occurrence %d non esiste, la frase %q compare %d volte", n, occ, a.After, len(positions))
		}
		pos := positions[occ-1] + len(a.After)
		if insideTag(current, pos) {
			return Result{}, fmt.Errorf("annotazione %d: il punto dopo %q cade dentro un tag HTML", n, a.After)
		}
		if used[pos] {
			return Result{}, fmt.Errorf("annotazione %d: un'altra annotazione va nello stesso punto, dopo %q", n, a.After)
		}
		used[pos] = true
		if strings.HasPrefix(current[pos:], " "+a.Text) {
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
		if insideTag(s, from) {
			from += strings.Index(s[from:], ">") + 1
		}
	}
	to := len(s)
	if loc := blockClose.FindStringIndex(s[end:]); loc != nil {
		to = end + loc[1]
	} else if end+120 < len(s) {
		to = end + 120
		if insideTag(s, to) {
			to += strings.Index(s[to:], ">") + 1
		}
	}
	return s[from:to]
}
```

- [ ] **Passo 4: verifica che passino**

Esegui: `go test ./internal/compose/...`
Atteso: `ok  .../internal/compose`.

- [ ] **Passo 5: commit (solo dopo conferma del dev)**

```bash
git add plugins/wm-skills/mcp/internal/compose/
git commit -m "feat(nota-in-testa-description): funzione che compone la description senza toccare il testo esistente"
```

---

### Task 2: anteprima completa e leggibile

**File:**
- Modifica: `plugins/wm-skills/mcp/internal/preview/preview.go` (intero file)
- Crea: `plugins/wm-skills/mcp/internal/preview/readable.go`
- Modifica: `plugins/wm-skills/mcp/internal/preview/preview_test.go:45-67` (sostituisce
  `TestDiffTruncatesLongStringOnRuneBoundary`)
- Crea: `plugins/wm-skills/mcp/internal/preview/readable_test.go`

**Interfacce:**
- Consuma: niente.
- Produce: `func Readable(s string) string`; `Diff(current, requested map[string]any) string` e
  `NewResource(requested map[string]any) string` con le stesse firme di oggi, senza troncamento.

- [ ] **Passo 1: scrivi i test che falliscono**

`plugins/wm-skills/mcp/internal/preview/readable_test.go`:

```go
package preview

import "testing"

func TestReadableRendersStructure(t *testing.T) {
	in := `<h2>Secondo ciclo</h2><p>Esito: <strong>DA CORREGGERE</strong></p><ul><li>Bloccante 1</li><li>Vedi <a href="https://github.com/x/pull/1">PR</a></li></ul>`
	want := "SECONDO CICLO\n\nEsito: DA CORREGGERE\n\n- Bloccante 1\n- Vedi PR (https://github.com/x/pull/1)"
	if got := Readable(in); got != want {
		t.Fatalf("resa sbagliata:\n got %q\nwant %q", got, want)
	}
}

func TestReadableDecodesEntitiesInHeadings(t *testing.T) {
	if got := Readable("<h3>Rischi &amp; rollback</h3><p>a&nbsp;b</p>"); got != "RISCHI & ROLLBACK\n\na b" {
		t.Fatalf("entità non decodificate: %q", got)
	}
}

func TestReadableNestedLists(t *testing.T) {
	in := "<ul><li>uno<ul><li>due</li></ul></li></ul>"
	if got := Readable(in); got != "- uno\n  - due" {
		t.Fatalf("elenco annidato sbagliato: %q", got)
	}
}

func TestReadableLeavesPlainTextAlone(t *testing.T) {
	if got := Readable("testo semplice"); got != "testo semplice" {
		t.Fatalf("un testo senza tag deve restare com'è: %q", got)
	}
}
```

In `plugins/wm-skills/mcp/internal/preview/preview_test.go` sostituisci per intero
`TestDiffTruncatesLongStringOnRuneBoundary` (righe 45-67) con:

```go
func TestDiffShowsLongStringInFull(t *testing.T) {
	s := strings.Repeat("x", 119) + strings.Repeat("à", 500) + "FINE"
	got := Diff(map[string]any{"campo": nil}, map[string]any{"campo": s})

	if !utf8.ValidString(got) {
		t.Fatalf("UTF-8 non valido:\n%q", got)
	}
	if !strings.Contains(got, s) {
		t.Fatalf("l'anteprima deve contenere il valore per intero, senza troncamenti:\n%s", got)
	}
}

func TestDiffRendersDescriptionAsReadableText(t *testing.T) {
	got := Diff(
		map[string]any{"description": "<p>vecchio</p>"},
		map[string]any{"description": "<h2>Nuovo</h2><p>vecchio</p>"},
	)
	if strings.Contains(got, "<h2>") || strings.Contains(got, "<p>") {
		t.Fatalf("la description va mostrata come testo, non come HTML:\n%s", got)
	}
	if !strings.Contains(got, "NUOVO") || !strings.Contains(got, "prima:") || !strings.Contains(got, "dopo:") {
		t.Fatalf("mancano etichette o contenuto:\n%s", got)
	}
}

func TestDiffWarnsWhenDescriptionLosesMostOfItsText(t *testing.T) {
	got := Diff(
		map[string]any{"description": strings.Repeat("a", 1000)},
		map[string]any{"description": strings.Repeat("a", 100)},
	)
	if !strings.Contains(got, "⚠️ description: 1000 → 100 caratteri — si perde il 90% del testo attuale.") {
		t.Fatalf("manca l'avviso sul testo perso:\n%s", got)
	}
	if strings.Index(got, "⚠️") > strings.Index(got, "prima:") {
		t.Fatalf("l'avviso va in cima all'anteprima:\n%s", got)
	}
}

func TestDiffNoWarningWhenDescriptionGrows(t *testing.T) {
	got := Diff(map[string]any{"description": "a"}, map[string]any{"description": "ab"})
	if strings.Contains(got, "si perde") {
		t.Fatalf("nessun avviso se il testo cresce:\n%s", got)
	}
}
```

- [ ] **Passo 2: verifica che falliscano**

Esegui: `go test ./internal/preview/...`
Atteso: FAIL, `undefined: Readable` e `TestDiffShowsLongStringInFull` fallito per troncamento.

- [ ] **Passo 3: implementa**

`plugins/wm-skills/mcp/internal/preview/readable.go`:

```go
package preview

import (
	"html"
	"regexp"
	"strings"
)

var (
	tagRe       = regexp.MustCompile(`(?s)<(/?)([a-zA-Z0-9]+)([^>]*)>`)
	hrefRe      = regexp.MustCompile(`(?i)href\s*=\s*["']([^"']*)["']`)
	spacesRe    = regexp.MustCompile(`[ \t\r\n]+`)
	blankRunsRe = regexp.MustCompile(`\n{3,}`)
)

// Readable rende un campo HTML come testo da leggere in una sessione di
// terminale: titoli in maiuscolo su una riga a sé, elenchi con «- » e rientro,
// paragrafi separati da una riga vuota, link come testo con l'indirizzo fra
// parentesi, entità decodificate. Serve all'anteprima: il dev deve vedere
// cosa verrà scritto senza leggere tag.
func Readable(s string) string {
	var b strings.Builder
	var hrefs []string
	heading := false
	depth := 0
	last := 0

	writeText := func(raw string) {
		t := spacesRe.ReplaceAllString(html.UnescapeString(raw), " ")
		t = strings.ReplaceAll(t, "\u00a0", " ")
		if heading {
			t = strings.ToUpper(t)
		}
		b.WriteString(t)
	}

	for _, m := range tagRe.FindAllStringSubmatchIndex(s, -1) {
		writeText(s[last:m[0]])
		last = m[1]
		closing := s[m[2]:m[3]] == "/"
		name := strings.ToLower(s[m[4]:m[5]])
		attrs := s[m[6]:m[7]]

		switch name {
		case "h1", "h2", "h3", "h4", "h5", "h6":
			heading = !closing
			b.WriteString("\n\n")
		case "p", "div", "table":
			b.WriteString("\n\n")
		case "br", "tr":
			b.WriteString("\n")
		case "ul", "ol":
			// solo la profondità: l'a capo lo scrive ogni <li>, altrimenti
			// fra una voce e la sua sottolista resterebbe una riga vuota
			if closing {
				depth--
			} else {
				depth++
			}
		case "li":
			if !closing {
				b.WriteString("\n" + strings.Repeat("  ", max(depth-1, 0)) + "- ")
			}
		case "td", "th":
			if !closing {
				b.WriteString(" | ")
			}
		case "a":
			if !closing {
				href := ""
				if hm := hrefRe.FindStringSubmatch(attrs); hm != nil {
					href = hm[1]
				}
				hrefs = append(hrefs, href)
			} else if n := len(hrefs); n > 0 {
				if hrefs[n-1] != "" {
					b.WriteString(" (" + hrefs[n-1] + ")")
				}
				hrefs = hrefs[:n-1]
			}
		}
	}
	writeText(s[last:])

	lines := strings.Split(b.String(), "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " ")
		// il rientro degli elenchi va conservato, gli spazi iniziali del
		// testo che segue un tag no
		if trimmed := strings.TrimLeft(lines[i], " "); !strings.HasPrefix(trimmed, "- ") {
			lines[i] = trimmed
		}
	}
	out := blankRunsRe.ReplaceAllString(strings.Join(lines, "\n"), "\n\n")
	return strings.Trim(out, "\n")
}
```

`plugins/wm-skills/mcp/internal/preview/preview.go` (sostituisce il file):

```go
// Package preview costruisce il testo che il dev legge prima di autorizzare una
// scrittura. La differenza è calcolata sui dati veri letti da Orchestrator, non
// ricostruita a memoria, e mostra ogni campo per intero: un'anteprima troncata
// fa approvare alla cieca.
package preview

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"
)

// htmlFields sono i campi resi da un editor visuale su Orchestrator: in
// anteprima si mostrano come testo leggibile, non come HTML.
var htmlFields = map[string]bool{"description": true, "customer_request": true}

// Diff descrive cosa cambierebbe applicando requested allo stato current.
func Diff(current, requested map[string]any) string {
	names := sortedKeys(requested)

	var blocks []string
	for _, n := range names {
		before := format(n, current[n])
		after := format(n, requested[n])
		if before == after {
			continue
		}
		blocks = append(blocks, n+"\n"+labelled("prima", before)+"\n"+labelled("dopo", after))
	}

	if len(blocks) == 0 {
		return "Nessuna modifica: i valori richiesti coincidono con quelli attuali."
	}

	var b strings.Builder
	b.WriteString("ANTEPRIMA — nulla è stato scritto\n\n")
	if w := lossWarning(current["description"], requested["description"]); w != "" {
		b.WriteString(w + "\n\n")
	}
	b.WriteString(strings.Join(blocks, "\n\n"))
	b.WriteString("\n\nPer applicare, richiama lo stesso tool con confirm: true.")
	return b.String()
}

// NewResource descrive una creazione, dove non esiste uno stato precedente.
func NewResource(requested map[string]any) string {
	var b strings.Builder
	b.WriteString("ANTEPRIMA — nulla è stato creato\n\n")
	for _, n := range sortedKeys(requested) {
		b.WriteString(labelled(n, format(n, requested[n])) + "\n")
	}
	b.WriteString("\nPer creare, richiama lo stesso tool con confirm: true.")
	return b.String()
}

// lossWarning avvisa quando la description nuova è meno della metà di quella
// attuale: è il segno tipico di una sostituzione usata al posto di
// un'aggiunta. Non blocca nulla, rende l'errore impossibile da non vedere.
func lossWarning(before, after any) string {
	b, okB := before.(string)
	a, okA := after.(string)
	if !okB || !okA {
		return ""
	}
	nb, na := utf8.RuneCountInString(b), utf8.RuneCountInString(a)
	if nb == 0 || na*2 >= nb {
		return ""
	}
	return fmt.Sprintf("⚠️ description: %d → %d caratteri — si perde il %d%% del testo attuale.", nb, na, 100-na*100/nb)
}

// labelled scrive «etichetta: valore» sulla stessa riga se il valore è una
// riga sola, altrimenti l'etichetta e sotto il valore rientrato.
func labelled(label, value string) string {
	if !strings.Contains(value, "\n") {
		return fmt.Sprintf("  %s: %s", label, value)
	}
	return fmt.Sprintf("  %s:\n    %s", label, strings.ReplaceAll(value, "\n", "\n    "))
}

func sortedKeys(m map[string]any) []string {
	names := make([]string, 0, len(m))
	for k := range m {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

func format(field string, v any) string {
	switch t := v.(type) {
	case nil:
		return "(vuoto)"
	case string:
		if htmlFields[field] {
			t = Readable(t)
		}
		if t == "" {
			return "(vuoto)"
		}
		return t
	case float64:
		if t == float64(int64(t)) {
			return fmt.Sprintf("%d", int64(t))
		}
		return fmt.Sprintf("%g", t)
	default:
		return fmt.Sprintf("%v", v)
	}
}
```

- [ ] **Passo 4: verifica che passino tutti i test del pacchetto e del modulo**

Esegui: `go test ./...`
Atteso: tutto `ok`. Se un test di un altro pacchetto (per esempio `internal/tools`) controllava
il vecchio formato fra virgolette o il troncamento, aggiorna **l'asserzione** al formato nuovo
(`etichetta: valore` per intero), non il codice.

- [ ] **Passo 5: commit (solo dopo conferma del dev)**

```bash
git add plugins/wm-skills/mcp/internal/preview/
git commit -m "feat(nota-in-testa-description): anteprima con i campi per intero e resi come testo leggibile"
```

---

### Task 3: `update_story` sceglie fra sostituzione e aggiunta

**File:**
- Modifica: `plugins/wm-skills/mcp/internal/tools/stories.go` (struct `storyFieldsInput`,
  handler di `update_story`, `storyFieldsInputSchema`, chiamate di `create_story`)
- Modifica: `plugins/wm-skills/mcp/internal/tools/handlers_test.go` (server di prova con corpo
  configurabile, test nuovi)

**Interfacce:**
- Consuma: `compose.Apply`, `compose.Annotation`, `compose.Result` (Task 1); `preview.Diff`,
  `preview.Readable` (Task 2).
- Produce: parametri `prepend` e `annotations` di `update_story`; `storyFieldsInputSchema(e
  *enums.Values, forUpdate bool) map[string]any`.

- [ ] **Passo 1: scrivi i test che falliscono**

In `handlers_test.go`, trasforma `newRecordingServer` in una funzione con corpo configurabile e
lascia la vecchia come scorciatoia:

```go
func newRecordingServer(t *testing.T) (*recordingServer, *httptest.Server) {
	return newRecordingServerWithBody(t, `{"id":1,"name":"Titolo attuale","status":"todo"}`)
}

// newRecordingServerWithBody risponde a ogni richiesta con body: serve ai test
// che devono leggere una description precisa dal ticket.
func newRecordingServerWithBody(t *testing.T, body string) (*recordingServer, *httptest.Server) {
	t.Helper()
	rs := &recordingServer{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rs.mu.Lock()
		defer rs.mu.Unlock()
		switch r.Method {
		case http.MethodGet:
			rs.reads++
		case http.MethodPost, http.MethodPatch, http.MethodDelete:
			rs.writes++
			rs.lastWriteMeth = r.Method
			rs.lastWritePath = r.URL.Path
			var b map[string]any
			_ = json.NewDecoder(r.Body).Decode(&b)
			rs.lastWriteBody = b
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return rs, srv
}

// resultText unisce il testo di tutti i blocchi del risultato.
func resultText(res *mcp.CallToolResult) string {
	var parts []string
	for _, c := range res.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			parts = append(parts, tc.Text)
		}
	}
	return strings.Join(parts, "\n")
}
```

(aggiungi `"strings"` agli import di `handlers_test.go`).

Poi aggiungi i test:

```go
const storyWithCycle = `{"id":8543,"status":"progress","description":"<h2>Primo ciclo</h2><ul><li>perde la quota</li></ul>"}`

func TestUpdateStoryAppendSendsComposedDescription(t *testing.T) {
	rs, srv := newRecordingServerWithBody(t, storyWithCycle)
	session := testSession(t, map[string]bool{"stories": true, "me": true}, srv.URL)

	res := callTool(t, session, "update_story", map[string]any{
		"story_id":    8543,
		"status":      "todo",
		"prepend":     "<h2>Secondo ciclo</h2>",
		"annotations": []map[string]any{{"after": "perde la quota", "text": "<strong>✅ Risolto</strong>"}},
		"confirm":     true,
	})
	if res.IsError {
		t.Fatalf("update_story in aggiunta non deve fallire: %s", resultText(res))
	}

	rs.mu.Lock()
	defer rs.mu.Unlock()
	if rs.writes != 1 || rs.lastWriteMeth != http.MethodPatch {
		t.Fatalf("attesa una sola PATCH, ricevute %d scritture (%s)", rs.writes, rs.lastWriteMeth)
	}
	want := "<h2>Secondo ciclo</h2><h2>Primo ciclo</h2><ul><li>perde la quota <strong>✅ Risolto</strong></li></ul>"
	if rs.lastWriteBody["description"] != want || rs.lastWriteBody["status"] != "todo" {
		t.Fatalf("corpo della PATCH sbagliato: %+v", rs.lastWriteBody)
	}
	if _, leaked := rs.lastWriteBody["prepend"]; leaked {
		t.Fatalf("prepend non deve arrivare a Orchestrator: %+v", rs.lastWriteBody)
	}
}

func TestUpdateStoryAppendPreviewShowsEverythingReadable(t *testing.T) {
	rs, srv := newRecordingServerWithBody(t, storyWithCycle)
	session := testSession(t, map[string]bool{"stories": true, "me": true}, srv.URL)

	res := callTool(t, session, "update_story", map[string]any{
		"story_id":    8543,
		"prepend":     "<h2>Secondo ciclo</h2>",
		"annotations": []map[string]any{{"after": "perde la quota", "text": "<strong>✅ Risolto</strong>"}},
	})
	text := resultText(res)
	if res.IsError {
		t.Fatalf("l'anteprima non deve fallire: %s", text)
	}
	for _, want := range []string{"SECONDO CICLO", "PRIMO CICLO", "- perde la quota ✅ Risolto", "annotazione 1", "prima:", "dopo:"} {
		if !strings.Contains(text, want) {
			t.Fatalf("all'anteprima manca %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "<h2>") {
		t.Fatalf("l'anteprima deve essere testo leggibile, non HTML:\n%s", text)
	}
	rs.mu.Lock()
	defer rs.mu.Unlock()
	if rs.writes != 0 {
		t.Fatalf("senza confirm nessuna scrittura, ricevute %d", rs.writes)
	}
}

func TestUpdateStoryRejectsDescriptionTogetherWithPrepend(t *testing.T) {
	rs, srv := newRecordingServerWithBody(t, storyWithCycle)
	session := testSession(t, map[string]bool{"stories": true, "me": true}, srv.URL)

	res := callTool(t, session, "update_story", map[string]any{
		"story_id":    8543,
		"description": "<p>tutto nuovo</p>",
		"prepend":     "<h2>x</h2>",
		"confirm":     true,
	})
	if !res.IsError || !strings.Contains(resultText(res), "non insieme") {
		t.Fatalf("description e prepend insieme vanno rifiutati: %s", resultText(res))
	}
	rs.mu.Lock()
	defer rs.mu.Unlock()
	if rs.writes != 0 {
		t.Fatalf("nessuna scrittura attesa, ricevute %d", rs.writes)
	}
}

func TestUpdateStoryAppendAlreadyPresentWritesNothing(t *testing.T) {
	body := `{"id":8543,"status":"todo","description":"<h2>Secondo ciclo</h2><h2>Primo ciclo</h2>"}`
	rs, srv := newRecordingServerWithBody(t, body)
	session := testSession(t, map[string]bool{"stories": true, "me": true}, srv.URL)

	res := callTool(t, session, "update_story", map[string]any{
		"story_id": 8543,
		"prepend":  "<h2>Secondo ciclo</h2>",
		"confirm":  true,
	})
	if res.IsError || !strings.Contains(resultText(res), "già presente") {
		t.Fatalf("un blocco già in testa non va riaggiunto: %s", resultText(res))
	}
	rs.mu.Lock()
	defer rs.mu.Unlock()
	if rs.writes != 0 {
		t.Fatalf("nessuna scrittura attesa, ricevute %d", rs.writes)
	}
}

func TestUpdateStoryAppendAlreadyPresentStillWritesOtherFields(t *testing.T) {
	body := `{"id":8543,"status":"progress","description":"<h2>Secondo ciclo</h2><h2>Primo ciclo</h2>"}`
	rs, srv := newRecordingServerWithBody(t, body)
	session := testSession(t, map[string]bool{"stories": true, "me": true}, srv.URL)

	res := callTool(t, session, "update_story", map[string]any{
		"story_id": 8543,
		"status":   "todo",
		"prepend":  "<h2>Secondo ciclo</h2>",
		"confirm":  true,
	})
	if res.IsError {
		t.Fatalf("non deve fallire: %s", resultText(res))
	}
	rs.mu.Lock()
	defer rs.mu.Unlock()
	if rs.writes != 1 || rs.lastWriteBody["status"] != "todo" {
		t.Fatalf("lo status va scritto comunque: %+v", rs.lastWriteBody)
	}
	if _, has := rs.lastWriteBody["description"]; has {
		t.Fatalf("la description non va riscritta se non c'è nulla da aggiungere: %+v", rs.lastWriteBody)
	}
}

func TestUpdateStoryAppendOnNullDescription(t *testing.T) {
	rs, srv := newRecordingServerWithBody(t, `{"id":8543,"status":"todo","description":null}`)
	session := testSession(t, map[string]bool{"stories": true, "me": true}, srv.URL)

	res := callTool(t, session, "update_story", map[string]any{
		"story_id": 8543,
		"prepend":  "<h2>Primo ciclo</h2>",
		"confirm":  true,
	})
	if res.IsError {
		t.Fatalf("non deve fallire: %s", resultText(res))
	}
	rs.mu.Lock()
	defer rs.mu.Unlock()
	if rs.lastWriteBody["description"] != "<h2>Primo ciclo</h2>" {
		t.Fatalf("su description nulla il blocco diventa tutto il testo: %+v", rs.lastWriteBody)
	}
}

func TestUpdateStoryRejectsUnknownParameter(t *testing.T) {
	rs, srv := newRecordingServer(t)
	session := testSession(t, map[string]bool{"stories": true, "me": true}, srv.URL)

	res := callTool(t, session, "update_story", map[string]any{
		"story_id": 8543,
		"prepen":   "<h2>refuso</h2>",
		"confirm":  true,
	})
	if !res.IsError {
		t.Fatalf("un parametro sconosciuto deve dare errore, non essere ignorato")
	}
	rs.mu.Lock()
	defer rs.mu.Unlock()
	if rs.writes != 0 {
		t.Fatalf("nessuna scrittura attesa, ricevute %d", rs.writes)
	}
}
```

- [ ] **Passo 2: verifica che falliscano**

Esegui: `go test ./internal/tools/ -run 'TestUpdateStory'`
Atteso: FAIL sui test nuovi (parametri `prepend`/`annotations` ignorati, `prepen` accettato); i
due test `TestUpdateStory…` già esistenti passano.

- [ ] **Passo 3: implementa**

In `stories.go`:

1. Import: aggiungi
   `"github.com/webmappsrl/claude-marketplace/plugins/wm-skills/mcp/internal/compose"`.

2. In `storyFieldsInput`, dopo `Tags`, aggiungi:

```go
	Prepend     string               `json:"prepend,omitempty" jsonschema:"HTML da mettere in testa alla description attuale, che resta identica"`
	Annotations []compose.Annotation `json:"annotations,omitempty" jsonschema:"testi da inserire dopo frasi esatte della description attuale"`
```

3. Sostituisci il corpo dell'handler di `update_story` e la sua `Description`:

```go
	mcp.AddTool(server, &mcp.Tool{
		Name: "update_story",
		Description: "Modifica un ticket Orchestrator. Senza confirm mostra, campo per campo e per intero, " +
			"cosa cambierebbe, senza scrivere nulla. Per AGGIUNGERE alla description usa prepend (in testa) " +
			"e annotations (dopo frasi esatte): il testo esistente resta identico. description invece " +
			"SOSTITUISCE tutto il campo: usala solo per riformattare o invalidare. " +
			"Scrivere customer_request invia una notifica al cliente: non va mai autorizzato in automatico.",
		InputSchema: storyFieldsInputSchema(deps.Enums, true),
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in storyFieldsInput) (*mcp.CallToolResult, any, error) {
		appending := in.Prepend != "" || len(in.Annotations) > 0
		if appending && in.Description != "" {
			return nil, nil, fmt.Errorf("description sostituisce tutto il campo, prepend e annotations aggiungono: usa l'una o gli altri, non insieme")
		}

		fields := nonEmptyFields(in)
		for _, k := range []string{"story_id", "confirm", "prepend", "annotations"} {
			delete(fields, k)
		}

		if err := validateStoryFields(deps.Enums, fields); err != nil {
			return nil, nil, err
		}

		path := fmt.Sprintf("/api/stories/%d", in.StoryID)
		currentRaw, err := deps.Client.Do(ctx, "GET", path, nil)
		if err != nil {
			return nil, nil, err
		}
		var current map[string]any
		if err := json.Unmarshal(currentRaw, &current); err != nil {
			return nil, nil, fmt.Errorf("ticket illeggibile: %w", err)
		}

		// La composizione si rifà a ogni chiamata, anche a quella con confirm:
		// parte sempre dalla description letta adesso, non da quella vista
		// nell'anteprima.
		var composed compose.Result
		if appending {
			currentText, _ := current["description"].(string)
			composed, err = compose.Apply(currentText, in.Prepend, in.Annotations)
			if err != nil {
				return nil, nil, err
			}
			if composed.Changed {
				fields["description"] = composed.Text
			}
			if len(fields) == 0 {
				return plainText("Già presente: il blocco e le annotazioni sono già nella description. Nessuna scrittura."), nil, nil
			}
		}

		summary := previewNotice(fields) + appendSummary(appending, composed) + preview.Diff(current, fields)
		if !in.Confirm {
			return plainText(summary), nil, nil
		}

		updated, err := deps.Client.Do(ctx, "PATCH", path, fields)
		if err != nil {
			return nil, nil, err
		}
		return dataResult("Scritto.\n"+summary, updated)
	})
```

4. In `create_story` cambia `InputSchema: storyFieldsInputSchema(deps.Enums)` in
   `storyFieldsInputSchema(deps.Enums, false)` e, dopo `delete(fields, "confirm")`, aggiungi
   `delete(fields, "prepend")` e `delete(fields, "annotations")`.

5. Aggiungi, dopo `previewNotice`:

```go
// appendSummary elenca cosa l'aggiunta mette nella description, prima della
// differenza completa: il blocco in testa e, per ogni annotazione, il blocco
// in cui finisce reso come testo, così il dev vede se l'etichetta è al posto
// giusto.
func appendSummary(appending bool, r compose.Result) string {
	if !appending {
		return ""
	}
	var b strings.Builder
	b.WriteString("Aggiunta alla description\n")
	if r.Prepended {
		b.WriteString("  in testa: sì (il testo esistente resta identico)\n")
	} else {
		b.WriteString("  in testa: niente (vuoto o già presente)\n")
	}
	for i, p := range r.Placed {
		fmt.Fprintf(&b, "  annotazione %d, dopo %q:\n    %s\n", i+1, p.Annotation.After,
			strings.ReplaceAll(preview.Readable(p.Context), "\n", "\n    "))
	}
	for _, s := range r.Skipped {
		fmt.Fprintf(&b, "  già presente, non riaggiunta: %q dopo %q\n", s.Text, s.After)
	}
	return b.String() + "\n"
}

// plainText restituisce un unico blocco di testo semplice, con gli a capo
// veri: l'anteprima va letta dal dev così com'è, non dentro un JSON.
func plainText(s string) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: s}}}
}
```

6. In `storyFieldsInputSchema` cambia la firma in
   `func storyFieldsInputSchema(e *enums.Values, forUpdate bool) map[string]any`, aggiorna il
   commento («… di create_story e update_story; con forUpdate aggiunge i parametri
   dell'aggiunta») e sostituisci il `return` finale con:

```go
	properties := map[string]any{
		// ... le proprietà di oggi, invariate: story_id, name, description,
		// customer_request, type, status, user_id, creator_id, estimated_hours,
		// tags, confirm
	}
	if forUpdate {
		properties["prepend"] = map[string]any{
			"type":        "string",
			"description": "HTML da mettere in testa alla description attuale, che resta identica. Non insieme a description.",
		}
		properties["annotations"] = map[string]any{
			"type": "array",
			"description": "testi da inserire subito dopo frasi esatte della description attuale (es. «✅ Risolto» accanto a un bloccante). Non insieme a description.",
			"items": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"after":      map[string]any{"type": "string", "description": "frase esatta, come compare nell'HTML della description attuale"},
					"text":       map[string]any{"type": "string", "description": "HTML da inserire dopo la frase, preceduto da uno spazio"},
					"occurrence": map[string]any{"type": "integer", "description": "quale presenza usare se la frase compare più volte (1 = la prima)"},
				},
				"required":             []string{"after", "text"},
				"additionalProperties": false,
			},
		}
	}
	return map[string]any{
		"type":                 "object",
		"properties":           properties,
		"additionalProperties": false,
	}
```

   (la mappa `properties` contiene esattamente le voci che oggi stanno sotto `"properties"`,
   spostate in una variabile senza modificarle).

- [ ] **Passo 4: verifica che passino tutti i test**

Esegui: `go test ./...`
Atteso: tutto `ok`. Se `TestUpdateStoryWithoutConfirmEmitsNoWriteRequest` o altri test leggono
l'anteprima come JSON `{"text": …}`, aggiorna l'asserzione per usare `resultText`.

- [ ] **Passo 5: commit (solo dopo conferma del dev)**

```bash
git add plugins/wm-skills/mcp/internal/tools/
git commit -m "feat(nota-in-testa-description): update_story aggiunge alla description con prepend e annotations"
```

---

### Task 4: le skill aggiungono invece di ricopiare

**File:**
- Modifica: `plugins/wm-skills/skills/wm-review-ticket/SKILL.md` (Fase 6b, righe 250-303 circa)
- Modifica: `plugins/wm-skills/skills/wm-plan/SKILL.md` (righe 222, 423, 1475-1479 circa)
- Modifica: `plugins/wm-skills/shared/orchestrator-fallback.md` (sezione «Aggiornamento ticket»,
  riga 83 circa)
- Modifica: `plugins/wm-skills/mcp/CLAUDE.md` (sezione «Vincoli da conoscere»)
- Modifica: `docs/knowledge/orchestrator-integrazione.md` (paragrafo «In scrittura i due campi»,
  righe 37-41)

**Interfacce:**
- Consuma: parametri `prepend`, `annotations` di `update_story` (Task 3); anteprima completa
  (Task 2).
- Produce: nessuna interfaccia di codice.

- [ ] **Passo 1: `wm-review-ticket`, paragrafo sulla sovrascrittura**

In `wm-review-ticket/SKILL.md` sostituisci il paragrafo che inizia con
`**La `description` si sovrascrive per intero.**` (4 righe) con:

```markdown
**L'esito si aggiunge, non si riscrive.** `update_story` con `description` sostituirebbe tutto il
campo e cancellerebbe i cicli precedenti. Usa invece i parametri dell'aggiunta: la sezione del
ciclo corrente va in `prepend`, le etichette sui cicli precedenti in `annotations`. Il tool legge
la `description`, compone il testo nuovo e lascia identico tutto il resto: non ricopiare mai la
`description` attuale.
```

- [ ] **Passo 2: `wm-review-ticket`, etichette**

Nello stesso file, nel punto 2 della struttura, sostituisci la frase
`aggiungi a ogni loro punto un'etichetta in grassetto con la data, senza toccare il testo
originale:` con:

```markdown
aggiungi a ogni loro punto un'etichetta in grassetto con la data, passandola in `annotations`:
`after` è la frase esatta del punto, come compare nell'HTML letto con `get_story`, e `text` è
l'etichetta. Se la frase compare più volte, indica `occurrence` (1 = la prima). Le etichette:
```

e subito dopo la frase `aggiungi in coda \`<strong>⚠️ Sostituito dal <N>-esimo ciclo (<data>)</strong>\`.`
aggiungi `Anche questa è un'annotazione, con \`after\` uguale al testo del titolo.`

- [ ] **Passo 3: `wm-review-ticket`, anteprima e chiamata**

Sostituisci il paragrafo `Prima di chiamare il tool mostra al dev la sezione nuova per intero …
non basta per rivedere il contenuto.` con:

```markdown
L'anteprima del tool mostra ogni campo per intero, reso come testo leggibile, e per ogni
etichetta il punto in cui finisce. Mostrala al dev così com'è: non riassumerla, non sostituirla
con un rimando a testo già mostrato, anche se ti sembra ripetitiva.
```

e sostituisci il paragrafo `Chiama \`update_story\` con \`story_id: <ID>\`, \`status: <status
scelto>\` e \`description: <description completa>\` … e \`confirm: true\`.` con:

```markdown
Chiama `update_story` con `story_id: <ID>`, `status: <status scelto>`, `prepend: <sezione del
ciclo corrente in HTML>` e `annotations: <etichette sui cicli precedenti>`, **senza**
`description` e senza `confirm`: esito e status partono in una sola scrittura. Mostra al dev
l'anteprima del tool, attendi conferma esplicita, poi richiama `update_story` con gli stessi campi
e `confirm: true`. Se il tool rifiuta un'annotazione (frase che non c'è, che compare più volte,
che cade dentro un tag), correggi `after` o `occurrence` e rifai l'anteprima: non ripiegare mai
su `description`.
```

- [ ] **Passo 4: `wm-plan`, regola sulle scritture (riga 222)**

In `wm-plan/SKILL.md`, alla fine del paragrafo `**Regola sulle scritture.** …` aggiungi:

```markdown
L'anteprima del tool mostra ogni campo per intero, con i campi HTML resi come testo leggibile:
riportala al dev così com'è, etichetta e contenuto di ogni campo, anche se ripetitiva. Mai
riassumerla, mai un rimando del tipo «il testo è quello scritto sopra».

**Aggiungere o sostituire la `description`.** Per aggiungere informazione — note dev, esito di
una review, qualsiasi aggiunta — usa `update_story` con `prepend` (in testa) e, se serve,
`annotations` (dopo frasi esatte): il testo esistente resta identico e non va ricopiato. Usa
`description` solo per riformattare o invalidare il campo, perché lo sostituisce per intero. I
due modi non si combinano nella stessa chiamata.
```

- [ ] **Passo 5: `wm-plan`, aggiornamenti espliciti (riga 423)**

Sostituisci la frase finale `Se il campo è \`description\`, il valore inviato sostituisce tutto il
testo: aggiungi la nota in testa alla \`description\` attuale, come in \`update-context:
orchestrator\`.` con:

```markdown
Una nota da aggiungere alla `description` («scrivi nelle note dev che…») va in `prepend`, non in
`description`: vedi `## Orchestrator` → «Aggiungere o sostituire la `description`».
```

- [ ] **Passo 6: `wm-plan`, Checklist (righe 1475-1479)**

Sostituisci la voce `- [ ] Solo dopo approvazione esplicita, chiama \`update_story\` con i campi
\`status\`, \`description\`, \`customer_request\` …` e il blocco `**Importante — i due campi si
comportano in modo diverso:**` con le sue due voci, con:

```markdown
- [ ] Solo dopo approvazione esplicita, chiama `update_story` con `status`, `customer_request` e
  `prepend: <bozza delle note dev in HTML>` — **non** `description` — prima senza `confirm` per
  mostrare l'anteprima, poi con `confirm: true`. Una sola chiamata: status, risposta al cliente e
  note partono insieme.

  **I due campi si comportano in modo diverso:**
  - `customer_request`: manda solo il testo pulito. Il backend chiama `addResponse()`, che aggiunge
    in testa la risposta con autore e data e notifica il cliente.
  - `description`: la PATCH la sostituisce per intero (Orchestrator oc:8549). Per questo le note
    dev vanno in `prepend`: il tool le mette in testa e lascia identici l'overview scritta in
    `Fase: ticket` e le review precedenti.
```

- [ ] **Passo 7: file di ripiego**

In `plugins/wm-skills/shared/orchestrator-fallback.md`, subito sotto il titolo
`### Aggiornamento ticket (PATCH — richiede conferma esplicita)`, aggiungi:

```markdown
**Non usare questo ripiego per aggiungere informazione alla `description`** (note dev, esito di
una review): la PATCH sostituisce tutto il campo, e senza il server MCP l'unico modo di non
perdere il testo sarebbe ricopiarlo a mano. Fermati e chiedi al dev di far ripartire il server.
```

- [ ] **Passo 8: vincolo nel CLAUDE.md del server MCP**

In `plugins/wm-skills/mcp/CLAUDE.md`, sezione «Vincoli da conoscere prima di toccare il codice»,
dopo la voce su `confirm` aggiungi:

```markdown
- **L'anteprima mostra ogni campo per intero**, con i campi HTML resi come testo leggibile
  (`preview.Readable`): niente troncamenti. Un'anteprima tagliata fa approvare alla cieca ed è
  così che `wm-plan` ha cancellato la `description` di 11 ticket senza che nessuno se ne
  accorgesse.
- **Aggiungere alla `description` si fa nel codice** (`internal/compose`), non chiedendo al
  modello di ricopiare il testo esistente: `update_story` con `prepend`/`annotations` legge,
  compone e manda una sola PATCH.
```

- [ ] **Passo 9: pagina di conoscenza**

In `docs/knowledge/orchestrator-integrazione.md` sostituisci il paragrafo delle righe 37-41 (`In
scrittura i due campi di una Story …`) con:

```markdown
In scrittura i due campi di una Story si comportano in modo diverso. `customer_request` passa da
`addResponse()`, che aggiunge il testo in testa con autore e data e notifica il cliente.
`description` invece **si sovrascrive per intero** da oc:8549, in produzione dal 15 settembre
2026: prima passava da `addDevNote()`, che aggiungeva in testa.

Per questo il server MCP distingue le due operazioni. `update_story` con `description`
sostituisce, e serve solo per riformattare o invalidare il campo. Con `prepend` e `annotations`
aggiunge: legge la `description`, mette il blocco in testa e le etichette dopo frasi esatte,
lascia identico il resto e manda una sola PATCH insieme agli altri campi. Il testo esistente non
passa mai dal modello: una copia fatta dal modello su testi lunghi può saltarne dei pezzi, e
l'anteprima troncata di allora non lo mostrava.

- **Chiedere all'agente di ricopiare la `description`** (PR #20, superata): funzionava su testi
  brevi, ma su ticket da 18.000 caratteri una copia del modello può riassumere, e nessuno lo vede.
- **Un endpoint di Orchestrator per aggiungere in testa**: scartato, la PATCH esiste già e il
  modo di usarla è una scelta degli strumenti; e oc:8549 aveva escluso un'opzione per scegliere
  fra aggiungere e sostituire.
```

- [ ] **Passo 10: verifica**

Esegui dalla root del repo:
```bash
grep -n "ricopia\|carattere per carattere\|copiata identica" plugins/wm-skills/skills/wm-review-ticket/SKILL.md plugins/wm-skills/skills/wm-plan/SKILL.md
claude plugin validate .
./.github/scripts/verifica-diagramma.sh
```
Atteso: `grep` non trova nulla; `✔ Validation passed`; `✔ Diagramma allineato`.

- [ ] **Passo 11: commit (solo dopo conferma del dev)**

```bash
git add plugins/wm-skills/skills/ plugins/wm-skills/shared/orchestrator-fallback.md plugins/wm-skills/mcp/CLAUDE.md docs/knowledge/orchestrator-integrazione.md
git commit -m "feat(nota-in-testa-description): le skill aggiungono alla description invece di ricopiarla"
```

---

### Task 5: binario e verifica finale

**File:**
- Modifica: `plugins/wm-skills/bin/orchestrator-mcp` (rigenerato, mai a mano)

**Interfacce:**
- Consuma: tutto il codice dei Task 1-3.
- Produce: il binario distribuito col plugin.

- [ ] **Passo 1: test completi**

Esegui (da `plugins/wm-skills/mcp`): `go vet ./... && go test ./...`
Atteso: nessun rilievo di `vet`, tutti i pacchetti `ok`.

- [ ] **Passo 2: ricompila**

Esegui: `./build.sh`
Atteso: il file `../bin/orchestrator-mcp` aggiornato, `file ../bin/orchestrator-mcp` riporta
`Mach-O 64-bit executable arm64`.

- [ ] **Passo 3: il binario espone i parametri nuovi**

Esegui (da `plugins/wm-skills/mcp`):
```bash
printf '%s\n' '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"t","version":"0"}}}' '{"jsonrpc":"2.0","method":"notifications/initialized"}' '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' | ../bin/orchestrator-mcp --auth-file /dev/null 2>/dev/null | grep -o '"prepend"' | head -1
```
Atteso: `"prepend"`. Se il server non parte senza credenziali valide, salta questo passo e dillo
al dev: la verifica vera è la prova a mano del passo 4.

- [ ] **Passo 4: prova a mano sull'istanza locale (facoltativa, solo `localhost:8099`)**

Solo se l'istanza locale di Orchestrator è avviata: con il server MCP configurato su
`http://localhost:8099`, chiama `update_story` su un ticket locale con `prepend` e un'annotazione,
prima senza `confirm` (controlla l'anteprima leggibile) e poi con `confirm: true`; rileggi il
ticket con `get_story` e verifica che il testo precedente sia intatto. **Mai contro
`orchestrator.maphub.it`.**

- [ ] **Passo 5: validazione del repo**

Esegui dalla root del repo:
```bash
claude plugin validate .
./.github/scripts/verifica-diagramma.sh
```
Atteso: `✔ Validation passed`; `✔ Diagramma allineato`.

- [ ] **Passo 6: commit (solo dopo conferma del dev)**

```bash
git add plugins/wm-skills/bin/orchestrator-mcp
git commit -m "build(nota-in-testa-description): ricompila il server MCP"
```
