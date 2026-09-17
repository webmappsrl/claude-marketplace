> Ticket: oc:8577

# wm-plan propone i tag adatti a un ticket — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** dare a `wm-plan` una sotto-fase che propone al dev i tag Orchestrator da associare al ticket, con la parte meccanica in uno script di sola lettura.

**Architecture:** uno script bash (`tag-candidati.sh`) interroga `GET /api/tags?search=` una volta per termine, scarta i tag con `description` non nulla, quelli già presenti sul ticket e quelli che portano un trimestre nel nome, e stampa JSON con `candidati` e `scoperti`. La skill non vede mai le descrizioni. La sotto-fase `environment-setup: tag-suggestion` in `SKILL.md` mostra i candidati al dev e scrive su Orchestrator solo dopo un sì esplicito, uno per tag.

**Tech Stack:** bash, `curl`, `jq`. Nessuna dipendenza nuova: entrambi sono già usati dagli script esistenti del repo.

**Spec:** `docs/features/8577-wm-plan-propone-i-tag-adatti/overview.md`

## Global Constraints

- Repo unico: `claude-marketplace`. Nessun submodule.
- Documentazione, commenti e messaggi di commit in italiano; termini tecnici in inglese.
- **Nessun commit viene eseguito.** Gli step "Commit" sono istruzioni testuali per il dev.
- Lo script è di **sola lettura**: nessuna chiamata di scrittura verso Orchestrator.
- Nessuna prova contro `https://orchestrator.maphub.it` che scriva. Le letture sono ammesse.
- Le `description` dei tag non escono mai dallo script.
- `claude plugin validate .` deve passare prima di dichiarare il lavoro pronto.
- Stile degli script: `set -uo pipefail`, intestazione con uso ed exit code, output a campi separati, come `plugins/wm-skills/scripts/claude-md-lint.sh`.

---

### Task 1: Script — ricerca per termine e filtro sulle descrizioni

**Files:**
- Create: `plugins/wm-skills/scripts/tag-candidati.sh`
- Create: `plugins/wm-skills/scripts/tests/tag-candidati.test.sh`
- Create: `plugins/wm-skills/scripts/tests/fixtures/tags-marketplace.json`
- Create: `plugins/wm-skills/scripts/tests/fixtures/tags-misti.json`

**Interfaces:**
- Consumes: niente (primo task).
- Produces: `tag-candidati.sh <story_id> <termine>...`, che stampa su stdout un oggetto JSON
  `{"candidati":[{"id":N,"name":"...","termine":"..."}],"scoperti":["..."]}`.
  Punto di iniezione per i test: se la variabile d'ambiente `TAG_FIXTURE_DIR` è valorizzata,
  lo script legge `"$TAG_FIXTURE_DIR/tags-<termine>.json"` invece di chiamare l'API, e
  `"$TAG_FIXTURE_DIR/story-<id>.json"` per il ticket (usato dal Task 2).

- [ ] **Step 1: Scrivere le fixture**

`plugins/wm-skills/scripts/tests/fixtures/tags-marketplace.json` — risposta reale di
`GET /api/tags?search=marketplace`:

```json
[{"description":null,"id":648,"name":"claude-marketplace"}]
```

`plugins/wm-skills/scripts/tests/fixtures/tags-misti.json` — un'etichetta e un dossier, come
rende davvero l'API:

```json
[{"description":null,"id":662,"name":"[26Q3]FORESTAS"},
 {"description":"Contesto lungo del dossier, con credenziali in chiaro.","id":527,"name":"[26Q1][forestas]"}]
```

- [ ] **Step 2: Scrivere il test che fallisce**

`plugins/wm-skills/scripts/tests/tag-candidati.test.sh`:

```bash
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
    printf '✘ %s\n   atteso:  %s\n   ottenuto: %s\n' "$1" "$2" "$3"
    FALLITI=$((FALLITI+1))
  fi
}

# Un'etichetta pura diventa candidato.
OUT=$("$SCRIPT" 8577 marketplace | jq -c '.candidati')
verifica "etichetta proposta come candidato" \
  '[{"id":648,"name":"claude-marketplace","termine":"marketplace"}]' "$OUT"

# Un tag con description non nulla viene scartato, l'etichetta accanto no.
OUT=$("$SCRIPT" 8577 misti | jq -c '[.candidati[].id]')
verifica "tag con descrizione scartato" '[662]' "$OUT"

# La descrizione non compare mai nell'output.
OUT=$("$SCRIPT" 8577 misti | grep -c "credenziali" || true)
verifica "nessuna descrizione nell'output" "0" "$OUT"

exit $([ "$FALLITI" -eq 0 ] && echo 0 || echo 1)
```

- [ ] **Step 3: Eseguire il test e verificare che fallisca**

Run: `bash plugins/wm-skills/scripts/tests/tag-candidati.test.sh`
Expected: FAIL — `tag-candidati.sh` non esiste ancora.

- [ ] **Step 4: Scrivere lo script minimo**

`plugins/wm-skills/scripts/tag-candidati.sh`:

```bash
#!/usr/bin/env bash
# Propone i tag Orchestrator candidati per un ticket. SOLA LETTURA: nessuna scrittura.
#
# Uso:  tag-candidati.sh <story_id> <termine>...
#
# Output: JSON su stdout
#   {"candidati":[{"id":N,"name":"...","termine":"..."}],"scoperti":["..."]}
#
# Un tag è candidato se: lo trova la ricerca, ha description nulla (è un'etichetta, non un
# dossier), e non è già associato al ticket. Le description non escono mai da qui: servono
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

# ---------- raccolta ----------
CANDIDATI='[]'
SCOPERTI='[]'

for termine in "${TERMINI[@]}"; do
  grezzo=$(api_get "/api/tags?search=$(printf '%s' "$termine" | jq -sRr @uri)" "tags-$termine")
  # solo etichette: description nulla. La description non viene mai riportata.
  trovati=$(printf '%s' "$grezzo" \
    | jq --arg t "$termine" '[ .[] | select(.description == null)
                               | {id, name, termine: $t} ]')
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
```

- [ ] **Step 5: Rendere eseguibili e rilanciare il test**

Run:
```bash
chmod +x plugins/wm-skills/scripts/tag-candidati.sh plugins/wm-skills/scripts/tests/tag-candidati.test.sh
bash plugins/wm-skills/scripts/tests/tag-candidati.test.sh
```
Expected: PASS — tre righe `✔`.

- [ ] **Step 6: Commit (istruzione per il dev, non eseguire)**

```bash
git add plugins/wm-skills/scripts/tag-candidati.sh plugins/wm-skills/scripts/tests/
git commit -m "feat(oc:8577): script dei tag candidati, con filtro sulle descrizioni"
```

---

### Task 2: Script — scartare i tag già presenti sul ticket

> ⚠️ L'implementazione ha deviato da questo task: [notes.md](notes.md#il-trimestre-esce-dalle-caratteristiche-proposte)

**Files:**
- Modify: `plugins/wm-skills/scripts/tag-candidati.sh`
- Modify: `plugins/wm-skills/scripts/tests/tag-candidati.test.sh`
- Create: `plugins/wm-skills/scripts/tests/fixtures/story-8577.json`

**Interfaces:**
- Consumes: `tag-candidati.sh` del Task 1, con la funzione `api_get`.
- Produces: stesso contratto di output; i candidati già associati al ticket non compaiono più.

- [ ] **Step 1: Scrivere la fixture del ticket**

`plugins/wm-skills/scripts/tests/fixtures/story-8577.json`, com'è davvero oc:8577:

```json
{"id":8577,"name":"wm-plan propone i tag adatti a un ticket","tags":[{"id":661,"name":"26Q3"}]}
```

E `plugins/wm-skills/scripts/tests/fixtures/tags-26Q3.json`:

```json
[{"description":null,"id":661,"name":"26Q3"}]
```

- [ ] **Step 2: Aggiungere il test che fallisce**

In fondo a `tag-candidati.test.sh`, prima della riga `exit`:

```bash
# Un tag già presente sul ticket non viene riproposto: il termine risulta scoperto.
OUT=$("$SCRIPT" 8577 26Q3 | jq -c '.')
verifica "tag già sul ticket non riproposto" \
  '{"candidati":[],"scoperti":["26Q3"]}' "$OUT"
```

- [ ] **Step 3: Eseguire il test e verificare che fallisca**

Run: `bash plugins/wm-skills/scripts/tests/tag-candidati.test.sh`
Expected: FAIL — il tag 661 compare ancora fra i candidati.

- [ ] **Step 4: Implementare lo scarto**

In `tag-candidati.sh`, subito dopo il blocco `TERMINI=("$@")`, aggiungere la lettura dei tag
del ticket:

```bash
# Tag già sul ticket: si scartano dai candidati, altrimenti la caratteristica "trimestre"
# riproporrebbe quasi sempre il tag che Orchestrator assegna alla creazione.
GIA=$(api_get "/api/stories/$STORY_ID" "story-$STORY_ID" | jq '[.tags[]?.id]')
```

e dentro il ciclo, sostituire la riga che calcola `trovati` con:

```bash
  trovati=$(printf '%s' "$grezzo" \
    | jq --arg t "$termine" --argjson gia "$GIA" \
         '[ .[] | select(.description == null) | select(.id as $i | $gia | index($i) | not)
            | {id, name, termine: $t} ]')
```

- [ ] **Step 5: Rilanciare il test**

Run: `bash plugins/wm-skills/scripts/tests/tag-candidati.test.sh`
Expected: PASS — quattro righe `✔`.

- [ ] **Step 6: Prova sui ticket veri (richiesta esplicita del dev)**

Run:
```bash
plugins/wm-skills/scripts/tag-candidati.sh 8577 marketplace 26Q3
plugins/wm-skills/scripts/tag-candidati.sh 8545 marketplace 26Q3
```
Expected: su oc:8577 il candidato `claude-marketplace` (id 648) e `26Q3` fra gli scoperti,
perché il ticket ce l'ha già. Nessuna descrizione nell'output. Mostrare l'esito al dev e
attendere il suo giudizio prima di passare al Task 3: è la prova che i candidati proposti
sono quelli che avrebbe scelto lui.

- [ ] **Step 7: Commit (istruzione per il dev, non eseguire)**

```bash
git add plugins/wm-skills/scripts/
git commit -m "fix(oc:8577): i tag già presenti sul ticket non vengono riproposti"
```

---

### Task 3: Script — varianti del nome prima di dichiarare un termine scoperto

**Files:**
- Modify: `plugins/wm-skills/scripts/tag-candidati.sh`
- Modify: `plugins/wm-skills/scripts/tests/tag-candidati.test.sh`
- Create: `plugins/wm-skills/scripts/tests/fixtures/tags-claude marketplace.json`

**Interfaces:**
- Consumes: `tag-candidati.sh` dei Task 1 e 2.
- Produces: stesso contratto; un termine finisce in `scoperti` solo se nessuna variante trova nulla.

- [ ] **Step 1: Scrivere la fixture della variante**

`plugins/wm-skills/scripts/tests/fixtures/tags-claude marketplace.json` — la ricerca con lo
spazio non trova nulla, quella con il trattino sì:

```json
[]
```

- [ ] **Step 2: Aggiungere il test che fallisce**

```bash
# Cercando con lo spazio non si trova nulla, ma la variante col trattino sì:
# il termine NON è scoperto, è un candidato.
OUT=$("$SCRIPT" 8577 "claude marketplace" | jq -c '[.candidati[].id, .scoperti[]]')
verifica "variante del nome trovata prima di dichiarare scoperto" '[648]' "$OUT"
```

- [ ] **Step 3: Eseguire il test e verificare che fallisca**

Run: `bash plugins/wm-skills/scripts/tests/tag-candidati.test.sh`
Expected: FAIL — `"claude marketplace"` finisce fra gli scoperti.

- [ ] **Step 4: Implementare le varianti**

In `tag-candidati.sh`, prima del ciclo principale, aggiungere:

```bash
# Varianti plausibili di un nome, per non proporre la creazione di un doppione:
# trattini e spazi si scambiano, le parentesi quadre si tolgono. Il confronto dell'API
# ignora già le maiuscole.
varianti() { # termine -> una variante per riga, senza ripetizioni
  local t="$1"
  printf '%s\n' "$t" "${t//-/ }" "${t// /-}" "${t//[/}" | sed 's/]//g' | awk 'NF && !v[$0]++'
}
```

e sostituire il corpo del ciclo con:

```bash
for termine in "${TERMINI[@]}"; do
  trovati='[]'
  while IFS= read -r variante; do
    grezzo=$(api_get "/api/tags?search=$(printf '%s' "$variante" | jq -sRr @uri)" "tags-$variante")
    trovati=$(printf '%s' "$grezzo" \
      | jq --arg t "$termine" --argjson gia "$GIA" --argjson acc "$trovati" \
           '$acc + [ .[] | select(.description == null)
                     | select(.id as $i | $gia | index($i) | not)
                     | {id, name, termine: $t} ]')
    [ "$(printf '%s' "$trovati" | jq 'length')" -gt 0 ] && break
  done < <(varianti "$termine")

  if [ "$(printf '%s' "$trovati" | jq 'length')" -eq 0 ]; then
    SCOPERTI=$(printf '%s' "$SCOPERTI" | jq --arg t "$termine" '. + [$t]')
  else
    CANDIDATI=$(jq -n --argjson a "$CANDIDATI" --argjson b "$trovati" '$a + $b')
  fi
done
```

- [ ] **Step 5: Rilanciare il test**

Run: `bash plugins/wm-skills/scripts/tests/tag-candidati.test.sh`
Expected: PASS — cinque righe `✔`.

- [ ] **Step 6: Commit (istruzione per il dev, non eseguire)**

```bash
git add plugins/wm-skills/scripts/
git commit -m "feat(oc:8577): varianti del nome prima di dichiarare scoperto un termine"
```

---

### Task 4: La sotto-fase `environment-setup: tag-suggestion` in `SKILL.md`

> ⚠️ L'implementazione ha deviato da questo task: [notes.md](notes.md#una-fase-sola-diventa-due)

**Files:**
- Modify: `plugins/wm-skills/skills/wm-plan/SKILL.md` (nuova sotto-sezione dopo
  `### environment-setup: docker-check`)

**Interfaces:**
- Consumes: `plugins/wm-skills/scripts/tag-candidati.sh` dei Task 1-3, con il suo contratto
  di output `{"candidati":[{id,name,termine}],"scoperti":[termine]}`.
- Produces: la sotto-fase che il diagramma del Task 5 deve rispecchiare.

- [ ] **Step 1: Scrivere la sotto-sezione**

Aggiungere in `SKILL.md`, subito dopo la fine di `### environment-setup: docker-check`:

````markdown
### environment-setup: tag-suggestion

Propone al dev i tag Orchestrator da associare al ticket. **Non eseguita in tag-mode**: lì il
tag padre lo decide `wm-tag`.

Salta la fase se non c'è un ticket (`caso-b` senza creazione, o lavoro senza ID).

**I termini di ricerca** si ricavano da quello che la fase ha già in mano, senza chiedere nulla:

| Caratteristica | Da dove | Esempio |
|---|---|---|
| Repository | `basename` dell'URL di `git remote get-url origin`, senza `.git` | `claude-marketplace` |
| Area | `stack_type` e `stack_ui` di `environment-setup: project-detection` | `backend`, `webapp`, `app` |

**La ricerca e il filtro non si fanno nel prompt**, si fanno con lo script:

```bash
plugins/wm-skills/scripts/tag-candidati.sh <ID> <termine>...
```

Restituisce `candidati` (tag che esistono, sono etichette e non sono già sul ticket) e
`scoperti` (termini per cui non esiste nulla, varianti del nome comprese). **Le descrizioni dei
tag non entrano mai nel context:** lo script le usa solo per scartare, e alcune contengono
credenziali in chiaro.

**Se lo script fallisce** (exit diverso da 0, o nessun output), **fermati e chiedi al dev**:

> ⚠️ Non sono riuscito a leggere i tag da Orchestrator: \<motivo\>. Riprovo, o vado avanti senza?

Non proseguire in silenzio: un avviso che scorre via mentre il dev legge altro lo lascia
convinto che i tag ci siano. Se il dev sceglie di proseguire, registralo in `notes.md`
(sezione `## Decisioni`), creando il file se non esiste.

**Per ogni candidato, uno per volta**, mostra nome e provenienza e chiedi:

> Il ticket oc:\<ID\> non ha il tag `claude-marketplace` (trovato cercando "marketplace",
> caratteristica: repository). Lo associo?

Solo dopo un **sì esplicito** chiama `attach_story_to_tag`, prima senza `confirm` per
l'anteprima e poi con `confirm: true`. Nessuna approvazione in blocco, nessun silenzio-assenso.

**Per i termini scoperti**, valuta se proponere la creazione del tag. Il criterio non è la
forma del nome ma **la caratteristica che vale la pena tracciare**: la domanda è cosa qualcuno
cercherà fra sei mesi. Proponi solo se la caratteristica **ricorrerà** — un raggruppamento
destinato a contenere un solo ticket non è un raggruppamento. `wm-plan` crea **solo etichette**,
mai tag dossier con una descrizione: quelli sono di `wm-tag`.

La creazione passa da `create_tag`, con lo stesso doppio passaggio anteprima/conferma.

**Al termine, registra in `notes.md`** (sezione `## Decisioni`) ogni tag associato o creato: se
un giorno si torna indietro, serve la lista di cosa è stato fatto, perché due tag su
Orchestrator non si possono fondere.

**La fase non blocca mai il workflow:** un tag mancante non ferma un lavoro.
````

- [ ] **Step 2: Verificare che il nome della sotto-fase non rompa il controllo del diagramma**

Run: `./.github/scripts/verifica-diagramma.sh`
Expected: PASS — lo script confronta solo le intestazioni `## Fase:`, e questa è una
sotto-sezione `###`. Se fallisce, il nome è stato scritto come fase di primo livello: correggerlo.

- [ ] **Step 3: Validare il plugin**

Run: `claude plugin validate .`
Expected: nessun errore.

- [ ] **Step 4: Commit (istruzione per il dev, non eseguire)**

```bash
git add plugins/wm-skills/skills/wm-plan/SKILL.md
git commit -m "feat(oc:8577): sotto-fase tag-suggestion in wm-plan"
```

---

### Task 5: Diagramma, confine con `wm-tag`, pagina di conoscenza

**Files:**
- Modify: `docs/guide/wm-plan-diagramma/index.html`
- Modify: `CLAUDE.md` (tabella *Coupling tra skill*, indice `## Conoscenza`)
- Create: `docs/knowledge/wm-plan-tag-suggestion.md`

**Interfaces:**
- Consumes: la sotto-fase del Task 4.
- Produces: niente che altri task usino — è l'ultimo.

- [ ] **Step 1: Aggiornare il diagramma**

Nel blocco `<pre class="mermaid">` di `docs/guide/wm-plan-diagramma/index.html`, il nodo di
`Fase: environment-setup` deve dire anche della proposta dei tag. **Il template è congelato
nella struttura**: si aggiorna il contenuto del nodo, non il layout né gli stili. I vincoli
stanno in `.claude/rules/wm-plan-diagramma.md`, che si carica da sé quando si tocca il file —
leggerlo prima di modificare.

- [ ] **Step 2: Verificare la coerenza**

Run: `./.github/scripts/verifica-diagramma.sh`
Expected: `✔ Diagramma allineato`.

- [ ] **Step 3: Scrivere la pagina di conoscenza**

`docs/knowledge/wm-plan-tag-suggestion.md`:

```markdown
# Proposta dei tag in wm-plan

## Come funziona oggi

Dopo `environment-setup`, `wm-plan` propone i tag da associare al ticket. I termini di ricerca
sono trimestre, repository e area, tutti ricavati da dati già in mano alla fase. La ricerca e il
filtro stanno in `plugins/wm-skills/scripts/tag-candidati.sh`, di sola lettura; la skill mostra i
candidati e scrive solo dopo un sì esplicito del dev, un tag per volta.

Un tag con una `description` non nulla non è mai un candidato: è un dossier di `wm-tag`, non
un'etichetta.

## Perché così

- **Il filtro sta in uno script, non nel prompt** (oc:8577): scartare per `description` e per id
  è confronto di stringhe, e uno script lo fa uguale tutte le volte. È lo stesso criterio per cui
  il lint del `CLAUDE.md` è uscito dal prompt del doctor.
- **Le descrizioni non entrano nel context** (oc:8577): `GET /api/tags` le manda sempre e alcune
  contengono credenziali in chiaro. Misurato: `search=RDO` rende 84.714 caratteri per 7 tag, di
  cui 83.391 di sole descrizioni.
- **Nessun tag senza un sì esplicito** (oc:8577): su Orchestrator due tag non si possono fondere,
  quindi un doppione si ripara solo a mano, ticket per ticket.
- **Il nome di un tag nuovo esprime una caratteristica che vale la pena tracciare** (oc:8577),
  non uno schema sintattico: conta cosa qualcuno cercherà fra sei mesi.
- **Se Orchestrator non risponde, la fase si ferma e chiede** (oc:8577): un avviso che scorre via
  mentre il dev legge altro lo lascia convinto che i tag ci siano.

## Come ci siamo arrivati

- **Leggere tutti i tag e confrontarne le descrizioni** (oc:8577, superata): è la prima forma in
  cui la feature era stata pensata. Abbandonata alla misura: l'elenco costava circa 21.000 token
  a ticket e portava in context credenziali in chiaro.
```

- [ ] **Step 4: Aggiornare il `CLAUDE.md`**

Nella tabella *Coupling tra skill*, aggiungere la riga:

```markdown
| `wm-plan` (tag-suggestion) | `wm-tag` | `wm-tag` crea tag dossier, con descrizione; `wm-plan` propone e crea **solo etichette**, cioè tag con `description` nulla. Il confine è il filtro sulla descrizione: se cambia in una delle due, verifica l'altra. |
```

Nella tabella `## Conoscenza`, aggiungere:

```markdown
| Proposta dei tag in `wm-plan` | Termini di ricerca, filtro sulle descrizioni, conferma per singolo tag | [docs/knowledge/wm-plan-tag-suggestion.md](docs/knowledge/wm-plan-tag-suggestion.md) |
```

- [ ] **Step 5: Controllo di forma prima di scrivere**

Invocare `wm-context-guard` passandogli il percorso `docs/knowledge/wm-plan-tag-suggestion.md`
e il testo proposto. Verificare ogni rilievo che cita una voce esistente. Una contraddizione va
portata al dev, non risolta da soli. Se l'agente non risponde, scrivere comunque: è un ausilio,
non un gate.

- [ ] **Step 6: Validare**

Run:
```bash
claude plugin validate .
./.github/scripts/verifica-diagramma.sh
bash plugins/wm-skills/scripts/tests/tag-candidati.test.sh
```
Expected: tutti e tre senza errori.

- [ ] **Step 7: Commit (istruzione per il dev, non eseguire)**

```bash
git add docs/ CLAUDE.md
git commit -m "docs(oc:8577): diagramma, confine con wm-tag e pagina di conoscenza"
```
