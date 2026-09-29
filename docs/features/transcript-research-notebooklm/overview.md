# wm-transcript-research su NotebookLM

## Cosa cambia

`wm-transcript-research` smette di leggere le trascrizioni da solo e le fa leggere a NotebookLM,
organizzato in **un notebook per giorno** e **un notebook per tag**, nell'account NotebookLM di
ciascun dev. Una call si carica una volta sola per dev: da lì in poi interrogarla costa solo una
domanda. L'agente interroga i notebook che servono in parallelo e mette insieme le risposte in
ordine di data.

**I notebook.**

- `scrum AAAA-MM-GG` — le call dello scrum di quella giornata, dalla cartella Drive delle
  trascrizioni. Si crea la prima volta che serve e si completa con le call arrivate dopo.
- `tag <nome del tag>` — le trascrizioni del tag (righe `**Fonte:**` e Google Doc linkati nel tag).
  Lo crea `wm-tag` alla creazione del tag e ne scrive il link nella descrizione; per un tag che non
  lo ha, si crea alla prima richiesta.

Ogni notebook riceve alla creazione, con `chat_configure`, il **template del team**, scritto nel
prompt dell'agente: una base comune (citazioni verbatim con chi parla e marcatore, niente
ricuciture, etichette `deciso` / `proposto` / `obiezione` / `fatto riferito`, ordine di data, «non
determinabile») più una parte per i notebook dei tag (cliente e Webmapp distinti, etichette
`richiesta` e `da chiarire`). Il template detta regole, **mai un formato**: un formato imposto fa
perdere a NotebookLM i riferimenti veri alle fonti (provato il 23/09). Le stesse regole valgono per
le domande che il dev fa direttamente dal link di un notebook.

**Quando si usano, in `wm-plan`.**

1. **All'avvio**, prima di qualsiasi domanda, parte in background la preparazione del notebook del
   giorno: creato se manca, completato con le call di oggi non ancora caricate. Le altre richieste
   all'agente aspettano che abbia finito, perché due richieste insieme creerebbero lo stesso notebook
   due volte.
2. **Caso B (nuovo ticket):** dopo la descrizione del dev, e prima di formulare la bozza, l'agente
   chiede al notebook del giorno cosa si è detto su quell'argomento; la bozza parte da lì, con le
   citazioni. Altri giorni o tag solo su richiesta esplicita del dev.
3. **`reverse-interaction`:** le domande vanno ai notebook della finestra — un notebook per ogni
   giorno con almeno una call, da qualche giorno prima della creazione del ticket a oggi, creati o
   completati se mancano — e ai notebook dei tag del ticket **che esistono già** (un tag senza
   notebook si segnala al dev, che può chiederne la creazione). Quando l'API di Orchestrator esporrà
   i giorni in cui il ticket è stato in `progress` (oc:8636), la finestra si calcolerà su quelli.
   Tutte le domande in una sola chiamata per notebook, i notebook in parallelo.
4. **Domande dirette** («cosa si è deciso su questo ticket?», «guarda cosa si è deciso in call»):
   si allarga ai giorni o ai tag indicati, creando i notebook che mancano.

**Attribuzioni.** La data di una citazione viene dal notebook stesso (un giorno, o il tag); titolo e
id Drive dalla tabella di `source_list_drive` a partire dal `source_id` del riferimento, mai dal
testo della risposta; ogni citazione si confronta con il suo `cited_text`. `wm-plan` e `wm-tag`
verificano poi ogni citazione con una ricerca Drive sulla frase
(`plugins/wm-skills/shared/verifica-citazioni.md`).

**Il merge** lo fa l'agente: una sezione di risposte per notebook, in ordine di data, con i
ribaltamenti fra giorni diversi segnalati. È solido perché ogni risposta è datata per costruzione.

**`wm-tag`** continua a leggere la trascrizione e a scrivere le macro aree come oggi; in più crea il
notebook del tag, ne mette il link nella descrizione accanto alla riga `**Fonte:**`, verifica le
citazioni del «Cosa» prima di creare il tag, e usa l'agente per le domande successive sulla call.

**Se NotebookLM non è disponibile** (server non avviato, credenziali scadute), l'agente restituisce
`RICERCA FALLITA: NotebookLM non disponibile (<motivo>)` con l'indicazione `! nlm login` quando il
motivo sono le credenziali; `wm-plan` lo dice al dev in modo evidente e prosegue senza trascrizioni.

**Pulizia.** I notebook del giorno più vecchi di 90 giorni si propongono al dev per la
cancellazione, sempre con conferma; i notebook dei tag restano finché esiste il tag.

**Distribuzione.** Il server NotebookLM lo dichiara il plugin (`notebooklm` nel suo `.mcp.json`);
ogni dev installa `notebooklm-mcp-cli` e fa `nlm login` con l'account `@webmapp.it`.

## Perché

Il 23/09 l'agente, leggendo le call direttamente da Drive, ha attribuito citazioni alla call
sbagliata in due modi diversi:

- quando una call supera il limite di output dei tool MCP (circa 25k token, cioè tutti gli scrum
  lunghi, da un'ora in su), Claude Code non gliela consegna: l'agente ha citato quella call con il
  testo di un'altra;
- anche su call corte, lette senza errori, ha scambiato gli id fra l'elenco dei risultati e i
  documenti letti, perché il testo restituito da Drive non contiene titolo né data, e ne è uscito un
  «ribaltamento» di una decisione che non c'è mai stato.

Le regole scritte nel prompt per evitarlo non sono state rispettate nelle prove. Sulle stesse call
NotebookLM ha dato 11 citazioni su 11 alla call giusta, ciascuna con l'id della fonte e il testo
originale del passaggio, a circa 9k token contro i 69–112k dell'agente.

Con un notebook per giorno e per tag ogni call si carica una volta sola per dev, i notebook restano
piccoli, e la data di ogni citazione è data dal notebook.

## Requisiti

**Agente `wm-transcript-research`**

- [ ] Nel frontmatter: `search_files` di Drive e i tool di NotebookLM che servono (elenco e
      creazione dei notebook, `chat_configure`, `source_add`, `source_list_drive`, `notebook_query_start`, `notebook_query_status`);
      non `read_file_content`, non `Bash`.
- [ ] Notebook `scrum AAAA-MM-GG` e `tag <nome del tag>`, trovati con `notebook_list`, con il
      template applicato alla creazione; i notebook del giorno si creano se mancano, quelli dei tag
      solo come indicato sotto.
- [ ] Modalità **prepara** (usata all'avvio di `wm-plan`): crea o completa il notebook del giorno e
      non fa domande.
- [ ] Elenco delle call di un giorno con `search_files` nella cartella delle trascrizioni, filtrando
      sulla sola parte `AAAA/MM/GG` del titolo, `excludeContentSnippets: true`, `pageSize: 100`,
      tutte le pagine (`nextPageToken`); si aggiungono le call che il notebook non ha.
- [ ] Fonti del tag: righe `**Fonte:**` e Google Doc per id; Fogli Google, siti web e GitHub esclusi
      e dichiarati come documenti collegati non caricati. Il notebook di un tag si crea solo su
      richiesta esplicita («crea se manca») o da `wm-tag`; altrimenti un tag senza notebook si salta
      e si dichiara.
- [ ] Finestra: solo i giorni con almeno una call; provvisoriamente per date, in attesa di oc:8636.
- [ ] Tutte le domande in una sola domanda per notebook; notebook diversi in parallelo, facendo partire
      tutte le domande con `notebook_query_start` e leggendole poi con `notebook_query_status`; una
      seconda chiamata solo per i punti scoperti; domanda più stretta se la risposta supera il limite
      di output.
- [ ] Prima di un «non determinabile», una seconda domanda mirata.
- [ ] Attribuzioni: data dal notebook, titolo e id Drive da `source_list_drive`, controllo sul
      `cited_text`; una citazione che non combacia ha una seconda domanda solo su quel punto, e si
      scarta se neanche quella combacia; una battuta interrotta si segnala.
- [ ] Merge in ordine di data, con i ribaltamenti fra notebook diversi segnalati.
- [ ] Risposta nel formato di sempre; `Copertura` con i notebook interrogati e, per ciascuno, le call
      dell'elenco e quelle caricate; i link dei notebook.
- [ ] Segnala i notebook `scrum …` più vecchi di 90 giorni; non cancella mai.
- [ ] Doppioni: dopo aver creato un notebook rilegge l'elenco; con più notebook dello stesso nome usa
      il più vecchio e riporta gli altri sotto `Notebook doppi:`.
- [ ] NotebookLM non disponibile: `RICERCA FALLITA` con il motivo e `! nlm login` se servono le
      credenziali.

**Skill `wm-plan`**

- [ ] All'avvio, in background, l'agente in modalità prepara sul notebook del giorno; le altre
      richieste all'agente aspettano che abbia risposto.
- [ ] Notebook doppi segnalati dall'agente: `wm-plan` propone di cancellarli in chiusura, col sì del
      dev.
- [ ] Caso B: dopo la descrizione del dev, domanda al notebook del giorno prima della bozza; altri
      giorni o tag solo su richiesta esplicita.
- [ ] `reverse-interaction`: all'agente la finestra (dai giorni prima della creazione del ticket a
      oggi) e i nomi dei tag del ticket; per un tag che ha il link del notebook, quel notebook.
- [ ] Domande dirette sulle call: l'agente con i giorni o i tag indicati.
- [ ] Verifica di ogni citazione con `shared/verifica-citazioni.md`.
- [ ] Alla prima ricerca, i link dei notebook al dev, con l'invito a interrogarli da sé.
- [ ] Avviso evidente se la ricerca non è disponibile.
- [ ] A fine workflow, proposta di cancellare i notebook del giorno segnalati come vecchi, con
      conferma.

**Skill `wm-tag`**

- [ ] Alla creazione del tag, lettura della trascrizione e macro aree come oggi.
- [ ] Notebook `tag <nome del tag>` creato tramite l'agente, link nella descrizione accanto alla riga
      `**Fonte:**`.
- [ ] Verifica delle citazioni del «Cosa» prima di creare il tag.
- [ ] Domande successive sulla call tramite l'agente sul notebook del tag.

**Plugin e regola condivisa**

- [ ] Server `notebooklm` nel `.mcp.json` del plugin.
- [ ] `plugins/wm-skills/shared/verifica-citazioni.md`.

**Documentazione**

- [ ] `docs/howto/installare-notebooklm.md`: installazione, `nlm login` con l'account `@webmapp.it`,
      rimozione di un server NotebookLM configurato a mano (nasconde quello del plugin), controllo
      con `/mcp`; reinstallazione del plugin dopo una modifica al `.mcp.json`.
- [ ] `docs/knowledge/wm-transcript-research.md` aggiornata.
- [ ] `docs/howto/provare-wm-transcript-research.md`: prove sul caso del 23/09 (su più notebook del
      giorno), su oc:8543 (notebook del giorno più notebook del tag 678), sul caso B, sulle citazioni
      di un tag, su una domanda senza risposta. Le prove si fanno aprendo la sessione nel repo del
      prodotto, non in `claude-marketplace`: claude-mem vi inietta il contesto del lavoro sulle
      skill.
- [ ] `CLAUDE.md`: riga della procedura di installazione e riga dei coupling.

## Rischi

- **`notebooklm-mcp-cli` non è un prodotto ufficiale Google** (reale): un aggiornamento di
  NotebookLM può romperlo, e le credenziali scadono (è successo il 23/09). Mitigato dal fallimento
  dichiarato.
- **Un'installazione in più per ogni dev** (reale). Mitigato dal howto e dal server dichiarato dal
  plugin.
- **Caricamento ripetuto per ogni dev** (reale): i notebook stanno nell'account di ciascuno. Dopo i
  primi giorni ognuno carica solo il giorno nuovo.
- **Tempi di risposta di NotebookLM** (reale): 20–60 secondi per notebook di un giorno. Claude Code
  esegue una alla volta le chiamate allo stesso server, perché il server non dichiara i tool di sola
  lettura: con `notebook_query` dieci notebook costano circa sei minuti e mezzo. Mitigato con
  `notebook_query_start`, che risponde subito, e `notebook_query_status` per leggere i risultati.
- **Merge fatto dal modello** (reale): un ribaltamento fra due giorni si vede solo confrontando le
  risposte. Mitigato dalla data per costruzione di ogni risposta.
- **Citazioni ricucite** da NotebookLM: mitigato dal controllo sul `cited_text` e dalla verifica su
  Drive.
- **Il comportamento di Claude Code non è un contratto** (limite di output, server MCP letti dalla
  copia installata del plugin): le regole di fallimento non dipendono da quei dettagli.

## Out of scope

- Condivisione dei notebook fra dev.
- Rilascio di una nuova versione di `wm-skills`: segue la checklist di rilascio, a parte.

## Moduli toccati

Tutto nel repo `claude-marketplace`:

- `plugins/wm-skills/agents/wm-transcript-research.md`
- `plugins/wm-skills/.mcp.json`
- `plugins/wm-skills/shared/verifica-citazioni.md`
- `plugins/wm-skills/skills/wm-plan/SKILL.md` — avvio, caso B, `reverse-interaction`, chiusura
- `plugins/wm-skills/skills/wm-tag/SKILL.md`
- `docs/knowledge/wm-transcript-research.md`
- `docs/howto/installare-notebooklm.md`
- `docs/howto/provare-wm-transcript-research.md`
- `CLAUDE.md`

- `docs/guide/wm-plan-diagramma/index.html` — solo contenuto: nodo di `wm-transcript-research`,
  freccia da `Fase: ticket`, paragrafi di `Fase: ticket` e `reverse-interaction`. Le fasi non
  cambiano.
