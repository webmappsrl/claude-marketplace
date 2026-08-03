> Ticket: oc:8341

# Esecuzione automatica PHPStan pre-PR/merge in wm-plan Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Far eseguire automaticamente PHPStan a `wm-plan` dentro `execution: review-gate`, prima della conferma di commit, quando il repo target è Laravel con PHPStan configurato sia via file di config sia via step CI GitHub Actions — bloccando di default sugli errori introdotti dal diff corrente o su fallimenti infrastrutturali, con un meccanismo di override motivato e tracciato in `notes.md`.

**Architecture:** `plugins/wm-skills/skills/wm-plan/SKILL.md` è un file di istruzioni Markdown eseguito da un agente LLM, non codice applicativo — non esiste un framework di test automatizzato per queste istruzioni. Ogni "task" di questo piano produce una modifica testuale verificabile tramite lettura del contenuto risultante (grep sui pattern attesi) e tramite `claude plugin validate .`, che è il solo tool di verifica automatica disponibile per la struttura dello SKILL.md.

**Tech Stack:** Markdown (istruzioni skill), Bash (snippet di comando dentro le istruzioni), `claude plugin validate`.

## Global Constraints

- Nessun commit o branch automatico durante l'esecuzione — i commit sono decisi ed eseguiti solo dal developer dopo approvazione esplicita (regola generale `wm-plan`).
- Commit message scope: `feat(oc:8341): ...`.
- Unico file da modificare: `plugins/wm-skills/skills/wm-plan/SKILL.md`.
- Il timeout per l'esecuzione di PHPStan è fisso a 5 minuti (300 secondi) — valore approvato in Fase: estimation/challenge, non configurabile in questo ciclo.
- Il nome del servizio Docker da usare per `docker compose exec` è sempre `$DOCKER_PROJECT_DIR_NAME`, già risolto da `environment-setup: docker-check` — nessuna euristica di rilevamento servizio aggiuntiva.
- La detection dello step CI è un grep case-insensitive permissivo sulla keyword `phpstan` — falsi positivi accettati, falsi negativi accettati come limite noto (vedi overview.md → Rischi).

---

### Task 1: Aggiungere detection `has_phpstan_ci` in `environment-setup: project-detection`

**Files:**
- Modify: `plugins/wm-skills/skills/wm-plan/SKILL.md:552-584` (sezione `### environment-setup: project-detection`)

**Interfaces:**
- Consumes: nessuno (task indipendente)
- Produces: flag interno `has_phpstan_ci` (`true`/`false`), consumato da Task 2 in `execution: review-gate`

- [ ] **Step 1: Aggiungere lo snippet di detection al blocco bash esistente**

Nel blocco bash che inizia a riga 556 (`Esegui i seguenti check e imposta i flag interni:`), dopo il commento `# 4. Rileva cartelle frontend Laravel` (riga 567), aggiungi:

```bash
# 5. Rileva PHPStan configurato (file di config + step CI)
ls phpstan.neon phpstan.neon.dist 2>/dev/null
grep -ilr "phpstan" .github/workflows/ 2>/dev/null
```

- [ ] **Step 2: Aggiungere la riga del flag alla tabella "Flag interni da impostare"**

Nella tabella a riga 572-584, dopo la riga `| \`stack_ui\` | \`false\` | nessun segnale |`, aggiungi:

```markdown
| `has_phpstan_ci` | `true` | trovato `phpstan.neon`/`phpstan.neon.dist` nella root **E** keyword `phpstan` (case-insensitive) in almeno un file `.github/workflows/*.yml` |
| `has_phpstan_ci` | `false` | manca almeno una delle due condizioni |
```

- [ ] **Step 3: Verificare il contenuto risultante**

Run: `grep -n "has_phpstan_ci" /Users/bongiu/Documents/claude-marketplace/plugins/wm-skills/skills/wm-plan/SKILL.md`
Expected: 2 righe (la riga `true` e la riga `false` della tabella), più lo snippet bash che contiene `phpstan.neon` e `.github/workflows/`.

- [ ] **Step 4: Validare la skill**

Run: `cd /Users/bongiu/Documents/claude-marketplace && claude plugin validate .`
Expected: nessun errore di validazione (il file resta un Markdown/frontmatter valido).

- [ ] **Step 5: Commit**

```bash
git add plugins/wm-skills/skills/wm-plan/SKILL.md
git commit -m "feat(oc:8341): rileva has_phpstan_ci in environment-setup: project-detection"
```

---

### Task 2: Aggiungere step PHPStan in `execution: review-gate`

**Files:**
- Modify: `plugins/wm-skills/skills/wm-plan/SKILL.md:1033-1085` (sezione `### execution: review-gate`)

**Interfaces:**
- Consumes: flag `has_phpstan_ci` (Task 1), flag `has_docker` e variabile `$DOCKER_PROJECT_DIR_NAME` (già esistenti in `environment-setup: docker-check`)
- Produces: nuova sotto-sezione `#### review-gate: phpstan-check`, richiamata da `review-gate: dialog` prima del punto 3 (attesa conferma commit) e capace di scrivere in `docs/features/<feature-slug>/notes.md` (sezione "Decisioni") e di proporre la creazione di un ticket Orchestrator (stesso flusso di `## Orchestrator API → Creazione ticket`)

- [ ] **Step 1: Inserire la nuova sotto-sezione `review-gate: phpstan-check` subito dopo `review-gate: subagent` (dopo riga 1067, prima di `#### review-gate: dialog` a riga 1069)**

```markdown
#### review-gate: phpstan-check

Eseguito **solo se `has_phpstan_ci: true`**, subito dopo `review-gate: subagent` e prima di `review-gate: dialog`.

**Esecuzione:**

```bash
if [ "$has_docker" = "true" ]; then
  timeout 300 docker compose -f local.compose.yml exec -T "$DOCKER_PROJECT_DIR_NAME" vendor/bin/phpstan analyse --error-format=json
else
  timeout 300 vendor/bin/phpstan analyse --error-format=json
fi
PHPSTAN_EXIT=$?
```

- **`PHPSTAN_EXIT` diverso da 0 e diverso da 1** (comando non trovato, crash, timeout — `timeout` restituisce `124` allo scadere) → **fallimento infrastrutturale**. Tratta questo caso come un blocco (vedi "review-gate: phpstan-override" sotto), con motivazione di default proposta: "PHPStan non è riuscito a completare l'analisi (way exit code $PHPSTAN_EXIT) — verificare ambiente/timeout."
- **`PHPSTAN_EXIT` == 1** (PHPStan ha girato e ha trovato errori) → prosegui al cross-check diff sotto.
- **`PHPSTAN_EXIT` == 0** → nessun errore, nessun blocco, prosegui direttamente a `review-gate: dialog`.

**Cross-check diff vs preesistenti (solo se `PHPSTAN_EXIT == 1`):**

```bash
git diff --name-only > /tmp/wm-plan-diff-files.txt
# Incrocia i file riportati negli errori PHPStan (`.file` nel JSON) con /tmp/wm-plan-diff-files.txt
```

- Per ogni errore riportato da PHPStan, verifica se il file corrispondente è presente in `/tmp/wm-plan-diff-files.txt`.
- **Se almeno un errore è su un file del diff corrente** → blocco (vedi "review-gate: phpstan-override" sotto), motivazione di default: "PHPStan ha trovato N errori sui file modificati in questo task."
- **Se tutti gli errori sono su file fuori dal diff corrente** (debito preesistente) → nessun blocco. Presenta all'utente:

  > "PHPStan ha trovato N errori preesistenti su file non toccati da questa feature ([lista file]). Vuoi che crei un ticket Orchestrator separato per tracciare questo debito tecnico?"

  Se sì, crea il ticket seguendo `## Orchestrator API → Creazione ticket` (preview + conferma, `type: "Task"`, `name` sintetico, `customer_request` con l'elenco degli errori). Poi prosegui a `review-gate: dialog` senza bloccare il commit.

#### review-gate: phpstan-override

Attivato quando `review-gate: phpstan-check` rileva un blocco (errori sul diff corrente o fallimento infrastrutturale).

1. Presenta all'utente gli errori/il fallimento riscontrato.
2. Proponi (o deduci dal contesto) una motivazione per un eventuale bypass — es. "Errore preesistente in un file limitrofo non toccato direttamente" o "Falso positivo noto di PHPStan su questo pattern".
3. Mostra la motivazione proposta in preview e chiedi al dev di confermarla o modificarla.
4. Chiedi conferma esplicita e distinta dalla conferma generica di commit di `review-gate: dialog`, con questo messaggio:

   > "PHPStan blocca il commit per [errori di codice sul diff / fallimento infrastrutturale]. Vuoi bypassare questo blocco specifico? Verrà registrato in notes.md con la motivazione: \"[motivazione confermata]\"."

5. Solo se il dev conferma esplicitamente il bypass (non basta il "procedi" generico del punto 3 di `review-gate: dialog`), consenti di proseguire e registra in `docs/features/<feature-slug>/notes.md` (sezione "Decisioni") una riga con: motivazione, timestamp, e la responsabilità esplicita attribuita al dev.
6. Se il dev non conferma il bypass, il workflow resta bloccato su questo punto: nessun commit finché gli errori non sono risolti o il bypass non viene confermato.
```

- [ ] **Step 2: Aggiornare `review-gate: dialog` per richiamare il check PHPStan prima della richiesta di conferma commit**

Nel punto 2 di `review-gate: dialog` (riga 1071-1078), aggiungi una frase introduttiva prima del messaggio di conferma esistente:

```markdown
1. Presenta all'utente il riepilogo prodotto dal subagente (o dal fallback).
1bis. Se `has_phpstan_ci: true`, esegui `review-gate: phpstan-check` ora, prima di procedere al punto 2. Se ne emerge un blocco, gestiscilo con `review-gate: phpstan-override` prima di continuare.
2. Chiedi conferma esplicita con questo messaggio:
```

(Sostituisci l'attuale numerazione `1.`/`2.` con `1.`/`1bis.`/`2.` come sopra, rinumerando i passi successivi 3-6 di conseguenza in `3.`/`4.`/`5.`/`6.` — restano identici nel contenuto, cambia solo il numero.)

- [ ] **Step 3: Verificare il contenuto risultante**

Run: `grep -n "phpstan-check\|phpstan-override\|PHPSTAN_EXIT" /Users/bongiu/Documents/claude-marketplace/plugins/wm-skills/skills/wm-plan/SKILL.md`
Expected: righe corrispondenti alle due nuove sotto-sezioni e alla variabile `PHPSTAN_EXIT`, tutte presenti nel blocco `execution: review-gate`.

- [ ] **Step 4: Validare la skill**

Run: `cd /Users/bongiu/Documents/claude-marketplace && claude plugin validate .`
Expected: nessun errore di validazione.

- [ ] **Step 5: Commit**

```bash
git add plugins/wm-skills/skills/wm-plan/SKILL.md
git commit -m "feat(oc:8341): esegui PHPStan in execution: review-gate con hard-block e override motivato"
```

---

### Task 3: Aggiornare CLAUDE.md con la nuova feature e bump versione

**Files:**
- Modify: `plugins/wm-skills/.claude-plugin/plugin.json` (bump `version`)
- Modify: `plugins/wm-skills/skills/wm-plan/SKILL.md` (riga `### header: versione` → `**Versione installata:**`)
- Modify: `/Users/bongiu/Documents/claude-marketplace/CLAUDE.md` (sezione `## Feature disponibili`)

**Interfaces:**
- Consumes: nessuno
- Produces: nessuno (task di chiusura/documentazione)

- [ ] **Step 1: Leggere la versione corrente**

Run: `grep '"version"' /Users/bongiu/Documents/claude-marketplace/plugins/wm-skills/.claude-plugin/plugin.json`
Expected: mostra la versione attuale (es. `"1.1.2"`).

- [ ] **Step 2: Bump patch version in `plugin.json`**

Incrementa il numero di patch (es. `1.1.2` → `1.1.3`) nel campo `version` di `plugins/wm-skills/.claude-plugin/plugin.json`.

- [ ] **Step 3: Aggiornare la riga versione in `SKILL.md`**

Nella sezione `### header: versione` di `plugins/wm-skills/skills/wm-plan/SKILL.md`, aggiorna `**Versione installata:** v<vecchia>` con `**Versione installata:** v<nuova>` (stesso valore del bump).

- [ ] **Step 4: Aggiungere la riga in `CLAUDE.md` → `## Feature disponibili`**

Aggiungi in cima alla tabella "Feature disponibili" (senza toccare le righe precedenti):

```markdown
| Esecuzione automatica PHPStan pre-PR/merge in wm-plan | oc:8341 | `plugins/wm-skills/skills/wm-plan/SKILL.md` | `execution: review-gate` esegue PHPStan automaticamente su repo Laravel con PHPStan in CI; hard-block su errori del diff corrente o fallimenti infrastrutturali; override motivato e tracciato in notes.md; errori preesistenti fuori dal diff propongono un ticket Orchestrator dedicato invece di bloccare |
```

- [ ] **Step 5: Verificare il contenuto risultante**

Run: `grep -n "8341" /Users/bongiu/Documents/claude-marketplace/CLAUDE.md /Users/bongiu/Documents/claude-marketplace/plugins/wm-skills/skills/wm-plan/SKILL.md`
Expected: righe corrispondenti in entrambi i file.

- [ ] **Step 6: Validare la skill**

Run: `cd /Users/bongiu/Documents/claude-marketplace && claude plugin validate .`
Expected: nessun errore di validazione.

- [ ] **Step 7: Commit**

```bash
git add plugins/wm-skills/.claude-plugin/plugin.json plugins/wm-skills/skills/wm-plan/SKILL.md CLAUDE.md
git commit -m "chore(oc:8341): bump wm-skills a v<nuova> e documenta feature PHPStan review-gate"
```
