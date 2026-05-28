# wm-plan Challenge Refactor — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Spostare la Challenge dopo overview.md, eseguirla tramite subagente avversariale isolato, e rendere la documentazione obbligatoria anche senza ticket.

**Architecture:** Modifica puramente testuale a `SKILL.md` — nessun codice sorgente coinvolto. Le modifiche cambiano l'ordine delle fasi nel workflow e introducono una nuova sezione per il subagente avversariale.

**Tech Stack:** Markdown, Claude Code plugin system

---

### Task 1: Modifica SKILL.md — ordine fasi, subagente avversariale, doc obbligatoria

**Files:**
- Modify: `plugins/wm-skills/skills/wm-plan/SKILL.md`

> **Nota:** queste modifiche sono già state applicate nella sessione corrente. I passi sotto documentano cosa è stato fatto e permettono verifica.

- [x] **Step 1: Spostare Fase 3 dopo Fase 4**

  Rimuovere il blocco `## Fase 3 — Challenge` dalla posizione originale (tra Fase 2 e Fase 4) e reinserirlo dopo `## Fase 4 — Scrivi overview.md`, prima di `## Fase 5 — Scrivi plan.md`.

- [x] **Step 2: Riscrivere la Fase 3 con subagente avversariale**

  Il nuovo blocco Fase 3 deve contenere:
  - Istruzione di lanciare un subagente con prompt avversariale
  - Prompt esatto da passare al subagente (solo percorso file + istruzioni avversariali)
  - Divieto esplicito di aggiungere contesto dalla conversazione precedente
  - Sezione "Dialogo asse per asse" per gestire il report del subagente
  - Sezione "Aggiornamento overview" se la Challenge trova buchi

  Il prompt del subagente deve essere:
  ```
  Leggi il file `docs/features/<feature-slug>/overview.md`.

  Sei un revisore adversariale. Il tuo unico obiettivo è trovare criticità,
  assunzioni fragili, rischi nascosti e scenari di fallimento in questa feature.
  Non bilanciare con aspetti positivi. Non difendere le scelte fatte.
  Assumi che la soluzione abbia problemi e trovali.

  Analizza questi 5 assi:
  1. Assunzioni fragili — Quali ipotesi potrebbero essere false? Cosa succede se lo sono?
  2. Rischi architetturali — Dove questo design crea accoppiamento, rigidità o debito tecnico?
  3. Blind spot — Cosa non viene considerato? Edge case, utenti atipici, comportamenti inattesi?
  4. Worst case — Se qualcosa va storto in produzione, qual è lo scenario peggiore? È recuperabile?
  5. Difficoltà di rollback — Quanto è facile tornare indietro? Migrazioni, API breaking change, dipendenze esterne?

  Per ogni asse scrivi almeno un punto concreto. Non puoi scrivere "nessun rischio" senza motivazione esplicita.
  ```

- [x] **Step 3: Rendere la documentazione obbligatoria anche senza ticket**

  In `## Fase 4`:
  - Aggiungere nota esplicita: "La documentazione va creata sempre, anche senza ticket."
  - Aggiornare la struttura obbligatoria dell'overview: annotare `> Ticket: oc:<ID>` come "ometti se non c'è ticket"

  In `## Checklist di completamento`:
  - Aggiornare i tre bullet dei file obbligatori: "(riferimento ticket presente se applicabile)"
  - Aggiungere nota in grassetto: "Questi tre file sono obbligatori sempre, con o senza ticket Orchestrator."

- [x] **Step 4: Verificare la skill con il plugin validator**

  ```bash
  claude plugin validate .
  ```

  Expected: nessun errore di schema o frontmatter.

- [ ] **Step 5: Commit**

  ```bash
  git add plugins/wm-skills/skills/wm-plan/SKILL.md
  git commit -m "refactor(wm-plan-challenge-refactor): move Challenge after overview, use adversarial subagent, make docs mandatory"
  ```

---

### Task 2: Creare documentazione feature

**Files:**
- Create: `docs/features/wm-plan-challenge-refactor/overview.md`
- Create: `docs/features/wm-plan-challenge-refactor/plan.md`
- Create: `docs/features/wm-plan-challenge-refactor/notes.md`

- [x] **Step 1: Creare overview.md**

  File creato in sessione. Contenuto: cosa cambia, perché, requisiti, rischi, out of scope, moduli toccati.

- [x] **Step 2: Creare plan.md**

  Questo file.

- [ ] **Step 3: Creare notes.md**

  ```markdown
  # Notes — wm-plan Challenge refactor

  ## Deviazioni dal piano
  Nessuna deviazione rilevante.

  ## Bug trovati
  - La skill originale non esplicitava l'obbligo di creare documentazione in assenza di ticket — buco scoperto durante la sessione di refactor.

  ## Decisioni
  - Il subagente Challenge legge overview.md dal filesystem invece di riceverlo nel prompt: unico modo per garantire isolamento reale del contesto.
  - Prompt avversariale esplicito necessario perché senza framing il modello tende a produrre analisi equilibrate invece che critiche.

  ## Follow-up
  Nessuno.
  ```

- [ ] **Step 4: Commit**

  ```bash
  git add docs/features/wm-plan-challenge-refactor/
  git commit -m "docs(wm-plan-challenge-refactor): add feature documentation"
  ```
