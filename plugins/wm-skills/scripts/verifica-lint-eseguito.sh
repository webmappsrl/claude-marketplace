#!/usr/bin/env bash
# Impedisce a wm-context-doctor di chiudere senza aver eseguito claude-md-lint.sh.
#
# Il lint scrive un marker con il proprio timestamp a ogni esecuzione. Qui si controlla
# che ne esista almeno uno recente: se non c'è, exit 2 blocca la chiusura del subagente
# e il messaggio su stderr gli dice cosa fare.
#
# È la parte che l'ambiente garantisce e il prompt no: un'istruzione si può saltare,
# un hook no.

set -uo pipefail
FRESCO=$((15 * 60))   # secondi
ORA=$(date +%s)

for m in "${TMPDIR:-/tmp}"/wm-context-doctor-*.run; do
  [ -f "$m" ] || continue
  T=$(cat "$m" 2>/dev/null || echo 0)
  [ $((ORA - T)) -lt $FRESCO ] && exit 0
done

cat >&2 <<'MSG'
Non risulta eseguito il lint meccanico, che è il primo e l'ultimo passo del tuo lavoro.

Eseguilo ora e confronta il suo output con la relazione che stai per chiudere:

  ${CLAUDE_PLUGIN_ROOT}/scripts/claude-md-lint.sh <percorso/CLAUDE.md> [repo vicino]...

Per ogni riga che restituisce, la tua relazione deve averla affrontata: un rilievo `error` o
`warn` non si contraddice senza una prova; un `hint` è un'euristica e si può contestare, ma
motivando. Le divergenze si dichiarano, non si omettono.
MSG
exit 2
