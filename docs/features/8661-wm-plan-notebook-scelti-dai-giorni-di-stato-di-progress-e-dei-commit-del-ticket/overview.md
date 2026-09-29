> Ticket: oc:8661

# wm-plan: notebook scelti dai giorni di stato, di progress e dei commit del ticket

## Cosa cambia

Quando `wm-plan` cerca nelle call cosa si è detto su un ticket, non passa più a
`wm-transcript-research` una finestra di date («da qualche giorno prima della creazione del ticket
a oggi»), ma i **giorni del ticket**: i giorni in cui al ticket è successo qualcosa. L'agente
interroga il notebook `scrum AAAA-MM-GG` di ciascuno di quei giorni che ha almeno una call.

I giorni del ticket sono l'unione di:

- **giorni di cambio stato**: le date dei `from` di `intervals` in
  `GET /api/stories/{story}/status-history` (oc:8636), compreso il giorno di creazione, che è
  sempre il `from` del primo periodo;
- **giorni in `progress`**: le date di `days` nella stessa chiamata con `?status=progress`;
- **giorni con un commit del ticket**: le date dei commit che hanno `(oc:<ID>)` nel messaggio, in
  tutti i branch dei repository indicati;
- **oggi**.

Li calcola tutti un tool nuovo del server MCP, **`get_story_days`**, che riceve l'id del ticket e,
facoltativamente, i repository in cui cercare i commit (senza, usa il repository in cui gira
`wm-plan` con i suoi submodule), e restituisce l'elenco già unito, ordinato e senza doppioni.
`wm-plan` chiede, riceve i giorni pronti e li passa all'agente: il calcolo non passa dal modello.

I giorni del ticket valgono per **ogni ricerca sulle call di quel ticket** nella sessione: in
`reverse-interaction` e nelle domande dirette del dev («cosa si è deciso su questo ticket?»). Si
esce dall'elenco solo quando il dev chiede giorni precisi («guarda anche il 10/09»). L'elenco si
ricalcola a ogni ricerca, quindi è sempre aggiornato.

## Perché

La finestra di date porta dentro anche i giorni in cui il ticket era fermo (`waiting`, `todo`),
e per un ticket lungo sono molti notebook da interrogare senza nessuna informazione in più. Di
un ticket si parla in call quando succede qualcosa: quando lo si assegna, quando ci si lavora,
quando torna da una review, quando lo si chiude. oc:8636 è stato fatto per preparare questo
lavoro.

I commit si aggiungono perché lo stato su Orchestrator lo aggiorna una persona a mano: un giorno
con un commit del ticket è un giorno di lavoro vero anche se lo stato non era `progress`.

Il calcolo sta in un tool Go e non nelle istruzioni della skill perché è meccanico: nel codice è
testato e dà sempre lo stesso risultato, mentre il modello che scrive ogni volta i comandi `git`
può dimenticare un submodule o sbagliare il filtro senza che nessuno se ne accorga.

## Requisiti

**Server MCP (`plugins/wm-skills/mcp/`)**

- [ ] Tool `get_story_days` nel gruppo `stories`, con `story_id` obbligatorio e `repos` facoltativo
      (percorsi dei repository in cui cercare i commit). Senza `repos` cerca nel repository in cui
      gira `wm-plan`, cioè la directory di lavoro del server, e nei suoi submodule, se è un
      repository git; se non lo è, usa solo Orchestrator e lo dice in `warnings`. Restituisce `story_id`, `current_status`,
      `days` (date `AAAA-MM-GG` in ordine crescente, senza doppioni) e `warnings`.
- [ ] Giorni di Orchestrator: `GET /api/stories/{story}/status-history` e la stessa con
      `?status=progress`, con `Client.Do`. Le date sono la parte `AAAA-MM-GG` delle stringhe
      restituite, che sono già nel fuso `Europe/Rome`: nessuna conversione.
- [ ] Giorni dei commit: per ogni repository, `git fetch --all` con `GIT_TERMINAL_PROMPT=0` e un
      timeout (il comando non chiede mai credenziali e non resta fermo), poi `git log --all` con
      `(oc:<ID>)` cercato come testo esatto nel messaggio, data d'autore nel fuso del computer.
- [ ] Oggi sempre compreso.
- [ ] Nessun errore blocca il tool: se l'endpoint non risponde (401, 403, 404, rete) o un
      `git fetch` fallisce, il tool restituisce comunque i giorni che è riuscito a calcolare e
      spiega il problema in `warnings`, con il messaggio già tradotto da `Client.Do`.
- [ ] Test con `recordingServer`, `testSession` e repository git creati nel test: l'esempio della
      description di oc:8636 (cambi il 21 e il 23, `waiting` il 22 → `2026-09-21`,
      `2026-09-23`, il 22 escluso); un ticket senza righe di stato (solo il giorno di creazione);
      un periodo che inizia alle 23:30 italiane; un commit con `(oc:<ID>)` su un branch non unito
      e uno con il numero in un altro contesto, che non deve contare; l'endpoint che risponde 404,
      con i giorni dei commit comunque restituiti.
- [ ] Binario ricompilato con `build.sh`.

**Agente `wm-transcript-research`**

- [ ] La richiesta «domande su una finestra di giorni» diventa «domande sui **giorni del ticket**»:
      riceve un elenco di date, non un intervallo, e interroga il notebook di ogni data dell'elenco
      che ha almeno una call. Oggi è sempre compreso; senza indicazioni i giorni sono solo oggi. Il
      testo sulla finestra si toglie del tutto dal prompt.
- [ ] Ricerca delle call su Drive per un elenco di date sparse (una clausola `title contains` per
      data, tutte le pagine).
- [ ] `Copertura`: la riga `finestra: <data>–<data>` diventa `giorni del ticket: <elenco>`, con i
      giorni con almeno una call.
- [ ] Pulizia: un notebook `scrum …` è vecchio se è stato **creato** più di 90 giorni fa (data di
      creazione in `notebook_list`), non se è vecchio il giorno del suo nome. I notebook
      interrogati nella ricerca in corso non si segnalano mai come vecchi.

**Skill `wm-plan`**

- [ ] In `reverse-interaction` chiama `get_story_days` con l'id del ticket, senza `repos` (vale il
      repository in cui lavora, con i submodule), e passa i giorni all'agente al
      posto della finestra. Mostra al dev i giorni e gli eventuali `warnings` in modo evidente.
- [ ] Le domande dirette del dev sulle call di quel ticket usano gli stessi giorni, ricalcolati; i
      giorni indicati dal dev si aggiungono.
- [ ] Dopo uno split (`caso-a-split-execution`), per i ticket nuovi si usano anche i giorni del
      ticket originale.
- [ ] Senza ticket, e nel caso B, i giorni sono solo oggi (come ora).
- [ ] `get_story_days` nella tabella `## Orchestrator`.

**Documentazione**

- [ ] `docs/knowledge/wm-transcript-research.md`: i giorni del ticket al posto della finestra, la
      frase «quando l'API li esporrà (oc:8636)» tolta, la pulizia per data di creazione.
- [ ] `docs/howto/provare-wm-transcript-research.md`: le prove che passano una finestra
      all'agente passano un elenco di giorni.

## Rischi

- **Il `git fetch` fallisce o è lento** (reale: rete assente, credenziali, passphrase): il tool non
  chiede nulla e ha un timeout; si usano i branch già presenti in locale e lo si dice in
  `warnings`, che `wm-plan` mostra al dev.
- **Ticket creati prima del 03/07/2024** (reale, raro per un ticket su cui si lavora oggi):
  l'endpoint non ha righe di stato e restituisce un solo periodo dalla creazione; i giorni sono il
  giorno di creazione, i commit e oggi.
- **Un giorno vecchio senza notebook** (reale): l'agente lo crea e lo riempie alla prima ricerca,
  come fa già; costa solo quella volta. Con la pulizia per data di creazione, un notebook ricreato
  non viene subito proposto per la cancellazione.
- **Ticket nati da uno split** (reale): non hanno storia propria; si usano i giorni del ticket
  originale, noti al momento dello split.
- **Cambi di stato fatti in blocco**, per esempio alla chiusura di un rilascio (reale, di poco
  conto): aggiungono qualche giorno senza informazioni; accettato.
- **Il tool ora lancia `git`** oltre a chiamare Orchestrator: i comandi sono solo di lettura
  (`fetch`, `log`) e girano solo sui percorsi passati da `wm-plan` o sul repository in cui gira.
- **Sessione aperta fuori da un repository** (reale): Claude Code avvia il server MCP con la
  directory in cui è aperta la sessione (verificato il 29/09 sui server attivi: `forestas`,
  `claude-marketplace`, `webmapp-app`, ma anche la home). Aperta nella home, il valore predefinito
  di `repos` non trova un repository: il tool usa solo Orchestrator e lo scrive in `warnings`.

## Out of scope

- Modifiche all'endpoint di Orchestrator.
- Commit su repository che non sono sul computer del dev.

## Moduli toccati

Tutti nel repo `claude-marketplace`:

- `plugins/wm-skills/mcp/internal/tools/storydays.go` (nuovo) — tool `get_story_days`, registrato
  in `registry.go` nel gruppo `stories`
- `plugins/wm-skills/mcp/internal/tools/storydays_test.go` (nuovo) — test del tool
- `plugins/wm-skills/bin/orchestrator-mcp` — binario ricompilato
- `plugins/wm-skills/agents/wm-transcript-research.md` — giorni del ticket al posto della finestra,
  pulizia per data di creazione
- `plugins/wm-skills/skills/wm-plan/SKILL.md` — `reverse-interaction`, domande dirette, split,
  tabella `## Orchestrator`
- `docs/knowledge/wm-transcript-research.md`, `docs/howto/provare-wm-transcript-research.md`
- `docs/guide/wm-plan-diagramma/index.html` — solo se cambia il testo di `reverse-interaction`
