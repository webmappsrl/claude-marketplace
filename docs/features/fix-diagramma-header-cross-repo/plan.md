# Fix diagramma header cross-repo Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Far sì che `### header: diagramma` in `wm-plan/SKILL.md` legga sempre l'URL Artifact dal `CLAUDE.md` remoto di `claude-marketplace` via fetch diretto, invece che dal `CLAUDE.md` del repo target — così l'header mostra il link corretto anche quando `wm-plan` viene invocato da un repo diverso (es. `camminiditalia`).

**Architecture:** Sostituzione della singola sotto-sezione `### header: diagramma` in `plugins/wm-skills/skills/wm-plan/SKILL.md`: da "leggi il CLAUDE.md del repo target" a "fetch remoto di `https://raw.githubusercontent.com/webmappsrl/claude-marketplace/main/CLAUDE.md`, estrai sezione, gestisci 3 esiti (successo / fetch fallito / sezione assente)".

**Tech Stack:** Nessuno — modifica di sole istruzioni Markdown consumate da Claude Code (nessun codice eseguibile, nessuna suite di test automatizzata). Verifica tramite esecuzione manuale della skill.

## Global Constraints

- Il fetch remoto va fatto **sempre**, anche quando `wm-plan` è invocato dentro il repo `claude-marketplace` stesso — nessun fallback su lettura locale (decisione presa in `Fase: challenge`, scelta di semplicità).
- Distinguere esplicitamente **fetch HTTP fallito** (rete assente, timeout, status non-2xx) da **fetch riuscito ma sezione/URL assente o non valida** — messaggi diversi per i due casi.
- Nessun commit automatico: i comandi `git commit` nel piano sono istruzioni testuali per l'utente, non azioni da eseguire in autonomia.
- Nessun ticket Orchestrator — scope commit: nessun prefisso `oc:<ID>`, usa solo un messaggio descrittivo.

---

### Task 1: Sostituire `### header: diagramma` con fetch remoto

**Files:**
- Modify: `plugins/wm-skills/skills/wm-plan/SKILL.md:52-57`

**Interfaces:**
- Consumes: nessuna dipendenza da altre sezioni del piano (task unico).
- Produces: comportamento osservabile dell'header di `wm-plan` alla prima invocazione di sessione — tre possibili righe di output:
  - `📊 Diagramma di flusso: <URL>`
  - `📊 Diagramma di flusso: non ancora pubblicato`
  - `⚠️ Diagramma di flusso non verificabile`

- [ ] **Step 1: Leggere il blocco attuale da sostituire**

Il blocco corrente (righe 52-57 di `plugins/wm-skills/skills/wm-plan/SKILL.md`) è:

```markdown
### header: diagramma

Leggi la sezione `## Diagramma di flusso wm-plan` dal `CLAUDE.md` del repo target (root del progetto).

- Se la sezione esiste e contiene un URL: mostra `📊 Diagramma di flusso: <URL>`
- Se la sezione non esiste o non contiene un URL valido (Artifact non ancora pubblicato): mostra `📊 Diagramma di flusso: non ancora pubblicato` — non bloccare l'esecuzione della skill in nessun caso
```

- [ ] **Step 2: Sostituire il blocco con la nuova versione**

Usa l'Edit tool per rimpiazzare esattamente il testo dello Step 1 con:

```markdown
### header: diagramma

Esegui il fetch del `CLAUDE.md` di `claude-marketplace` direttamente da GitHub (sempre da remoto, anche se `wm-plan` è invocato all'interno del repo `claude-marketplace` stesso — nessuna lettura da filesystem locale):

```bash
curl -s --max-time 5 "https://raw.githubusercontent.com/webmappsrl/claude-marketplace/main/CLAUDE.md"
```

Dal contenuto restituito, cerca la sezione `## Diagramma di flusso wm-plan` ed estrai l'URL indicato dopo `**URL Artifact:**`.

- **Se il comando `curl` fallisce** (nessun output, errore di rete, timeout, o risposta HTTP non 2xx): mostra `⚠️ Diagramma di flusso non verificabile` — non bloccare l'esecuzione della skill in nessun caso.
- **Se il fetch riesce ma la sezione non esiste o non contiene un URL valido** (Artifact non ancora pubblicato): mostra `📊 Diagramma di flusso: non ancora pubblicato`.
- **Se il fetch riesce e la sezione contiene un URL valido:** mostra `📊 Diagramma di flusso: <URL>`.
```

- [ ] **Step 3: Verifica manuale — caso successo**

Esegui:
```bash
curl -s --max-time 5 "https://raw.githubusercontent.com/webmappsrl/claude-marketplace/main/CLAUDE.md" | grep -A 2 "## Diagramma di flusso wm-plan"
```
Expected: output contiene la riga `**URL Artifact:** https://claude.ai/code/artifact/53f16a0c-0074-44a3-8846-281b0faf5b77` (o l'URL corrente).

- [ ] **Step 4: Verifica manuale — caso fetch fallito**

Esegui lo stesso comando con un host non risolvibile per simulare il fallimento:
```bash
curl -s --max-time 5 "https://raw.githubusercontent.com/webmappsrl/repo-inesistente-xyz/main/CLAUDE.md"
```
Expected: output vuoto o errore — conferma che la condizione "fetch fallito" del nuovo testo della skill (output vuoto/non 2xx) è raggiungibile e distinguibile dal caso "sezione assente" (che richiederebbe invece un fetch riuscito su un file senza quella sezione).

- [ ] **Step 5: Verifica end-to-end — invocare wm-plan da un repo diverso**

In un repo diverso da `claude-marketplace` (es. `camminiditalia`, quello riportato dall'utente come bug originale), invoca `/wm-skills:wm-plan` e osserva l'header.
Expected: la riga `📊 Diagramma di flusso: ...` mostra l'URL reale (`https://claude.ai/code/artifact/53f16a0c-0074-44a3-8846-281b0faf5b77`), non più `non ancora pubblicato`.

- [ ] **Step 6: Commit**

```bash
git add plugins/wm-skills/skills/wm-plan/SKILL.md
git commit -m "fix: leggi link diagramma wm-plan sempre da claude-marketplace remoto"
```
