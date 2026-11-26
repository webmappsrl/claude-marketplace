#!/usr/bin/env bash
# Trova gli identificatori italiani nel codice appena scritto.
#
# La regola «gli identificatori sono in inglese» è scritta in tre posti di wm-plan/SKILL.md
# e nel CLAUDE.md del repo, e non basta: una regola si legge all'inizio del lavoro e si
# perde dopo qualche ora passata a parlare italiano. Questo invece gira a ogni scrittura.
#
# Due modi di invocarlo:
#
#   identifier-check.sh <file>...        # a mano, o da un test: analizza i file interi
#   identifier-check.sh                  # da un hook PostToolUse: legge il JSON su stdin
#
# Nel secondo caso analizza **solo il testo appena scritto** (`content` di Write,
# `new_string` di Edit), non tutto il file: il codice che c'era già non è compito di
# questa scrittura, e segnalarlo trasformerebbe ogni modifica in un refactor.
#
# Exit code 2 con il messaggio su stderr: è il modo con cui un hook PostToolUse parla al
# modello. Non annulla la scrittura — è già avvenuta — gli chiede di correggerla.

set -uo pipefail

HERE="$(cd "$(dirname "$0")" && pwd)"
SCAN="$HERE/lib/identifier-scan.pl"
export WORDS="$HERE/data/italian-words.txt"
export ALLOWLIST="$HERE/data/identifier-allowlist.txt"

for needed in "$SCAN" "$WORDS"; do
  [ -f "$needed" ] || { echo "identifier-check: manca $needed" >&2; exit 1; }
done

FOUND=""

if [ "$#" -gt 0 ]; then
  for path in "$@"; do
    [ -f "$path" ] || continue
    hits="$(perl "$SCAN" "$path" < "$path" | sort -u)"
    [ -n "$hits" ] && FOUND="$FOUND$(printf '%s\n' "$hits" | sed "s|^|$path	|")
"
  done
else
  # L'hook passa su stdin il JSON dell'evento. Serve il path — per l'estensione — e il
  # testo scritto: `content` per Write, `new_string` per Edit, `edits[]` per MultiEdit.
  EVENT="$(cat)"

  if ! command -v jq >/dev/null 2>&1; then
    # Un controllo che tace quando non sa leggere l'evento è peggio di uno che non c'è:
    # si crede che stia girando. Lo dice, e lascia passare.
    echo "identifier-check: manca jq, il controllo sugli identificatori non è stato eseguito." >&2
    exit 0
  fi

  FILE_PATH="$(printf '%s' "$EVENT" | jq -r '.tool_input.file_path // ""')"
  WRITTEN="$(printf '%s' "$EVENT" | jq -r '
    [ .tool_input.content,
      .tool_input.new_string,
      (.tool_input.edits // [] | map(.new_string) | join("\n")) ]
    | map(select(. != null and . != "")) | join("\n")')"

  [ -n "$FILE_PATH" ] || exit 0
  [ -n "$WRITTEN" ] || exit 0

  hits="$(printf '%s' "$WRITTEN" | perl "$SCAN" "$FILE_PATH" | sort -u)"
  [ -n "$hits" ] && FOUND="$(printf '%s\n' "$hits" | sed "s|^|$FILE_PATH	|")
"
fi

[ -n "$FOUND" ] || exit 0

{
  echo "Identificatori italiani nel codice appena scritto:"
  echo
  printf '%s' "$FOUND" | awk -F'\t' 'NF >= 3 { printf "  %s  →  %s  (parola italiana: %s)\n", $1, $3, $2 }' | sort -u
  echo
  cat <<'MSG'
Nel codice gli identificatori si scrivono sempre in inglese, anche quando il dominio è
italiano: `$quote` e non `$preventivo`, `getUserTracks()` e non `getTracceUtente()`. Vale
per variabili, funzioni, metodi, classi, campi del database, chiavi di array e di
configurazione, nomi di file e di branch. L'italiano resta nei commenti, nei docblock e
nei testi rivolti a persone — quelli questo controllo non li guarda nemmeno.

Rinomina ora quello che hai appena scritto. Se invece il nome italiano è obbligato — una
colonna che esiste già, una chiave imposta da un servizio terzo, un termine amministrativo
senza equivalente esatto — aggiungilo con il suo motivo a:

  plugins/wm-skills/scripts/data/identifier-allowlist.txt

Se è un falso positivo, cioè la parola è anche inglese, la correzione va al dizionario,
data/italian-words.txt, non all'allowlist.

Non rinominare gli identificatori italiani che hai solo letto e non scritto tu: quello è
un refactor fuori dallo scope del ticket, e va in notes.md sotto «Follow-up».
MSG
} >&2

exit 2
