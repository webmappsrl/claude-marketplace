> Ticket: oc:8670

# wm-geohub-import-check — piano di implementazione

> **Per chi esegue:** sotto-skill richiesta: `superpowers:subagent-driven-development` (consigliata)
> o `superpowers:executing-plans`, un task alla volta. I passi usano le checkbox (`- [ ]`).
> **Nessun commit, `git add`, `git push` o branch automatico**: i commit indicati sono istruzioni
> per il dev, che li esegue dopo il review-gate.

**Obiettivo:** una skill che, dati gli URL del frontend di un'app su Geohub e della stessa app su
uno shard wm-package, confronta ciò che il frontend legge e produce un report di verifica dell'import, da
passare a `wm-tag` per aprire i ticket.

> ⚠️ L'implementazione ha deviato da architettura e tecnologie: [notes.md](notes.md#riscrittura-in-go)

**Architettura:** uno script Python con tre comandi separati — `risolvi` (URL → indirizzi), `scarica`
(tutto su disco), `confronta` (file su disco → `diff.json` e `report.md`) — così i test girano senza
rete sul confronto e con un server HTTP locale sullo scaricamento. La skill orchestra i comandi,
parla col dev e passa il report a `wm-tag`, che acquista il materiale «report di verifica dell'import».

**Tecnologie:** Python 3, sola libreria standard (`urllib`, `json`, `concurrent.futures`, `re`,
`hashlib`); test in bash con `jq`, sul modello di `plugins/wm-skills/scripts/tests/tag-candidati.test.sh`.

**Spec:** [overview.md](overview.md)

## Vincoli globali

- Python 3 con la sola libreria standard; nessuna dipendenza da installare.
- Sola lettura: lo script fa solo GET; nessuna scrittura su Geohub, shard o Orchestrator.
- La verifica dei certificati TLS non si disattiva mai, né per default né con un'opzione.
- Frontmatter della skill: solo `name: wm-geohub-import-check` e `description`.
- Tutto il testo (skill, messaggi dello script, report, commenti, documentazione) in italiano; i
  termini tecnici restano in inglese.
- Report e dati scaricati nella cartella di lavoro della sessione, mai nel repo.
- Prove del ramo della verifica dell'import di `wm-tag` solo su `http://localhost:8099`, mai su
  `https://orchestrator.maphub.it`.
- Prima di dichiarare pronto: `claude plugin validate .` e tutti i test verdi.
- Commit: `feat(oc:8670): …`.

## Indirizzi dei file (valori verificati il 30/09/2026)

| Risorsa | Shard «vecchi» (`geohub`, `geohubdev`) | Altri shard |
|---|---|---|
| config | `{awsApi}/conf/{id}.json` | `{awsApi}/{id}/config.json` |
| icone | `{awsApi}/icons/{id}.json` | `{awsApi}/{id}/icons.json` |
| POI | `{awsApi}/pois/{id}.geojson` | `{awsApi}/{id}/pois.geojson` |
| traccia | `{awsApi}/tracks/{trackId}.json` | `{awsApi}/tracks/{trackId}.json` (cartella condivisa fra app) |
| elenco tracce | `{elasticApi}/?app=geohub_app_{id}` | idem; `hits` è una lista |
| app dello shard | — | `{origin}/api/v2/app/all` |
| nome di una where Geohub | `{origin}/api/taxonomy/where/{whereId}` | — |

Regola dell'host, da wm-core `EnvironmentService.init()`: host che contiene una chiave di
`redirects` → quella voce; altrimenti `^(\d+)\.([a-zA-Z0-9-]+)(?:\.mobile)?(?:\.[^.]+)+$` con il
secondo gruppo non in `['app','geohub','mobile']` → id e shard; altrimenti id = primo sottodominio e
shard `geohub`.

## Da verificare in review

Casi che la spec implica e che un utente incontrerà; ognuno ha il suo test nel task indicato.

1. Una risorsa che risponde 404 o va in timeout → «non scaricata» nel report, mai differenza né
   uguaglianza (Task 2, Task 3).
2. Elastic dello shard con 0 tracce (Scout non reindicizzato) → il report lo dice con la possibile
   causa, niente eccezioni (Task 3).
3. Un POI dello shard senza `geohub_id`, creato a mano → «solo sullo shard», niente eccezioni
   (Task 3).
4. `name: null` da una parte e valore dall'altra → differenza, come la chiave assente (Task 3).
5. `environment.ts` con una struttura che il parser non riconosce → la skill si ferma con il
   messaggio, mai indirizzi indovinati (Task 1).

---

### Task 1: fixture e comando `risolvi`

**File:**
- Crea: `plugins/wm-skills/scripts/geohub-import-check.py`
- Crea: `plugins/wm-skills/skills/wm-geohub-import-check/shards.json`
- Crea: `plugins/wm-skills/scripts/tests/geohub-import-check.test.sh`
- Crea: `plugins/wm-skills/scripts/tests/fixtures/geohub-import-check/environment.ts` (copia del
  30/09/2026) e `environment-rotto.ts` (senza il blocco `export const shards`)
- Crea: `plugins/wm-skills/scripts/tests/fixtures/geohub-import-check/{geohub,shard}/` — sottoinsieme
  della fotografia in `<scratchpad>/foto-2026-09-30/`, ritagliato con `jq`:
  - POI: Geohub 2565 ↔ shard 350, due POI senza problemi, i POI shard con `geohub_id` 106258 e
    106260 (solo sullo shard), un POI shard con `geohub_id` tolto (caso 3 della review);
  - tracce: Geohub 31723 ↔ shard 39, «La Spezia – Itinerario 2» (metriche a 0 sullo shard), una
    traccia senza problemi; `elastic.json` delle due parti ridotto a queste tre;
  - `config.json`, `icons.json`, `app-all.json` interi.
  In testa al test, un commento dice da quale fotografia vengono e che non vanno rigenerate dopo le
  correzioni di oc:8663-8665: sono la prova che lo script trova quei bug.

**Interfacce:**
- Produce: `geohub-import-check.py risolvi <url>` → stdout JSON
  `{"shard","app_id","origin","awsApi","elasticApi","layout":"vecchio"|"nuovo","fonte"}`; exit 2 con
  messaggio su stderr se lo shard non c'è o `environment.ts` non si legge.
- Variabili d'ambiente per i test: `GIC_ENV_TS_URL` (default l'URL raw di GitHub) e
  `GIC_SHARDS_FALLBACK` (default `shards.json` della skill).
- Produce: `shards.json` = `{"copiato_il":"2026-09-30","fonte":"<url raw>","shards":{…},"redirects":{…}}`.

- [ ] **Passo 1: test che falliscono**, in `geohub-import-check.test.sh` con la funzione
  `verifica` del modello, usando `GIC_ENV_TS_URL=file://…/environment.ts`:
  - `risolvi https://28.app.geohub.webmapp.it/` → `.shard=="geohub"`, `.app_id==28`,
    `.layout=="vecchio"`, `.awsApi=="https://wmfe.s3.eu-central-1.amazonaws.com/geohub"`;
  - `risolvi https://3.maphubdev.maphub.it/` → `maphubdev`, `3`, `nuovo`,
    `.awsApi=="https://dev.maphub.it/wmfe/maphubdev"`;
  - `risolvi https://fiemaps.it/` → `geohub`, `29` (da `redirects`);
  - `risolvi https://3.inesistente.maphub.it/` → exit 2, stderr contiene `inesistente`;
  - con `environment-rotto.ts` → exit 2, stderr contiene `environment.ts`;
  - con `GIC_ENV_TS_URL` irraggiungibile → `.fonte` contiene `copia locale del 2026-09-30`.
- [ ] **Passo 2:** eseguire `bash plugins/wm-skills/scripts/tests/geohub-import-check.test.sh` e
  controllare che fallisca.
- [ ] **Passo 3:** implementare `risolvi` e il parser di `environment.ts`: estrarre i blocchi
  `export const shards` e `export const redirects` e leggerne le voci `nome: { chiave: 'valore' }`
  con espressioni regolari; nessun blocco trovato o zero voci → errore. Scrivere `shards.json`
  dalla copia del 30/09.
- [ ] **Passo 4:** rieseguire il test e controllare che passi.
- [ ] **Passo 5 (istruzione per il dev):** `git commit -m "feat(oc:8670): risoluzione degli URL del frontend e fixture della verifica dell'import"`

### Task 2: comando `scarica`

> ⚠️ L'implementazione ha deviato da questo task: [notes.md](notes.md#task-2-scaricamento)

**File:**
- Modifica: `plugins/wm-skills/scripts/geohub-import-check.py`
- Modifica: `plugins/wm-skills/scripts/tests/geohub-import-check.test.sh`

**Interfacce:**
- Consuma: `risolvi` del Task 1.
- Produce: `geohub-import-check.py scarica --geohub <url> --shard <url> --out <dir>` → in `<dir>`:
  `geohub/` e `shard/` con `config.json`, `icons.json`, `pois.geojson`, `elastic.json`,
  `tracks/<id>.json`; `shard/app-all.json`; `geohub/where/<id>.json` per ogni where citata dalle
  tracce Geohub; `immagini/` con i quattro campioni; `manifest.json` =
  `{"geohub":{…risolvi…},"shard":{…},"scaricato_il":"<ISO>","non_scaricate":[{"url","motivo"}],"campioni_immagini":[{"tipo","geohub_url","shard_url"}]}`.
- Exit 3 con messaggio se l'app dello shard in `app-all.json` non ha `properties.geohub_id` uguale
  all'id Geohub, prima di scaricare il resto.

- [ ] **Passo 1: test che falliscono.** Il test avvia `python3 -m http.server` su una porta libera
  sopra la cartella delle fixture e usa un `environment.ts` di prova i cui `awsApi`, `elasticApi` e
  `origin` puntano a `http://127.0.0.1:<porta>/…`. Asserzioni:
  - dopo `scarica` esistono i tre file di traccia per parte e `pois.geojson` dalle due parti;
  - una traccia elencata in Elastic ma assente dalle fixture compare in
    `.non_scaricate` con `motivo` che contiene `404`, e il comando esce con 0 (caso 1 della review);
  - con `app-all.json` modificato (`geohub_id` 99) il comando esce con 3 e non crea `tracks/`.
- [ ] **Passo 2:** eseguire il test e controllare che fallisca.
- [ ] **Passo 3:** implementare `scarica`: GET con `urllib`, timeout 30 s, 2 nuovi tentativi, al
  massimo 8 richieste in parallelo; l'elenco delle tracce dello shard viene da Elastic, non dalla
  cartella `tracks/`; campione delle immagini = primo originale (`url`) trovato per ciascun tipo
  (evidenza e galleria di tracce e POI), abbinato alla feature corrispondente dell'altra parte.
- [ ] **Passo 4:** rieseguire il test e controllare che passi.
- [ ] **Passo 5 (istruzione per il dev):** `git commit -m "feat(oc:8670): scaricamento dei file pubblici delle due app"`

### Task 3: comando `confronta` e regole delle differenze attese

> ⚠️ L'implementazione ha deviato da questo task: [notes.md](notes.md#task-3-confronto)

**File:**
- Modifica: `plugins/wm-skills/scripts/geohub-import-check.py`
- Crea: `plugins/wm-skills/skills/wm-geohub-import-check/differenze-attese.json`
- Modifica: `plugins/wm-skills/scripts/tests/geohub-import-check.test.sh`

**Interfacce:**
- Consuma: la cartella prodotta da `scarica` (le fixture hanno la stessa forma).
- Produce: `geohub-import-check.py confronta <dir> --regole <file>` → `<dir>/diff.json`:
  `{"regole_versione","gruppi":[{"codice","risorsa","campo","tipo","casi","esempi":[{"geohub_id","shard_id","geohub","shard"}],"possibile_causa"}],"solo_geohub":[…],"solo_shard":[…],"chiavi_da_una_parte":[{"risorsa","chiave","parte","feature"}],"ignorate":[{"regola","casi"}],"non_scaricate":[…]}`.
  `tipo` ∈ `valore`, `tipo_diverso`, `assente`, `conteggio_immagini`, `immagine_diversa`.
- Produce: `differenze-attese.json` = `{"versione":1,"regole":[{"codice","descrizione","si_applica_a",…}]}`.
  Regole iniziali: id nuovi; `APP.geohubId`; prefissi degli URL di immagini, file e icone
  (normalizzazione, il campo resta confrontato); mappatura del tema; `created_at`/`updated_at`;
  where OSMFeatures in più; chiavi solo dello shard `geohub_id`, `geohub_synced_at`,
  `import_method`, `source`, `source_id`, `mbtiles`, `slope`, `taxonomy_wheres_show_first`, `code`,
  `icon`, `type`, `color`, `taxonomy_where`, `taxonomyWheres`; nome in Elastic oggetto con una sola
  lingua ↔ stringa con lo stesso testo.

- [ ] **Passo 1: test che falliscono**, sulle fixture, senza rete:
  - un gruppo con `campo=="name"`, `risorsa=="pois.geojson"`, `tipo=="assente"` che contiene
    l'esempio `geohub_id 2565` / `shard_id 350` (oc:8665);
  - un gruppo `campo=="related_url"`, `tipo=="tipo_diverso"` con l'esempio della traccia
    `31723` / `39` e `shard=="[]"` (oc:8664);
  - un gruppo `conteggio_immagini` sui POI (oc:8663);
  - un gruppo con le metriche della traccia «La Spezia – Itinerario 2» a 0 e
    `possibile_causa` che contiene `Process Track Data`;
  - `.solo_shard` contiene i `geohub_id` 106258 e 106260 e il POI senza `geohub_id` (caso 3);
  - `.ignorate` contiene la regola del nome in Elastic con `casi` uguale al numero di tracce
    delle fixture, e nessun gruppo riguarda il nome in Elastic;
  - `.chiavi_da_una_parte` contiene `gpx_url` una sola volta, con `parte=="geohub"`;
  - nessun gruppo per `geohub_synced_at`, `created_at` o per gli URL delle immagini;
  - le where si confrontano per nome italiano: nessuna differenza per le where OSMFeatures in più
    sulla traccia 39; una where Geohub tolta dalla fixture dello shard produce un gruppo `assente`;
  - una fixture con `name: null` su un POI dello shard produce `assente` (caso 4);
  - `elastic.json` dello shard svuotato → gruppo con `possibile_causa` che contiene
    `Reindicizza Scout`, exit 0 (caso 2);
  - una traccia in `non_scaricate` non compare né fra i gruppi né fra le feature uguali (caso 1);
  - `.regole_versione == 1`.
- [ ] **Passo 2:** eseguire il test e controllare che fallisca.
- [ ] **Passo 3:** implementare `confronta`. Abbinamento: POI per `properties.id` su Geohub e
  `properties.geohub_id` sullo shard; tracce per `properties.id` su Geohub e `properties.geohub_id`
  nel file della traccia dello shard (Elastic dello shard ha solo l'id dello shard). Confronto
  ricorsivo campo per campo, che distingue tipo e valore; i traducibili in tutte le lingue; le where
  Geohub (id → nome da `geohub/where/<id>.json`) devono comparire fra i nomi di tutte le where dello
  shard; per le immagini presenza dell'evidenza e numero della galleria su tutte le feature, più il
  confronto `sha256` dei campioni. Le regole normalizzano i valori prima del confronto e contano i
  casi; una chiave presente da una parte sola si registra una volta per chiave.
- [ ] **Passo 4:** rieseguire il test e controllare che passi.
- [ ] **Passo 5 (istruzione per il dev):** `git commit -m "feat(oc:8670): confronto campo per campo con le differenze attese"`

### Task 4: report Markdown

**File:**
- Modifica: `plugins/wm-skills/scripts/geohub-import-check.py`
- Modifica: `plugins/wm-skills/scripts/tests/geohub-import-check.test.sh`

**Interfacce:**
- Consuma: `diff.json` e `manifest.json` dei Task 2-3.
- Produce: `confronta` scrive anche `<dir>/report.md` e stampa su stdout il riepilogo (una riga per
  gruppo: risorsa, campo, tipo, casi, due esempi con i due id e i due valori).

- [ ] **Passo 1: test che falliscono** sul `report.md` generato dalle fixture:
  - prima riga `# Verifica import Geohub 28 → maphubdev 3`, poi la data dello scaricamento;
  - sezioni, in quest'ordine: `## Differenze`, `## Presenti da una parte sola`,
    `## Chiavi presenti da una parte sola`, `## Non scaricate`, `## Differenze ignorate per regola`,
    `## Verifiche da fare a mano`;
  - `## Differenze ignorate per regola` riporta `versione 1` e i conteggi;
  - `## Verifiche da fare a mano` contiene `geohub-import`, `Editor`, `UGC`, `CORS`, e il rimando
    alla scaletta;
  - con `non_scaricate` vuoto la sezione dice `Nessuna`.
- [ ] **Passo 2:** eseguire il test e controllare che fallisca.
- [ ] **Passo 3:** implementare la scrittura del report e del riepilogo.
- [ ] **Passo 4:** rieseguire il test e controllare che passi.
- [ ] **Passo 5 (istruzione per il dev):** `git commit -m "feat(oc:8670): report di verifica dell'import"`

### Task 5: la skill `wm-geohub-import-check`

> ⚠️ L'implementazione ha deviato da questo task: [notes.md](notes.md#task-5-skill)

**File:**
- Crea: `plugins/wm-skills/skills/wm-geohub-import-check/SKILL.md`
- Crea: `plugins/wm-skills/skills/wm-geohub-import-check/tag-correzioni.json` =
  `{"tag":[{"id":651,"name":"[MAPHUB]Fix import"}]}`

**Interfacce:**
- Consuma: i tre comandi dello script (`${CLAUDE_PLUGIN_ROOT}/scripts/geohub-import-check.py`).
- Produce per `wm-tag` (Task 6): percorso di `report.md`, con accanto `diff.json`, `manifest.json`
  e una copia di `tag-correzioni.json`.

- [ ] **Passo 1:** scrivere `SKILL.md` con frontmatter `name: wm-geohub-import-check` e una
  `description` che dica quando usarla (verifica dell'import di un'app da Geohub su uno shard) e
  quando no (nel flusso di `wm-plan`). Fasi `##`:
  - `ingresso`: chiede i due URL del frontend; esegue `risolvi` e mostra shard, id e fonte degli
    indirizzi (in evidenza se è la copia locale);
  - `prerequisiti`: una sola domanda di conferma sui passi della scaletta — coda `geohub-import`
    vuota, taxonomy where importate e sincronizzate, «Reindicizza Scout», «Aggiorna Tracks su AWS»,
    «Rigenera pois.geojson», «Process Track Data» sulle tracce senza metriche;
  - `scaricamento` e `confronto`: esegue `scarica` e `confronta` nella cartella di lavoro della
    sessione; exit 2 o 3 → si ferma e riporta il messaggio;
  - `riepilogo`: mostra lo stdout del riepilogo e il percorso del report, senza leggere nel context
    `diff.json` intero;
  - `tag esistente`: `list_tags` con `search: "COLLAUDO"`, filtro sul nome dell'app; se c'è un tag,
    una domanda con le tre scelte — aggiornarlo (`update_tag` con anteprima e conferma, problemi
    spariti segnati come risolti), crearne uno nuovo, nessun tag;
  - `passaggio a wm-tag`: una domanda; dopo il sì invoca `wm-skills:wm-tag` con il percorso del
    report e la scelta sul tag.
  Regole in testa: una domanda per messaggio; nessuna scrittura senza anteprima e conferma; la
  skill non corregge i dati.
- [ ] **Passo 2:** eseguire `claude plugin validate .` e controllare che non segnali errori.
- [ ] **Passo 3 (istruzione per il dev):** `git commit -m "feat(oc:8670): skill wm-geohub-import-check"`

### Task 6: `wm-tag` accetta il report di verifica dell'import

> ⚠️ L'implementazione ha deviato da questo task: [notes.md](notes.md#task-6-wm-tag)

**File:**
- Modifica: `plugins/wm-skills/skills/wm-tag/SKILL.md` — `## Fase: input` (righe 33-54),
  `## Fase: client-extraction` (91-100), `## Fase: tag-naming` (103-119),
  `## Fase: tag-description` (122-165), `## Fase: tag-creation` (213-242),
  `## Fase: candidate-review` (346-400)

**Interfacce:**
- Consuma: il pacchetto del Task 5 (report, `diff.json`, `manifest.json`, `tag-correzioni.json`,
  scelta sul tag esistente).

- [ ] **Passo 1:** in `Fase: input` aggiungere il terzo formato, **report di verifica dell'import**: fonte = riga
  `Verifica import Geohub <id> → <shard> <id> del <data>` da `manifest.json`; il tag deve contenere da
  sé tutti i dati, perché il report sparisce con la sessione.
- [ ] **Passo 2:** `client-extraction`: per il report, il cliente è il nome dell'app nel config di
  Geohub, proposto al dev. `tag-naming`: default `[COLLAUDO][<APP>][<ANNO>]<N>`, N contato con
  `list_tags search: "COLLAUDO"`; se il dev ha scelto di aggiornare un tag esistente, si salta.
- [ ] **Passo 3:** `tag-description`: ogni gruppo di `diff.json` è una macro area; **Cosa** = risorsa,
  campo, tipo, casi e due esempi con i due id e i due valori; **Come** ed **Esiste** come oggi, con
  una sola chiamata a `wm-codebase-research` sul repo wm-package; differenze ignorate per regola e
  verifiche manuali nelle note trasversali.
- [ ] **Passo 4:** `tag-creation`: per il report la verifica delle citazioni è un `jq` che cerca in
  `diff.json` ogni coppia di id citata; niente notebook NotebookLM e niente riga `**Notebook:**`;
  con la scelta «aggiorna» si usa `update_tag`, sempre con anteprima e conferma.
- [ ] **Passo 5:** prima di `candidate-review`, per il report: `get_tag` su ogni tag di
  `tag-correzioni.json`, letto su file con `jq` portando nel context solo id, nome e stato; la
  descrizione si apre solo per i ticket che sembrano trattare lo stesso problema. Candidato con
  ticket aperto → «già trattato in oc:<ID> (<stato>): lo associo al tag di verifica dell'import invece di
  crearne uno?», `attach_story_to_tag` dopo il sì; ticket `done`/`released`/`pending_release` →
  segnalato come ricomparso (non rilasciato, import non rilanciato, o regressione) e decide il dev.
  In tag-mode la `customer_request` si scrive come descrizione del problema, perché notifica.
- [ ] **Passo 6:** eseguire `claude plugin validate .` e controllare che non segnali errori.
- [ ] **Passo 7 (istruzione per il dev):** `git commit -m "feat(oc:8670): wm-tag accetta il report di verifica dell'import"`

### Task 7: prove dal vivo

> ⚠️ L'implementazione ha deviato da questo task: [notes.md](notes.md#task-7-prove-dal-vivo)

Nessun file nuovo; l'esito va in `notes.md`.

- [ ] **Passo 1: Orchestrator locale allineato.** Nel repo `orchestrator`, con Docker avviato,
  eseguire `scripts/deploy_local_with_prod.sh`. Poi confrontare in sola lettura produzione e locale:
  id, nome e stato dei ticket del tag 651 e di oc:8663, oc:8664, oc:8665. Se non coincidono, fermarsi
  e dirlo al dev.
- [ ] **Passo 2:** configurare il server MCP su `http://localhost:8099` da `.mcp.json.example` e
  riavviare la sessione (il dev esegue il riavvio).
- [ ] **Passo 3:** lanciare la skill su `https://28.app.geohub.webmapp.it/` e
  `https://3.maphubdev.maphub.it/`. Atteso, finché oc:8663-8665 sono aperti: 211 POI su 211 senza
  `name`; `related_url` stringa su 57 tracce e su 149 POI; immagine in evidenza su 126 POI su Geohub
  e 31 sullo shard; 4 tracce con metriche a 0; POI 106258 e 106260 solo sullo shard. Se i ticket sono
  già stati corretti, registrarlo e affidarsi ai test del Task 3.
- [ ] **Passo 4:** passare il report a `wm-tag` sull'istanza locale e controllare, con anteprima e
  conferma a ogni scrittura: tag `[COLLAUDO][…]` creato; proposta di associare 8663-8665 invece di
  crearli; un ticket del tag messo `done` in locale segnalato come ricomparso; ticket nuovi proposti
  solo per le differenze senza ticket.
- [ ] **Passo 5:** ripristinare `.mcp.json` sulla produzione e dirlo al dev.

### Task 8: documentazione del repo

**File:**
- Modifica: `CLAUDE.md` — elenco delle skill in `## Cos'è questo repo`; riga nella tabella
  `### Coupling tra skill`: `wm-geohub-import-check` → `wm-tag`, contratto = pacchetto del report
  (`report.md`, `diff.json`, `manifest.json`, `tag-correzioni.json`) e materiale «report di verifica dell'import»
  di `wm-tag`; riga in `## Conoscenza`.
- Crea: `docs/knowledge/verifica-import-geohub.md` — come funziona oggi, perché così (HTTP invece
  di SSH, confronto su tutte le feature, regole che normalizzano, fotografia per le fixture, tag
  aggiornato invece di duplicato), con `(oc:8670)`.

- [ ] **Passo 1:** scrivere pagina e righe; farle controllare a `wm-context-guard` (in
  `update-context`).
- [ ] **Passo 2:** eseguire `claude plugin validate .` e l'intero
  `bash plugins/wm-skills/scripts/tests/geohub-import-check.test.sh`; entrambi verdi.
- [ ] **Passo 3 (istruzione per il dev):** `git commit -m "feat(oc:8670): documentazione della verifica dell'import"`
