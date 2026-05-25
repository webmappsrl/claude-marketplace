# Webmapp Claude Marketplace

Marketplace di plugin Claude Code di Webmapp. Include:

- **`superpowers`** — il framework di Jesse Vincent ([obra/superpowers](https://github.com/obra/superpowers)), agganciato dal repo upstream
- **`team-skills`** — le skill interne del team Webmapp (stile codice, checklist PR, processo deploy)

## Installazione (una tantum, per ciascuno del team)

In Claude Code:

```
/plugin marketplace add webmappsrl/claude-marketplace
/plugin install superpowers@claude-marketplace
/plugin install team-skills@claude-marketplace
```

## Aggiornamenti

```
/plugin marketplace update claude-marketplace
```

Gli aggiornamenti di Superpowers arrivano automaticamente dal repo `obra/superpowers` perché lo referenziamo con `ref: main`. Per pinnare a una versione stabile, cambiare `ref` con un tag specifico (es. `v5.1.0`) nel `marketplace.json`.

Il plugin `team-skills` invece **non ha** un campo `version` nel suo `plugin.json`: questo significa che ogni nuovo commit su `main` è automaticamente trattato come nuova versione e arriva al team al primo `marketplace update`. Quando vorrete passare a release cadenzate, aggiungete `"version": "1.0.0"` nel `plugin.json` e ricordatevi di bumparlo a ogni release.

## Aggiungere una nuova skill del team

1. Creare una nuova cartella sotto `plugins/team-skills/skills/<nome-skill>/`
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
│   └── marketplace.json          ← catalogo dei plugin
├── plugins/
│   └── team-skills/
│       ├── .claude-plugin/
│       │   └── plugin.json       ← manifesto del plugin del team
│       └── skills/
│           ├── our-code-style/SKILL.md
│           ├── our-pr-checklist/SKILL.md
│           └── our-deploy-process/SKILL.md
├── LICENSE
└── README.md
```

## Licenza

MIT — vedi `LICENSE`.
