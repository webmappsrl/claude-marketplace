#!/usr/bin/env bash
# Il formato della riga di un requisito è scritto in tre posti, e devono dire la stessa cosa:
#
#   1. plugins/wm-skills/skills/wm-plan/SKILL.md      lo prescrive  (la fonte autoritativa)
#   2. plugins/wm-skills/scripts/feature-docs-check.sh  lo riconosce con una regex awk
#   3. plugins/wm-skills/mods/wm-plan-anchors/parse.ts  lo riconosce con una regex JavaScript
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
TS_SRC="plugins/wm-skills/mods/wm-plan-anchors/parse.ts"
ESITO=0

for F in "$SKILL" "$AWK_SRC" "$TS_SRC"; do
  [ -f "$F" ] || { echo "❌ manca $F"; exit 1; }
done

# ── le righe d'esempio, dalla fonte autoritativa ────────────────────────────
ESEMPI="$(grep -oE '^[[:space:]]*- <a id="r[0-9]+"></a>\*\*R[0-9]+\*\*.*$' "$SKILL" \
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

# ── le due regex, estratte dai parser ───────────────────────────────────────
# awk: la riga che apre il blocco del requisito in feature-docs-check.sh
AWK_RE="$(grep -oE '/\^- <a id="r\[0-9\]\+"><\\/a>/' "$AWK_SRC" | head -1)"
[ -n "$AWK_RE" ] || { echo "❌ non trovo la regex del requisito in $AWK_SRC"; exit 1; }
AWK_ERE='^- <a id="r[0-9]+"></a>'

# JavaScript: la costante RIGA_REQUISITO in parse.ts, tradotta in ERE
TS_RE="$(grep -oE 'export const RIGA_REQUISITO = /.*/$' "$TS_SRC" | sed 's|^.* = /||; s|/$||')"
[ -n "$TS_RE" ] || { echo "❌ non trovo RIGA_REQUISITO in $TS_SRC"; exit 1; }
TS_ERE="$(printf '%s' "$TS_RE" \
  | sed 's|\\s|[[:space:]]|g; s|\\d|[0-9]|g; s|\\/|/|g; s|\\\*|[*]|g')"

prova() { # <ere> <riga> → 0 se riconosciuta
  printf '%s\n' "$2" | grep -qE "$1"
}

# ── le righe prescritte devono essere riconosciute da entrambi ──────────────
while IFS= read -r RIGA; do
  [ -n "$RIGA" ] || continue
  prova "$AWK_ERE" "$RIGA" || {
    echo "❌ $AWK_SRC non riconosce una riga che $SKILL prescrive:"
    echo "   $RIGA"
    echo "   → allinea la regex in $AWK_SRC"
    ESITO=1
  }
  prova "$TS_ERE" "$RIGA" || {
    echo "❌ $TS_SRC non riconosce una riga che $SKILL prescrive:"
    echo "   $RIGA"
    echo "   → allinea RIGA_REQUISITO in $TS_SRC"
    ESITO=1
  }
done <<< "$ESEMPI"

# ── i contro-esempi non devono essere riconosciuti da nessuno dei due ───────
# Senza questi, una regex che accetta qualunque riga passerebbe il controllo.
CONTRO=(
  '- [ ] il vecchio formato a checkbox, senza identificativo'
  '- **R1** — identificativo senza ancora'
  '- <a id="x1"></a>**R1** — ancora che non comincia per r'
)
for RIGA in "${CONTRO[@]}"; do
  if prova "$AWK_ERE" "$RIGA"; then
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
