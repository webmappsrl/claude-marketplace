#!/usr/bin/env bash
# Nessun $1, $2, … nei file delle skill. Quando carica una skill, Claude Code sostituisce $0-$9 e
# ${0}-${9} con gli argomenti con cui la si lancia, anche dentro i blocchi di codice: una funzione
# bash con "$1" arriva al modello con l'indirizzo passato alla skill al posto del parametro, e
# nessuno se ne accorge finché il comando non fallisce. Gli script con parametri vanno in
# plugins/<plugin>/scripts/ e la skill li chiama.
set -uo pipefail
cd "$(dirname "$0")/../.."

TROVATI=$(grep -rnE '\$[0-9]|\$\{[0-9]' plugins/*/skills --include='*.md' || true)
if [ -n "$TROVATI" ]; then
  echo "✘ Parametri posizionali nei file delle skill: Claude Code li sostituisce con gli argomenti di avvio."
  echo "  Sposta il codice in uno script sotto plugins/<plugin>/scripts/ e chiamalo dalla skill."
  echo "$TROVATI" | sed 's/^/  /'
  exit 1
fi
echo "✔ Nessun parametro posizionale nei file delle skill."
