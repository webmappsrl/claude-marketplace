> Ticket: oc:8102

# Docker environment check in wm-plan — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Migrare tutte le fasi di `wm-plan` a slug inglesi stabili, spostare domain-mapping e ux-ui-detection in una nuova `Fase: environment-setup`, e aggiungere docker-check e formal-review.

**Architecture:** Modifica puramente testuale di `plugins/wm-skills/skills/wm-plan/SKILL.md`. Quattro operazioni in sequenza: (1) rinomina header, (2) aggiorna riferimenti interni, (3) sposta contenuto domain-mapping e ux-ui-detection in environment-setup, (4) inserisce nuove sottofasi docker-check e formal-review.

**Tech Stack:** Markdown

## Global Constraints

- MAI usare `docker compose down` o `docker rm` — solo `docker compose stop`
- La fase environment-setup è FAIL-SOFT: qualsiasi errore → `⚠️` + prosegui
- Commit convention: `feat(oc:8102): ...`
- NO commit automatici

---

### Task 1: Rinominare tutti gli header delle fasi e sottofasi

**Files:**
- Modify: `plugins/wm-skills/skills/wm-plan/SKILL.md`

- [ ] **Step 1: Rinomina header fasi principali**

Eseguire queste sostituzioni con Edit (una alla volta):

| Vecchio | Nuovo |
|---------|-------|
| `## Fase 0 — Ticket Orchestrator` | `## Fase: ticket` |
| `## Fase 1 — Leggi CLAUDE.md` | `## Fase: init-context` |
| `## Fase 2 — Reverse Interaction (obbligatoria, non skippabile)` | `## Fase: reverse-interaction (obbligatoria, non skippabile)` |
| `## Fase 4 — Scrivi \`overview.md\`` | `## Fase: overview` |
| `## Fase 3 — Challenge (review adversariale)` | `## Fase: challenge` |
| `## Fase 5 — Scrivi \`plan.md\`` | `## Fase: write-plan` |
| `## Fase 6 — Esecuzione` | `## Fase: execution` |
| `## Fase 7 — Mantieni \`notes.md\`` | `## Fase: notes` |
| `## Fase 8 — Aggiorna CLAUDE.md del progetto target` | `## Fase: update-context` |

- [ ] **Step 2: Rinomina sotto-header**

| Vecchio | Nuovo |
|---------|-------|
| `### Caso A — L'utente fornisce un ID ticket` | `### ticket: caso-a` |
| `### Domanda progress (Caso A)` | `### ticket: progress` |
| `### Caso B — L'utente non ha un ticket` | `### ticket: caso-b` |
| `### Aggiornamenti espliciti durante il workflow` | `### ticket: aggiornamenti-espliciti` |
| `### Da estrarre in entrambi i casi` | `### ticket: estrazione` |
| `### 2b — Dialogo` | `### reverse-interaction: dialog` |
| `### Esecuzione tramite subagente isolato` | `### challenge: subagent` |
| `### Dialogo asse per asse` | `### challenge: dialog` |
| `### Aggiornamento overview (se necessario)` | `### challenge: overview-update` |
| `### 6a — Design (se applicabile)` | `### execution: design (se applicabile)` |
| `### 6b — Creazione branch (obbligatoria, prima di scrivere qualsiasi file)` | `### execution: branch (obbligatoria, prima di scrivere qualsiasi file)` |
| `### 6c — Implementazione` | `### execution: implementation` |
| `### 6d — Gate di revisione (obbligatorio, non skippabile)` | `### execution: review-gate (obbligatorio, non skippabile)` |
| `### Aggiornamento Orchestrator (solo se esiste un ticket oc:\<ID\>)` | `### update-context: orchestrator (solo se esiste un ticket oc:\<ID\>)` |

- [ ] **Step 3: Verifica header**

```bash
grep -n "^## Fase\|^### " plugins/wm-skills/skills/wm-plan/SKILL.md | grep -v "^[0-9]*:### environment-setup\|^[0-9]*:### ticket\|^[0-9]*:### reverse\|^[0-9]*:### challenge\|^[0-9]*:### execution\|^[0-9]*:### update-context\|^[0-9]*:### write-plan\|^[0-9]*:### Orchestrator API\|^[0-9]*:### Configurazione\|^[0-9]*:### Login\|^[0-9]*:### Migrazione\|^[0-9]*:### Lettura\|^[0-9]*:### Creazione\|^[0-9]*:### Aggiornamento ticket\|^[0-9]*:### Status\|^[0-9]*:### Campi\|^[0-9]*:### Regola"
```

Output atteso: nessuna riga (tutti gli header sono stati rinominati).

---

### Task 2: Aggiornare i riferimenti interni nel testo

**Files:**
- Modify: `plugins/wm-skills/skills/wm-plan/SKILL.md`

- [ ] **Step 1: Aggiorna riferimenti a Fase 0/ticket**

| Vecchio | Nuovo |
|---------|-------|
| `usato in Fase 2 e Fase 4` | `usato in Fase: reverse-interaction e Fase: overview` |
| `può orientare le domande in Fase 2` | `può orientare le domande in Fase: reverse-interaction` |
| `prima di procedere alla Fase 1.` | `prima di procedere alla Fase: init-context.` |
| `procedi direttamente alla Fase 1 senza modificare il ticket.` | `procedi direttamente alla Fase: init-context senza modificare il ticket.` |
| `**Il ticket va creato prima di procedere alla Fase 1.**` | `**Il ticket va creato prima di procedere alla Fase: init-context.**` |

- [ ] **Step 2: Aggiorna riferimenti a Fase 1/init-context**

| Vecchio | Nuovo |
|---------|-------|
| `Questo flag rimane attivo per tutto il workflow e influenza Fase 2a.` | `Questo flag rimane attivo per tutto il workflow e influenza environment-setup: ux-ui-detection.` |
| `**Livello 1 — Stack UI rilevato in Fase 1 + file coinvolti**` | `**Livello 1 — Stack UI rilevato in Fase: init-context + file coinvolti**` |

- [ ] **Step 3: Aggiorna riferimenti a Fase 2/reverse-interaction**

| Vecchio | Nuovo |
|---------|-------|
| `Procedi alla Fase 3 solo dopo aver fatto almeno 5 domande` | `Procedi alla Fase: overview solo dopo aver fatto almeno 5 domande` |

- [ ] **Step 4: Aggiorna riferimenti a Fase 3/challenge e Fase 4/overview**

| Vecchio | Nuovo |
|---------|-------|
| `[Motivazione business/tecnica emersa dal ticket e dalla Fase 2]` | `[Motivazione business/tecnica emersa dal ticket e dalla Fase: reverse-interaction]` |
| `[Criticità emerse dalla Fase 3 con indicazione di come vengono mitigate]` | `[Criticità emerse dalla Fase: challenge con indicazione di come vengono mitigate]` |
| `aggiorna \`overview.md\` prima di procedere alla Fase 5.` | `aggiorna \`overview.md\` prima di procedere alla Fase: write-plan.` |

- [ ] **Step 5: Aggiorna riferimenti a Fase 5/write-plan**

| Vecchio | Nuovo |
|---------|-------|
| `generati in Fase 4 (uno per ogni repo coinvolto)` | `generati in Fase: overview (uno per ogni repo coinvolto)` |
| `Rischi e decisioni emerse dalla Challenge (Fase 3) e recepite nell'overview` | `Rischi e decisioni emerse dalla Fase: challenge e recepite nell'overview` |
| `Vincoli tecnici emersi dal dialogo in Fase 2` | `Vincoli tecnici emersi dalla Fase: reverse-interaction` |

- [ ] **Step 6: Aggiorna riferimenti a Fase 6/execution**

| Vecchio | Nuovo |
|---------|-------|
| `rilevata in Fase 2 — solitamente inglese` | `rilevata in Fase: reverse-interaction — solitamente inglese` |
| `file di lingua rilevati in Fase 2` | `file di lingua rilevati in Fase: reverse-interaction` |
| `Se in Fase 2 non è stata rilevata la configurazione i18n` | `Se in Fase: reverse-interaction non è stata rilevata la configurazione i18n` |
| `completa la Fase 7 (scrivi \`notes.md\`) e la Fase 8 (aggiorna \`CLAUDE.md\`)` | `completa la Fase: notes e la Fase: update-context` |

- [ ] **Step 7: Aggiorna sezione Composizione con altre skill**

| Vecchio | Nuovo |
|---------|-------|
| `invocata automaticamente in Fase 2c quando rilevati componenti UI/UX` | `invocata automaticamente in environment-setup: ux-ui-detection quando rilevati componenti UI/UX` |
| `applica in Fase 5 (scrittura plan) e Fase 6 (esecuzione)` | `applica in Fase: write-plan e Fase: execution` |
| `applica dopo la Fase 7, prima di aprire la PR` | `applica dopo la Fase: notes, prima di aprire la PR` |

- [ ] **Step 8: Aggiorna sezione update-context**

| Vecchio | Nuovo |
|---------|-------|
| `le scelte non ovvie emerse dalla Challenge (Fase 3) e dai notes (Fase 7)` | `le scelte non ovvie emerse dalla Fase: challenge e dalla Fase: notes` |

- [ ] **Step 9: Verifica finale riferimenti numerici**

```bash
grep -n "Fase [0-9]" plugins/wm-skills/skills/wm-plan/SKILL.md
```

Output atteso: nessuna riga.

---

### Task 3: Spostare domain-mapping e ux-ui-detection, inserire environment-setup

**Files:**
- Modify: `plugins/wm-skills/skills/wm-plan/SKILL.md`

- [ ] **Step 1: Inserire la nuova sezione Fase: environment-setup**

Trovare il `---` che chiude `## Fase: ticket` (prima di `## Fase: init-context`) e inserire la nuova sezione:

Sostituire:
```markdown
---

## Fase: init-context
```

Con:
```markdown
---

## Fase: environment-setup

Questa fase rileva il tipo di progetto e normalizza l'ambiente prima di qualsiasi altra operazione. I flag impostati qui rimangono attivi per tutto il workflow.

**⚠️ Questa fase è FAIL-SOFT.** Qualsiasi errore (Docker daemon non attivo, CLI mancante, `.env` non leggibile) produce `⚠️ Environment setup non disponibile — proseguo con il workflow.` e il workflow continua normalmente. La fase non blocca mai il flusso.

### environment-setup: project-detection

Esegui i seguenti check e imposta i flag interni:

```bash
# 1. Rileva Laravel/wmpackage con Docker
grep -s "DOCKER_PROJECT_DIR_NAME" .env

# 2. Rileva stack frontend
cat package.json 2>/dev/null | jq -r '(.dependencies // {}) + (.devDependencies // {}) | keys[]' | grep -E "^(@angular/core|vue|@vue/core|react)$"

# 3. Rileva submodule
git submodule status 2>/dev/null

# 4. Rileva cartelle frontend Laravel
ls resources/views/ resources/js/components/ 2>/dev/null | head -3
```

**Flag interni da impostare:**

| Flag | Valore | Condizione |
|------|--------|-----------|
| `stack_type` | `laravel` | `.env` con `DOCKER_PROJECT_DIR_NAME`, nessun frontend JS |
| `stack_type` | `frontend` | `package.json` con Vue/Angular/React |
| `stack_type` | `fullstack` | entrambi i segnali |
| `stack_type` | `other` | nessun segnale |
| `has_docker` | `true` / `false` | presenza `DOCKER_PROJECT_DIR_NAME` in `.env` |
| `has_submodules` | `true` / `false` | output non vuoto di `git submodule status` |
| `stack_ui` | `angular` | trovato `@angular/core` |
| `stack_ui` | `vue` | trovato `vue` o `@vue/core` |
| `stack_ui` | `react` | trovato `react` |
| `stack_ui` | `laravel-blade` | `stack_type: laravel` + trovata `resources/views/` |
| `stack_ui` | `false` | nessun segnale |

### environment-setup: domain-mapping

Prima di fare qualsiasi domanda, ispeziona il progetto e identifica:

1. **Submodule presenti** — esegui `git submodule status` e leggi `.gitmodules`. Nota quali repo sono inclusi (es. `wm-package` per backend, `wm-core` / `map-core` per frontend).
2. **Dominio della feature** — sulla base del ticket e del codice, classifica la feature:
   - **Custom** — logica specifica di questo progetto, il codice va nel repo principale
   - **Package/submodule** — logica generica riusabile, il codice va nel submodule appropriato
   - **Misto** — parte nel repo principale, parte nel submodule (specifica quale parte va dove)
3. **Dichiarazione esplicita** — scrivi un messaggio con:
   - Submodule trovati e loro scopo
   - Classificazione della feature (custom / package / misto)
   - Per ogni file o modulo che verrà toccato, il repo di destinazione esplicito

Questa classificazione rimane attiva per tutto il workflow: overview.md, plan.md e ogni step del piano devono sempre indicare il repo di destinazione per ogni file.

### environment-setup: ux-ui-detection

Dopo aver impostato i flag in project-detection, valuta la necessità di design UX:

**Livello 0 — Richiesta esplicita (priorità massima)**

Se la richiesta dell'utente o il titolo/body del ticket contiene parole come: `UI`, `UX`, `interfaccia`, `componente`, `layout`, `form`, `modal`, `stile`, `CSS`, `design`, `animazione`, `schermata` → confidenza alta, procedi direttamente all'invocazione.

**Livello 1 — Stack UI rilevato in Fase: init-context + file coinvolti**

Se `stack_ui != false` E almeno uno dei file/moduli toccati dalla feature appartiene a:
- Estensioni: `.vue`, `.html`, `.css`, `.scss`, `.component.ts`, `.component.html`, `.component.scss`
- Cartelle: `src/components/`, `src/views/`, `resources/views/`, `resources/js/components/`, `src/app/`

→ confidenza alta, procedi all'invocazione automatica.

**Livello 2 — Solo stack UI, nessun file frontend esplicito**

Se `stack_ui != false` ma i file toccati sono solo backend/logica → confidenza bassa.

**Livello 3 — Nessun segnale**

Se `stack_ui: false` e nessun file frontend → nessuna azione UX.

---

**Flusso di invocazione:**

```
SE confidenza alta:
  → scrivi: "Rilevati componenti UI ([file/stack trovati]) — cerco skill UX specializzata."
  → cerca tra le skill disponibili: ui-ux-pro-max
  → SE trovata:
      invoca la skill con questo contesto:
      - Titolo ticket e tipo
      - Stack rilevato (vue / angular / react / laravel-blade)
      - Lista file/componenti UI coinvolti
      - customer_request del ticket (se disponibile)
      Ricevuto il parere della skill:
      - Le raccomandazioni UX → aggiunte come requisiti funzionali nella sezione "Requisiti" di overview.md con prefisso `[UX]`
      - I rischi UX → aggiunti nella sezione "Rischi" di overview.md con prefisso `[UX]`
  → SE non trovata:
      scrivi: "⚠️ La skill ui-ux-pro-max non è installata.
      Per ottenerla: /plugin install ui-ux-pro-max@wm-marketplace
      Procedo con giudizio interno UX."

SE confidenza bassa:
  → chiedi: "Ho rilevato uno stack frontend ([stack]) ma i file toccati sembrano principalmente backend.
     Questa feature ha componenti UI? Vuoi che invochi la skill UX specializzata?"
  → SE sì: stesso flusso lookup sopra
  → SE no: procedi senza UX detection
```

### environment-setup: docker-check

Eseguito **solo se `has_docker: true`**.

**Verifica container attivo:**

```bash
DOCKER_PROJECT_DIR_NAME=$(grep "^DOCKER_PROJECT_DIR_NAME=" .env | cut -d= -f2)
docker compose -f local.compose.yml ps --format json 2>/dev/null
```

Analizza l'output per trovare container con `State: running` associati al progetto corrente.

- **Se già up:** mostra `✅ Container \`<nome>\` già attivo.` e prosegui alla Fase: init-context.
- **Se non up:** procedi al rilevamento conflitti.

**Rilevamento conflitti di porta:**

```bash
# Leggi le porte dal .env (ignora variabili mancanti)
grep -E "^DOCKER_(SERVE|PHP|PSQL|VITE|KIBANA)_PORT=" .env 2>/dev/null

# Confronta con i container running
docker ps --format "table {{.Names}}\t{{.Ports}}" 2>/dev/null
```

**Se ci sono container in conflitto**, mostra:

> **Container in conflitto rilevati:**
>
> | Container | Porta in conflitto |
> |-----------|-------------------|
> | `<nome1>` | `<porta>` |
>
> Fermo questi container e avvio `<DOCKER_PROJECT_DIR_NAME>`?

Attendi conferma esplicita. Solo dopo conferma:

```bash
# Ferma i container in conflitto — MAI docker compose down o docker rm
docker stop <container-id>   # oppure: docker compose stop nella dir del progetto in conflitto

# Avvia il container del progetto corrente
echo "⏳ Avvio container \`$DOCKER_PROJECT_DIR_NAME\`... potrebbe richiedere qualche minuto."
docker compose -f local.compose.yml up -d
```

Al termine: `✅ Container \`<DOCKER_PROJECT_DIR_NAME>\` avviato, ambiente normalizzato.`

---

## Fase: init-context
```

- [ ] **Step 2: Rimuovere domain-mapping e ux-ui-detection da Fase: reverse-interaction**

Nella sezione `## Fase: reverse-interaction`, rimuovere completamente:
- Il blocco `### reverse-interaction: domain-mapping` (ex `### 2a`) con tutto il suo contenuto
- Il blocco `### reverse-interaction: ux-ui-detection` (ex `### 2c`) con tutto il suo contenuto

Aggiungere dopo l'header `## Fase: reverse-interaction (obbligatoria, non skippabile)` una riga di riferimento:

```markdown
> I flag `stack_ui`, `stack_type`, `has_docker`, `has_submodules`, la classificazione domain-mapping e la valutazione UX sono già stati impostati in Fase: environment-setup.
```

- [ ] **Step 3: Rimuovere stack-ui-detection da Fase: init-context**

Nella sezione `## Fase: init-context`, rimuovere completamente il blocco `### init-context: stack-ui-detection` con tutto il suo contenuto.

Aggiungere in fondo alla sezione:

```markdown
> I flag di stack (`stack_ui`, `stack_type`) sono già stati impostati in Fase: environment-setup.
```

---

### Task 4: Aggiungere execution: formal-review

**Files:**
- Modify: `plugins/wm-skills/skills/wm-plan/SKILL.md`

- [ ] **Step 1: Aggiungere la sottofase formal-review**

Nella sezione `### execution: review-gate`, trovare il punto 3 che già menziona la review formale:

```markdown
> 💡 **Review formale opzionale:** vuoi eseguire una code review strutturata prima dei commit? Invoca `wm-skills:wm-review-ticket oc:<ID>` per finder paralleli e aggiornamento automatico del ticket. Rispondi **sì** per eseguirla ora, **no** per procedere direttamente ai commit.
```

Subito dopo, aggiungere la nuova sottofase:

```markdown
### execution: formal-review

Se l'utente risponde **sì** alla proposta di review formale in `execution: review-gate`:

1. Invoca `wm-skills:wm-review-ticket oc:<ID>`
2. Attendi il completamento della review
3. Se emergono correzioni → applicale prima di procedere ai commit
4. Torna al punto 6 di `execution: review-gate` (esegui i commit)
```

---

### Task 5: Validare e committare

- [ ] **Step 1: Validare la skill**

```bash
claude plugin validate .
```

Output atteso: nessun errore.

- [ ] **Step 2: Verifica struttura slug**

```bash
grep -n "^## Fase" plugins/wm-skills/skills/wm-plan/SKILL.md
```

Output atteso (in ordine):
```
## Fase: ticket
## Fase: environment-setup
## Fase: init-context
## Fase: reverse-interaction (obbligatoria, non skippabile)
## Fase: overview
## Fase: challenge
## Fase: write-plan
## Fase: execution
## Fase: notes
## Fase: update-context
```

- [ ] **Step 3: Verifica nessun riferimento numerico rimasto**

```bash
grep -n "Fase [0-9]" plugins/wm-skills/skills/wm-plan/SKILL.md
```

Output atteso: nessuna riga.

- [ ] **Step 4: Commit (solo dopo approvazione esplicita del developer)**

```bash
git add plugins/wm-skills/skills/wm-plan/SKILL.md \
        docs/features/8102-docker-environment-check-in-wm-plan/
git commit -m "feat(oc:8102): migrate wm-plan phases to slugs and add environment-setup phase"
```
