# wm-plan UX/UI Skill Detection — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Aggiungere a `wm-plan` la rilevazione automatica di feature UX/UI e l'invocazione della skill `ui-ux-pro-max`, più l'aggiunta del plugin al marketplace Webmapp.

**Architecture:** Due modifiche indipendenti e sequenziali: (1) aggiunta di `ui-ux-pro-max` a `marketplace.json`; (2) inserimento della logica di rilevazione UX in `SKILL.md` — in Fase 1 (stack detection via `package.json`) e in Fase 2a (file-level + intent detection + invocazione skill).

**Tech Stack:** Markdown (SKILL.md), JSON (marketplace.json). Nessun codice eseguibile — entrambi i file sono letti da Claude Code come istruzioni.

---

## File Structure

| File | Tipo modifica | Responsabilità |
|---|---|---|
| `.claude-plugin/marketplace.json` | Modify | Aggiunta plugin `ui-ux-pro-max` |
| `plugins/wm-skills/skills/wm-plan/SKILL.md` | Modify | Logica rilevazione UX in Fase 1 e Fase 2a |

---

## Task 1: Aggiungi `ui-ux-pro-max` al marketplace

**Files:**
- Modify: `.claude-plugin/marketplace.json`

- [ ] **Step 1: Apri il file e individua la sezione `plugins`**

Il file si trova in `.claude-plugin/marketplace.json`. Attualmente contiene due plugin: `superpowers` e `wm-skills`.

- [ ] **Step 2: Aggiungi il terzo plugin**

Aggiungi dopo l'entry `wm-skills`:

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

Il risultato finale della sezione `plugins` deve essere:

```json
"plugins": [
  {
    "name": "superpowers",
    "source": {
      "source": "github",
      "repo": "obra/superpowers",
      "ref": "main"
    },
    "description": "Agentic skills framework (upstream, di Jesse Vincent / obra)"
  },
  {
    "name": "wm-skills",
    "source": "./plugins/wm-skills",
    "description": "Skill interne del team Webmapp: stile, PR, deploy"
  },
  {
    "name": "ui-ux-pro-max",
    "source": {
      "source": "github",
      "repo": "nextlevelbuilder/ui-ux-pro-max-skill",
      "ref": "main"
    },
    "description": "Skill UX/UI professionale: design system, palette, tipografia, checklist pre-consegna"
  }
]
```

- [ ] **Step 3: Valida il JSON**

```bash
cat .claude-plugin/marketplace.json | jq .
```

Expected: output JSON formattato senza errori.

- [ ] **Step 4: Valida il plugin**

```bash
claude plugin validate .
```

Expected: nessun errore di schema.

---

## Task 2: Aggiungi rilevazione stack UI in Fase 1 di SKILL.md

**Files:**
- Modify: `plugins/wm-skills/skills/wm-plan/SKILL.md`

La Fase 1 attuale termina con:
> "Tieni queste informazioni attive per tutto il workflow."

- [ ] **Step 1: Aggiungi il blocco di rilevazione stack UI alla fine di Fase 1**

Dopo la riga `- Tieni queste informazioni attive per tutto il workflow.`, aggiungi:

```markdown
### Rilevazione stack UI (eseguita in Fase 1)

Dopo aver letto `CLAUDE.md`, rileva se il progetto ha uno stack frontend:

```bash
cat package.json 2>/dev/null | jq -r '(.dependencies // {}) + (.devDependencies // {}) | keys[]' | grep -E "^(@angular/core|vue|@vue/core|react)$"
```

Se `package.json` non esiste o non contiene dipendenze frontend, verifica la presenza di cartelle frontend Laravel:

```bash
ls resources/views/ resources/js/components/ 2>/dev/null | head -3
```

**Imposta internamente il flag `stack_ui`:**
- `stack_ui: angular` — se trovato `@angular/core`
- `stack_ui: vue` — se trovato `vue` o `@vue/core`
- `stack_ui: react` — se trovato `react`
- `stack_ui: laravel-blade` — se trovata `resources/views/` senza dipendenze JS frontend
- `stack_ui: false` — se nessun segnale trovato

Questo flag rimane attivo per tutto il workflow e influenza Fase 2a.
```

- [ ] **Step 2: Verifica che la Fase 1 abbia ancora senso come blocco coeso**

Rileggi la sezione `## Fase 1 — Leggi CLAUDE.md` nel file modificato. La nuova sottosezione deve apparire dopo i tre bullet esistenti, non prima.

---

## Task 3: Aggiungi logica UX detection in Fase 2a di SKILL.md

**Files:**
- Modify: `plugins/wm-skills/skills/wm-plan/SKILL.md`

La Fase 2a attuale termina con:
> "Questa classificazione rimane attiva per tutto il workflow: overview.md, plan.md e ogni step del piano devono sempre indicare il repo di destinazione per ogni file."

- [ ] **Step 1: Aggiungi il blocco UX detection dopo la classificazione dominio in Fase 2a**

Dopo il paragrafo "Questa classificazione rimane attiva per tutto il workflow…", aggiungi:

```markdown
### UX/UI Detection (eseguita al termine di Fase 2a)

Dopo aver classificato i moduli toccati, esegui il seguente controllo in ordine di priorità:

**Livello 0 — Richiesta esplicita (priorità massima)**

Se la richiesta dell'utente o il titolo/body del ticket contiene parole come: `UI`, `UX`, `interfaccia`, `componente`, `layout`, `form`, `modal`, `stile`, `CSS`, `design`, `animazione`, `schermata` → confidenza alta, procedi direttamente all'invocazione.

**Livello 1 — Stack UI rilevato in Fase 1 + file coinvolti**

Se `stack_ui != false` E almeno uno dei file/moduli toccati dalla feature appartiene a:
- Estensioni: `.vue`, `.html`, `.css`, `.scss`, `.component.ts`, `.component.html`, `.component.scss`
- Cartelle: `src/components/`, `src/views/`, `resources/views/`, `resources/js/components/`, `src/app/`

→ confidenza alta, procedi all'invocazione automatica.

**Livello 2 — Solo stack UI, nessun file frontend esplicito**

Se `stack_ui != false` ma i file toccati sono solo backend/logica → confidenza bassa.

**Livello 3 — Nessun segnale**

Se `stack_ui: false` e nessun file frontend → nessuna azione UX.

---

**Flusso di invocazione:**

```
SE confidenza alta:
  → scrivi: "Rilevati componenti UI ([file/stack trovati]) — cerco skill UX specializzata."
  → cerca tra le skill disponibili: `ui-ux-pro-max`
  → SE trovata:
      invoca la skill con questo contesto:
      - Titolo ticket e tipo
      - Stack rilevato (vue / angular / react / laravel-blade)
      - Lista file/componenti UI coinvolti
      - customer_request del ticket (se disponibile)
      Il parere UX ricevuto confluisce nelle sezioni "Requisiti" e "Rischi" di overview.md
  → SE non trovata:
      scrivi: "⚠️ La skill ui-ux-pro-max non è installata.
      Per ottenerla: /plugin install ui-ux-pro-max@wm-marketplace
      Procedo con giudizio interno UX — installa la skill per i prossimi ticket."
      Applica il tuo giudizio interno su UX per i Requisiti e Rischi dell'overview.

SE confidenza bassa:
  → chiedi: "Ho rilevato uno stack frontend ([stack]) ma i file toccati sembrano principalmente backend.
     Questa feature ha componenti UI? Vuoi che invochi la skill UX specializzata?"
  → SE sì: stesso flusso lookup sopra
  → SE no: procedi senza UX detection
```
```

- [ ] **Step 2: Verifica coerenza con il resto di Fase 2a**

Il nuovo blocco deve apparire come terza sottosezione di `## Fase 2 — Reverse Interaction`, dopo `### 2a — Mappatura domini e submodule` e prima di `### 2b — Dialogo`. Controlla che i titoli delle sottosezioni siano allineati.

---

## Task 4: Aggiorna la sezione "Composizione con altre skill"

**Files:**
- Modify: `plugins/wm-skills/skills/wm-plan/SKILL.md`

- [ ] **Step 1: Aggiungi `ui-ux-pro-max` alla sezione composizione**

Trova la sezione:
```markdown
## Composizione con altre skill Webmapp

- **`wm-skills:our-code-style`** — applica in Fase 5 (scrittura plan) e Fase 6 (esecuzione)
- **`wm-skills:our-pr-checklist`** — applica dopo la Fase 7, prima di aprire la PR
- **`wm-skills:our-deploy-post-merge`** — applica dopo il merge della PR
```

Aggiungi in cima alla lista:

```markdown
- **`ui-ux-pro-max`** — invocata automaticamente in Fase 2a quando rilevati componenti UI/UX (Vue, Angular, HTML/CSS). Richiede `/plugin install ui-ux-pro-max@wm-marketplace` se non installata.
```

---

## Task 5: Valida e committa

**Files:**
- Read: `plugins/wm-skills/skills/wm-plan/SKILL.md`
- Read: `.claude-plugin/marketplace.json`

- [ ] **Step 1: Valida il plugin**

```bash
claude plugin validate .
```

Expected: nessun errore.

- [ ] **Step 2: Mostra il diff completo all'utente e attendi approvazione esplicita prima di procedere**

```bash
git diff
```

Presenta un riepilogo strutturato dei file modificati. Attendi conferma esplicita ("procedi", "sì", o equivalente) prima di eseguire qualsiasi commit.

- [ ] **Step 3: Solo dopo approvazione — esegui i commit**

```bash
git add .claude-plugin/marketplace.json
git commit -m "feat(wm-plan): add ui-ux-pro-max plugin to marketplace"

git add plugins/wm-skills/skills/wm-plan/SKILL.md
git commit -m "feat(wm-plan): add automatic UX/UI skill detection in Phase 1 and 2a"
```

- [ ] **Step 4: Apri PR verso `develop`**

```bash
git push -u origin <nome-branch>
gh pr create --base develop --title "feat(wm-plan): automatic UX/UI skill detection" --body "$(cat <<'EOF'
## Summary
- Aggiunto plugin `ui-ux-pro-max` al marketplace Webmapp
- wm-plan rileva automaticamente stack UI (Vue/Angular) via `package.json` in Fase 1
- wm-plan invoca `ui-ux-pro-max` in Fase 2a quando rileva componenti UI nei file toccati
- Fallback graceful se la skill non è installata (messaggio + giudizio interno)

## Test plan
- [ ] Invocare wm-plan su un progetto Vue: verificare che in Fase 1 rilevi `stack_ui: vue`
- [ ] Verificare che in Fase 2a, con file `.vue` nei moduli toccati, invochi automaticamente la skill
- [ ] Invocare wm-plan su un progetto Laravel puro: verificare che `stack_ui: false` e nessuna azione UX
- [ ] Verificare il messaggio di fallback disinstallando temporaneamente `ui-ux-pro-max`

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```
