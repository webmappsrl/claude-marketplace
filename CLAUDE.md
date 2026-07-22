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

## Decisioni architetturali

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
