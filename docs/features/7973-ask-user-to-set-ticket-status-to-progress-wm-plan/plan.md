> Ticket: oc:7973

# Ask user to set ticket status to progress when work begins in wm-plan skill — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Modificare `SKILL.md` della skill `wm-plan` per chiedere all'utente se vuole impostare lo status del ticket a "progress" e assegnarsi il ticket al termine della Fase 0, e unificare le credenziali Orchestrator in un unico file JSON.

**Architecture:** Tutte le modifiche sono in un unico file Markdown (`SKILL.md`). Le modifiche toccano tre aree: (1) la sezione `## Orchestrator API` per aggiornare login e lettura credenziali al file JSON unificato, (2) la Fase 0 per aggiungere la domanda sul progress dopo il riepilogo del ticket in entrambi i casi A e B.

**Tech Stack:** Markdown, bash (curl, jq), file system (`~/.config/webmapp/`)

---

### Task 1: Aggiorna sezione Configurazione — riferimento al file JSON unificato

**Files:**
- Modify: `plugins/wm-skills/skills/wm-plan/SKILL.md` (linee 22-23)

- [ ] **Step 1: Sostituisci il riferimento al file token nella sezione Configurazione**

Trova questo testo (linee 22-23):
```
- **Token:** salvato in `~/.config/webmapp/orchestrator-token` (plain text, una riga). Se il file non esiste o la chiamata restituisce 401, esegui il login (vedi sotto).
```

Sostituiscilo con:
```
- **Auth:** salvato in `~/.config/webmapp/orchestrator-auth.json` (JSON con campi `token`, `id`, `name`, `email`). Se il file non esiste o la chiamata restituisce 401, esegui il login (vedi sotto). Se esiste solo il file legacy `~/.config/webmapp/orchestrator-token`, esegui la migrazione (vedi sotto).
```

---

### Task 2: Aggiorna la procedura di Login per salvare file JSON unificato

**Files:**
- Modify: `plugins/wm-skills/skills/wm-plan/SKILL.md` (sezione `### Login`)

- [ ] **Step 1: Sostituisci l'intera sezione `### Login`**

Trova questo testo:
```markdown
### Login (solo se token assente o scaduto)

Chiedi email e password all'utente, poi:

```bash
ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-https://orchestrator.maphub.it}"
curl -s -X POST "$ORCHESTRATOR_URL/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"<email>","password":"<password>"}' \
  | jq -r '.token'
```

Salva il token restituito:

```bash
mkdir -p ~/.config/webmapp
echo "<token>" > ~/.config/webmapp/orchestrator-token
```
```

Sostituiscilo con:
```markdown
### Login (solo se auth assente o scaduto)

Chiedi email e password all'utente, poi:

```bash
ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-https://orchestrator.maphub.it}"
TOKEN=$(curl -s -X POST "$ORCHESTRATOR_URL/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"<email>","password":"<password>"}' \
  | jq -r '.token')
USER=$(curl -s -X GET "$ORCHESTRATOR_URL/api/me" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Accept: application/json")
echo $USER | jq --arg token "$TOKEN" '. + {token: $token}' > ~/.config/webmapp/orchestrator-auth.json
```

### Migrazione da file legacy (solo se `orchestrator-auth.json` assente ma `orchestrator-token` presente)

```bash
ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-https://orchestrator.maphub.it}"
TOKEN=$(cat ~/.config/webmapp/orchestrator-token)
USER=$(curl -s -X GET "$ORCHESTRATOR_URL/api/me" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Accept: application/json")
echo $USER | jq --arg token "$TOKEN" '. + {token: $token}' > ~/.config/webmapp/orchestrator-auth.json
```

Se `GET /api/me` risponde 401 durante la migrazione, il token legacy è scaduto: cancella `orchestrator-token` ed esegui il login completo sopra.
```

---

### Task 3: Aggiorna tutti i comandi curl che leggono il token

**Files:**
- Modify: `plugins/wm-skills/skills/wm-plan/SKILL.md` (sezioni Lettura ticket, Creazione ticket, Aggiornamento ticket, Fase 0 Caso A)

- [ ] **Step 1: Aggiorna la riga di lettura token nella sezione `### Lettura ticket`**

Trova:
```bash
TOKEN=$(cat ~/.config/webmapp/orchestrator-token)
curl -s -X GET "$ORCHESTRATOR_URL/api/stories/<ID>" \
```

Sostituisci con:
```bash
TOKEN=$(jq -r '.token' ~/.config/webmapp/orchestrator-auth.json)
curl -s -X GET "$ORCHESTRATOR_URL/api/stories/<ID>" \
```

- [ ] **Step 2: Aggiorna la riga di lettura token nella sezione `### Creazione ticket`**

Trova:
```bash
TOKEN=$(cat ~/.config/webmapp/orchestrator-token)
curl -s -X POST "$ORCHESTRATOR_URL/api/stories" \
```

Sostituisci con:
```bash
TOKEN=$(jq -r '.token' ~/.config/webmapp/orchestrator-auth.json)
curl -s -X POST "$ORCHESTRATOR_URL/api/stories" \
```

- [ ] **Step 3: Aggiorna la riga di lettura token nella sezione `### Aggiornamento ticket`**

Trova:
```bash
TOKEN=$(cat ~/.config/webmapp/orchestrator-token)
curl -s -X PATCH "$ORCHESTRATOR_URL/api/stories/<ID>" \
```

Sostituisci con:
```bash
TOKEN=$(jq -r '.token' ~/.config/webmapp/orchestrator-auth.json)
curl -s -X PATCH "$ORCHESTRATOR_URL/api/stories/<ID>" \
```

- [ ] **Step 4: Aggiorna la riga di lettura token nella Fase 0 Caso A**

Trova:
```bash
TOKEN=$(cat ~/.config/webmapp/orchestrator-token 2>/dev/null)
```

Sostituisci con:
```bash
TOKEN=$(jq -r '.token' ~/.config/webmapp/orchestrator-auth.json 2>/dev/null)
```

Aggiorna anche la riga successiva che descrive il fallback:

Trova:
```
Se il token non esiste o la risposta è 401, esegui il login (vedi `## Orchestrator API → Login`) e ritenta.
```

Sostituisci con:
```
Se il file auth non esiste, controlla se esiste `~/.config/webmapp/orchestrator-token` ed esegui la migrazione (vedi `## Orchestrator API → Migrazione da file legacy`). Se la risposta è 401, cancella il file auth ed esegui il login completo.
```

---

### Task 4: Aggiungi la domanda "metti in progress" alla fine della Fase 0 Caso A

**Files:**
- Modify: `plugins/wm-skills/skills/wm-plan/SKILL.md` (sezione `### Caso A`)

- [ ] **Step 1: Individua la fine della sezione Caso A**

La sezione termina con:
```
Mostra all'utente un riepilogo del ticket letto prima di procedere alla Fase 1.
```

- [ ] **Step 2: Aggiungi il blocco "metti in progress" dopo il riepilogo**

Sostituisci:
```
Mostra all'utente un riepilogo del ticket letto prima di procedere alla Fase 1.
```

Con:
```
Mostra all'utente un riepilogo del ticket letto prima di procedere alla Fase 1.

### Domanda progress (Caso A)

Dopo il riepilogo, chiedi:

> "Vuoi impostare lo status del ticket a **progress** e assegnartelo?"

Se l'utente risponde sì, esegui il PATCH seguendo `## Orchestrator API → Aggiornamento ticket` con:

```bash
ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-https://orchestrator.maphub.it}"
TOKEN=$(jq -r '.token' ~/.config/webmapp/orchestrator-auth.json)
USER_ID=$(jq -r '.id' ~/.config/webmapp/orchestrator-auth.json)
curl -s -X PATCH "$ORCHESTRATOR_URL/api/stories/<ID>" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d "{\"status\": \"progress\", \"user_id\": $USER_ID}"
```

Se il PATCH fallisce (risposta non 2xx o errore di rete), avvisa l'utente con un messaggio ("⚠️ Impossibile aggiornare lo status del ticket — procedo comunque con il workflow.") e continua.

Se l'utente risponde no, procedi direttamente alla Fase 1 senza modificare il ticket.
```

---

### Task 5: Aggiungi la domanda "metti in progress" alla fine della Fase 0 Caso B

**Files:**
- Modify: `plugins/wm-skills/skills/wm-plan/SKILL.md` (sezione `### Caso B`)

- [ ] **Step 1: Individua il punto di inserimento nel Caso B**

La sezione Caso B include:
```
**Il ticket va creato prima di procedere alla Fase 1.** I campi `description` e `customer_request` potranno essere aggiornati a fine workflow (Checklist) con le informazioni emerse dalle fasi successive.

Se l'utente non vuole creare il ticket ora, procedi senza ID: usa solo il titolo kebab-case come slug.
```

- [ ] **Step 2: Aggiungi il blocco "metti in progress" dopo la creazione del ticket**

Sostituisci:
```
**Il ticket va creato prima di procedere alla Fase 1.** I campi `description` e `customer_request` potranno essere aggiornati a fine workflow (Checklist) con le informazioni emerse dalle fasi successive.

Se l'utente non vuole creare il ticket ora, procedi senza ID: usa solo il titolo kebab-case come slug.
```

Con:
```
**Il ticket va creato prima di procedere alla Fase 1.** I campi `description` e `customer_request` potranno essere aggiornati a fine workflow (Checklist) con le informazioni emerse dalle fasi successive.

Una volta creato il ticket e salvato l'ID, chiedi:

> "Vuoi impostare lo status del ticket a **progress** e assegnartelo?"

Se l'utente risponde sì, esegui il PATCH seguendo `## Orchestrator API → Aggiornamento ticket` con:

```bash
ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-https://orchestrator.maphub.it}"
TOKEN=$(jq -r '.token' ~/.config/webmapp/orchestrator-auth.json)
USER_ID=$(jq -r '.id' ~/.config/webmapp/orchestrator-auth.json)
curl -s -X PATCH "$ORCHESTRATOR_URL/api/stories/<ID>" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d "{\"status\": \"progress\", \"user_id\": $USER_ID}"
```

Se il PATCH fallisce, avvisa l'utente con un messaggio ("⚠️ Impossibile aggiornare lo status del ticket — procedo comunque con il workflow.") e continua.

Se l'utente non vuole creare il ticket ora, procedi senza ID: usa solo il titolo kebab-case come slug. La domanda progress non viene posta.
```

---

### Task 6: Valida il file SKILL.md

**Files:**
- Read: `plugins/wm-skills/skills/wm-plan/SKILL.md`

- [ ] **Step 1: Verifica che tutti i riferimenti a `orchestrator-token` siano stati rimossi**

```bash
grep -n "orchestrator-token" plugins/wm-skills/skills/wm-plan/SKILL.md
```

Output atteso: nessun risultato (o solo il riferimento nella sezione Migrazione, dove è intenzionale).

- [ ] **Step 2: Verifica che `orchestrator-auth.json` sia presente nelle sezioni giuste**

```bash
grep -n "orchestrator-auth.json" plugins/wm-skills/skills/wm-plan/SKILL.md
```

Output atteso: almeno 6 occorrenze (Configurazione, Login, Lettura ticket, Creazione ticket, Aggiornamento ticket, Fase 0 Caso A, Task 4, Task 5).

- [ ] **Step 3: Valida la sintassi del plugin**

```bash
claude plugin validate .
```

Output atteso: nessun errore di validazione.

- [ ] **Step 4: Commit**

```bash
git add plugins/wm-skills/skills/wm-plan/SKILL.md
git commit -m "feat(oc:7973): ask user to set ticket to progress in Fase 0, unify auth to JSON file"
```
