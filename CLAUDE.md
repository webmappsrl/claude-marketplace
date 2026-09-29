# claude-marketplace — CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Regola che precede tutte le altre

**Non eseguire `git commit`, `git add`, `git push` né creare branch come parte di un lavoro.**
Scrivi i file, fermati, dichiara che il lavoro è pronto. Se un'istruzione che stai seguendo
prevede un commit, quell'istruzione è sbagliata.

**Unica eccezione:** il dev chiede il commit come atto a sé, dopo aver visto cosa è cambiato —
«committa» detto guardando il risultato, non «fai X e committa».

Il commit è il momento in cui il dev può dire di no: committare dentro il lavoro glielo toglie.

## Quando un difetto emerge in un repo, si corregge la skill

Lavorando su un `CLAUDE.md` di prodotto — con `wm-context-doctor`, `wm-context-guard` o a mano —
i difetti che emergono sono quasi sempre difetti dello strumento che non li ha trovati. **La prima
domanda è se il doctor o il guard l'avrebbero visto. Se no, la correzione va prima nello strumento,
poi nel file.**

Correggere solo il file chiude un caso e lascia il buco aperto per tutti i repo successivi. E non
è nemmeno la strada più corta: il difetto va comunque portato nella skill, quindi il giro si fa due
volte invece di una.

Il ciclo è: ripristina il file allo stato di partenza → lancia il doctor → valuta cosa trova e cosa
no → **correggi la skill** → riavvia la sessione, perché le definizioni degli agenti si caricano
all'avvio (`shared/claude-md-rules.md` no, quello si legge a runtime) → rilancia. Solo quando lo
strumento regge si esegue il piano sul file.

## Lingua: mai modi di dire inglesi tradotti

Vale per tutto ciò che si scrive — documentazione, commenti, messaggi di commit, prompt delle
skill, risposte al dev. **Un'espressione idiomatica inglese non si traduce alla lettera**: il team
è italiano e quei modi di dire non li conosce, quindi parola per parola non si capiscono. «Alzare
il pavimento» per *raise the floor*, «strato sottile» per *thin layer*, «a colpo d'occhio» per *at
a glance*: sono frasi che sembrano italiane e non lo sono.

I **termini tecnici** restano in inglese: commit, branch, merge, build, deploy, review, gate. È
l'errore opposto, e fa lo stesso danno.

Se non diresti quella frase parlando con un collega, non scriverla: di' la cosa in modo esplicito,
anche se è più lungo.

## Seconda regola: mai scritture sulla produzione di Orchestrator

**Non provare mai una scrittura contro `https://orchestrator.maphub.it`.** Alcune hanno effetti
verso l'esterno che nessuno può annullare: scrivere `customer_request` su una story fa scattare
`addResponse()`, che **notifica il cliente**; i link PDF firmati dei preventivi restano validi
fino a 90 giorni e non si revocano.

Si prova **sull'istanza locale, `http://localhost:8099`**, avviata con Docker dal repo
`webmappsrl/orchestrator`. La configurazione del server MCP sta in `.mcp.json.example`, da
copiare in `.mcp.json` (ignorato da git, contiene percorsi personali) adattando il percorso del
file di credenziali.

Esiste anche un'istanza di sviluppo, ma le manca un certificato HTTPS valido: usala solo come
verifica facoltativa, e **non rendere l'accettazione dei certificati non validi il
comportamento predefinito di alcuno strumento** — va chiesta ogni volta, per non farla diventare
normale anche verso la produzione.

## Terza regola: online va solo `docs/guide/`

**Non pubblicare niente di `docs/` che non stia sotto `docs/guide/`.** Il cantiere in
`docs/features/`, la conoscenza in `docs/knowledge/` e le procedure in `docs/howto/` sono
interni e non devono finire online — su un repo cliente significherebbe pubblicare le sue note
di sviluppo.

La pubblicazione passa solo da `.github/workflows/pages.yml`, che dichiara `docs/guide/` come
sorgente. **Non configurare GitHub Pages dall'interfaccia del repository**: accetta come
sorgente solo la root o l'intera `docs/`, quindi pubblicherebbe anche il resto.

Una guida si scrive assumendo che la legga chiunque: nessun percorso interno, nessun nome di
branch, nessun dato di un altro cliente.

## Cos'è questo repo

Marketplace di plugin Claude Code del team Webmapp, pubblicato su GitHub come
`webmappsrl/claude-marketplace`. Distribuisce due plugin:

- **`superpowers`** — framework agentico di Jesse Vincent ([obra/superpowers](https://github.com/obra/superpowers)),
  referenziato direttamente dall'upstream con `ref: main`
- **`wm-skills`** — skill interne Webmapp, mantenute in questo repo sotto `plugins/wm-skills/`

Le skill sono `wm-plan`, `wm-review-ticket`, `wm-tag`. **Quando si attiva ciascuna lo dice il
campo `description` del suo `SKILL.md`**, che è la fonte autoritativa: non ripeterlo qui, o le
due versioni divergeranno. Ogni skill può comporre skill di `superpowers`, installato a parte.

## Regole del repo

- **Prima di dichiarare pronto un lavoro esegui `claude plugin validate .`**: è l'unico
  controllo che copre frontmatter delle skill e manifesti. Vale sempre, a maggior ragione dopo
  aver toccato i file di config.
- **Prima di ogni release esegui la checklist di rilascio**, obbligatoria e nell'ordine dato:
  [docs/howto/rilascio-wm-skills.md](docs/howto/rilascio-wm-skills.md). È lì che vivono il bump
  di `version` in `plugins/wm-skills/.claude-plugin/plugin.json`, l'allineamento della versione
  mostrata da `wm-plan` e del binario MCP, e il tag sul commit.
- **Quando modifichi una skill, controlla se ha un coupling con un'altra e aggiornale nello
  stesso lavoro** — il dettaglio è nella tabella qui sotto.
- **Il campo `name` in `.claude-plugin/marketplace.json` non può** contenere prefissi come
  `claude-*` né impersonare nomi Anthropic ufficiali: il plugin system lo blocca con errore di
  schema.
- **Documentazione, commenti, messaggi di commit e descrizioni delle PR sono in italiano.** I
  termini tecnici restano in inglese: commit, branch, merge, gate, build, deploy, review. Nomi
  di file, slug e identificatori seguono la stessa regola dei termini tecnici.

### File da non modificare a mano

- `plugins/wm-skills/bin/orchestrator-mcp` — binario generato: si rigenera con `build.sh`.
- `docs/guide/wm-plan-diagramma/index.html` — il template grafico è congelato nella struttura e
  si aggiorna solo nel contenuto, quando cambiano le fasi di `wm-plan`: i vincoli stanno in
  `.claude/rules/wm-plan-diagramma.md`, che si carica quando tocchi quei file.

### Coupling tra skill

| Skill | Con | Contratto condiviso |
|---|---|---|
| `wm-plan` | `wm-review-ticket` | `wm-plan` è fonte autoritativa del contratto artefatti `docs/features/<slug>/`. `wm-review-ticket` lo referenzia, non lo duplica: modificare la struttura richiede solo aggiornare `wm-plan`. |
| `wm-tag` ↔ `wm-plan` | tag-mode | `wm-tag` invoca `wm-plan` in tag-mode, e `caso-c` in `Fase: ticket` cede il controllo a `wm-tag`. La divisione dei compiti sta nei due `SKILL.md`. |
| `wm-plan` (challenge) | `wm-plan` (review-gate) | Entrambe le sotto-fasi isolano il giudizio in un subagente cieco (solo path e istruzioni, nessun riassunto della conversazione). Se il pattern di isolamento cambia in una, verifica l'altra. |
| `wm-plan` (proposta tag) | `wm-tag` | `wm-tag` crea tag dossier, con descrizione; `wm-plan` propone e crea solo etichette, cioè tag con `description` nulla. Il confine è il filtro sulla descrizione: se cambia in una delle due, verifica l'altra. |
| `wm-plan` ↔ `wm-tag` | NotebookLM | Entrambe usano `wm-transcript-research` (che contiene il template dei notebook), la verifica `shared/verifica-citazioni.md` e i nomi dei notebook (`scrum AAAA-MM-GG`, `tag <nome del tag>`). Se cambia uno di questi, verifica l'altra skill. |

## Convenzioni per le skill

- **Prefisso `wm-` obbligatorio**, in kebab-case, sia sul nome della cartella sia sul campo
  `name` del frontmatter.
- **Frontmatter**: solo `name` e `description`; niente campi extra.
- **Tono** imperativo e diretto, rivolto a Claude come agente che esegue istruzioni; sezioni
  `##` per le fasi, elenchi puntati per i passi atomici.

## Ambiente

- **Go 1.27.1** per il server MCP (`plugins/wm-skills/mcp/go.mod`).
- **Il binario versionato è compilato solo per `darwin/arm64`** (`build.sh`): su Linux o Mac
  Intel va ricompilato cambiando `GOOS`/`GOARCH`, altrimenti il server MCP non parte.
- Il server MCP non ascolta su alcuna porta: comunica su stdio, e il file delle credenziali si
  passa con `--auth-file`.

## Comandi

| Cosa | Comando |
|---|---|
| Validare marketplace, plugin e skill | `claude plugin validate .` |
| Test del server MCP (quando tocchi `mcp/`) | `cd plugins/wm-skills/mcp && go test ./...` |
| Ricompilare il binario MCP (versionato nel repo) | `plugins/wm-skills/mcp/build.sh` |
| Verificare che diagramma e skill siano allineati | `./.github/scripts/verifica-diagramma.sh` |

In CI girano il controllo di coerenza fra skill e diagramma e la pubblicazione delle guide.

## Procedure

| Cosa devi fare | Procedura |
|---|---|
| Rilasciare una versione di `wm-skills` | [docs/howto/rilascio-wm-skills.md](docs/howto/rilascio-wm-skills.md) |
| Aggiungere una skill al plugin | [docs/howto/aggiungere-una-skill.md](docs/howto/aggiungere-una-skill.md) |
| Provare una skill in locale senza pushare | [docs/howto/test-in-locale.md](docs/howto/test-in-locale.md) |
| Aggiornare o pinnare `superpowers` | [docs/howto/aggiornare-superpowers.md](docs/howto/aggiornare-superpowers.md) |
| Provare `wm-transcript-research` dopo averne toccato il prompt | [docs/howto/provare-wm-transcript-research.md](docs/howto/provare-wm-transcript-research.md) |
| Installare NotebookLM per `wm-transcript-research` | [docs/howto/installare-notebooklm.md](docs/howto/installare-notebooklm.md) |

## Orchestrator

I ticket vivono su Orchestrator, piattaforma Laravel interna. Formato ID: `oc:<numero>`.

**Se manca un endpoint si estende l'API, non si costruisce un aggiramento lato skill**: quegli
endpoint sono nati per queste skill, non sono un servizio terzo.

**Nei documenti generati dalle skill**: ogni file `docs/features/` inizia con
`> Ticket: oc:<ID>`; lo slug di una feature è `<ID>-<titolo-in-kebab-case>`; lo scope dei commit
è `feat(oc:<ID>): …` / `fix(oc:<ID>): …` / `refactor(oc:<ID>): …`.

Il resto — formato dei campi, specifica OpenAPI, ripiego se il server MCP non parte — in
[docs/knowledge/orchestrator-integrazione.md](docs/knowledge/orchestrator-integrazione.md).

## Trappole

Stanno in `.claude/rules/`, con il frontmatter `paths:` che le carica quando si toccano i file
corrispondenti: `wm-plan-diagramma` tiene i vincoli del template grafico pubblicato su GitHub
Pages, che è congelato nella struttura e si aggiorna solo nel contenuto.

## Conoscenza

| Argomento | Cosa copre | Pagina |
|---|---|---|
| Integrazione con Orchestrator | API costruita per le skill, server MCP, formato dei campi testuali | [docs/knowledge/orchestrator-integrazione.md](docs/knowledge/orchestrator-integrazione.md) |
| Header di sessione di `wm-plan` | Versione mostrata, check aggiornamenti, link al diagramma | [docs/knowledge/wm-plan-header-sessione.md](docs/knowledge/wm-plan-header-sessione.md) |
| Gate di review di `wm-plan` | PHPStan automatico, isolamento del riepilogo diff, review formale | [docs/knowledge/wm-plan-review-gate.md](docs/knowledge/wm-plan-review-gate.md) |
| Stima delle ore in `wm-plan` | Criterio per componente, pianificazione misurata, ri-stima | [docs/knowledge/wm-plan-stima-ore.md](docs/knowledge/wm-plan-stima-ore.md) |
| Fasi e rilevamento ambiente di `wm-plan` | Slug al posto dei numeri, `environment-setup`, `docker-check` | [docs/knowledge/wm-plan-fasi-e-ambiente.md](docs/knowledge/wm-plan-fasi-e-ambiente.md) |
| Come `wm-plan` tratta i ticket | Assegnazione e stato, split dei ticket Help desk, `notes.md` | [docs/knowledge/wm-plan-gestione-ticket.md](docs/knowledge/wm-plan-gestione-ticket.md) |
| `wm-tag` e la tag-mode | Naming dei tag, `repos.json`, cosa salta `wm-plan` in tag-mode | [docs/knowledge/wm-tag-e-tag-mode.md](docs/knowledge/wm-tag-e-tag-mode.md) |
| Delega ad agenti | Quali fasi vanno a un agente, i due tipi di delega, come si contesta un esito | [docs/knowledge/wm-skills-delega-agentica.md](docs/knowledge/wm-skills-delega-agentica.md) |
| `wm-review-ticket` | Contratto artefatti letto a runtime, stash pre-checkout | [docs/knowledge/wm-review-ticket.md](docs/knowledge/wm-review-ticket.md) |
| Le trascrizioni come fonte | NotebookLM che legge le call, notebook per giorno e per tag, attribuzioni e verifica delle citazioni | [docs/knowledge/wm-transcript-research.md](docs/knowledge/wm-transcript-research.md) |
| Proposta dei tag in `wm-plan` | I due momenti, filtro sulle descrizioni, conferma per singolo tag | [docs/knowledge/wm-plan-proposta-tag.md](docs/knowledge/wm-plan-proposta-tag.md) |
