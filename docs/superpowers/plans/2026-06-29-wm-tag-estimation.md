> Ticket: oc:8157

# wm-tag skill e fase estimation in wm-plan — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Creare la skill `wm-tag` che trasforma trascrizioni/brief cliente in tag Orchestrator con ticket figli analizzati; aggiungere a `wm-plan` il `caso-c` (entry point verso wm-tag), la `Fase: estimation` per Feature, e la modalità `tag-mode`.

**Architecture:** Tre modifiche indipendenti su due file SKILL.md esistenti più un file nuovo. Le skill sono documenti Markdown con frontmatter YAML — nessun codice da compilare, nessun test automatizzato. La validazione avviene tramite `claude plugin validate .` e test manuale in sessione locale.

**Tech Stack:** Markdown, YAML frontmatter, API REST Orchestrator (curl + jq), Google Drive MCP tool, plugin system Claude Code.

## Global Constraints

- Prefisso obbligatorio `wm-` per ogni skill wm-skills
- Commit scope: `feat(oc:8157): ...`
- Ogni operazione POST/PATCH su Orchestrator richiede preview tabellare + conferma esplicita, senza eccezioni
- La `Fase: estimation` si esegue solo per ticket di tipo Feature, mai per Bug o altri tipi
- In tag-mode l'overview non va mai nel repo filesystem — solo nel campo `description` del ticket
- Naming tag: `[RDO][CLIENTE][ANNO]<N>` — N calcolato dinamicamente cercando tag esistenti via API
- `~/.config/webmapp/repos.json` aggiornato incrementalmente, mai riscritto da zero

---

## File map

| File | Azione | Responsabilità |
|---|---|---|
| `plugins/wm-skills/skills/wm-tag/SKILL.md` | Creare | Intera skill wm-tag: input, repo-map, client-extraction, tag-naming, tag-description, tag-creation, ticket-list, ticket-loop |
| `plugins/wm-skills/skills/wm-plan/SKILL.md` | Modificare | Tre aggiunte: caso-c in Fase: ticket, Fase: estimation, comportamento tag-mode |
| `CLAUDE.md` (root marketplace) | Modificare | Tabella coupling aggiornata, sezione Feature disponibili, tabella skill wm-skills |

---

## Task 1: Modifica wm-plan — caso-c nella Fase: ticket

**Files:**
- Modify: `plugins/wm-skills/skills/wm-plan/SKILL.md` (sezione `## Fase: ticket`, all'inizio)

**Interfaces:**
- Produces: quando l'utente sceglie C al menu iniziale, `wm-plan` invoca `wm-skills:wm-tag` e cede il controllo

- [ ] **Step 1: Leggi la sezione Fase: ticket di wm-plan**

  Apri `plugins/wm-skills/skills/wm-plan/SKILL.md` e individua la sezione `## Fase: ticket`. Attualmente presenta `caso-a` (ticket esistente) e `caso-b` (nuovo ticket). Identifica il punto esatto dove aggiungere `caso-c`.

- [ ] **Step 2: Aggiungi l'introduzione con menu a tre opzioni**

  All'inizio della sezione `## Fase: ticket`, prima di `### ticket: caso-a`, inserisci:

  ````markdown
  All'inizio del workflow, presenta sempre questo menu all'utente:

  > Come vuoi procedere?
  > - **A)** Ho un ticket esistente (`oc:<ID>`)
  > - **B)** Voglio creare un nuovo ticket
  > - **C)** Ho una trascrizione/brief cliente → crea tag con più ticket

  In base alla scelta:
  - **A** → segui `### ticket: caso-a`
  - **B** → segui `### ticket: caso-b`
  - **C** → segui `### ticket: caso-c`
  ````

- [ ] **Step 3: Aggiungi la sottosezione ticket: caso-c**

  Dopo la sottosezione `### ticket: caso-b` e prima di `### ticket: aggiornamenti-espliciti`, inserisci:

  ````markdown
  ### ticket: caso-c

  L'utente ha una trascrizione Meet, un brief cliente o qualsiasi materiale da cui estrarre più ticket raggruppati in un tag Orchestrator.

  Invoca immediatamente `wm-skills:wm-tag`, passando come contesto qualsiasi testo o link già fornito dall'utente in questa sessione. Da questo momento il controllo del flusso passa interamente a `wm-tag` — non proseguire con nessun'altra fase di `wm-plan`.
  ````

- [ ] **Step 4: Valida la sintassi del file**

  ```bash
  claude plugin validate .
  ```

  Atteso: nessun errore. Se ci sono errori di frontmatter o struttura, correggili prima di procedere.

- [ ] **Step 5: Commit**

  ```bash
  git add plugins/wm-skills/skills/wm-plan/SKILL.md
  git commit -m "feat(oc:8157): add caso-c to wm-plan Fase: ticket — entry point to wm-tag"
  ```

---

## Task 2: Modifica wm-plan — Fase: estimation

**Files:**
- Modify: `plugins/wm-skills/skills/wm-plan/SKILL.md` (nuova sezione dopo `## Fase: challenge`, prima di `## Fase: write-plan`)

**Interfaces:**
- Consumes: overview approvato (da file `docs/features/<slug>/overview.md` in modalità normale, o da description ticket in tag-mode); campo `type` del ticket Orchestrator
- Produces: valore numerico `estimated_hours` scritto nel ticket via PATCH; stima visibile all'utente con motivazione

- [ ] **Step 1: Individua il punto di inserimento**

  In `plugins/wm-skills/skills/wm-plan/SKILL.md` trova la riga `## Fase: write-plan`. La nuova sezione va inserita immediatamente prima di essa.

- [ ] **Step 2: Inserisci la sezione Fase: estimation**

  Inserisci il seguente blocco Markdown immediatamente prima di `## Fase: write-plan`:

  ````markdown
  ---

  ## Fase: estimation

  **Eseguita solo se il ticket è di tipo Feature.** Per Bug, Task o altri tipi, salta questa fase e procedi direttamente a `Fase: write-plan`.

  In tag-mode, questa fase viene eseguita prima di fermarsi (non si procede a write-plan).

  ### estimation: analisi

  Basandoti sull'overview approvato, produci una stima ragionata in ore con questa struttura:

  > **Stima proposta: \<N\> ore**
  >
  > Motivazione:
  > - \<componente 1\>: \<X\>h — \<motivazione tecnica\>
  > - \<componente 2\>: \<Y\>h — \<motivazione tecnica\>
  > - Buffer rischio: \<Z\>h — \<motivazione: complessità, dipendenze esterne, incertezze\>
  >
  > Confidenza: alta / media / bassa
  > *(alta = requisiti chiari e stack noto; media = qualche incertezza tecnica; bassa = dipendenze esterne o requisiti aperti)*

  Regole per la stima:
  - Includi sempre un buffer rischio (minimo 10% del totale, mai zero)
  - La confidenza deve essere coerente con i rischi emersi nella Fase: challenge
  - Non stimare meno di 1h per qualsiasi feature che tocchi più di un file

  ### estimation: conferma

  Chiedi al dev:

  > "Accetti questa stima di **\<N\> ore**, o vuoi modificarla?"

  Aspetta risposta esplicita. Se il dev propone un valore diverso, usalo senza discutere — la stima finale è sempre quella approvata dal dev.

  ### estimation: scrittura su Orchestrator

  Mostra il preview della modifica e chiedi conferma prima di eseguire:

  > **Aggiornamento ticket oc:\<ID\>**
  >
  > | Campo | Valore |
  > |-------|--------|
  > | `estimated_hours` | `<N>` |
  >
  > Procedo?

  Solo dopo la conferma:

  ```bash
  ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-https://orchestrator.maphub.it}"
  TOKEN=$(jq -r '.token' ~/.config/webmapp/orchestrator-auth.json)
  curl -s -X PATCH "$ORCHESTRATOR_URL/api/stories/<ID>" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -H "Accept: application/json" \
    -d '{"estimated_hours": <N>}'
  ```

  Se il PATCH fallisce (risposta non 2xx), avvisa l'utente con `⚠️ Impossibile aggiornare la stima su Orchestrator — procedo comunque.` e continua.

  ---
  ````

- [ ] **Step 3: Valida**

  ```bash
  claude plugin validate .
  ```

  Atteso: nessun errore.

- [ ] **Step 4: Commit**

  ```bash
  git add plugins/wm-skills/skills/wm-plan/SKILL.md
  git commit -m "feat(oc:8157): add Fase: estimation to wm-plan for Feature tickets"
  ```

---

## Task 3: Modifica wm-plan — comportamento tag-mode

**Files:**
- Modify: `plugins/wm-skills/skills/wm-plan/SKILL.md` (nuova sezione `## Modalità tag-mode` da aggiungere dopo `## Contratto artefatti`)

**Interfaces:**
- Consumes: contesto passato da `wm-tag` — titolo ticket, tipo, repo di destinazione, ID tag padre, flag `tag-mode: true`
- Produces: ticket Orchestrator creato con `description` contenente `## Overview`; ticket associato al tag padre via `tags: [<tag-id>]`

- [ ] **Step 1: Individua il punto di inserimento**

  Trova la sezione `## Orchestrator API` in `wm-plan/SKILL.md`. La nuova sezione `## Modalità tag-mode` va inserita immediatamente prima di essa.

- [ ] **Step 2: Inserisci la sezione**

  ````markdown
  ---

  ## Modalità tag-mode

  `wm-plan` può essere invocato da `wm-skills:wm-tag` in **tag-mode**. In questo caso riceve nel contesto della conversazione:
  - Titolo del ticket da creare
  - Tipo (Feature / Bug / Task)
  - Repo di destinazione (ricavabile da `~/.config/webmapp/repos.json`)
  - ID del tag padre su Orchestrator
  - Flag esplicito `tag-mode: true`

  ### tag-mode: flusso

  In tag-mode, `wm-plan` esegue **solo** queste fasi nell'ordine:

  1. `Fase: ticket` — crea il ticket su Orchestrator (usando il titolo e tipo già forniti da `wm-tag`)
  2. `Fase: environment-setup` — rileva stack e repo (consulta `repos.json` per la path del repo di destinazione)
  3. `Fase: init-context` — legge il CLAUDE.md del repo di destinazione
  4. `Fase: reverse-interaction` — dialogo socratico completo (minimo 5 domande, nessuna eccezione)
  5. `Fase: overview` — produce l'overview con la struttura canonica
  6. `Fase: challenge` — analisi adversariale sull'overview
  7. `Fase: estimation` — solo se Feature (stima in ore, approvata dal dev)
  8. **Scrittura overview nel ticket** — vedi sezione sotto
  9. **Stop** — restituisce il controllo a `wm-tag`

  Le fasi `write-plan`, `execution`, `notes`, `update-context` **non vengono eseguite**.

  **Importante:** l'overview **non viene salvata nel filesystem** del repo. Nessun file `docs/features/` viene creato o modificato.

  ### tag-mode: scrittura overview nel ticket

  Dopo l'approvazione dell'overview (e dell'estimation se Feature), costruisci il payload per il PATCH del ticket. La `description` del ticket deve contenere il testo completo dell'overview in una sezione `## Overview`:

  ```
  ## Overview

  ### Cosa cambia
  <testo>

  ### Perché
  <testo>

  ### Requisiti
  - [ ] ...

  ### Rischi
  <testo>

  ### Out of scope
  <testo>

  ### Moduli toccati
  <testo>
  ```

  Mostra preview e chiedi conferma (regola generale scritture):

  > **Aggiornamento ticket oc:\<ID\>**
  >
  > | Campo | Valore |
  > |-------|--------|
  > | `description` | `## Overview\n### Cosa cambia\n<prime 200 caratteri...>` |
  > | `tags` | `[<tag-id>]` |
  >
  > Procedo?

  Solo dopo la conferma:

  ```bash
  ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-https://orchestrator.maphub.it}"
  TOKEN=$(jq -r '.token' ~/.config/webmapp/orchestrator-auth.json)
  curl -s -X PATCH "$ORCHESTRATOR_URL/api/stories/<ID>" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -H "Accept: application/json" \
    -d '{"description": "<overview-markdown>", "tags": [<tag-id>]}'
  ```

  Al termine mostra: `✅ Ticket oc:<ID> aggiornato e associato al tag <nome-tag>.`

  ---
  ````

- [ ] **Step 3: Valida**

  ```bash
  claude plugin validate .
  ```

- [ ] **Step 4: Commit**

  ```bash
  git add plugins/wm-skills/skills/wm-plan/SKILL.md
  git commit -m "feat(oc:8157): add tag-mode to wm-plan — overview to ticket description, skip write-plan+"
  ```

---

## Task 4: Crea skill wm-tag

**Files:**
- Create: `plugins/wm-skills/skills/wm-tag/SKILL.md`

**Interfaces:**
- Consumes: testo in chat o link Google Drive; `~/.config/webmapp/orchestrator-auth.json`; `~/.config/webmapp/repos.json`
- Produces: tag Orchestrator con descrizione; ticket figli con overview in description e stima; controllo ceduto a `wm-plan` per ogni ticket

- [ ] **Step 1: Crea la cartella**

  ```bash
  mkdir -p plugins/wm-skills/skills/wm-tag
  ```

- [ ] **Step 2: Scrivi SKILL.md**

  Crea `plugins/wm-skills/skills/wm-tag/SKILL.md` con il seguente contenuto:

  ````markdown
  ---
  name: wm-tag
  description: "Usa quando hai una trascrizione Meet, un brief cliente o qualsiasi materiale che descrive richieste da trasformare in ticket Orchestrator raggruppati in un tag. Analizza il materiale, crea il tag con le macro aree, e per ogni task identificato esegue il flusso wm-plan in modalità tag-mode (overview → description del ticket, poi stop)."
  ---

  # wm-tag — Trascrizione cliente → Tag + Ticket Orchestrator

  Questa skill trasforma un brief o una trascrizione cliente in un tag Orchestrator con ticket figli strutturati e stimati. Per ogni ticket invoca `wm-skills:wm-plan` in tag-mode, che esegue il flusso completo di analisi (fino all'overview e alla stima) senza toccare il filesystem del repo.

  ---

  ## Orchestrator API

  Segui le stesse regole di `wm-skills:wm-plan` per auth, login e migrazione da file legacy. URL base: `${ORCHESTRATOR_URL:-https://orchestrator.maphub.it}`. Auth: `~/.config/webmapp/orchestrator-auth.json`.

  **Regola generale scritture:** qualsiasi POST o PATCH — tag o story — richiede sempre un preview tabellare dei campi e conferma esplicita dell'utente prima di eseguire la chiamata HTTP. Nessuna eccezione.

  ### API Tag

  **Lista tag (ricerca):**
  ```bash
  curl -s "$ORCHESTRATOR_URL/api/tags?search=<query>" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Accept: application/json"
  ```

  **Creazione tag:**
  ```bash
  curl -s -X POST "$ORCHESTRATOR_URL/api/tags" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -H "Accept: application/json" \
    -d '{"name": "<nome>", "description": "<descrizione>"}'
  ```

  Salva l'`id` restituito — serve per associare i ticket al tag.

  ---

  ## Fase: input

  Accetta due formati:

  - **Testo in chat** — l'utente incolla direttamente la trascrizione o il brief. Fonte: `"Testo fornito in chat"`.
  - **Link Google Drive** — usa `mcp__claude_ai_Google_Drive__read_file_content` se disponibile, altrimenti WebFetch sull'URL. Fonte: l'URL completo del documento.

  Se l'utente non ha ancora fornito il materiale, chiedi: "Incolla la trascrizione o fornisci il link Google Drive del documento."

  Tieni traccia della fonte — verrà linkata in testa alla descrizione del tag.

  ---

  ## Fase: repo-map

  Prima di qualsiasi analisi, costruisce o aggiorna `~/.config/webmapp/repos.json`.

  ```bash
  # Individua la cartella padre del working directory
  PARENT=$(dirname "$PWD")

  # Trova tutti i repo git nella cartella padre
  find "$PARENT" -maxdepth 2 -name ".git" -type d | sed 's|/.git||'
  ```

  **Algoritmo:**
  1. Leggi `~/.config/webmapp/repos.json` se esiste (struttura: `{"nome-repo": "/path/assoluta"}`)
  2. Per ogni repo trovato non presente nel dizionario, aggiungi `"basename-cartella": "/path/assoluta"`
  3. Salva il file aggiornato:
     ```bash
     # Esempio di aggiornamento incrementale con jq
     jq --arg name "<basename>" --arg path "<path>" '. + {($name): $path}' \
       ~/.config/webmapp/repos.json > /tmp/repos_tmp.json && \
       mv /tmp/repos_tmp.json ~/.config/webmapp/repos.json
     ```
  4. Mostra all'utente la mappa completa:
     > **Repository disponibili:**
     > | Nome | Path |
     > |------|------|
     > | `<nome>` | `<path>` |

  Questa mappa è disponibile per tutto il flusso. Quando serve ispezionare codice in un repo specifico, leggi la path da `repos.json` e naviga lì.

  ---

  ## Fase: client-extraction

  Analizza il testo e identifica il nome del cliente.

  - Cerca ricorrenze di nomi propri, ragioni sociali, acronimi usati come riferimento al committente
  - Proponi il nome trovato con motivazione:
    > "Ho trovato il nome cliente: **CAMMINI** — citato 6 volte nel testo come destinatario delle richieste. È corretto?"
  - Attendi conferma esplicita. Se l'utente corregge, usa il nome corretto per tutto il resto del flusso.
  - Il nome cliente viene usato in MAIUSCOLO nel naming del tag (es. `CAMMINI`, `PARCHI`, `REGIONE-VENETO`).

  ---

  ## Fase: tag-naming

  Costruisce il nome del tag seguendo la convenzione: `[RDO][CLIENTE][ANNO]<N>`

  Esempio: `[RDO][CAMMINI][2026]1`

  **Calcolo di N:**

  ```bash
  ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-https://orchestrator.maphub.it}"
  TOKEN=$(jq -r '.token' ~/.config/webmapp/orchestrator-auth.json)
  ANNO=$(date +%Y)

  # Cerca tag esistenti per questo cliente e anno
  curl -s "$ORCHESTRATOR_URL/api/tags?search=RDO" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Accept: application/json" | \
    jq --arg cliente "<CLIENTE>" --arg anno "$ANNO" \
    '[.[] | select(.name | contains($cliente) and contains($anno))] | length'
  ```

  Il valore restituito è il numero di tag esistenti per quel cliente+anno. `N = conteggio + 1`.

  Proponi il nome al dev e attendi conferma:
  > "Nome tag proposto: `[RDO][CAMMINI][2026]2` (esistono già 1 tag per CAMMINI nel 2026). Confermo?"

  ---

  ## Fase: tag-description

  Analizza il testo del brief/trascrizione e produce la descrizione del tag in Markdown:

  ```markdown
  **Fonte:** <URL Drive o "Testo fornito in chat">

  ## Macro aree

  - **<Area 1>**: <descrizione sintetica — cosa vuole il cliente in quest'area>
  - **<Area 2>**: <descrizione sintetica>
  ...

  ## Informazioni principali

  <Contesto generale: chi è il cliente, obiettivo del progetto, vincoli noti (tecnologici, temporali, di budget), deadline esplicite se presenti nel testo>
  ```

  Mostra la descrizione all'utente e attendi approvazione esplicita prima di creare il tag.

  ---

  ## Fase: tag-creation

  Mostra il preview tabellare e chiedi conferma:

  > **Creazione tag**
  >
  > | Campo | Valore |
  > |-------|--------|
  > | `name` | `[RDO][CLIENTE][ANNO]N` |
  > | `description` | `<prime 200 caratteri della descrizione>...` |
  >
  > Procedo?

  Solo dopo la conferma, esegui il POST:

  ```bash
  ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-https://orchestrator.maphub.it}"
  TOKEN=$(jq -r '.token' ~/.config/webmapp/orchestrator-auth.json)
  curl -s -X POST "$ORCHESTRATOR_URL/api/tags" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -H "Accept: application/json" \
    -d "{\"name\": \"<nome-tag>\", \"description\": \"<descrizione-escaped>\"}"
  ```

  Salva l'`id` restituito come `<TAG_ID>` per l'associazione dei ticket.
  Al termine: `✅ Tag \`<nome-tag>\` creato (ID: <TAG_ID>).`

  ---

  ## Fase: ticket-list

  Analizza il testo e individua tutti i task distinti che il cliente ha richiesto. Per ogni task identifica:
  - Titolo sintetico
  - Tipo: Feature (nuova funzionalità), Bug (problema da risolvere), Task (attività tecnica senza utente finale diretto)
  - Repo di destinazione stimato (backend / frontend / altro — usando `repos.json` come riferimento)

  Presenta la lista proposta in formato tabellare:

  | # | Titolo ticket | Tipo | Repo |
  |---|---|---|---|
  | 1 | ... | Feature | backend |
  | 2 | ... | Bug | frontend |

  L'utente può:
  - **Approvare** la lista così com'è
  - **Modificare** un titolo o tipo
  - **Unire** due ticket in uno
  - **Eliminare** un ticket dalla lista
  - **Aggiungere** un ticket non rilevato

  Attendi approvazione esplicita della lista prima di procedere al loop.

  ---

  ## Fase: ticket-loop

  Per ogni ticket nella lista approvata, nell'ordine:

  1. Annuncia: "Processo ticket \<N\>/\<TOT\>: **\<titolo\>**"
  2. Invoca `wm-skills:wm-plan` passando questo contesto:
     - Titolo del ticket
     - Tipo (Feature / Bug / Task)
     - Repo di destinazione (path da `repos.json`)
     - ID tag padre (`<TAG_ID>`)
     - Flag `tag-mode: true`
  3. `wm-plan` esegue il flusso completo in tag-mode (reverse-interaction, overview, challenge, estimation se Feature) e scrive l'overview nella description del ticket Orchestrator associandolo al tag
  4. Al termine di ogni ticket, chiedi:
     > "Ticket \<N\>/\<TOT\> completato. Procedo con il prossimo (**\<titolo-prossimo\>**), o vuoi fermarti qui?"
  5. Se l'utente vuole fermarsi, interrompi il loop. I ticket rimanenti restano in lista e possono essere ripresi in una sessione successiva rilanciando `wm-skills:wm-tag` con lo stesso tag ID.
  ````

- [ ] **Step 3: Valida**

  ```bash
  claude plugin validate .
  ```

  Atteso: nessun errore su `wm-tag/SKILL.md`.

- [ ] **Step 4: Commit**

  ```bash
  git add plugins/wm-skills/skills/wm-tag/SKILL.md
  git commit -m "feat(oc:8157): add wm-tag skill — transcript to Orchestrator tag + tickets"
  ```

---

## Task 5: Aggiorna CLAUDE.md marketplace

**Files:**
- Modify: `CLAUDE.md` (root del repo `claude-marketplace`)

**Interfaces:**
- Consumes: nessuna dipendenza da altri task (può essere eseguito in parallelo al Task 4)
- Produces: tabella coupling aggiornata, skill wm-tag nella tabella, Feature disponibili aggiornata

- [ ] **Step 1: Aggiorna la tabella coupling**

  Nella sezione `### Coupling tra skill` di `CLAUDE.md`, aggiungi due righe alla tabella:

  | Skill A | Skill B | Contratto condiviso |
  |---|---|---|
  | `wm-tag` | `wm-plan` | `wm-tag` invoca `wm-plan` in tag-mode passando titolo, tipo, repo e TAG_ID. `wm-plan` è responsabile di reverse-interaction, overview, challenge, estimation e scrittura della description del ticket. `wm-tag` gestisce tag, lista ticket e loop. |
  | `wm-plan` | `wm-tag` | `caso-c` in Fase: ticket switcha su `wm-tag` cedendo il controllo. |

- [ ] **Step 2: Aggiorna la tabella Skill wm-skills disponibili**

  Nella sezione `## Skill wm-skills disponibili`, aggiungi la riga:

  | `wm-tag` | Analizzare trascrizioni/brief cliente per creare tag Orchestrator con ticket figli strutturati e stimati. |

- [ ] **Step 3: Aggiorna la sezione Decisioni architetturali**

  Aggiungi il blocco in cima alla sezione `## Decisioni architetturali`:

  ````markdown
  ### wm-tag skill e fase estimation in wm-plan (oc:8157)
  - **tag-mode in wm-plan**: quando invocato da `wm-tag`, `wm-plan` salta write-plan/execution/notes/update-context — l'overview va nella description del ticket, non nel filesystem
  - **Fase: estimation solo per Feature**: i bug non si stimano in ore (costo nella diagnosi, non nella fix) — solo Feature ricevono `estimated_hours`
  - **repos.json per navigazione multi-repo**: dizionario persistente `~/.config/webmapp/repos.json` aggiornato incrementalmente — non riscritto da zero per preservare path manuali
  - **Naming tag `[RDO][CLIENTE][ANNO]N`**: N calcolato dinamicamente contando tag esistenti per stesso cliente+anno su Orchestrator — evita conflitti senza coordinazione manuale
  - **Regola scritture estesa a tag**: la regola preview+conferma di `wm-plan` per le story si applica identicamente a tutti i POST/PATCH su Orchestrator, incluse le operazioni sui tag
  ````

- [ ] **Step 4: Aggiorna la sezione Feature disponibili**

  Nella tabella `## Feature disponibili`, aggiungi la riga:

  | wm-tag skill e fase estimation in wm-plan | oc:8157 | `plugins/wm-skills/skills/wm-tag/SKILL.md`, `plugins/wm-skills/skills/wm-plan/SKILL.md` | Nuova skill `wm-tag` per trascrizione → tag + ticket; `caso-c` in Fase: ticket; `Fase: estimation` per Feature; tag-mode in wm-plan |

- [ ] **Step 5: Commit**

  ```bash
  git add CLAUDE.md
  git commit -m "docs(oc:8157): update CLAUDE.md — wm-tag coupling, skill table, arch decisions"
  ```

---

## Task 6: Validazione finale e test locale

**Files:**
- Nessun file modificato — solo verifica

- [ ] **Step 1: Validazione plugin**

  ```bash
  claude plugin validate .
  ```

  Atteso: nessun errore su nessun file.

- [ ] **Step 2: Installa il marketplace locale**

  In una sessione Claude Code separata, dalla root del repo:

  ```
  /plugin marketplace add .
  /plugin install wm-skills@wm-marketplace
  ```

- [ ] **Step 3: Test smoke — wm-plan caso-c**

  Apri una nuova sessione Claude Code e invoca `/wm-skills:wm-plan`. Verifica che il menu mostri tre opzioni (A, B, C). Scegli C e verifica che Claude passi il controllo a `wm-tag`.

- [ ] **Step 4: Test smoke — wm-tag input**

  Invoca `/wm-skills:wm-tag` e incolla un testo di test con richieste fittizie di un cliente. Verifica che la skill:
  - Costruisca `repos.json` mostrando i repo trovati
  - Estragga il nome cliente e chieda conferma
  - Proponga il nome tag nella convenzione `[RDO][CLIENTE][ANNO]N`
  - Produca la descrizione del tag con macro aree
  - Mostri il preview prima di creare il tag

- [ ] **Step 5: Test smoke — Fase: estimation in wm-plan**

  Avvia un flusso `wm-plan` su un ticket Feature esistente. Verifica che dopo la Fase: challenge compaia la Fase: estimation con stima in ore e richiesta di conferma prima del PATCH su Orchestrator.

- [ ] **Step 6: Ripristina marketplace remoto**

  ```
  /plugin marketplace remove wm-marketplace
  /plugin marketplace add webmappsrl/claude-marketplace
  ```
