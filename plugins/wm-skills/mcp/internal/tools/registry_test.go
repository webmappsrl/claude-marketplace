package tools

import "testing"

func TestActiveGroupsAlwaysIncludesCore(t *testing.T) {
	g := ActiveGroups("")
	if !g["stories"] || !g["me"] {
		t.Fatalf("stories e me devono essere sempre attivi: %v", g)
	}
	if g["crm"] {
		t.Fatalf("crm non deve essere attivo senza richiesta: %v", g)
	}
}

func TestActiveGroupsParsesList(t *testing.T) {
	g := ActiveGroups("tags, crm")
	if !g["tags"] || !g["crm"] {
		t.Fatalf("gruppi richiesti non attivati: %v", g)
	}
	if !g["stories"] {
		t.Fatalf("il nucleo resta attivo anche con una lista esplicita: %v", g)
	}
}

func TestActiveGroupsIgnoresUnknown(t *testing.T) {
	g := ActiveGroups("inesistente")
	if g["inesistente"] {
		t.Fatalf("un gruppo sconosciuto non va attivato: %v", g)
	}
}
