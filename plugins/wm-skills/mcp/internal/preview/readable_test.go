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

func TestReadableKeepsUnknownAngleBracketsAsText(t *testing.T) {
	if got := Readable("<p>a<5 e b>10</p>"); got != "a<5 e b>10" {
		t.Fatalf("un confronto fra angolari non è un tag: %q", got)
	}
	if got := Readable("<p>Vedi List<String> generics</p>"); got != "Vedi List<String> generics" {
		t.Fatalf("un tipo generico non è un tag: %q", got)
	}
}

func TestReadablePreservesCodeBlockWhitespace(t *testing.T) {
	in := "<pre><code>func f() {\n  return 1\n}</code></pre>"
	want := "func f() {\n  return 1\n}"
	if got := Readable(in); got != want {
		t.Fatalf("il blocco di codice deve restare intatto:\n got %q\nwant %q", got, want)
	}
}

func TestReadableHeadingInsideListItemStaysOnSameLine(t *testing.T) {
	in := "<ul><li><h3>Titolo</h3>testo</li></ul>"
	want := "- TITOLO testo"
	if got := Readable(in); got != want {
		t.Fatalf("l'intestazione dentro un elenco non va a capo:\n got %q\nwant %q", got, want)
	}
}

// m3 — quando il testo dopo il tag chiuso porta già il suo spazio (caso
// comune: "<h3>Titolo</h3> testo"), lo spazio aggiunto per separare
// l'intestazione dal resto non va sommato a quello del testo: un solo spazio,
// non due.
func TestReadableHeadingInsideListItemCollapsesDoubleSpace(t *testing.T) {
	in := "<ul><li><h3>Titolo</h3> testo</li></ul>"
	want := "- TITOLO testo"
	if got := Readable(in); got != want {
		t.Fatalf("due spazi vanno ridotti a uno:\n got %q\nwant %q", got, want)
	}
}
