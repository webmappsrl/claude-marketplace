# Webmapp Claude Marketplace

Marketplace di plugin Claude Code di Webmapp. Include:

- **`superpowers`** — il framework di Jesse Vincent ([obra/superpowers](https://github.com/obra/superpowers)), agganciato dal repo upstream
- **`wm-skills`** — le skill interne del team Webmapp (stile codice, checklist PR, processo deploy)

## Installazione (una tantum, per ciascuno del team)

In Claude Code:

```
/plugin marketplace add webmappsrl/claude-marketplace
/plugin install superpowers@claude-marketplace
/plugin install wm-skills@claude-marketplace
```

## Aggiornamenti

```
/plugin marketplace update claude-marketplace
```

Gli aggiornamenti di Superpowers arrivano automaticamente dal repo `obra/superpowers` perché lo referenziamo con `ref: main`. Per pinnare a una versione stabile, cambiare `ref` con un tag specifico (es. `v5.1.0`) nel `marketplace.json`.

Il plugin `wm-skills` invece **non ha** un campo `version` nel suo `plugin.json`: questo significa che ogni nuovo commit su `main` è automaticamente trattato come nuova versione e arriva al team al primo `marketplace update`. Quando vorrete passare a release cadenzate, aggiungete `"version": "1.0.0"` nel `plugin.json` e ricordatevi di bumparlo a ogni release.

## Skill disponibili

### `wm-skills`

| Skill | Quando usarla |
|---|---|
| `wm-plan` | Prima di implementare qualsiasi feature o refactor non banale. Orchestra l'intero ciclo: ticket Orchestrator → analisi → overview → piano → esecuzione → documentazione. |

Le skill di `superpowers` (brainstorming, writing-plans, executing-plans, ecc.) sono usate internamente da `wm-plan` — raramente serve invocarle direttamente.

## Aggiungere una nuova skill del team

1. Creare una nuova cartella sotto `plugins/wm-skills/skills/<nome-skill>/`
2. Aggiungere un `SKILL.md` con frontmatter YAML (`name`, `description`)
3. Commit + push su `main`
4. Il team esegue `/plugin marketplace update claude-marketplace`

## Validare prima di pushare

```
claude plugin validate .
```

## Struttura del repo

```
.
├── .claude-plugin/
│   └── marketplace.json                      ← catalogo dei plugin
├── plugins/
│   └── wm-skills/
│       ├── .claude-plugin/
│       │   └── plugin.json                   ← manifesto del plugin del team
│       └── skills/
│           └── wm-plan/
│               └── SKILL.md
├── LICENSE
└── README.md
```

## Licenza

MIT — vedi `LICENSE`.
