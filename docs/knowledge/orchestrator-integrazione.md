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

## Come ci siamo arrivati

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
