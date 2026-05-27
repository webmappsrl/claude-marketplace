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

## Orchestrator API

Queste istruzioni valgono per tutte le chiamate HTTP a Orchestrator. Usale ogni volta che una fase richiede di leggere o scrivere un ticket.

### Configurazione

- **URL base:** leggi `$ORCHESTRATOR_URL` dall'environment. Se non è impostata usa `https://orchestrator.maphub.it` come default.
- **Token:** salvato in `~/.config/webmapp/orchestrator-token` (plain text, una riga). Se il file non esiste o la chiamata restituisce 401, esegui il login (vedi sotto).

### Login (solo se token assente o scaduto)

Chiedi email e password all'utente, poi:

```bash
ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-https://orchestrator.maphub.it}"
curl -s -X POST "$ORCHESTRATOR_URL/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"<email>","password":"<password>"}' \
  | jq -r '.token'
```

Salva il token restituito:

```bash
mkdir -p ~/.config/webmapp
echo "<token>" > ~/.config/webmapp/orchestrator-token
```

### Lettura ticket (GET — nessuna conferma richiesta)

```bash
ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-https://orchestrator.maphub.it}"
TOKEN=$(cat ~/.config/webmapp/orchestrator-token)
curl -s -X GET "$ORCHESTRATOR_URL/api/stories/<ID>" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Accept: application/json"
```

Se risponde 401: cancella il token e ripeti il login prima di ritentare.

### Creazione ticket (POST — richiede conferma esplicita)

Prima di eseguire, mostra all'utente i campi che verranno inviati e chiedi conferma esplicita. Solo dopo:

```bash
ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-https://orchestrator.maphub.it}"
TOKEN=$(cat ~/.config/webmapp/orchestrator-token)
curl -s -X POST "$ORCHESTRATOR_URL/api/stories" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '<json payload>'
```

Salva l'ID restituito (`$.id`) come `<ID>` del ticket per il resto del workflow.

### Aggiornamento ticket (PATCH — richiede conferma esplicita)

Prima di eseguire, mostra sempre questo messaggio:

> "Sto per aggiornare il ticket oc:\<ID\> con questi campi:
> - `<campo>`: `<valore>`
> Procedo?"

Solo dopo la conferma:

```bash
ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-https://orchestrator.maphub.it}"
TOKEN=$(cat ~/.config/webmapp/orchestrator-token)
curl -s -X PATCH "$ORCHESTRATOR_URL/api/stories/<ID>" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '<json payload>'
```

### Status disponibili (letti dinamicamente)

Prima di proporre uno status, leggi i valori aggiornati da:

```
https://raw.githubusercontent.com/webmappsrl/orchestrator/main/app/Enums/StoryStatus.php
```

Estrai i valori dell'enum e presentali all'utente come lista numerata. Suggerisci quello più appropriato al contesto ma aspetta sempre la scelta esplicita dell'utente.

### Campi accettati (letti dinamicamente)

Prima di costruire un payload POST o PATCH, leggi i campi validati da:

```
https://raw.githubusercontent.com/webmappsrl/orchestrator/main/app/Http/Requests/Api/StoryApiRequest.php
```

Usa solo i campi presenti nelle `rules()` del Form Request. Non inviare campi non dichiarati.

### Regola generale scritture

**Qualsiasi operazione di scrittura su Orchestrator (POST o PATCH) richiede conferma esplicita con preview della modifica prima di eseguire la chiamata HTTP. Nessuna eccezione.**

---

## Fase 0 — Ticket Orchestrator

Il team Webmapp traccia il lavoro su Orchestrator. Ogni ticket ha un ID numerico referenziato come `oc:<ID>` (es. `oc:7815`).

### Caso A — L'utente fornisce un ID ticket

Se l'utente scrive `oc:<ID>` (con o senza contenuto aggiuntivo), leggi il ticket via API seguendo `## Orchestrator API → Lettura ticket`:

```bash
ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-https://orchestrator.maphub.it}"
TOKEN=$(cat ~/.config/webmapp/orchestrator-token 2>/dev/null)
curl -s -X GET "$ORCHESTRATOR_URL/api/stories/<ID>" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Accept: application/json"
```

Se il token non esiste o la risposta è 401, esegui il login (vedi `## Orchestrator API → Login`) e ritenta.

Dal JSON restituito estrai:
- `name` → titolo, usato per il `<feature-slug>`: `<ID>-<titolo-in-kebab-case>`
- `customer_request` → contesto del problema, usato in Fase 2 e Fase 4
- `description` → note tecniche già raccolte, può orientare le domande in Fase 2
- `type` → orienta il tono dell'overview

Mostra all'utente un riepilogo del ticket letto prima di procedere alla Fase 1.

### Caso B — L'utente non ha un ticket

Procedi con il workflow usando le informazioni disponibili. Al termine della Fase 3 (Challenge), prima di scrivere i documenti, proponi il testo completo di un ticket con questo formato:

```
name: <titolo sintetico della feature>

type: Feature

customer_request:
<descrizione del problema e della soluzione emersa dalle fasi 0-3, 5-10 righe, linguaggio non tecnico>

description:
<approccio tecnico scelto, moduli coinvolti, vincoli emersi dalla Challenge>
```

Chiedi all'utente di confermarlo o modificarlo. Una volta approvato, crea il ticket via API seguendo `## Orchestrator API → Creazione ticket`. Salva l'ID restituito e usalo come `<ID>` per tutto il resto del workflow.

Se l'utente non vuole creare il ticket ora, procedi senza ID: usa solo il titolo kebab-case come slug.

### Aggiornamenti espliciti durante il workflow

Se in qualsiasi momento l'utente chiede di aggiornare un campo del ticket (es. "aggiorna lo status a progress", "scrivi nelle note dev che…"), esegui un PATCH seguendo `## Orchestrator API → Aggiornamento ticket`. Mostra sempre il preview della modifica e attendi conferma prima di inviare.

### Da estrarre in entrambi i casi

- `<feature-slug>`: `<ID>-<titolo-in-kebab-case>` (con ticket) o `<titolo-in-kebab-case>` (senza)
- **Riferimento ticket in tutti i documenti:** ogni file creato nelle fasi successive deve riportare in testa `> Ticket: oc:<ID>` (ometti se non c'è ID).
- **Commit convention:** tutti i commit usano `oc:<ID>` come scope (es. `feat(oc:7815): add OSM POI import action`). Senza ticket, usa il titolo kebab-case come scope.

---

## Fase 1 — Leggi CLAUDE.md

Leggi il file `CLAUDE.md` nella root del progetto target.

- Se non esiste, segnalalo all'utente e procedi con le informazioni disponibili.
- Estrai: stack tecnologico, convenzioni di test, struttura cartelle, istruzioni specifiche al team.
- Tieni queste informazioni attive per tutto il workflow.

---

## Fase 2 — Reverse Interaction (obbligatoria, non skippabile)

### 2a — Mappatura domini e submodule

Prima di fare qualsiasi domanda, ispeziona il progetto e identifica:

1. **Submodule presenti** — esegui `git submodule status` e leggi `.gitmodules`. Nota quali repo sono inclusi (es. `wm-package` per backend, `wm-core` / `map-core` per frontend).
2. **Dominio della feature** — sulla base del ticket e del codice, classifica la feature:
   - **Custom** — logica specifica di questo progetto, il codice va nel repo principale
   - **Package/submodule** — logica generica riusabile, il codice va nel submodule appropriato
   - **Misto** — parte nel repo principale, parte nel submodule (specifica quale parte va dove)
3. **Dichiarazione esplicita** — prima della prima domanda scrivi un messaggio con:
   - Submodule trovati e loro scopo
   - Classificazione della feature (custom / package / misto)
   - Per ogni file o modulo che verrà toccato, il repo di destinazione esplicito

Questa classificazione rimane attiva per tutto il workflow: overview.md, plan.md e ogni step del piano devono sempre indicare il repo di destinazione per ogni file.

### 2b — Dialogo

Conduci un dialogo socratico con l'utente: **una domanda alla volta**, aspetta la risposta, poi formula la successiva tenendo conto di ciò che hai appena sentito. Non presentare mai più domande in un unico messaggio.

**Regole:**
- Una domanda per messaggio. Sempre.
- Minimo 5 domande totali. Puoi farne di più se le risposte aprono nuovi buchi di contesto.
- Le `Note di sviluppo` del ticket possono orientare le domande, non ridurne il numero minimo.
- L'unica eccezione per scendere sotto 5 è una giustificazione esplicita scritta nel tuo messaggio (es. "Il ticket e le note di sviluppo coprono già questi aspetti — le 5 domande sarebbero ridondanti perché…").
- Ogni domanda deve essere costruita sulla risposta precedente, non preparata in anticipo.
- Procedi alla Fase 3 solo dopo aver fatto almeno 5 domande e ricevuto tutte le risposte.
- **Non chiedere ciò che puoi leggere nel codice o nel database.** Per ogni potenziale domanda segui questo protocollo obbligatorio in tre passi — non puoi saltarli:
  1. **Cerca nel codice** — modelli, migration, config, CLAUDE.md. Hai trovato la risposta? Usala, non fare la domanda.
  2. **Cerca nel db locale** — che contiene un dump veritiero dei dati di produzione. Interrogalo con `php artisan tinker`, query SQL diretta, o equivalente per lo stack del progetto. Hai trovato la risposta? Usala, non fare la domanda.
  3. **Solo se entrambe le ricerche sono fallite** — formula la domanda arricchita dal contesto trovato nei passi 1 e 2: cita esplicitamente cosa hai già capito e cosa rimane aperto. Non fare domande che ignorano ciò che hai già trovato.

- **Ogni domanda deve includere un consiglio da best practice.** Non aspettare che l'utente lo chieda. Dopo aver posto la domanda aggiungi sempre una riga "💡 Best practice:" con la raccomandazione tecnica più rilevante per quel problema specifico, così l'utente può decidere con più contesto. Questa riga è obbligatoria — una domanda senza consiglio è incompleta.
  Prima di ogni domanda scrivi esplicitamente: *"Ho cercato nel codice [cosa hai cercato e dove] e nel db [query eseguita] — non ho trovato risposta sufficiente, quindi chiedo:"*. Se non scrivi questa riga, non puoi fare la domanda.

**Aree da coprire nel dialogo (adatta e riordina in base alle risposte):**
- Perché ora? Qual è il trigger business/tecnico che rende necessaria questa feature?
- Chi la usa? Quali utenti o sistemi interagiscono con essa, e in quali condizioni limite?
- Cosa non deve fare? Scope esplicito di ciò che è out-of-scope.
- Ci sono vincoli tecnici noti? (performance, compatibilità, dipendenze legacy, deadline)
- Come si misura il successo? Quali test o comportamenti osservabili confermano che è fatta bene?
- La feature introduce testi visibili all'utente? Se sì, **ispeziona autonomamente** il repo per determinare:
  - Lingua di default (cerca `config/app.php` → `locale`, `i18n.config.*` → `defaultLocale`, `nuxt.config.*` → `i18n.defaultLocale`, `.env` → `APP_LOCALE`, o equivalenti per lo stack)
  - Lingue disponibili (elenca le cartelle/file in `resources/lang/`, `lang/`, `locales/`, `src/locales/`, `public/locales/` o equivalenti)
  Documenta quanto trovato prima di fare domande. Chiedi all'utente solo se non riesci a determinare né la lingua di default né le lingue disponibili.

---

## Fase 3 — Challenge (review adversariale)

La Challenge si svolge in due momenti:

**Momento 1 — Analisi completa (un unico messaggio)**

Presenta tutti e 5 gli assi in un solo messaggio. Per ogni asse scrivi almeno un punto concreto — non puoi scrivere "nessun rischio" senza motivazione esplicita. L'obiettivo è trovare debolezze, non difendere la soluzione.

1. **Assunzioni fragili** — Quali ipotesi stai facendo che potrebbero essere false? Cosa succede se lo sono?
2. **Rischi architetturali** — Dove questo design crea accoppiamento, rigidità o debito tecnico futuro?
3. **Blind spot** — Cosa non stai considerando? Edge case, utenti atipici, comportamenti inattesi?
4. **Worst case** — Se qualcosa va storto in produzione, qual è lo scenario peggiore? È recuperabile?
5. **Difficoltà di rollback** — Quanto è facile tornare indietro? Ci sono migrazioni di dati, API breaking change, o dipendenze esterne che rendono il rollback costoso?

**Momento 2 — Dialogo asse per asse**

Dopo aver presentato l'analisi, affronta gli assi uno alla volta in ordine di criticità (dal più critico al meno). Per ogni asse:
- Riassumi il rischio in una riga
- Proponi come intendi gestirlo
- Chiedi all'utente se vuole modificare l'approccio su quel punto

Aspetta la risposta prima di passare all'asse successivo. Se la risposta apre nuovi scenari non coperti dall'analisi iniziale, aggiorna l'asse o aggiungi considerazioni prima di procedere. Solo dopo aver discusso tutti gli assi critici procedi alla Fase 4.

---

## Fase 4 — Scrivi `overview.md`

**I file di documentazione seguono il codice, non il repo principale.**

Per ogni repo coinvolto dalla feature (principale + eventuali submodule) crea un `overview.md` separato nella cartella `docs/features/<feature-slug>/` di quel repo. Non accentrare tutto nel repo principale se il codice è distribuito.

| Dove va il codice | Dove va la documentazione |
|---|---|
| Repo principale | `docs/features/<feature-slug>/overview.md` nel repo principale |
| Submodule `wm-package` | `docs/features/<feature-slug>/overview.md` in `wm-package/` |
| Submodule `wm-core` | `docs/features/<feature-slug>/overview.md` in `wm-core/` |

Se la feature è interamente custom (nessun submodule coinvolto), un solo `overview.md` nel repo principale è sufficiente.

Il `<feature-slug>` è `<ID>-<titolo-in-kebab-case>` (es. `7815-creazione-poi-tramite-osm-id`). Se non c'è ticket, usa solo il titolo kebab-case. Lo slug è lo stesso in tutti i repo per mantenere la tracciabilità.

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

**Prima di invocare writing-plans**, leggi esplicitamente tutti gli `overview.md` generati in Fase 4 (uno per ogni repo coinvolto) e costruisci un briefing strutturato da passare come contesto a writing-plans. Il briefing deve contenere:

- Ticket: `oc:<ID>` e titolo
- Repo coinvolti e classificazione (custom / package / misto)
- Requisiti dalla sezione "Requisiti" di ogni overview.md
- Rischi e decisioni emerse dalla Challenge (Fase 3) e recepite nell'overview
- File da creare/modificare per repo, dalla sezione "Moduli toccati"
- Vincoli tecnici emersi dal dialogo in Fase 2

Questo briefing è lo spec che writing-plans usa per generare il piano — senza di esso writing-plans parte cieco.

**REQUIRED SUB-SKILL:** Invoca `superpowers:writing-plans` con il briefing sopra come contesto, e questi override:

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

Aspetta che l'utente confermi di aver completato la fase di design (o decida di saltarla) prima di procedere con 6b (creazione branch).

### 6b — Creazione branch (obbligatoria, prima di scrivere qualsiasi file)

<HARD-GATE>
Nessun file può essere creato o modificato prima che esista un branch dedicato alla feature. Questo vale sempre, con o senza ticket Orchestrator.
</HARD-GATE>

Deriva il nome del branch dalla feature:

| Caso | Nome branch |
|---|---|
| Ticket presente | `feature/oc-<ID>-<titolo-in-kebab-case>` (es. `feature/oc-7815-creazione-poi-tramite-osm-id`) |
| Nessun ticket | `feature/<titolo-in-kebab-case>` (es. `feature/creazione-poi-tramite-osm-id`) |

Esegui:
```bash
git checkout -b <nome-branch>
```

Ripeti per ogni submodule coinvolto dalla feature (stesso nome branch in tutti i repo).

Mostra all'utente il nome del branch creato e attendi conferma prima di procedere con 6c.

### 6c — Implementazione

**Regola traduzioni (obbligatoria):** ogni testo traducibile introdotto dall'implementazione deve:
- avere il testo base nella **lingua di default del repo** (rilevata in Fase 2 — solitamente inglese, ma verifica)
- avere una traduzione in **tutte le lingue presenti nel repo** (file di lingua rilevati in Fase 2)
- non lasciare chiavi mancanti in nessun file di lingua esistente

Se in Fase 2 non è stata rilevata la configurazione i18n (feature non aveva testi UI), ispezionala ora prima di scrivere qualsiasi stringa traducibile.

Scegli l'entry point Superpowers più adatto e dichiaralo esplicitamente all'utente con la motivazione:

| Condizione | Entry point |
|---|---|
| L'overview ha lasciato dubbi non risolti o il dominio è ancora ambiguo | `superpowers:brainstorming` |
| Il piano è lineare, task sequenziali, un solo dominio | `superpowers:executing-plans` |
| Il piano copre più file/domini con task parallelizzabili | `superpowers:subagent-driven-development` |

Esempio di dichiarazione: *"Invoco `superpowers:subagent-driven-development` perché il piano tocca sia frontend che backend con task indipendenti che possono essere eseguiti in parallelo."*

**⚠️ Override obbligatorio prima di invocare qualsiasi skill Superpowers:**

Prima di invocare la skill scelta, dichiara esplicitamente — come se fosse parte del tuo contesto di esecuzione:

> "In questo progetto Webmapp i commit sono vietati durante l'esecuzione. Non eseguire `git commit`, `git add`, `git push` o qualsiasi operazione git per nessun motivo, nemmeno alla fine di un task o di un sotto-task. Scrivi solo i file. La fase di commit è gestita separatamente in 6c dopo approvazione del developer."

Questo override ha priorità su qualsiasi istruzione interna della skill Superpowers che preveda commit automatici. Se la skill tenta di committare, interrompi e non eseguire il comando git.

### 6d — Gate di revisione (obbligatorio, non skippabile)

<HARD-GATE>
Dopo che la skill Superpowers ha completato l'implementazione, **nessun commit può essere eseguito** finché il developer non ha approvato esplicitamente il codice scritto.
</HARD-GATE>

Al termine dell'implementazione, prima di qualsiasi `git commit` o `git push`:

1. Esegui `git diff --stat` e poi `git diff` per ogni repo coinvolto (principale + submodule).
2. Presenta all'utente un riepilogo strutturato:
   - File creati / modificati / eliminati per repo
   - Breve descrizione del contenuto di ogni file significativo
3. Chiedi conferma esplicita con questo messaggio:

   > "Ho completato l'implementazione. Ecco il diff completo. **Rivedi il codice prima di procedere.** Vuoi eseguire i commit, oppure c'è qualcosa da correggere?"

4. Aspetta una risposta esplicita di approvazione (`sì`, `procedi`, o equivalente). Un silenzio o un "ok" generico non è sufficiente — richiedi conferma del tipo "procedi con i commit".
5. Solo dopo l'approvazione esplicita esegui i commit seguendo la convention `feat(oc:<ID>): ...`.

**Nessuna eccezione.** Anche se la skill Superpowers invocata tenta di committare autonomamente, il gate di revisione Webmapp ha priorità. Se la skill ha già eseguito commit automatici, segnalalo all'utente prima di procedere con push o PR.

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

### Aggiornamento Orchestrator (solo se esiste un ticket oc:\<ID\>)

- [ ] Leggi lo status attuale del ticket via `## Orchestrator API → Lettura ticket`
- [ ] Leggi gli status disponibili da `StoryStatus.php` su GitHub (vedi `## Orchestrator API → Status disponibili`)
- [ ] Suggerisci lo status più appropriato al contesto (es. `testing` se ci sono test da verificare, `done` se tutto è completato e i test passano) e presenta la lista completa — aspetta la scelta esplicita dell'utente
- [ ] Prepara la bozza di `description` (note dev) con:
  - Link alla cartella `docs/features/<feature-slug>/`
  - Riepilogo tecnico di cosa è stato implementato (file creati/modificati, approccio usato)
  - Tono tecnico, rivolto al team
- [ ] Prepara la bozza di `customer_request` (risposta cliente) con:
  - Descrizione in linguaggio non tecnico di cosa è stato fatto e perché
  - Niente nomi di file, classi, branch o dettagli implementativi
  - Tono chiaro e orientato al beneficio per l'utente finale
- [ ] Mostra entrambe le bozze all'utente e chiedi approvazione esplicita — la `customer_request` è letta dal cliente, richiede revisione attenta
- [ ] Solo dopo approvazione esplicita, esegui il PATCH seguendo `## Orchestrator API → Aggiornamento ticket` con i tre campi: `status`, `description`, `customer_request`
