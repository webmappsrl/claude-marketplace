# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Cos'è questo repo

Marketplace di plugin Claude Code del team Webmapp, pubblicato su GitHub come
`webmappsrl/claude-marketplace`. Distribuisce due plugin:

- **`superpowers`** — framework agentico di Jesse Vincent ([obra/superpowers](https://github.com/obra/superpowers)),
  referenziato direttamente dall'upstream con `ref: main`
- **`wm-skills`** — skill interne Webmapp (convenzioni di codice, checklist PR, processo deploy),
  mantenute in questo repo sotto `plugins/wm-skills/`

## Struttura e ruolo dei file di config

```
.claude-plugin/
  marketplace.json          catalogo del marketplace: nome, owner, lista plugin con sorgente
plugins/
  wm-skills/
    .claude-plugin/
      plugin.json           manifesto del plugin: name, description, author, license, keywords
    skills/
      <nome-skill>/
        SKILL.md            contenuto della skill (frontmatter YAML + corpo Markdown)
```

### `.claude-plugin/marketplace.json`

Definisce il marketplace. Il campo `name` **non deve** contenere prefissi come `claude-*` o
impersonare nomi Anthropic ufficiali — il plugin system li blocca con errore di schema.

### `plugins/wm-skills/.claude-plugin/plugin.json`

**Non ha campo `version` di proposito.** Questo significa che ogni commit su `main` viene trattato
come nuova versione e arriva al team al primo `marketplace update`. Se in futuro volete passare a
release cadenzate, aggiungete `"version": "1.0.0"` e bumpate a ogni release.

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

| Skill A | Skill B | Contratto condiviso |
|---|---|---|
| `wm-plan` | `wm-review-ticket` | `wm-plan` è fonte autoritativa del contratto artefatti `docs/features/<slug>/`. `wm-review-ticket` lo referenzia, non lo duplica. Modificare la struttura artefatti richiede solo aggiornare `wm-plan`. |
| `wm-tag` | `wm-plan` | `wm-tag` invoca `wm-plan` in tag-mode passando titolo, tipo, repo e TAG_ID. `wm-plan` è responsabile di reverse-interaction, overview, challenge, estimation e scrittura della description del ticket. `wm-tag` gestisce tag, lista ticket e loop. |
| `wm-plan` | `wm-tag` | `caso-c` in Fase: ticket switcha su `wm-tag` cedendo il controllo del flusso. |
| `wm-plan` (challenge) | `wm-plan` (review-gate) | Entrambe le sotto-fasi isolano il giudizio in un subagente cieco (solo path/istruzioni, nessun riassunto della conversazione precedente). Se il pattern di isolamento cambia in una sotto-fase, verificare se va aggiornato anche nell'altra. |

### Convenzioni di naming

- **Prefisso obbligatorio `wm-`**: tutte le skill di `wm-skills` devono avere il nome in kebab-case con prefisso `wm-` (es. `wm-plan`, `wm-review-ticket`). Questo vale sia per il nome della cartella che per il campo `name` nel frontmatter di `SKILL.md`.

### Convenzioni di stile per le skill

- **Tono**: imperativo e diretto, rivolto a Claude come agente che esegue istruzioni
- **Lunghezza**: da poche decine di righe (checklist) a qualche centinaio (workflow articolati)
- **Struttura**: sezioni Markdown (`##`) per separare fasi o categorie; elenchi puntati per step atomici
- **Frontmatter**: solo `name` e `description` sono obbligatori; evitare campi extra non necessari

## Validare prima del commit

```bash
claude plugin validate .
```

Verifica la correttezza di `marketplace.json`, `plugin.json` e tutti i `SKILL.md`. Da eseguire
sempre prima di pushare, specialmente dopo aver modificato i file di config.

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

## Skill wm-skills disponibili

| Skill | Quando si attiva |
|---|---|
| `wm-plan` | Implementare, aggiungere o refactorare feature non banali (multi-file, architetturali). Non per bug fix semplici o domande di lettura. |
| `wm-review-ticket` | Eseguire la code review di un ticket Orchestrator — sia quando un collega assegna un ticket da rivedere, sia al termine di una feature wm-plan prima del merge. |
| `wm-tag` | Analizzare trascrizioni/brief cliente per creare tag Orchestrator con ticket figli strutturati e stimati. |

Ogni skill può dipendere o comporre skill di `superpowers` (già installato come plugin separato).

## Integrazione con Orchestrator (sistema ticket Webmapp)

I ticket Webmapp vivono su **Orchestrator** (`webmappsrl/orchestrator`), piattaforma Laravel interna.

**Formato ID ticket:** `oc:<numero>` (es. `oc:7815`)

**Come fornire un ticket a Claude:** incollare il contenuto del ticket nella chat (Titolo, Richiesta, Note di sviluppo). Integrazione diretta via API/MCP pianificata per il futuro.

**Convenzioni nei documenti generati da skill:**
- Ogni file `docs/features/` inizia con `> Ticket: oc:<ID>`
- Feature slug: `<ID>-<titolo-in-kebab-case>` (es. `7815-creazione-poi-tramite-osm-id`)
- Commit scope: `feat(oc:<ID>): ...` / `fix(oc:<ID>): ...` / `refactor(oc:<ID>): ...`

## Aggiornare superpowers

`superpowers` è pinato a `ref: main` di `obra/superpowers`. Per pinnare a una versione stabile,
cambiare `ref` con un tag specifico (es. `"ref": "v5.1.0"`) in `.claude-plugin/marketplace.json`.

## Diagramma di flusso wm-plan

**URL Artifact:** versionato direttamente in `plugins/wm-skills/skills/wm-plan/SKILL.md` → `## Header di sessione` → `### header: diagramma` (non più qui in `CLAUDE.md`). Motivo: l'URL cambia solo quando si modifica la skill, condizione che coincide sempre con un aggiornamento di `SKILL.md` — spostarlo lì elimina il fetch remoto necessario per leggerlo e lo mantiene intrinsecamente allineato ad ogni `/plugin marketplace update`, senza rischio di staleness.

Questo Artifact mostra un diagramma Mermaid del workflow `wm-plan` (fasi ed esecuzione).

**Regola di rigenerazione (obbligatoria per Claude Code):** dopo qualsiasi modifica ai file di questo repo effettuata in una sessione, prima di concludere il lavoro, rigenera l'Artifact del diagramma di flusso usando lo stesso `file_path` e la stessa chiamata di pubblicazione già usata in precedenza (redeploy sullo stesso URL, mai un nuovo Artifact). Questo vale per qualsiasi modifica al repo, non solo per modifiche a `wm-plan`.

**Gestione errori (fail-soft):** la pubblicazione avviene per tentativo diretto, senza check preventivo dell'account attivo. Se il redeploy fallisce, avvisa l'utente con `⚠️ Impossibile aggiornare l'Artifact del diagramma — potrebbe servire switchare all'account Claude del team Webmapp.` e prosegui comunque con il resto della sessione, senza bloccare.

**Template grafico congelato (obbligatorio per Claude Code):** il layout definito in oc:8283 — due colonne di pari altezza (diagramma Mermaid a sinistra, dettaglio fasi a destra con scroll interno), legenda a piena larghezza sotto, palette e tipografia definite — è la versione approvata e **non va ridisegnato** ad ogni rigenerazione. Ogni redeploy successivo deve limitarsi ad aggiornare il **contenuto** (nodi del diagramma, paragrafi di dettaglio, se una fase viene aggiunta/rinominata/rimossa in `wm-plan/SKILL.md`) mantenendo invariati struttura HTML, CSS e stile visivo. Se una modifica al repo non tocca il workflow di `wm-plan`, non è necessario alcun cambiamento al contenuto del diagramma stesso, ma il redeploy va comunque eseguito (vedi regola di rigenerazione sopra) per mantenere l'Artifact "vivo" sullo stesso URL.

Alla prima pubblicazione riuscita (o ad ogni redeploy con URL diverso, caso che non dovrebbe verificarsi con un redeploy corretto), aggiorna il campo **URL Artifact** in `SKILL.md` (non più qui) con il link reale.

## Decisioni architetturali

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
| Ask user to set ticket status to progress in wm-plan | oc:7973 | `plugins/wm-skills/skills/wm-plan/SKILL.md` | Chiede all'utente di mettere il ticket in progress al termine della Fase 0; unifica le credenziali Orchestrator in `orchestrator-auth.json` |
| wm-review-ticket skill | oc:8068 | `plugins/wm-skills/skills/wm-review-ticket/SKILL.md`, `plugins/wm-skills/skills/wm-plan/SKILL.md` | Nuova skill per code review strutturata di ticket Orchestrator; contratto artefatti via WebFetch su wm-plan; stash automatico pre-checkout; review opzionale in wm-plan Fase 6d |
| wm-plan slug e environment-setup | oc:8102 | `plugins/wm-skills/skills/wm-plan/SKILL.md` | Migrazione fasi a slug inglesi; nuova Fase: environment-setup con project-detection, domain-mapping, ux-ui-detection, docker-check |
| wm-tag skill e fase estimation in wm-plan | oc:8157 | `plugins/wm-skills/skills/wm-tag/SKILL.md`, `plugins/wm-skills/skills/wm-plan/SKILL.md` | Nuova skill `wm-tag` per trascrizione → tag + ticket; `caso-c` in Fase: ticket; `Fase: estimation` per Feature; tag-mode in wm-plan |
| Rivedere criteri di stima ore in wm-plan | oc:8278 | `plugins/wm-skills/skills/wm-plan/SKILL.md` | Classificazione per-componente (scrittura pura/decisioni aperte) con buffer per-componente invece di forfettario; pianificazione misurata via timestamp; marcatore versione `[stima v2 — per-componente]`; `execution: re-estimation` per revisioni mid-execution |
| Clear del context in wm-plan dopo reverse-interaction e dopo implementation | oc:8282 | `plugins/wm-skills/skills/wm-plan/SKILL.md` | Isolamento sempre attivo del riepilogo diff in `execution: review-gate` tramite subagente cieco (stesso pattern di `challenge`); fallback esplicito se `planning_start_at` non è stato registrato in `Fase: estimation` |
| Intestazione ASCII per skill wm-plan con info versione/aggiornamenti e diagramma di flusso | oc:8283 | `plugins/wm-skills/skills/wm-plan/SKILL.md`, `CLAUDE.md` | Nuova sezione "Header di sessione" (banner ASCII, check versione via hash HEAD vs remoto, modalità sviluppo locale, link Artifact diagramma); nuova sezione `## Diagramma di flusso wm-plan` in CLAUDE.md con URL e regola di rigenerazione post-modifica repo |
| Fix diagramma header cross-repo in wm-plan | — (nessun ticket) | `plugins/wm-skills/skills/wm-plan/SKILL.md` | `### header: diagramma` legge sempre il `CLAUDE.md` di `claude-marketplace` via fetch remoto (`curl -sf`), non più il `CLAUDE.md` del repo target — corregge il falso "non ancora pubblicato" quando wm-plan è invocato da altri repo |
