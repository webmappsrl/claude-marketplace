# Split automatico ticket Help desk multi-richiesta in wm-plan — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Far sì che `wm-plan`, quando legge in `caso-a` un ticket Orchestrator di tipo Help desk contenente più richieste distinte del cliente, proponga automaticamente di splittarlo in più ticket tracciabili, mantenendo il ticket originale ridotto alla prima richiesta.

**Architecture:** Nessun codice eseguibile — questa feature è interamente prosa Markdown dentro `plugins/wm-skills/skills/wm-plan/SKILL.md`, letta e seguita da Claude a runtime. "Testare" significa: (a) validare la sintassi/schema del plugin con `claude plugin validate .`, (b) verificare per lettura che il testo copra ogni requisito dello spec, (c) verificare la coerenza dei riferimenti incrociati (nomi di sotto-fasi, endpoint, campi) con il resto del file.

**Tech Stack:** Markdown (frontmatter YAML + corpo), Bash/curl per le chiamate Orchestrator API, `jq` per parsing JSON, `claude plugin validate` come unico tool di verifica automatica.

## Global Constraints

- Prefisso e naming: nessuna nuova skill creata, si modifica solo `wm-plan` esistente — nessun vincolo di naming aggiuntivo.
- Ogni scrittura HTTP (POST/PATCH) su Orchestrator richiede sempre preview tabellare + conferma esplicita, nessuna eccezione (regola generale già esistente in `## Orchestrator API → Regola generale scritture`).
- Il testo di ogni partizione deve essere **verbatim** — nessuna riformulazione del testo del cliente.
- `type` dei nuovi ticket è sempre `"Help desk"` — nessuna riclassificazione automatica.
- Prima di pubblicare, eseguire `claude plugin validate .` dalla root del repo (da `CLAUDE.md` → `## Validare prima del commit`).
- Dopo qualsiasi modifica al repo, rigenerare l'Artifact del diagramma di flusso `wm-plan` con lo stesso `file_path` (redeploy sullo stesso URL) — regola obbligatoria da `CLAUDE.md` → `## Diagramma di flusso wm-plan`.
- Bump di `version` in `plugins/wm-skills/.claude-plugin/plugin.json` richiesto ad ogni release, secondo la checklist in `CLAUDE.md` → `## Versioning del plugin wm-skills`. Questa è una nuova sotto-fase retro-compatibile in una skill esistente → bump **minor**.

---

### Task 1: Aggiungere la sotto-fase di rilevamento split in `ticket: caso-a`

**Files:**
- Modify: `plugins/wm-skills/skills/wm-plan/SKILL.md` (sezione `### ticket: caso-a`, subito dopo il blocco `Dal JSON restituito estrai:` e prima della riga `Mostra all'utente un riepilogo del ticket letto prima di procedere alla Fase: init-context.`)

**Interfaces:**
- Consumes: risposta JSON di `GET /api/stories/<ID>` già disponibile in `caso-a` (campi `id`, `name`, `type`, `customer_request`, `creator_id`, `tags`)
- Produces: sotto-fase referenziabile come `ticket: caso-a-split-detection`, che se non scatta lascia invariato il flusso esistente (prosegue al riepilogo standard di `caso-a`)

- [ ] **Step 1: Leggi il contesto esatto del punto di inserimento**

Apri `plugins/wm-skills/skills/wm-plan/SKILL.md` e individua il blocco esatto (già noto da esplorazione precedente):

```markdown
Dal JSON restituito estrai:
- `name` → titolo, usato per il `<feature-slug>`: `<ID>-<titolo-in-kebab-case>`
- `customer_request` → contesto del problema, usato in Fase: reverse-interaction e Fase: overview
- `description` → note tecniche già raccolte, può orientare le domande in Fase: reverse-interaction
- `type` → orienta il tono dell'overview

Mostra all'utente un riepilogo del ticket letto prima di procedere alla Fase: init-context.
```

- [ ] **Step 2: Inserisci la nuova sotto-fase tra le due righe**

Usa l'editor per inserire questo blocco subito dopo `- \`type\` → orienta il tono dell'overview` e prima di `Mostra all'utente un riepilogo...`:

```markdown

Se `type == "Help desk"`, esegui `### ticket: caso-a-split-detection` prima di mostrare qualunque riepilogo. Altrimenti procedi direttamente al riepilogo standard sotto.

### ticket: caso-a-split-detection

Analizza `customer_request` e determina, con giudizio diretto sul testo (nessuna euristica meccanica su keyword o conteggio paragrafi), se contiene **più richieste distinte** scritte dal cliente in un unico ticket.

- **Se rilevi una sola richiesta:** non fare nulla, prosegui al riepilogo standard di `caso-a` come oggi.
- **Se rilevi N ≥ 2 richieste distinte:** estrai N partizioni di testo **verbatim** (nessuna riformulazione), in ordine di apparizione nel testo originale. Genera anche un titolo sintetico per ciascuna partizione. Poi mostra:

  > **Rilevate {N} richieste distinte nel ticket oc:{ID} ("{name originale}"):**
  >
  > | # | Ruolo | Titolo proposto | Testo (verbatim) |
  > |---|---|---|---|
  > | 1 | Ticket originale (rinominato) | `{titolo sintetico 1}` | "{partizione 1}" |
  > | 2 | Nuovo ticket | `{titolo sintetico 2}` | "{partizione 2}" |
  > | ... | | | |
  >
  > Procedo con lo split? (puoi modificare titoli o testo prima di confermare)

  Attendi conferma esplicita.

  - **Se l'utente rifiuta:** il ticket originale resta intatto, prosegui al riepilogo standard di `caso-a` come oggi (nessuno split).
  - **Se l'utente conferma (con eventuali modifiche a titoli/testo):** procedi con `### ticket: caso-a-split-execution`.

### ticket: caso-a-split-execution

Esegui le scritture seguendo comunque `## Orchestrator API → Regola generale scritture` (preview tabellare + conferma esplicita per ogni singola chiamata, anche se il contenuto è già stato approvato nel riepilogo di `caso-a-split-detection`):

1. **PATCH sul ticket originale `oc:{ID}`:**

```bash
ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-https://orchestrator.maphub.it}"
TOKEN=$(jq -r '.token' ~/.config/webmapp/orchestrator-auth.json)
curl -s -X PATCH "$ORCHESTRATOR_URL/api/stories/<ID>" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '{"name": "<titolo sintetico 1>", "customer_request": "<partizione 1 verbatim>"}'
```

Tag e `creator_id` del ticket originale non vengono toccati da questa PATCH.

2. **POST per ciascuna partizione 2..N:**

```bash
ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-https://orchestrator.maphub.it}"
TOKEN=$(jq -r '.token' ~/.config/webmapp/orchestrator-auth.json)
curl -s -X POST "$ORCHESTRATOR_URL/api/stories" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '{"name": "<titolo sintetico N>", "type": "Help desk", "customer_request": "<partizione N verbatim>", "creator_id": <creator_id originale>, "tags": [<id tag originali>]}'
```

`type` è sempre `"Help desk"` — nessuna riclassificazione automatica in questa fase. Salva l'`id` restituito per ogni ticket creato. Mostra `✅ Ticket oc:<nuovo-ID> creato.` per ciascuno.

3. **Tag di raggruppamento opzionale (uso interno dev):**

Chiedi:

> "Vuoi raggruppare questi {N} ticket in un tag?"

- **Se sì:**
  - Cerca tra i tag del ticket originale uno riconoscibile come identificativo cliente (es. `ass_cammini_italia`). Se trovato, proponi nome default `<tag-cliente>-<titolo originale kebab-case>`.
  - Se nessun tag è riconoscibile come cliente, chiedi al dev di indicare manualmente il nome cliente da usare.
  - Il dev può modificare il nome proposto prima della creazione.
  - Crea il tag (stesso endpoint di `wm-skills:wm-tag`):

```bash
ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-https://orchestrator.maphub.it}"
TOKEN=$(jq -r '.token' ~/.config/webmapp/orchestrator-auth.json)
curl -s -X POST "$ORCHESTRATOR_URL/api/tags" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '{"name": "<nome-tag>", "description": "<customer_request originale completo, pre-split>"}'
```

  - Associa il tag restituito a tutti i ticket del gruppo (originale + nuovi) tramite PATCH `tags` su ciascuno (preview + conferma per ciascuna PATCH, come da regola generale).
- **Se no:** salta questo step, nessun tag creato.

4. **Selezione ticket su cui continuare:**

Mostra l'elenco di tutti i ticket del gruppo:

> **Ticket generati dallo split di oc:{ID originale}:**
>
> | Ticket | Titolo |
> |---|---|
> | oc:{ID originale} | {titolo 1} |
> | oc:{nuovo ID 2} | {titolo 2} |
> | ... | |
>
> Su quale vuoi continuare il workflow ora? (oppure "nessuno" per tornare al menu)

- **Se l'utente sceglie un ticket:** prosegui il workflow in `Fase: init-context` usando i dati già noti di quel ticket (senza rifare la GET).
- **Se l'utente risponde "nessuno":** torna al menu A/B/C di `Fase: ticket`.
```

- [ ] **Step 3: Verifica la sintassi Markdown e i riferimenti incrociati**

Rileggi il file intero attorno alla modifica per assicurarti che:
- Il blocco `### ticket: progress` (che segue nel file originale) sia rimasto intatto e ancora coerente come prossimo step dopo `caso-a-split-execution` (il flusso "nessuno split" o "split completato" deve comunque poter arrivare a `ticket: progress` sul ticket scelto).
- Nessun heading duplicato (`###` univoci nel file).

```bash
grep -n "^### " plugins/wm-skills/skills/wm-plan/SKILL.md | sort | uniq -d
```

Expected: nessun output (nessun heading duplicato).

- [ ] **Step 4: Valida il plugin**

```bash
cd /Users/bongiu/Documents/claude-marketplace && claude plugin validate .
```

Expected: nessun errore di schema/validazione.

- [ ] **Step 5: Commit**

```bash
git add plugins/wm-skills/skills/wm-plan/SKILL.md
git commit -m "feat(wm-plan): auto-detect and split multi-request Help desk tickets"
```

---

### Task 2: Bump versione plugin e riga versione in SKILL.md

**Files:**
- Modify: `plugins/wm-skills/.claude-plugin/plugin.json`
- Modify: `plugins/wm-skills/skills/wm-plan/SKILL.md` (riga `**Versione installata:** v<version>` in `### header: versione`)

**Interfaces:**
- Consumes: valore attuale `version` in `plugin.json` (`1.0.0`, verificato in questa sessione)
- Produces: nuovo valore `version` (`1.1.0`) allineato in entrambi i file

- [ ] **Step 1: Leggi la versione attuale**

```bash
jq -r '.version' plugins/wm-skills/.claude-plugin/plugin.json
```

Expected output: `1.0.0`

- [ ] **Step 2: Bump minor a 1.1.0 in `plugin.json`**

Modifica il campo `"version": "1.0.0"` in `"version": "1.1.0"` dentro `plugins/wm-skills/.claude-plugin/plugin.json`.

- [ ] **Step 3: Aggiorna la riga versione in `wm-plan/SKILL.md`**

Cerca nel file la riga (dentro `### header: versione`, sezione descrittiva statica se presente — se la riga non esiste come testo statico ma solo come comando dinamico `jq`, salta questo step: la skill legge `plugin.json` a runtime e non serve altro aggiornamento):

```bash
grep -n "Versione installata" plugins/wm-skills/skills/wm-plan/SKILL.md
```

Se il grep restituisce solo righe di codice bash (`INSTALLED_VERSION=$(jq ...)`), nessuna modifica manuale è necessaria: la versione è già letta dinamicamente da `plugin.json` a runtime — questo step è già soddisfatto dal bump fatto in Step 2.

- [ ] **Step 4: Verifica coerenza**

```bash
jq -r '.version' plugins/wm-skills/.claude-plugin/plugin.json
```

Expected: `1.1.0`

- [ ] **Step 5: Valida il plugin**

```bash
claude plugin validate .
```

Expected: nessun errore.

- [ ] **Step 6: Commit e tag**

```bash
git add plugins/wm-skills/.claude-plugin/plugin.json
git commit -m "chore: bump wm-skills to v1.1.0"
git tag v1.1.0
```

**Nota:** `git push` (incluso `git push origin v1.1.0`) non viene eseguito automaticamente — richiede conferma esplicita dell'utente prima di eseguire un push, per le linee guida generali di sicurezza sulle azioni condivise.

---

### Task 3: Aggiornare `CLAUDE.md` con la nuova feature e rigenerare l'Artifact del diagramma

**Files:**
- Modify: `CLAUDE.md` (sezione `## Feature disponibili`)
- Redeploy: Artifact esistente all'URL già pubblicato in `wm-plan/SKILL.md` → `### header: diagramma` (`https://claude.ai/code/artifact/53f16a0c-0074-44a3-8846-281b0faf5b77`)

**Interfaces:**
- Consumes: riga di tabella esistente in `CLAUDE.md` → `## Feature disponibili` (formato: `| Feature | Ticket | Moduli toccati | Note |`)
- Produces: nessuna nuova interfaccia — solo documentazione e redeploy dell'Artifact esistente

- [ ] **Step 1: Aggiungi la riga della feature in `CLAUDE.md`**

Aggiungi in fondo alla tabella `## Feature disponibili` questa riga:

```markdown
| Split automatico ticket Help desk multi-richiesta in wm-plan | — (nessun ticket) | `plugins/wm-skills/skills/wm-plan/SKILL.md` | Rilevamento automatico di più richieste distinte in ticket Help desk letti in `caso-a`; split in ticket separati (originale ridotto alla prima richiesta + nuovi ticket per le successive), `creator_id` replicato dal cliente originale, tag di raggruppamento opzionale per uso interno dev |
```

- [ ] **Step 2: Verifica la riga aggiunta**

```bash
grep -n "Split automatico ticket Help desk" CLAUDE.md
```

Expected: una riga trovata.

- [ ] **Step 3: Rigenera l'Artifact del diagramma**

Usa lo stesso file HTML sorgente già usato per l'ultimo redeploy dell'Artifact `wm-plan` (verifica il path nella cronologia di sessione o rigenera dal contenuto attuale delle fasi in `wm-plan/SKILL.md` se il sorgente non è più disponibile in scratchpad) e ripubblica con lo **stesso `file_path`** passato allo strumento Artifact, così l'URL resta invariato (`https://claude.ai/code/artifact/53f16a0c-0074-44a3-8846-281b0faf5b77`). Nessun nuovo nodo obbligatorio nel diagramma per questa feature (è una sotto-fase interna a `caso-a`, non una nuova Fase top-level) — il redeploy è comunque dovuto per la regola "qualsiasi modifica al repo" in `CLAUDE.md` → `## Diagramma di flusso wm-plan`.

Se il redeploy fallisce: mostra `⚠️ Impossibile aggiornare l'Artifact del diagramma — potrebbe servire switchare all'account Claude del team Webmapp.` e prosegui comunque (fail-soft, da `CLAUDE.md`).

- [ ] **Step 4: Commit della modifica a `CLAUDE.md`**

```bash
git add CLAUDE.md
git commit -m "docs: register wm-plan Help desk auto-split feature"
```
