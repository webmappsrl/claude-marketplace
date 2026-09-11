# Server MCP per Orchestrator

## Cosa cambia

Le skill `wm-skills` smettono di parlare con Orchestrator tramite comandi `curl` scritti dentro i `SKILL.md` e usano un server MCP dedicato, distribuito come parte del plugin `wm-skills`.

Il server è un tramite: prende una chiamata, la gira all'API di Orchestrator e riporta la risposta in forma leggibile. Non contiene logica di lavoro — regole di autorizzazione, ruoli, notifiche al cliente, stati dei preventivi restano dentro Orchestrator.

## Perché

Oggi ogni regola sul modo di chiamare Orchestrator è una frase in `SKILL.md` che Claude deve ricordarsi di seguire: leggere l'elenco dei tipi ammessi prima di costruire un messaggio, mostrare l'anteprima prima di scrivere, non inventare campi. Sono garanzie di comportamento, non di struttura, e si rompono in silenzio: la stessa regola va ripetuta in tre skill e basta dimenticarne una.

Con un server MCP le stesse garanzie diventano strutturali. Un valore fuori dall'elenco ammesso non è esprimibile perché lo schema lo rifiuta; l'anteprima non è una tabella ricostruita a memoria ma la differenza calcolata dal server sui dati veri; l'autenticazione non è una procedura da ripetere in ogni skill ma un dettaglio interno.

Beneficio secondario, non trascurabile: la sezione `Orchestrator API` di `wm-plan` (139 righe) esiste quasi interamente per spiegare a Claude cose che il server farebbe da sé, e si riduce a poche righe.

## Requisiti

- [ ] Server MCP in Go, compilato per `darwin/arm64`, binario versionato nel repo dentro il plugin `wm-skills`
- [ ] Comunicazione sui canali standard del processo (stdio): nessuna porta in ascolto, nessun conflitto possibile con i container Docker del team
- [ ] Autenticazione leggendo `~/.config/webmapp/orchestrator-auth.json`, riletto a ogni chiamata; il server non esegue mai l'accesso e non chiede mai la password
- [ ] **Nessun interruttore di ambiente nel plugin distribuito**: il `.mcp.json` del plugin dichiara un solo server, fisso sulla produzione. Solo il repo `claude-marketplace` dichiara per conto suo un secondo server `orchestrator-dev` puntato all'istanza locale, che non viene mai distribuito. I due hanno nomi diversi, quindi i tool hanno nomi diversi e l'ambiente è visibile nella richiesta di autorizzazione senza doverlo stampare
- [ ] Tool dichiarati a mano (nome, descrizione, gruppo, se è scrittura); campi e tipi letti dalle strutture Go che descrivono le risorse Orchestrator, non da una specifica scaricata
- [ ] **I valori ammessi di `type` e `status` si leggono dagli enum PHP** (`StoryType.php`, `StoryStatus.php` su GitHub), non dalla specifica OpenAPI: verificato che Scramble non li espone — nella specifica quei due campi risultano senza tipo e senza elenco
- [ ] Gli elenchi di valori ammessi finiscono **nello schema del tool**, non solo in una convalida interna: è lì che un valore inventato smette di essere esprimibile. Verificato che l'SDK permette uno schema costruito a runtime
- [ ] Tool di lettura senza conferma
- [ ] Tool di scrittura con parametro `confirm` (valore predefinito `false`): senza conferma mostrano la differenza e non scrivono, con conferma scrivono
- [ ] Le scritture con effetti non annullabili — eliminazione di un preventivo, collegamento PDF pubblico — portano nei parametri il **titolo del preventivo a scopo informativo**, perché la richiesta di autorizzazione mostri cosa si sta colpendo invece di un solo numero. Non è una barriera: nessun parametro derivabile dai dati può fermare l'agente che li ha letti. La barriera reale è l'approvazione umana del programma, e questi tool non vanno mai messi fra quelli approvati in automatico
- [ ] Gruppi di tool attivabili da configurazione (variabile d'ambiente); `get_story`, `create_story`, `update_story` e `me` sempre caricati
- [ ] `customer_request` fa scattare la notifica al cliente: la descrizione del tool lo dichiara esplicitamente, e vale la stessa regola — mai fra i tool approvati in automatico
- [ ] Errori tradotti in messaggi utilizzabili, distinguendo dati rifiutati, mancanza di autorizzazione, risorsa inesistente e servizio irraggiungibile
- [ ] Prove isolate su convalida, anteprima, traduzione degli errori e lettura delle credenziali
- [ ] Prove con risposte registrate di Orchestrator
- [ ] **Cancello:** collaudo manuale del binario contro l'istanza locale, approvato dal dev, prima di toccare le skill
- [ ] Riscrittura di `wm-plan`, `wm-tag` e `wm-review-ticket` sui tool, con sfoltimento della sezione `Orchestrator API`
- [ ] Le istruzioni `curl` attuali vengono spostate in `plugins/wm-skills/shared/orchestrator-fallback.md`, letto su richiesta e solo se i tool non rispondono
- [ ] Lista di controllo del rilascio aggiornata: il binario va ricompilato a ogni versione

## Superficie dei tool

| Gruppo | Tool | Attivo |
|---|---|---|
| `stories` | `get_story`, `create_story`, `update_story` | sempre |
| `me` | `me` | sempre |
| `tags` | `list_tags`, `get_tag`, `create_tag`, `update_tag`, `attach_story_to_tag`, `detach_story_from_tag` | su richiesta |
| `tasks` | `list_tasks`, `get_task`, `create_task`, `update_task` | su richiesta |
| `crm` | clienti e preventivi in lettura e scrittura, prodotti, collegamenti PDF pubblici | su richiesta |

`attach_story_to_tag` sostituisce il modo attuale di associare i tag, che riscrive l'intero elenco con una modifica del campo `tags` e può cancellare tag già presenti.

## Rischi

**Il beneficio è di disciplina, non di funzionalità.** Il server non permette di fare nulla che oggi sia impossibile: rende affidabile ciò che oggi dipende dal fatto che Claude segua le istruzioni. Se il valore atteso fosse "nuove capacità", il progetto deluderebbe.

**Costo di contesto in ogni sessione.** Le definizioni dei tool sempre attivi pesano qualche centinaio di parole in ogni sessione, anche in quelle che non toccano Orchestrator. Accettato consapevolmente: chi apre `wm-plan` sta quasi sempre lavorando su un ticket, e un tool caricato su richiesta rischia di non essere caricato quando serve.

**Proprietà del codice.** Il Go non è familiare al team: il server nasce scritto e manutenuto da Claude. Mitigato dal fatto che la logica è minima e che Go si legge quasi come pseudocodice, ma resta un punto da ricordare se un giorno il server dovesse crescere.

**I valori ammessi si leggono da file sorgente PHP, non da un contratto dichiarato.** La specifica OpenAPI non espone gli enum, quindi il server legge `StoryType.php` e `StoryStatus.php` con espressioni regolari. Se un giorno quegli enum cambiassero forma, la lettura smetterebbe di funzionare in silenzio e la convalida tornerebbe a essere assente. Mitigazione: la convalida interna resta come rete, e il server segnala quando non riesce a leggere gli elenchi. Miglioramento a monte possibile ma non dipendenza di questo lavoro: far sì che Scramble esponga gli enum nella specifica.

**Il pacchetto che leggeva la specifica OpenAPI per campi e tipi è stato eliminato, non riparato.** La specifica reale di produzione usa per alcuni campi la sintassi OpenAPI 3.1 `"type": ["string","null"]`, che quel pacchetto non interpretava: il server si rifiutava di avviarsi per caricare un dato — `deps.Spec` — che nessuna riga del programma leggeva davvero. Campi e tipi vengono invece dalle strutture Go dichiarate a mano; nessuna specifica viene scaricata o letta all'avvio.

**Modifiche contemporanee non rilevate.** Fra il calcolo dell'anteprima e la conferma, un'altra persona può modificare lo stesso ticket da Nova: si approverebbe una differenza e se ne applicherebbe un'altra. Rischio accettato consapevolmente — siamo cinque, la finestra è di secondi, e oggi con i comandi a mano il problema esiste identico senza che sia mai stato un disturbo.

**Un endpoint nuovo non diventa un tool da solo.** Conseguenza voluta della scelta di dichiarare i tool a mano: la specifica tiene allineati campi e valori ammessi, non la superficie. Un endpoint nuovo va esposto da qualcuno che decida nome, gruppo e descrizione.

**Il server potrebbe non partire** su una macchina con il plugin aggiornato. Il fallimento deve essere rumoroso e con un messaggio chiaro. Le istruzioni `curl` non vengono buttate: si spostano in `shared/orchestrator-fallback.md`, che le skill leggono **solo** quando i tool non rispondono. Costo zero nelle sessioni normali, aggiramento manuale disponibile quando serve.

## Out of scope

- **`GET /stories`**, oggi assente mentre tutte le altre risorse hanno un elenco con filtri. Serve per l'uso diretto in chat, non per le skill: va aperto un ticket separato su `webmappsrl/orchestrator`.
- **Server MCP ospitato su Orchestrator.** Valutata e rimandata: risolverebbe la proprietà del codice e l'installazione, ma richiede lavoro sul servizio in produzione. Se il server locale dimostra il suo valore, la migrazione è un secondo tempo e i tool restano gli stessi.
- **Scrittura verso la produzione durante le prove.** Il collaudo avviene contro `orchestrator-dev`, dichiarato solo in questo repo. Quando il comportamento è quello atteso si compila il binario e si rilascia il plugin, che parla solo con la produzione.
- **Accettazione dei certificati non validi come comportamento predefinito**, per l'istanza di sviluppo che oggi ne è priva: va richiesta esplicitamente ogni volta.
- **Nucleo condiviso fra le skill** per ciò che non riguarda l'API (struttura dei documenti, convenzioni di commit e rami): lavoro distinto, da riprendere dopo.

## Moduli toccati

**Nuovi**, dentro `plugins/wm-skills/`:
- sorgente Go del server, binario compilato, script di compilazione
- dichiarazione del server nel manifesto del plugin

**Modificati:**
- `plugins/wm-skills/skills/wm-plan/SKILL.md` — tool al posto dei `curl`, sfoltimento di `Orchestrator API`
- `plugins/wm-skills/skills/wm-tag/SKILL.md` — tool al posto dei `curl`
- `plugins/wm-skills/skills/wm-review-ticket/SKILL.md` — tool al posto dei `curl`
- `plugins/wm-skills/.claude-plugin/plugin.json` — versione
- `CLAUDE.md` — lista di controllo del rilascio, voce nelle funzionalità disponibili, decisioni architetturali
