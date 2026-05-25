---
name: wm-plan
description: "Use when asked to implement, build, add, or refactor a non-trivial feature — anything that touches multiple files, changes architecture, or introduces new behaviour. Do NOT invoke for simple bug fixes, typo corrections, or read-only questions."
---

# Webmapp Feature Workflow

Workflow obbligatorio prima che venga scritta qualsiasi riga di codice per feature o refactor non banali. Segui le fasi in ordine senza saltarne nessuna.

<HARD-GATE>
Nessun codice può essere scritto prima che `overview.md` e `plan.md` esistano nel filesystem e siano stati esplicitamente approvati dall'utente. Questo vale sempre, indipendentemente dalla semplicità percepita del task.
</HARD-GATE>

---

## Fase 0 — Ticket Orchestrator

Il team Webmapp traccia il lavoro su Orchestrator. Ogni ticket ha un ID numerico referenziato come `oc:<ID>` (es. `oc:7815`).

**Chiedi all'utente:**
> "Qual è l'ID del ticket Orchestrator per questa feature? Incolla anche il contenuto del ticket (Titolo, Richiesta, Note di sviluppo)."

Se l'utente non ha un ticket, procedi comunque con il workflow. Al termine della Fase 3 (Challenge), prima di scrivere i documenti, proponi il testo completo di un ticket da creare su Orchestrator con questo formato:

```
Titolo: <titolo sintetico della feature>

Tipo: Feature

Richiesta:
<descrizione del problema e della soluzione emersa dalle fasi 0-3, 5-10 righe>

Note di sviluppo:
<approccio tecnico scelto, moduli coinvolti, vincoli emersi dalla Challenge>
```

Chiedi all'utente di confermarlo o modificarlo prima di procedere alla Fase 4. Una volta confermato, usa il testo del ticket proposto come riferimento per overview.md e plan.md (senza ID numerico finché non viene creato su Orchestrator).

**Da estrarre dal ticket:**
- `Titolo` → usato per il `<feature-slug>`: `<ID>-<titolo-in-kebab-case>` (es. `7815-creazione-poi-tramite-osm-id`)
- `Richiesta` → contesto del problema da risolvere, usato in Fase 2 e Fase 4
- `Note di sviluppo` → contesto tecnico già raccolto, può ridurre le domande in Fase 2 (ma non eliminarle)
- `Tipo` → Feature / Bug / etc., orienta il tono dell'overview

**Riferimento ticket in tutti i documenti:** ogni file creato nelle fasi successive deve riportare in testa `> Ticket: oc:<ID>`.

**Commit convention:** tutti i commit del workflow usano `oc:<ID>` come scope (es. `feat(oc:7815): add OSM POI import action`).

---

## Fase 1 — Leggi CLAUDE.md

Leggi il file `CLAUDE.md` nella root del progetto target.

- Se non esiste, segnalalo all'utente e procedi con le informazioni disponibili.
- Estrai: stack tecnologico, convenzioni di test, struttura cartelle, istruzioni specifiche al team.
- Tieni queste informazioni attive per tutto il workflow.

---

## Fase 2 — Reverse Interaction (obbligatoria, non skippabile)

Conduci un dialogo socratico con l'utente: **una domanda alla volta**, aspetta la risposta, poi formula la successiva tenendo conto di ciò che hai appena sentito. Non presentare mai più domande in un unico messaggio.

**Regole:**
- Una domanda per messaggio. Sempre.
- Minimo 5 domande totali. Puoi farne di più se le risposte aprono nuovi buchi di contesto.
- Le `Note di sviluppo` del ticket possono orientare le domande, non ridurne il numero minimo.
- L'unica eccezione per scendere sotto 5 è una giustificazione esplicita scritta nel tuo messaggio (es. "Il ticket e le note di sviluppo coprono già questi aspetti — le 5 domande sarebbero ridondanti perché…").
- Ogni domanda deve essere costruita sulla risposta precedente, non preparata in anticipo.
- Procedi alla Fase 3 solo dopo aver fatto almeno 5 domande e ricevuto tutte le risposte.
- **Non chiedere ciò che puoi leggere nel codice o nel database.** Prima di formulare ogni domanda, verifica se la risposta è ricavabile da: codebase (modelli, migration, config, CLAUDE.md) oppure dal database locale — che contiene un dump veritiero dei dati di produzione. Per domande sulla struttura dati, relazioni tra entità o valori reali, interroga direttamente il db locale (`php artisan tinker`, query SQL, o equivalente per lo stack del progetto) prima di chiedere all'utente. Se trovi la risposta, non fare la domanda — dichiarala in una riga ("Ho verificato nel db che X — procedo con questa assunzione"). Le domande sono riservate a contesto business, decisioni di prodotto e vincoli non visibili né nel codice né nei dati.

**Aree da coprire nel dialogo (adatta e riordina in base alle risposte):**
- Perché ora? Qual è il trigger business/tecnico che rende necessaria questa feature?
- Chi la usa? Quali utenti o sistemi interagiscono con essa, e in quali condizioni limite?
- Cosa non deve fare? Scope esplicito di ciò che è out-of-scope.
- Ci sono vincoli tecnici noti? (performance, compatibilità, dipendenze legacy, deadline)
- Come si misura il successo? Quali test o comportamenti osservabili confermano che è fatta bene?

---

## Fase 3 — Challenge (review adversariale)

Attacca il tuo approccio su questi 5 assi. L'obiettivo è trovare debolezze, non difendere la soluzione.

Per ogni asse scrivi almeno un punto concreto — non puoi scrivere "nessun rischio" senza motivazione esplicita.

1. **Assunzioni fragili** — Quali ipotesi stai facendo che potrebbero essere false? Cosa succede se lo sono?
2. **Rischi architetturali** — Dove questo design crea accoppiamento, rigidità o debito tecnico futuro?
3. **Blind spot** — Cosa non stai considerando? Edge case, utenti atipici, comportamenti inattesi?
4. **Worst case** — Se qualcosa va storto in produzione, qual è lo scenario peggiore? È recuperabile?
5. **Difficoltà di rollback** — Quanto è facile tornare indietro? Ci sono migrazioni di dati, API breaking change, o dipendenze esterne che rendono il rollback costoso?

Dopo la Challenge, comunica all'utente le criticità emerse e chiedi se vuole modificare l'approccio prima di procedere.

---

## Fase 4 — Scrivi `overview.md`

Crea il file `docs/features/<feature-slug>/overview.md`.

Il `<feature-slug>` è `<ID>-<titolo-in-kebab-case>` (es. `7815-creazione-poi-tramite-osm-id`). Se non c'è ticket, usa solo il titolo kebab-case.

**Struttura obbligatoria:**

```markdown
> Ticket: oc:<ID>

# <Titolo dal ticket>

## Cosa cambia
[Descrizione concisa di cosa il sistema farà di diverso dopo questa feature]

## Perché
[Motivazione business/tecnica emersa dal ticket e dalla Fase 2]

## Requisiti
- [ ] Requisito funzionale 1
- [ ] Requisito funzionale 2
...

## Rischi
[Criticità emerse dalla Fase 3 con indicazione di come vengono mitigate]

## Out of scope
[Cosa esplicitamente NON viene fatto in questo ciclo]

## Moduli toccati
[Lista di file, moduli o servizi che vengono modificati o creati]
```

Mostra il file all'utente e attendi approvazione esplicita prima di procedere.

---

## Fase 5 — Scrivi `plan.md`

**REQUIRED SUB-SKILL:** Invoca `superpowers:writing-plans` per generare il piano, con questi override:

- **Percorso di salvataggio:** `docs/features/<feature-slug>/plan.md`
- **Header obbligatorio:** il piano inizia con `> Ticket: oc:<ID>`
- **Commit convention:** tutti i commit usano `feat(oc:<ID>): ...` / `fix(oc:<ID>): ...` / `refactor(oc:<ID>): ...`
- **⚠️ No commit o branch automatici:** i commit nel piano sono istruzioni testuali per l'utente, non azioni da eseguire autonomamente. Claude non esegue `git commit`, `git push` o crea branch senza conferma esplicita dell'utente per ogni singolo commit.

Durante la scrittura del piano applica la skill `wm-skills:our-code-style` per allineare le scelte implementative alle convenzioni Webmapp.

Mostra il piano all'utente e attendi approvazione esplicita prima di procedere.

---

## Fase 6 — Esecuzione

### 6a — Design (se applicabile)

Se il piano include componenti UI/UX (nuove interfacce, layout, prototipi, slide, one-pager), prima di eseguire il codice proponi all'utente di usare **Claude Design** (`claude.ai/design`):

> "Questa feature ha componenti visual. Ti consiglio di prototipare il design su claude.ai/design prima di implementare — puoi poi trasferire il risultato direttamente a Claude Code con un'istruzione."

Aspetta che l'utente confermi di aver completato la fase di design (o decida di saltarla) prima di procedere con 6b.

### 6b — Implementazione

Scegli l'entry point Superpowers più adatto e dichiaralo esplicitamente all'utente con la motivazione:

| Condizione | Entry point |
|---|---|
| L'overview ha lasciato dubbi non risolti o il dominio è ancora ambiguo | `superpowers:brainstorming` |
| Il piano è lineare, task sequenziali, un solo dominio | `superpowers:executing-plans` |
| Il piano copre più file/domini con task parallelizzabili | `superpowers:subagent-driven-development` |

Esempio di dichiarazione: *"Invoco `superpowers:subagent-driven-development` perché il piano tocca sia frontend che backend con task indipendenti che possono essere eseguiti in parallelo."*

---

## Fase 7 — Mantieni `notes.md`

Crea e aggiorna `docs/features/<feature-slug>/notes.md` durante e dopo l'esecuzione.

**Regole:**
- Il file deve esistere al termine del workflow. Un notes.md con "Nessuna deviazione rilevante" è valido. Un notes.md assente non lo è.
- Registra: deviazioni dal piano, bug trovati durante l'implementazione, decisioni prese on-the-fly, follow-up da fare in cicli successivi.

**Struttura consigliata:**

```markdown
> Ticket: oc:<ID>

# Notes — <Titolo feature>

## Deviazioni dal piano
[Task che hanno richiesto approcci diversi da quanto pianificato, con motivazione]

## Bug trovati
[Problemi scoperti durante l'implementazione, anche se già risolti]

## Decisioni
[Scelte tecniche non ovvie prese durante l'esecuzione]

## Follow-up
[Cose da fare in cicli successivi, tech debt consapevolmente accettato]
```

---

## Fase 8 — Aggiorna CLAUDE.md del progetto target

Aggiorna il `CLAUDE.md` nella root del progetto target con le informazioni prodotte dal workflow. Se il file non esiste, crealo.

**Sezione `## Feature disponibili`** — aggiorna o crea con una riga per ogni feature completata:

```markdown
## Feature disponibili

| Feature | Ticket | Moduli toccati | Note |
|---|---|---|---|
| <Titolo feature> | oc:<ID> | `path/modulo1`, `path/modulo2` | <una riga: cosa fa> |
```

Se la sezione esiste già, aggiungi la nuova riga senza toccare quelle precedenti.

**Sezione `## Decisioni architetturali`** — aggiungi le scelte non ovvie emerse dalla Challenge (Fase 3) e dai notes (Fase 7) che un futuro Claude dovrebbe conoscere per non ripercorrere gli stessi ragionamenti:

```markdown
## Decisioni architetturali

### <Titolo feature> (oc:<ID>)
- <decisione 1: cosa e perché>
- <decisione 2: cosa e perché>
```

Se la sezione esiste già, aggiungi il nuovo blocco in cima (le decisioni recenti sono le più rilevanti).

Mostra le modifiche al `CLAUDE.md` all'utente prima di scriverle.

---

## Composizione con altre skill Webmapp

- **`wm-skills:our-code-style`** — applica in Fase 5 (scrittura plan) e Fase 6 (esecuzione)
- **`wm-skills:our-pr-checklist`** — applica dopo la Fase 7, prima di aprire la PR
- **`wm-skills:our-deploy-post-merge`** — applica dopo il merge della PR

---

## Checklist di completamento

Prima di dichiarare il workflow concluso, verifica che esistano tutti e tre i file:

- [ ] `docs/features/<feature-slug>/overview.md` — approvato dall'utente, riferimento ticket presente
- [ ] `docs/features/<feature-slug>/plan.md` — approvato dall'utente, riferimento ticket presente
- [ ] `docs/features/<feature-slug>/notes.md` — compilato (anche solo con "Nessuna deviazione"), riferimento ticket presente
- [ ] `CLAUDE.md` del progetto target aggiornato — sezione "Feature disponibili" e "Decisioni architetturali"
