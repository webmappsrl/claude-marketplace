package tools

import (
	"strings"
	"testing"
)

func TestAttachPathIsBuiltCorrectly(t *testing.T) {
	got := attachPath(42, 8503)
	if got != "/api/tags/42/stories/8503" {
		t.Fatalf("percorso sbagliato: %s", got)
	}
}

func TestTagDescriptionMentionsMarkdown(t *testing.T) {
	// La descrizione del tag è resa da un editor Markdown, a differenza della
	// description di una story che è HTML: il tool deve dirlo, altrimenti chi
	// lo usa applica la convenzione sbagliata.
	if !strings.Contains(strings.ToLower(tagDescriptionHint), "markdown") {
		t.Fatalf("la descrizione del campo deve citare Markdown: %s", tagDescriptionHint)
	}
}

func TestListTagsPathEncodesSpecialCharacters(t *testing.T) {
	// I tag di questo team si chiamano "[RDO][CLIENTE][ANNO]N": senza
	// codifica le parentesi cambiano silenziosamente la richiesta inviata.
	got := listTagsPath("[RDO][ACME][2026]1")
	want := "/api/tags?search=%5BRDO%5D%5BACME%5D%5B2026%5D1"
	if got != want {
		t.Fatalf("percorso non codificato correttamente:\n got:  %s\n want: %s", got, want)
	}
}

func TestListTagsPathEncodesSpacesAndPercent(t *testing.T) {
	got := listTagsPath("100% RDO cliente")
	want := "/api/tags?search=100%25+RDO+cliente"
	if got != want {
		t.Fatalf("percorso non codificato correttamente:\n got:  %s\n want: %s", got, want)
	}
}

func TestListTagsPathWithoutFilterHasNoQuery(t *testing.T) {
	if got := listTagsPath(""); got != "/api/tags" {
		t.Fatalf("senza filtro non deve comparire la query string: %s", got)
	}
}
