package main

import (
	"os"
	"testing"
)

// Il comportamento d'insieme lo copre scripts/tests/geohub-import-check.test.sh sulle fixture
// reali; qui ci sono i pezzi in cui Go si comporta diversamente da Python se non si sta attenti.

func ambienteDiProva(t *testing.T) *Ambiente {
	t.Helper()
	testo, err := os.ReadFile("../../../scripts/tests/fixtures/geohub-import-check/environment.ts")
	if err != nil {
		t.Fatal(err)
	}
	amb, err := leggiEnvironment(string(testo))
	if err != nil {
		t.Fatal(err)
	}
	return amb
}

func TestRisolvi(t *testing.T) {
	amb := ambienteDiProva(t)
	casi := []struct {
		url, shard string
		id         int64
		layout     string
	}{
		{"https://28.app.geohub.webmapp.it/", "geohub", 28, "vecchio"},
		{"https://3.maphubdev.maphub.it/", "maphubdev", 3, "nuovo"},
		{"3.MAPHUBDEV.maphub.it", "maphubdev", 3, "nuovo"},
		{"https://fiemaps.it/", "geohub", 29, "vecchio"},
	}
	for _, c := range casi {
		r, err := risolvi(c.url, amb)
		if err != nil {
			t.Fatalf("%s: %v", c.url, err)
		}
		if r.Shard != c.shard || r.AppID != c.id || r.Layout != c.layout {
			t.Errorf("%s: %s %d %s, atteso %s %d %s", c.url, r.Shard, r.AppID, r.Layout, c.shard, c.id, c.layout)
		}
	}
	if _, err := risolvi("https://www.esempio.it/", amb); err == nil {
		t.Error("un host senza id deve dare errore")
	}
}

func TestEnvironmentSenzaBlocchi(t *testing.T) {
	if _, err := leggiEnvironment("export const altro = {\n};"); err == nil {
		t.Error("senza i blocchi shards e redirects deve dare errore")
	}
}

func TestSvg(t *testing.T) {
	a := `<svg a='1'  b="2"><g/></svg>`
	b := "    <svg a=\"1\" b=\"2\">\n        <g></g>\n    </svg>\n"
	if svg(a) != svg(b) {
		t.Errorf("%q e %q devono coincidere", svg(a), svg(b))
	}
	// Senza riferimenti all'indietro in RE2: un tag aperto e uno chiuso diversi restano com'erano.
	if got := svg("<g></h>"); got != "<g></h>" {
		t.Errorf("tag diversi riscritti: %q", got)
	}
}

func TestNomeWhere(t *testing.T) {
	if nomeWhere("Massa Carrara") != nomeWhere(" massa-carrara ") {
		t.Error("spazi, trattini e maiuscole non devono contare")
	}
	if nomeWhere("Massa Carrara") != nomeWhere("Massa Carrara") {
		t.Error("lo spazio non separabile deve contare come spazio, come in Python")
	}
}

func TestDumpPy(t *testing.T) {
	got := dumpPy(map[string]any{"en": "b", "it": "à", "de": nil, "n": 350.0, "x": 74.6})
	want := `{"it": "à", "de": null, "en": "b", "n": 350, "x": 74.6}`
	if got != want {
		t.Errorf("%s, atteso %s", got, want)
	}
}

func TestJSONInStringa(t *testing.T) {
	for v, atteso := range map[string]bool{"[]": true, " false ": true, `{"a":1}`: true, "[x": false, "testo": false} {
		if jsonInStringa(v) != atteso {
			t.Errorf("%q: atteso %v", v, atteso)
		}
	}
	if jsonInStringa(3.0) {
		t.Error("un numero non è JSON scritto come testo")
	}
}

func TestMostraContaICaratteri(t *testing.T) {
	lungo := ""
	for range 250 {
		lungo += "è"
	}
	got := mostra(lungo).(string)
	if n := len([]rune(got)); n != 201 {
		t.Errorf("%d caratteri, attesi 200 più «…»", n)
	}
}
