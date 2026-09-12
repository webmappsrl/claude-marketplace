# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Regola che precede tutte le altre

**Non eseguire `git commit`, `git add`, `git push` né creare branch come parte di un lavoro.**
Scrivi i file, fermati, dichiara che il lavoro è pronto. Se un'istruzione che stai seguendo
prevede un commit, quell'istruzione è sbagliata.

**Unica eccezione:** il dev chiede il commit come atto a sé, dopo aver visto cosa è cambiato —
«committa» detto guardando il risultato, non «fai X e committa».

Il commit è il momento in cui il dev può dire di no: committare dentro il lavoro glielo toglie.

## Seconda regola: mai scritture sulla produzione di Orchestrator

**Non provare mai una scrittura contro `https://orchestrator.maphub.it`.** Alcune hanno effetti
verso l'esterno che nessuno può annullare: scrivere `customer_request` su una story fa scattare
`addResponse()`, che **notifica il cliente**; i link PDF firmati dei preventivi restano validi
fino a 90 giorni e non si revocano.

Si prova **sull'istanza locale, `http://localhost:8099`**, avviata con Docker dal repo
`webmappsrl/orchestrator`. La configurazione del server MCP sta in `.mcp.json.example`, da
copiare in `.mcp.json` (ignorato da git) adattando il percorso del file di credenziali.

Esiste anche un'istanza di sviluppo, ma le manca un certificato HTTPS valido: usala solo come
verifica facoltativa, e **non rendere l'accettazione dei certificati non validi il
comportamento predefinito di alcuno strumento** — va chiesta ogni volta, per non farla diventare
normale anche verso la produzione.

## Cos'è questo repo

Marketplace di plugin Claude Code del team Webmapp, pubblicato su GitHub come
`webmappsrl/claude-marketplace`. Distribuisce due plugin:

- **`superpowers`** — framework agentico di Jesse Vincent ([obra/superpowers](https://github.com/obra/superpowers)),
  referenziato direttamente dall'upstream con `ref: main`
- **`wm-skills`** — skill interne Webmapp (convenzioni di codice, checklist PR, processo deploy),
  mantenute in questo repo sotto `plugins/wm-skills/`

## Vincoli sui file di config

- Il campo `name` in `.claude-plugin/marketplace.json` **non può** contenere prefissi come
  `claude-*` né impersonare nomi Anthropic ufficiali: il plugin system lo blocca con errore di
  schema.
- `plugins/wm-skills/.claude-plugin/plugin.json` ha un `version` semver, aggiornato ad ogni
  release insieme al tag `v<version>` sul commit corrispondente.

## Ambiente

- **Go 1.27.1** per il server MCP (`plugins/wm-skills/mcp/go.mod`).
- **Il binario versionato è compilato solo per `darwin/arm64`** (`build.sh`): su Linux o Mac
  Intel va ricompilato cambiando `GOOS`/`GOARCH`, altrimenti il server MCP non parte.
- Il server MCP non ascolta su alcuna porta: comunica su stdio, e il file delle credenziali si
  passa con `--auth-file`.

## File da non modificare a mano

- `plugins/wm-skills/bin/orchestrator-mcp` — binario generato: si rigenera con `build.sh`.
- `docs/guide/wm-plan-diagramma/index.html` — il template grafico è congelato; i vincoli stanno
  in `.claude/rules/wm-plan-diagramma.md`, che si carica quando tocchi quei file.

## Comandi

| Cosa | Comando |
|---|---|
| Validare marketplace, plugin e skill | `claude plugin validate .` |
| Test del server MCP | `cd plugins/wm-skills/mcp && go test ./...` |
| Ricompilare il binario MCP (versionato nel repo) | `plugins/wm-skills/mcp/build.sh` |
| Verificare che diagramma e skill siano allineati | `./.github/scripts/verifica-diagramma.sh` |

`claude plugin validate .` va eseguito **sempre** prima di un commit: è l'unico controllo che
copre frontmatter e manifesti. I test Go vanno lanciati quando si tocca `mcp/`, e il binario
ricompilato ad ogni release, perché viaggia nel plugin.

In CI gira il controllo di coerenza fra skill e diagramma, e la pubblicazione delle guide.

## Lingua

Documentazione, commenti, messaggi di commit e descrizioni delle PR sono **in italiano**. I
termini tecnici restano in inglese: commit, branch, merge, gate, build, deploy, review. Nomi
di file, slug e identificatori seguono la stessa regola dei termini tecnici.

## Versioning del plugin wm-skills

La versione installata mostrata nell'header di sessione di `wm-plan` (`Header di sessione` →
`### header: versione` in `plugins/wm-skills/skills/wm-plan/SKILL.md`) **è un valore statico
scritto direttamente in quella skill**, non letto a runtime da `plugin.json`, dalla cache dei
plugin o da git. Motivo: la cache dei plugin (`~/.claude/plugins/cache/...`) non è un repository
git e il path del repo installato varia a seconda di come l'utente ha aggiunto il marketplace —
provare a risolverlo a runtime si è rivelato fragile. Il valore viaggia con il resto del contenuto
della skill, quindi resta intrinsecamente allineato ad ogni `/plugin marketplace update`.

**Checklist di release (obbligatoria, in quest'ordine):**
1. Bump `version` in `plugins/wm-skills/.claude-plugin/plugin.json` (semver: patch per fix, minor per nuove skill/feature retro-compatibili, major per breaking change nel contratto artefatti o nel nome delle skill).
2. Aggiorna la riga `**Versione installata:** v<version>` in `plugins/wm-skills/skills/wm-plan/SKILL.md` → `### header: versione` con lo stesso valore del bump.
2-bis. Aggiorna la costante `Version` in `plugins/wm-skills/mcp/internal/version/version.go` e ricompila il binario con `plugins/wm-skills/mcp/build.sh` — il binario compilato è versionato nel repo e va aggiornato a ogni rilascio.
3. Commit di entrambi i file su `main`.
4. Tag del commit: `git tag v<version> && git push origin v<version>`.

## Aggiungere una nuova skill del team

1. Creare la cartella:
   ```
   plugins/wm-skills/skills/<nome-skill>/
   ```
2. Creare `SKILL.md` con questo frontmatter YAML obbligatorio:
   ```markdown
   ---
   name: nome-skill-in-kebab-case
   description: Una frase che spiega quando Claude deve invocare questa skill.
   ---

   Corpo della skill in Markdown...
   ```
3. Il campo `description` è il testo che Claude usa per decidere se invocare la skill:
   deve essere una frase con soggetto ("Use when…" oppure "Usa quando…"), non un titolo.
4. Validare prima del commit (vedi sotto), poi fare commit + push su `main`.

### Coupling tra skill

Alcune skill condividono un contratto su artefatti o comportamenti. Quando modifichi una skill, verifica se esiste un coupling documentato qui e aggiorna tutte le skill coinvolte in sincronia.

| Skill | Con | Contratto condiviso |
|---|---|---|
| `wm-plan` | `wm-review-ticket` | `wm-plan` è fonte autoritativa del contratto artefatti `docs/features/<slug>/`. `wm-review-ticket` lo referenzia, non lo duplica. Modificare la struttura artefatti richiede solo aggiornare `wm-plan`. |
| `wm-tag` ↔ `wm-plan` | tag-mode | `wm-tag` invoca `wm-plan` passando titolo, tipo, repo e TAG_ID; `wm-plan` si occupa di reverse-interaction, overview, challenge, estimation e description del ticket, `wm-tag` di tag, lista ticket e loop. Nel verso opposto, `caso-c` in `Fase: ticket` cede il controllo a `wm-tag`. |
| `wm-plan` (challenge) | `wm-plan` (review-gate) | Entrambe le sotto-fasi isolano il giudizio in un subagente cieco (solo path/istruzioni, nessun riassunto della conversazione precedente). Se il pattern di isolamento cambia in una sotto-fase, verificare se va aggiornato anche nell'altra. |

### Convenzioni di naming

- **Prefisso obbligatorio `wm-`**: tutte le skill di `wm-skills` devono avere il nome in kebab-case con prefisso `wm-` (es. `wm-plan`, `wm-review-ticket`). Questo vale sia per il nome della cartella che per il campo `name` nel frontmatter di `SKILL.md`.

### Convenzioni di stile per le skill

- **Tono**: imperativo e diretto, rivolto a Claude come agente che esegue istruzioni
- **Lunghezza**: da poche decine di righe (checklist) a qualche centinaio (workflow articolati)
- **Struttura**: sezioni Markdown (`##`) per separare fasi o categorie; elenchi puntati per step atomici
- **Frontmatter**: solo `name` e `description` sono obbligatori; evitare campi extra non necessari

## Validare

```bash
claude plugin validate .
```

Verifica `marketplace.json`, `plugin.json` e tutti i `SKILL.md`. Da eseguire sempre **prima di
dichiarare un lavoro pronto**, specialmente dopo aver toccato i file di config.

## Workflow di test in locale (senza pushare)

Per iterare su una skill senza dover fare push ogni volta, aggiungere il marketplace come path
locale dalla directory del repo:

```bash
# Dentro Claude Code, dalla root del repo:
/plugin marketplace add .
/plugin install wm-skills@wm-marketplace
```

Questo punta al filesystem locale. Ogni modifica a un `SKILL.md` è immediatamente disponibile
ricaricando la sessione, senza commit né push.

Per tornare alla versione remota:
```bash
/plugin marketplace remove wm-marketplace
/plugin marketplace add webmappsrl/claude-marketplace
```

## Skill wm-skills

Tre skill: `wm-plan`, `wm-review-ticket`, `wm-tag`. **Quando si attiva ciascuna lo dice il campo
`description` del suo `SKILL.md`**, che è la fonte autoritativa: non ripeterlo qui, o le due
versioni divergeranno. Ogni skill può comporre skill di `superpowers`, installato a parte.


## Orchestrator

I ticket vivono su Orchestrator, piattaforma Laravel interna. Formato ID: `oc:<numero>`.

**Se manca un endpoint si estende l'API, non si costruisce un aggiramento lato skill**: quegli
endpoint sono nati per queste skill, non sono un servizio terzo.

**Nei documenti generati dalle skill**: ogni file `docs/features/` inizia con
`> Ticket: oc:<ID>`; lo slug di una feature è `<ID>-<titolo-in-kebab-case>`; lo scope dei commit
è `feat(oc:<ID>): …` / `fix(oc:<ID>): …` / `refactor(oc:<ID>): …`.

Il resto — formato dei campi, specifica OpenAPI, ripiego se il server MCP non parte — in
[docs/knowledge/orchestrator-integrazione.md](docs/knowledge/orchestrator-integrazione.md).

## File `.mcp.json` locale

Il repo ignora `.mcp.json` (contiene percorsi assoluti personali, es. il file di auth di sviluppo): copia `.mcp.json.example` in `.mcp.json` e adatta i percorsi al tuo ambiente prima di usare il server MCP `orchestrator-dev`.

## Decisioni architetturali

### Server MCP per Orchestrator
- **Go compilato invece di un ambiente da installare**: il binario viaggia nel plugin, quindi nessuno deve installare nulla; Go produce un eseguibile di poche decine di MB senza dipendenze da scaricare a runtime, contro l'ambiente di esecuzione che un runtime interpretato si porterebbe dietro
- **Comunicazione sui canali standard del processo (stdio), nessuna porta in ascolto**: esclude per costruzione ogni conflitto con i container Docker del team
- **Elenchi dei valori ammessi (`type`, `status`) letti dagli enum PHP e non dalla specifica OpenAPI**: verificato in esecuzione che la specifica generata da Scramble non li espone (quei due campi risultano senza tipo e senza elenco); il server legge direttamente `StoryType.php` e `StoryStatus.php` con espressioni regolari. Il pacchetto `internal/spec`, che avrebbe dovuto leggere l'intera specifica OpenAPI per campi e tipi, è stato **eliminato** durante l'esecuzione (non riparato): la specifica reale di produzione usa per alcuni campi la sintassi OpenAPI 3.1 `"type": ["string","null"]`, che il pacchetto non interpretava, e il server si rifiutava di avviarsi per un dato — `deps.Spec` — che nessuna riga del programma leggeva. Campi e tipi vengono invece dalle strutture Go dichiarate a mano
- **Gli elenchi dei valori ammessi finiscono nello schema del tool, non solo in un controllo interno**: verificato che l'SDK MCP accetta uno schema costruito a runtime — un valore fuori elenco (es. il tipo `Task`, inesistente) diventa così inesprimibile per costruzione, rifiutato dallo schema prima ancora che la chiamata parta, invece di affidarsi a una convalida interna che si potrebbe dimenticare di eseguire
- **Nessuna conferma rafforzata per le scritture non annullabili** (eliminazione di un preventivo, collegamento PDF pubblico, `customer_request` che notifica il cliente): un meccanismo a codice con scadenza è stato scartato perché richiederebbe uno stato da conservare lato server; nessun parametro derivabile dai dati può comunque fermare l'agente che li ha appena letti, quindi una conferma "rafforzata" darebbe una falsa sicurezza. Al posto delle conferme rafforzate: il titolo/nome della risorsa colpita viene portato nei parametri a scopo informativo dentro la richiesta di autorizzazione, e la regola operativa — questi tool non vanno mai fra quelli approvati in automatico — resta una responsabilità umana, non tecnica
- **Produzione e istanza locale sono due server MCP distinti, con nomi diversi (`orchestrator` e `orchestrator-dev`)**: il plugin distribuito dichiara solo `orchestrator`, fisso sulla produzione; `orchestrator-dev` esiste solo nel `.mcp.json` di questo repo, per il collaudo. Nomi diversi rendono l'ambiente visibile nella richiesta di autorizzazione senza doverlo stampare a parte. Il file delle credenziali è stato reso configurabile (`--auth-file`, non più fissato dentro il client) e separato per lo stesso motivo dopo un difetto scoperto nel collaudo dal vivo: i due server leggevano lo stesso file e non potevano essere usati in parallelo con identità diverse
- **Le vecchie istruzioni `curl` sopravvivono come ripiego** in `plugins/wm-skills/shared/orchestrator-fallback.md`, letto dalle skill solo su richiesta e solo se i tool MCP non rispondono: costo zero nelle sessioni normali, aggiramento manuale disponibile se il server non parte

### Esecuzione automatica PHPStan pre-PR/merge in wm-plan (oc:8341)
- **Blocco duro per errori sul diff corrente e per fallimenti infrastrutturali, stesso trattamento per entrambi**: inizialmente si era considerato un trattamento più permissivo (fail-soft) per i fallimenti infrastrutturali (comando non trovato, crash, timeout) rispetto agli errori di qualità reali — la decisione finale del dev in Fase: reverse-interaction ha uniformato i due casi allo stesso hard-block di default, per evitare un bypass implicito su problemi ambientali che potrebbero mascherare un errore reale
- **Debito tecnico preesistente non blocca**: l'output PHPStan viene incrociato con `git diff --name-only` per distinguere errori sui file toccati dal diff corrente da errori preesistenti sul resto del codebase — solo i primi bloccano il commit; per i secondi `wm-plan` propone la creazione di un ticket Orchestrator separato, evitando che un progetto con debito tecnico pregresso renda il gate permanentemente bloccante e inutilizzabile
- **Override sempre motivato e loggato, nessun kill-switch persistente**: il dev può bypassare un blocco solo dopo una conferma esplicita e distinta (non basta il "procedi" generico del gate), con una motivazione proposta/dedotta e confermata in preview, sempre registrata in `notes.md` con la responsabilità esplicita attribuita al dev — deliberatamente nessuna configurazione di disattivazione permanente per repo, per non introdurre un meccanismo che possa silenziare il check senza tracciabilità
- **Servizio Docker riusa `$DOCKER_PROJECT_DIR_NAME` già risolto in `environment-setup: docker-check`**: nessuna euristica aggiuntiva di rilevamento servizio — evita la fragilità di un `docker compose config --services` che potrebbe selezionare un servizio senza `vendor/bin/phpstan` disponibile
- **Detection CI intenzionalmente permissiva**: grep case-insensitive sulla keyword `phpstan` in `.github/workflows/*.yml`, accettando il rischio di falsi positivi (step disabilitato ma matchato) a fronte di un rischio più basso di falsi negativi (check che dovrebbe attivarsi ma non lo fa) — un falso positivo residuo è comunque a basso impatto perché gestito dallo stesso meccanismo di bypass-con-motivazione
- **Timeout fisso a 5 minuti**: valore di partenza stimato per progetti Laravel di dimensioni tipiche Webmapp, senza dati misurati specifici sui repo del team — regolabile in futuro se si rivela troppo stretto
- **Accoppiamento a un solo tool (PHPStan) accettato consapevolmente**: nessuna astrazione "quality-gate pluggable" generica in questo ciclo — se in futuro serviranno altri strumenti (ESLint, Pint, Psalm), andrà valutata una generalizzazione, non prevista ora

### Formato description per Story vs Tag su Orchestrator: HTML vs Markdown
- **`Story.description` (ticket) è HTML, non Markdown**: verificato leggendo il repo `orchestrator` — l'editor Nova è `Marshmallow\Tiptap\Tiptap` (`app/Traits/fieldTrait.php` → `descriptionField()`, usato anche per `customer_request` e `answer_to_ticket`), e i dati reali in DB confermano markup HTML (`<p>`, `<pre><code>`, `<a href>`). Skill che scrivono `description` su una Story (`wm-plan` in `ticket: caso-b` e tag-mode, `wm-review-ticket` in Fase 6b) devono generare HTML — inviare Markdown grezzo lo fa apparire letteralmente come testo nell'editor Tiptap
- **`Tag.description` è invece Markdown**: l'editor Nova è `Datomatic\NovaMarkdownTui\MarkdownTui` con `EditorType::MARKDOWN` (`app/Nova/Tag.php`), e i dati reali in DB sono Markdown grezzo (`##`, liste `-`, emoji), senza tag HTML. `wm-tag`, che scrive la `description` del tag, resta quindi in Markdown — nessuna conversione
- **Nessuna sanitizzazione HTML lato backend** su nessuno dei due campi (`StoryApiRequest`/`TagApiRequest` validano solo `string` generico): un errore di formato non genera un errore API, si traduce solo in un rendering sbagliato nell'editor corrispondente — quindi va rispettato per convenzione, non per vincolo di validazione

### Fix path cross-repo header wm-plan (versione via cache plugin, URL diagramma via SKILL.md statico)
- **`### header: versione` risolve il repo via cache plugin, non più via path relativo alla cwd**: il comando precedente (`realpath plugins/wm-skills/skills/wm-plan/SKILL.md` relativo alla cwd) funzionava solo se `wm-plan` veniva invocato da dentro il repo `claude-marketplace` — da qualsiasi altro repo falliva silenziosamente e il check versione andava sempre in `⚠️ Check versione non disponibile`. Fix: `find ~/.claude/plugins/cache -maxdepth 5 -path '*/wm-skills/*/skills/wm-plan/SKILL.md'` individua il path della skill installata indipendentemente dalla cwd, poi `git -C "$(dirname ...)" rev-parse --show-toplevel` deriva la root del repo in modo robusto (niente `../../../..` fisso)
- **`### header: diagramma` non fa più fetch remoto di `CLAUDE.md`**: l'URL dell'Artifact è ora un valore statico scritto direttamente in `SKILL.md` invece che in `CLAUDE.md` — dato che l'URL cambia solo quando si modifica la skill stessa (stessa sessione in cui si aggiorna `SKILL.md`), tenerlo in `SKILL.md` lo mantiene intrinsecamente sincronizzato ad ogni `/plugin marketplace update`, elimina la dipendenza da rete/GitHub per questa parte dell'header, e non introduce rischio di staleness (a differenza di quanto temuto inizialmente in fase di discussione)
- **Superata la decisione precedente "Fetch remoto sempre" per il diagramma**: la decisione originale di `curl -sf` su `raw.githubusercontent.com/.../CLAUDE.md` per leggere l'URL è superata da questo fix — resta valida solo per il check versione (`### header: versione`, hash `LOCAL_HASH` vs `REMOTE_HASH` via GitHub API), che continua a dover leggere un dato realmente remoto (l'ultimo commit su `main`)

### Intestazione ASCII per skill wm-plan con info versione/aggiornamenti e diagramma di flusso (oc:8283)
- **Header mostrato solo alla prima invocazione di sessione**: nessun meccanismo di stato programmato — è una regola comportamentale in prosa che Claude segue, dato che le skill non hanno un tracciamento di sessione nativo garantito
- **Check versione senza stato locale aggiuntivo**: confronto tra hash HEAD del repo git installato e hash ultimo commit remoto via GitHub API — evita il problema di un file di stato che verrebbe sovrascritto ad ogni `marketplace update` o che partirebbe senza baseline su nuove installazioni
- **Modalità sviluppo locale esplicita nell'header**: se il branch non è `main` o `SKILL.md` ha modifiche non committate, il check versione viene sostituito da un'indicazione visibile ("🔧 modalità sviluppo locale"), non disattivato silenziosamente — pensato per chi lavora su `wm-plan` in locale (workflow `/plugin marketplace add .`)
- **Trigger di rigenerazione Artifact non ristretto ai soli file di `wm-plan`**: scatta su qualsiasi modifica al repo `claude-marketplace`, anche se il diagramma rappresenta solo il workflow di `wm-plan` — scelta esplicita dell'utente, accettato lo spreco occasionale di redeploy non necessari
- **Pubblicazione Artifact per tentativo diretto, nessun check preventivo di account**: dato che due account Claude diversi possono condividere la stessa email (caso reale riscontrato: webmapp e net7 sulla stessa mail), non esiste un segnale di sessione affidabile per distinguerli — l'unico modo verificato è il fallimento del redeploy stesso, con avviso esplicito di switchare account
- **Drift documentale tra Artifact e stato reale di `SKILL.md` accettato come rischio consapevole**: nessuna verifica automatica di sincronia (es. hash di pubblicazione) — la responsabilità di notare e correggere è dell'utente, decisione presa esplicitamente in Fase: challenge
- **URL Artifact versionato in `CLAUDE.md`**, non in un file di config locale (`~/.config/webmapp/`) — è un link condiviso da tutto il team, non un dato personale, quindi va distribuito automaticamente via git a chiunque cloni il repo

### Clear del context in wm-plan dopo reverse-interaction e dopo implementation (oc:8282)
- **Nessun clear reale del context principale**: il pivot deciso durante la Fase: reverse-interaction ha scartato l'idea di "svuotare" il context — l'isolamento si ottiene solo delegando a subagenti nei punti dove serve un giudizio non contaminato dal ragionamento pregresso (motivazione primaria: indipendenza di giudizio, non economia di context)
- **`Fase: overview` resta invariata**: scritta dal context principale, perché in quel punto del workflow il context è ancora leggero e chi ha condotto `reverse-interaction` ha un vantaggio informativo che un subagente isolato non recupererebbe da un riassunto di seconda mano
- **`execution: review-gate` isola sempre il riepilogo del diff**, senza soglie né eccezioni per la skill di implementazione usata — anche la ridondanza con le review per-task di `subagent-driven-development` è accettata come controllo doppio intenzionale
- **`--find-renames --find-copies` obbligatori** nel subagente di review-gate, per non descrivere un file rinominato come "nuovo + cancellato"
- **Il riepilogo del subagente non sostituisce mai la lettura del diff da parte del developer** — resta un ausilio di orientamento, il gate reale è l'approvazione esplicita del developer sul diff completo
- **Coordinamento tra sottoagenti paralleli durante l'implementazione è fuori scope**: già gestito da `subagent-driven-development` (esecuzione sequenziale, non parallela, con ledger di progresso) — eventuali miglioramenti vanno proposti upstream su `obra/superpowers`, non in `wm-skills`

### Rivedere criteri di stima ore in wm-plan (oc:8278)
- **Classificazione per-componente invece di buffer forfettario**: ogni componente della stima è "scrittura pura" (zero domande aperte dopo overview+challenge, buffer 0%) o "decisioni aperte" (UX/reverse-engineering legacy, buffer 20-30%) — motivato da analisi Orchestrator su ticket luglio 2026 che mostrava overstima sistematica su task ben specificati
- **Buffer di integrazione trasversale 5%**: separato dal buffer per-componente, copre il rischio di interazione tra componenti che nessun componente singolo cattura
- **Tempo di pianificazione misurato, non stimato**: timestamp reale registrato in `Fase: ticket` (`planning_start_at`) e confrontato con quello a fine `Fase: estimation` — mostrato al dev come "Misurato + Stimato = Totale", mai un numero unico fuso
- **Marcatore di versione `[stima v2 — per-componente]`** su ogni stima scritta su Orchestrator: garantisce che i dati storici restino distinguibili tra criterio vecchio e nuovo per calibrazioni future
- **Coefficiente di velocità per-dev esplicitamente rimandato**: Orchestrator non espone oggi un endpoint di listing/aggregazione stimato-vs-effettivo per utente, e il campione per dev è troppo piccolo (8-15 ticket) per un coefficiente affidabile
- **`execution: re-estimation`**: se durante l'esecuzione emerge un imprevisto stimabile, si propone al dev una revisione della stima con conferma esplicita, prima del PATCH `estimated_hours`

### wm-tag skill e fase estimation in wm-plan (oc:8157)
- **tag-mode in wm-plan**: quando invocato da `wm-tag`, `wm-plan` salta write-plan/execution/notes/update-context — l'overview va nella description del ticket, non nel filesystem
- **Fase: estimation solo per Feature**: i bug non si stimano in ore (costo nella diagnosi, non nella fix) — solo Feature ricevono `estimated_hours`
- **repos.json per navigazione multi-repo**: dizionario persistente `~/.config/webmapp/repos.json` aggiornato incrementalmente — non riscritto da zero per preservare path manuali
- **Naming tag `[RDO][CLIENTE][ANNO]N`**: N calcolato dinamicamente contando tag esistenti per stesso cliente+anno su Orchestrator — evita conflitti senza coordinazione manuale
- **Regola scritture estesa a tag**: la regola preview+conferma di `wm-plan` per le story si applica identicamente a tutti i POST/PATCH su Orchestrator, incluse le operazioni sui tag

### wm-plan slug e environment-setup (oc:8102)
- **Slug al posto dei numeri nelle fasi**: `## Fase: ticket`, `## Fase: environment-setup`, ecc. — inserire nuove fasi non richiede mai rinumerazione
- **`Fase: environment-setup` centralizza il rilevamento ambiente**: `project-detection`, `domain-mapping`, `ux-ui-detection`, `docker-check` eseguiti prima di `init-context` e `reverse-interaction`
- **`init-context` mantenuta separata**: Claude legge `CLAUDE.md` come primo atto di comprensione del progetto — semanticamente distinto dal rilevamento tecnico
- **docker-check FAIL-SOFT**: qualsiasi errore → `⚠️` + prosegui, mai bloccare il workflow; usa `docker compose stop` mai `down`/`rm`
- **Review opzionale in wm-plan ora `execution: formal-review`**: sottofase esplicita invece di hint testuale in `execution: review-gate`

### wm-review-ticket skill (oc:8068)
- **Contratto artefatti via WebFetch su GitHub raw**: `wm-review-ticket` non duplica la struttura `docs/features/` ma la legge da `wm-plan/SKILL.md` su GitHub al runtime — nessun drift possibile
- **`wm-plan` fonte autoritativa del contratto**: la tabella coupling in `CLAUDE.md` è l'unico punto da aggiornare se la struttura `docs/features/` cambia
- **Stash automatico pre-checkout**: se il working tree è dirty, la skill fa `git stash` automatico e `git stash pop` al termine — nessun rischio di perdita lavoro in corso
- **Review opzionale in wm-plan Fase 6d**: formulata come domanda sì/no esplicita, non hint passivo

### Ask user to set ticket status to progress in wm-plan (oc:7973)
- **File auth unificato JSON** (`orchestrator-auth.json`) invece di token plain text: permette di salvare anche `user_id`, `name`, `email` in un unico file, necessari per il PATCH `user_id` senza chiamate aggiuntive
- **`GET /api/me`** per ricavare l'utente corrente invece di usare `creator_id` del ticket: garantisce che l'assegnazione sia sempre all'utente autenticato, anche su ticket creati da altri
- **`notes.md` come registro decisioni a posteriori**: le modifiche richieste dopo l'approvazione del piano vanno registrate in `notes.md` sezione "Decisioni"

## Feature disponibili

| Feature | Ticket | Moduli toccati | Note |
|---|---|---|---|
| Server MCP per Orchestrator | — (nessun ticket) | `plugins/wm-skills/mcp/`, `plugins/wm-skills/.mcp.json`, le tre `SKILL.md` | Server MCP in Go che espone l'API di Orchestrator come tool tipizzati: valori ammessi (`type`, `status`) letti dagli enum PHP (`StoryType.php`, `StoryStatus.php`), campi e tipi dalle strutture Go; anteprima obbligatoria prima delle scritture tramite il parametro `confirm`; gruppi di tool attivabili. Le skill non costruiscono più chiamate HTTP a mano |
| Tipi ticket Orchestrator letti dinamicamente in wm-plan e wm-tag | — (nessun ticket) | `plugins/wm-skills/skills/wm-plan/SKILL.md`, `plugins/wm-skills/skills/wm-tag/SKILL.md` | Rimosso il tipo inesistente `Task` (una POST `/api/stories` con `"type": "Task"` falliva con `422 — Il valore selezionato per type non è valido`). Nuova sezione `## Orchestrator API → Tipi disponibili (letti dinamicamente)` che rimanda a `StoryType.php` su GitHub, stesso pattern già usato per gli status: nessun valore di tipo è scritto nelle skill, così l'aggiunta o la rinomina di un tipo su Orchestrator non le fa invecchiare in silenzio |
| Esecuzione automatica PHPStan pre-PR/merge in wm-plan | oc:8341 | `plugins/wm-skills/skills/wm-plan/SKILL.md` | `execution: review-gate` esegue PHPStan automaticamente su repo Laravel con PHPStan in CI; hard-block su errori del diff corrente o fallimenti infrastrutturali; override motivato e tracciato in notes.md; errori preesistenti fuori dal diff propongono un ticket Orchestrator dedicato invece di bloccare |
| Ask user to set ticket status to progress in wm-plan | oc:7973 | `plugins/wm-skills/skills/wm-plan/SKILL.md` | Chiede all'utente di mettere il ticket in progress al termine della Fase 0; unifica le credenziali Orchestrator in `orchestrator-auth.json` |
| wm-review-ticket skill | oc:8068 | `plugins/wm-skills/skills/wm-review-ticket/SKILL.md`, `plugins/wm-skills/skills/wm-plan/SKILL.md` | Nuova skill per code review strutturata di ticket Orchestrator; contratto artefatti via WebFetch su wm-plan; stash automatico pre-checkout; review opzionale in wm-plan Fase 6d |
| wm-plan slug e environment-setup | oc:8102 | `plugins/wm-skills/skills/wm-plan/SKILL.md` | Migrazione fasi a slug inglesi; nuova Fase: environment-setup con project-detection, domain-mapping, ux-ui-detection, docker-check |
| wm-tag skill e fase estimation in wm-plan | oc:8157 | `plugins/wm-skills/skills/wm-tag/SKILL.md`, `plugins/wm-skills/skills/wm-plan/SKILL.md` | Nuova skill `wm-tag` per trascrizione → tag + ticket; `caso-c` in Fase: ticket; `Fase: estimation` per Feature; tag-mode in wm-plan |
| Rivedere criteri di stima ore in wm-plan | oc:8278 | `plugins/wm-skills/skills/wm-plan/SKILL.md` | Classificazione per-componente (scrittura pura/decisioni aperte) con buffer per-componente invece di forfettario; pianificazione misurata via timestamp; marcatore versione `[stima v2 — per-componente]`; `execution: re-estimation` per revisioni mid-execution |
| Clear del context in wm-plan dopo reverse-interaction e dopo implementation | oc:8282 | `plugins/wm-skills/skills/wm-plan/SKILL.md` | Isolamento sempre attivo del riepilogo diff in `execution: review-gate` tramite subagente cieco (stesso pattern di `challenge`); fallback esplicito se `planning_start_at` non è stato registrato in `Fase: estimation` |
| Intestazione ASCII per skill wm-plan con info versione/aggiornamenti e diagramma di flusso | oc:8283 | `plugins/wm-skills/skills/wm-plan/SKILL.md`, `CLAUDE.md` | Nuova sezione "Header di sessione" (banner ASCII, check versione via hash HEAD vs remoto, modalità sviluppo locale, link Artifact diagramma); nuova sezione `## Diagramma di flusso wm-plan` in CLAUDE.md con URL e regola di rigenerazione post-modifica repo |
| Fix diagramma header cross-repo in wm-plan | — (nessun ticket) | `plugins/wm-skills/skills/wm-plan/SKILL.md` | `### header: diagramma` legge sempre il `CLAUDE.md` di `claude-marketplace` via fetch remoto (`curl -sf`), non più il `CLAUDE.md` del repo target — corregge il falso "non ancora pubblicato" quando wm-plan è invocato da altri repo |
| Split automatico ticket Help desk multi-richiesta in wm-plan | — (nessun ticket) | `plugins/wm-skills/skills/wm-plan/SKILL.md` | Rilevamento automatico di più richieste distinte in ticket Help desk letti in `caso-a`; split in ticket separati (originale ridotto alla prima richiesta + nuovi ticket per le successive), `creator_id` replicato dal cliente originale, tag di raggruppamento opzionale per uso interno dev |
| Fix formato description ticket (HTML invece di Markdown) in wm-plan e wm-review-ticket | — (nessun ticket) | `plugins/wm-skills/skills/wm-plan/SKILL.md`, `plugins/wm-skills/skills/wm-review-ticket/SKILL.md` | Confermato leggendo il repo `orchestrator`: `Story.description` (ticket) è HTML via editor Nova Tiptap, `Tag.description` è invece Markdown via editor Nova MarkdownTui. `wm-plan` (`ticket: caso-b`, tag-mode) e `wm-review-ticket` (Fase 6b) ora costruiscono la `description` del ticket in HTML; `wm-tag` resta in Markdown per la `description` del tag, nessuna modifica lì |
