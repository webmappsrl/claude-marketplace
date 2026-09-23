package preview

import (
	"html"
	"regexp"
	"strings"
)

var (
	tagRe        = regexp.MustCompile(`(?s)<(/?)([a-zA-Z0-9]+)([^>]*)>`)
	hrefRe       = regexp.MustCompile(`(?i)href\s*=\s*["']([^"']*)["']`)
	spacesRe     = regexp.MustCompile(`[ \t\r\n]+`)
	blankRunsRe  = regexp.MustCompile(`\n{3,}`)
	multiSpaceRe = regexp.MustCompile(` {2,}`)
)

// knownTags sono i nomi HTML che Readable interpreta come struttura. Un
// nome fuori da questo elenco non è un tag: è testo dell'utente che
// contiene «<» e «>» (es. "b<5 e c>10", "List<String>") e va scritto per
// intero, letteralmente.
var knownTags = map[string]bool{
	"a": true, "abbr": true, "b": true, "blockquote": true, "br": true,
	"code": true, "del": true, "div": true, "em": true,
	"h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true,
	"hr": true, "i": true, "img": true, "li": true, "mark": true, "ol": true,
	"p": true, "pre": true, "s": true, "span": true, "strong": true,
	"sub": true, "sup": true, "table": true, "tbody": true, "td": true,
	"th": true, "thead": true, "tr": true, "u": true, "ul": true,
}

// span identifica, con offset di byte sulla stringa finale costruita da
// Readable, il tratto scritto mentre eravamo dentro <pre>: serve al passo
// finale di trim per riconoscere le righe di codice e lasciarle intatte,
// perché lì gli spazi sono indentazione, non spaziatura da collassare.
type span struct{ start, end int }

// Readable rende un campo HTML come testo da leggere in una sessione di
// terminale: titoli in maiuscolo su una riga a sé, elenchi con «- » e rientro,
// paragrafi separati da una riga vuota, link come testo con l'indirizzo fra
// parentesi, entità decodificate. Serve all'anteprima: il dev deve vedere
// cosa verrà scritto senza leggere tag.
func Readable(s string) string {
	var b strings.Builder
	var hrefs []string
	var verbatimSpans []span
	heading := false
	verbatim := false
	depth := 0
	last := 0

	writeText := func(raw string) {
		decoded := html.UnescapeString(raw)
		if verbatim {
			if decoded == "" {
				return
			}
			start := b.Len()
			b.WriteString(decoded)
			verbatimSpans = append(verbatimSpans, span{start, b.Len()})
			return
		}
		t := spacesRe.ReplaceAllString(decoded, " ")
		t = strings.ReplaceAll(t, " ", " ")
		if heading {
			t = strings.ToUpper(t)
		}
		b.WriteString(t)
	}

	for _, m := range tagRe.FindAllStringSubmatchIndex(s, -1) {
		writeText(s[last:m[0]])
		last = m[1]
		closing := s[m[2]:m[3]] == "/"
		name := strings.ToLower(s[m[4]:m[5]])
		attrs := s[m[6]:m[7]]

		if !knownTags[name] {
			// non è un tag noto: si scrive il testo così com'è, entità
			// comprese, non si interpreta come struttura.
			writeText(s[m[0]:m[1]])
			continue
		}

		switch name {
		case "h1", "h2", "h3", "h4", "h5", "h6":
			heading = !closing
			if depth > 0 {
				// dentro un elenco l'intestazione resta sulla stessa riga
				// della voce: solo lo spazio per separarla da ciò che segue.
				if closing {
					b.WriteString(" ")
				}
			} else {
				b.WriteString("\n\n")
			}
		case "p", "div", "table":
			b.WriteString("\n\n")
		case "pre":
			b.WriteString("\n\n")
			verbatim = !closing
		case "br", "tr":
			b.WriteString("\n")
		case "ul", "ol":
			// solo la profondità: l'a capo lo scrive ogni <li>, altrimenti
			// fra una voce e la sua sottolista resterebbe una riga vuota
			if closing {
				depth--
			} else {
				depth++
			}
		case "li":
			if !closing {
				b.WriteString("\n" + strings.Repeat("  ", max(depth-1, 0)) + "- ")
			}
		case "td", "th":
			if !closing {
				b.WriteString(" | ")
			}
		case "a":
			if !closing {
				href := ""
				if hm := hrefRe.FindStringSubmatch(attrs); hm != nil {
					href = hm[1]
				}
				hrefs = append(hrefs, href)
			} else if n := len(hrefs); n > 0 {
				if hrefs[n-1] != "" {
					b.WriteString(" (" + hrefs[n-1] + ")")
				}
				hrefs = hrefs[:n-1]
			}
		}
	}
	writeText(s[last:])

	str := b.String()
	lines := strings.Split(str, "\n")
	offset := 0
	for i, l := range lines {
		lineStart, lineEnd := offset, offset+len(l)
		offset = lineEnd + 1 // +1: il separatore '\n' consumato dallo Split

		verbatimLine := false
		for _, sp := range verbatimSpans {
			if lineStart < sp.end && lineEnd > sp.start {
				verbatimLine = true
				break
			}
		}
		if verbatimLine {
			continue // riga di codice dentro <pre>: si conserva com'è
		}

		lines[i] = strings.TrimRight(l, " ")
		// il rientro degli elenchi va conservato, gli spazi iniziali del
		// testo che segue un tag no. Nel contenuto, invece, due o più spazi
		// consecutivi si riducono a uno solo: capita quando un tag chiuso
		// (es. </h3> dentro un <li>) aggiunge già uno spazio di separazione
		// e il testo che segue ne porta un altro suo, letterale.
		trimmed := strings.TrimLeft(lines[i], " ")
		indent := lines[i][:len(lines[i])-len(trimmed)]
		if strings.HasPrefix(trimmed, "- ") {
			lines[i] = indent + "- " + multiSpaceRe.ReplaceAllString(trimmed[2:], " ")
		} else {
			lines[i] = multiSpaceRe.ReplaceAllString(trimmed, " ")
		}
	}
	out := blankRunsRe.ReplaceAllString(strings.Join(lines, "\n"), "\n\n")
	return strings.Trim(out, "\n")
}
