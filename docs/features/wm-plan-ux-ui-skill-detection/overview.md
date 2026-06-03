# Design: rilevazione automatica UX/UI in wm-plan

## Obiettivo

Aggiungere a `wm-plan` la capacità di rilevare automaticamente quando una feature tocca componenti UI/UX (Vue, Angular, HTML, CSS) e invocare una skill specializzata (`ui-ux-pro-max`) per ottenere un parere prima di pianificare l'implementazione.

---

## Componenti coinvolti

| File | Modifica |
|---|---|
| `plugins/wm-skills/skills/wm-plan/SKILL.md` | Aggiunta logica di rilevazione UX in Fase 1 e Fase 2a |
| `.claude-plugin/marketplace.json` | Aggiunta plugin `ui-ux-pro-max` |

---

## Rilevazione — tre livelli

### Livello 0 — Richiesta esplicita (priorità massima, qualsiasi fase)

Se l'utente menziona esplicitamente UI/UX nella sua richiesta (es. "migliora l'interfaccia", "ridisegna il componente", "fix CSS", "cambia il layout", "aggiungi animazione"), wm-plan invoca immediatamente la skill UX **senza attendere la rilevazione automatica**, indipendentemente dai file coinvolti e dallo stack.

### Livello 1 — Repo-level (Fase 1, lettura CLAUDE.md)

Durante la lettura del `CLAUDE.md` del progetto target, wm-plan legge `package.json` (se presente):

```bash
cat package.json | jq '.dependencies + .devDependencies | keys[]' 2>/dev/null
```

Segnali che impostano `stack_ui: true`:
- `@angular/core` → Angular
- `vue` o `@vue/core` → Vue
- `react` → React (non primario per Webmapp ma gestito)

Se `package.json` non esiste (es. progetto Laravel puro), fallback a: cerca `resources/views/` o `resources/js/` nel filesystem.

Il flag `stack_ui` rimane attivo per tutta la sessione.

### Livello 2 — File-level (Fase 2a, dopo classificazione moduli)

Dopo aver identificato i file toccati dalla feature, controlla:

**Segnali forti** → confidenza alta, invocazione automatica:
- Estensioni: `.vue`, `.html`, `.css`, `.scss`, `.component.ts`, `.component.html`, `.component.scss`
- Cartelle: `src/components/`, `src/views/`, `resources/views/`, `resources/js/components/`, `src/app/`
- Pattern Angular: `*.module.css`, `*.styles.ts`

**Segnali deboli** → confidenza bassa, chiede conferma:
- Solo file `.css` di utility/variabili (es. `variables.scss`, `_tokens.css`)
- Menzione di "interfaccia" nel ticket senza file concreti ancora identificati

### Livello 3 — Intent-level (Fase 2a)

Se `stack_ui: true` (dal livello 1) e i moduli toccati includono anche solo parzialmente componenti frontend, tratta come confidenza alta anche senza estensioni esplicite.

---

## Invocazione skill UX

### Lookup dinamico

Prima di invocare, wm-plan cerca la skill installata:

```
skill disponibile: ui-ux-pro-max
```

Se trovata → invoca con contesto (ticket + moduli toccati UI).  
Se non trovata → fallback (vedi sotto).

### Flusso completo

```
SE confidenza alta E skill installata:
  → log: "Rilevati componenti UI ([file/cartelle trovate]) — invoco ui-ux-pro-max"
  → invoca skill con: titolo ticket, lista file UI coinvolti, stack (Vue/Angular)
  → il parere UX confluisce in "Requisiti" e "Rischi" di overview.md

SE confidenza alta E skill NON installata:
  → messaggio: "⚠️ Rilevati componenti UI ma la skill UX non è installata.
     Installa con: /plugin install ui-ux-pro-max@wm-marketplace
     Procedo con giudizio interno — ti consiglio di installare la skill per i prossimi ticket."
  → continua con giudizio interno su UX

SE confidenza bassa:
  → chiede: "Ho trovato [file X] — potrebbe richiedere attenzione UX.
     Vuoi che invochi la skill ui-ux-pro-max?"
  → SE sì: stesso flusso lookup sopra
  → SE no: procede normalmente
```

### Contesto passato alla skill

Quando invocata, wm-plan passa:
- Titolo e tipo del ticket
- Lista dei file/componenti UI coinvolti
- Stack rilevato (Vue / Angular / entrambi)
- Eventuale `customer_request` del ticket (contesto utente finale)

---

## Aggiornamento marketplace.json

Aggiungere ai `plugins`:

```json
{
  "name": "ui-ux-pro-max",
  "source": {
    "source": "github",
    "repo": "nextlevelbuilder/ui-ux-pro-max-skill",
    "ref": "main"
  },
  "description": "Skill UX/UI professionale: design system, palette, tipografia, checklist pre-consegna"
}
```

Comando di installazione per il team:
```
/plugin install ui-ux-pro-max@wm-marketplace
```

---

## Posizione nel workflow wm-plan

| Fase | Aggiunta |
|---|---|
| Fase 1 (leggi CLAUDE.md) | Rilevazione stack UI a livello repo (`stack_ui` flag) |
| Fase 2a (classificazione moduli) | Rilevazione file-level + intent-level, invocazione skill |
| Fase 4 (overview.md) | Il parere UX ricevuto popola "Requisiti" e "Rischi" |

---

## Out of scope

- Non viene invocata durante la Fase 6 (esecuzione) — il parere UX deve influenzare la pianificazione, non solo l'implementazione
- Non blocca il workflow se la skill non è installata (graceful degradation)
- Non sostituisce il brainstorming o la challenge adversariale

---

## Criteri di successo

- Su un ticket con file `.vue` o `.component.ts`, wm-plan invoca automaticamente la skill senza che l'utente lo chieda
- Su un ticket puramente backend, wm-plan non menziona UX
- Se la skill non è installata, il messaggio di fallback include il comando di installazione corretto
