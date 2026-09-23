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
	// M6 — con una sola presenza il messaggio va al singolare, "1 volta", non
	// "1 volte".
	if !strings.Contains(err.Error(), "compare 1 volta") || strings.Contains(err.Error(), "1 volte") {
		t.Fatalf("atteso «compare 1 volta» al singolare, ottenuto %v", err)
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

// I1: dopo aver scritto l'annotazione sul secondo <p>, la riapplicazione
// (con e senza occurrence) deve dare Skipped, non un inserimento duplicato.
func TestApplyReapplyAfterPrependContainsPhrase(t *testing.T) {
	prepend := "<h2>Secondo</h2><ul><li>manca il test</li></ul>"
	base := "<p>manca il test</p><p>manca il test</p>"

	step1, err := Apply(base, prepend, []Annotation{{After: "manca il test", Text: "✅", Occurrence: 2}})
	if err != nil {
		t.Fatal(err)
	}
	wantStep1 := prepend + "<p>manca il test</p><p>manca il test ✅</p>"
	if step1.Text != wantStep1 {
		t.Fatalf("prima applicazione sbagliata:\n got %s\nwant %s", step1.Text, wantStep1)
	}

	// Riapplicazione con occurrence: deve individuare il secondo <p> di
	// current (non contare l'occorrenza nel prepend) e trovarla già annotata.
	step2, err := Apply(step1.Text, prepend, []Annotation{{After: "manca il test", Text: "✅", Occurrence: 2}})
	if err != nil {
		t.Fatal(err)
	}
	if step2.Changed || step2.Prepended || len(step2.Skipped) != 1 || step2.Text != step1.Text {
		t.Fatalf("riapplicazione con occurrence non deve aggiungere nulla: %+v", step2)
	}

	// Riapplicazione senza occurrence: in current[len(prepend):] la frase
	// compare ancora 2 volte (una senza annotazione, una con), quindi resta
	// ambigua — non deve mai dare errore «già presente» basato sul prepend.
	_, err = Apply(step1.Text, prepend, []Annotation{{After: "manca il test", Text: "✅"}})
	if err == nil || !strings.Contains(err.Error(), "compare 2 volte") {
		t.Fatalf("atteso errore «compare 2 volte» cercando solo in current, ottenuto %v", err)
	}
}

// M1: il controllo di "già presente" non deve scambiare un testo che inizia
// come l'annotazione per l'annotazione stessa.
func TestApplyDoesNotConfusePartialMatchWithAlreadyPresent(t *testing.T) {
	res, err := Apply("<p>quota ✅ Risolto (22/09)</p>", "", []Annotation{{After: "quota", Text: "✅"}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Text != "<p>quota ✅ ✅ Risolto (22/09)</p>" {
		t.Fatalf("doveva inserire, non considerarlo già presente: %s", res.Text)
	}

	again, err := Apply("<p>quota ✅</p>", "", []Annotation{{After: "quota", Text: "✅"}})
	if err != nil {
		t.Fatal(err)
	}
	if again.Changed || len(again.Skipped) != 1 {
		t.Fatalf("qui invece è davvero già presente: %+v", again)
	}
}

// M3: il ripiego a 120 byte non deve tagliare una rune multi-byte a metà.
func TestApplyBlockAroundKeepsUTF8Boundary(t *testing.T) {
	desc := strings.Repeat("è", 100) + "X" + strings.Repeat("è", 100)
	res, err := Apply(desc, "", []Annotation{{After: "X", Text: "✅"}})
	if err != nil {
		t.Fatal(err)
	}
	if !utf8.ValidString(res.Placed[0].Context) {
		t.Fatalf("Context non è UTF-8 valido: %q", res.Placed[0].Context)
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
