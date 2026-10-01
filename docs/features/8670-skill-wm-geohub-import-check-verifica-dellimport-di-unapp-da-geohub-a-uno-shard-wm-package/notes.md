> Ticket: oc:8670

# Notes — Skill wm-geohub-import-check

## Deviazioni dal piano

Vedi la sezione seguente, task per task.

## Divergenze dal piano, task per task

### Task 2 scaricamento
- Opzione `--no-immagini` di `scarica`, usata dai test: gli URL delle immagini nelle fixture sono quelli veri e i test non toccano la rete. La scelta dei campioni è provata a parte sulle fixture, lo scaricamento dei campioni solo dal vivo.
- `scarica` prende anche i nomi delle where dei POI, non solo delle tracce: servono al confronto per nome sui POI.

### Task 3 confronto
- La regola `prefissi_url` considera uguali due URL http(s) invece di normalizzarne il prefisso: i percorsi delle immagini non hanno nulla in comune (`EcMedia/6269.jpg` su Geohub, `3/media/81/5` sullo shard). Presenza, numero e campioni delle immagini restano controllati a parte.
- POI collegati a una traccia confrontati per id Geohub, ordine e nome; il resto dei loro campi solo nel `pois.geojson`. Il confronto generico ripeteva ogni differenza dei POI dentro ogni traccia.
- Liste di tassonomie abbinate per `identifier`, non per posizione.
- Un valore JSON scritto come testo sullo shard (`"[]"`, `"{…}"`, `"false"`, `"true"`, `"null"`) dove Geohub non ha nulla è `tipo_diverso`: è la forma di oc:8664. Il caso `"false"` (90 POI) è emerso solo dal vivo.
- Nomi delle where confrontati senza distinguere trattini, spazi e maiuscole («Massa Carrara» su Geohub, «Massa-Carrara» in OSMFeatures).
- Regola `where_geometria` «da segnalare» e sezione del report «Differenze attese da segnalare»; regole alla versione 2 (decisione del dev, vedi Decisioni).

### Task 5 skill
- Fase `correzioni-manuali` fra riepilogo e tag: le differenze con una possibile causa si risolvono con l'action della scaletta prima di passare a `wm-tag` (decisione del dev).
- Ricerca del tag esistente ristretta a `search: "[COLLAUDO][<APP>]"`: con `search: "COLLAUDO"` `list_tags` ha portato nel context le descrizioni di tag di altri clienti, una con credenziali in chiaro.
- Tolto `tag-correzioni.json` e il pacchetto che lo passava a `wm-tag` (decisione del dev, vedi Task 6).

### Task 6 wm-tag
- Tolta la ricerca dei ticket già esistenti prima del vaglio, con le proposte «già trattato» e «ricomparso»: i doppioni li gestisce il dev, eliminandoli e lasciando che si ricreino nel tag.
- Le differenze attese da segnalare vanno nelle note trasversali della descrizione del tag.

### Task 7 prove dal vivo
- Il passo 1 (`deploy_local_with_prod.sh`) l'ha fatto il dev; l'allineamento è stato confermato in sola lettura (oc:8663-8665 e i 20 ticket del tag 651 identici fra produzione e locale).
- Il server `orchestrator-dev` usa il file di credenziali di produzione: il database locale ripristinato dalla produzione accetta lo stesso token.
- La prova è andata avanti in tre giri, con le action lanciate dal dev su Maphub dev fra un giro e l'altro («Reindicizza Scout», «Sincronizza Taxonomy Where su EC Features»).

### Riscrittura in Go
- Lo script Python è diventato il comando Go `geohub-import-check`: sorgente in `plugins/wm-skills/mcp/cmd/geohub-import-check/`, binario versionato in `plugins/wm-skills/bin/`, compilato da `mcp/build.sh` insieme al server MCP (solo `darwin/arm64`). Stessi comandi, stessi file prodotti, stessi codici di uscita e messaggi; la skill chiama il binario.
- La suite bash usa il binario; il test sui campioni di immagini, che importava il modulo Python, passa da un comando non documentato `campioni <cartella>`. Test Go in `main_test.go` per i punti in cui Go differisce da Python: espressioni regolari senza riferimenti all'indietro (tag SVG vuoti), spazi Unicode, lunghezze in caratteri e non in byte.
- Negli esempi del report le chiavi degli oggetti sono in ordine alfabetico con `it` per prima, invece dell'ordine del file: Go non lo conserva, e mettere `it` per prima fa restare il testo italiano negli esempi tagliati a 200 caratteri. Verificato confrontando i due output sulle fixture: l'unica differenza rimasta è l'ordine delle chiavi senza testo (`forward`/`backward`).
- Un nome di where che arriva non in JSON finisce fra le non scaricate, invece di fermare lo scaricamento come faceva Python; i nomi delle where si scaricano in parallelo.
- La suite usa ancora `python3` solo per servire le fixture in HTTP (`http.server`) e trovare una porta libera: è strumentazione del test, non codice della skill.

## Bug trovati

## Decisioni

- Tag Orchestrator del ticket: nessuno proposto né associato, su scelta del dev («ignora i tag»).
- Prefisso del tag: resta `[COLLAUDO]`. Il dev ha valutato `[IMPORT-CHECK]` (in «collaudo» il team intende la verifica di una release) e poi è tornato a `[COLLAUDO]`. Resta invece il nome del materiale di `wm-tag`, «report di verifica dell'import» al posto di «report di collaudo».
- Il tag si crea solo per i problemi che restano (richiesta del dev durante la prova dal vivo): le differenze con una possibile causa si risolvono prima con l'action della scaletta, lanciata dal dev, e si rilancia la verifica. Aggiunta la fase `correzioni-manuali` alla skill. Esempio che l'ha motivata: le metriche a 0 in Elastic su 4 tracce e 125 delle 128 where mancanti in Elastic sono sparite dopo «Reindicizza Scout».

- Regola `where_geometria`, «attesa ma da segnalare» (decisione del dev sul terzo giro della verifica di app 28 → 3): su POI 2404 (Geohub Savona e Stellanello, shard Imperia e Diano Arentino, punto sul confine) e sulle tracce 31688 e 31697 (comuni Castelnuovo Magra e Levanto) le where assegnate a mano su Geohub non coincidono con quelle calcolate dalla geometria sullo shard. Non sono un errore dell'import, ma restano visibili nel report. Regole alla versione 2.
- Nessuna ricerca dei ticket già esistenti prima del vaglio (decisione del dev durante la prova, dopo l'approvazione del piano): i doppioni li gestisce il dev, eliminandoli e lasciando che si ricreino nel tag. Tolti `tag-correzioni.json` e le proposte «già trattato» e «ricomparso» di `wm-tag`.
- Correzioni manuali estese ai comandi su Laravel (decisione del dev, prima della creazione del tag): oltre alle action di Nova, ogni problema rimasto e ogni differenza da segnalare si valuta come intervento rapido con un comando artisan già esistente lanciato dal dev via SSH. Niente tinker né script scritti per l'occasione (decisione del dev): un problema che si correggerebbe solo così è un possibile ticket. Solo ciò che non si risolve così diventa una macro area del tag.
- Regole versione 3 contro il rumore (approvate dal dev): chiavi che il frontend non legge (verificato su wm-core: `gpx_url`, `kml_url`, `geojson_url` e `user_id` solo nella definizione dei tipi), numeri scritti come testo, SVG uguali a meno di spazi, apici e tag vuoti, `slope` calcolata solo dallo shard, chiavi vuote su tutte le feature. Sulla verifica di app 28 → 3 i gruppi scendono da 72 a 39. `searchable` dei POI, `taxonomyIdentifiers` ed `excerpt` restano: il frontend li usa.
- Interventi manuali limitati alla piattaforma Laravel dello shard (decisione del dev, ultima): solo action di Nova e modifiche di pochi record dall'interfaccia. Esclusi i comandi da terminale, artisan via SSH compreso, oltre a tinker e script: supera le due decisioni precedenti su artisan e tinker.
- Codice di riferimento dai repo `webmappsrl` su GitHub, sul branch che gira sull'ambiente verificato: `develop` per gli shard di sviluppo, `main` per quelli di produzione (decisione del dev; su `main` di wm-package, fermo al 07/07/2026, manca per esempio `SyncEcTaxonomyWhere`, installato su Maphub dev da `develop`): la ricerca delle cause leggeva la copia locale di wm-package (commit `eb1e01f5` del 15/09), che poteva non coincidere con il codice installato, e aveva portato a un'ipotesi sbagliata su `start`/`end`.
- Diagramma di flusso della skill, come per `wm-plan` (richiesta del dev, indicata da Giuseppe in call; la ricerca nelle trascrizioni non l'aveva trovata): `docs/guide/wm-geohub-import-check-diagramma/index.html`, stessa struttura della pagina di `wm-plan` ma con lo stile del design system Webmapp (richiesta del dev): colori `petrolio` e del logo, Montserrat Medium e Roboto, spigoli vivi, niente ombre, un solo tema chiaro. Il controllo in CI `verifica-diagramma.sh` ora vale per tutte e due le skill, e la regola `.claude/rules/wm-plan-diagramma.md` copre anche la nuova. Per renderlo possibile le fasi della skill hanno nomi senza spazi (`codice-riferimento`, `correzioni-manuali`, `tag-esistente`, `passaggio-wm-tag`).
- Procedura d'uso per il team in `docs/howto/verificare-un-import-geohub.md`: la skill la useranno Carla, Alessandro e chiunque faccia un collaudo (call del 30/09/2026 delle 12:50).
- Dalle call del 30/09: le tassonomie forse vanno confrontate solo per presenza («all'andata c'erano, al ritorno ce ne sono»), da capire meglio; la sezione «Differenze attese da segnalare» resta per le where OSMFeatures; i due POI presenti solo sullo shard si ignorano.
- Fase `pulizia` (decisione del dev): finito `wm-tag`, o se il dev non gli passa il report, la skill cancella la copia del codice (circa 260 MB nella verifica di app 28 → 3) e lascia report, differenze e file scaricati.
- Codice di riferimento al commit fissato da chi pubblica, non più al branch (decisione del dev, dopo un test con un agente): il frontend è un solo bundle per tutti gli shard, pubblicato da `wm-webapp` `main` (`scripts/deploy-default.js:2-5`), che fissa wm-core a `b01b05e`; Maphub, sia su `main` sia su `develop`, fissa wm-package a `a685012` (29/09/2026), che non è né `main` (`b201c51`, 07/07) né `develop` (`fbae771`, 30/09) di wm-package. Supera la regola del branch secondo lo shard.
- Riepilogo corto al dev (decisione del dev, dopo un test con un agente): l'elenco completo dei gruppi di differenze, cioè lo stdout di `confronta`, non si mostra durante la verifica. Al dev arrivano il numero di gruppi per risorsa, quelli con una possibile causa e il percorso del report.
- Comunicazione chiara e semplice, senza muri di testo (richiesta del dev): regola in testa alla skill, che vale in ogni fase. Messaggi brevi, una cosa per messaggio, elenchi di poche righe, dettaglio solo su richiesta, parole di tutti i giorni invece dei nomi dei campi.
- Spiegazioni in tre passi (richiesta del dev, dopo un test in cui l'agente aveva dovuto rispiegare una macro area): cosa non funziona per chi usa l'app, perché, cosa si fa; i nomi tecnici dopo. In `wm-tag`, per il report di verifica dell'import: titoli delle macro aree che dicono l'effetto, «Cosa» che comincia con una frase semplice, una riga su come si risolve.
- Ogni intervento proposto si chiude con una domanda a due scelte (richiesta del dev): farlo subito e comunicarlo, così dopo il nuovo confronto il problema non entra nel tag, oppure lasciarlo da fare, e allora va nel tag fra gli «Interventi manuali da fare». `interventi.md` e la sezione del tag hanno due elenchi, fatti e da fare.
- `wm-tag`, fase `repo-map`: la mappa dei repository non si mostra più al dev (richiesta del dev, dopo un test). Si dicono in una riga solo i repo nuovi aggiunti; se non ce ne sono, niente. Vale solo quando il materiale è un report di verifica dell'import (decisione del dev): con le trascrizioni la mappa si mostra come prima.
- Differenze «non visibili in questa app» (decisione del dev, dopo un test in cui il tag proponeva un ticket per il filtro dei POI per luogo, che sull'app di Geohub non è attivo): prima di cercare gli interventi, per ogni gruppo si stabilisce quale funzione del frontend legge il campo, quale chiave del config la accende e se è accesa nel config di Geohub. Con la chiave spenta il gruppo non è una macro area: resta nel report e in una riga delle note del tag.
- Al dev solo gli interventi da fare (decisione del dev, dopo un test): niente conteggi dei gruppi, niente spiegazioni su come la skill ha classificato le differenze, niente riassunto di quanti problemi si correggono solo nel codice. Il riepilogo è una riga (confronto fatto, percorso del report); il passo sulle differenze non visibili lavora in silenzio e scarta anche i campi che il frontend non legge affatto.
- Lingua sempre italiana (richiesta del dev, dopo un test in cui la skill aveva proposto un intervento in inglese): regola esplicita in testa alla skill e nella sezione del report di verifica di `wm-tag`. Si cambia lingua solo su richiesta esplicita del dev.
- Regola `poi_type_generico`, da segnalare (decisione del dev): un POI senza tipo su Geohub riceve sullo shard il tipo generico «Punto di interesse» (`EcPoi.php:276-280`). Regole alla versione 4.
- Visibilità di una funzione decisa sulla condizione reale del frontend, non solo su una chiave del config (decisione del dev, dopo un test in cui la skill proponeva l'icona di una mappa di base nel selettore, che nell'app su Geohub non c'è): la ricerca dice a quale condizione il frontend mostra la funzione (chiave, numero di elementi, ruolo, piattaforma) e se è vera nel config di Geohub. Le differenze di cui non si sa se l'app mostri la funzione non diventano macro aree né proposte: vanno fra le note «da verificare».
- Una macro area per problema e proporzione sempre detta (decisione del dev, dopo un test in cui excerpt e partenza/arrivo erano uniti, e «in 3 tracce mancano» aveva richiesto tre domande per capire che erano 3 su 3): `wm-tag` unisce gruppi solo con una causa comune provata; lo script aggiunge `su_geohub` ai gruppi «assente» e il report scrive «N casi su M».
- Prerequisiti letti dalla scaletta a ogni esecuzione (decisione del dev): la fase `prerequisiti` legge l'Artifact della scaletta di oc:8652 e mostra i passi di quel momento; l'elenco del 30/09/2026 resta solo come ripiego, dichiarato. Nella skill c'è l'indirizzo dell'Artifact senza la chiave di condivisione (`?sk=…`): il repo è pubblico.
- Script riscritto in Go (richiesta di Giuseppe, confermata dal dev): nel repo gli script delle skill sono in bash e il codice vero è in Go come il server MCP; lo script Python da circa 700 righe era l'eccezione. Ri-stima approvata dal dev e scritta su Orchestrator: da 5,6h a 8,6h (+3h per la riscrittura). Non comprende il lavoro aggiunto su richiesta durante le prove.

## Follow-up

- Confronto delle tassonomie per sola presenza invece che per valore (proposta di Giuseppe in call del 30/09/2026, 12:50): da valutare dopo un'altra verifica.
- Scaricare da GitHub solo le cartelle utili invece dell'intero repo (la copia di `geohub` pesa 208 MB). La cancellazione della copia alla fine è stata fatta: fase `pulizia`.
- Il render del diagramma non è stato verificato in un browser in questa sessione (strumento non disponibile): da controllare alla prima pubblicazione.
