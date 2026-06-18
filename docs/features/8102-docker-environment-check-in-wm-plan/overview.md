> Ticket: oc:8102

# Docker environment check in wm-plan

## Cosa cambia
`wm-plan` acquisisce due miglioramenti strutturali:

1. **Slug al posto dei numeri nelle fasi** — tutti gli header `## Fase N — ...` e sottofasi `### Na —` diventano `## Fase: <slug>` e `### fase: sottofase`. Inserire nuove fasi in futuro non richiede rinumerazione.
2. **Nuova `## Fase: environment-setup`** — inserita tra `Fase: ticket` e `Fase: init-context`, raggruppa tutto il rilevamento dell'ambiente di progetto (stack, submodule, UX/UI, Docker) in un unico posto, prima del dialogo con l'utente.

## Perché
- **Slug:** i numeri creano accoppiamento fragile — ogni inserimento richiede rinumerazione di header, riferimenti interni e documenti già prodotti. Gli slug sono stabili per definizione.
- **environment-setup:** logica di rilevamento ambiente era dispersa in Fase 1 (stack UI) e Fase 2a (domain mapping, ux-ui-detection). Centralizzarla in una fase dedicata rende il workflow più leggibile e fa sì che i flag siano disponibili per tutte le fasi successive.
- **docker-check:** i developer Webmapp lavorano su più progetti Laravel che condividono le stesse porte Docker. La fase automatizza lo stop dei container in conflitto e l'avvio di quello corretto.

## Requisiti

### Slug (refactor fasi esistenti)
- [ ] Tutti gli header di fase diventano `## Fase: <slug>` secondo questa mappa:
  - `Fase 0 — Ticket Orchestrator` → `Fase: ticket`
  - `Fase 1 — Leggi CLAUDE.md` → `Fase: init-context`
  - `Fase 2 — Reverse Interaction` → `Fase: reverse-interaction`
  - `Fase 4 — Scrivi overview.md` → `Fase: overview`
  - `Fase 3 — Challenge` → `Fase: challenge`
  - `Fase 5 — Scrivi plan.md` → `Fase: write-plan`
  - `Fase 6 — Esecuzione` → `Fase: execution`
  - `Fase 7 — Mantieni notes.md` → `Fase: notes`
  - `Fase 8 — Aggiorna CLAUDE.md` → `Fase: update-context`
- [ ] Tutti i sotto-header diventano `### fase: sottofase` secondo questa mappa:
  - `### Caso A` → `### ticket: caso-a`
  - `### Domanda progress (Caso A)` → `### ticket: progress`
  - `### Caso B` → `### ticket: caso-b`
  - `### Aggiornamenti espliciti durante il workflow` → `### ticket: aggiornamenti-espliciti`
  - `### Da estrarre in entrambi i casi` → `### ticket: estrazione`
  - `### Rilevazione stack UI (eseguita in Fase 1)` → rimossa (spostata in environment-setup)
  - `### 2a — Mappatura domini e submodule` → rimossa (spostata in environment-setup)
  - `### 2c — UX/UI Detection` → rimossa (spostata in environment-setup)
  - `### 2b — Dialogo` → `### reverse-interaction: dialog`
  - `### Esecuzione tramite subagente isolato` → `### challenge: subagent`
  - `### Dialogo asse per asse` → `### challenge: dialog`
  - `### Aggiornamento overview (se necessario)` → `### challenge: overview-update`
  - `### 6a — Design` → `### execution: design`
  - `### 6b — Creazione branch` → `### execution: branch`
  - `### 6c — Implementazione` → `### execution: implementation`
  - `### 6d — Gate di revisione` → `### execution: review-gate`
  - (nuova) → `### execution: formal-review`
  - `### Aggiornamento Orchestrator` → `### update-context: orchestrator`
- [ ] Tutti i riferimenti interni nel testo aggiornati con i nuovi slug

### Fase: environment-setup (nuova)

**environment-setup: project-detection**
- [ ] Rileva `stack_type` leggendo `.env` (cerca `DOCKER_PROJECT_DIR_NAME`) e `package.json`
  - `laravel` — `.env` con `DOCKER_PROJECT_DIR_NAME`, nessun frontend
  - `frontend` — `package.json` con Vue/Angular/React
  - `fullstack` — entrambi
  - `other` — nessun segnale
- [ ] Imposta `has_docker: true/false` (presenza `DOCKER_PROJECT_DIR_NAME` in `.env`)
- [ ] Imposta `has_submodules: true/false` (output di `git submodule status`)
- [ ] Imposta `stack_ui: angular | vue | react | laravel-blade | false`

**environment-setup: domain-mapping** (contenuto spostato da Fase 2a)
- [ ] Esegue `git submodule status` e legge `.gitmodules`
- [ ] Classifica la feature: custom / package / misto
- [ ] Dichiara esplicitamente il repo di destinazione per ogni file/modulo coinvolto

**environment-setup: ux-ui-detection** (contenuto spostato da Fase 2c)
- [ ] Valuta confidenza UX (Livello 0/1/2/3) basandosi sui flag di project-detection
- [ ] Se confidenza alta → cerca e invoca `ui-ux-pro-max`
- [ ] Se confidenza bassa → chiede conferma all'utente
- [ ] Aggiunge requisiti e rischi UX all'overview con prefisso `[UX]`

**environment-setup: docker-check** (solo se `has_docker: true`)
- [ ] Legge `DOCKER_PROJECT_DIR_NAME` dal `.env`
- [ ] Verifica container attivo con `docker compose -f local.compose.yml ps --format json`
- [ ] Se già up → "✅ Container `<nome>` già attivo", prosegui
- [ ] Se non up → rileva conflitti di porta (`DOCKER_SERVE_PORT`, `DOCKER_PHP_PORT`, `DOCKER_PSQL_PORT`, `DOCKER_VITE_PORT`, `DOCKER_KIBANA_PORT`)
- [ ] Mostra lista container in conflitto, chiede conferma esplicita
- [ ] Dopo conferma: `docker compose stop` (mai `down`/`rm`) → `docker compose -f local.compose.yml up -d`
- [ ] Messaggio finale "✅ Container `<nome>` avviato, ambiente normalizzato"
- [ ] **FAIL-SOFT**: qualsiasi errore → `⚠️` + prosegui

### execution: formal-review (nuova sottofase esplicita)
- [ ] Al termine di `execution: review-gate`, se l'utente vuole una review formale, invocare esplicitamente `wm-skills:wm-review-ticket oc:<ID>`

## Rischi
- **Riferimenti interni**: il testo cita le fasi per nome — ogni riferimento va aggiornato. Un riferimento mancato produce istruzioni incoerenti per Claude.
- **Ordine fasi nel file**: attualmente Fase 4 (overview) appare prima di Fase 3 (challenge) nel file — con gli slug questo non crea confusione ma va documentato nell'ordine corretto.
- **`DOCKER_PROJECT_DIR_NAME` non espanso**: leggere sempre `.env` reale, non `.env-example`.
- **Nome container**: usare `docker compose ps --format json` per trovare i container effettivi, non assumere il nome dalla variabile.
- **Spostamento domain-mapping e ux-ui-detection**: il contenuto va rimosso da Fase 2 e inserito in environment-setup — nessuna duplicazione.

## Out of scope
- File compose diversi da `local.compose.yml`
- Ambienti non locali (CI, staging, production)
- Porte non dichiarate nel `.env`

## Moduli toccati
- `plugins/wm-skills/skills/wm-plan/SKILL.md` — unico file modificato
