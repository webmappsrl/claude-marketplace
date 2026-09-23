# Notes — wm-transcript-research su NotebookLM

## Deviazioni dal piano

- **Nessun branch dedicato.** Il lavoro è su `develop`: il `CLAUDE.md` del repo vieta di creare
  branch dentro un lavoro, e prevale sulla fase `execution: branch` di `wm-plan`.
- **Il template dei notebook sta nel prompt dell'agente**, non in `shared/`: l'agente non ha tool per
  leggere file, e non deve averne (un agente con `Bash` ha letto i file del repo in una prova del
  23/09). Deciso prima dell'approvazione del piano; overview allineata.
- **Controllo del Task 2:** il `grep -c` si aspettava 2 righe, ma «verifica non possibile» va a capo
  nel file; il contenuto è completo. Difetto del controllo, non del file.

- **Architettura per giorno e per tag implementata la sera del 23/09** nei file (agente, `wm-plan`,
  `wm-tag`, conoscenza, howto, diagramma); `plan.md` riscritto dopo, sulla versione implementata.
  Finestra provvisoria in attesa di oc:8636: dai giorni prima
  della creazione del ticket a oggi, solo i giorni con almeno una call. **Non ancora provata con
  l'agente vero.**

## Bug trovati

- **Elenco delle call fermo alla prima pagina di Drive** (prova 2, agente vero): `search_files`
  restituisce i risultati a pagine; l'agente ha preso la prima (6 call su 13) senza seguire
  `nextPageToken`, la call del 16/09 15:39 con la risposta è rimasta fuori, e il risultato è stato un
  «non determinabile» falso, con una `Copertura` che sembrava completa. Corretto nel prompt:
  `pageSize: 100`, tutte le pagine, e in `Copertura` il numero di call dell'elenco accanto a quelle
  caricate. Le citazioni proposte da NotebookLM su quelle 6 call sono state scartate correttamente
  dal controllo sul `cited_text`.
- **Scelta delle call a giudizio** (prova 3, agente vero): l'agente ha caricato 1 scrum su più di 25
  della finestra perché giudicava la domanda «specifica». Corretto nel prompt: le call della
  finestra si caricano tutte, l'unica esclusione è il limite di fonti.
- **Server del plugin nascosto da un server con lo stesso comando:** finché in `~/.claude.json`
  c'era `notebooklm-mcp` configurato a mano, il server `notebooklm` del plugin non compariva. Tolto
  quello manuale, compare. È il passo 3 del howto di installazione.
- **Notebook del giorno duplicati** (prove 2 e 3 lanciate insieme, architettura per giorno): le due
  ricerche non trovano il notebook e lo creano entrambe, così `scrum 2026-09-14` … `scrum 2026-09-18`
  esistono due volte. Corretto: dopo ogni creazione l'agente rilegge l'elenco, usa il più vecchio e
  riporta gli altri in `Notebook doppi:`; `wm-plan` aspetta la risposta di «prepara» prima delle
  altre richieste; i doppi segnalati si propongono da cancellare in chiusura (deciso dal dev: a
  regime capiterà di rado).
- **Citazione di Piccioli del 15/09 persa** (prova 2, architettura per giorno): NotebookLM ha dato
  un riferimento che non combaciava col `cited_text` e l'agente l'ha scartata. Corretto: seconda
  domanda solo su quel punto prima di scartarla. Verificato dopo il riavvio: nella prova 2 ripetuta
  da sola la citazione c'è, attribuita al 15/09 15:29, e i doppioni sono stati riconosciuti e
  segnalati senza crearne di nuovi.
- **Domande ai notebook eseguite una alla volta** (`wm-plan` su oc:8543, 23/09 sera): la ricerca ha
  preso 10 minuti e mezzo, di cui 6 e mezzo per 10 `notebook_query` partite insieme ma servite in
  fila, circa 40 secondi l'una. Da script fuori da Claude Code tre `notebook_query` in parallelo
  tornano insieme (32–39 secondi), quindi la fila la fa Claude Code: il server non dichiara
  `readOnlyHint` sui tool. Con `notebook_query_start` tre avvii tornano in 3 secondi e le risposte
  sono pronte insieme dopo circa 50. Corretto nel prompt: avvio con `notebook_query_start`, lettura
  con `notebook_query_status`. Provato dopo il riavvio sulla domanda della prova 2 allargata al
  14–23/09: 8 avvii in 9 secondi, tutte le 8 risposte lette 56 secondi dopo il primo avvio (prima
  sarebbero stati circa 5 minuti); ricerca intera in 1 minuto e 52 secondi, 66k token, 35 tool call;
  citazioni come nella prova 2.
- **`wm-plan` su oc:8543 con le domande avviate insieme (23/09 sera):** preparazione del notebook del
  giorno in 14 secondi (già completo); ricerca di `reverse-interaction` su 9 notebook (8 giorni più il
  tag) in 3 minuti e mezzo, di cui 1 minuto e 18 secondi di domande, contro i 10 minuti e mezzo di
  prima. Tag `wm-package` e `26Q3` senza notebook, saltati e dichiarati. `wm-plan` ha verificato su
  Drive le citazioni usate e ha trovato la contraddizione fra la call del 23/09 («la inverte sempre»)
  e il ticket (flag disattivabile). **Difetto:** l'agente ha riportato una citazione senza
  riferimento «come contesto», con speaker e ora sbagliati (16/09 12:46 con l'id della call delle
  15:39), contro la regola «se la citazione non ha riferimenti, non si riporta»; la verifica di
  `wm-plan` l'ha scoperta e non l'ha usata. Corretto nel prompt: la regola è in cima a
  `## Le attribuzioni`, senza eccezioni («non compare in nessuna parte della risposta»).
  Riprovato dopo il riavvio (`wm-plan` su oc:8543): la citazione del 16/09 torna attribuita alla
  call delle 12:46 con il suo id (`1Duzcm5k…`), speaker Giuseppe Bonfanti, verificata su Drive;
  nessuna citazione «come contesto». 10 notebook, 4 minuti e 42 secondi.
- **I server MCP del plugin si leggono dalla copia installata**, non dal repo: dopo aver modificato
  `.mcp.json` serve reinstallare il plugin (`/plugin uninstall` e `/plugin install`) e riavviare.
  La reinstallazione copia i file del repo così come sono, anche non committati.

Durante l'esecuzione dei Task 1–6, nessuno. I difetti trovati nelle prove del 23/09, prima del piano, sono già
nell'overview: formato imposto nel template che fa perdere i riferimenti, Fogli che fanno superare il
limite di output, resoconto di una call intera oltre il limite, etichette di NotebookLM inaffidabili
sulla struttura di una call intera.

## Decisioni

- **Da dove viene questo lavoro.** È partito come «rendere opzionale la ricerca sulle trascrizioni
  perché costa troppo»: misurato con `/usage`, una ricerca verbosa vale circa l'1% della finestra di
  5 ore, quindi il costo non era il problema. Le prove hanno mostrato invece attribuzioni sbagliate
  dell'agente; una prima correzione (tool Go di lettura a blocchi, regole di lettura) è stata
  scartata dal dev e ripristinata senza commit dopo che NotebookLM, già collegato come MCP, ha dato
  risultati migliori al primo colpo.
- **Autorizzazione sui dati:** il dev ha stabilito che caricare le call in NotebookLM con gli account
  `@webmapp.it` è accettabile: il team sa cosa carica, e l'accesso resta agli account di lavoro.
- **Accesso dei colleghi:** confermato dal dev; sono invitati e coorganizzatori dello scrum.
- **`wm-tag` non affida la scrittura del tag a NotebookLM:** nella prova sul tag 678 ha perso 2 macro
  aree su 13 e sbagliato le etichette in 6; resta la lettura attuale, con in più notebook del tag e
  verifica delle citazioni.

## Follow-up

**Prova di `wm-plan` su oc:8543 (23/09 sera): superata** sui cinque punti, controllati sulle tool
call: slug e id della fonte del tag passati all'agente, notebook riusato senza ricaricare le 29 call,
7 citazioni verificate su Drive da `wm-plan` prima della domanda, link del notebook mostrato, riga
«dalle call risulta» con data, ora e speaker. La ricerca ha trovato una contraddizione reale fra la
review del terzo ciclo e lo scrum del 23/09 08:27. Difetto: 10 `notebook_query`, una per domanda,
circa 9 minuti e 75k token; corretto nel prompt (tutte le domande in una chiamata). La sessione di
prova era aperta in `claude-marketplace` e quindi contaminata da claude-mem nel tono, non nei
meccanismi controllati.

**Architettura rivista la sera del 23/09** (overview riscritta, piano da aggiornare): un notebook per
giorno (`scrum AAAA-MM-GG`) e uno per tag (`tag <nome>`) nell'account di ciascun dev, preparazione
del notebook del giorno in background all'avvio di `wm-plan`, caso B che interroga il notebook del
giorno prima della bozza, merge delle risposte fatto dall'agente in ordine di data. La finestra di
`reverse-interaction` va calcolata sui giorni in cui il ticket è stato in `progress`: l'API di
Orchestrator non li espone, ticket aperto **oc:8636** («API: giorni in cui un ticket è stato in un
certo stato»). Decisione del dev: il resto del lavoro riprende quando l'endpoint esiste.

**Stato (23/09, fine serata):** prove del Task 7 fatte, esiti nel howto; lavoro pronto per
l'approvazione del dev, che decide il commit. Le prove di `wm-plan` vanno fatte aprendo Claude Code
nel repo del prodotto, non in `claude-marketplace`: claude-mem inietta all'avvio il riassunto delle
sessioni recenti dello stesso progetto, e una sessione aperta qui legge di essere un test e si
comporta da collaudatore.

**Dopo il commit:**

- `get_tag` restituisce da 70 KB a 1,4 MB perché include tutti i ticket del tag, mentre alle skill
  servono solo i link della descrizione: estendere l'API/il tool con una variante senza ticket.
- Migliorare le etichette di NotebookLM sulla struttura di una call intera, per un eventuale uso in
  `wm-tag` oltre le domande mirate.
- Rilascio di una nuova versione di `wm-skills` con la checklist di rilascio.
