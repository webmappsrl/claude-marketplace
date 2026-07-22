> Ticket: oc:8282

# Clear del context in wm-plan dopo reverse-interaction e dopo implementation — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Isolare sempre il riepilogo del diff in `execution: review-gate` di `wm-plan` tramite un subagente cieco (stesso pattern di `challenge: subagent`), e rendere non skippabile la registrazione del timestamp di pianificazione in `Fase: estimation`.

**Architecture:** Modifiche testuali a `plugins/wm-skills/skills/wm-plan/SKILL.md` (nessun codice eseguibile — è una skill Claude Code in Markdown) più aggiornamento della tabella coupling e delle sezioni "Feature disponibili"/"Decisioni architetturali" in `CLAUDE.md` (root repo).

**Tech Stack:** Markdown, Claude Code plugin/skill format, `claude plugin validate` come unico "test" disponibile per questo tipo di file.

## Global Constraints

- Nessun commit o branch automatico: ogni step di commit in questo piano è un'istruzione testuale per lo sviluppatore, non un'azione da eseguire autonomamente durante `execution: implementation`.
- Commit convention: `feat(oc:8282): <descrizione>`.
- Validare con `claude plugin validate .` dalla root del repo dopo ogni modifica a `SKILL.md`.
- Non toccare `Fase: overview`, `challenge: subagent`, `subagent-driven-development` (skill esterna upstream) — sono esplicitamente out of scope.
- Il subagente di review-gate non deve mai ricevere un riassunto della conversazione precedente — solo repo/branch coinvolti + istruzioni di analisi, sullo stesso modello di `challenge: subagent` (righe 628-650 di `SKILL.md`).

---

### Task 1: Isolare il riepilogo del diff in `execution: review-gate`

**Files:**
- Modify: `plugins/wm-skills/skills/wm-plan/SKILL.md:857-880` (sezione `### execution: review-gate`)

**Interfaces:**
- Consumes: nessuna dipendenza da altri task di questo piano.
- Produces: la sezione `### execution: review-gate` aggiornata — i task successivi (2, 3, 4) non dipendono dal suo output testuale specifico, solo dal fatto che la sezione esista con l'header invariato.

- [ ] **Step 1: Sostituire il contenuto della sezione `### execution: review-gate`**

Apri `plugins/wm-skills/skills/wm-plan/SKILL.md` e sostituisci il blocco che va da `### execution: review-gate (obbligatorio, non skippabile)` (riga 857) fino a (esclusa) `### execution: formal-review` (riga 882) con:

```markdown
### execution: review-gate (obbligatorio, non skippabile)

<HARD-GATE>
Dopo che la skill Superpowers ha completato l'implementazione, **nessun commit può essere eseguito** finché il developer non ha approvato esplicitamente il codice scritto.
</HARD-GATE>

Al termine dell'implementazione, prima di qualsiasi `git commit` o `git push`, il riepilogo del diff viene prodotto da un **subagente isolato**, non dal context principale — lo stesso principio di `challenge: subagent`: chi ha scritto/coordinato il codice tende a confermare le proprie scelte invece di valutarle a freddo; un subagente che non ha visto il ragionamento implementativo ha più probabilità di notare un'incongruenza tra intenzione e codice reale.

**Questo isolamento si applica sempre**, indipendentemente da quanto sia stata "pesante" l'implementazione e da quale entry point Superpowers sia stato usato (anche se `subagent-driven-development` ha già eseguito review isolate per-task e una review finale whole-branch, il riepilogo di review-gate resta comunque affidato a un subagente dedicato — è un controllo ridondante ma intenzionale, non un costo da evitare con soglie o condizioni).

#### review-gate: subagent

Lancia un subagente con questo prompt — e **nient'altro** (nessun riassunto della conversazione precedente, nessuna spiegazione di cosa è stato implementato o perché):

```
Nel repository al percorso <path-repo> (branch <nome-branch>), analizza le
modifiche non ancora committate.

Esegui, in questo ordine:
1. `git diff --stat`
2. `git diff --name-status --find-renames --find-copies` (per distinguere
   correttamente rename/copy da coppie "nuovo file + file cancellato")
3. `git diff` (diff completo)

Produci un riepilogo strutturato:
- File creati / modificati / eliminati / rinominati (usa l'esito del punto 2
  per non descrivere un rename come "nuovo + cancellato")
- Breve descrizione del contenuto di ogni file significativo

Ripeti per ogni repo coinvolto se ne viene indicato più di uno (es. submodule).
```

Se sono coinvolti più repo (principale + submodule), passa al subagente l'elenco dei path da analizzare — questa è informazione strutturale (dove guardare), non il ragionamento implementativo che l'isolamento deve escludere.

**Fallback fail-soft:** se lo spawn del subagente fallisce (errore tool, timeout, rate limit), segnala `⚠️ Impossibile isolare il riepilogo del diff — procedo mostrando il diff nel context principale.` ed esegui tu stesso, nel context principale, `git diff --stat` + `git diff` per ogni repo coinvolto, come comportamento di fallback. Non bloccare mai il workflow per questo motivo.

#### review-gate: dialog

1. Presenta all'utente il riepilogo prodotto dal subagente (o dal fallback).
2. Chiedi conferma esplicita con questo messaggio:

   > "Ho completato l'implementazione. Ecco il riepilogo del diff (prodotto da un subagente isolato, senza contesto sulla conversazione precedente). **Rivedi comunque il diff completo prima di procedere** — il riepilogo è un ausilio di orientamento, non sostituisce la lettura del codice. Vuoi eseguire i commit, oppure c'è qualcosa da correggere?"
   >
   > 💡 **Review formale opzionale:** vuoi eseguire una code review strutturata prima dei commit? Invoca `wm-skills:wm-review-ticket oc:<ID>` per finder paralleli e aggiornamento automatico del ticket. Rispondi **sì** per eseguirla ora, **no** per procedere direttamente ai commit.
   >
   > ℹ️ **Differenza con la review formale:** questo riepilogo del subagente è un controllo obbligatorio e leggero (solo diff strutturato). `wm-review-ticket` è un'analisi più approfondita e opzionale.

3. Aspetta una risposta esplicita di approvazione (`sì`, `procedi`, o equivalente). Un silenzio o un "ok" generico non è sufficiente — richiedi conferma del tipo "procedi con i commit".
4. Solo dopo l'approvazione esplicita, **prima di eseguire i commit**, completa la Fase: notes e la Fase: update-context — così tutti i file vengono inclusi nello stesso commit.
5. Esegui i commit seguendo la convention `feat(oc:<ID>): ...`.
6. Dopo i commit, apri la PR verso **`develop`** (non `main`) — è il branch di integrazione Webmapp.

**Nessuna eccezione.** Il subagente produce solo il riepilogo — non decide né esegue commit. Anche se la skill Superpowers invocata tenta di committare autonomamente, il gate di revisione Webmapp ha priorità. Se la skill ha già eseguito commit automatici, segnalalo all'utente prima di procedere con push o PR.
```

- [ ] **Step 2: Verificare che l'header della sezione successiva sia rimasto intatto**

Esegui:
```bash
grep -n "^### execution: formal-review" plugins/wm-skills/skills/wm-plan/SKILL.md
```
Expected: una sola riga trovata, subito dopo il blocco appena scritto — confirma che non è stato accidentalmente cancellato o duplicato l'header della sezione successiva.

- [ ] **Step 3: Validare il plugin**

Esegui dalla root del repo:
```bash
claude plugin validate .
```
Expected: nessun errore di schema su `marketplace.json`, `plugin.json` o `SKILL.md`.

- [ ] **Step 4: Commit**

```bash
git add plugins/wm-skills/skills/wm-plan/SKILL.md
git commit -m "feat(oc:8282): isola il riepilogo del diff in review-gate tramite subagente"
```

---

### Task 2: Fallback esplicito per `planning_start_at` mancante in `Fase: estimation`

**Files:**
- Modify: `plugins/wm-skills/skills/wm-plan/SKILL.md:684-697` (sezione `### estimation: analisi`, subito prima della tabella "Stima proposta")

**Interfaces:**
- Consumes: nessuna dipendenza da Task 1.
- Produces: nessuna dipendenza per i task successivi — è una modifica indipendente nella stessa fase (`Fase: estimation`).

- [ ] **Step 1: Aggiungere il controllo di presenza del timestamp**

Nel file `plugins/wm-skills/skills/wm-plan/SKILL.md`, individua questo blocco esistente (circa riga 691-697):

```markdown
**Calcola il tempo di pianificazione misurato:**

```bash
NOW=$(date -u +"%Y-%m-%dT%H:%M:%S%z")
```

Confronta `NOW` con `planning_start_at` (registrato in Fase: ticket) e calcola la differenza in ore — questo è un dato misurato, non stimato.
```

Sostituiscilo con:

```markdown
**Calcola il tempo di pianificazione misurato:**

```bash
NOW=$(date -u +"%Y-%m-%dT%H:%M:%S%z")
```

Confronta `NOW` con `planning_start_at` (registrato in Fase: ticket) e calcola la differenza in ore — questo è un dato misurato, non stimato.

**Se `planning_start_at` non è stato registrato** (es. la Fase: ticket è stata eseguita senza eseguire il comando `date -u` previsto), non calcolare una quota "misurata" fittizia e non presentarla come se il dato fosse disponibile. Segnala esplicitamente all'utente:

> "⚠️ Non ho registrato il timestamp di inizio pianificazione in Fase: ticket — la quota 'Misurato' non è disponibile per questa stima. Propongo solo la quota 'Stimato', senza il confronto Misurato + Stimato = Totale."

Poi procedi con la tabella sottostante omettendo la riga "Pianificazione" e il calcolo `<M>h`.
```

- [ ] **Step 2: Verificare la modifica**

Esegui:
```bash
grep -n "planning_start_at" plugins/wm-skills/skills/wm-plan/SKILL.md
```
Expected: compaiono sia il riferimento originale in `Fase: ticket` sia il nuovo controllo appena aggiunto in `Fase: estimation`.

- [ ] **Step 3: Validare il plugin**

```bash
claude plugin validate .
```
Expected: nessun errore.

- [ ] **Step 4: Commit**

```bash
git add plugins/wm-skills/skills/wm-plan/SKILL.md
git commit -m "feat(oc:8282): segnala esplicitamente se planning_start_at non è stato registrato"
```

---

### Task 3: Documentare il coupling tra `challenge: subagent` e `execution: review-gate` in CLAUDE.md

**Files:**
- Modify: `CLAUDE.md:34` (tabella "Coupling tra skill", root del repo)

**Interfaces:**
- Consumes: nessuna dipendenza da Task 1/2 (è un aggiornamento di documentazione indipendente, ma logicamente segue la modifica di Task 1).
- Produces: nessuna dipendenza per Task 4.

- [ ] **Step 1: Leggere la tabella coupling esistente**

Esegui:
```bash
grep -n "Coupling tra skill" -A 10 CLAUDE.md
```

- [ ] **Step 2: Aggiungere una riga alla tabella**

Nella sezione `### Coupling tra skill` di `CLAUDE.md`, dopo l'ultima riga della tabella esistente (quella su `wm-plan` / `wm-tag` / caso-c), aggiungi:

```markdown
| `wm-plan` (challenge) | `wm-plan` (review-gate) | Entrambe le sotto-fasi isolano il giudizio in un subagente cieco (solo path/istruzioni, nessun riassunto della conversazione precedente). Se il pattern di isolamento cambia in una sotto-fase, verificare se va aggiornato anche nell'altra. |
```

- [ ] **Step 3: Verificare la modifica**

```bash
grep -n "review-gate" CLAUDE.md
```
Expected: la nuova riga compare nella tabella coupling.

- [ ] **Step 4: Commit**

```bash
git add CLAUDE.md
git commit -m "feat(oc:8282): documenta coupling tra challenge e review-gate in CLAUDE.md"
```

---

### Task 4: Aggiornare "Feature disponibili" e "Decisioni architetturali" in CLAUDE.md

**Files:**
- Modify: `CLAUDE.md` — sezione `## Feature disponibili` (tabella) e sezione `## Decisioni architetturali` (in cima, sopra il blocco più recente esistente)

**Interfaces:**
- Consumes: dipende da Task 1, 2, 3 completati (descrive il risultato finale della feature).
- Produces: nessuno — è l'ultimo task del piano.

- [ ] **Step 1: Aggiungere la riga alla tabella "Feature disponibili"**

Individua la tabella sotto `## Feature disponibili` in `CLAUDE.md` e aggiungi in fondo:

```markdown
| Clear del context in wm-plan dopo reverse-interaction e dopo implementation | oc:8282 | `plugins/wm-skills/skills/wm-plan/SKILL.md` | Isolamento sempre attivo del riepilogo diff in `execution: review-gate` tramite subagente cieco (stesso pattern di `challenge`); fallback esplicito se `planning_start_at` non è stato registrato in `Fase: estimation` |
```

- [ ] **Step 2: Aggiungere il blocco in cima a "Decisioni architetturali"**

Subito dopo l'header `## Decisioni architetturali` (prima del blocco più recente esistente), inserisci:

```markdown
### Clear del context in wm-plan dopo reverse-interaction e dopo implementation (oc:8282)
- **Nessun clear reale del context principale**: il pivot deciso durante la Fase: reverse-interaction ha scartato l'idea di "svuotare" il context — l'isolamento si ottiene solo delegando a subagenti nei punti dove serve un giudizio non contaminato dal ragionamento pregresso (motivazione primaria: indipendenza di giudizio, non economia di context)
- **`Fase: overview` resta invariata**: scritta dal context principale, perché in quel punto del workflow il context è ancora leggero e chi ha condotto `reverse-interaction` ha un vantaggio informativo che un subagente isolato non recupererebbe da un riassunto di seconda mano
- **`execution: review-gate` isola sempre il riepilogo del diff**, senza soglie né eccezioni per la skill di implementazione usata — anche la ridondanza con le review per-task di `subagent-driven-development` è accettata come controllo doppio intenzionale
- **`--find-renames --find-copies` obbligatori** nel subagente di review-gate, per non descrivere un file rinominato come "nuovo + cancellato"
- **Il riepilogo del subagente non sostituisce mai la lettura del diff da parte del developer** — resta un ausilio di orientamento, il gate reale è l'approvazione esplicita del developer sul diff completo
- **Coordinamento tra sottoagenti paralleli durante l'implementazione è fuori scope**: già gestito da `subagent-driven-development` (esecuzione sequenziale, non parallela, con ledger di progresso) — eventuali miglioramenti vanno proposti upstream su `obra/superpowers`, non in `wm-skills`
</markdown>
```

(Nota: l'ultima riga del blocco sopra contiene un errore di battitura innocuo — "</markdown>" non va scritto nel file, è solo un delimitatore di questo step. Copia solo il contenuto Markdown del blocco, non l'ultima riga.)

- [ ] **Step 3: Verificare le modifiche**

```bash
grep -n "oc:8282" CLAUDE.md
```
Expected: compare sia nella tabella "Feature disponibili" sia nel blocco "Decisioni architetturali".

- [ ] **Step 4: Validare il plugin**

```bash
claude plugin validate .
```
Expected: nessun errore.

- [ ] **Step 5: Commit**

```bash
git add CLAUDE.md
git commit -m "feat(oc:8282): aggiorna CLAUDE.md con feature e decisioni architetturali"
```

---

## Note finali

- Non è previsto alcun task di "test automatico" nel senso tradizionale: questo piano modifica solo istruzioni testuali per Claude (Markdown), non codice eseguibile. La verifica di correttezza è: (a) `claude plugin validate .` per la validità dello schema, (b) `grep` puntuali per confermare che le stringhe attese siano presenti, (c) revisione umana del testo prodotto, dato che l'unico modo per "eseguire" queste istruzioni è farle seguire da Claude in una sessione futura.
- I 4 task sono sequenziali per semplicità di revisione (ognuno tocca una sezione distinta), ma sono indipendenti tra loro salvo Task 4 che riepiloga il risultato di 1-3.
