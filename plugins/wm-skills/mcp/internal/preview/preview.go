// Package preview costruisce il testo che il dev legge prima di autorizzare una
// scrittura. La differenza è calcolata sui dati veri letti da Orchestrator, non
// ricostruita a memoria.
package preview

import (
	"fmt"
	"sort"
	"strings"
)

// Diff descrive cosa cambierebbe applicando requested allo stato current.
func Diff(current, requested map[string]any) string {
	names := make([]string, 0, len(requested))
	for k := range requested {
		names = append(names, k)
	}
	sort.Strings(names)

	var changed []string
	for _, n := range names {
		before := format(current[n])
		after := format(requested[n])
		if before == after {
			continue
		}
		changed = append(changed, fmt.Sprintf("  %-18s %s  →  %s", n, before, after))
	}

	if len(changed) == 0 {
		return "Nessuna modifica: i valori richiesti coincidono con quelli attuali."
	}

	var b strings.Builder
	b.WriteString("ANTEPRIMA — nulla è stato scritto\n")
	b.WriteString(strings.Join(changed, "\n"))
	b.WriteString("\n\nPer applicare, richiama lo stesso tool con confirm: true.")
	return b.String()
}

// NewResource descrive una creazione, dove non esiste uno stato precedente.
func NewResource(requested map[string]any) string {
	names := make([]string, 0, len(requested))
	for k := range requested {
		names = append(names, k)
	}
	sort.Strings(names)

	var b strings.Builder
	b.WriteString("ANTEPRIMA — nulla è stato creato\n")
	for _, n := range names {
		fmt.Fprintf(&b, "  %-18s %s\n", n, format(requested[n]))
	}
	b.WriteString("\nPer creare, richiama lo stesso tool con confirm: true.")
	return b.String()
}

func format(v any) string {
	switch t := v.(type) {
	case nil:
		return "(vuoto)"
	case string:
		if t == "" {
			return "(vuoto)"
		}
		// Il troncamento va fatto su un confine di rune: tagliare per byte
		// (t[:120]) può spezzare un carattere multi-byte (es. accentato) a
		// metà e produrre una sequenza UTF-8 non valida.
		if runes := []rune(t); len(runes) > 120 {
			return fmt.Sprintf("%q… (%d caratteri)", string(runes[:120]), len(runes))
		}
		return fmt.Sprintf("%q", t)
	case float64:
		if t == float64(int64(t)) {
			return fmt.Sprintf("%d", int64(t))
		}
		return fmt.Sprintf("%g", t)
	default:
		return fmt.Sprintf("%v", v)
	}
}
