#!/usr/bin/env bash
# Il formato della riga di un requisito è scritto in tre posti, e devono dire la stessa cosa:
#
#   1. plugins/wm-skills/skills/wm-plan/SKILL.md      lo prescrive  (la fonte autoritativa)
#   2. plugins/wm-skills/scripts/feature-docs-check.sh  lo riconosce con una regex awk
#   3. plugins/wm-skills/mods/wm-plan-anchors/documents.ts  lo indicizza per il pane della mod
#
# Se la forma cambia in uno solo, gli altri due smettono di riconoscere i requisiti senza dire
# niente: lo script di controllo dei documenti passa su un overview che non legge, e il pane
# della mod resta vuoto. Questo controllo fa fallire la CI invece.
#
# Metodo: prende le righe d'esempio dallo SKILL.md — non le riscrive qui, o ci sarebbe un quarto
# posto da tenere allineato — e verifica che le due regex le riconoscano. Insieme passa dei
# contro-esempi che NON devono essere riconosciuti: senza quelli, una regex che accetta tutto
# passerebbe il controllo.
#
# LIMITE, da conoscere: verifica che i due parser riconoscano le stesse righe, non che le
# interpretino allo stesso modo. Che il testo estratto, le sottovoci e le continuazioni rientrate
# siano lette uguali non lo dice: quello lo coprono i test del parser (claude plugin test).
set -uo pipefail
cd "$(dirname "$0")/../.."

SKILL="plugins/wm-skills/skills/wm-plan/SKILL.md"
AWK_SRC="plugins/wm-skills/scripts/feature-docs-check.sh"
TS_SRC="plugins/wm-skills/mods/wm-plan-anchors/documents.ts"
ESITO=0

for F in "$SKILL" "$AWK_SRC" "$TS_SRC"; do
  [ -f "$F" ] || { echo "❌ manca $F"; exit 1; }
done

# ── le righe d'esempio, dalla fonte autoritativa ────────────────────────────
ESEMPI="$(grep -oE '^[[:space:]]*(- |#+ )<a id="r[0-9]+"></a>\*\*R[0-9]+\*\*.*$' "$SKILL" \
  | sed 's/^[[:space:]]*//' | sort -u)"
ESEMPI="$ESEMPI
$(grep -oE '`- <a id="r[0-9]+"></a>\*\*R[0-9]+\*\* — \[UX\][^`]*`' "$SKILL" | tr -d '`' | sort -u)"
ESEMPI="$(printf '%s\n' "$ESEMPI" | grep -v '^$' | sort -u)"

QUANTI="$(printf '%s\n' "$ESEMPI" | grep -c . || true)"
if [ "$QUANTI" -eq 0 ]; then
  echo "❌ nessuna riga di requisito trovata in $SKILL"
  echo "   → il formato è cambiato e questo controllo è diventato cieco: aggiorna l'estrazione qui"
  exit 1
fi

# ── il parser awk, eseguito davvero ─────────────────────────────────────────
# Non si ricopia qui la sua regex: una copia si scollega dall'originale senza che nessuno se ne
# accorga, ed è esattamente il guasto che questo controllo deve impedire. Si estrae il programma
# awk dallo script e lo si fa girare su un overview costruito con le righe prescritte.
REQ_AWK="$(awk "/^REQ_AWK='/{dentro=1; sub(/^REQ_AWK='/, \"\"); } dentro { if (/'$/ && !/^REQ_AWK/) { sub(/'$/, \"\"); print; exit } print }" "$AWK_SRC")"
[ -n "$REQ_AWK" ] || { echo "❌ non trovo REQ_AWK in $AWK_SRC"; exit 1; }

awk_riconosce() { # <riga> → 0 se il parser le dà un identificativo
  printf '## Requisiti\n\n%s\n' "$1" | awk "$REQ_AWK" 2>/dev/null | grep -q "^T\t"
}

# JavaScript: l'indicizzatore non ha una regex per «la riga di un requisito» — segna le àncore
# e gli heading, e una riga è indicizzata se porta un'àncora. Si estrae quella costante e si
# verifica che la riga la contenga: è ciò che decide se quel requisito finirà nel pane.
TS_RE="$(grep -oE 'const ANCHOR = /.*/$' "$TS_SRC" | sed 's|^.* = /||; s|/$||')"
[ -n "$TS_RE" ] || { echo "❌ non trovo la costante ANCHOR in $TS_SRC"; exit 1; }
TS_ERE="$(printf '%s' "$TS_RE" \
  | sed 's|\\s|[[:space:]]|g; s|\\d|[0-9]|g; s|\\/|/|g; s|\\\*|[*]|g')"

prova() { # <ere> <riga> → 0 se riconosciuta
  printf '%s\n' "$2" | grep -qE "$1"
}

# ── le righe prescritte devono essere riconosciute da entrambi ──────────────
while IFS= read -r RIGA; do
  [ -n "$RIGA" ] || continue
  awk_riconosce "$RIGA" || {
    echo "❌ $AWK_SRC non riconosce una riga che $SKILL prescrive:"
    echo "   $RIGA"
    echo "   → allinea la regex in $AWK_SRC"
    ESITO=1
  }
  prova "$TS_ERE" "$RIGA" || {
    echo "❌ $TS_SRC non riconosce una riga che $SKILL prescrive:"
    echo "   $RIGA"
    echo "   → allinea ANCHOR in $TS_SRC"
    ESITO=1
  }
done <<< "$ESEMPI"

# ── i contro-esempi non devono essere riconosciuti da nessuno dei due ───────
# Senza questi, una regex che accetta qualunque riga passerebbe il controllo.
CONTRO=(
  '- [ ] il vecchio formato a checkbox, senza identificativo'
  '- **R1** — identificativo senza ancora'
  '## Requisiti'
)
for RIGA in "${CONTRO[@]}"; do
  if awk_riconosce "$RIGA"; then
    echo "❌ $AWK_SRC riconosce una riga che non è un requisito:"
    echo "   $RIGA"
    ESITO=1
  fi
  if prova "$TS_ERE" "$RIGA"; then
    echo "❌ $TS_SRC riconosce una riga che non è un requisito:"
    echo "   $RIGA"
    ESITO=1
  fi
done

[ "$ESITO" -eq 0 ] && echo "✔ formato del requisito: $QUANTI righe da $SKILL, riconosciute da entrambi i parser."
exit $ESITO
