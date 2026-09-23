# Integrazione con Orchestrator

## Stato attuale

I ticket Webmapp vivono su **Orchestrator** (`webmappsrl/orchestrator`), piattaforma Laravel
interna. Le skill ci parlano attraverso il server MCP `orchestrator`, distribuito col plugin:
non costruiscono chiamate HTTP a mano.

### L'API è costruita per queste skill, non è un servizio terzo

Gli endpoint sotto `/api` — documentati in OpenAPI su
<https://orchestrator.maphub.it/docs/api.json> — sono nati per `wm-plan` e le altre skill.
Tre conseguenze operative:

- **Se manca un endpoint, si aggiunge.** Non si costruiscono aggiramenti lato skill: si apre un
  ticket su `webmappsrl/orchestrator` e si estende l'API. Esempio noto: manca `GET /stories`
  (lista e ricerca), mentre `/tasks`, `/quotes` e `/customers` hanno tutti un index con filtri
  e ordinamento.
- **Chi modifica l'API modifica il contratto delle skill.** Un cambio di campi, enum o regole di
  autorizzazione va propagato nello stesso giro alle skill che lo consumano.
- **La specifica OpenAPI è la fonte autoritativa** per campi, tipi ed enum quando serve la
  superficie completa — preferibile a leggere i singoli file PHP.

### Formato dei campi testuali

`Story.description` e `Story.customer_request` sono **HTML**: l'editor Nova è Tiptap, e i dati
in produzione contengono `<p>`, `<pre><code>`, `<a href>`. Inviare Markdown grezzo lo fa
apparire letteralmente come testo.

`Tag.description` è invece **Markdown**: l'editor è MarkdownTui, e i dati reali sono Markdown
grezzo. Nessuna conversione.

Nessuno dei due campi viene sanitizzato lato backend — `StoryApiRequest` e `TagApiRequest`
validano solo `string` generico. Un errore di formato non produce un errore API: si traduce in
un rendering sbagliato nell'editor, quindi va rispettato per convenzione e non per vincolo.

In scrittura i due campi di una Story si comportano in modo diverso. `customer_request` passa da
`addResponse()`, che aggiunge il testo in testa con autore e data e notifica il cliente.
`description` invece **si sovrascrive per intero** da oc:8549, in produzione dal 15 settembre
2026: prima passava da `addDevNote()`, che aggiungeva in testa.

Per questo il server MCP distingue le due operazioni. `update_story` con `description`
sostituisce, e serve solo per riformattare o invalidare il campo. Con `prepend` e `annotations`
aggiunge: legge la `description`, mette il blocco in testa e le etichette dopo frasi esatte,
lascia identico il resto e manda una sola PATCH insieme agli altri campi. Il testo esistente non
passa mai dal modello: una copia fatta dal modello su testi lunghi può saltarne dei pezzi, e
l'anteprima troncata di allora non lo mostrava.

## Come ci siamo arrivati

- **Chiedere all'agente di ricopiare la `description`** (PR #20, superata): funzionava su testi
  brevi, ma su ticket da 18.000 caratteri una copia del modello può riassumere, e nessuno lo vede.
- **Un endpoint di Orchestrator per aggiungere in testa**: scartato, la PATCH esiste già e il
  modo di usarla è una scelta degli strumenti; e oc:8549 aveva escluso un'opzione per scegliere
  fra aggiungere e sostituire.
- **Chiamate HTTP costruite a mano dentro le skill** (oc:7961) — il primo approccio: ogni skill
  sapeva quali endpoint chiamare e con quali campi. Caduto: la superficie dell'API si ripeteva
  in tre `SKILL.md` e invecchiava in silenzio ad ogni cambio del backend. Sostituito dal server
  MCP con i tool tipizzati.

- **Il server MCP è un binario Go compilato, non un ambiente da installare**: viaggia dentro il
  plugin, quindi nessuno deve installare nulla — un eseguibile di poche decine di MB contro
  l'ambiente di esecuzione che un runtime interpretato si porterebbe dietro.
- **Comunica su stdio, senza mettersi in ascolto su una porta**: esclude per costruzione ogni
  conflitto con i container Docker del team.
- **Produzione e istanza locale sono due server MCP distinti, con nomi diversi** (`orchestrator`
  e `orchestrator-dev`): il plugin distribuito dichiara solo `orchestrator`, fisso sulla
  produzione; `orchestrator-dev` esiste solo nel `.mcp.json` di questo repo, per il collaudo.
  Nomi diversi rendono l'ambiente visibile già nella richiesta di autorizzazione, senza doverlo
  stampare a parte. Il file delle credenziali è stato reso configurabile (`--auth-file`, non più
  fissato dentro il client) dopo un difetto emerso nel collaudo dal vivo: i due server leggevano
  lo stesso file e non potevano essere usati in parallelo con identità diverse.
- **Gli elenchi dei valori ammessi finiscono nello schema del tool, non solo in un controllo
  interno**: verificato che l'SDK MCP accetta uno schema costruito a runtime, un valore fuori
  elenco diventa inesprimibile per costruzione — rifiutato prima ancora che la chiamata parta,
  invece di dipendere da una convalida che si può dimenticare di eseguire.
- **Nessuna conferma rafforzata per le scritture non annullabili** (eliminare un preventivo,
  creare un link PDF pubblico, scrivere `customer_request`): un meccanismo a codice con scadenza
  richiederebbe uno stato lato server, e nessun parametro derivabile dai dati può fermare
  l'agente che quei dati li ha appena letti — darebbe una falsa sicurezza. Al suo posto: il
  titolo della risorsa colpita viene portato nei parametri a scopo informativo dentro la
  richiesta di autorizzazione, e la regola — questi tool non vanno mai fra quelli approvati in
  automatico — resta una responsabilità umana, non tecnica.
- **Gli elenchi dei valori ammessi (`type`, `status`) si leggono dagli enum PHP, non dalla
  specifica OpenAPI**: verificato in esecuzione che la specifica generata da Scramble non li
  espone — quei due campi risultano senza tipo e senza elenco. Il server legge direttamente
  `StoryType.php` e `StoryStatus.php`.
- **Il pacchetto `internal/spec` è stato eliminato, non riparato**: doveva leggere l'intera
  specifica OpenAPI per campi e tipi, ma la specifica di produzione usa per alcuni campi la
  sintassi OpenAPI 3.1 `"type": ["string","null"]`, che non interpretava — e il server si
  rifiutava di avviarsi per un dato che nessuna riga del programma leggeva. Campi e tipi
  vengono dalle strutture Go dichiarate a mano.
- **Le vecchie istruzioni `curl` sopravvivono come ripiego** in
  `plugins/wm-skills/shared/orchestrator-fallback.md`, letto dalle skill solo su richiesta e
  solo se i tool MCP non rispondono: costo zero nelle sessioni normali, aggiramento manuale
  disponibile se il server non parte.
