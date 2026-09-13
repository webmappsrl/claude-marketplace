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

**Versione installata:** v1.3.0

Questo valore è statico, scritto direttamente in questa skill (stesso pattern dell'URL del diagramma in `### header: diagramma`): si aggiorna manualmente ad ogni release, come da checklist in `docs/howto/rilascio-wm-skills.md` del repo `claude-marketplace`. Non richiede alcuna risoluzione di path a runtime (niente ricerca nella cache dei plugin né `git`), quindi mostra sempre il dato senza rischio di "check non disponibile".

Mostra `Versione installata: v1.3.0` come prima riga di questa sotto-sezione, poi prosegui con il check di aggiornamento disponibile:

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

L'URL è quello della pagina pubblicata su GitHub Pages dal repo `claude-marketplace`. È
stabile e non cambia mai: la pagina si aggiorna con il push che modifica il sorgente, quindi
non c'è nessun redeploy da fare né nessun URL da riscrivere qui.

**URL:** https://webmappsrl.github.io/claude-marketplace/wm-plan-diagramma/

Mostra `📊 Diagramma di flusso: <URL>`.

### header: context

Mostra l'occupazione del context in **valore assoluto**, come informazione. Non attiva e non
disattiva nulla.

```bash
TRANSCRIPT=$(ls -t ~/.claude/projects/"$(pwd | tr '/' '-')"/*.jsonl 2>/dev/null | head -1)
grep -v '"isSidechain":true' "$TRANSCRIPT" 2>/dev/null | grep -o '"usage":{[^}]*}' | tail -1
```

Il transcript più recente della directory di progetto potrebbe appartenere a un'altra
sessione (più istanze aperte sullo stesso repo): per questo la misura resta puramente
indicativa, coerente col fatto che è solo informativa.

Somma `input_tokens`, `cache_creation_input_tokens` e `cache_read_input_tokens`, e mostra:

`🧮 Context occupato: ~<N>k token`

- **Mai una percentuale:** la dimensione della finestra non è leggibile dal transcript e
  andrebbe indovinata (200k o 1M, e il modello può cambiare a metà sessione).
- **Mai un numero inventato:** se il transcript non è raggiungibile o il formato è inatteso,
  mostra `⚠️ Misura del context non disponibile.` e prosegui.
- Il filtro `isSidechain` esclude le righe dei subagenti che condividono il transcript.

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
- Tipo (uno dei valori ammessi per `type` in `create_story`/`update_story`)
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
7. `Fase: ticket` — crea il ticket su Orchestrator **solo ora**, con tutti i campi disponibili: `name`, `type`, `customer_request` (derivato dalla trascrizione), `description` (overview completa), `estimated_hours`; il tag padre si associa **dopo**, con `attach_story_to_tag`, mai passando `tags` a `create_story`
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

Chiama `create_story` con `name`, `type`, `customer_request`, `description` (l'overview in HTML) ed `estimated_hours`, prima senza `confirm` per mostrare al dev l'anteprima calcolata dal tool, poi con `confirm: true` dopo l'approvazione esplicita.

Salvato l'`id` del ticket restituito, associalo al tag padre con `attach_story_to_tag` (`tag_id`, `story_id`) — mai passando `tags` a `create_story`/`update_story`, che sostituirebbe l'intero elenco — di nuovo prima senza `confirm` poi con `confirm: true`.

Al termine mostra: `✅ Ticket oc:<ID> creato e associato al tag <nome-tag>.`

---

## Orchestrator

Le operazioni su Orchestrator si fanno con i tool del server `orchestrator`, distribuito con questo plugin. Non costruire chiamate HTTP a mano.

| Operazione | Tool |
|---|---|
| Leggere un ticket | `get_story` |
| Creare un ticket | `create_story` |
| Modificare un ticket | `update_story` |
| Utente corrente | `me` |
| Tag: elenco, lettura, creazione, modifica | `list_tags`, `get_tag`, `create_tag`, `update_tag` |
| Associare o togliere un ticket da un tag | `attach_story_to_tag`, `detach_story_from_tag` |

**Regola sulle scritture.** I tool di scrittura accettano `confirm`. Chiamali **sempre prima senza `confirm`**: restituiscono la differenza rispetto allo stato attuale senza scrivere nulla. Mostra quella differenza al dev, attendi un'approvazione esplicita, e solo allora richiama lo stesso tool con `confirm: true`. Non costruire tu la tabella dell'anteprima: quella del tool è calcolata sui dati veri.

**Tipi e stati.** Non scrivere valori a memoria: il tool rifiuta i valori fuori elenco e ti dice quali sono ammessi.

**Formato dei campi.** `description` e `customer_request` di un ticket sono resi da un editor visuale: vanno scritti in HTML, non in Markdown. La `description` di un tag è invece in Markdown.

**Associazione a un tag.** Usa `attach_story_to_tag`, mai il campo `tags` di `update_story`: quel campo sostituisce l'elenco completo e cancella i tag già presenti.

**Credenziali.** Il server legge `~/.config/webmapp/orchestrator-auth.json`. Se un tool segnala credenziali assenti o scadute, guida il dev a rifare l'accesso; il server non lo fa da sé e non chiede mai la password.

**Se i tool non sono disponibili** (server non avviato o in errore), segnalalo al dev, poi leggi `${CLAUDE_PLUGIN_ROOT}/shared/orchestrator-fallback.md` e segui quelle istruzioni per questa sessione. Non ricostruire le chiamate a memoria: quel file è l'unica forma ammessa di ripiego.

---

## Delega ad agenti

Il meccanismo — tipi di delega, contratto di ritorno, tetto all'output, fallback, misura del
context — è descritto in `${CLAUDE_PLUGIN_ROOT}/shared/agent-delegation.md`. Leggerlo alla prima
delega della sessione, non riscriverne il contenuto qui.

Fasi che delegano, e a chi:

| Fase | Agente | Tipo |
|---|---|---|
| `environment-setup` | `wm-env-detect` | informata |
| `reverse-interaction` | `wm-codebase-research` | informata |
| `reverse-interaction` | `wm-transcript-research` | informata |
| `challenge` | subagente già previsto | **cieca** |
| `estimation` | `wm-estimate` | **cieca** |
| `review-gate` | subagente già previsto | **cieca** |
| `update-context` | `wm-context-guard` | informata |

Le deleghe sono **statiche**: si applicano sempre in quelle fasi. Nessuna soglia, nessuna
attivazione basata sulla misura del context.

Restano nel context principale, per scelta motivata in
`${CLAUDE_PLUGIN_ROOT}/shared/agentic-feasibility.md`: `init-context`, `overview`, `write-plan`,
`notes`, `review-gate: phpstan-check`, `environment-setup: docker-check` e ogni dialogo con
il dev.

`wm-context-doctor` non fa parte di questo workflow: il dev lo invoca esplicitamente quando
vuole una revisione d'insieme del `CLAUDE.md` di un repo (contraddizioni accumulate, voci
obsolete, sezioni cresciute troppo). Propone un piano da approvare, senza modificare nulla.

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

Se l'utente scrive `oc:<ID>` (con o senza contenuto aggiuntivo), leggi il ticket chiamando `get_story` con `story_id: <ID>`.

Se il tool segnala credenziali assenti o scadute, guida il dev a rifare l'accesso (vedi `## Orchestrator → Credenziali`).

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

Esegui le scritture seguendo comunque la regola generale sulle scritture del server `orchestrator` (anteprima senza `confirm` + conferma esplicita per ogni singola chiamata, anche se il contenuto è già stato approvato nel riepilogo di `caso-a-split-detection`):

1. **Aggiorna il ticket originale `oc:{ID}`:** chiama `update_story` con `story_id: {ID}`, `name: "<titolo sintetico 1>"`, `customer_request: "<partizione 1 verbatim>"`, prima senza `confirm` per l'anteprima, poi con `confirm: true`.

Tag e `creator_id` del ticket originale non vengono toccati da questa chiamata.

2. **Crea un ticket per ciascuna partizione 2..N:** chiama `create_story` con `name: "<titolo sintetico N>"`, `type: "Help desk"`, `customer_request: "<partizione N verbatim>"`, `creator_id: <creator_id originale>`, prima senza `confirm` poi con `confirm: true`. Poi associa ciascuno ai tag del ticket originale con `attach_story_to_tag` (uno per tag).

`type` è sempre `"Help desk"` — nessuna riclassificazione automatica in questa fase. Salva l'`id` restituito per ogni ticket creato. Mostra `✅ Ticket oc:<nuovo-ID> creato.` per ciascuno.

3. **Tag di raggruppamento opzionale (uso interno dev):**

Chiedi:

> "Vuoi raggruppare questi {N} ticket in un tag?"

- **Se sì:**
  - Cerca tra i tag del ticket originale uno riconoscibile come identificativo cliente (es. `ass_cammini_italia`). Se trovato, proponi nome default `<tag-cliente>-<titolo originale kebab-case>`.
  - Se nessun tag è riconoscibile come cliente, chiedi al dev di indicare manualmente il nome cliente da usare.
  - Il dev può modificare il nome proposto prima della creazione.
  - Crea il tag chiamando `create_tag` con `name` e `description` (il `customer_request` originale completo, pre-split, in Markdown), prima senza `confirm` poi con `confirm: true`.
  - Associa il tag restituito a tutti i ticket del gruppo (originale + nuovi) con `attach_story_to_tag`, uno per ticket (anteprima + conferma per ciascuna chiamata).
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

Se l'utente risponde sì, chiama `me` per ottenere `user_id`, poi `update_story` con `story_id`, `status: "progress"` e quel `user_id`, prima senza `confirm` per mostrare la differenza, poi con `confirm: true` dopo l'approvazione.

Se il tool restituisce un errore, avvisa l'utente con un messaggio ("⚠️ Impossibile aggiornare lo status del ticket — procedo comunque con il workflow.") e continua.

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

Chiedi all'utente di confermarlo o modificarlo. Una volta approvato, crea il ticket chiamando `create_story` con `name`, `type`, `customer_request` e `description`, prima senza `confirm` per mostrare l'anteprima calcolata dal tool, poi con `confirm: true`. Il campo `description` è reso da un editor visuale (nessun parsing Markdown): se il testo approvato usa formattazione (titoli, liste, grassetto), convertila in tag HTML equivalenti (`<h3>`, `<ul><li>`, `<strong>`) prima di passarla al tool — non inviare Markdown grezzo. Salva l'ID restituito e usalo come `<ID>` per tutto il resto del workflow.

**Il ticket va creato prima di procedere alla Fase: init-context.** I campi `description` e `customer_request` potranno essere aggiornati a fine workflow (Checklist) con le informazioni emerse dalle fasi successive.

Una volta creato il ticket e salvato l'ID, chiedi:

> "Vuoi impostare lo status del ticket a **progress** e assegnartelo?"

Se l'utente risponde sì, chiama `me` per ottenere `user_id`, poi `update_story` con `story_id`, `status: "progress"` e quel `user_id`, prima senza `confirm` per mostrare la differenza, poi con `confirm: true` dopo l'approvazione.

Se il tool restituisce un errore, avvisa l'utente con un messaggio ("⚠️ Impossibile aggiornare lo status del ticket — procedo comunque con il workflow.") e continua.

Se l'utente non vuole creare il ticket ora, procedi senza ID: usa solo il titolo kebab-case come slug. La domanda progress non viene posta.

### ticket: caso-c

L'utente ha una trascrizione Meet, un brief cliente o qualsiasi materiale da cui estrarre più ticket raggruppati in un tag Orchestrator.

Invoca immediatamente `wm-skills:wm-tag`, passando come contesto qualsiasi testo o link già fornito dall'utente in questa sessione. Da questo momento il controllo del flusso passa interamente a `wm-tag` — non proseguire con nessun'altra fase di `wm-plan`.

### ticket: aggiornamenti-espliciti

Se in qualsiasi momento l'utente chiede di aggiornare un campo del ticket (es. "aggiorna lo status a progress", "scrivi nelle note dev che…"), chiama `update_story`. Chiamalo sempre prima senza `confirm` per mostrare la differenza al dev, e solo dopo l'approvazione esplicita richiamalo con `confirm: true`.

### ticket: estrazione

- `<feature-slug>`: `<ID>-<titolo-in-kebab-case>` (con ticket) o `<titolo-in-kebab-case>` (senza)
- **Riferimento ticket in tutti i documenti:** ogni file creato nelle fasi successive deve riportare in testa `> Ticket: oc:<ID>` (ometti se non c'è ID).
- **Commit convention:** tutti i commit usano `oc:<ID>` come scope (es. `feat(oc:7815): add OSM POI import action`). Senza ticket, usa il titolo kebab-case come scope.

---

## Fase: environment-setup

Questa fase rileva il tipo di progetto e normalizza l'ambiente prima di qualsiasi altra operazione. I flag impostati qui rimangono attivi per tutto il workflow.

**⚠️ Questa fase è FAIL-SOFT.** Qualsiasi errore (Docker daemon non attivo, CLI mancante, `.env` non leggibile) produce `⚠️ Environment setup non disponibile — proseguo con il workflow.` e il workflow continua normalmente. La fase non blocca mai il flusso.

### environment-setup: project-detection

Invoca l'agente `wm-env-detect` sul repo corrente. Restituisce i valori già risolti nel
formato dichiarato nel suo prompt: `stack_type`, `has_docker`, `has_submodules`, `stack_ui`,
`has_phpstan_ci`, `DOCKER_PROJECT_DIR_NAME`, elenco submodule.

Tieni questi valori attivi per tutto il workflow: `DOCKER_PROJECT_DIR_NAME` in particolare
viene riusato a fine workflow in `review-gate: phpstan-check`.

Se l'agente restituisce `RILEVAMENTO FALLITO`, mostra
`⚠️ Environment setup non disponibile — proseguo con il workflow.` e prosegui: questa fase
è fail-soft e non blocca mai.

**Flag interni prodotti dall'agente:**

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
| `has_phpstan_ci` | `true` | trovato `phpstan.neon`/`phpstan.neon.dist` nella root **E** keyword `phpstan` (case-insensitive) in almeno un file `.github/workflows/*.yml` |
| `has_phpstan_ci` | `false` | manca almeno una delle due condizioni |

### environment-setup: domain-mapping

Prima di fare qualsiasi domanda, ispeziona il progetto e identifica:

1. **Submodule presenti** — usa il valore `submodule:` già restituito da `wm-env-detect` in `environment-setup: project-detection`, senza rilanciare `git submodule status`. Leggi `.gitmodules` solo se serve il dettaglio dello scopo di ogni submodule (es. `wm-package` per backend, `wm-core` / `map-core` per frontend), non per rideterminare quali sono presenti.
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
- **Non chiedere ciò che puoi leggere nel codice o nel database.** La ricerca si delega
  all'agente `wm-codebase-research`, che legge nel proprio context e restituisce solo
  conclusioni con prova verbatim.

  **Una chiamata sola, a inizio fase**, con tutte le domande prevedibili dal ticket. Un
  agente per ogni domanda renderebbe il dialogo a scatti, ed è il difetto peggiore in una
  fase che vive di continuità. Se durante il dialogo emerge un buco imprevisto, una chiamata
  puntuale in più è ammessa come eccezione, non come regola.

  **Verifica obbligatoria di ogni prova ricevuta**, prima di usarne il contenuto:

  ```bash
  sed -n '<riga-inizio>,<riga-fine>p' <file>
  ```

  Confronta l'output con l'`Estratto` del dossier. Se anche una sola prova non combacia, il
  dossier è inattendibile: rifai la ricerca nel context principale e segnalalo al dev.

  Se l'agente risponde `Risposta: non determinabile dal repo`, quella domanda va fatta al
  dev: è l'esito che rende utile la delega, non un fallimento.

  Se l'agente restituisce `RICERCA FALLITA`, esegui tu la ricerca nel context principale.

  Se una prova combacia ma la conclusione non convince il dev o sembra fraintendere la
  domanda, rimanda le obiezioni allo stesso `wm-codebase-research` invece di rifare la
  ricerca nel principale — vedi `## Revisione con l'agente` in
  `${CLAUDE_PLUGIN_ROOT}/shared/agent-delegation.md`.

- **Non chiedere ciò che il team ha già deciso in call.** Insieme alla ricerca sul codice,
  interroga `wm-transcript-research` sulle trascrizioni dello scrum: passagli **numero e
  titolo del ticket**, la sua data di creazione e le domande che stai per fare al dev.

  Vale la stessa regola dell'altra ricerca: **una chiamata sola a inizio fase**, per non
  spezzare il dialogo.

  L'agente restituisce le citazioni prima della conclusione: **leggi le citazioni**, non
  solo la risposta. Su un parlato a più voci la conclusione di un agente è
  un'interpretazione, e chi ha partecipato alla call se ne accorge in un attimo.

  Se una citazione risponde a una domanda che avevi in programma, **non farla come se nulla
  fosse**: dì al dev cosa risulta e chiedi conferma — «allo scrum del 09/09 risulta che
  avete deciso X, confermi?». È diverso dal chiedere da zero, e gli fa risparmiare il
  fastidio di ripetersi.

  Se l'agente risponde `Risposta: non determinabile dalle trascrizioni`, la domanda va fatta
  al dev: è l'esito che rende utile la delega, non un fallimento.

  **Guarda sempre la `Copertura`.** Un «non trovato» su tre call lette su venti non è un
  «non se n'è parlato»: è una ricerca incompleta, e se la domanda è importante chiedi
  all'agente una ricerca verbosa prima di girarla al dev.

  Se l'agente restituisce `RICERCA FALLITA`, prosegui il dialogo senza le trascrizioni e
  dillo al dev: la fonte non era raggiungibile, non è che non ci fosse nulla.

- **Ogni domanda deve includere un consiglio da best practice.** Non aspettare che l'utente lo chieda. Dopo aver posto la domanda aggiungi sempre una riga "💡 Best practice:" con la raccomandazione tecnica più rilevante per quel problema specifico, così l'utente può decidere con più contesto. Questa riga è obbligatoria — una domanda senza consiglio è incompleta.
  Prima di ogni domanda scrivi esplicitamente cosa risulta dalle **due** ricerche: *"Dal dossier di `wm-codebase-research` risulta [conclusione/non determinabile, con riferimento alla prova verificata]; dalle call risulta [citazione con data e speaker / nulla] — quindi chiedo:"*. Se non scrivi questa riga, non puoi fare la domanda.

  La riga sulle call non è un ornamento: obbliga a **guardare le citazioni prima di formulare la domanda**, invece di farla e accorgersi dopo che la risposta c'era già. Una fonte che gira e che nessuno legge prima di parlare è peggio di non averla, perché costa e non evita nulla. Se l'agente delle trascrizioni non ha trovato niente, scrivilo: «dalle call, nulla» è un'informazione, e dice al dev che su quel punto non esiste una decisione pregressa.

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

**Eseguita solo se il ticket è di tipo `Feature`.** Per qualsiasi altro tipo (fra quelli ammessi da `create_story`/`update_story`), salta questa fase e procedi direttamente a `Fase: write-plan`.

In tag-mode, questa fase viene eseguita prima di fermarsi (non si procede a write-plan).

### estimation: analisi

La stima è prodotta dall'agente **cieco** `wm-estimate`, che riceve i percorsi di
`overview.md` e `plan.md` e **nient'altro** — nessun riassunto della conversazione, stesso
principio di `challenge: subagent`. Chi ha condotto il dialogo e scritto l'overview stima
ottimista in modo sistematico: il valore dell'agente sta nel non aver vissuto quella
conversazione.

Restano nel context principale:

- **la quota misurata** della pianificazione, che l'agente non può conoscere:

  ```bash
  NOW=$(date -u +"%Y-%m-%dT%H:%M:%S%z")
  ```

  Differenza con `planning_start_at` registrato in `Fase: ticket`. Se non è stato registrato,
  vale il fallback già previsto: dichiararlo al dev e presentare solo la quota stimata.

- **la conferma del dev**, invariata.

Presenta al dev la tabella dell'agente con in testa la riga della pianificazione misurata, e
chiudi sempre con la scomposizione, mai un numero unico fuso:

> **Misurato: \<M\>h + Stimato: \<S\>h = Totale: \<N\>h**

Se l'agente restituisce `STIMA NON POSSIBILE`, produci tu la stima nel context principale
seguendo gli stessi criteri (tempo di esecuzione per componente, buffer di novità di dominio
in valore assoluto, 5% di integrazione trasversale) e dichiara al dev che la stima non è
indipendente.

### estimation: conferma

Chiedi al dev, riportando sempre la scomposizione (mai un numero unico fuso):

> "Accetti questa stima — **Misurato: \<M\>h + Stimato: \<S\>h = Totale: \<N\>h** — la sostituisci con un tuo valore, o la fai rifare a `wm-estimate` con le tue obiezioni?"

Aspetta risposta esplicita. Se il dev propone un valore diverso, usalo senza discutere — la stima finale è sempre quella approvata dal dev. Se il dev modifica solo la quota stimata (implementazione) lasciando invariata quella misurata (pianificazione), aggiorna solo `<S>` e ricalcola `<N>`. Se il dev ritiene il ragionamento sbagliato (es. buffer eccessivo, componente sottostimato), rimanda le sue obiezioni a `wm-estimate` invece di correggere il numero a mano — vedi `## Revisione con l'agente` in `${CLAUDE_PLUGIN_ROOT}/shared/agent-delegation.md`.

### estimation: scrittura su Orchestrator

Chiama `update_story` con `story_id` ed `estimated_hours: <N>`, prima senza `confirm` per mostrare al dev la differenza calcolata dal tool rispetto al valore attuale, poi con `confirm: true` dopo l'approvazione esplicita. Nota interna da tenere a mente nel dialogo con il dev (non un campo separato da inviare): `Misurato: <M>h + Stimato: <S>h = Totale: <N>h`.

**Nessun marcatore di versione della metodologia va scritto nella nota.** La stima di un ticket parte **sempre** dal presupposto che l'esecuzione avvenga con un LLM: è il default, non una variante da segnalare. Un marcatore suggerirebbe che esista ancora una baseline alternativa in dev-hours umane, che non è più il caso.

Se il tool restituisce un errore, avvisa l'utente con `⚠️ Impossibile aggiornare la stima su Orchestrator — procedo comunque.` e continua.

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

Se il dev conferma, chiama `update_story` con `estimated_hours` pari al nuovo totale, prima senza `confirm` poi con `confirm: true` dopo l'approvazione esplicita.

Registra sempre l'evento in `Fase: notes` (sezione "Decisioni"), indipendentemente dal fatto che il dev abbia accettato o rifiutato la revisione.

### execution: divergenze

Quando l'implementazione devia da un task di `plan.md`, **annotalo subito, non a fine lavoro**: una divergenza si descrive bene mentre la si vive, male ricostruendola dopo, quando resta solo il ricordo di averla avuta.

Il contenuto della divergenza va scritto **una sola volta**, in `docs/features/<feature-slug>/notes.md`, sotto una sezione `## Divergenze dal piano, task per task` con un sottotitolo `### <task>` per ciascuna.

In `plan.md`, all'inizio del task divergente, va una riga sola che rimanda alla nota:

```markdown
> ⚠️ L'implementazione ha deviato da questo task: [notes.md](notes.md#task-4-registrazione-condizionale)
```

Motivo: tiene separate due domande — "cosa avevamo deciso" (il piano) e "cosa è successo" (le note) — senza che il piano possa ingannare chi lo apre a metà, senza sapere che quel pezzo è superato. Le alternative sono tutte peggiori: riscrivere il piano cancella la storia del cambiamento, lasciarlo muto trae in inganno, duplicare il testo nei due file crea due verità che divergono al primo aggiornamento.

**Nota tecnica:** l'ancora segue lo slug GitHub del titolo (minuscolo, punteggiatura rimossa, spazi convertiti in trattini).

**Limite:** se la maggioranza dei task ha un rimando, il piano non descrive più il lavoro nemmeno in prima approssimazione — in quel caso conviene riscriverlo invece di continuare ad annotarlo.

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

#### review-gate: phpstan-check

Eseguito **solo se `has_phpstan_ci: true`**, subito dopo `review-gate: subagent` e prima di `review-gate: dialog`.

**Questa fase non è delegabile.** È un hard-block: un agente che fallisce e restituisce
"nessun errore" trasformerebbe il blocco in un pass silenzioso. Vedi
`${CLAUDE_PLUGIN_ROOT}/shared/agentic-feasibility.md`.

**Esecuzione:**

```bash
rm -f /tmp/wm-phpstan.json
if [ "$has_docker" = "true" ]; then
  timeout 300 docker compose -f local.compose.yml exec -T "$DOCKER_PROJECT_DIR_NAME" vendor/bin/phpstan analyse --error-format=json > /tmp/wm-phpstan.json
else
  timeout 300 vendor/bin/phpstan analyse --error-format=json > /tmp/wm-phpstan.json
fi
PHPSTAN_EXIT=$?
jq -r '.files | to_entries[] | "\(.key): \(.value.messages | length) errori"' /tmp/wm-phpstan.json 2>/dev/null
```

L'output completo resta sul filesystem (`/tmp/wm-phpstan.json`) e si consulta solo per gli errori che riguardano il diff.

- **`PHPSTAN_EXIT` diverso da 0 e diverso da 1** (comando non trovato, crash, timeout — `timeout` restituisce `124` allo scadere) → **fallimento infrastrutturale**. Tratta questo caso come un blocco (vedi `review-gate: phpstan-override` sotto), con motivazione di default proposta: "PHPStan non è riuscito a completare l'analisi (exit code $PHPSTAN_EXIT) — verificare ambiente/timeout."
- **`PHPSTAN_EXIT` == 1** (PHPStan ha girato e ha trovato errori) → prosegui al cross-check diff sotto.
- **`PHPSTAN_EXIT` == 0** → nessun errore, nessun blocco, prosegui direttamente a `review-gate: dialog`.

**Cross-check diff vs preesistenti (solo se `PHPSTAN_EXIT == 1`):**

```bash
git diff --name-only > /tmp/wm-plan-diff-files.txt
# Incrocia i file riportati negli errori PHPStan (`.file` nel JSON) con /tmp/wm-plan-diff-files.txt
```

- Per ogni errore riportato da PHPStan, verifica se il file corrispondente è presente in `/tmp/wm-plan-diff-files.txt`.
- **Se almeno un errore è su un file del diff corrente** → blocco (vedi `review-gate: phpstan-override` sotto), motivazione di default: "PHPStan ha trovato N errori sui file modificati in questo task."
- **Se tutti gli errori sono su file fuori dal diff corrente** (debito preesistente) → nessun blocco. Presenta all'utente:

  > "PHPStan ha trovato N errori preesistenti su file non toccati da questa feature ([lista file]). Vuoi che crei un ticket Orchestrator separato per tracciare questo debito tecnico?"

  Se sì, crea il ticket chiamando `create_story` con `name` sintetico e `customer_request` con l'elenco degli errori, prima senza `confirm` per l'anteprima poi con `confirm: true`. Per il campo `type` scegli, fra i valori ammessi dal tool, quello che rappresenta un intervento di manutenzione pianificato — non inventare un valore né darne uno per scontato. Poi prosegui a `review-gate: dialog` senza bloccare il commit.

#### review-gate: phpstan-override

Attivato quando `review-gate: phpstan-check` rileva un blocco (errori sul diff corrente o fallimento infrastrutturale).

1. Presenta all'utente gli errori/il fallimento riscontrato.
2. Proponi (o deduci dal contesto) una motivazione per un eventuale bypass — es. "Errore preesistente in un file limitrofo non toccato direttamente" o "Falso positivo noto di PHPStan su questo pattern".
3. Mostra la motivazione proposta in preview e chiedi al dev di confermarla o modificarla.
4. Chiedi conferma esplicita e distinta dalla conferma generica di commit di `review-gate: dialog`, con questo messaggio:

   > "PHPStan blocca il commit per [errori di codice sul diff / fallimento infrastrutturale]. Vuoi bypassare questo blocco specifico? Verrà registrato in notes.md con la motivazione: \"[motivazione confermata]\"."

5. Solo se il dev conferma esplicitamente il bypass (non basta il "procedi" generico del punto 3 di `review-gate: dialog`), consenti di proseguire e registra in `docs/features/<feature-slug>/notes.md` (sezione "Decisioni") una riga con: motivazione, timestamp, e la responsabilità esplicita attribuita al dev.
6. Se il dev non conferma il bypass, il workflow resta bloccato su questo punto: nessun commit finché gli errori non sono risolti o il bypass non viene confermato.

#### review-gate: dialog

1. Presenta all'utente il riepilogo prodotto dal subagente (o dal fallback).
1bis. Se `has_phpstan_ci: true`, esegui `review-gate: phpstan-check` ora, prima di procedere al punto 2. Se ne emerge un blocco, gestiscilo con `review-gate: phpstan-override` prima di continuare.
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
- **Divergenze dal piano:** vanno annotate in `execution: divergenze` nel momento in cui accadono, non qui. In questa fase **verifica che i rimandi funzionino**: ogni riga `> ⚠️ L'implementazione ha deviato` in `plan.md` deve puntare a un titolo che esiste davvero in `notes.md`. Un rimando rotto non produce nessun errore in Markdown — resta lì e nessuno se ne accorge — quindi il controllo va fatto a macchina:

  ```bash
  cd docs/features/<feature-slug>
  # estrae le ancore citate nel piano e le confronta con i titoli presenti nelle note
  grep -o 'notes\.md#[a-z0-9-]*' plan.md | sed 's/.*#//' | sort -u > /tmp/wm-ancore-citate
  grep '^### ' notes.md | sed 's/^### //' \
    | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9 -]//g; s/ /-/g' | sort -u > /tmp/wm-ancore-esistenti
  comm -23 /tmp/wm-ancore-citate /tmp/wm-ancore-esistenti
  ```

  Se il comando stampa qualcosa, quei rimandi sono rotti: correggi il titolo nelle note o l'ancora nel piano prima di proseguire.

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

Registra nel repo target ciò che di questo lavoro serve a chi lo toccherà domani.

**L'identificativo del lavoro è il nome della cartella degli artefatti** creata in
`Fase: overview`, cioè `docs/features/<slug>/`. Leggilo, non ricalcolarlo: la regola
`<ID>-<titolo-in-kebab-case>` serve a *creare* quella cartella, non a ri-derivarne il nome
ogni volta. Se il titolo del ticket nel frattempo è cambiato, ri-derivare produrrebbe un
nome diverso da quello reale, e il legame si romperebbe in silenzio.

```bash
ls -d docs/features/*/ | grep "<ID>"      # con ticket
```

### update-context: pagina di conoscenza

Il sapere che deve sopravvivere a questo lavoro va in `docs/knowledge/<argomento>.md`,
**organizzato per argomento e non per ticket**: chi domani tocca quel pezzo di sistema non si
chiede «cosa fu deciso nel ticket X», si chiede «come funziona e cosa è già stato provato».

**Quale pagina.** Cerca fra le pagine esistenti una che copra il tema toccato:

```bash
ls docs/knowledge/ 2>/dev/null
```

- **Esiste già** → aggiornala (vedi sotto).
- **Non esiste, ed è la prima volta che si tocca questo tema** → creala con il nome del lavoro,
  cioè lo stesso slug della cartella degli artefatti.
- **Non esiste, ma un'altra pagina copre lo stesso tema con un altro nome** → proponi al dev di
  fondere le due in una pagina di argomento, con un nome che descrive il tema e non il lavoro.
  **Un argomento nasce al secondo lavoro che lo tocca**, non al primo: non inventare tassonomie
  in anticipo.

**Struttura della pagina** — lo stato attuale in cima, la storia sotto:

```markdown
# <Argomento>

## Come funziona oggi
<cosa vale adesso: il comportamento corrente e i vincoli che lo governano>

## Perché così
- **<scelta>** (oc:<ID>): <motivazione>

## Come ci siamo arrivati
- **<scelta precedente>** (oc:<ID>, superata): <perché è stata abbandonata>
```

Ogni voce porta il ticket da cui proviene: è il legame con il cantiere, che resta la fonte
completa di com'è andata.

**Aggiornare una pagina esistente è una riscrittura, non un'aggiunta.** Apri la pagina,
stabilisci cosa è ancora valido, riscrivi `## Come funziona oggi` e sposta in
`## Come ci siamo arrivati` ciò che il tuo lavoro ha superato, **con il motivo**. Una decisione
caduta non si cancella: chi domani proporrà di nuovo quella strada deve poter leggere perché
era stata abbandonata.

**Mostra sempre al dev il prima e il dopo** di una pagina riscritta, non solo la versione
nuova: sovrascrivere una pagina densa è l'operazione che in questa fase può fare più danni, e
il dev è l'unico che può accorgersene.

### update-context: procedure e guide

Non tutto ciò che resta è conoscenza. Prima di scrivere, stabilisci a quale domanda risponde
il testo — il criterio completo è in `${CLAUDE_PLUGIN_ROOT}/shared/claude-md-rules.md`:

- **perché funziona così** → `docs/knowledge/<argomento>.md`
- **come si fa** (procedura per chi lavora sul repo) → `docs/howto/<procedura>.md`
- **cosa fa il prodotto e come si usa** (per l'utente finale) → `docs/guide/<argomento>/`

**Le guide cliente.** Se in `Fase: estimation` è stato dichiarato un deliverable extra di
documentazione utente, è qui che va scritto: `docs/guide/<argomento>/`, con gli screenshot
nella stessa cartella del testo. Una guida non nomina file, classi, branch o dettagli
implementativi — vale la stessa regola della risposta al cliente in
`update-context: orchestrator`.

Se il deliverable era stato stimato e non viene prodotto, dillo al dev invece di ometterlo in
silenzio: era una voce della stima che hai approvato insieme.

**Una guida è destinata a essere pubblicata**, mentre tutto ciò che le sta accanto in `docs/`
è interno. Scrivila dando per scontato che la legga chiunque: nessun percorso interno, nessun
nome di branch, nessun riferimento ad altri clienti. Se il repo non ha ancora una
pubblicazione configurata, non improvvisarla: scrivi la guida e segnalalo al dev, perché la
scelta di cosa esporre non è tua.

### update-context: indice nel CLAUDE.md

Nel `CLAUDE.md` del repo target va **una riga sola** per argomento, sotto `## Conoscenza`:

```markdown
## Conoscenza

| Argomento | Cosa copre | Pagina |
|---|---|---|
| <Argomento> | <un gancio di una frase> | [docs/knowledge/<argomento>.md](docs/knowledge/<argomento>.md) |
```

Se il lavoro ha aggiornato una pagina già presente nell'indice, **la riga non cambia**: cambia
la pagina. L'indice elenca argomenti, non lavori, quindi cresce molto più lentamente di quanto
crescesse la vecchia tabella delle feature.

**Il rimando è un link Markdown o un percorso in prosa, mai `@percorso/file.md`**: in un
`CLAUDE.md` quella sintassi non è un link ma un import, e il file viene caricato all'avvio di
ogni sessione — annullando il motivo stesso di averlo separato.

**Non esiste una seconda sezione.** Le vecchie `## Feature disponibili` e
`## Decisioni architetturali` avevano la stessa granularità — una voce per lavoro — quindi
erano due indici dello stesso insieme e si ripetevano per costruzione. L'indice è uno.

**Cosa non entra nell'indice.** Il criterio è: *è nato da un lavoro?* Se sì sta in una pagina
di conoscenza. Se no — struttura delle cartelle, convenzioni di naming, come si valida prima
del commit — non è la decisione di una feature, è il repo: sta in `## Regole del repo`, nel
corpo del `CLAUDE.md`, fuori dall'indice.

**Se il lavoro introduce una regola da seguire d'ora in poi**, le due cose vanno in due posti
diversi e non è una duplicazione: il *perché* della scelta, con le alternative scartate, va
nella pagina di conoscenza col suo ticket; l'*obbligo* — cosa fare ogni volta che si tocca il
repo — va in `## Regole del repo`, in forma imperativa e senza il ragionamento. Proponi sempre
al dev la riga da aggiungere: è la sezione che governa come si lavora nel suo repo, e la scrive
lui.

### update-context: le trappole non sono regole del repo

Una parte di ciò che un lavoro lascia non è né conoscenza né politica: è una **trappola** — un
comportamento che fa sbagliare chi non lo conosce, e che non si deduce leggendo il codice.
`->rules()` su un campo Media blocca ogni salvataggio del form; un worker Horizon già avviato
ignora la classe Job che hai appena modificato; `identifier` non è in `$fillable` e passarlo al
costruttore lo scarta in silenzio.

La differenza da una regola del repo è **quando scatta**. Una regola vale sempre («la
documentazione è in italiano»); una trappola vale solo quando tocchi una certa cosa. Metterle
insieme trasforma `## Regole del repo` in un contenitore indifferenziato che cresce a ogni
ticket — il primo passo verso il file che si dovrà smontare.

Le trappole vivono quindi in **regole path-scoped**, un file per soggetto sotto
`.claude/rules/`, così si caricano solo quando si tocca il codice che le riguarda invece di
pesare su ogni sessione:

```markdown
---
paths:
  - "src/Nova/**"
---

# Trappole: <soggetto>

- <cosa fa sbagliare>: <la conseguenza, in una riga> — <cosa fare invece> (oc:<ID>)
```

Nel `CLAUDE.md` va **solo la riga di rimando** nella sezione `## Trappole`, perché il soggetto
sia noto anche a chi crea un file nuovo senza averne letto nessuno. Il dettaglio del criterio,
compresi i `paths:` tipici di un frontend, è nelle regole condivise che hai letto.

I soggetti nascono dal repo, non da una tassonomia decisa prima, e **un soggetto nasce al
secondo lavoro che lo tocca**. Verifica che ogni cartella citata nei `paths:` esista: un pattern
che non corrisponde a nulla è una regola che non si carica mai, e nessuno lo segnala.

Il formato è vincolante, ed è quello che tiene il file piccolo:

- **una riga per trappola**, imperativa, con la conseguenza;
- **l'ID del ticket a fine riga**, mai come titolo di sezione: serve a risalire alla storia,
  non a organizzarla;
- **la cronaca resta nel `notes.md`** del cantiere — i round di review, cosa è stato provato e
  scartato, le decisioni di scope. Nel `CLAUDE.md` va solo ciò che, se sparisse, farebbe
  sbagliare qualcuno.

**Una trappola con un effetto verso l'esterno non va in queste sezioni, va in cima al file**,
insieme alle altre regole che precedono tutto: un test che scrive davvero su un registro
condiviso, una suite che cancella un database, una chiamata che notifica un cliente. Chi arriva
al file dal fondo non le leggerebbe in tempo.

### update-context: repo non ancora in questa forma

Se il `CLAUDE.md` del repo target ha ancora `## Feature disponibili` e/o
`## Decisioni architetturali` e non ha `## Conoscenza`, **non migrare di tua iniziativa**: la
riorganizzazione di un `CLAUDE.md` esistente è un'operazione a sé, che il dev avvia quando
vuole invocando `wm-context-doctor`.

In quel caso:

1. scrivi comunque `docs/knowledge/<argomento>.md` — è un file nuovo, non tocca nulla di esistente;
2. aggiungi la riga nella sezione che il repo già usa (`## Feature disponibili`), con il link
   alla pagina al posto della descrizione lunga, e **non** scrivere un blocco in
   `## Decisioni architetturali`: il suo contenuto è ora nella pagina;
3. segnalalo al dev **una volta sola**, senza insistere:

   > `ℹ️ Questo repo usa ancora la forma vecchia del CLAUDE.md. Quanto emerso da questo lavoro è in docs/knowledge/<argomento>.md; per riorganizzare il resto puoi invocare wm-context-doctor.`

**Controllo di forma prima di scrivere.**

Prepara il testo da aggiungere, poi invoca `wm-context-guard` passandogli il percorso del
`CLAUDE.md` e il testo proposto. L'agente applica le regole di
`${CLAUDE_PLUGIN_ROOT}/shared/claude-md-rules.md` e restituisce i rilievi: ripetizione,
contraddizione, duplicato dal codice, fuori posto.

Verifica ogni rilievo che cita una voce esistente con l'estratto e la riga forniti, poi
correggi il testo di conseguenza. Una **contraddizione** non si risolve mai da soli: va
portata al dev, perché rimuovere o marcare come superata una voce esistente tocca testo già
approvato.

Se il dev ritiene un rilievo sbagliato, puoi rimandare l'obiezione a `wm-context-guard`
(vedi `## Revisione con l'agente` in `${CLAUDE_PLUGIN_ROOT}/shared/agent-delegation.md`); se
resta in disaccordo dopo due giri, decide il dev e si procede.

Se l'agente risponde `NESSUN RILIEVO`, procedi. Se risponde `CONTROLLO FALLITO: <motivo>`
o non risponde affatto (timeout, errore di spawn), scrivi comunque: il controllo è un
ausilio, non un gate.

Il `wm-context-guard` applica le stesse regole che valgono qui: se segnala che una voce
esistente dice già ciò che stai per scrivere, o che lo contraddice, quella è informazione da
portare al dev prima di scrivere, non da risolvere in autonomia.

Mostra le modifiche al `CLAUDE.md` all'utente prima di scriverle.

---

## Composizione con altre skill Webmapp

- **`ui-ux-pro-max`** — invocata automaticamente in environment-setup: ux-ui-detection quando rilevati componenti UI/UX (Vue, Angular, HTML/CSS). Richiede `/plugin install ui-ux-pro-max@wm-marketplace` se non installata.

---

## Checklist di completamento

Prima di dichiarare il workflow concluso, verifica che esistano tutti e tre i file:

- [ ] `docs/features/<feature-slug>/overview.md` — approvato dall'utente (riferimento ticket presente se applicabile)
- [ ] `docs/features/<feature-slug>/plan.md` — approvato dall'utente (riferimento ticket presente se applicabile)
- [ ] `docs/features/<feature-slug>/notes.md` — compilato (anche solo con "Nessuna deviazione") (riferimento ticket presente se applicabile)

**Questi tre file sono obbligatori sempre, con o senza ticket Orchestrator.**
- [ ] `CLAUDE.md` del progetto target aggiornato — **nella forma che quel repo usa davvero**: una riga sotto `## Conoscenza` se il repo è già in questa struttura, altrimenti nella sezione che usa oggi (`## Feature disponibili`), senza migrarlo di iniziativa. Più l'eventuale riga di regola proposta al dev, se il lavoro ne ha introdotta una da seguire d'ora in poi
- [ ] Sorgente del diagramma di flusso `wm-plan` aggiornato, se questa sessione ha modificato il workflow della skill nel repo `claude-marketplace` — la pagina si ripubblica da sé al push, nessun redeploy manuale. Un controllo in CI verifica che le fasi della skill e i nodi del diagramma coincidano: se hai aggiunto o rinominato una fase e non hai toccato la pagina, fallisce

### update-context: orchestrator (solo se esiste un ticket oc:\<ID\>)

- [ ] Leggi lo status attuale del ticket con `get_story`
- [ ] Prova a impostare uno status non ammesso con `update_story` (senza `confirm`) e usa l'elenco che il tool restituisce nell'errore, oppure consulta lo schema del tool
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
- [ ] Solo dopo approvazione esplicita, chiama `update_story` con i campi `status`, `description`, `customer_request` — prima senza `confirm` per mostrare la differenza, poi con `confirm: true`

  **Importante:** manda solo il testo pulito nei campi `customer_request` e `description` — il backend chiama internamente `addResponse()` e `addDevNote()` che gestiscono formato HTML, timestamp, prepend e notifiche. Non costruire HTML manualmente.
