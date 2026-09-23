package preview

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestDiffShowsChangedFields(t *testing.T) {
	current := map[string]any{"status": "assigned", "name": "Titolo", "user_id": nil}
	requested := map[string]any{"status": "progress", "user_id": float64(12)}

	got := Diff(current, requested)

	if !strings.Contains(got, "status") || !strings.Contains(got, "assigned") || !strings.Contains(got, "progress") {
		t.Fatalf("la differenza deve mostrare vecchio e nuovo valore:\n%s", got)
	}
	if !strings.Contains(got, "user_id") {
		t.Fatalf("manca un campo modificato:\n%s", got)
	}
}

func TestDiffMarksUnchangedFields(t *testing.T) {
	current := map[string]any{"status": "progress", "name": "Titolo"}
	requested := map[string]any{"status": "progress"}

	got := Diff(current, requested)

	if !strings.Contains(strings.ToLower(got), "nessuna modifica") {
		t.Fatalf("un valore identico non è una modifica:\n%s", got)
	}
}

func TestDiffShowsEmptyPreviousValue(t *testing.T) {
	current := map[string]any{"estimated_hours": nil}
	requested := map[string]any{"estimated_hours": float64(3.5)}

	got := Diff(current, requested)

	if !strings.Contains(got, "(vuoto)") {
		t.Fatalf("un valore assente va reso esplicito:\n%s", got)
	}
}

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
		"description",
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

func TestNewResourceListsAllFields(t *testing.T) {
	got := NewResource(map[string]any{"name": "Nuovo ticket", "type": "Feature"})

	if !strings.Contains(got, "name") || !strings.Contains(got, "Nuovo ticket") {
		t.Fatalf("la creazione deve elencare i campi inviati:\n%s", got)
	}
}

// I1 — un cambio di solo markup (stesso testo reso, HTML grezzo diverso) deve
// comunque comparire come modifica: altrimenti l'anteprima dice «Nessuna
// modifica» mentre la PATCH, che manda il campo grezzo, scrive qualcosa di
// diverso.
func TestDiffShowsFormattingOnlyChangeAsModification(t *testing.T) {
	got := Diff(
		map[string]any{"description": "<p><b>Bloccante</b></p>"},
		map[string]any{"description": "<p><strong>Bloccante</strong></p>"},
		"description",
	)

	if strings.Contains(strings.ToLower(got), "nessuna modifica") {
		t.Fatalf("un markup diverso è comunque una modifica, anche se il testo reso è identico:\n%s", got)
	}
	if !strings.Contains(got, "cambia solo la formattazione (HTML), il testo resta uguale") {
		t.Fatalf("manca la riga che spiega che cambia solo il markup:\n%s", got)
	}
}

// I2 — senza dichiarare htmlFields, un campo chiamato "description" ma
// Markdown (come quello di un tag) non va reso come HTML: le sue entità
// restano scritte come le ha inviate l'utente, invece di essere decodificate
// da Readable.
func TestDiffLeavesNonHTMLDescriptionUntouched(t *testing.T) {
	got := Diff(
		map[string]any{"description": ""},
		map[string]any{"description": "## Cosa\n\n- a\n- Uso &amp; commerciale"},
	)

	if !strings.Contains(got, "&amp;") {
		t.Fatalf("una description Markdown non dichiarata come campo HTML non va decodificata da Readable:\n%s", got)
	}
	if strings.Contains(got, "- a\n    - Uso & commerciale") {
		t.Fatalf("l'entità non va decodificata: solo un campo HTML dichiarato lo farebbe:\n%s", got)
	}
}
