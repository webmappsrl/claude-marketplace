package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

// Errore ferma la skill: il messaggio va al dev così com'è, con il codice di uscita.
type Errore struct {
	Messaggio string
	Codice    int
}

func (e *Errore) Error() string { return e.Messaggio }

func errore(codice int, formato string, arg ...any) *Errore {
	return &Errore{Messaggio: fmt.Sprintf(formato, arg...), Codice: codice}
}

// leggiJSON legge un file JSON; nil se manca o non è JSON valido.
func leggiJSON(percorso string) any {
	dati, err := os.ReadFile(percorso)
	if err != nil {
		return nil
	}
	var v any
	if json.Unmarshal(dati, &v) != nil {
		return nil
	}
	return v
}

// scriviJSON scrive v con rientro di due spazi, senza sostituire i caratteri non ASCII né < > &.
func scriviJSON(percorso string, v any) error {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return err
	}
	return os.WriteFile(percorso, bytes.TrimRight(b.Bytes(), "\n"), 0o644)
}

func comeMappa(v any) map[string]any {
	m, _ := v.(map[string]any)
	return m
}

func comeLista(v any) []any {
	l, _ := v.([]any)
	return l
}

func comeTesto(v any) (string, bool) {
	s, ok := v.(string)
	return s, ok
}

// vuoto: None, "", [] e {}. Zero e false non sono vuoti.
func vuoto(v any) bool {
	switch x := v.(type) {
	case nil:
		return true
	case string:
		return x == ""
	case []any:
		return len(x) == 0
	case map[string]any:
		return len(x) == 0
	}
	return false
}

// vero segue la verità di Python: nil, "", 0, false, [] e {} sono falsi.
func vero(v any) bool {
	switch x := v.(type) {
	case bool:
		return x
	case float64:
		return x != 0
	}
	return !vuoto(v)
}

// tipo dà lo stesso nome che darebbe lo script Python, per confrontare i tipi delle due parti.
func tipo(v any) string {
	switch v.(type) {
	case nil:
		return "NoneType"
	case bool:
		return "bool"
	case float64:
		return "numero"
	case string:
		return "str"
	case []any:
		return "list"
	case map[string]any:
		return "dict"
	}
	return fmt.Sprintf("%T", v)
}

// intero: un numero o un testo di sole cifre diventa un intero; il resto resta com'è.
func intero(v any) any {
	switch x := v.(type) {
	case float64:
		return int64(x)
	case int64:
		return x
	case int:
		return int64(x)
	case string:
		if x != "" && strings.Trim(x, "0123456789") == "" {
			if n, err := strconv.ParseInt(x, 10, 64); err == nil {
				return n
			}
		}
	}
	return v
}

// chiave rende confrontabili come chiavi di mappa gli id letti dai JSON (350, 350.0, "350").
func chiave(v any) string {
	switch x := intero(v).(type) {
	case nil:
		return "\x00nil"
	case int64:
		return strconv.FormatInt(x, 10)
	case string:
		return x
	default:
		return dumpPy(x)
	}
}

// dumpPy scrive un valore come json.dumps(ensure_ascii=False) di Python: ", " e ": " fra gli
// elementi, caratteri non ASCII lasciati come sono. Go non conserva l'ordine delle chiavi del
// file: si scrivono in ordine alfabetico, ma «it» per prima, perché negli esempi tagliati a 200
// caratteri resti il testo italiano e non quello inglese.
func dumpPy(v any) string {
	var b strings.Builder
	scriviPy(&b, v)
	return b.String()
}

func scriviPy(b *strings.Builder, v any) {
	switch x := v.(type) {
	case nil:
		b.WriteString("null")
	case bool:
		if x {
			b.WriteString("true")
		} else {
			b.WriteString("false")
		}
	case float64:
		b.WriteString(numeroPy(x))
	case int64:
		b.WriteString(strconv.FormatInt(x, 10))
	case int:
		b.WriteString(strconv.Itoa(x))
	case string:
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		enc.SetEscapeHTML(false)
		_ = enc.Encode(x)
		b.WriteString(strings.TrimRight(buf.String(), "\n"))
	case []any:
		b.WriteString("[")
		for i, e := range x {
			if i > 0 {
				b.WriteString(", ")
			}
			scriviPy(b, e)
		}
		b.WriteString("]")
	case map[string]any:
		chiavi := make([]string, 0, len(x))
		for k := range x {
			chiavi = append(chiavi, k)
		}
		sort.Slice(chiavi, func(i, j int) bool {
			if (chiavi[i] == "it") != (chiavi[j] == "it") {
				return chiavi[i] == "it"
			}
			return chiavi[i] < chiavi[j]
		})
		b.WriteString("{")
		for i, k := range chiavi {
			if i > 0 {
				b.WriteString(", ")
			}
			scriviPy(b, k)
			b.WriteString(": ")
			scriviPy(b, x[k])
		}
		b.WriteString("}")
	default:
		scriviPy(b, fmt.Sprint(x))
	}
}

// numeroPy: 350 per un intero, 74.6 per un decimale, come Python.
func numeroPy(f float64) string {
	if f == math.Trunc(f) && math.Abs(f) < 1e15 {
		return strconv.FormatInt(int64(f), 10)
	}
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// strPy: il valore come lo scrive str() di Python, usato per id e nomi nei messaggi e nei nomi dei file.
func strPy(v any) string {
	switch x := v.(type) {
	case nil:
		return "None"
	case bool:
		if x {
			return "True"
		}
		return "False"
	case string:
		return x
	}
	return dumpPy(v)
}

// copia: copia profonda di un valore letto da JSON.
func copia(v any) any {
	switch x := v.(type) {
	case map[string]any:
		m := make(map[string]any, len(x))
		for k, e := range x {
			m[k] = copia(e)
		}
		return m
	case []any:
		l := make([]any, len(x))
		for i, e := range x {
			l[i] = copia(e)
		}
		return l
	}
	return v
}

func ordinaPerTesto(chiavi []string) []string {
	sort.Strings(chiavi)
	return chiavi
}
