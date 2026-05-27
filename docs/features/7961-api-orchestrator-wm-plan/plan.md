> Ticket: oc:7961

# Integrazione API Orchestrator in wm-plan — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Integrare le API REST di Orchestrator direttamente in `wm-plan` così che la skill legga i ticket via API, li crei quando mancanti, li aggiorni a fine workflow e gestisca qualsiasi modifica esplicita richiesta dall'utente — tutto con conferma obbligatoria per le scritture.

**Architecture:** Modifiche al solo file `plugins/wm-skills/skills/wm-plan/SKILL.md`. Si aggiunge una sezione `## Orchestrator API` con le istruzioni per autenticazione e chiamate HTTP, poi si modificano Fase 0 e la Checklist di completamento per usare le API. Le operazioni di lettura sono automatiche; ogni scrittura richiede conferma esplicita con preview della modifica.

**Tech Stack:** Markdown (SKILL.md), REST API Orchestrator (Laravel Sanctum), `curl` via Bash, token salvato in `~/.config/webmapp/orchestrator-token`, `ORCHESTRATOR_URL` env var.

---

## File Structure

| File | Azione | Responsabilità |
|------|--------|----------------|
| `plugins/wm-skills/skills/wm-plan/SKILL.md` | Modifica | Aggiunge sezione API helper + modifica Fase 0 + modifica Checklist |

---

### Task 1: Sezione `## Orchestrator API` — helper autenticazione e chiamate HTTP

Aggiunge al SKILL.md una sezione riutilizzabile che descrive come autenticarsi e fare chiamate. Tutte le altre fasi la referenziano.

**Files:**
- Modify: `plugins/wm-skills/skills/wm-plan/SKILL.md`

- [ ] **Step 1: Leggi il SKILL.md attuale per capire dove inserire la sezione**

Il file è già stato letto. La sezione va inserita **subito dopo il frontmatter e l'`<HARD-GATE>`**, prima di `## Fase 0`, così è visibile a tutto il workflow.

- [ ] **Step 2: Inserisci la sezione `## Orchestrator API` nel SKILL.md**

Inserisci il seguente blocco tra la riga `---` che chiude l'`<HARD-GATE>` e la riga `## Fase 0`:

```markdown
## Orchestrator API

Queste istruzioni valgono per tutte le chiamate HTTP a Orchestrator. Usale ogni volta che una fase richiede di leggere o scrivere un ticket.

### Configurazione

- **URL base:** leggi `$ORCHESTRATOR_URL` dall'environment. Se non è impostata usa `https://orchestrator.maphub.it` come default.
- **Token:** salvato in `~/.config/webmapp/orchestrator-token` (plain text, una riga). Se il file non esiste o la chiamata restituisce 401, esegui il login (vedi sotto).

### Login (solo se token assente o scaduto)

```bash
curl -s -X POST "$ORCHESTRATOR_URL/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"<email>","password":"<password>"}' \
  | jq -r '.token'
```

Chiedi email e password all'utente, esegui la chiamata, salva il token restituito:

```bash
mkdir -p ~/.config/webmapp
echo "<token>" > ~/.config/webmapp/orchestrator-token
```

### Lettura ticket (GET — nessuna conferma richiesta)

```bash
TOKEN=$(cat ~/.config/webmapp/orchestrator-token)
curl -s -X GET "$ORCHESTRATOR_URL/api/stories/<ID>" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Accept: application/json"
```

Se risponde 401: cancella il token e ripeti il login prima di ritentare.

### Creazione ticket (POST — richiede conferma esplicita)

Prima di eseguire, mostra all'utente un riepilogo con i campi che verranno inviati e chiedi conferma esplicita. Solo dopo:

```bash
TOKEN=$(cat ~/.config/webmapp/orchestrator-token)
curl -s -X POST "$ORCHESTRATOR_URL/api/stories" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '<json payload>'
```

Salva l'ID restituito (`$.id`) come `<ID>` del ticket per il resto del workflow.

### Aggiornamento ticket (PATCH — richiede conferma esplicita)

Prima di eseguire, mostra all'utente un riepilogo con i campi che verranno modificati e i loro valori, e chiedi conferma esplicita con questo formato:

> "Sto per aggiornare il ticket oc:<ID> con questi campi:
> - `<campo>`: `<valore>`
> Procedo?"

Solo dopo la conferma:

```bash
TOKEN=$(cat ~/.config/webmapp/orchestrator-token)
curl -s -X PATCH "$ORCHESTRATOR_URL/api/stories/<ID>" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '<json payload>'
```

### Status disponibili (letti dinamicamente)

Prima di proporre uno status, leggi i valori aggiornati da:
`https://raw.githubusercontent.com/webmappsrl/orchestrator/main/app/Enums/StoryStatus.php`

Estrai i valori dell'enum e presentali all'utente come lista numerata. Suggerisci quello più appropriato al contesto ma aspetta sempre la scelta esplicita dell'utente.

### Campi accettati (letti dinamicamente)

Prima di costruire un payload POST o PATCH, leggi i campi validati da:
`https://raw.githubusercontent.com/webmappsrl/orchestrator/main/app/Http/Requests/Api/StoryApiRequest.php`

Usa solo i campi presenti nelle `rules()` del Form Request. Non inviare campi non dichiarati.

### Regola generale scritture

**Qualsiasi operazione di scrittura su Orchestrator (POST o PATCH) richiede conferma esplicita con preview della modifica prima di eseguire la chiamata HTTP.** Nessuna eccezione.
```

- [ ] **Step 3: Verifica locale**

```bash
# Dalla root del repo claude-marketplace
claude plugin validate .
```

Expected: nessun errore di validazione.

- [ ] **Step 4: Commit**

```
feat(oc:7961): add Orchestrator API helper section to wm-plan
```

---

### Task 2: Modifica Fase 0 — lettura automatica ticket via API

Sostituisce il comportamento attuale (incolla manualmente) con lettura automatica quando l'utente fornisce solo l'ID. Mantiene il flusso manuale come fallback.

**Files:**
- Modify: `plugins/wm-skills/skills/wm-plan/SKILL.md`

- [ ] **Step 1: Sostituisci il corpo di `## Fase 0` con la versione API-aware**

Sostituisci l'intera sezione `## Fase 0 — Ticket Orchestrator` (dalla riga `## Fase 0` fino alla riga `---` successiva) con:

```markdown
## Fase 0 — Ticket Orchestrator

Il team Webmapp traccia il lavoro su Orchestrator. Ogni ticket ha un ID numerico referenziato come `oc:<ID>` (es. `oc:7815`).

### Caso A — L'utente fornisce un ID ticket

Se l'utente scrive `oc:<ID>` (con o senza contenuto aggiuntivo), leggi il ticket via API seguendo le istruzioni in `## Orchestrator API`:

```bash
TOKEN=$(cat ~/.config/webmapp/orchestrator-token 2>/dev/null)
curl -s -X GET "$ORCHESTRATOR_URL/api/stories/<ID>" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Accept: application/json"
```

Se il token non esiste o la risposta è 401, esegui il login prima di ritentare (vedi `## Orchestrator API → Login`).

Dal JSON restituito estrai:
- `name` → titolo, usato per il `<feature-slug>`: `<ID>-<titolo-in-kebab-case>`
- `customer_request` → contesto del problema, usato in Fase 2 e Fase 4
- `description` → note tecniche già raccolte, può orientare le domande in Fase 2
- `type` → orienta il tono dell'overview

Mostra all'utente un riepilogo del ticket letto prima di procedere alla Fase 1.

### Caso B — L'utente non ha un ticket

Procedi con il workflow usando le informazioni disponibili. Al termine della Fase 3 (Challenge), prima di scrivere i documenti, proponi il testo completo di un ticket con questo formato:

```
Titolo: <titolo sintetico della feature>

Tipo: Feature

customer_request:
<descrizione del problema e della soluzione emersa dalle fasi 0-3, 5-10 righe, linguaggio non tecnico>

description:
<approccio tecnico scelto, moduli coinvolti, vincoli emersi dalla Challenge>
```

Chiedi all'utente di confermarlo o modificarlo. Una volta approvato, crea il ticket via API seguendo `## Orchestrator API → Creazione ticket`. Salva l'ID restituito e usalo come `<ID>` per tutto il resto del workflow.

Se l'utente non vuole creare il ticket ora, procedi senza ID: usa solo il titolo kebab-case come slug (`feature/<titolo-in-kebab-case>`).

### Da estrarre in entrambi i casi

- `<feature-slug>`: `<ID>-<titolo-in-kebab-case>` (con ticket) o `<titolo-in-kebab-case>` (senza)
- **Riferimento ticket in tutti i documenti:** ogni file creato nelle fasi successive deve riportare in testa `> Ticket: oc:<ID>` (ometti se non c'è ID).
- **Commit convention:** tutti i commit usano `oc:<ID>` come scope (es. `feat(oc:7815): add OSM POI import action`). Senza ticket, usa il titolo kebab-case come scope.

### Aggiornamenti espliciti durante il workflow

Se in qualsiasi momento l'utente chiede di aggiornare un campo del ticket (es. "aggiorna lo status a progress", "scrivi nelle note dev che…"), esegui un PATCH seguendo `## Orchestrator API → Aggiornamento ticket`. Mostra sempre il preview della modifica e attendi conferma prima di inviare.
```

- [ ] **Step 2: Verifica locale**

```bash
claude plugin validate .
```

Expected: nessun errore.

- [ ] **Step 3: Test manuale in sessione locale**

```bash
/plugin marketplace add .
/plugin install wm-skills@wm-marketplace
```

Apri una nuova sessione e scrivi: `implementa oc:7961`. Verifica che la skill:
1. Tenti di leggere il ticket via curl (o chieda le credenziali se non c'è token)
2. Mostri il riepilogo del ticket letto prima di procedere

- [ ] **Step 4: Commit**

```
feat(oc:7961): update Phase 0 to fetch ticket via API automatically
```

---

### Task 3: Modifica Checklist di completamento — aggiornamento ticket a fine workflow

Aggiunge il passo di aggiornamento Orchestrator alla checklist finale, con preview obbligatorio per `answer_to_ticket` e `customer_request`.

**Files:**
- Modify: `plugins/wm-skills/skills/wm-plan/SKILL.md`

- [ ] **Step 1: Aggiungi i passi API alla checklist di completamento**

Nella sezione `## Checklist di completamento`, dopo le checkbox esistenti, aggiungi:

```markdown
### Aggiornamento Orchestrator (solo se esiste un ticket oc:<ID>)

- [ ] Leggi lo status attuale del ticket via GET `/api/stories/<ID>`
- [ ] Leggi gli status disponibili da `StoryStatus.php` su GitHub (vedi `## Orchestrator API → Status disponibili`)
- [ ] Suggerisci lo status più appropriato al contesto (es. `testing` se ci sono test da verificare, `done` se tutto è completato e i test passano) e presenta la lista completa — aspetta la scelta esplicita dell'utente
- [ ] Prepara la bozza di `answer_to_ticket` (note dev) con:
  - Link alla cartella `docs/features/<feature-slug>/`
  - Riepilogo tecnico di cosa è stato implementato (file creati/modificati, approccio usato)
  - Tono tecnico, rivolto al team
- [ ] Prepara la bozza di `customer_request` (risposta cliente) con:
  - Descrizione in linguaggio non tecnico di cosa è stato fatto e perché
  - Niente nomi di file, classi, branch o dettagli implementativi
  - Tono chiaro e orientato al beneficio per l'utente finale
- [ ] Mostra entrambe le bozze all'utente e chiedi approvazione esplicita per ognuna — la `customer_request` è letta dal cliente, richiede revisione attenta
- [ ] Solo dopo approvazione esplicita, esegui il PATCH seguendo `## Orchestrator API → Aggiornamento ticket` con i tre campi: `status`, `answer_to_ticket`, `customer_request`
```

- [ ] **Step 2: Verifica locale**

```bash
claude plugin validate .
```

Expected: nessun errore.

- [ ] **Step 3: Test manuale**

Simula la fine di un workflow in sessione locale: verifica che la skill presenti le bozze e aspetti la conferma prima di eseguire il PATCH.

- [ ] **Step 4: Commit**

```
feat(oc:7961): add Orchestrator ticket update to completion checklist
```

---

## Self-Review

**Spec coverage:**
- ✅ Lettura automatica ticket dato ID (Fase 0, Caso A)
- ✅ Login automatico se token assente/scaduto (sezione API helper)
- ✅ `ORCHESTRATOR_URL` configurabile via env (sezione API helper)
- ✅ Status list letta dinamicamente da GitHub (sezione API helper)
- ✅ Campi accettati letti dinamicamente da `StoryApiRequest.php` (sezione API helper)
- ✅ Creazione ticket via POST se non esiste (Fase 0, Caso B)
- ✅ Aggiornamenti espliciti su richiesta durante il workflow (Fase 0, fondo)
- ✅ Aggiornamento `status` + `answer_to_ticket` + `customer_request` a fine workflow (Checklist)
- ✅ Conferma obbligatoria con preview per ogni scrittura (regola generale + checklist)
- ✅ `customer_request` in linguaggio non tecnico con nota esplicita (checklist)

**Placeholder scan:** nessun TBD, nessun "implement later", tutti i comandi curl sono completi.

**Type consistency:** i nomi di campo (`answer_to_ticket`, `customer_request`, `status`) sono coerenti tra sezione API helper, Fase 0 e Checklist.
