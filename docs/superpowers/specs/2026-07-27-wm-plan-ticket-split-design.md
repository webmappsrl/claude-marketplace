# Split automatico di ticket Help desk multi-richiesta in wm-plan

> Nessun ticket Orchestrator — feature interna alla skill `wm-plan`.
> Caso reale di riferimento: `oc:8303` ("Quesiti sui POI", type Help desk, due richieste distinte nel `customer_request`).

## Cosa cambia

`wm-plan`, in `Fase: ticket → ticket: caso-a` (lettura di un ticket esistente), rileva automaticamente quando un ticket di tipo **Help desk** contiene più richieste distinte scritte dal cliente in un unico `customer_request`, e propone di normalizzare la situazione splittando su più ticket — mantenendo il ticket originale (rinominato e ridotto alla prima richiesta) e creando un nuovo ticket per ciascuna richiesta successiva.

## Perché

I clienti scrivono spesso ticket Help desk che in realtà contengono più richieste indipendenti (es. `oc:8303`: una richiesta UX sulla creazione POI + una domanda su una feature di gestione autonoma POI). Oggi il dev deve accorgersene manualmente e gestire lo split a mano, con perdita di tracciabilità (una sola richiesta finisce lavorata, l'altra si perde nel testo). Normalizzare subito in fase di lettura del ticket rende ogni richiesta un'unità di lavoro tracciabile indipendentemente.

## Requisiti

- [ ] Il rilevamento scatta solo per ticket con `type == "Help desk"`, subito dopo la lettura API in `caso-a`, prima di mostrare qualunque riepilogo all'utente.
- [ ] Se il ticket non è Help desk o contiene una sola richiesta, il flusso resta identico a quello attuale (nessun cambiamento visibile).
- [ ] Se vengono rilevate N ≥ 2 richieste distinte, il testo di ciascuna partizione è **verbatim** (nessuna riformulazione), estratto in ordine di apparizione nel `customer_request` originale.
- [ ] Prima di qualsiasi scrittura, viene mostrato un riepilogo unico con tabella (ruolo, titolo proposto, testo verbatim per ciascuna partizione) e richiesta esplicita di conferma per procedere con lo split. Il dev può modificare titoli o testo prima di confermare. Se rifiuta, il ticket originale resta intatto e il flusso prosegue come oggi (nessuno split).
- [ ] La prima partizione (in ordine di apparizione) resta sempre sul ticket originale: `PATCH` su `name` (titolo sintetico proposto) e `customer_request` (partizione 1 verbatim). Tag e `creator_id` del ticket originale non vengono toccati dalla PATCH.
- [ ] Per ogni partizione successiva (2..N) viene creato un nuovo ticket via `POST /api/stories` con: `name` (titolo sintetico proposto), `type: "Help desk"` (sempre — nessuna proposta di riclassificazione automatica; sarà un successivo lavoro di wm-plan sul singolo ticket a cambiarlo se necessario), `customer_request` (partizione verbatim), `creator_id` = stesso valore del ticket originale (così il cliente resta visibile come creatore nella sua area Orchestrator), `tags` = stessi tag del ticket originale.
- [ ] Ogni PATCH/POST individuale segue comunque la regola generale scritture già esistente (`## Orchestrator API → Regola generale scritture`): preview tabellare + conferma esplicita per ogni singola chiamata, anche se il contenuto è già stato approvato nel riepilogo complessivo. Nessuna eccezione per lo split.
- [ ] Dopo la creazione di tutti i ticket del gruppo, si chiede al dev se vuole raggrupparli in un tag Orchestrator (uso interno dev, stesso endpoint `POST /api/tags` già usato da `wm-tag`):
  - Nome di default: `<customername>-<titolo originale kebab-case>`, dove `<customername>` è ricavato da un tag esistente sul ticket originale riconoscibile come identificativo cliente (es. `ass_cammini_italia`). Se nessun tag è riconoscibile come cliente, si chiede al dev di indicarlo manualmente.
  - Il dev può modificare il nome prima della creazione.
  - `description` del tag = `customer_request` originale completo (pre-split, testo intero del cliente).
  - Il tag va associato a tutti i ticket del gruppo (originale + nuovi).
  - Se il dev rifiuta, si salta questo step senza creare alcun tag.
- [ ] Al termine, viene mostrato l'elenco di tutti i ticket del gruppo (originale + nuovi, con ID e titolo) e si chiede esplicitamente su quale continuare il workflow ora (oppure "nessuno" per tornare al menu A/B/C di `Fase: ticket`). La scelta prosegue in `Fase: init-context` usando i dati già noti del ticket scelto (senza rifare la GET).

## Rischi

- **Falsi positivi nel rilevamento**: un ticket Help desk con un solo argomento ma scritto in più paragrafi potrebbe essere letto come "più richieste". Mitigato dalla conferma esplicita del dev prima di qualsiasi scrittura — nessuno split avviene senza approvazione.
- **Partizionamento impreciso**: il confine tra due richieste nel testo del cliente potrebbe non essere netto (es. una frase di transizione ambigua). Il dev vede il testo verbatim di ogni partizione nel riepilogo e può modificarlo prima di confermare.
- **Tag "cliente" non riconoscibile**: l'euristica di riconoscimento del tag cliente per il nome del tag di raggruppamento può fallire silenziosamente su ticket con tag non ovvi. Mitigato dal fallback esplicito (richiesta manuale al dev), mai un default indovinato senza conferma.

## Out of scope

- Riclassificazione automatica del `type` dei nuovi ticket (restano sempre Help desk; la riclassificazione avviene in un secondo momento, quando wm-plan lavora sul singolo ticket).
- Split di ticket con `type` diverso da Help desk (Feature/Bug/Task non vengono mai analizzati per split automatico).
- Split del campo `description` (note tecniche dev) — solo `customer_request` viene partizionato.
- Coefficienti o euristiche rigide (keyword, conteggio paragrafi) per il rilevamento — il giudizio è diretto sul testo, non basato su pattern meccanici.

## Moduli toccati

- `plugins/wm-skills/skills/wm-plan/SKILL.md` — nuova sotto-fase `ticket: caso-a-split-detection` inserita in `Fase: ticket → ticket: caso-a`, prima del riepilogo ticket esistente.
