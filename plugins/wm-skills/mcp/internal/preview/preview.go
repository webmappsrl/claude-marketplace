// Package preview costruisce il testo che il dev legge prima di autorizzare una
// scrittura. La differenza è calcolata sui dati veri letti da Orchestrator, non
// ricostruita a memoria, e mostra ogni campo per intero: un'anteprima troncata
// fa approvare alla cieca.
package preview

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"unicode/utf8"
)

// Diff descrive cosa cambierebbe applicando requested allo stato current.
// htmlFields elenca i nomi dei campi resi da un editor visuale su
// Orchestrator (per esempio "description" e "customer_request" di una
// story): solo quelli, esplicitamente passati dal chiamante, vengono resi
// come testo leggibile invece che mostrati grezzi. Un campo chiamato allo
// stesso modo ma non HTML — la description Markdown di un tag, per
// esempio — resta invariata se il chiamante non lo dichiara.
func Diff(current, requested map[string]any, htmlFields ...string) string {
	html := toSet(htmlFields)
	names := sortedKeys(requested)

	var blocks []string
	for _, n := range names {
		rawBefore, rawAfter := current[n], requested[n]
		if rawEqual(rawBefore, rawAfter) {
			continue
		}
		before := format(n, rawBefore, html)
		after := format(n, rawAfter, html)

		block := n + "\n"
		if before == after && html[n] {
			// il valore grezzo è cambiato ma la resa leggibile no: è un
			// cambio di solo markup (es. <b> → <strong>), non di contenuto.
			// Va segnalato: altrimenti l'anteprima sembra dire che non
			// cambia nulla mentre la PATCH parte comunque.
			block += "  cambia solo la formattazione (HTML), il testo resta uguale\n"
		}
		block += labelled("prima", before) + "\n" + labelled("dopo", after)
		blocks = append(blocks, block)
	}

	if len(blocks) == 0 {
		return "Nessuna modifica: i valori richiesti coincidono con quelli attuali."
	}

	var b strings.Builder
	b.WriteString("ANTEPRIMA — nulla è stato scritto\n\n")
	if w := lossWarning(current["description"], requested["description"]); w != "" {
		b.WriteString(w + "\n\n")
	}
	b.WriteString(strings.Join(blocks, "\n\n"))
	b.WriteString("\n\nPer applicare, richiama lo stesso tool con confirm: true.")
	return b.String()
}

// NewResource descrive una creazione, dove non esiste uno stato precedente.
// htmlFields ha lo stesso significato che in Diff.
func NewResource(requested map[string]any, htmlFields ...string) string {
	html := toSet(htmlFields)
	var b strings.Builder
	b.WriteString("ANTEPRIMA — nulla è stato creato\n\n")
	for _, n := range sortedKeys(requested) {
		b.WriteString(labelled(n, format(n, requested[n], html)) + "\n")
	}
	b.WriteString("\nPer creare, richiama lo stesso tool con confirm: true.")
	return b.String()
}

// rawEqual confronta i valori grezzi (quelli che finiscono davvero nella
// PATCH), non la loro resa leggibile: un campo è cambiato se cambia il dato
// inviato a Orchestrator, anche quando due rese diverse producono lo stesso
// testo (o viceversa).
func rawEqual(a, b any) bool {
	return reflect.DeepEqual(a, b)
}

func toSet(names []string) map[string]bool {
	set := make(map[string]bool, len(names))
	for _, n := range names {
		set[n] = true
	}
	return set
}

// lossWarning avvisa quando la description nuova è meno della metà di quella
// attuale: è il segno tipico di una sostituzione usata al posto di
// un'aggiunta. Non blocca nulla, fa sì che l'errore si noti.
func lossWarning(before, after any) string {
	b, okB := before.(string)
	a, okA := after.(string)
	if !okB || !okA {
		return ""
	}
	nb, na := utf8.RuneCountInString(b), utf8.RuneCountInString(a)
	if nb == 0 || na*2 >= nb {
		return ""
	}
	return fmt.Sprintf("⚠️ description: %d → %d caratteri — si perde il %d%% del testo attuale.", nb, na, 100-na*100/nb)
}

// labelled scrive «etichetta: valore» sulla stessa riga se il valore è una
// riga sola, altrimenti l'etichetta e sotto il valore rientrato.
func labelled(label, value string) string {
	if !strings.Contains(value, "\n") {
		return fmt.Sprintf("  %s: %s", label, value)
	}
	return fmt.Sprintf("  %s:\n    %s", label, strings.ReplaceAll(value, "\n", "\n    "))
}

func sortedKeys(m map[string]any) []string {
	names := make([]string, 0, len(m))
	for k := range m {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

func format(field string, v any, html map[string]bool) string {
	switch t := v.(type) {
	case nil:
		return "(vuoto)"
	case string:
		if html[field] {
			t = Readable(t)
		}
		if t == "" {
			return "(vuoto)"
		}
		return t
	case float64:
		if t == float64(int64(t)) {
			return fmt.Sprintf("%d", int64(t))
		}
		return fmt.Sprintf("%g", t)
	default:
		return fmt.Sprintf("%v", v)
	}
}
