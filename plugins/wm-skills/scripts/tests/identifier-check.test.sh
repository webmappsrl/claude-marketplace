#!/usr/bin/env bash
# Test di identifier-check.sh. Lavora in una cartella temporanea, mai nel repo vero.
#
# Nota sulla forma: i file di prova si scrivono con `write <nome>` su una riga a sé e non
# dentro `"$(...)"`. Un heredoc dentro una sostituzione di comando rompe il parsing di
# bash appena il testo contiene un apostrofo — e qui il testo è italiano, quindi ne è pieno.
set -uo pipefail
HERE="$(cd "$(dirname "$0")" && pwd)"
SCRIPT="$HERE/../identifier-check.sh"
FAILED=0

WORK="$(mktemp -d)"; trap 'rm -rf "$WORK"' EXIT

check() { # nome, atteso, ottenuto
  if [ "$2" = "$3" ]; then printf '✔ %s\n' "$1"
  else printf '✘ %s\n   atteso:   %s\n   ottenuto: %s\n' "$1" "$2" "$3"; FAILED=$((FAILED+1)); fi
}

write() { cat > "$WORK/$1"; }

# Gli identificatori segnalati per un file, in ordine alfabetico e senza doppioni.
flagged() { # <nome file>
  "$SCRIPT" "$WORK/$1" 2>&1 | names
}

# Gli identificatori segnalati per un evento di hook già scritto in un file.
flagged_event() { # <nome file col json>
  "$SCRIPT" < "$WORK/$1" 2>&1 | names
}

names() {
  awk -F'→' '/→/ { split($2, part, "  "); gsub(/^[ \t]+|[ \t]+$/, "", part[2]); print part[2] }' \
    | sort -u | paste -sd, -
}

code() { # <nome file>
  "$SCRIPT" "$WORK/$1" >/dev/null 2>&1
  echo "$?"
}

# ── quello che il controllo deve vedere ─────────────────────────────────────

write a.ts <<'EOF'
const elencoVoci = []
EOF
check "una variabile italiana è segnalata" "elencoVoci" "$(flagged a.ts)"

write b.ts <<'EOF'
function calcolaPrezzo(amount) { return amount }
EOF
check "una funzione italiana è segnalata" "calcolaPrezzo" "$(flagged b.ts)"

write c.php <<'EOF'
<?php
$cliente = $request->user();
EOF
check "una variabile PHP con il dollaro" '$cliente' "$(flagged c.php)"

write d.ts <<'EOF'
type Voce = { titolo: string; corpo: string }
EOF
check "un tipo e i suoi campi" "corpo,titolo,Voce" "$(flagged d.ts)"

write e.php <<'EOF'
<?php
return ['prezzo_totale' => 0, 'quote_total' => 1];
EOF
check "una chiave di array, che è una stringa ma nomina un campo" "'prezzo_totale'" "$(flagged e.php)"

write e2.php <<'EOF'
<?php
Schema::table('quotes', function ($table) { $table->decimal('prezzo'); });
EOF
check "una colonna di una migration" "'prezzo'" "$(flagged e2.php)"

# ── quello che NON deve vedere ──────────────────────────────────────────────

write f.ts <<'EOF'
// Il corpo della voce, con il titolo e l'elenco dei prezzi.
/* anche qui: indice, riga, colonna */
const body = ''
EOF
check "un commento in italiano non è un identificatore" "" "$(flagged f.ts)"

write g.ts <<'EOF'
const label = 'il prezzo totale del preventivo'
const other = "nessuna voce trovata"
EOF
check "una stringa in italiano non è un identificatore" "" "$(flagged g.ts)"

write h.ts <<'EOF'
const head = /^>\s*Stato:\s*(.+?)\s*$/m.exec(md)
EOF
check "una regex che contiene una parola italiana" "" "$(flagged h.ts)"

write i.tsx <<'EOF'
const view = <Text>Nessun overview sotto docs/features/</Text>
EOF
check "il testo dentro un elemento JSX" "" "$(flagged i.tsx)"

write j.tsx <<'EOF'
const x = a > bTitolo && c < d
EOF
check "il codice con i segni di maggiore e minore resta intero" "bTitolo" "$(flagged j.tsx)"

write k.ts <<'EOF'
const userTracks = getUserTracks()
function buildQuoteTotal(quote) { return quote.amount }
EOF
check "un nome inglese non è segnalato" "" "$(flagged k.ts)"

write l.php <<'EOF'
<?php
$comune = $address->municipality;
EOF
check "una parola dell allowlist passa" "" "$(flagged l.php)"

write m.ts <<'EOF'
const getUno = 1
EOF
check "un token di tre lettere non fa rumore" "" "$(flagged m.ts)"

# ── il markdown: contano solo i blocchi di codice ───────────────────────────

write n.md <<'EOF'
Il task introduce il calcolo del prezzo totale del preventivo.

```php
function calcolaTotale($quote) { return $quote->amount; }
```
EOF
check "un nome italiano dentro un blocco di codice del piano" "calcolaTotale" "$(flagged n.md)"

write o.md <<'EOF'
Il requisito dice che il prezzo e lo stato della richiesta si leggono dall'elenco.
EOF
check "la prosa italiana del piano, fuori dai blocchi, è lecita" "" "$(flagged o.md)"

write o2.md <<'EOF'
Esempio di overview:

```markdown
## Requisiti
- <a id="r1"></a>**R1** — il prezzo totale compare nell elenco delle voci
```
EOF
check "un blocco markdown di esempio non è codice" "" "$(flagged o2.md)"

# ── la modalità hook: solo il testo appena scritto ───────────────────────────

if command -v jq >/dev/null 2>&1; then
  write write.json <<'EOF'
{"tool_name":"Write","tool_input":{"file_path":"/x/Quote.php","content":"<?php\nfunction calcolaPrezzo() {}\n"}}
EOF
  check "un evento Write" "calcolaPrezzo" "$(flagged_event write.json)"

  write edit.json <<'EOF'
{"tool_name":"Edit","tool_input":{"file_path":"/x/Quote.php","old_string":"$customer","new_string":"$cliente"}}
EOF
  check "un evento Edit guarda new_string, non il file" '$cliente' "$(flagged_event edit.json)"

  write multi.json <<'EOF'
{"tool_name":"MultiEdit","tool_input":{"file_path":"/x/a.ts","edits":[{"new_string":"const primaRiga = 1"},{"new_string":"const secondaRiga = 2"}]}}
EOF
  check "un evento MultiEdit guarda tutte le modifiche" "primaRiga,secondaRiga" "$(flagged_event multi.json)"

  write read.json <<'EOF'
{"tool_name":"Read","tool_input":{"file_path":"/x/a.ts"}}
EOF
  "$SCRIPT" < "$WORK/read.json" >/dev/null 2>&1
  check "un evento senza testo scritto non dice niente" "0" "$?"
else
  printf '… modalità hook saltata: manca jq\n'
fi

# ── gli exit code ───────────────────────────────────────────────────────────

write p.ts <<'EOF'
const userTracks = []
EOF
check "un file pulito esce con 0" "0" "$(code p.ts)"

write q.ts <<'EOF'
const elencoVoci = []
EOF
check "un rilievo esce con 2, il codice con cui un hook parla al modello" "2" "$(code q.ts)"

echo
if [ "$FAILED" -eq 0 ]; then echo "Tutti i test passano."; else echo "$FAILED test falliti."; fi
exit "$FAILED"
