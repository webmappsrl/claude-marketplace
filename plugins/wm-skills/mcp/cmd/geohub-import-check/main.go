// Verifica dell'import di un'app da Geohub a uno shard wm-package. SOLA LETTURA: solo GET.
//
// Usato dalla skill wm-geohub-import-check (oc:8670).
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const uso = `Verifica dell'import di un'app da Geohub a uno shard wm-package. SOLA LETTURA: solo GET.

Confronta ciò che il frontend legge dalle due parti. Tre comandi:
  risolvi <url>                                 URL del frontend -> shard, id e indirizzi (JSON)
  scarica --geohub <url> --shard <url> --out D  scarica i file pubblici delle due app in D
  confronta D --regole <file>                   confronta i file in D e scrive diff.json e report.md
`

// opzione: il valore che segue `nome` fra gli argomenti, o "".
func opzione(argv []string, nome string) string {
	for i, a := range argv {
		if a == nome && i+1 < len(argv) {
			return argv[i+1]
		}
	}
	return ""
}

func presente(argv []string, nome string) bool {
	for _, a := range argv {
		if a == nome {
			return true
		}
	}
	return false
}

func stampaJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}

func esegui(argv []string) (int, error) {
	switch {
	case len(argv) >= 2 && argv[0] == "risolvi":
		amb, err := caricaEnvironment()
		if err != nil {
			return 0, err
		}
		r, err := risolvi(argv[1], amb)
		if err != nil {
			return 0, err
		}
		return 0, stampaJSON(r)
	case len(argv) >= 1 && argv[0] == "scarica":
		g, s, out := opzione(argv, "--geohub"), opzione(argv, "--shard"), opzione(argv, "--out")
		if g == "" || s == "" || out == "" {
			return 0, errore(1, "scarica vuole --geohub <url> --shard <url> --out <cartella>.")
		}
		m, err := scarica(g, s, out, !presente(argv, "--no-immagini"))
		if err != nil {
			return 0, err
		}
		fmt.Printf("Scaricato in %s: %d risorse non scaricate.\n", out, len(m.NonScaricate))
		return 0, nil
	case len(argv) >= 2 && argv[0] == "confronta":
		regole := opzione(argv, "--regole")
		if regole == "" {
			regole = filepath.Join(cartellaSkill(), "differenze-attese.json")
		}
		d, _, err := confronta(argv[1], regole)
		if err != nil {
			return 0, err
		}
		fmt.Println(riepilogo(d))
		fmt.Printf("Report: %s\n", filepath.Join(argv[1], "report.md"))
		return 0, nil
	case len(argv) >= 2 && argv[0] == "campioni":
		// Non documentato: i campioni di immagini che scarica sceglierebbe, per i test.
		return 0, stampaJSON(campioniImmagini(argv[1]))
	}
	fmt.Fprint(os.Stderr, uso)
	return 1, nil
}

func main() {
	codice, err := esegui(os.Args[1:])
	if err != nil {
		if e, ok := err.(*Errore); ok {
			fmt.Fprintf(os.Stderr, "❌ %s\n", e.Messaggio)
			os.Exit(e.Codice)
		}
		fmt.Fprintf(os.Stderr, "❌ %v\n", err)
		os.Exit(2)
	}
	os.Exit(codice)
}
