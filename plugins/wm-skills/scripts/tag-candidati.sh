#!/usr/bin/env bash
# Propone i tag Orchestrator candidati per un ticket. SOLA LETTURA: nessuna scrittura.
#
# Uso:  tag-candidati.sh <story_id> <termine>...
#
# Output: JSON su stdout
#   {"candidati":[{"id":N,"name":"...","termine":"..."}],"scoperti":["..."]}
#
# Un tag è candidato se: lo trova la ricerca, ha description nulla (è un'etichetta, non un
# dossier), non è già associato al ticket, e non porta un trimestre nel nome — il trimestre
# lo assegna Orchestrator alla creazione, e un nome come "[26Q3]FORESTAS" è il tag di un
# cliente, non del trimestre: proporlo sposterebbe il lavoro nelle cose di quel cliente. Le description non escono mai da qui: servono
# solo a scartare, e alcune contengono credenziali in chiaro.
#
# Exit: 0 ok, 2 uso sbagliato, 3 credenziali assenti, 4 API non raggiungibile
#
# Test: TAG_FIXTURE_DIR fa leggere le risposte da file invece che dall'API.

set -uo pipefail

[ $# -ge 2 ] || { echo "uso: $0 <story_id> <termine>..." >&2; exit 2; }
STORY_ID="$1"; shift
TERMINI=("$@")

AUTH="${ORCHESTRATOR_AUTH_FILE:-$HOME/.config/webmapp/orchestrator-auth.json}"
BASE="${ORCHESTRATOR_BASE_URL:-https://orchestrator.maphub.it}"
TIMEOUT="${TAG_TIMEOUT:-15}"

api_get() { # percorso, nome-fixture -> JSON su stdout
  local percorso="$1" fixture="${2:-}"
  if [ -n "${TAG_FIXTURE_DIR:-}" ]; then
    # il nome della fixture è il termine leggibile, non quello codificato nell'URL
    cat "$TAG_FIXTURE_DIR/$fixture.json" 2>/dev/null || echo '[]'
    return 0
  fi
  [ -f "$AUTH" ] || { echo "credenziali assenti: $AUTH" >&2; exit 3; }
  local token
  token=$(jq -r '.token' "$AUTH")
  curl -sS --max-time "$TIMEOUT" -H "Authorization: Bearer $token" \
       -H "Accept: application/json" "$BASE$percorso" || exit 4
}

# Tag già sul ticket: si scartano dai candidati, altrimenti la caratteristica "trimestre"
# riproporrebbe quasi sempre il tag che Orchestrator assegna alla creazione.
GIA=$(api_get "/api/stories/$STORY_ID" "story-$STORY_ID" | jq '[.tags[]?.id]')

# Varianti plausibili di un nome, per non proporre la creazione di un doppione: trattini e
# spazi si scambiano, le parentesi quadre si tolgono. Il confronto dell'API ignora già le
# maiuscole. Si prova una variante per volta e ci si ferma alla prima che trova qualcosa.
varianti() { # termine -> una variante per riga, senza ripetizioni
  local t="$1"
  printf '%s\n' "$t" "${t//-/ }" "${t// /-}" "${t//[/}" | sed 's/]//g' | awk 'NF && !v[$0]++'
}

CANDIDATI='[]'
SCOPERTI='[]'

for termine in "${TERMINI[@]}"; do
  trovati='[]'
  while IFS= read -r variante; do
    grezzo=$(api_get "/api/tags?search=$(printf '%s' "$variante" | jq -sRr @uri)" "tags-$variante")
    # solo etichette: description nulla. La description non viene mai riportata.
    trovati=$(printf '%s' "$grezzo" \
      | jq --arg t "$termine" --argjson gia "$GIA" \
           '[ .[] | select(.description == null)
              | select(.id as $i | $gia | index($i) | not)
              | select(.name | test("[0-9]{2}[Qq][1-4]") | not)
              | {id, name, termine: $t} ]')
    [ "$(printf '%s' "$trovati" | jq 'length')" -gt 0 ] && break
  done < <(varianti "$termine")

  if [ "$(printf '%s' "$trovati" | jq 'length')" -eq 0 ]; then
    SCOPERTI=$(printf '%s' "$SCOPERTI" | jq --arg t "$termine" '. + [$t]')
  else
    CANDIDATI=$(jq -n --argjson a "$CANDIDATI" --argjson b "$trovati" '$a + $b')
  fi
done

# deduplica per id, mantenendo il primo termine che l'ha trovato
CANDIDATI=$(printf '%s' "$CANDIDATI" | jq 'unique_by(.id)')

jq -n --argjson c "$CANDIDATI" --argjson s "$SCOPERTI" \
  '{candidati: $c, scoperti: $s}'
