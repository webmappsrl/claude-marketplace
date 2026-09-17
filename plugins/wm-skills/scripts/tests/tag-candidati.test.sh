#!/usr/bin/env bash
# Test dello script tag-candidati.sh. Non tocca la rete: le risposte dell'API
# arrivano dalle fixture in tests/fixtures/ tramite TAG_FIXTURE_DIR.
set -uo pipefail
QUI="$(cd "$(dirname "$0")" && pwd)"
SCRIPT="$QUI/../tag-candidati.sh"
export TAG_FIXTURE_DIR="$QUI/fixtures"
FALLITI=0

verifica() { # nome, atteso, ottenuto
  if [ "$2" = "$3" ]; then
    printf '✔ %s\n' "$1"
  else
    printf '✘ %s\n   atteso:   %s\n   ottenuto: %s\n' "$1" "$2" "$3"
    FALLITI=$((FALLITI+1))
  fi
}

# Un'etichetta pura diventa candidato.
OUT=$("$SCRIPT" 8577 marketplace | jq -c '.candidati')
verifica "etichetta proposta come candidato" \
  '[{"id":648,"name":"claude-marketplace","termine":"marketplace"}]' "$OUT"

# Un tag con description non nulla viene scartato, l'etichetta accanto no.
OUT=$("$SCRIPT" 8577 misti | jq -c '[.candidati[].id]')
verifica "tag con descrizione scartato" '[648]' "$OUT"

# La descrizione non compare mai nell'output.
OUT=$("$SCRIPT" 8577 misti | grep -c "credenziali" || true)
verifica "nessuna descrizione nell'output" "0" "$OUT"

# Un tag già presente sul ticket non viene riproposto: il termine risulta scoperto.
OUT=$("$SCRIPT" 8577 26Q3 | jq -c '.')
verifica "tag già sul ticket non riproposto" \
  '{"candidati":[],"scoperti":["26Q3"]}' "$OUT"

# Ogni tag che porta il trimestre nel nome è fuori: il trimestre lo assegna Orchestrator.
# Qui [26Q3]FORESTAS è un tag cliente che contiene il trimestre, e non va mai proposto.
OUT=$("$SCRIPT" 8577 trimestrati | jq -c '.')
verifica "tag col trimestre nel nome ignorati" \
  '{"candidati":[],"scoperti":["trimestrati"]}' "$OUT"

# Cercando con lo spazio non si trova nulla, ma la variante col trattino sì:
# il termine NON è scoperto, è un candidato.
OUT=$("$SCRIPT" 8577 "claude marketplace" | jq -c '[.candidati[].id, .scoperti[]]')
verifica "variante del nome trovata prima di dichiarare scoperto" '[648]' "$OUT"

exit $([ "$FALLITI" -eq 0 ] && echo 0 || echo 1)
