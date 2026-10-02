> Ticket: oc:8670

# Skill wm-geohub-import-check: verifica dell'import di un'app da Geohub a uno shard wm-package

## Cosa cambia

Nasce la skill `wm-geohub-import-check` nel plugin `wm-skills`. Dati due indirizzi del frontend
— l'app su Geohub (`https://28.app.geohub.webmapp.it/`) e la stessa app importata su uno shard
wm-package (`https://3.maphubdev.maphub.it/`) — confronta **quello che il frontend legge** dalle due
parti e produce l'elenco delle differenze, tolte quelle che ci sono in ogni import per costruzione.

Il confronto lo fa uno script, non il modello: scarica i file pubblici su disco, abbina le feature
tramite `geohub_id`, confronta campo per campo e scrive il report. Al context arriva solo il
riepilogo.

A fine controllo la skill propone di passare il report a `wm-tag`, che acquista un terzo tipo di
materiale, il **report di verifica dell'import**: ne ricava un tag `[COLLAUDO][…]` con una macro area per ogni
gruppo di differenze, da cui si aprono i ticket di correzione come oggi.

La skill è **a sé**: non entra nel flusso di `wm-plan`, perché la migrazione di un'app non capita a
ogni lavoro.

## Perché

La migrazione di prova dell'app 28 (Itinera Romanica PLUS) su Maphub dev, il 29/09/2026, ha fatto
emergere a mano tre bug dell'import che i conteggi non mostravano: media dei POI mancanti (oc:8663),
`related_url` importato come testo invece che come oggetto (oc:8664), nomi assenti nel `pois.geojson`
(oc:8665). In oc:8459 il conteggio «EcPoi 32» era stato registrato come OK senza confronto.

Allo scrum del 30/09/2026 è stato deciso di farne una skill che «deve tirare fuori le differenze»
confrontando i file su AWS, il config e l'API Elastic delle due parti, lasciando a chi fa l'import il
giudizio su ciascuna. La scaletta di migrazione (oc:8652) prevede una verifica post-import per ogni
app: oggi è manuale e a campione.

## Requisiti

**Ingresso e individuazione delle due app**

- [ ] La skill riceve due URL del frontend, uno di Geohub e uno dello shard, e da ciascuno ricava id
  dell'app e shard con la stessa regola di wm-core (`EnvironmentService.init()`): `<id>.app.geohub…`
  → shard `geohub`; `<id>.<shard>.<dominio>` → quello shard; domini del cliente dalla mappa
  `redirects`.
- [ ] Gli indirizzi di ogni shard (`awsApi`, `elasticApi`, origin) e la mappa `redirects` si leggono
  a ogni esecuzione da wm-types, `src/environment.ts` su GitHub (repo pubblico,
  `raw.githubusercontent.com/webmappsrl/wm-types/main/src/environment.ts`): `redirects` cambia a
  ogni switch di un dominio cliente, e una copia ferma farebbe confrontare Geohub con Geohub. Il
  report dice da dove sono stati letti.
- [ ] Se GitHub non risponde, lo script usa la copia locale `shards.json`, con la data della copia, e
  lo dice in modo evidente. Se la struttura di `environment.ts` è cambiata e la lettura fallisce, o
  lo shard non si trova, la skill si ferma e lo dice: mai indirizzi indovinati.
- [ ] La skill verifica che l'app dello shard abbia `geohub_id` uguale all'id Geohub, leggendo
  `GET <origin>/api/v2/app/all` su file. Se non combacia si ferma, prima di confrontare due app che
  non c'entrano.
- [ ] Prima di scaricare, la skill chiede al dev di confermare che sono stati fatti i passi della
  scaletta che rigenerano ciò che il frontend legge — coda `geohub-import` vuota, import e
  sincronizzazione delle taxonomy where, «Reindicizza Scout», «Aggiorna Tracks su AWS», «Rigenera
  pois.geojson», «Process Track Data» sulle tracce senza metriche. Senza, i file sono quelli scritti
  durante l'import e il report si riempie di falsi allarmi.
- [ ] Nel report, le differenze tipiche di un passo saltato portano l'indicazione del passo come
  **possibile causa**, e restano comunque differenze del report: Elastic dello shard vuoto →
  «Reindicizza Scout»; tracce e POI senza where amministrative italiane → sincronizzazione delle
  where; metriche a 0 → «Process Track Data» (su app 28 → 3: 4 tracce su 57, per esempio «La Spezia
  – Itinerario 2», 74,6 km su Geohub e 0 sullo shard).

**Cosa si confronta** — solo ciò che legge il frontend, tutto via HTTP pubblico, in sola lettura

- [ ] `config.json` e `icons.json` delle due app.
- [ ] `pois.geojson`: ogni POI abbinato e confrontato campo per campo. La chiave di abbinamento è
  diversa dalle due parti: su Geohub è `id`, che è già l'id Geohub; sullo shard è
  `properties.geohub_id`.
- [ ] Elenco delle tracce da Elastic (`<elasticApi>/?app=geohub_app_<id>`): stesso numero e stesse
  tracce. Sullo shard Elastic espone solo l'id dello shard: l'abbinamento passa dal `geohub_id` del
  file della traccia. Sullo shard `hits` è una lista, non `{total, hits}`: il formato di Geohub va
  verificato.
- [ ] Dettaglio di **ogni** traccia (`tracks/<id>.json`), campo per campo: nome, descrizione,
  `related_url`, distanza, salita, discesa, durate, quote, immagini, POI collegati, taxonomy where,
  layer, attività, temi, difficoltà. Sullo shard la cartella `tracks/` è condivisa fra le app: le
  tracce dell'app sono quelle dell'elenco di Elastic, non il contenuto della cartella.
- [ ] Le taxonomy where si confrontano per nome e livello, non per chiave: sullo shard le chiavi sono
  diverse (`"15"` per le where Geohub, `"R41977"` per OSMFeatures) e ogni voce porta `_source` e
  `_admin_level` (formato di oc:8588). Le where OSMFeatures che Geohub non ha sono attese; una where
  di Geohub che manca sullo shard è una differenza.
- [ ] Una chiave presente da una parte sola si riporta **una volta per chiave**, con il numero di
  feature in cui manca, non una riga per feature. Le chiavi note come innocue stanno fra le
  differenze attese; le altre — per esempio `excerpt`, `gpx_url`, `kml_url`, `geojson_url`, presenti
  solo su Geohub nel `pois.geojson` — restano nel report finché qualcuno non stabilisce se il
  frontend le usa.
- [ ] Il confronto controlla anche il **tipo** del valore, non solo il testo: `related_url` identico
  ma stringa invece di oggetto è una differenza (oc:8664).
- [ ] Feature presenti da una parte sola: elencate con il loro id.
- [ ] I tile pbf sono esclusi: sono binari e derivano dagli stessi dati delle tracce.

**Immagini**

- [ ] Su **tutte** le feature, non a campione: presenza dell'immagine in evidenza e numero di
  immagini in galleria uguali. È questo il controllo che trova oc:8663 (su app 28 → 3: immagine in
  evidenza su 126 POI su Geohub, su 31 sullo shard); il campione sul contenuto non lo troverebbe.
- [ ] Il contenuto si controlla a campione, **una immagine per tipo** (evidenza e galleria delle
  tracce, evidenza e galleria dei POI), confrontando il file **originale** (`url`) e non le versioni
  ridimensionate, che lo shard rigenera: sulla copertina della traccia Geohub 31723 → Maphub 39
  l'originale ha lo stesso hash dalle due parti, le versioni `400x200` no. Se gli originali non
  coincidono il report riporta entrambi gli URL da guardare.

**Differenze attese**

- [ ] Le differenze che ogni import produce per costruzione (id nuovi, `APP.geohubId` che vale l'id
  dello shard, URL di bucket e percorsi diversi, mappatura delle chiavi del tema, `created_at` e
  `updated_at`, le where OSMFeatures in più, e le chiavi che esistono solo sullo shard come
  `geohub_id`, `geohub_synced_at`, `import_method`, `source`, `source_id`, `mbtiles`, `slope`,
  `taxonomy_wheres_show_first`) stanno in un file di regole della skill, letto dallo script, da
  ampliare quando un import ne rivela una nuova innocua.
- [ ] Nei soli campi di Elastic, un oggetto con una sola lingua da una parte e una stringa con lo
  stesso testo dall'altra sono uguali: su app 28 → 3 il nome delle 57 tracce è
  `{"it": "Boucle 1 – …"}` su Geohub e `"Boucle 1 – …"` sullo shard. Testo diverso o più lingue
  restano differenze, e la regola non vale per nessun altro file: `related_url` stringa invece di
  oggetto resta una differenza (oc:8664).
- [ ] I campi traducibili si confrontano in tutte le lingue presenti; una lingua che manca da una
  parte è una differenza.
- [ ] Una regola normalizza un valore (per esempio il prefisso di un URL), non elimina un campo dal
  confronto: un'immagine assente o in numero diverso resta una differenza anche se gli URL sono
  coperti da una regola.
- [ ] Il report riporta quante differenze ha ignorato ciascuna regola e la versione del file di
  regole usata.
- [ ] Una regola può essere «da segnalare»: la differenza è attesa e non passa a `wm-tag` come macro
  area, ma il report la mostra con casi ed esempi in una sezione propria. È il caso delle where di
  Geohub che la geometria dello shard non tocca, quando lo shard ha già le where OSMFeatures
  (regola `where_geometria`, versione 2 delle regole).

**Report**

- [ ] Il report è un file Markdown nella cartella di lavoro della sessione, fuori dal repo, con
  accanto il JSON completo delle differenze.
- [ ] Il riepilogo in conversazione raggruppa le differenze per tipo, con il numero di casi e due
  esempi che riportano gli id delle due parti e i due valori.
- [ ] Il report si chiude con le verifiche che via HTTP non si fanno, come promemoria da fare a mano
  con il rimando alla scaletta: non deve far credere di aver verificato l'import.
  - coda Horizon `geohub-import` e log;
  - proprietario dell'app con ruolo Editor;
  - UGC: non confrontati, e il promemoria lo dice esplicitamente quando l'app ne ha (l'app 28 ha 585
    UGC POI e 40 UGC tracce);
  - aprire la webapp dello shard e controllare che carichi e funzioni: su Maphub dev una GET da
    script andava a buon fine mentre la webapp falliva per le OPTIONS non gestite (CORS), e il
    grafico altimetrico dava `ERR_MODULE_NOT_FOUND`;
  - home e mappa a vista.
- [ ] Una risorsa che non si scarica (timeout, 404) compare come «non scaricata», mai come
  differenza né come uguaglianza.

**Passaggio a `wm-tag`**

- [ ] Il tag si crea solo per i problemi che non si risolvono a mano in poco tempo. Per ogni gruppo
  rimasto, e per le differenze da segnalare, la skill cerca un intervento rapido dalla piattaforma
  Laravel dello shard — un'action di Nova o la modifica di pochi record dall'interfaccia, mai comandi
  da terminale — e lo propone con l'action e la pagina esatte, dicendo se corregge solo i dati o
  anche la causa. Gli interventi li
  fa il dev; poi si rilancia la verifica, e a `wm-tag` va il report dell'ultimo giro.
- [ ] Cause e interventi da Nova si cercano nel codice che gira davvero, da GitHub, al commit fissato
  da chi pubblica l'applicazione: `wm-core` al submodule del `main` di `wm-webapp` (il frontend è un
  solo bundle per tutti gli shard), `wm-package` al submodule del repo Laravel dello shard di arrivo
  (`develop` per uno shard di sviluppo, `main` per gli altri), `geohub` su `main`. Mai le copie locali
  del dev. Il tag riporta i commit. Se un clone fallisce, la skill si ferma.
- [ ] Il confronto resta fra i frontend di partenza e di arrivo indicati dal dev, sull'ambiente che
  indicano (Maphub o Maphub dev); tag e ticket si scrivono sempre sull'Orchestrator di produzione,
  con anteprima e conferma, mai con una scrittura diretta.
- [ ] Gli interventi manuali fatti, con action, pagina di Nova ed effetto misurato — anche quelli senza
  effetto — si registrano e finiscono nella descrizione del tag, in una sezione «Interventi manuali».
- [ ] A fine controllo la skill chiede se passare il report a `wm-tag`; solo dopo un sì lo invoca
  con il percorso del report.
- [ ] Prima di passarlo, la skill cerca con `list_tags` un tag `[COLLAUDO][<APP>]` già esistente e,
  se c'è, lo dice al dev, che sceglie: aggiornare quel tag (descrizione riscritta con il nuovo
  report, con i problemi spariti segnati come risolti, anteprima e conferma di `update_tag`),
  crearne uno nuovo (per esempio verifica in produzione dopo quella su dev), o non creare nessun
  tag. Una verifica ripetuta dopo le correzioni, come prevede la scaletta, non deve produrre un tag in
  più a ogni giro.
- [ ] `wm-tag` accetta il **report di verifica dell'import** come terzo tipo di materiale:
  - la fonte in testa alla descrizione è una riga come `Verifica import Geohub 28 → maphubdev 3 del
    30/09/2026`, non il percorso del file;
  - il tag contiene da sé tutti i dati che servono, perché il file del report sparisce con la
    sessione;
  - il cliente si ricava dal nome dell'app nel config di Geohub e si propone al dev;
  - il nome predefinito è `[COLLAUDO][<APP>][<ANNO>]<N>`, sempre modificabile dal dev, con N
    calcolato cercando `COLLAUDO` e non `RDO`;
  - ogni gruppo di differenze è una macro area: il **Cosa** cita la differenza con numero di casi,
    id delle due parti e valori; **Come** ed **Esiste** come oggi, con la causa cercata in
    wm-package da `wm-codebase-research`;
  - le differenze ignorate per regola e le verifiche manuali vanno fra le note trasversali;
  - la verifica delle citazioni controlla che ogni esempio citato sia nel JSON del report, non su
    Drive; non si crea il notebook NotebookLM.
- [ ] `wm-tag` non cerca ticket già esistenti che trattino lo stesso problema: i doppioni li gestisce
  il dev, eliminandoli e lasciando che si ricreino dentro il tag.
- [ ] Dalla lista dei candidati in poi `wm-tag` non cambia.

**Verifica della skill**

- [ ] Lanciata su Geohub 28 → maphubdev 3, finché oc:8663, oc:8664 e oc:8665 sono aperti, la skill
  trova da sola i tre bug con i numeri visti negli import di prova del 29 e del 30/09:
  - 211 POI su 211 senza `name` nel `pois.geojson` (oc:8665);
  - `related_url` di tipo stringa su 57 tracce su 57 e su 149 POI, e la stringa `"[]"` nella traccia
    39 (oc:8664);
  - immagine in evidenza su 126 POI su Geohub contro 31 sullo shard (oc:8663).

  Trova inoltre la differenza nel numero di POI (209 nel `pois.geojson` di Geohub, 211 sullo shard),
  la cui causa non è ancora nota.
- [ ] Lo script ha test automatici su fixture ricavate dalla fotografia dei file pubblici delle due
  app scattata il 30/09/2026, prima delle correzioni di oc:8663-8665 (config, icons,
  `pois.geojson`, Elastic, 57 file di traccia per parte, `app/all` dello shard). Le fixture sono un
  sottoinsieme che contiene ogni bug e ogni differenza attesa — POI 2565 → 350, traccia 31723 → 39,
  una traccia con metriche a 0, i POI 106258 e 106260 presenti solo sullo shard — e i test
  verificano che lo script li trovi e che non segnali le differenze attese. La prova dal vivo resta
  una verifica in più, non l'unico criterio: dopo le correzioni non è più ripetibile.
- [ ] Il ramo «report di verifica dell'import» di `wm-tag` si prova durante lo sviluppo sull'Orchestrator
  locale (`http://localhost:8099`), mai in produzione, con anteprima e conferma su ogni scrittura
  come in uso normale.
  - Prima delle prove il database locale va allineato all'ultima versione di quello di produzione
    (nel repo `orchestrator`, `scripts/deploy_local_with_prod.sh`, che esegue `db:restore` e
    `migrate`), e l'allineamento si conferma confrontando in sola lettura le due istanze: i ticket
    del tag `[MAPHUB]Fix import` (id, nome e stato) e oc:8663, oc:8664, oc:8665 devono coincidere.
    Se non coincidono, le prove non partono.
  - La prova controlla che `wm-tag` crei il tag `[COLLAUDO][…]` con una macro area per ogni gruppo
    di differenze rimasto e proponga un ticket per ciascuna.
- [ ] In uso normale tag e ticket si creano in produzione, e nessuna scrittura parte senza anteprima
  e conferma esplicita del dev, un'operazione alla volta.

## Rischi

Verificati sui file pubblici di app 28 → 3 il 30/09/2026; quelli che non si verificano nel caso
d'uso normale sono stati scartati.

- **Prove di `wm-tag` su Orchestrator di produzione.** Il ramo nuovo scrive tag e associazioni, e le
  prove di sviluppo lascerebbero dati da ripulire. Mitigato: prove sull'istanza locale, con il
  database allineato alla produzione e l'allineamento confermato prima di cominciare.
- **Formato diverso dello stesso campo in Elastic.** Il nome delle 57 tracce è un oggetto con la
  lingua su Geohub e una stringa sullo shard: il controllo sul tipo produrrebbe 57 falsi allarmi.
  Mitigato: regola limitata ai campi di Elastic e a un oggetto con una sola lingua, con il conteggio
  dei casi ignorati.
- **Prova di accettazione non ripetibile.** Correggendo oc:8663-8665 i dati dal vivo cambiano.
  Mitigato: fotografia dei file del 30/09/2026 e fixture che contengono ogni bug.
- **Un tag in più a ogni verifica ripetuta.** La scaletta prevede di rilanciare l'import dopo ogni
  correzione. Mitigato: la skill cerca il tag di verifica dell'import dell'app e propone di aggiornarlo.
- **Indirizzi degli shard disallineati da wm-types.** `redirects` cambia a ogni switch di un dominio.
  Mitigato: lettura da GitHub a ogni esecuzione, copia locale solo come ripiego dichiarato.
- **Regole delle differenze attese troppo larghe.** Una regola che ignorasse interi campi URL
  nasconderebbe immagini rotte. Mitigato: le regole normalizzano il valore (prefisso dell'URL), non
  eliminano il campo; presenza e numero delle immagini si controllano a parte, e il report riporta
  la versione del file di regole usata.
- **Passi della scaletta confermati per abitudine.** Mitigato in parte: le differenze tipiche di un
  passo saltato indicano il passo come possibile causa, senza togliere la differenza dal report.

## Out of scope

- Confronto dei database via tunnel SSH: resta come passo successivo, per risalire alla causa quando
  il confronto via HTTP segnala qualcosa.
- Screenshot e confronto visivo di home e mappa.
- Confronto degli UGC: richiedono autenticazione. Il promemoria lo ricorda quando l'app ne ha.
- Correzione dei dati: la skill segnala, non corregge.
- Ripresa completa di un tag di verifica dell'import (`wm-tag` in modalità tag esistente, con la verifica dei
  blocchi). Entra solo l'aggiornamento della descrizione con il nuovo report.
- Inserimento della skill nel flusso di `wm-plan`.
- Ricerca dei ticket già esistenti che trattano lo stesso problema: i doppioni li gestisce il dev.

## Moduli toccati

Tutti nel repo `claude-marketplace`. Geohub, maphub, wm-package e Orchestrator non si toccano:
Orchestrator si usa con i tool che esistono già, senza migrazioni.

- `plugins/wm-skills/skills/wm-geohub-import-check/SKILL.md` — nuovo
- `plugins/wm-skills/skills/wm-geohub-import-check/shards.json` — nuovo, indirizzi degli shard
- `plugins/wm-skills/skills/wm-geohub-import-check/differenze-attese.json` — nuovo, regole delle
  differenze attese
- `plugins/wm-skills/scripts/geohub-import-check.py` — nuovo, download e confronto (Python 3, sola
  libreria standard)
- `plugins/wm-skills/scripts/tests/geohub-import-check.test.sh` e `tests/fixtures/` — nuovi
- `plugins/wm-skills/skills/wm-tag/SKILL.md` — terzo tipo di materiale
- `CLAUDE.md` — elenco delle skill e riga nella tabella dei coupling (`wm-geohub-import-check` →
  `wm-tag`, report di verifica dell'import)
- `docs/knowledge/` — pagina sulla verifica dell'import
