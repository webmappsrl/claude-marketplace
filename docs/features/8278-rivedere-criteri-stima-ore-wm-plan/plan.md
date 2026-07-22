> Ticket: oc:8278

# Rivedere criteri di stima ore in wm-plan Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Riscrivere la `Fase: estimation` di `plugins/wm-skills/skills/wm-plan/SKILL.md` per eliminare il bias di overstima sistematica, introducendo classificazione per-componente (scrittura pura / decisioni aperte), buffer per-componente invece di forfettario, tempo di pianificazione misurato via timestamp, e tracciabilità storica del criterio usato.

**Architecture:** Nessun codice eseguibile: la feature modifica solo il contenuto testuale (Markdown) della skill `wm-plan`, che è un file di istruzioni interpretato da Claude a runtime. Ogni task sostituisce con precisione una sezione/sottosezione esistente. La "verifica" di ogni task non è un test automatico ma un controllo strutturale (grep dei marcatori richiesti + rilettura della sezione) perché non esiste codice da eseguire.

**Tech Stack:** Markdown, nessuna dipendenza.

## Global Constraints

- File unico da modificare: `plugins/wm-skills/skills/wm-plan/SKILL.md`
- Non toccare la logica di `Fase: challenge` in sé — solo il suo output diventa input alla classificazione in `Fase: estimation`
- Non modificare il meccanismo di stima per Bug/Task (restano non stimati in ore, invariato)
- Il coefficiente di velocità per-dev NON va implementato in questo ciclo (out of scope esplicito nell'overview)
- Nessun commit o branch automatico: ogni step di commit in questo piano è un'istruzione testuale per l'utente, non un'azione da eseguire autonomamente
- Commit convention: `feat(oc:8278): ...`
- Validare sempre con `claude plugin validate .` prima di considerare un task concluso, per garantire che il frontmatter e la struttura Markdown della skill restino validi

---

### Task 1: Registrare il timestamp reale di inizio pianificazione in `Fase: ticket`

**Files:**
- Modify: `plugins/wm-skills/skills/wm-plan/SKILL.md:248-258`

**Interfaces:**
- Produces: un timestamp `planning_start_at` (formato ISO, es. `2026-07-22T07:45:47+02:00`) che il Task 3 (estimation: analisi) userà per calcolare le ore di pianificazione misurate.

- [ ] **Step 1: Leggi il contenuto attuale della sezione**

Il contenuto attuale (righe 248-258) è:

```markdown
## Fase: ticket

Il team Webmapp traccia il lavoro su Orchestrator. Ogni ticket ha un ID numerico referenziato come `oc:<ID>` (es. `oc:7815`).

All'inizio del workflow, presenta sempre questo menu all'utente:

> Come vuoi procedere?
> - **A)** Ho un ticket esistente (`oc:<ID>`)
> - **B)** Voglio creare un nuovo ticket
> - **C)** Ho una trascrizione/brief cliente → crea tag con più ticket

In base alla scelta:
```

- [ ] **Step 2: Sostituisci con la versione che registra il timestamp di inizio**

Nuovo contenuto:

```markdown
## Fase: ticket

Il team Webmapp traccia il lavoro su Orchestrator. Ogni ticket ha un ID numerico referenziato come `oc:<ID>` (es. `oc:7815`).

**Registra subito il timestamp di inizio pianificazione** (`planning_start_at`), eseguendo:

```bash
date -u +"%Y-%m-%dT%H:%M:%S%z"
```

Tieni questo valore attivo per tutto il workflow — serve in `Fase: estimation` per calcolare il tempo di pianificazione effettivamente trascorso (misurato, non stimato).

All'inizio del workflow, presenta sempre questo menu all'utente:

> Come vuoi procedere?
> - **A)** Ho un ticket esistente (`oc:<ID>`)
> - **B)** Voglio creare un nuovo ticket
> - **C)** Ho una trascrizione/brief cliente → crea tag con più ticket

In base alla scelta:
```

- [ ] **Step 3: Verifica la modifica**

```bash
grep -n "planning_start_at" plugins/wm-skills/skills/wm-plan/SKILL.md
```

Expected: due occorrenze nella sezione `## Fase: ticket` (definizione + riferimento futuro).

- [ ] **Step 4: Valida la skill**

```bash
claude plugin validate .
```

Expected: nessun errore di schema.

- [ ] **Step 5: Commit**

```bash
git add plugins/wm-skills/skills/wm-plan/SKILL.md
git commit -m "feat(oc:8278): registra timestamp inizio pianificazione in Fase: ticket"
```

---

### Task 2: Riscrivere `estimation: analisi` con classificazione per-componente e buffer per-componente

**Files:**
- Modify: `plugins/wm-skills/skills/wm-plan/SKILL.md:676-693`

**Interfaces:**
- Consumes: `planning_start_at` prodotto dal Task 1.
- Produces: la struttura di breakdown "Misurato + Stimato = Totale" con classificazione per componente, che il Task 3 (estimation: conferma) e il Task 4 (estimation: scrittura su Orchestrator) riutilizzano.

- [ ] **Step 1: Leggi il contenuto attuale della sezione**

Contenuto attuale (righe 676-693):

```markdown
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
```

- [ ] **Step 2: Sostituisci con la nuova struttura di classificazione per-componente**

Nuovo contenuto:

```markdown
### estimation: analisi

Basandoti sull'overview approvato e sull'esito della Fase: challenge, classifica ogni componente della stima in una di due categorie:

- **Scrittura pura** — zero domande aperte residue su "come deve comportarsi" dopo overview + challenge (specifica già completa: campi, endpoint, comportamento, edge case). Buffer 0%.
- **Decisioni aperte** — restano scelte UX/comportamentali da prendere, o reverse-engineering di comportamento legacy non documentato. Buffer 20-30%.

**Calcola il tempo di pianificazione misurato:**

```bash
NOW=$(date -u +"%Y-%m-%dT%H:%M:%S%z")
```

Confronta `NOW` con `planning_start_at` (registrato in Fase: ticket) e calcola la differenza in ore — questo è un dato misurato, non stimato.

Produci la stima con questa struttura:

> **Stima proposta**
>
> | Componente | Classificazione | Ore | Buffer | Note |
> |---|---|---|---|---|
> | Pianificazione (ticket → reverse-interaction → overview → challenge → estimation) | — | \<M\>h (misurata) | — | Timestamp reali: \<planning_start_at\> → \<NOW\> |
> | \<componente 1\> | Scrittura pura / Decisioni aperte | \<X\>h | 0% / 20-30% | \<motivazione tecnica\> |
> | \<componente 2\> | Scrittura pura / Decisioni aperte | \<Y\>h | 0% / 20-30% | \<motivazione tecnica\> |
> | Buffer integrazione trasversale (solo se più di un componente) | — | — | 5% sul totale implementazione | Rischio di interazione tra componenti, non attribuibile a un singolo componente |
>
> **Misurato: \<M\>h + Stimato: \<S\>h = Totale: \<N\>h**
>
> Confidenza: alta / media / bassa
> *(alta = requisiti chiari e stack noto; media = qualche incertezza tecnica; bassa = dipendenze esterne o requisiti aperti)*

Regole per la stima:
- Il buffer rischio è sempre per-componente (0% scrittura pura, 20-30% decisioni aperte), mai un unico valore forfettario finale
- Aggiungi il buffer di integrazione trasversale (5% sul totale implementazione, esclusa la pianificazione) solo quando la feature coinvolge più di un componente
- La confidenza deve essere coerente con i rischi emersi nella Fase: challenge
- Non stimare meno di 0.5h per qualsiasi feature che tocchi più di un file (era 1h)
- Il coefficiente di velocità per-dev NON va applicato in questo ciclo — resta out of scope
```

- [ ] **Step 3: Verifica la modifica**

```bash
grep -n "Scrittura pura\|Decisioni aperte\|Misurato:\|integrazione trasversale" plugins/wm-skills/skills/wm-plan/SKILL.md
```

Expected: tutte e quattro le stringhe trovate nella sezione `estimation: analisi`.

- [ ] **Step 4: Valida la skill**

```bash
claude plugin validate .
```

Expected: nessun errore di schema.

- [ ] **Step 5: Commit**

```bash
git add plugins/wm-skills/skills/wm-plan/SKILL.md
git commit -m "feat(oc:8278): classificazione per-componente e buffer per-componente in estimation: analisi"
```

---

### Task 3: Riscrivere `estimation: conferma` per distinguere Misurato/Stimato

**Files:**
- Modify: `plugins/wm-skills/skills/wm-plan/SKILL.md:695-701`

**Interfaces:**
- Consumes: la struttura "Misurato + Stimato = Totale" prodotta dal Task 2.
- Produces: il testo di conferma che il dev vede e approva, usato dal Task 4 per il payload di scrittura.

- [ ] **Step 1: Leggi il contenuto attuale della sezione**

Contenuto attuale (righe 695-701):

```markdown
### estimation: conferma

Chiedi al dev:

> "Accetti questa stima di **\<N\> ore**, o vuoi modificarla?"

Aspetta risposta esplicita. Se il dev propone un valore diverso, usalo senza discutere — la stima finale è sempre quella approvata dal dev.
```

- [ ] **Step 2: Sostituisci con la versione che distingue Misurato/Stimato**

Nuovo contenuto:

```markdown
### estimation: conferma

Chiedi al dev, riportando sempre la scomposizione (mai un numero unico fuso):

> "Accetti questa stima — **Misurato: \<M\>h + Stimato: \<S\>h = Totale: \<N\>h** — o vuoi modificarla?"

Aspetta risposta esplicita. Se il dev propone un valore diverso, usalo senza discutere — la stima finale è sempre quella approvata dal dev. Se il dev modifica solo la quota stimata (implementazione) lasciando invariata quella misurata (pianificazione), aggiorna solo `<S>` e ricalcola `<N>`.
```

- [ ] **Step 3: Verifica la modifica**

```bash
grep -n "Misurato: .M.h" plugins/wm-skills/skills/wm-plan/SKILL.md
```

Expected: una occorrenza nella sezione `estimation: conferma`.

- [ ] **Step 4: Valida la skill**

```bash
claude plugin validate .
```

Expected: nessun errore di schema.

- [ ] **Step 5: Commit**

```bash
git add plugins/wm-skills/skills/wm-plan/SKILL.md
git commit -m "feat(oc:8278): distingui Misurato/Stimato in estimation: conferma"
```

---

### Task 4: Aggiungere marcatore di versione del criterio in `estimation: scrittura su Orchestrator`

**Files:**
- Modify: `plugins/wm-skills/skills/wm-plan/SKILL.md:703-727`

**Interfaces:**
- Consumes: il totale `<N>` e la scomposizione Misurato/Stimato approvata nel Task 3.

- [ ] **Step 1: Leggi il contenuto attuale della sezione**

Contenuto attuale (righe 703-727):

```markdown
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
```

- [ ] **Step 2: Sostituisci aggiungendo il marcatore di versione**

Nuovo contenuto:

```markdown
### estimation: scrittura su Orchestrator

Mostra il preview della modifica e chiedi conferma prima di eseguire:

> **Aggiornamento ticket oc:\<ID\>**
>
> | Campo | Valore |
> |-------|--------|
> | `estimated_hours` | `<N>` |
>
> Nota interna (non inviata al campo `estimated_hours`, va aggiunta come nota/description se il campo lo consente): `[stima v2 — per-componente] Misurato: <M>h + Stimato: <S>h = Totale: <N>h`
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

Il marcatore `[stima v2 — per-componente]` identifica le stime prodotte con questo criterio, distinguendole nei dati storici Orchestrator da quelle prodotte con il criterio precedente (buffer forfettario unico) — necessario per validare in cicli futuri se il nuovo criterio riduce davvero il bias di overstima.

Se il PATCH fallisce (risposta non 2xx), avvisa l'utente con `⚠️ Impossibile aggiornare la stima su Orchestrator — procedo comunque.` e continua.
```

- [ ] **Step 3: Verifica la modifica**

```bash
grep -n "stima v2 — per-componente" plugins/wm-skills/skills/wm-plan/SKILL.md
```

Expected: almeno due occorrenze (nel preview e nella spiegazione).

- [ ] **Step 4: Valida la skill**

```bash
claude plugin validate .
```

Expected: nessun errore di schema.

- [ ] **Step 5: Commit**

```bash
git add plugins/wm-skills/skills/wm-plan/SKILL.md
git commit -m "feat(oc:8278): marcatore di versione del criterio in estimation: scrittura su Orchestrator"
```

---

### Task 5: Aggiungere revisione stima mid-execution in `Fase: execution`

**Files:**
- Modify: `plugins/wm-skills/skills/wm-plan/SKILL.md:789` (inserire nuova sottosezione `### execution: re-estimation` subito dopo `### execution: implementation`, prima di `### execution: review-gate`)

**Interfaces:**
- Consumes: la regola generale scritture (preview + conferma) già definita in `## Orchestrator API → Regola generale scritture`.
- Produces: un flusso di aggiornamento `estimated_hours` invocabile durante l'esecuzione, riusando lo stesso pattern PATCH di `estimation: scrittura su Orchestrator`.

- [ ] **Step 1: Leggi il contenuto attuale attorno al punto di inserimento**

Contenuto attuale (righe 812-816):

```markdown
Questo override ha priorità su qualsiasi istruzione interna della skill Superpowers che preveda commit automatici. Se la skill tenta di committare, interrompi e non eseguire il comando git.

### execution: review-gate (obbligatorio, non skippabile)
```

- [ ] **Step 2: Inserisci la nuova sottosezione tra le due**

Nuovo contenuto (sostituisce il blocco sopra):

```markdown
Questo override ha priorità su qualsiasi istruzione interna della skill Superpowers che preveda commit automatici. Se la skill tenta di committare, interrompi e non eseguire il comando git.

### execution: re-estimation (solo se il ticket è di tipo Feature)

Se durante l'implementazione emerge un problema non previsto nell'overview e nella stima originale, e questo problema è stimabile in ore (non un semplice imprevisto trascurabile), proponi al dev una revisione della stima prima di proseguire:

> "Ho trovato \<descrizione problema non previsto\>. Stimo un impatto aggiuntivo di **\<X\>h**, portando il totale da \<N\> a \<N+X\>h. Vuoi che aggiorni la stima su Orchestrator?"

Se il dev conferma, applica `## Orchestrator API → Aggiornamento ticket` (preview + conferma esplicita) per il PATCH `estimated_hours` con il nuovo totale, mantenendo lo stesso marcatore di versione (`[stima v2 — per-componente]`) usato in `estimation: scrittura su Orchestrator`.

Registra sempre l'evento in `Fase: notes` (sezione "Decisioni"), indipendentemente dal fatto che il dev abbia accettato o rifiutato la revisione.

### execution: review-gate (obbligatorio, non skippabile)
```

- [ ] **Step 3: Verifica la modifica**

```bash
grep -n "execution: re-estimation" plugins/wm-skills/skills/wm-plan/SKILL.md
```

Expected: una occorrenza, posizionata tra `execution: implementation` e `execution: review-gate`.

- [ ] **Step 4: Valida la skill**

```bash
claude plugin validate .
```

Expected: nessun errore di schema.

- [ ] **Step 5: Commit**

```bash
git add plugins/wm-skills/skills/wm-plan/SKILL.md
git commit -m "feat(oc:8278): aggiungi execution: re-estimation per revisioni stima mid-execution"
```

---

### Task 6: Registrare i falsi negativi di classificazione in `Fase: notes`

**Files:**
- Modify: `plugins/wm-skills/skills/wm-plan/SKILL.md:852-859`

**Interfaces:**
- Consumes: la classificazione "scrittura pura / decisioni aperte" prodotta dal Task 2.

- [ ] **Step 1: Leggi il contenuto attuale della sezione**

Contenuto attuale (righe 852-859):

```markdown
## Fase: notes

Crea e aggiorna `docs/features/<feature-slug>/notes.md` durante e dopo l'esecuzione.

**Regole:**
- Il file deve esistere al termine del workflow. Un notes.md con "Nessuna deviazione rilevante" è valido. Un notes.md assente non lo è.
- Registra: deviazioni dal piano, bug trovati durante l'implementazione, decisioni prese on-the-fly, follow-up da fare in cicli successivi.
- **Modifiche richieste a posteriori** (dopo l'approvazione del piano ma prima del commit): registrale nella sezione "Decisioni" con una riga che descrive cosa è cambiato e perché — anche se la modifica è stata recepita nel codice, la traccia in notes serve per capire perché il piano è stato superato.
```

- [ ] **Step 2: Sostituisci aggiungendo la regola sui falsi negativi di classificazione**

Nuovo contenuto:

```markdown
## Fase: notes

Crea e aggiorna `docs/features/<feature-slug>/notes.md` durante e dopo l'esecuzione.

**Regole:**
- Il file deve esistere al termine del workflow. Un notes.md con "Nessuna deviazione rilevante" è valido. Un notes.md assente non lo è.
- Registra: deviazioni dal piano, bug trovati durante l'implementazione, decisioni prese on-the-fly, follow-up da fare in cicli successivi.
- **Modifiche richieste a posteriori** (dopo l'approvazione del piano ma prima del commit): registrale nella sezione "Decisioni" con una riga che descrive cosa è cambiato e perché — anche se la modifica è stata recepita nel codice, la traccia in notes serve per capire perché il piano è stato superato.
- **Falsi negativi di classificazione stima** (solo per ticket Feature): se un componente classificato "scrittura pura" in `Fase: estimation` si rivela durante l'esecuzione una "decisione aperta" (richiede scelte UX/comportamentali non previste), registralo esplicitamente in una riga della sezione "Follow-up" o "Decisioni" — questo dato è necessario per calibrare il criterio di classificazione nei cicli successivi.
```

- [ ] **Step 3: Verifica la modifica**

```bash
grep -n "Falsi negativi di classificazione stima" plugins/wm-skills/skills/wm-plan/SKILL.md
```

Expected: una occorrenza nella sezione `Fase: notes`.

- [ ] **Step 4: Valida la skill**

```bash
claude plugin validate .
```

Expected: nessun errore di schema.

- [ ] **Step 5: Commit**

```bash
git add plugins/wm-skills/skills/wm-plan/SKILL.md
git commit -m "feat(oc:8278): registra falsi negativi di classificazione stima in Fase: notes"
```

---

## Self-Review

**1. Spec coverage** (dai Requisiti di `overview.md`):
- Classificazione per-componente → Task 2 ✅
- Buffer per-componente (0% / 20-30%) → Task 2 ✅
- Buffer integrazione trasversale 5% → Task 2 ✅
- Minimo forfettario 1h → 0.5h → Task 2 ✅
- Coefficiente per-dev NON implementato → nessun task lo introduce, confermato in Global Constraints ✅
- Marcatore di versione del criterio → Task 4 ✅
- Falsi negativi di classificazione in Fase: notes → Task 6 ✅
- Revisione stima mid-execution → Task 5 ✅
- Pianificazione misurata via timestamp → Task 1 (cattura) + Task 2 (calcolo) ✅
- Distinzione Misurato/Stimato nel totale mostrato al dev → Task 3 ✅

Nessun gap residuo.

**2. Placeholder scan:** nessun "TBD"/"implementare dopo" — ogni step riporta il testo Markdown completo da scrivere.

**3. Coerenza terminologica:** "Misurato" / "Stimato" / "Totale" usati in modo identico in Task 2, 3, 4; "Scrittura pura" / "Decisioni aperte" usati in modo identico in Task 2 e 6; `[stima v2 — per-componente]` usato in modo identico in Task 4 e 5; `planning_start_at` definito in Task 1 e riusato in Task 2. Nessuna discrepanza.
