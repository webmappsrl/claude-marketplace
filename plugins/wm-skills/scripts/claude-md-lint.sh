#!/usr/bin/env bash
# Lint meccanico di un CLAUDE.md: tutto ciò che, ripetuto, dà lo stesso risultato.
#
# Uso:  claude-md-lint.sh <percorso/CLAUDE.md> [repo vicino]...
#
# Output: un rilievo per riga, campi separati da TAB
#         <severita>\t<id>\t<messaggio>
#   error  fatto verificabile e falso: non si contraddice senza una prova
#   warn   difetto strutturale certo
#   hint   euristica: riconosce per forma, non per senso — contestabile con motivazione
#
# Exit:   0 nessun error, 1 almeno un error, 2 uso sbagliato
#
# Silenziare un rilievo: una riga con il suo id in <repo>/.claude/claude-md-lint-ignore
# Resta nell'output come `suppressed`, così si vede ancora ma non va rispiegato ogni volta.

set -uo pipefail

[ $# -ge 1 ] || { echo "uso: $0 <percorso/CLAUDE.md> [repo vicino]..." >&2; exit 2; }
FILE="$1"; shift
VICINI=("$@")
[ -f "$FILE" ] || { echo "error	file-assente	$FILE non esiste"; exit 1; }
REPO="$(cd "$(dirname "$FILE")" && pwd)"
IGNORE="$REPO/.claude/claude-md-lint-ignore"

ERRORI=0
emetti() { # severita, id, messaggio
  local sev="$1" id="$2" msg="$3"
  if [ -f "$IGNORE" ] && grep -qxF "$id" "$IGNORE" 2>/dev/null; then
    printf 'suppressed\t%s\t%s\n' "$id" "$msg"; return
  fi
  printf '%s\t%s\t%s\n' "$sev" "$id" "$msg"
  [ "$sev" = error ] && ERRORI=$((ERRORI+1))
  return 0
}

# ---------- stato ----------
BYTE=$(wc -c < "$FILE" | tr -d ' ')
RIGHE=$(wc -l < "$FILE" | tr -d ' ')
SEZIONI=$(grep -c '^## ' "$FILE")
PRIMA=$(grep -m1 '^## ' "$FILE" | sed 's/^## //')
printf 'info\tstato\t%s byte, %s righe, %s sezioni, prima sezione: %s\n' "$BYTE" "$RIGHE" "$SEZIONI" "$PRIMA"
[ "$RIGHE" -gt 200 ] && emetti warn "file-troppo-lungo" "$RIGHE righe, oltre il limite di 200"

# sezione più pesante
awk '/^## /{n=$0; next} {b[n]+=length($0)+1} END{for(s in b) printf "%d\t%s\n", b[s], s}' "$FILE" \
  | sort -rn | head -1 | while IFS=$'\t' read -r peso nome; do
      pct=$(( peso * 100 / (BYTE>0?BYTE:1) ))
      printf 'info\tsezione-piu-pesante\t%s (%s byte, %s%%)\n' "$nome" "$peso" "$pct"
      [ "$pct" -ge 50 ] && printf 'warn\tsezione-fuori-scala\t%s occupa il %s%% del file\n' "$nome" "$pct"
    done

# ---------- indice ----------
grep -o '](docs/knowledge/[^)]*\.md)' "$FILE" | sed 's/](//; s/)//' | sort -u | while read -r f; do
  [ -f "$REPO/$f" ] || emetti error "missing-page:$f" "l'indice cita $f, che non esiste"
done
if [ -d "$REPO/docs/knowledge" ]; then
  ls "$REPO"/docs/knowledge/*.md 2>/dev/null | sed "s|$REPO/||" | sort -u | while read -r f; do
    grep -q "$(basename "$f")" "$FILE" || emetti warn "orphan-page:$f" "$f esiste ma nessuno la cita"
  done
fi
if [ -d "$REPO/docs/features" ]; then
  for d in "$REPO"/docs/features/*/; do
    [ -d "$d" ] || continue
    # una cartella vuota non è un cantiere: git non traccia le directory, quindi
    # è un residuo locale e segnalarla manda a cercare un lavoro che non c'è
    [ -n "$(ls -A "$d" 2>/dev/null)" ] || continue
    s=$(basename "$d"); id=$(echo "$s" | grep -o '^[0-9]\{1,\}')
    [ -n "$id" ] || continue
    grep -rqs "oc:$id" "$FILE" "$REPO/docs/knowledge/" "$REPO/docs/howto/" 2>/dev/null \
      || emetti warn "uncovered-feature:$s" "nessuna conoscenza cita oc:$id"
  done
fi
grep -o '`[a-zA-Z0-9_]\{1,\}/[A-Za-z0-9_/.-]*\.[a-z]\{2,4\}`' "$FILE" | tr -d '`' | sort -u | while read -r f; do
  if [ ! -e "$REPO/$f" ]; then
    trovato=""
    for sub in core src app; do [ -e "$REPO/$sub/$f" ] && trovato="$sub"; done
    [ -n "$trovato" ] || emetti hint "missing-path:$f" "percorso citato fra backtick che non esiste in questo repo"
  fi
done

# ---------- doppio indice ----------
TMPD=$(mktemp -d)
awk -v d="$TMPD" '/^## /{s=$0; gsub(/[^A-Za-z0-9]/,"_",s); next}
     match($0, /oc:[0-9]+/) {print substr($0, RSTART, RLENGTH) >> (d "/" s)}' "$FILE" 2>/dev/null
for a in "$TMPD"/*; do
  [ -f "$a" ] || continue
  for b in "$TMPD"/*; do
    [ -f "$b" ] && [ "$a" \< "$b" ] || continue
    comuni=$(sort -u "$a" | comm -12 - <(sort -u "$b") | wc -l | tr -d ' ')
    if [ "$comuni" -ge 3 ]; then
      na=$(basename "$a" | tr '_' ' '); nb=$(basename "$b" | tr '_' ' ')
      emetti warn "double-index:$na|$nb" "$comuni ticket in comune fra due sezioni: stessa granularità"
    fi
  done
done
rm -rf "$TMPD"

# ---------- scheletro ----------
for parte in "Cos'è questo repo|cos-e-questo-repo" "Regole|regole" "Comandi|comandi" "Convenzioni|convenzioni" "Conoscenza|conoscenza" "Trappole|trappole"; do
  eti="${parte%%|*}"; slug="${parte##*|}"
  grep -qi "^## .*${eti}" "$FILE" || emetti warn "missing-section:$slug" "manca la sezione «${eti}»"
done
case "$PRIMA" in
  *Cos*|*cos*) : ;;
  *) emetti hint "first-section:$PRIMA" "la prima sezione non è l'inquadramento del repo" ;;
esac

# ---------- trappole rimaste nelle pagine ----------
if [ -d "$REPO/docs/knowledge" ]; then
  grep -nE 'va aggiornat|vanno aggiornat|va allineat|aggiorna anche|aggiornare anche|non toccarl|non toccare|non rimuover|non confonder|non ripristinar|non importarl|assicurati|verifica (sempre|prima|che)|controlla (che|anche|sempre|prima|entrambi)|deve (restare|puntare|essere aggiornat|coincidere)|devono (restare|puntare|coincidere|essere)' \
    "$REPO"/docs/knowledge/*.md 2>/dev/null | while IFS= read -r r; do
      loc=$(echo "$r" | cut -d: -f1,2 | sed "s|$REPO/||")
      emetti hint "trap-in-page:$loc" "frase imperativa in una pagina di conoscenza: se è un'istruzione va in .claude/rules/"
    done
fi

# ---------- divieti con effetto esterno fuori dal file ----------
for src in "$REPO"/.claude/rules/*.md "$REPO"/docs/knowledge/*.md; do
  [ -f "$src" ] || continue
  grep -nEi 'a tutti i client|tutti i clienti|a tutti gli altri|multi-tenant|notifica il cliente|pubblicherebbe|pubblicare a tutti|irreversibil|non si revoca|non si annulla' "$src" 2>/dev/null \
    | while IFS= read -r r; do
        loc="$(echo "$src" | sed "s|$REPO/||"):$(echo "$r" | cut -d: -f1)"
        if grep -qEi 'a tutti i client|tutti i clienti|a tutti gli altri|multi-tenant|notifica il cliente|pubblicherebbe|irreversibil|non si revoca' "$FILE"; then
          continue   # il divieto è già anche nel file principale: qui è il dettaglio, ed è corretto
        fi
        emetti hint "divieto-fuori:$loc" "vincolo con possibile effetto esterno fuori dal CLAUDE.md: se lo è, va anche fra le regole in cima"
      done
done

# ---------- il titolo nomina il repo? ----------
TITOLO=$(head -1 "$FILE" | sed 's/^# //')
NOMEREPO=$(basename "$REPO")
echo "$TITOLO" | grep -qi "$NOMEREPO" \
  || emetti warn "titolo-non-nomina-il-repo:$TITOLO" "il titolo non contiene «${NOMEREPO}»: spesso è il residuo del template da cui il repo è nato"

# ---------- una rule che si carica sempre non è path-scoped ----------
for rule in "$REPO"/.claude/rules/*.md; do
  [ -f "$rule" ] || continue
  rn=$(basename "$rule")
  # conta i paths: quanti sono e quanto sono larghi
  larghi=$(awk '/^paths:/{p=1;next} /^---/{p=0} p&&/^ *- /{gsub(/[" -]/,"");print}' "$rule" 2>/dev/null \
    | grep -cE '^[a-zA-Z0-9_]+/\*\*$|^\*\*|^src/\*\*$|^app/\*\*$')
  soggetti=$(grep -c '^- \*\*' "$rule" 2>/dev/null)
  if [ "${larghi:-0}" -ge 2 ] && [ "${soggetti:-0}" -ge 4 ]; then
    emetti warn "rule-troppo-larga:$rn" "$soggetti trappole scollegate con $larghi paths generici: se si carica quasi sempre non è path-scoped, va spezzata o riportata fra le regole"
  fi
done

# ---------- reciprocità coi repo vicini ----------
NOME=$(basename "$REPO")
for v in "${VICINI[@]:-}"; do
  [ -n "$v" ] && [ -f "$v/CLAUDE.md" ] || continue
  vn=$(basename "$v")
  # si cerca in tutto il file tranne l'inquadramento: lì il nome dell'altro repo
  # compare per forza («l'altro prodotto è X») e non vuol dire che ci sia una regola
  senza_intro() { awk '/^## /{sez=$0} sez !~ /[Cc]os.è questo repo/' "$1"; }
  if senza_intro "$v/CLAUDE.md" | grep -qE "\b$NOME\b"; then
    senza_intro "$FILE" | grep -qE "\b$vn\b" \
      || emetti warn "missing-reciprocal:$vn" "$vn dichiara una regola che nomina questo repo, qui non c'è la corrispondente"
  fi
done

# ---------- cifre che invecchiano ----------
grep -nE '\b[0-9]{1,4} (spec|test|file|righe|voci|casi)\b' "$FILE" 2>/dev/null | while IFS= read -r r; do
  n=$(echo "$r" | cut -d: -f1)
  emetti hint "aging-count:CLAUDE.md:$n" "conteggio che invecchia al primo file aggiunto"
done

# ---------- marker per l'hook ----------
MARK="${TMPDIR:-/tmp}/wm-context-doctor-$(echo "$REPO" | shasum | cut -c1-12).run"
date +%s > "$MARK"

[ "$ERRORI" -gt 0 ] && exit 1
exit 0
