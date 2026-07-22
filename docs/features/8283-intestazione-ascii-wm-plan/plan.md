> Ticket: oc:8283

# Intestazione ASCII per skill wm-plan con info versione/aggiornamenti e diagramma di flusso Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Aggiungere a `wm-plan` un header di sessione (banner ASCII, data ultima modifica, check aggiornamenti, link diagramma di flusso) e l'istruzione permanente in `CLAUDE.md` per mantenere aggiornato il diagramma.

**Architecture:** Questa feature non produce codice eseguibile: `wm-plan` è una skill Markdown, quindi l'"implementazione" consiste in istruzioni testuali che Claude segue a runtime (comandi bash espliciti + logica condizionale in prosa). I "test" sono validazione strutturale (`claude plugin validate .`) e dry-run manuale dei comandi bash per verificare che producano l'output atteso.

**Tech Stack:** Markdown (SKILL.md, CLAUDE.md), bash (git, curl/GitHub API), Claude Code Artifact tool (Mermaid).

## Global Constraints

- Nessun commit o branch automatico in questo piano — ogni commit è un'istruzione testuale per l'utente, mai eseguita autonomamente
- Tutti i commit usano lo scope `feat(oc:8283): ...`
- Nessuna modifica introduce dipendenze/tool esterni oltre a quelli già usati in questo repo (git, curl, jq)
- Le istruzioni aggiunte a `SKILL.md` devono restare fail-soft: nessun errore di rete/git può bloccare l'esecuzione del resto della skill
- Header mostrato solo alla prima invocazione di `wm-plan` in una sessione — la skill deve dichiarare esplicitamente questa regola in prosa (non esiste un meccanismo di stato di sessione nativo da programmare: è un'istruzione comportamentale che Claude segue)

---

### Task 1: Sezione "Header di sessione" in wm-plan/SKILL.md — banner ASCII

**Files:**
- Modify: `plugins/wm-skills/skills/wm-plan/SKILL.md` (inserire nuova sezione subito dopo il frontmatter YAML e prima di `# Webmapp Feature Workflow`)

**Interfaces:**
- Consumes: nessuno (primo task)
- Produces: sezione `## Header di sessione` con sotto-sezione `### header: banner`, punto di aggancio per Task 2 e Task 3 che aggiungeranno sotto-sezioni successive nello stesso blocco

- [ ] **Step 1: Inserire la sezione con il banner ASCII**

Inserisci questo blocco subito dopo la riga `---` di chiusura del frontmatter YAML e prima di `# Webmapp Feature Workflow`:

```markdown
## Header di sessione

**Mostra questa sezione solo alla prima invocazione di `wm-plan` in questa conversazione.** Se `wm-plan` è già stato invocato in precedenza in questa stessa sessione (es. richiamato una seconda volta per un altro ticket), salta l'intero Header di sessione e vai direttamente a `Fase: ticket`.

Alla prima invocazione, mostra questo banner come primissimo output, prima di qualsiasi altro testo:

\`\`\`
 __      __ _____ ___    ______ _        _
 \ \    / /|  \/  || _ \ |  ____| |      | |
  \ \  / / | .  . || |_) || |__  | |  __ _| |_
   \ \/ /  | |\/| ||  __/ |  __| | | / _\` | |__|
    \  /   | |  | || |    | |____| || (_| | |_
     \/    \_|  |_/|_|    |______|_| \__,_|\__|
\`\`\`

Dopo il banner, mostra le informazioni di stato raccolte nelle sotto-sezioni seguenti (`### header: versione`, `### header: diagramma`), poi procedi a `Fase: ticket`.
```

- [ ] **Step 2: Validare la struttura dello SKILL.md**

Run: `claude plugin validate .` (dalla root del repo)
Expected: nessun errore di schema riportato per `wm-skills` (il frontmatter YAML resta invariato, solo aggiunto testo nel corpo)

- [ ] **Step 3: Verifica visiva del banner**

Run: stampa il blocco banner con `cat` in un terminale per controllare l'allineamento monospace:
```bash
cat <<'EOF'
 __      __ _____ ___    ______ _        _
 \ \    / /|  \/  || _ \ |  ____| |      | |
  \ \  / / | .  . || |_) || |__  | |  __ _| |_
   \ \/ /  | |\/| ||  __/ |  __| | | / _` | |__|
    \  /   | |  | || |    | |____| || (_| | |_
     \/    \_|  |_/|_|    |______|_| \__,_|\__|
EOF
```
Expected: le 6 righe restano allineate a colonna fissa, nessun carattere troncato

- [ ] **Step 4: Commit**

```bash
git add plugins/wm-skills/skills/wm-plan/SKILL.md
git commit -m "feat(oc:8283): add session header banner to wm-plan"
```

---

### Task 2: Check versione (hash HEAD vs remoto) e modalità sviluppo locale

**Files:**
- Modify: `plugins/wm-skills/skills/wm-plan/SKILL.md` (aggiungere `### header: versione` dentro `## Header di sessione`, subito dopo il blocco banner del Task 1)

**Interfaces:**
- Consumes: sezione `## Header di sessione` creata in Task 1
- Produces: sotto-sezione `### header: versione` — logica di confronto hash e rilevamento modalità locale, usata come riferimento testuale (nessuna funzione/nome esportato, essendo istruzioni Markdown)

- [ ] **Step 1: Aggiungere la sotto-sezione con la logica di check versione**

Inserisci subito dopo il blocco banner (dentro `## Header di sessione`):

```markdown
### header: versione

Determina la path del repo marketplace installato (dove risiede questo `SKILL.md`) ed esegui:

\`\`\`bash
cd "$(dirname "$(realpath plugins/wm-skills/skills/wm-plan/SKILL.md 2>/dev/null || echo .)")/../../../.." 2>/dev/null
git rev-parse --abbrev-ref HEAD 2>/dev/null
git status --porcelain -- plugins/wm-skills/skills/wm-plan/SKILL.md 2>/dev/null
\`\`\`

**Se il branch non è `main` OPPURE `git status --porcelain` restituisce output non vuoto per `SKILL.md`:**

Mostra:
\`\`\`
🔧 modalità sviluppo locale — check versione saltato
\`\`\`
e salta il resto di questa sotto-sezione (nessun confronto hash remoto).

**Altrimenti (branch `main`, nessuna modifica locale non committata):**

\`\`\`bash
LOCAL_HASH=$(git rev-parse HEAD 2>/dev/null)
REMOTE_HASH=$(curl -s "https://api.github.com/repos/webmappsrl/claude-marketplace/commits/main" | jq -r '.sha' 2>/dev/null)
LOCAL_DATE=$(git log -1 --format=%ad --date=format:%Y-%m-%d -- plugins/wm-skills/skills/wm-plan/SKILL.md 2>/dev/null)
\`\`\`

- Se `LOCAL_HASH` e `REMOTE_HASH` non sono ottenibili (rete assente, comando fallito, output vuoto): mostra `⚠️ Check versione non disponibile.` e prosegui senza bloccare.
- Se `LOCAL_HASH` == `REMOTE_HASH`: mostra `✅ wm-plan aggiornato (ultima modifica: $LOCAL_DATE)`.
- Se `LOCAL_HASH` != `REMOTE_HASH`: mostra `⬆️ Aggiornamento disponibile per wm-plan (ultima modifica locale: $LOCAL_DATE) — esegui \`/plugin marketplace update\` per aggiornare.`
```

- [ ] **Step 2: Validare la struttura dello SKILL.md**

Run: `claude plugin validate .`
Expected: nessun errore di schema

- [ ] **Step 3: Dry-run dei comandi bash su questo stesso repo**

Run (dalla root del repo, sul branch corrente):
```bash
git rev-parse --abbrev-ref HEAD
git status --porcelain -- plugins/wm-skills/skills/wm-plan/SKILL.md
git rev-parse HEAD
curl -s "https://api.github.com/repos/webmappsrl/claude-marketplace/commits/main" | jq -r '.sha'
```
Expected: ogni comando restituisce un valore (branch name, eventuale output di status se ci sono modifiche non committate, hash locale a 40 caratteri, hash remoto a 40 caratteri) senza errori di shell

- [ ] **Step 4: Commit**

```bash
git add plugins/wm-skills/skills/wm-plan/SKILL.md
git commit -m "feat(oc:8283): add version check and local dev mode detection to wm-plan header"
```

---

### Task 3: Link Artifact diagramma nell'header (fail-soft)

**Files:**
- Modify: `plugins/wm-skills/skills/wm-plan/SKILL.md` (aggiungere `### header: diagramma` dentro `## Header di sessione`, dopo `### header: versione`)

**Interfaces:**
- Consumes: sezione `## Header di sessione` (Task 1), sotto-sezione `### header: versione` (Task 2)
- Produces: sotto-sezione `### header: diagramma` — legge la sezione `## Diagramma di flusso wm-plan` che sarà creata in `CLAUDE.md` dal Task 4

- [ ] **Step 1: Aggiungere la sotto-sezione con la logica di lettura del link**

Inserisci subito dopo `### header: versione`:

```markdown
### header: diagramma

Leggi la sezione `## Diagramma di flusso wm-plan` dal `CLAUDE.md` del repo target (root del progetto).

- Se la sezione esiste e contiene un URL: mostra `📊 Diagramma di flusso: <URL>`
- Se la sezione non esiste o non contiene un URL valido (Artifact non ancora pubblicato): mostra `📊 Diagramma di flusso: non ancora pubblicato` — non bloccare l'esecuzione della skill in nessun caso
```

- [ ] **Step 2: Validare la struttura dello SKILL.md**

Run: `claude plugin validate .`
Expected: nessun errore di schema

- [ ] **Step 3: Commit**

```bash
git add plugins/wm-skills/skills/wm-plan/SKILL.md
git commit -m "feat(oc:8283): add artifact diagram link to wm-plan header"
```

---

### Task 4: Sezione CLAUDE.md — URL Artifact e istruzione permanente di rigenerazione

**Files:**
- Modify: `CLAUDE.md` (aggiungere nuova sezione `## Diagramma di flusso wm-plan`, dopo la sezione `## Aggiornare superpowers` e prima di `## Decisioni architetturali`)

**Interfaces:**
- Consumes: nessuno da task precedenti (sezione letta da Task 3, ma scrittura indipendente)
- Produces: sezione `## Diagramma di flusso wm-plan` in `CLAUDE.md`, con placeholder URL che Task 5 aggiornerà con il link reale

- [ ] **Step 1: Aggiungere la sezione al CLAUDE.md**

Inserisci questa sezione in `CLAUDE.md`, subito dopo `## Aggiornare superpowers` (riga 133 dell'attuale file) e prima di `## Decisioni architetturali`:

```markdown
## Diagramma di flusso wm-plan

**URL Artifact:** _non ancora pubblicato_

Questo Artifact mostra un diagramma Mermaid del workflow `wm-plan` (fasi ed esecuzione). L'header di sessione di `wm-plan` (vedi `plugins/wm-skills/skills/wm-plan/SKILL.md` → `## Header di sessione` → `### header: diagramma`) legge l'URL da questa sezione.

**Regola di rigenerazione (obbligatoria per Claude Code):** dopo qualsiasi modifica ai file di questo repo effettuata in una sessione, prima di concludere il lavoro, rigenera l'Artifact del diagramma di flusso usando lo stesso `file_path` e la stessa chiamata di pubblicazione già usata in precedenza (redeploy sullo stesso URL, mai un nuovo Artifact). Questo vale per qualsiasi modifica al repo, non solo per modifiche a `wm-plan`.

**Gestione errori (fail-soft):** la pubblicazione avviene per tentativo diretto, senza check preventivo dell'account attivo. Se il redeploy fallisce, avvisa l'utente con `⚠️ Impossibile aggiornare l'Artifact del diagramma — potrebbe servire switchare all'account Claude del team Webmapp.` e prosegui comunque con il resto della sessione, senza bloccare.

Alla prima pubblicazione riuscita (o ad ogni redeploy con URL diverso, caso che non dovrebbe verificarsi con un redeploy corretto), aggiorna il campo **URL Artifact** qui sopra con il link reale.
```

- [ ] **Step 2: Verifica posizionamento**

Run: `grep -n "^## " CLAUDE.md`
Expected: la nuova riga `## Diagramma di flusso wm-plan` compare tra `## Aggiornare superpowers` e `## Decisioni architetturali` nell'elenco delle sezioni

- [ ] **Step 3: Commit**

```bash
git add CLAUDE.md
git commit -m "feat(oc:8283): add wm-plan flow diagram section and regeneration rule to CLAUDE.md"
```

---

### Task 4b: Estendere la Checklist di completamento di wm-plan con la sincronizzazione dell'Artifact

**Files:**
- Modify: `plugins/wm-skills/skills/wm-plan/SKILL.md` (sezione `## Checklist di completamento`, aggiungere una voce)

**Interfaces:**
- Consumes: sezione `## Diagramma di flusso wm-plan` in `CLAUDE.md` (Task 4)
- Produces: promemoria esplicito nella checklist finale del workflow, punto di aggancio per chi esegue `wm-plan` end-to-end

- [ ] **Step 1: Aggiungere la voce alla checklist**

Nella sezione `## Checklist di completamento` esistente di `plugins/wm-skills/skills/wm-plan/SKILL.md`, aggiungi questa riga all'elenco puntato principale (accanto alle voci `overview.md`/`plan.md`/`notes.md`/`CLAUDE.md`):

```markdown
- [ ] Artifact del diagramma di flusso `wm-plan` rigenerato (redeploy stesso URL) se questa sessione ha modificato file del repo `claude-marketplace` — vedi `CLAUDE.md` → `## Diagramma di flusso wm-plan`
```

- [ ] **Step 2: Validare la struttura dello SKILL.md**

Run: `claude plugin validate .`
Expected: nessun errore di schema

- [ ] **Step 3: Commit**

```bash
git add plugins/wm-skills/skills/wm-plan/SKILL.md
git commit -m "feat(oc:8283): add artifact regeneration check to wm-plan completion checklist"
```

---

### Task 5: Prima pubblicazione dell'Artifact (diagramma Mermaid) e aggiornamento URL in CLAUDE.md

**Files:**
- Create (fuori dal repo, file locale di lavoro): `/tmp` o scratchpad — file `.md` con il diagramma Mermaid da passare all'Artifact tool
- Modify: `CLAUDE.md` (aggiornare il placeholder `_non ancora pubblicato_` nella sezione `## Diagramma di flusso wm-plan` con l'URL reale)

**Interfaces:**
- Consumes: sezione `## Diagramma di flusso wm-plan` in `CLAUDE.md` (Task 4), struttura fasi di `wm-plan/SKILL.md` (per costruire il contenuto del diagramma)
- Produces: URL Artifact pubblicato, salvato in `CLAUDE.md` — consumato a runtime da `### header: diagramma` (Task 3)

- [ ] **Step 1: Chiedere conferma dell'account attivo prima di pubblicare**

Prima di procedere, chiedi esplicitamente all'utente: "Confermi di essere loggato con l'account Claude del team Webmapp prima che pubblichi l'Artifact?" — attendi conferma esplicita (questo è il redeploy fondativo: se sbagliato ora, il link condiviso appartiene all'account sbagliato fin dall'inizio).

- [ ] **Step 2: Scrivere il file Markdown con il diagramma Mermaid**

Crea un file (es. nello scratchpad di sessione) con questo contenuto, che rappresenta le fasi principali di `wm-plan` così come descritte in `plugins/wm-skills/skills/wm-plan/SKILL.md`:

```markdown
# Diagramma di flusso — wm-plan

\`\`\`mermaid
flowchart TD
    A[Fase: ticket] --> B[Fase: environment-setup]
    B --> C[Fase: init-context]
    C --> D[Fase: reverse-interaction]
    D --> E[Fase: overview]
    E --> F[Fase: challenge]
    F --> G{Tipo ticket?}
    G -->|Feature| H[Fase: estimation]
    G -->|Bug/Task| I[Fase: write-plan]
    H --> I
    I --> J[Fase: execution]
    J --> K[execution: branch]
    K --> L[execution: implementation]
    L --> M[execution: review-gate]
    M --> N[Fase: notes]
    N --> O[Fase: update-context]
    O --> P[Checklist di completamento]
\`\`\`
```

- [ ] **Step 3: Pubblicare l'Artifact**

Pubblica il file creato allo Step 2 come Artifact HTML/Markdown (Mermaid è supportato nativamente negli Artifact). Usa un `favicon` stabile (es. 📐) e un titolo chiaro ("wm-plan — Diagramma di flusso").

- [ ] **Step 4: Verificare il risultato della pubblicazione**

- Se la pubblicazione riesce: annota l'URL restituito.
- Se fallisce: mostra `⚠️ Impossibile pubblicare l'Artifact — verifica di essere loggato con l'account Claude del team Webmapp e riprova.` e non modificare `CLAUDE.md` (resta `_non ancora pubblicato_`).

- [ ] **Step 5: Aggiornare CLAUDE.md con l'URL reale**

Solo se lo Step 4 ha avuto successo, sostituisci in `CLAUDE.md`:

```markdown
**URL Artifact:** _non ancora pubblicato_
```

con:

```markdown
**URL Artifact:** <URL restituito dalla pubblicazione>
```

- [ ] **Step 6: Verifica finale**

Run: `grep -n "URL Artifact" CLAUDE.md`
Expected: la riga mostra l'URL reale, non più il placeholder

- [ ] **Step 7: Commit**

```bash
git add CLAUDE.md
git commit -m "feat(oc:8283): publish wm-plan flow diagram artifact and record its URL"
```

---

## Self-Review

- **Copertura spec:** tutti e 9 i requisiti dell'overview sono coperti — banner (Task 1), check versione + modalità locale (Task 2), link diagramma + fail-soft (Task 3), istruzione permanente + storage URL in CLAUDE.md (Task 4), checklist di sincronizzazione (Task 4b), prima pubblicazione reale con gestione errore account (Task 5).
- **Nessun placeholder:** ogni step include il testo Markdown/bash esatto da inserire, nessun "TBD" o "gestisci gli errori" generico — i comportamenti di errore sono scritti esplicitamente in ogni sotto-sezione.
- **Coerenza:** il nome della sezione `## Diagramma di flusso wm-plan` e il campo `**URL Artifact:**` sono usati identicamente in Task 3, Task 4 e Task 5.
