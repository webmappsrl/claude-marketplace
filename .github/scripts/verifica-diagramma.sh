#!/usr/bin/env bash
# Verifica che le fasi dichiarate in wm-plan/SKILL.md e i nodi del diagramma siano
# gli stessi. Il diagramma è scritto a mano leggendo la skill: senza questo controllo,
# aggiungere o rinominare una fase e dimenticare la pagina non produce alcun errore —
# il diagramma mostra un workflow che non esiste più e nessuno se ne accorge.
#
# LIMITE, da conoscere: il confronto è sui NOMI. Se una fase cambia comportamento
# mantenendo il titolo, il controllo passa e il diagramma resta vecchio. Stabilire se un
# paragrafo descriva ancora il comportamento reale è giudizio, non confronto di stringhe.

set -uo pipefail

SKILL="plugins/wm-skills/skills/wm-plan/SKILL.md"
PAGINA="docs/guide/wm-plan-diagramma/index.html"

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

ESITO=0

if [ -n "$MANCANTI" ]; then
  echo "❌ Fasi presenti nella skill e assenti dal diagramma:"
  echo "$MANCANTI" | sed 's/^/   - Fase: /'
  echo "   → aggiungi il nodo in $PAGINA"
  ESITO=1
fi

if [ -n "$INVENTATE" ]; then
  echo "❌ Fasi presenti nel diagramma e non più nella skill:"
  echo "$INVENTATE" | sed 's/^/   - Fase: /'
  echo "   → rimuovi o rinomina il nodo in $PAGINA"
  ESITO=1
fi

if [ "$ESITO" -eq 0 ]; then
  echo "✔ Diagramma allineato: $(wc -l < /tmp/fasi-skill | tr -d ' ') fasi, stesse nella skill e nel diagramma."
fi

exit $ESITO
