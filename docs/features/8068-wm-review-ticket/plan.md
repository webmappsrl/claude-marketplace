> Ticket: oc:8068

# wm-review-ticket — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Creare la skill `wm-review-ticket` nel plugin `wm-skills` e aggiornare `wm-plan` per referenziarla, distribuendo la logica di code review a tutto il team via marketplace.

**Architecture:** Due file SKILL.md da modificare/creare in `plugins/wm-skills/skills/`. `wm-review-ticket` è una skill autonoma che legge il ticket Orchestrator, fa checkout della PR, legge gli artefatti `wm-plan` come contesto di intent, esegue finder paralleli sul diff e aggiorna il ticket con l'esito. `wm-plan` viene aggiornato per documentare il contratto artefatti (fonte autoritativa) e citare `wm-review-ticket` in Fase 6d.

**Tech Stack:** Markdown (SKILL.md), Bash (curl, git, jq), Orchestrator API REST, GitHub raw URL per fetch contratto.

---

### Task 1: Creare `plugins/wm-skills/skills/wm-review-ticket/SKILL.md`

**Files:**
- Create: `plugins/wm-skills/skills/wm-review-ticket/SKILL.md`

- [ ] **Step 1: Creare la cartella e il file SKILL.md**

```bash
mkdir -p plugins/wm-skills/skills/wm-review-ticket
```

- [ ] **Step 2: Scrivere il contenuto completo di SKILL.md**

Contenuto da scrivere in `plugins/wm-skills/skills/wm-review-ticket/SKILL.md`:

````markdown
---
name: wm-review-ticket
description: Esegui una code review completa di un ticket Orchestrator. Usa quando un collega ti assegna un ticket/PR da rivedere, oppure al termine di una feature wm-plan prima del merge. Input: oc:<ID>.
---

## Contratto artefatti

Questa skill consuma gli artefatti prodotti da `wm-skills:wm-plan`. Per conoscere la struttura autoritativa di `docs/features/<slug>/`, leggi la sezione "Contratto artefatti" di wm-plan:

```bash
# Recupera il contratto da GitHub (funziona anche senza wm-plan installato localmente)
# Cerca la sezione "## Contratto artefatti" nel file scaricato
```

URL: `https://raw.githubusercontent.com/webmappsrl/claude-marketplace/main/plugins/wm-skills/skills/wm-plan/SKILL.md`

---

## Orchestrator API

Stesse istruzioni di `wm-skills:wm-plan` — auth, login, migrazione token legacy, lettura e scrittura ticket. Riferisciti alla sezione `## Orchestrator API` di wm-plan per i dettagli completi.

Sintesi operativa:

```bash
ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-https://orchestrator.maphub.it}"
TOKEN=$(jq -r '.token' ~/.config/webmapp/orchestrator-auth.json 2>/dev/null)
# Se il file non esiste o risponde 401: esegui login come da wm-plan
```

---

## Fase 0 — Autenticazione

Verifica che `~/.config/webmapp/orchestrator-auth.json` esista e contenga un token valido.

- Se non esiste ma esiste `~/.config/webmapp/orchestrator-token`: esegui la migrazione (vedi `## Orchestrator API → Migrazione` in wm-plan).
- Se non esiste nessun file auth: esegui il login completo (chiedi email e password, salva JSON con `token`, `id`, `name`, `email`).
- Verifica con una chiamata `GET /api/me` — se risponde 401, cancella il file e ripeti il login.

---

## Fase 1 — Lettura ticket

```bash
ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-https://orchestrator.maphub.it}"
TOKEN=$(jq -r '.token' ~/.config/webmapp/orchestrator-auth.json)
curl -s -X GET "$ORCHESTRATOR_URL/api/stories/<ID>" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Accept: application/json"
```

Estrai dal JSON:
- `name` → titolo, usato per ricavare il `<feature-slug>`: `<ID>-<titolo-in-kebab-case>`
- `customer_request` → richiesta originale del cliente — è il criterio principale per valutare la correttezza della review
- `description` → cerca un URL PR GitHub con regex: `https://github\.com/[^/]+/[^/]+/pull/\d+`

Mostra all'utente un riepilogo del ticket prima di procedere.

**Se non trovi URL PR nella description:**

Chiedi all'utente:
> "Non ho trovato un URL PR nel ticket. Fornisci uno dei seguenti:
> 1. URL PR GitHub (es. `https://github.com/webmappsrl/maphub/pull/42`)
> 2. Nome branch (es. `feature/oc-8068-...`)
> 3. Hash commit"

---

## Fase 2 — Recupero contratto artefatti

Esegui WebFetch su:
```
https://raw.githubusercontent.com/webmappsrl/claude-marketplace/main/plugins/wm-skills/skills/wm-plan/SKILL.md
```

Cerca la sezione `## Contratto artefatti` e memorizza la struttura `docs/features/<slug>/` per la Fase 4.

---

## Fase 3 — Setup repo e checkout branch PR

### 3a — Identifica il repo

Leggi il `CLAUDE.md` del progetto corrente. Cerca sezioni che descrivono i repo coinvolti (es. path locali di submodule, repo noti del team).

Estrai dall'URL PR il nome del repo (es. `webmappsrl/maphub` → cerca path locale in CLAUDE.md).

Se il repo non è riconosciuto nel CLAUDE.md:
> "Non ho trovato il path locale del repo `<nome-repo>` nel CLAUDE.md. Dove si trova sulla tua macchina?"

### 3b — Stash se working tree dirty

```bash
cd <repo-path>
if [ -n "$(git status --porcelain)" ]; then
  echo "⚠️ Working tree non pulito — eseguo git stash prima del checkout. Lo stash verrà ripristinato al termine della review."
  git stash push -m "wm-review-ticket: stash automatico per review oc:<ID>"
  STASH_CREATED=true
fi
```

### 3c — Checkout branch PR

Estrai il nome del branch dall'URL PR tramite GitHub API o dalla description del ticket.

```bash
# Tentativo 1: branch locale
git checkout <branch-name> 2>/dev/null && echo "OK" || \

# Tentativo 2: fetch dal remote
git fetch origin <branch-name> && git checkout <branch-name> 2>/dev/null || \

# Tentativo 3: cerca commit di merge
git log --all --oneline --grep="<branch-name>" | head -5
```

Se tutti i tentativi falliscono:
> "Non riesco a trovare il branch `<branch-name>`. Puoi fornire l'hash del commit di merge o il nome corretto del branch?"

---

## Fase 4 — Lettura contesto wm-plan

Cerca i docs della feature nel repo:

```bash
# Ricerca fuzzy per ID ticket
find docs/features/ -name "*.md" 2>/dev/null | grep "<ID>" | head -5
# Oppure per slug esatto
ls docs/features/<ID>-* 2>/dev/null
```

**Se trovati** (`overview.md`, `plan.md`, `notes.md`):
- Leggi `overview.md` — sezioni "Cosa cambia", "Requisiti", "Out of scope": definiscono cosa il codice DEVE fare
- Leggi `plan.md` — lista task implementativi: verifica che tutti siano stati eseguiti
- Leggi `notes.md` — deviazioni dal piano: sono contestualmente approvate dall'autore, ma rimani libero di sollevare dubbi su deviazioni rischiose

Mostra all'utente quali docs hai trovato prima di procedere alla review.

**Se non trovati:**
> "Non ho trovato docs wm-plan per questo ticket (cercato in `docs/features/`). Procedo con la review sul diff senza contesto di intent."

---

## Fase 5 — Review

### 5a — Ottieni il diff

```bash
# Se hai il branch checkout:
git diff main...<branch-name> --stat
git diff main...<branch-name>

# Se hai solo il commit di merge:
git show <hash> --stat
git show <hash>
```

### 5b — Finder paralleli

Lancia **5 finder paralleli** sul diff. Ogni finder legge il diff con `git diff` e i file completi con `git show <hash>:<path>` o leggendo direttamente dal filesystem dopo il checkout.

**Finder 1 — Correctness vs richiesta ticket:**
Confronta il diff con `customer_request` e i Requisiti di `overview.md`. Il codice risponde a quello che il cliente ha chiesto? Ci sono requisiti non implementati o implementati parzialmente?

**Finder 2 — Side effect e bug:**
Cerca comportamenti non intenzionali: race condition, null pointer, edge case non gestiti, regressioni su funzionalità esistenti, modifiche a file non dichiarati nei "Moduli toccati" dell'overview.

**Finder 3 — Deviazioni non documentate:**
Confronta l'implementazione con `plan.md`. Ci sono task saltati o approcciati diversamente non registrati in `notes.md`? Le deviazioni in `notes.md` sono giustificate o rischiose?

**Finder 4 — Cleanup:**
Codice duplicato, naming incoerente, funzioni troppo lunghe, import non usati, magic number. Non bloccanti ma da segnalare.

**Finder 5 — Altitude:**
Design e architettura. Accoppiamento, responsabilità dei moduli, scelte che creano debito tecnico.

Ogni finder restituisce candidati: `{file, line, summary, failure_scenario, severity: blocker|cleanup}`.

### 5c — Dedup e verifica

Deduplica i candidati sovrapposti. Per ogni candidato non ovvio, verifica che il problema non sia già risolto a HEAD (`git show HEAD:<path>`).

### 5d — Output

Presenta in italiano:

```
## Verdetto

[Una frase: APPROVATO / APPROVATO CON RISERVE / DA CORREGGERE]

## Finding bloccanti
[Solo bug correctness o regressioni user-facing]

### [Titolo finding]
- **File:** `path/to/file.ext:123`
- **Problema:** [descrizione tecnica]
- **Scenario utente:** [cosa vive l'utente se il bug si manifesta]
- **Suggerimento:** [come correggerlo]

## Finding cleanup
[Non bloccanti — miglioramenti facoltativi]

## Finding confutati
[Candidati scartati dalla verifica, con motivazione]
```

---

## Fase 6 — Ripristino e aggiornamento ticket

### 6a — Ripristina stash (se creato)

```bash
if [ "$STASH_CREATED" = true ]; then
  git stash pop || echo "⚠️ git stash pop fallito — conflitti da risolvere manualmente. Lo stash è intatto: usa 'git stash list' per trovarlo."
fi
```

### 6b — Aggiorna ticket Orchestrator

Prepara il riepilogo tecnico della review (finding, esito, eventuali azioni richieste) e mostra il preview all'utente prima di inviare.

Leggi gli status disponibili da:
```
https://raw.githubusercontent.com/webmappsrl/orchestrator/main/app/Enums/StoryStatus.php
```

**Se nessun bloccante:**
> "Review completata senza finding bloccanti. Quale status vuoi impostare?"
> [Lista status da StoryStatus.php — suggerisci `testing` come default]

**Se ci sono bloccanti:**
> "Trovati [N] finding bloccanti. Propongo di impostare lo status a `todo` per richiedere correzioni. Confermo?"

Mostra sempre preview tabellare prima della PATCH:

> **Aggiornamento ticket oc:\<ID\>**
>
> | Campo | Valore |
> |-------|--------|
> | `status` | `<status scelto>` |
> | `description` | `<riepilogo review — max 3 righe preview>` |
>
> Procedo?

Solo dopo conferma esplicita:

```bash
ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-https://orchestrator.maphub.it}"
TOKEN=$(jq -r '.token' ~/.config/webmapp/orchestrator-auth.json)
curl -s -X PATCH "$ORCHESTRATOR_URL/api/stories/<ID>" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '{"status": "<status>", "description": "<riepilogo review>"}'
```
````

- [ ] **Step 3: Validare il plugin**

```bash
claude plugin validate .
```

Expected: nessun errore di schema su `wm-review-ticket`.

- [ ] **Step 4: Commit**

```bash
git add plugins/wm-skills/skills/wm-review-ticket/SKILL.md
git commit -m "feat(oc:8068): add wm-review-ticket skill"
```

---

### Task 2: Aggiornare `plugins/wm-skills/skills/wm-plan/SKILL.md`

**Files:**
- Modify: `plugins/wm-skills/skills/wm-plan/SKILL.md`

Due modifiche indipendenti:
1. Aggiungere sezione `## Contratto artefatti` (fonte autoritativa per wm-review-ticket)
2. Aggiornare Fase 6d per citare `wm-review-ticket`

- [ ] **Step 1: Aggiungere sezione `## Contratto artefatti` prima di `## Fase 0`**

Inserire questo blocco subito dopo il frontmatter YAML (dopo `---`) e prima della prima sezione `## Fase 0`:

```markdown
## Contratto artefatti

Questa sezione è la fonte autoritativa degli artefatti prodotti da `wm-plan`. `wm-skills:wm-review-ticket` la legge via WebFetch per conoscere la struttura senza duplicarla.

Per ogni feature lavorata con `wm-plan`, vengono creati i seguenti file nella cartella `docs/features/<feature-slug>/` del repo target:

| File | Contenuto | Usato da wm-review-ticket per |
|------|-----------|-------------------------------|
| `overview.md` | Cosa cambia, Perché, Requisiti, Rischi, Out of scope, Moduli toccati | Criterio principale di correttezza: il codice risponde ai Requisiti? |
| `plan.md` | Lista task implementativi step-by-step | Verifica completezza: tutti i task sono stati eseguiti? |
| `notes.md` | Deviazioni dal piano, bug trovati, decisioni on-the-fly | Contesto deviazioni: sono giustificate o rischiose? |

**Slug:** `<ID>-<titolo-in-kebab-case>` (es. `8068-wm-review-ticket`). Se non c'è ticket, solo `<titolo-in-kebab-case>`.

**Ricerca fuzzy:** per trovare la cartella da un ID ticket:
```bash
find docs/features/ -maxdepth 1 -type d | grep "<ID>"
```

---
```

- [ ] **Step 2: Aggiornare Fase 6d — aggiungere riferimento a `wm-review-ticket`**

Trova il blocco del gate di revisione in Fase 6d e aggiungi dopo il messaggio di conferma al developer:

```markdown
> **Suggerimento review formale:** prima dei commit, puoi invocare `wm-skills:wm-review-ticket oc:<ID>` per una code review strutturata con finder paralleli e aggiornamento automatico del ticket.
```

- [ ] **Step 3: Validare il plugin**

```bash
claude plugin validate .
```

Expected: nessun errore di schema su `wm-plan`.

- [ ] **Step 4: Commit**

```bash
git add plugins/wm-skills/skills/wm-plan/SKILL.md
git commit -m "feat(oc:8068): add artefatti contract section and wm-review-ticket reference to wm-plan"
```

---

### Task 3: Aggiornare `CLAUDE.md` del marketplace

**Files:**
- Modify: `CLAUDE.md`

- [ ] **Step 1: Aggiungere `wm-review-ticket` alla tabella "Skill wm-skills disponibili"**

Trova la sezione `## Skill wm-skills disponibili` e aggiungi la riga:

```markdown
| `wm-review-ticket` | Eseguire la code review di un ticket Orchestrator — sia quando un collega assegna un ticket da rivedere, sia al termine di una feature wm-plan prima del merge. |
```

- [ ] **Step 2: Commit**

```bash
git add CLAUDE.md
git commit -m "docs(oc:8068): register wm-review-ticket in marketplace CLAUDE.md"
```

---

### Task 4: Test e validazione finale

- [ ] **Step 1: Installare il marketplace locale**

```bash
/plugin marketplace add .
/plugin install wm-skills@wm-marketplace
```

- [ ] **Step 2: Verificare che la skill appaia**

In una nuova sessione Claude Code, verifica che `wm-review-ticket` sia elencata tra le skill disponibili.

- [ ] **Step 3: Test smoke — invocazione con ticket reale**

Invoca la skill con un ticket noto che ha una PR associata:
```
wm-skills:wm-review-ticket oc:<ID-ticket-con-PR>
```

Verifica che:
- Legga il ticket da Orchestrator
- Trovi l'URL PR nella description
- Tenti il checkout del branch
- Cerchi i docs wm-plan in `docs/features/`
- Presenti il verdetto in italiano

- [ ] **Step 4: Tornare alla versione remota dopo i test**

```bash
/plugin marketplace remove wm-marketplace
/plugin marketplace add webmappsrl/claude-marketplace
```
