#!/usr/bin/env bash
# Verifica che le fasi dichiarate nello SKILL.md di ogni skill con un diagramma pubblicato
# e i nodi del suo diagramma siano gli stessi. Il diagramma è scritto a mano leggendo la skill: senza questo controllo,
# aggiungere o rinominare una fase e dimenticare la pagina non produce alcun errore —
# il diagramma mostra un workflow che non esiste più e nessuno se ne accorge.
#
# LIMITE, da conoscere: il confronto è sui NOMI. Se una fase cambia comportamento
# mantenendo il titolo, il controllo passa e il diagramma resta vecchio. Stabilire se un
# paragrafo descriva ancora il comportamento reale è giudizio, non confronto di stringhe.

set -uo pipefail

# Coppie skill:diagramma da controllare. Si possono passare come argomenti (servono ai test);
# senza argomenti si controllano tutte le skill che hanno un diagramma pubblicato.
COPPIE=("$@")
if [ "${#COPPIE[@]}" -eq 0 ]; then
  COPPIE=(
    "plugins/wm-skills/skills/wm-plan/SKILL.md:docs/guide/wm-plan-diagramma/index.html"
    "plugins/wm-skills/skills/wm-geohub-import-check/SKILL.md:docs/guide/wm-geohub-import-check-diagramma/index.html"
  )
fi

ESITO=0
for COPPIA in "${COPPIE[@]}"; do
  SKILL="${COPPIA%%:*}"
  PAGINA="${COPPIA#*:}"

  for f in "$SKILL" "$PAGINA"; do
    [ -f "$f" ] || { echo "❌ File assente: $f"; exit 1; }
  done

  # Le fasi della skill: le intestazioni "## Fase: <nome>"
  grep -o '^## Fase: [a-z-]*' "$SKILL" | sed 's/^## Fase: //' | sort -u > /tmp/fasi-skill

  # I nodi del diagramma: l'etichetta sta fra virgolette, in una qualsiasi delle forme
  # mermaid — ["..."], {{"..."}}, (["..."]), [["..."]] — quindi si cerca il testo citato,
  # non la parentesi che lo racchiude.
  sed -n '/<pre class="mermaid">/,/<\/pre>/p' "$PAGINA" \
    | grep -o '"Fase: [a-z-]*' | sed 's/"Fase: //' | sort -u > /tmp/fasi-diagramma

  MANCANTI=$(comm -23 /tmp/fasi-skill /tmp/fasi-diagramma)
  INVENTATE=$(comm -13 /tmp/fasi-skill /tmp/fasi-diagramma)

  if [ -n "$MANCANTI" ]; then
    echo "❌ $SKILL: fasi presenti nella skill e assenti dal diagramma:"
    echo "$MANCANTI" | sed 's/^/   - Fase: /'
    echo "   → aggiungi il nodo in $PAGINA"
    ESITO=1
  fi

  if [ -n "$INVENTATE" ]; then
    echo "❌ $SKILL: fasi presenti nel diagramma e non più nella skill:"
    echo "$INVENTATE" | sed 's/^/   - Fase: /'
    echo "   → rimuovi o rinomina il nodo in $PAGINA"
    ESITO=1
  fi

  if [ -z "$MANCANTI" ] && [ -z "$INVENTATE" ]; then
    echo "✔ $(basename "$(dirname "$SKILL")"): $(wc -l < /tmp/fasi-skill | tr -d ' ') fasi, stesse nella skill e nel diagramma."
  fi
done

exit $ESITO
