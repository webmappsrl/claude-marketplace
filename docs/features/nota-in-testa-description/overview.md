# Aggiungere informazione alla description di un ticket senza riscriverla

## Cosa cambia

`update_story`, nel server MCP di `wm-skills`, impara ad **aggiungere** informazione alla
`description` di un ticket senza che l'agente la ricopi. Resta l'unico tool che scrive su un
ticket, e sceglie cosa fare in base ai parametri che riceve:

- **`description`** → sostituisce il campo con il testo ricevuto, come oggi. Serve per
  riformattare o invalidare la `description`.
- **`prepend` e/o `annotations`** → chiama al suo interno una funzione di append: legge la
  `description` attuale, mette `prepend` in testa, inserisce ogni annotazione subito dopo una
  frase esatta del testo esistente, lascia identico tutto il resto, e usa il risultato come
  `description` della PATCH.
- **entrambi** → rifiuta la chiamata.

Gli altri campi (`status`, `customer_request`, ecc.) partono nella **stessa PATCH**: una sola
scrittura, che riesce o fallisce tutta insieme.

L'anteprima di `update_story` mostra **ogni campo che cambia con il contenuto completo**, prima e
dopo, senza troncamenti né rimandi.

Le skill seguono questa divisione:

- `wm-review-ticket` (Fase 6b) scrive l'esito con una sola chiamata `update_story`: il ciclo
  nuovo in `prepend`, le etichette sui cicli precedenti in `annotations`, lo status. Il formato
  dei cicli e delle etichette resta quello attuale.
- `wm-plan` (Checklist `update-context: orchestrator` e `ticket: aggiornamenti-espliciti`) scrive
  le note dev in `prepend`, insieme a status e `customer_request` nella stessa chiamata.
- In entrambe sparisce la regola «rileggi la `description` e ricopiala identica».

Orchestrator non cambia: la PATCH che sostituisce (oc:8549) resta com'è.

## Perché

Da oc:8549, in produzione dal 15 settembre 2026, `PATCH /api/stories/{id}` sostituisce la
`description` invece di aggiungere in testa. `wm-plan` mandava solo la nota di chiusura e così
ha cancellato l'analisi iniziale di 11 ticket.

La correzione della PR #20 chiede all'agente di rileggere la `description` e ricopiarla
identica. Su testi lunghi (oc:8543 ha 18.000 caratteri) una copia fatta dal modello può
riassumere o saltare pezzi, e nessuno se ne accorge: l'anteprima di `update_story` mostra solo
i primi 120 caratteri del campo. Solo il codice copia senza errori.

Non si aggiunge un endpoint a Orchestrator: la PATCH esiste già, e qui si decide solo come le
skill la usano.

## Requisiti

**Parametri e modalità**

- [ ] `update_story` accetta due parametri nuovi, facoltativi: `prepend` (HTML da mettere in
      testa) e `annotations` (elenco di `{after, text, occurrence}`; `occurrence` è facoltativo).
- [ ] Con `description` insieme a `prepend` o `annotations` il tool rifiuta la chiamata senza
      scrivere.
- [ ] Senza `prepend` né `annotations`, `update_story` si comporta esattamente come oggi.
- [ ] La descrizione del tool dice: per aggiungere si usano `prepend`/`annotations`;
      `description` sostituisce tutto il campo.

**Funzione di append**

- [ ] Legge la `description` attuale con GET; `null` vale come testo vuoto.
- [ ] Il risultato è `prepend` + `description` attuale con le annotazioni inserite. Nessuna
      intestazione di autore o data viene aggiunta.
- [ ] Ogni `after` si cerca **solo nella `description` attuale**, mai in `prepend`, confrontando
      l'HTML grezzo.
- [ ] Se `after` non compare, il tool si ferma senza scrivere e dice quale annotazione ha
      fallito.
- [ ] Se `after` compare più volte, serve `occurrence` (1 = la prima): senza, o con un numero
      oltre quelle presenti, il tool si ferma senza scrivere.
- [ ] Il punto di inserimento non può cadere dentro un tag HTML (fra `<` e `>`): in quel caso il
      tool si ferma senza scrivere.
- [ ] `text` viene inserito subito dopo `after`, preceduto da uno spazio. Le posizioni si
      calcolano tutte sulla `description` attuale; due annotazioni con lo stesso punto di
      inserimento vengono rifiutate.
- [ ] **Controllo di ripetizione:** se la `description` inizia già con `prepend`, o se un `text`
      si trova già subito dopo il suo `after`, quel pezzo non viene aggiunto di nuovo; se non
      resta nulla da aggiungere, il tool risponde «già presente, nessuna scrittura».
- [ ] Un test verifica che, togliendo dal risultato `prepend` e ogni «spazio + `text`»
      inserito, si ottenga la `description` di partenza carattere per carattere.

**Anteprima e conferma**

- [ ] Senza `confirm` il tool non scrive e mostra, per **ogni campo che cambia**, l'etichetta e
      il contenuto completo prima e dopo, esattamente come verrà scritto: nessun troncamento,
      nessun «il resto non cambia», nessun rimando. Vale anche per `description` in modalità
      sostituzione.
- [ ] I campi HTML (`description`, `customer_request`) nell'anteprima sono **resi come testo
      leggibile**, non come HTML né Markdown: titoli su una riga a sé in maiuscolo, elenchi con
      «- » e rientro, paragrafi separati da una riga vuota, grassetto come testo semplice, link
      come testo con l'indirizzo fra parentesi, entità decodificate. Nessun tag visibile.
- [ ] In modalità append l'anteprima elenca in più ogni annotazione con la riga in cui finisce.
- [ ] Se la `description` nuova è meno della metà di quella attuale, l'anteprima si apre con
      `⚠️ description: <prima> → <dopo> caratteri — si perde il <N>% del testo attuale.` Non
      blocca la scrittura.
- [ ] Con `confirm: true` il tool rilegge la `description`, ricompone con gli stessi parametri,
      rifà tutti i controlli e manda **una sola PATCH** con tutti i campi. Vince chi scrive per
      ultimo: nessun controllo sulle scritture contemporanee.

**Skill**

- [ ] `wm-review-ticket` 6b scrive esito e status con una sola chiamata `update_story` in
      modalità append e non chiede più di ricopiare la `description`.
- [ ] `wm-plan` scrive le note dev con `prepend` nella Checklist (insieme a status e
      `customer_request`) e negli aggiornamenti espliciti; `description` resta indicata solo per
      riformattare o invalidare.
- [ ] Lo schema di `update_story` e `create_story` dichiara `additionalProperties: false`: un
      parametro sconosciuto o scritto male dà errore invece di essere ignorato.
- [ ] Entrambe le skill, e la regola sulle scritture in `wm-plan` → `## Orchestrator`: ogni
      anteprima mostrata al dev riporta etichetta e contenuto completo di ogni campo, anche se
      ripetitivo, mai rimandi a testo già mostrato, e i campi HTML resi come testo leggibile.
- [ ] `plugins/wm-skills/shared/orchestrator-fallback.md`: quando il server MCP non parte,
      l'aggiunta alla `description` non si fa con una chiamata scritta a mano.

**Documentazione e build**

- [ ] `docs/knowledge/orchestrator-integrazione.md` descrive quando si aggiunge e quando si
      sostituisce.
- [ ] Il binario `plugins/wm-skills/bin/orchestrator-mcp` è ricompilato con `build.sh`.

## Rischi

- **Frasi ripetute o spezzate dall'HTML** (challenge): la ricerca di `after` è limitata al testo
  esistente, le ripetizioni si risolvono con `occurrence`, gli inserimenti dentro un tag sono
  rifiutati, e l'anteprima mostra la riga di ogni annotazione.
- **Binario vecchio con skill nuova**: skill e binario viaggiano nello stesso plugin, quindi
  capita solo aggiornando il plugin a sessione aperta, e si risolve riavviando. Con il binario
  1.4.2 `prepend` verrebbe ignorato (lo schema scritto a mano non dichiara
  `additionalProperties: false`): la `description` non viene toccata, ma la nota non viene
  scritta. Accettato, nessun controllo nelle skill; da questa versione un parametro sconosciuto
  dà errore.
- **Scrittura ripetuta dopo un timeout**: il controllo di ripetizione evita di aggiungere due
  volte lo stesso blocco o la stessa etichetta.
- **Scritture a metà** (nota scritta, status no, o notifica al cliente senza nota): una sola
  PATCH per tutti i campi.
- **Sostituzione usata al posto dell'aggiunta**: resta possibile, il codice non distingue una
  sostituzione voluta da una sbagliata. Difese: la descrizione del tool, le istruzioni delle
  skill, l'anteprima completa di prima e dopo.
- **Errori di markup invisibili nel testo reso**: l'anteprima leggibile non mostra un tag
  rotto; li previene il rifiuto degli inserimenti dentro un tag.
- **Modifiche contemporanee da Nova**: accettato, vince chi scrive per ultimo; la
  ricomposizione alla conferma riduce la finestra al tempo fra GET e PATCH.

## Out of scope

- Nessuna modifica a Orchestrator: nessun endpoint nuovo, nessun ripristino di `addDevNote()`.
- Nessuna intestazione di autore e data aggiunta dal tool.
- Nessun recupero dei ticket che hanno già perso l'analisi iniziale.
- Nessuna protezione dalle scritture contemporanee.
- La release di `wm-skills` (bump di versione, tag): segue la checklist di rilascio, a parte.

## Moduli toccati

Tutti nel repo `claude-marketplace`:

- `plugins/wm-skills/mcp/internal/tools/stories.go` — parametri nuovi di `update_story`, scelta
  della modalità, anteprima completa
- `plugins/wm-skills/mcp/internal/preview/preview.go` — fine del troncamento a 120 caratteri,
  resa dei campi HTML come testo leggibile, avviso sul testo perso
- funzione di append (file nuovo in `internal/`, da decidere nel piano) e i suoi test
- `plugins/wm-skills/mcp/internal/tools/handlers_test.go`, `stories_test.go`,
  `preview` test — test delle modalità e dell'anteprima
- `plugins/wm-skills/bin/orchestrator-mcp` — binario ricompilato
- `plugins/wm-skills/skills/wm-review-ticket/SKILL.md` — Fase 6b
- `plugins/wm-skills/skills/wm-plan/SKILL.md` — `## Orchestrator`, `ticket: aggiornamenti-espliciti`,
  Checklist `update-context: orchestrator`
- `plugins/wm-skills/shared/orchestrator-fallback.md`
- `docs/knowledge/orchestrator-integrazione.md`
