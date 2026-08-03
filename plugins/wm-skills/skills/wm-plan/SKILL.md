---
name: wm-plan
description: "Use when asked to implement, build, add, or refactor a non-trivial feature — anything that touches multiple files, changes architecture, or introduces new behaviour. Do NOT invoke for simple bug fixes, typo corrections, or read-only questions."
---

## Header di sessione

**Mostra questa sezione ad ogni invocazione di `wm-plan`, senza eccezioni.** Anche se `wm-plan` è già stato invocato in precedenza nella stessa conversazione (es. richiamato una seconda volta per un altro ticket), l'Header di sessione va comunque mostrato per intero prima di procedere a `Fase: ticket`.

**Nessuna narrazione prima, durante o dopo l'header.** Non annunciare cosa stai per fare ("ora leggo il CLAUDE.md", "faccio i check di versione", "vedo che hai eseguito wm-plan"), non commentare i comandi eseguiti, non introdurre l'header con un saluto. Il primo output di ogni invocazione deve essere il banner stesso, seguito immediatamente — senza testo intermedio — dalle righe prodotte da `### header: versione` e `### header: diagramma`. Il tutto è un unico blocco di output compatto, poi si passa direttamente a `Fase: ticket` senza commento di transizione.

Ad ogni invocazione, mostra questo banner come primissimo output, prima di qualsiasi altro testo:

```
┌──────────────────────────────┐
│         W M · P L A N         │
└──────────────────────────────┘
```

Subito dopo il banner, senza alcuna riga di commento tra l'uno e l'altro, mostra le informazioni di stato raccolte nelle sotto-sezioni seguenti (`### header: versione`, `### header: diagramma`), poi procedi a `Fase: ticket`.

### header: versione

**Versione installata:** v1.1.1

Questo valore è statico, scritto direttamente in questa skill (stesso pattern dell'URL del diagramma in `### header: diagramma`): si aggiorna manualmente ad ogni release, come da checklist in `CLAUDE.md` → `## Versioning del plugin wm-skills`. Non richiede alcuna risoluzione di path a runtime (niente ricerca nella cache dei plugin né `git`), quindi mostra sempre il dato senza rischio di "check non disponibile".

Mostra `Versione installata: v1.1.1` come prima riga di questa sotto-sezione, poi prosegui con il check di aggiornamento disponibile:

Determina la path del repo marketplace installato risolvendo la path del plugin cacheato, **indipendentemente dalla cwd** (questa skill può essere invocata da qualsiasi repo, non solo da `claude-marketplace`):

```bash
SKILL_PATH=$(find ~/.claude/plugins/cache -maxdepth 7 -path '*/wm-skills/*/skills/wm-plan/SKILL.md' 2>/dev/null | head -1)
REPO_PATH=$(git -C "$(dirname "$SKILL_PATH")" rev-parse --show-toplevel 2>/dev/null)
```

- **Se `SKILL_PATH` o `REPO_PATH` sono vuoti** (skill non installata via marketplace, es. sviluppo locale con `/plugin marketplace add .`): mostra `⚠️ Check aggiornamenti non disponibile.` e salta il resto di questa sotto-sezione.
- **Altrimenti**, usa `REPO_PATH` come base:

```bash
git -C "$REPO_PATH" rev-parse --abbrev-ref HEAD 2>/dev/null
git -C "$REPO_PATH" status --porcelain -- plugins/wm-skills/skills/wm-plan/SKILL.md 2>/dev/null
```

**Se il branch non è `main` OPPURE `git status --porcelain` restituisce output non vuoto per `SKILL.md`:**

Mostra:
```
🔧 modalità sviluppo locale — check aggiornamenti saltato
```
e salta il resto di questa sotto-sezione (nessun confronto hash remoto).

**Altrimenti (branch `main`, nessuna modifica locale non committata):**

```bash
LOCAL_HASH=$(git -C "$REPO_PATH" rev-parse HEAD 2>/dev/null)
REMOTE_HASH=$(curl -s "https://api.github.com/repos/webmappsrl/claude-marketplace/commits/main" | jq -r '.sha' 2>/dev/null)
LOCAL_DATE=$(git -C "$REPO_PATH" log -1 --format=%ad --date=format:%Y-%m-%d -- plugins/wm-skills/skills/wm-plan/SKILL.md 2>/dev/null)
```

- Se `LOCAL_HASH` e `REMOTE_HASH` non sono ottenibili (rete assente, comando fallito, output vuoto): mostra `⚠️ Check aggiornamenti non disponibile.` e prosegui senza bloccare.
- Se `LOCAL_HASH` == `REMOTE_HASH`: mostra `✅ wm-plan aggiornato (ultima modifica: $LOCAL_DATE)`.
- Se `LOCAL_HASH` != `REMOTE_HASH`: mostra `⬆️ Aggiornamento disponibile per wm-plan (ultima modifica locale: $LOCAL_DATE) — esegui \`/plugin marketplace update\` per aggiornare.`

### header: diagramma

L'URL dell'Artifact è un dato statico di questa skill, aggiornato qui stesso ad ogni redeploy (vedi regola di rigenerazione in `CLAUDE.md` → `## Diagramma di flusso wm-plan`). Nessun fetch remoto necessario: essendo scritto in `SKILL.md`, è sempre disponibile insieme al resto del contenuto della skill già caricato, e viaggia allineato ad ogni `/plugin marketplace update`.

**URL Artifact:** https://claude.ai/code/artifact/53f16a0c-0074-44a3-8846-281b0faf5b77

- **Se il valore sopra è un URL valido:** mostra `📊 Diagramma di flusso: <URL>`.
- **Se il valore sopra è assente o è un placeholder** (Artifact non ancora pubblicato la prima volta): mostra `📊 Diagramma di flusso: non ancora pubblicato`.

---

# Webmapp Feature Workflow

Workflow obbligatorio prima che venga scritta qualsiasi riga di codice per feature o refactor non banali. Segui le fasi in ordine senza saltarne nessuna.

<HARD-GATE>
Nessun codice può essere scritto prima che `overview.md` e `plan.md` esistano nel filesystem e siano stati esplicitamente approvati dall'utente. Questo vale sempre, indipendentemente dalla semplicità percepita del task.
</HARD-GATE>

---

## Contratto artefatti

Questa sezione è la fonte autoritativa degli artefatti prodotti da `wm-plan`. `wm-skills:wm-review-ticket` la legge via WebFetch per conoscere la struttura senza duplicarla.

Per ogni feature lavorata con `wm-plan`, vengono creati i seguenti file nella cartella `docs/features/<feature-slug>/` del repo target:

| File | Contenuto | Usato da wm-review-ticket per |
|------|-----------|-------------------------------|
| `overview.md` | Cosa cambia, Perché, Requisiti, Rischi, Out of scope, Moduli toccati | Criterio principale di correttezza: il codice risponde ai Requisiti? |
| `plan.md` | Lista task implementativi step-by-step | Verifica completezza: tutti i task sono stati eseguiti? |
| `notes.md` | Deviazioni dal piano, bug trovati, decisioni on-the-fly | Contesto deviazioni: sono giustificate o rischiose? |

**Slug:** `<ID>-<titolo-in-kebab-case>` (es. `8068-wm-review-ticket`). Se non c'è ticket, solo `<titolo-in-kebab-case>`.

**Ricerca fuzzy per ID:**
```bash
find docs/features/ -maxdepth 1 -type d | grep "<ID>"
```

---

## Modalità tag-mode

`wm-plan` può essere invocato da `wm-skills:wm-tag` in **tag-mode**. In questo caso riceve nel contesto della conversazione:
- Titolo del ticket da creare
- Tipo (Feature / Bug / Task)
- Repo di destinazione (ricavabile da `~/.config/webmapp/repos.json`)
- ID del tag padre su Orchestrator
- Flag esplicito `tag-mode: true`

### tag-mode: flusso

In tag-mode, `wm-plan` esegue **solo** queste fasi nell'ordine:

1. `Fase: environment-setup` — rileva stack e repo (consulta `repos.json` per la path del repo di destinazione)
2. `Fase: init-context` — legge il CLAUDE.md del repo di destinazione
3. `Fase: reverse-interaction` — dialogo socratico completo (minimo 5 domande, nessuna eccezione)
4. `Fase: overview` — produce l'overview con la struttura canonica
5. `Fase: challenge` — analisi adversariale sull'overview
6. `Fase: estimation` — solo se Feature (stima in ore, approvata dal dev)
7. `Fase: ticket` — crea il ticket su Orchestrator **solo ora**, con tutti i campi disponibili: `name`, `type`, `customer_request` (derivato dalla trascrizione), `description` (overview completa), `estimated_hours`, `tags` (tag padre)
8. **Stop** — restituisce il controllo a `wm-tag`

Le fasi `write-plan`, `execution`, `notes`, `update-context` **non vengono eseguite**.

**Importante:** l'overview **non viene salvata nel filesystem** del repo. Nessun file `docs/features/` viene creato o modificato.

### tag-mode: creazione ticket con overview

Dopo l'approvazione dell'overview (e dell'estimation se Feature), crea il ticket in un'unica chiamata POST con tutti i campi disponibili. Il campo `description` è renderizzato da un editor WYSIWYG lato Orchestrator (nessun parsing Markdown): costruisci il contenuto in **HTML**, non in Markdown, convertendo la struttura canonica dell'overview 1:1 in tag HTML (`<h2>`/`<h3>` per le sezioni, `<p>` per il testo, `<ul><li>` per liste e requisiti):

```html
<h2>Overview</h2>

<h3>Cosa cambia</h3>
<p><testo></p>

<h3>Perché</h3>
<p><testo></p>

<h3>Requisiti</h3>
<ul>
  <li><requisito></li>
</ul>

<h3>Rischi</h3>
<p><testo></p>

<h3>Out of scope</h3>
<p><testo></p>

<h3>Moduli toccati</h3>
<p><testo></p>
```

Mostra preview e chiedi conferma (regola generale scritture):

> **Creazione ticket**
>
> | Campo | Valore |
> |-------|--------|
> | `name` | `<titolo>` |
> | `type` | `<tipo>` |
> | `customer_request` | `<prime 200 caratteri...>` |
> | `description` | `<h2>Overview</h2><h3>Cosa cambia</h3><p><prime 200 caratteri...>` |
> | `estimated_hours` | `<N>` |
> | `tags` | `[<tag-id>]` |
>
> Procedo?

Solo dopo la conferma:

```bash
ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-https://orchestrator.maphub.it}"
TOKEN=$(jq -r '.token' ~/.config/webmapp/orchestrator-auth.json)
curl -s -X POST "$ORCHESTRATOR_URL/api/stories" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '{"name": "<titolo>", "type": "<tipo>", "customer_request": "<testo>", "description": "<overview-html>", "estimated_hours": <N>, "tags": [<tag-id>]}'
```

Al termine mostra: `✅ Ticket oc:<ID> creato e associato al tag <nome-tag>.`

---

## Orchestrator API

Queste istruzioni valgono per tutte le chiamate HTTP a Orchestrator. Usale ogni volta che una fase richiede di leggere o scrivere un ticket.

### Configurazione

- **URL base:** leggi `$ORCHESTRATOR_URL` dall'environment. Se non è impostata usa `https://orchestrator.maphub.it` come default.
- **Auth:** salvato in `~/.config/webmapp/orchestrator-auth.json` (JSON con campi `token`, `id`, `name`, `email`). Se il file non esiste o la chiamata restituisce 401, esegui il login (vedi sotto). Se esiste solo il file legacy `~/.config/webmapp/orchestrator-token`, esegui la migrazione (vedi sotto).

### Login (solo se auth assente o scaduto)

Chiedi email e password all'utente, poi:

```bash
ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-https://orchestrator.maphub.it}"
TOKEN=$(curl -s -X POST "$ORCHESTRATOR_URL/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"<email>","password":"<password>"}' \
  | jq -r '.token')
USER=$(curl -s -X GET "$ORCHESTRATOR_URL/api/me" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Accept: application/json")
mkdir -p ~/.config/webmapp
echo $USER | jq --arg token "$TOKEN" '. + {token: $token}' > ~/.config/webmapp/orchestrator-auth.json
```

### Migrazione da file legacy (solo se `orchestrator-auth.json` assente ma `orchestrator-token` presente)

```bash
ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-https://orchestrator.maphub.it}"
TOKEN=$(cat ~/.config/webmapp/orchestrator-token)
USER=$(curl -s -X GET "$ORCHESTRATOR_URL/api/me" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Accept: application/json")
echo $USER | jq --arg token "$TOKEN" '. + {token: $token}' > ~/.config/webmapp/orchestrator-auth.json
```

Se `GET /api/me` risponde 401 durante la migrazione, il token legacy è scaduto: cancella `orchestrator-token` ed esegui il login completo sopra.

### Lettura ticket (GET — nessuna conferma richiesta)

```bash
ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-https://orchestrator.maphub.it}"
TOKEN=$(jq -r '.token' ~/.config/webmapp/orchestrator-auth.json)
curl -s -X GET "$ORCHESTRATOR_URL/api/stories/<ID>" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Accept: application/json"
```

Se risponde 401: cancella il file auth e ripeti il login prima di ritentare.

### Creazione ticket (POST — richiede conferma esplicita)

Prima di eseguire, mostra un riepilogo tabellare e chiedi conferma esplicita:

> **Creazione ticket**
>
> | Campo | Valore |
> |-------|--------|
> | `<campo>` | `<valore>` |
>
> Procedo?

Solo dopo la conferma:

```bash
ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-https://orchestrator.maphub.it}"
TOKEN=$(jq -r '.token' ~/.config/webmapp/orchestrator-auth.json)
curl -s -X POST "$ORCHESTRATOR_URL/api/stories" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '<json payload>'
```

Salva l'ID restituito (`$.id`) come `<ID>` del ticket per il resto del workflow.

### Aggiornamento ticket (PATCH — richiede conferma esplicita)

Prima di eseguire, mostra sempre un riepilogo tabellare:

> **Aggiornamento ticket oc:\<ID\>**
>
> | Campo | Valore |
> |-------|--------|
> | `<campo>` | `<valore>` |
> | `<campo>` | `<valore>` |
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
- **A** → segui `### ticket: caso-a`
- **B** → segui `### ticket: caso-b`
- **C** → segui `### ticket: caso-c`

### ticket: caso-a

Se l'utente scrive `oc:<ID>` (con o senza contenuto aggiuntivo), leggi il ticket via API seguendo `## Orchestrator API → Lettura ticket`:

```bash
ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-https://orchestrator.maphub.it}"
TOKEN=$(jq -r '.token' ~/.config/webmapp/orchestrator-auth.json 2>/dev/null)
curl -s -X GET "$ORCHESTRATOR_URL/api/stories/<ID>" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Accept: application/json"
```

Se il file auth non esiste, controlla se esiste `~/.config/webmapp/orchestrator-token` ed esegui la migrazione (vedi `## Orchestrator API → Migrazione da file legacy`). Se la risposta è 401, cancella il file auth ed esegui il login completo.

Dal JSON restituito estrai:
- `name` → titolo, usato per il `<feature-slug>`: `<ID>-<titolo-in-kebab-case>`
- `customer_request` → contesto del problema, usato in Fase: reverse-interaction e Fase: overview
- `description` → note tecniche già raccolte, può orientare le domande in Fase: reverse-interaction
- `type` → orienta il tono dell'overview

Se `type == "Help desk"`, esegui `### ticket: caso-a-split-detection` prima di mostrare qualunque riepilogo. Altrimenti procedi direttamente al riepilogo standard sotto.

Mostra all'utente un riepilogo del ticket letto prima di procedere alla Fase: init-context.

### ticket: caso-a-split-detection

Analizza `customer_request` e determina, con giudizio diretto sul testo (nessuna euristica meccanica su keyword o conteggio paragrafi), se contiene **più richieste distinte** scritte dal cliente in un unico ticket.

- **Se rilevi una sola richiesta:** non fare nulla, prosegui al riepilogo standard di `caso-a` come oggi.
- **Se rilevi N ≥ 2 richieste distinte:** estrai N partizioni di testo **verbatim** (nessuna riformulazione), in ordine di apparizione nel testo originale. Genera anche un titolo sintetico per ciascuna partizione. Poi mostra:

  > **Rilevate {N} richieste distinte nel ticket oc:{ID} ("{name originale}"):**
  >
  > | # | Ruolo | Titolo proposto | Testo (verbatim) |
  > |---|---|---|---|
  > | 1 | Ticket originale (rinominato) | `{titolo sintetico 1}` | "{partizione 1}" |
  > | 2 | Nuovo ticket | `{titolo sintetico 2}` | "{partizione 2}" |
  > | ... | | | |
  >
  > Procedo con lo split? (puoi modificare titoli o testo prima di confermare)

  Attendi conferma esplicita.

  - **Se l'utente rifiuta:** il ticket originale resta intatto, prosegui al riepilogo standard di `caso-a` come oggi (nessuno split).
  - **Se l'utente conferma (con eventuali modifiche a titoli/testo):** procedi con `### ticket: caso-a-split-execution`.

### ticket: caso-a-split-execution

Esegui le scritture seguendo comunque `## Orchestrator API → Regola generale scritture` (preview tabellare + conferma esplicita per ogni singola chiamata, anche se il contenuto è già stato approvato nel riepilogo di `caso-a-split-detection`):

1. **PATCH sul ticket originale `oc:{ID}`:**

```bash
ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-https://orchestrator.maphub.it}"
TOKEN=$(jq -r '.token' ~/.config/webmapp/orchestrator-auth.json)
curl -s -X PATCH "$ORCHESTRATOR_URL/api/stories/<ID>" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '{"name": "<titolo sintetico 1>", "customer_request": "<partizione 1 verbatim>"}'
```

Tag e `creator_id` del ticket originale non vengono toccati da questa PATCH.

2. **POST per ciascuna partizione 2..N:**

```bash
ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-https://orchestrator.maphub.it}"
TOKEN=$(jq -r '.token' ~/.config/webmapp/orchestrator-auth.json)
curl -s -X POST "$ORCHESTRATOR_URL/api/stories" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '{"name": "<titolo sintetico N>", "type": "Help desk", "customer_request": "<partizione N verbatim>", "creator_id": <creator_id originale>, "tags": [<id tag originali>]}'
```

`type` è sempre `"Help desk"` — nessuna riclassificazione automatica in questa fase. Salva l'`id` restituito per ogni ticket creato. Mostra `✅ Ticket oc:<nuovo-ID> creato.` per ciascuno.

3. **Tag di raggruppamento opzionale (uso interno dev):**

Chiedi:

> "Vuoi raggruppare questi {N} ticket in un tag?"

- **Se sì:**
  - Cerca tra i tag del ticket originale uno riconoscibile come identificativo cliente (es. `ass_cammini_italia`). Se trovato, proponi nome default `<tag-cliente>-<titolo originale kebab-case>`.
  - Se nessun tag è riconoscibile come cliente, chiedi al dev di indicare manualmente il nome cliente da usare.
  - Il dev può modificare il nome proposto prima della creazione.
  - Crea il tag (stesso endpoint di `wm-skills:wm-tag`):

```bash
ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-https://orchestrator.maphub.it}"
TOKEN=$(jq -r '.token' ~/.config/webmapp/orchestrator-auth.json)
curl -s -X POST "$ORCHESTRATOR_URL/api/tags" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '{"name": "<nome-tag>", "description": "<customer_request originale completo, pre-split>"}'
```

  - Associa il tag restituito a tutti i ticket del gruppo (originale + nuovi) tramite PATCH `tags` su ciascuno (preview + conferma per ciascuna PATCH, come da regola generale).
- **Se no:** salta questo step, nessun tag creato.

4. **Selezione ticket su cui continuare:**

Mostra l'elenco di tutti i ticket del gruppo:

> **Ticket generati dallo split di oc:{ID originale}:**
>
> | Ticket | Titolo |
> |---|---|
> | oc:{ID originale} | {titolo 1} |
> | oc:{nuovo ID 2} | {titolo 2} |
> | ... | |
>
> Su quale vuoi continuare il workflow ora? (oppure "nessuno" per tornare al menu)

- **Se l'utente sceglie un ticket:** prosegui il workflow in `Fase: init-context` usando i dati già noti di quel ticket (senza rifare la GET).
- **Se l'utente risponde "nessuno":** torna al menu A/B/C di `Fase: ticket`.

### ticket: progress

Dopo il riepilogo, chiedi:

> "Vuoi impostare lo status del ticket a **progress** e assegnartelo?"

Se l'utente risponde sì, esegui il PATCH seguendo `## Orchestrator API → Aggiornamento ticket` con:

```bash
ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-https://orchestrator.maphub.it}"
TOKEN=$(jq -r '.token' ~/.config/webmapp/orchestrator-auth.json)
USER_ID=$(jq -r '.id' ~/.config/webmapp/orchestrator-auth.json)
curl -s -X PATCH "$ORCHESTRATOR_URL/api/stories/<ID>" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d "{\"status\": \"progress\", \"user_id\": $USER_ID}"
```

Se il PATCH fallisce (risposta non 2xx o errore di rete), avvisa l'utente con un messaggio ("⚠️ Impossibile aggiornare lo status del ticket — procedo comunque con il workflow.") e continua.

Se l'utente risponde no, procedi direttamente alla Fase: init-context senza modificare il ticket.

### ticket: caso-b

Chiedi all'utente una descrizione della feature (anche breve). Con quella, proponi subito il testo del ticket:

```
name: <titolo sintetico della feature>

type: Feature

customer_request:
<descrizione del problema in linguaggio non tecnico, 3-5 righe>

description:
<approccio tecnico iniziale, da raffinare dopo le fasi successive>
```

Chiedi all'utente di confermarlo o modificarlo. Una volta approvato, crea il ticket via API seguendo `## Orchestrator API → Creazione ticket`. Il campo `description` è renderizzato da un editor WYSIWYG (nessun parsing Markdown): se il testo approvato usa formattazione (titoli, liste, grassetto), convertila in tag HTML equivalenti (`<h3>`, `<ul><li>`, `<strong>`) prima di inviarla nel payload — non inviare Markdown grezzo. Salva l'ID restituito e usalo come `<ID>` per tutto il resto del workflow.

**Il ticket va creato prima di procedere alla Fase: init-context.** I campi `description` e `customer_request` potranno essere aggiornati a fine workflow (Checklist) con le informazioni emerse dalle fasi successive.

Una volta creato il ticket e salvato l'ID, chiedi:

> "Vuoi impostare lo status del ticket a **progress** e assegnartelo?"

Se l'utente risponde sì, esegui il PATCH seguendo `## Orchestrator API → Aggiornamento ticket` con:

```bash
ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-https://orchestrator.maphub.it}"
TOKEN=$(jq -r '.token' ~/.config/webmapp/orchestrator-auth.json)
USER_ID=$(jq -r '.id' ~/.config/webmapp/orchestrator-auth.json)
curl -s -X PATCH "$ORCHESTRATOR_URL/api/stories/<ID>" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d "{\"status\": \"progress\", \"user_id\": $USER_ID}"
```

Se il PATCH fallisce, avvisa l'utente con un messaggio ("⚠️ Impossibile aggiornare lo status del ticket — procedo comunque con il workflow.") e continua.

Se l'utente non vuole creare il ticket ora, procedi senza ID: usa solo il titolo kebab-case come slug. La domanda progress non viene posta.

### ticket: caso-c

L'utente ha una trascrizione Meet, un brief cliente o qualsiasi materiale da cui estrarre più ticket raggruppati in un tag Orchestrator.

Invoca immediatamente `wm-skills:wm-tag`, passando come contesto qualsiasi testo o link già fornito dall'utente in questa sessione. Da questo momento il controllo del flusso passa interamente a `wm-tag` — non proseguire con nessun'altra fase di `wm-plan`.

### ticket: aggiornamenti-espliciti

Se in qualsiasi momento l'utente chiede di aggiornare un campo del ticket (es. "aggiorna lo status a progress", "scrivi nelle note dev che…"), esegui un PATCH seguendo `## Orchestrator API → Aggiornamento ticket`. Mostra sempre il preview della modifica e attendi conferma prima di inviare.

### ticket: estrazione

- `<feature-slug>`: `<ID>-<titolo-in-kebab-case>` (con ticket) o `<titolo-in-kebab-case>` (senza)
- **Riferimento ticket in tutti i documenti:** ogni file creato nelle fasi successive deve riportare in testa `> Ticket: oc:<ID>` (ometti se non c'è ID).
- **Commit convention:** tutti i commit usano `oc:<ID>` come scope (es. `feat(oc:7815): add OSM POI import action`). Senza ticket, usa il titolo kebab-case come scope.

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

**Livello 1 — Stack UI rilevato in Fase: environment-setup + file coinvolti**

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
      Procedo con giudizio interno UX — installa la skill per i prossimi ticket."
      Applica il tuo giudizio interno su UX per i Requisiti e Rischi dell'overview.

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

Leggi il file `CLAUDE.md` nella root del progetto target.

- Se non esiste, segnalalo all'utente e procedi con le informazioni disponibili.
- Estrai: stack tecnologico, convenzioni di test, struttura cartelle, istruzioni specifiche al team.
- Tieni queste informazioni attive per tutto il workflow.

> I flag di stack (`stack_ui`, `stack_type`, `has_docker`, `has_submodules`) sono già stati impostati in Fase: environment-setup.

---

## Fase: reverse-interaction (obbligatoria, non skippabile)

> I flag `stack_ui`, `stack_type`, `has_docker`, `has_submodules`, la classificazione domain-mapping e la valutazione UX sono già stati impostati in Fase: environment-setup.

### reverse-interaction: dialog

Conduci un dialogo socratico con l'utente: **una domanda alla volta**, aspetta la risposta, poi formula la successiva tenendo conto di ciò che hai appena sentito. Non presentare mai più domande in un unico messaggio.

**Regole:**
- Una domanda per messaggio. Sempre.
- Minimo 5 domande totali. Puoi farne di più se le risposte aprono nuovi buchi di contesto.
- Le `Note di sviluppo` del ticket possono orientare le domande, non ridurne il numero minimo.
- L'unica eccezione per scendere sotto 5 è una giustificazione esplicita scritta nel tuo messaggio (es. "Il ticket e le note di sviluppo coprono già questi aspetti — le 5 domande sarebbero ridondanti perché…").
- Ogni domanda deve essere costruita sulla risposta precedente, non preparata in anticipo.
- Procedi alla Fase: overview solo dopo aver fatto almeno 5 domande e ricevuto tutte le risposte.
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

## Fase: overview

**I file di documentazione seguono il codice, non il repo principale.**

Per ogni repo coinvolto dalla feature (principale + eventuali submodule) crea un `overview.md` separato nella cartella `docs/features/<feature-slug>/` di quel repo. Non accentrare tutto nel repo principale se il codice è distribuito.

| Dove va il codice | Dove va la documentazione |
|---|---|
| Repo principale | `docs/features/<feature-slug>/overview.md` nel repo principale |
| Submodule `wm-package` | `docs/features/<feature-slug>/overview.md` in `wm-package/` |
| Submodule `wm-core` | `docs/features/<feature-slug>/overview.md` in `wm-core/` |

Se la feature è interamente custom (nessun submodule coinvolto), un solo `overview.md` nel repo principale è sufficiente.

Il `<feature-slug>` è `<ID>-<titolo-in-kebab-case>` (es. `7815-creazione-poi-tramite-osm-id`). Se non c'è ticket, usa solo il titolo kebab-case. Lo slug è lo stesso in tutti i repo per mantenere la tracciabilità.

**La documentazione va creata sempre, anche senza ticket.** L'assenza di un ID Orchestrator non è un motivo per saltare `overview.md`, `plan.md` o `notes.md`. Ometti solo la riga `> Ticket: oc:<ID>` nell'header.

**Struttura obbligatoria:**

```markdown
> Ticket: oc:<ID>   ← ometti se non c'è ticket

# <Titolo dal ticket>

## Cosa cambia
[Descrizione concisa di cosa il sistema farà di diverso dopo questa feature]

## Perché
[Motivazione business/tecnica emersa dal ticket e dalla Fase: reverse-interaction]

## Requisiti
- [ ] Requisito funzionale 1
- [ ] Requisito funzionale 2
...

## Rischi
[Criticità emerse dalla Fase: challenge con indicazione di come vengono mitigate]

## Out of scope
[Cosa esplicitamente NON viene fatto in questo ciclo]

## Moduli toccati
[Lista di file, moduli o servizi che vengono modificati o creati]
```

Mostra il file all'utente e attendi approvazione esplicita prima di procedere.

---

## Fase: challenge

La Challenge viene eseguita **dopo** la scrittura di `overview.md` e **prima** di scrivere `plan.md`. Questo garantisce che l'analisi avvenga su un documento concreto e che eventuali buchi trovati possano essere corretti nell'overview prima di pianificare l'implementazione.

### challenge: subagent

Lancia un subagente con questo prompt — e **nient'altro**:

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

**Non aggiungere al prompt nessun altro contesto, riassunto o spiegazione della conversazione precedente. Solo il percorso del file e le istruzioni sopra. Il subagente deve leggere `overview.md` autonomamente dal filesystem.**

### challenge: dialog

Ricevuto il report del subagente, presentalo all'utente e affronta gli assi uno alla volta in ordine di criticità (dal più critico al meno). Per ogni asse:
- Riassumi il rischio in una riga
- Proponi come intendi gestirlo
- Chiedi all'utente se vuole modificare l'approccio su quel punto

Aspetta la risposta prima di passare all'asse successivo.

### challenge: overview-update

Se dalla Challenge emergono buchi che cambiano requisiti, scope o approccio, aggiorna `overview.md` prima di procedere alla Fase: write-plan. Mostra le modifiche all'utente e attendi approvazione esplicita.

---

## Fase: estimation

**Eseguita solo se il ticket è di tipo Feature.** Per Bug, Task o altri tipi, salta questa fase e procedi direttamente a `Fase: write-plan`.

In tag-mode, questa fase viene eseguita prima di fermarsi (non si procede a write-plan).

### estimation: analisi

Basandoti sull'overview approvato e sull'esito della Fase: challenge, classifica ogni componente della stima in una di due categorie:

- **Scrittura pura** — zero domande aperte residue su "come deve comportarsi" dopo overview + challenge (specifica già completa: campi, endpoint, comportamento, edge case). Buffer 0%.
- **Decisioni aperte** — restano scelte UX/comportamentali da prendere, o reverse-engineering di comportamento legacy non documentato. Buffer 20-30%.

**Calcola il tempo di pianificazione misurato:**

```bash
NOW=$(date -u +"%Y-%m-%dT%H:%M:%S%z")
```

Confronta `NOW` con `planning_start_at` (registrato in Fase: ticket) e calcola la differenza in ore — questo è un dato misurato, non stimato.

**Se `planning_start_at` non è stato registrato** (es. la Fase: ticket è stata eseguita senza eseguire il comando `date -u` previsto), non calcolare una quota "misurata" fittizia e non presentarla come se il dato fosse disponibile. Segnala esplicitamente all'utente:

> "⚠️ Non ho registrato il timestamp di inizio pianificazione in Fase: ticket — la quota 'Misurato' non è disponibile per questa stima. Propongo solo la quota 'Stimato', senza il confronto Misurato + Stimato = Totale."

Poi procedi con la tabella sottostante omettendo la riga "Pianificazione" e il calcolo `<M>h`.

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

### estimation: conferma

Chiedi al dev, riportando sempre la scomposizione (mai un numero unico fuso):

> "Accetti questa stima — **Misurato: \<M\>h + Stimato: \<S\>h = Totale: \<N\>h** — o vuoi modificarla?"

Aspetta risposta esplicita. Se il dev propone un valore diverso, usalo senza discutere — la stima finale è sempre quella approvata dal dev. Se il dev modifica solo la quota stimata (implementazione) lasciando invariata quella misurata (pianificazione), aggiorna solo `<S>` e ricalcola `<N>`.

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

---

## Fase: write-plan

**Prima di invocare writing-plans**, leggi esplicitamente tutti gli `overview.md` generati in Fase: overview (uno per ogni repo coinvolto) e costruisci un briefing strutturato da passare come contesto a writing-plans. Il briefing deve contenere:

- Ticket: `oc:<ID>` e titolo
- Repo coinvolti e classificazione (custom / package / misto)
- Requisiti dalla sezione "Requisiti" di ogni overview.md
- Rischi e decisioni emerse dalla Fase: challenge e recepite nell'overview
- File da creare/modificare per repo, dalla sezione "Moduli toccati"
- Vincoli tecnici emersi dalla Fase: reverse-interaction

Questo briefing è lo spec che writing-plans usa per generare il piano — senza di esso writing-plans parte cieco.

**REQUIRED SUB-SKILL:** Invoca `superpowers:writing-plans` con il briefing sopra come contesto, e questi override:

- **Percorso di salvataggio:** `docs/features/<feature-slug>/plan.md`
- **Header obbligatorio:** il piano inizia con `> Ticket: oc:<ID>`
- **Commit convention:** tutti i commit usano `feat(oc:<ID>): ...` / `fix(oc:<ID>): ...` / `refactor(oc:<ID>): ...`
- **⚠️ No commit o branch automatici:** i commit nel piano sono istruzioni testuali per l'utente, non azioni da eseguire autonomamente. Claude non esegue `git commit`, `git push` o crea branch senza conferma esplicita dell'utente per ogni singolo commit.

Durante la scrittura del piano applica la skill `wm-skills:our-code-style` per allineare le scelte implementative alle convenzioni Webmapp.

Mostra il piano all'utente e attendi approvazione esplicita prima di procedere.

---

## Fase: execution

### execution: design (se applicabile)

Se il piano include componenti UI/UX (nuove interfacce, layout, prototipi, slide, one-pager), prima di eseguire il codice proponi all'utente di usare **Claude Design** (`claude.ai/design`):

> "Questa feature ha componenti visual. Ti consiglio di prototipare il design su claude.ai/design prima di implementare — puoi poi trasferire il risultato direttamente a Claude Code con un'istruzione."

Aspetta che l'utente confermi di aver completato la fase di design (o decida di saltarla) prima di procedere con 6b (creazione branch).

### execution: branch (obbligatoria, prima di scrivere qualsiasi file)

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

### execution: implementation

**Regola traduzioni (obbligatoria):** ogni testo traducibile introdotto dall'implementazione deve:
- avere il testo base nella **lingua di default del repo** (rilevata in Fase: reverse-interaction — solitamente inglese, ma verifica)
- avere una traduzione in **tutte le lingue presenti nel repo** (file di lingua rilevati in Fase: reverse-interaction)
- non lasciare chiavi mancanti in nessun file di lingua esistente

Se in Fase: reverse-interaction non è stata rilevata la configurazione i18n (feature non aveva testi UI), ispezionala ora prima di scrivere qualsiasi stringa traducibile.

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

### execution: re-estimation (solo se il ticket è di tipo Feature)

Se durante l'implementazione emerge un problema non previsto nell'overview e nella stima originale, e questo problema è stimabile in ore (non un semplice imprevisto trascurabile), proponi al dev una revisione della stima prima di proseguire:

> "Ho trovato \<descrizione problema non previsto\>. Stimo un impatto aggiuntivo di **\<X\>h**, portando il totale da \<N\> a \<N+X\>h. Vuoi che aggiorni la stima su Orchestrator?"

Se il dev conferma, applica `## Orchestrator API → Aggiornamento ticket` (preview + conferma esplicita) per il PATCH `estimated_hours` con il nuovo totale, mantenendo lo stesso marcatore di versione (`[stima v2 — per-componente]`) usato in `estimation: scrittura su Orchestrator`.

Registra sempre l'evento in `Fase: notes` (sezione "Decisioni"), indipendentemente dal fatto che il dev abbia accettato o rifiutato la revisione.

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

### execution: formal-review

Se l'utente risponde **sì** alla proposta di review formale in `execution: review-gate`:

1. Invoca `wm-skills:wm-review-ticket oc:<ID>`
2. Attendi il completamento della review
3. Se emergono correzioni → applicale prima di procedere ai commit
4. Torna al punto 6 di `execution: review-gate` (esegui i commit)

---

## Fase: notes

Crea e aggiorna `docs/features/<feature-slug>/notes.md` durante e dopo l'esecuzione.

**Regole:**
- Il file deve esistere al termine del workflow. Un notes.md con "Nessuna deviazione rilevante" è valido. Un notes.md assente non lo è.
- Registra: deviazioni dal piano, bug trovati durante l'implementazione, decisioni prese on-the-fly, follow-up da fare in cicli successivi.
- **Modifiche richieste a posteriori** (dopo l'approvazione del piano ma prima del commit): registrale nella sezione "Decisioni" con una riga che descrive cosa è cambiato e perché — anche se la modifica è stata recepita nel codice, la traccia in notes serve per capire perché il piano è stato superato.
- **Falsi negativi di classificazione stima** (solo per ticket Feature): se un componente classificato "scrittura pura" in `Fase: estimation` si rivela durante l'esecuzione una "decisione aperta" (richiede scelte UX/comportamentali non previste), registralo esplicitamente in una riga della sezione "Follow-up" o "Decisioni" — questo dato è necessario per calibrare il criterio di classificazione nei cicli successivi.

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

## Fase: update-context

Aggiorna il `CLAUDE.md` nella root del progetto target con le informazioni prodotte dal workflow. Se il file non esiste, crealo.

**Sezione `## Feature disponibili`** — aggiorna o crea con una riga per ogni feature completata:

```markdown
## Feature disponibili

| Feature | Ticket | Moduli toccati | Note |
|---|---|---|---|
| <Titolo feature> | oc:<ID> | `path/modulo1`, `path/modulo2` | <una riga: cosa fa> |
```

Se la sezione esiste già, aggiungi la nuova riga senza toccare quelle precedenti.

**Sezione `## Decisioni architetturali`** — aggiungi le scelte non ovvie emerse dalla Fase: challenge e dalla Fase: notes che un futuro Claude dovrebbe conoscere per non ripercorrere gli stessi ragionamenti:

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

- **`ui-ux-pro-max`** — invocata automaticamente in environment-setup: ux-ui-detection quando rilevati componenti UI/UX (Vue, Angular, HTML/CSS). Richiede `/plugin install ui-ux-pro-max@wm-marketplace` se non installata.
- **`wm-skills:our-code-style`** — applica in Fase: write-plan e Fase: execution
- **`wm-skills:our-pr-checklist`** — applica dopo la Fase: notes, prima di aprire la PR
- **`wm-skills:our-deploy-post-merge`** — applica dopo il merge della PR

---

## Checklist di completamento

Prima di dichiarare il workflow concluso, verifica che esistano tutti e tre i file:

- [ ] `docs/features/<feature-slug>/overview.md` — approvato dall'utente (riferimento ticket presente se applicabile)
- [ ] `docs/features/<feature-slug>/plan.md` — approvato dall'utente (riferimento ticket presente se applicabile)
- [ ] `docs/features/<feature-slug>/notes.md` — compilato (anche solo con "Nessuna deviazione") (riferimento ticket presente se applicabile)

**Questi tre file sono obbligatori sempre, con o senza ticket Orchestrator.**
- [ ] `CLAUDE.md` del progetto target aggiornato — sezione "Feature disponibili" e "Decisioni architetturali"
- [ ] Artifact del diagramma di flusso `wm-plan` rigenerato (redeploy stesso URL) se questa sessione ha modificato file del repo `claude-marketplace` — vedi `CLAUDE.md` → `## Diagramma di flusso wm-plan`

### update-context: orchestrator (solo se esiste un ticket oc:\<ID\>)

- [ ] Leggi lo status attuale del ticket via `## Orchestrator API → Lettura ticket`
- [ ] Leggi gli status disponibili da `StoryStatus.php` su GitHub (vedi `## Orchestrator API → Status disponibili`)
- [ ] Suggerisci lo status più appropriato al contesto (es. `testing` se ci sono test da verificare, `done` se tutto è completato e i test passano) e presenta la lista completa — aspetta la scelta esplicita dell'utente
- [ ] Prepara la bozza di `description` (note dev) con:
  - Link cliccabile HTML alla cartella `docs/features/<feature-slug>/` — il campo è interpretato come HTML, usa:
    ```
    <a href="https://github.com/<owner>/<repo>/tree/main/docs/features/<feature-slug>/">docs/features/<feature-slug>/</a>
    ```
    Ricava `<owner>/<repo>` eseguendo `git remote get-url origin` e normalizzando l'URL (rimuovi `.git` finale, gestisci sia formato HTTPS che SSH).
  - Riepilogo tecnico di cosa è stato implementato (file creati/modificati, approccio usato)
  - Tono tecnico, rivolto al team
- [ ] Prepara la bozza del messaggio di risposta cliente con:
  - Descrizione in linguaggio non tecnico di cosa è stato fatto e perché
  - Niente nomi di file, classi, branch o dettagli implementativi
  - Tono chiaro e orientato al beneficio per l'utente finale
- [ ] Mostra entrambe le bozze all'utente e chiedi approvazione esplicita — la risposta cliente è letta dal cliente, richiede revisione attenta
- [ ] Solo dopo approvazione esplicita, esegui il PATCH seguendo `## Orchestrator API → Aggiornamento ticket` con i campi: `status`, `description`, `customer_request`

  **Importante:** manda solo il testo pulito nei campi `customer_request` e `description` — il backend chiama internamente `addResponse()` e `addDevNote()` che gestiscono formato HTML, timestamp, prepend e notifiche. Non costruire HTML manualmente.
