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

func TestDiffTruncatesLongStringOnRuneBoundary(t *testing.T) {
	// 119 caratteri di un byte seguiti da caratteri accentati di due byte:
	// tagliare a byte 120 (come faceva t[:120]) cade a metà del primo "à".
	// %q rende comunque la stringa risultante UTF-8 valida, ma nasconde il
	// difetto scrivendo la coda del byte spezzato come sequenza di scampo
	// letterale "\xc3" al posto dell'ultimo carattere — è quella che il
	// risultato non deve contenere.
	s := strings.Repeat("x", 119) + strings.Repeat("à", 50)
	current := map[string]any{"campo": nil}
	requested := map[string]any{"campo": s}

	got := Diff(current, requested)

	if !utf8.ValidString(got) {
		t.Fatalf("la troncatura ha prodotto UTF-8 non valido:\n%q", got)
	}
	if strings.Contains(got, `\x`) {
		t.Fatalf("la troncatura ha spezzato un carattere multi-byte (sequenza di scampo residua):\n%s", got)
	}
	if !strings.Contains(got, "à\"…") {
		t.Fatalf("la troncatura deve terminare su un carattere accentato intero:\n%s", got)
	}
}

func TestNewResourceListsAllFields(t *testing.T) {
	got := NewResource(map[string]any{"name": "Nuovo ticket", "type": "Feature"})

	if !strings.Contains(got, "name") || !strings.Contains(got, "Nuovo ticket") {
		t.Fatalf("la creazione deve elencare i campi inviati:\n%s", got)
	}
}
