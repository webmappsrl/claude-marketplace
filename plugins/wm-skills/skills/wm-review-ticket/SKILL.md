---
name: wm-review-ticket
description: Esegui una code review completa di un ticket Orchestrator. Usa quando un collega ti assegna un ticket/PR da rivedere, oppure al termine di una feature wm-plan prima del merge. Input: oc:<ID>.
---

## Contratto artefatti

Questa skill consuma gli artefatti prodotti da `wm-skills:wm-plan`. Per conoscere la struttura autoritativa di `docs/features/<slug>/`, esegui WebFetch su:

```
https://raw.githubusercontent.com/webmappsrl/claude-marketplace/main/plugins/wm-skills/skills/wm-plan/SKILL.md
```

Cerca la sezione `## Contratto artefatti` nel file scaricato. Funziona anche se `wm-plan` non è installato localmente.

---

## Delega ad agenti

Il meccanismo è in `${CLAUDE_PLUGIN_ROOT}/shared/agent-delegation.md`. Dove questa skill deve
rispondere a una domanda sul codice per valutare una modifica, usa `wm-codebase-research`
invece di leggere i file nel context principale, e verifica ogni prova ricevuta con
`sed -n '<inizio>,<fine>p' <file>` prima di usarne il contenuto.

Non delegare il **giudizio** sulla review: la delega copre la raccolta delle informazioni,
non la valutazione.

Se il dev contesta una conclusione di `wm-codebase-research`, vedi `## Revisione con l'agente` in `${CLAUDE_PLUGIN_ROOT}/shared/agent-delegation.md`.

---

## Orchestrator API

Le operazioni su Orchestrator si fanno con i tool del server `orchestrator`: vedi `wm-skills:wm-plan` → `## Orchestrator`. La regola dell'anteprima prima della scrittura vale identica qui.

---

## Fase 1 — Lettura ticket

Chiama `get_story` con `story_id: <ID>`.

Se il tool segnala credenziali assenti o scadute, guida il dev a rifare l'accesso (vedi `wm-skills:wm-plan` → `## Orchestrator → Credenziali`).

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

Estrai dall'URL PR il nome del repo (es. `webmappsrl/maphub`) e cerca il path locale corrispondente nel CLAUDE.md.

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

Estrai il nome del branch dall'URL PR (tramite GitHub API se necessario) o dalla description del ticket.

```bash
# Tentativo 1: branch locale
git checkout <branch-name> 2>/dev/null && echo "checkout locale OK" || \

# Tentativo 2: fetch dal remote
git fetch origin <branch-name> 2>/dev/null && git checkout <branch-name> 2>/dev/null && echo "fetch+checkout OK" || \

# Tentativo 3: cerca commit di merge
git log --all --oneline --grep="<branch-name>" | head -5
```

Se tutti i tentativi falliscono:
> "Non riesco a trovare il branch `<branch-name>`. Puoi fornire l'hash del commit di merge o il nome corretto del branch?"

---

## Fase 4 — Lettura contesto wm-plan

Delega questa fase a `wm-codebase-research` (delega informata: passa il path del repo e un
elenco esplicito di domande, non un generico "leggi i docs"): input piccolo, output piccolo,
il lavoro di lettura di `overview.md`, `plan.md` e `notes.md` è il lavoro intermedio da
spostare fuori dal context principale. L'agente risponde nel formato `Domanda / Risposta /
Fonte / Estratto`, quindi le domande vanno poste una per una, per esempio:

- Cosa prevede `overview.md` in "Cosa cambia" e "Requisiti" per la feature `<slug>`?
- Cosa è dichiarato "Out of scope" in `overview.md`?
- Quali task elenca `plan.md` e risultano tutti completati?
- `notes.md` registra deviazioni dal piano? Quali, e con quale motivazione?

Verifica ogni prova ricevuta con `sed -n '<inizio>,<fine>p' <file>` prima di usarne il
contenuto — se non combacia, la ricerca si rifà nel context principale. Se l'agente risponde
`RICERCA FALLITA: <motivo>` o `Risposta: non determinabile dal repo` su una domanda chiave,
rifai quella parte di ricerca nel context principale.

Cerca i docs della feature nel repo:

```bash
# Ricerca fuzzy per ID ticket
find docs/features/ -maxdepth 1 -type d 2>/dev/null | grep "<ID>"
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
# Se hai il branch in checkout:
git diff main...<branch-name> --stat
git diff main...<branch-name>

# Se hai solo un commit di merge:
git show <hash> --stat
git show <hash>
```

### 5b — Finder paralleli

Lancia **5 finder paralleli** sul diff. Ogni finder legge il diff con `git diff` e i file completi leggendo direttamente dal filesystem dopo il checkout.

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

Deduplica i candidati sovrapposti. Per ogni candidato non ovvio, verifica che il problema non sia già risolto a HEAD.

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

### 6a — Ripristina stash (se creato in Fase 3b)

```bash
if [ "$STASH_CREATED" = true ]; then
  git stash pop || echo "⚠️ git stash pop fallito — conflitti da risolvere manualmente. Lo stash è intatto: usa 'git stash list' per trovarlo."
fi
```

### 6b — Aggiorna ticket Orchestrator

Prepara il riepilogo tecnico della review (finding, esito, eventuali azioni richieste) e mostra il preview all'utente prima di inviare.

Il campo `description` del ticket è renderizzato da un editor WYSIWYG (HTML, non Markdown): componi il riepilogo in HTML (`<h3>`/`<p>`/`<ul><li>`, `<strong>` per il verdetto) prima di inviarlo nel payload della PATCH — non inviare il Markdown dell'output di Fase 5d as-is.

Gli status disponibili sono elencati direttamente nello schema del tool `update_story` (campo `status`, letto dagli enum PHP di Orchestrator): non serve scaricare nulla da GitHub, il tool stesso rifiuta un valore fuori elenco.

**Se nessun bloccante:**
> "Review completata senza finding bloccanti. Quale status vuoi impostare?"
> [Mostra la lista ammessa dallo schema del tool `update_story` — suggerisci `testing` come default]

**Se ci sono bloccanti:**
> "Trovati [N] finding bloccanti. Propongo di impostare lo status a `todo` per richiedere correzioni. Confermo?"

Chiama `update_story` con `story_id: <ID>`, `status: <status scelto>` e `description: <riepilogo HTML>` senza `confirm`: mostra al dev la differenza calcolata dal tool rispetto allo stato attuale. Attendi conferma esplicita, poi richiama `update_story` con gli stessi campi e `confirm: true`.
