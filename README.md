# Acme Team Marketplace

Marketplace di plugin Claude Code per il team Acme. Include:

- **`superpowers`** — il framework di Jesse Vincent ([obra/superpowers](https://github.com/obra/superpowers)), agganciato dal repo upstream
- **`team-skills`** — le nostre skill interne (stile codice, checklist PR, processo deploy)

## Installazione (una tantum, per ciascuno del team)

In Claude Code:

```
/plugin marketplace add acme-org/team-marketplace
/plugin install superpowers@acme-team-marketplace
/plugin install team-skills@acme-team-marketplace
```

Sostituire `acme-org/team-marketplace` con il nostro repo reale.

## Aggiornamenti

```
/plugin marketplace update acme-team-marketplace
```

Gli aggiornamenti di Superpowers arrivano automaticamente dal repo `obra/superpowers` perché lo referenziamo con `ref: main`. Per pinnarci a una versione stabile (consigliato in produzione), cambiare `ref` con un tag specifico, es. `v5.1.0`, nel `marketplace.json`.

## Aggiungere una nuova skill del team

1. Creare una nuova cartella sotto `plugins/team-skills/skills/<nome-skill>/`
2. Aggiungere un `SKILL.md` con frontmatter YAML (`name`, `description`)
3. Bump della `version` in `plugins/team-skills/.claude-plugin/plugin.json` (es. `0.1.0` → `0.2.0`), altrimenti i membri del team non riceveranno la nuova skill
4. Commit + push su `main`
5. Il team esegue `/plugin marketplace update`

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
└── README.md
```
