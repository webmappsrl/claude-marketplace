# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Cos'è questo repo

Marketplace di plugin Claude Code del team Webmapp, pubblicato su GitHub come
`webmappsrl/claude-marketplace`. Distribuisce due plugin:

- **`superpowers`** — framework agentico di Jesse Vincent ([obra/superpowers](https://github.com/obra/superpowers)),
  referenziato direttamente dall'upstream con `ref: main`
- **`team-skills`** — skill interne Webmapp (convenzioni di codice, checklist PR, processo deploy),
  mantenute in questo repo sotto `plugins/team-skills/`

## Struttura e ruolo dei file di config

```
.claude-plugin/
  marketplace.json          catalogo del marketplace: nome, owner, lista plugin con sorgente
plugins/
  team-skills/
    .claude-plugin/
      plugin.json           manifesto del plugin: name, description, author, license, keywords
    skills/
      <nome-skill>/
        SKILL.md            contenuto della skill (frontmatter YAML + corpo Markdown)
```

### `.claude-plugin/marketplace.json`

Definisce il marketplace. Il campo `name` **non deve** contenere prefissi come `claude-*` o
impersonare nomi Anthropic ufficiali — il plugin system li blocca con errore di schema.

### `plugins/team-skills/.claude-plugin/plugin.json`

**Non ha campo `version` di proposito.** Questo significa che ogni commit su `main` viene trattato
come nuova versione e arriva al team al primo `marketplace update`. Se in futuro volete passare a
release cadenzate, aggiungete `"version": "1.0.0"` e bumpate a ogni release.

## Aggiungere una nuova skill del team

1. Creare la cartella:
   ```
   plugins/team-skills/skills/<nome-skill>/
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
/plugin install team-skills@webmapp-marketplace
```

Questo punta al filesystem locale. Ogni modifica a un `SKILL.md` è immediatamente disponibile
ricaricando la sessione, senza commit né push.

Per tornare alla versione remota:
```bash
/plugin marketplace remove webmapp-marketplace
/plugin marketplace add webmappsrl/claude-marketplace
```

## Skill team-skills disponibili

| Skill | Quando si attiva |
|---|---|
| `webmapp-feature-workflow` | Implementare, aggiungere o refactorare feature non banali (multi-file, architetturali). Non per bug fix semplici o domande di lettura. |

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
