# Note — Server MCP per Orchestrator

## Deviazioni dal piano

- **Il requisito "campi e tipi letti dalla specifica OpenAPI" non è stato implementato**, ed è risultato non necessario: il pacchetto `internal/spec` è stato eliminato interamente (`spec.go`, `spec_test.go`, `testdata/`), insieme al campo `Deps.Spec` e al caricamento della specifica in `main.go`. Campi e tipi vengono dalle strutture Go dichiarate a mano; gli elenchi dei valori ammessi (`type`, `status`) vengono dagli enum PHP letti con espressioni regolari (`internal/enums`), non dalla specifica.
- **Task 11-bis anticipato prima del Task 10** (invece che dopo): l'avviso sulle scritture che notificano il cliente (`previewNotice` in `update_story`) è stato completato prima del binario sottoposto al collaudo del cancello, così il binario collaudato al Task 11 era già completo e non è stato ricompilato subito dopo.
- **Creazione ticket senza il campo `tags`**: in `wm-plan` (tag-mode e `caso-a-split-execution`) e in origine nel brief del Task 12, la creazione non passa più `tags: [...]` nel payload; l'associazione ai tag avviene sempre con `attach_story_to_tag` dopo la creazione, per non rischiare di sovrascrivere l'elenco completo dei tag di un ticket esistente.
- **Riferimenti a "Tipi/Status disponibili" letti da GitHub rimossi da `wm-plan`**: non più necessari perché lo schema JSON di `create_story`/`update_story` espone già l'enum ammesso (letto da `internal/enums`); riformulati come "valori ammessi dal tool".

## Difetti trovati durante l'esecuzione

- **Troncamento per byte in `preview.format`** (Task 2-5, revisione): rompeva l'UTF-8 sugli accenti italiani. Corretto.
- **L'espressione di lettura degli enum accettava solo apici singoli**, con perdita silenziosa di valori PHP scritti con doppi apici (Task 2-5, revisione). Corretto.
- **Il server non si avviava contro la specifica OpenAPI reale di produzione** (Task 10): `internal/spec.Spec` definiva `Type string` per ogni proprietà, ma la specifica pubblicata usa per campi nullable la sintassi OpenAPI 3.1 `"type": ["string","null"]`. Causa dell'eliminazione del pacchetto, non di una riparazione.
- **Parametri di ricerca concatenati senza codifica** (Task 6-9, revisione — CRITICO): i nomi dei tag Webmapp (`[RDO][CLIENTE][ANNO]N`, con parentesi quadre) e le ragioni sociali dei clienti (con `&`) producevano richieste HTTP sbagliate in silenzio, senza errore visibile. Corretto in tutti i punti di costruzione di query string usando `net/url` (`url.Values` + `Encode()`).
- **File di credenziali condiviso fra produzione e sviluppo** (scoperto dal collaudo dal vivo, non dalle revisioni): i due server leggevano lo stesso file, quindi non erano usabili in parallelo con identità diverse. Corretto con un flag `--auth-file` esplicito (`client.NewWithAuthPath`), file separato per l'ambiente dev.
- **JSON annidato nelle risposte dei tool** (scoperto dal collaudo dal vivo su `get_story`): il corpo già-JSON di Orchestrator veniva incollato come stringa dentro il campo di testo della risposta MCP, poi l'intero risultato veniva serializzato di nuovo — doppio livello di escape, difficile da leggere per le skill che consumano la risposta. Corretto con `internal/tools/output.go` (`jsonOrText`, `dataResult`): il dato deserializzato finisce in `structuredContent`, il messaggio discorsivo resta testo puro.

## Decisioni prese

- Vedi `CLAUDE.md` → `## Decisioni architetturali` → `### Server MCP per Orchestrator` per le decisioni architetturali vere e proprie (Go compilato, comunicazione su stdio, enum PHP invece di OpenAPI, valori ammessi nello schema del tool, nessuna conferma rafforzata per operazioni non annullabili, server distinti per produzione/dev con credenziali separate, `curl` di ripiego).
- **Task 11 (cancello) superato**: collaudo manuale contro l'istanza locale (`localhost:8099`, utente 116 sul dump locale). `get_story 8420` letto correttamente; `update_story` senza `confirm` ha mostrato l'anteprima reale senza scrivere; `create_story` con `type: "Task"` è stato **rifiutato dallo schema** (`enum: Task does not equal any of: [Bug Feature Help desk Scrum]"`) prima ancora che la chiamata partisse — il bug originale che aveva motivato questo lavoro (oc:8503) è ora inesprimibile; `create_story` con `confirm` ha creato il ticket 8421 sul locale; `update_story` con `customer_request` ha mostrato l'avviso di notifica al cliente prima dell'anteprima. Produzione mai toccata durante il collaudo. Approvazione esplicita del dev ottenuta.
- **`attach_story_to_tag` al posto della modifica del campo `tags`**: quel campo sostituisce l'intero elenco e può cancellare tag già presenti — difetto già presente nelle skill prima di questo lavoro, corretto come effetto collaterale.
- **Dimensione del binario**: 8,2 MB anziché "intorno a 10 MB" stimato nel brief — differenza plausibile per via di `-trimpath -ldflags="-s -w"`, non indagata oltre perché non bloccante.

## Cose da fare in futuro

- **`GET /stories` manca lato Orchestrator** (oggi tutte le altre risorse hanno un elenco con filtri, le storie no): va aperto un ticket separato su `webmappsrl/orchestrator`. Serve per l'uso diretto in chat, non è richiesto dalle skill attuali.
- **Delegare la fase `environment-setup` di `wm-plan` a un subagente**: oggi eseguita nel contesto principale; valutare se isolarla come già avviene per `challenge` e per `execution: review-gate`, per coerenza col resto del workflow.
- **Nucleo condiviso fra le skill per le convenzioni non legate all'API** (struttura dei documenti, convenzioni di commit e rami): esplicitamente out of scope in questo lavoro, da riprendere come progetto distinto.
- **Server MCP ospitato su Orchestrator** invece che locale: risolverebbe la proprietà del codice e l'installazione, ma richiede lavoro su un servizio in produzione. Se il server locale dimostra il suo valore, valutare la migrazione, mantenendo gli stessi tool.
- **Miglioramento a monte su Scramble**: far sì che la specifica OpenAPI esponga gli enum di `type`/`status`, così da non dipendere dalla lettura dei file sorgente PHP.
