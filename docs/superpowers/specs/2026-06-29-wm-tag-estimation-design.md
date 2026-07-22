> Ticket: oc:8157

# Design — wm-tag skill e fase estimation in wm-plan

## Panoramica

Due artefatti da produrre in sincronia:

1. **Nuova skill `wm-tag`** — trasforma una trascrizione/brief cliente in un tag Orchestrator con ticket figli analizzati e stimati
2. **Modifiche a `wm-plan`** — aggiunge `caso-c` nella Fase: ticket (entry point verso `wm-tag`) e la nuova `Fase: estimation` per ticket Feature

---

## Artefatto 1 — Skill `wm-tag`

### Trigger e descrizione

Nome: `wm-tag`
Prefisso rispettato: `wm-` (convenzione wm-skills)
Attivazione: quando l'utente ha una trascrizione Meet, un brief cliente o qualsiasi materiale che descrive richieste da trasformare in ticket Orchestrator.

### Flusso completo

#### Fase: input

Accetta due formati:
- **Testo in chat** — incollato direttamente dall'utente
- **Link Google Drive** — letto via `mcp__claude_ai_Google_Drive__read_file_content` o WebFetch secondo disponibilità

In testa alla descrizione del tag va sempre linkata la fonte originale (URL Drive o label "Testo fornito in chat").

#### Fase: repo-map

Prima di qualsiasi analisi, costruisce o aggiorna il dizionario `~/.config/webmapp/repos.json`:

```bash
# Struttura del file
{
  "nome-repo": "/path/assoluta/al/repo",
  ...
}
```

**Algoritmo:**
1. Leggi `~/.config/webmapp/repos.json` se esiste
2. Individua la cartella padre del working directory corrente
3. Esplora i sottodir con `find <parent> -maxdepth 1 -type d -name "*.git" -o -maxdepth 2 -name ".git"` per trovare repo git
4. Per ogni repo trovato non presente nel dizionario, aggiungilo (nome = basename della cartella)
5. Salva il file aggiornato
6. Mostra all'utente la mappa costruita prima di procedere

La mappa rimane disponibile per tutto il flusso — ogni volta che serve ispezionare codice in un altro repo, si consulta il dizionario.

#### Fase: client-extraction

Estrae il nome del cliente dal testo del brief/trascrizione (es. "Cammini", "Parchi", "Regione Veneto").

- Proposta automatica con motivazione ("Ho trovato il nome cliente: **CAMMINI** — citato 4 volte nel testo")
- Conferma esplicita dell'utente prima di usarlo nel naming

#### Fase: tag-naming

Costruisce il nome del tag seguendo la convenzione: `[RDO][CLIENTE][ANNO]<N>`

```bash
# Cerca tag esistenti per stesso cliente+anno
ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-https://orchestrator.maphub.it}"
TOKEN=$(jq -r '.token' ~/.config/webmapp/orchestrator-auth.json)
curl -s "$ORCHESTRATOR_URL/api/tags?search=RDO" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Accept: application/json"
```

Filtra i risultati per `[CLIENTE][ANNO]`, trova il numero più alto, propone `N = max + 1`.
Mostra il nome proposto e attende conferma.

#### Fase: tag-description

Analizza il testo e produce la descrizione del tag in Markdown:

```markdown
**Fonte:** <link o "Testo fornito in chat">

## Macro aree

- **<Area 1>**: <descrizione sintetica>
- **<Area 2>**: <descrizione sintetica>
...

## Informazioni principali

<Contesto generale del cliente, obiettivi, vincoli noti, deadline se presenti>
```

Mostra la descrizione all'utente e attende approvazione prima di creare il tag.

#### Fase: tag-creation

Prima di creare il tag, mostra sempre un riepilogo tabellare e chiedi conferma esplicita:

> **Creazione tag**
>
> | Campo | Valore |
> |-------|--------|
> | `name` | `[RDO][CLIENTE][ANNO]N` |
> | `description` | `<anteprima primi 200 caratteri...>` |
>
> Procedo?

Solo dopo la conferma:

```bash
curl -s -X POST "$ORCHESTRATOR_URL/api/tags" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "<nome-tag>", "description": "<descrizione>"}'
```

Salva l'ID del tag restituito per l'associazione dei ticket.

**Regola generale scritture su Orchestrator (tag e story):** qualsiasi operazione POST o PATCH — creazione tag, aggiornamento tag, creazione ticket, aggiornamento ticket, associazione ticket a tag — richiede sempre un preview tabellare dei campi modificati e conferma esplicita dell'utente prima di eseguire la chiamata HTTP. Nessuna eccezione.

#### Fase: ticket-list

Individua tutti i task dalla trascrizione e presenta la lista proposta:

| # | Titolo ticket | Tipo | Repo stimato |
|---|---|---|---|
| 1 | ... | Feature / Bug | backend / frontend |
| 2 | ... | Feature | backend |

L'utente può approvare, modificare, unire o eliminare ticket dalla lista prima di procedere.

#### Fase: ticket-loop

Per ogni ticket approvato, invoca `wm-skills:wm-plan` in **tag-mode**:

- Passa il contesto: titolo ticket, tipo, repo di destinazione, ID tag padre
- `wm-plan` esegue il flusso canonico fino all'overview (incluse `reverse-interaction` e `challenge`)
- L'overview **non viene salvata nel repo** ma scritta nel campo `description` del ticket Orchestrator in una sezione `## Overview`
- Se il ticket è di tipo Feature, `wm-plan` esegue anche `Fase: estimation` prima di fermarsi
- Dopo ogni ticket, chiede all'utente se procedere al successivo o fermarsi

---

## Artefatto 2 — Modifiche a `wm-plan`

### Modifica A — `caso-c` nella Fase: ticket

Aggiunge una terza opzione al menu iniziale di `wm-plan`:

> Come vuoi procedere?
> - **A)** Ho un ticket esistente (`oc:<ID>`)
> - **B)** Voglio creare un nuovo ticket
> - **C)** Ho una trascrizione/brief cliente → crea tag con più ticket

Se l'utente sceglie C, `wm-plan` invoca immediatamente `wm-skills:wm-tag` passando l'input già fornito e cede il controllo del flusso.

### Modifica B — `Fase: estimation` (solo Feature)

Posizione nel flusso: dopo `Fase: challenge`, prima di `Fase: write-plan`.
Eseguita solo se il ticket è di tipo Feature (controllato dal campo `type`).

**Flusso:**

1. Legge l'overview approvato (da file o da description del ticket in tag-mode)
2. Produce una stima ragionata in ore con questa struttura:

```
**Stima proposta: <N> ore**

Motivazione:
- <componente 1>: <X>h — <motivazione>
- <componente 2>: <Y>h — <motivazione>
- Buffer rischio: <Z>h — <motivazione>

Confidenza: alta / media / bassa
```

3. Chiede al dev: "Accetti questa stima o vuoi modificarla?"
4. La stima accettata viene scritta nel campo `estimated_hours` del ticket Orchestrator via PATCH
5. Prosegue normalmente con `Fase: write-plan` (o si ferma se in tag-mode)

### Modifica C — tag-mode

Quando `wm-plan` è invocato da `wm-tag` riceve un flag `tag-mode: true` nel contesto.

Comportamento in tag-mode:
- Salta `Fase: write-plan`, `Fase: execution`, `Fase: notes`, `Fase: update-context`
- Dopo l'overview approvato (e la stima se Feature), scrive l'overview nel campo `description` del ticket Orchestrator in una sezione `## Overview`
- Associa il ticket al tag padre via `tags: [<tag-id>]` nel payload PATCH
- Si ferma e restituisce il controllo a `wm-tag`

---

## Coupling documentato

| Skill A | Skill B | Contratto |
|---|---|---|
| `wm-tag` | `wm-plan` | `wm-tag` invoca `wm-plan` con `tag-mode: true`. `wm-plan` è responsabile di overview, estimation e scrittura nel ticket. `wm-tag` gestisce tag, lista ticket e loop. |
| `wm-plan` | `wm-tag` | `caso-c` in Fase: ticket switcha su `wm-tag` cedendo il controllo. |

Da aggiungere nella tabella coupling di `CLAUDE.md`.

---

## Gestione multi-repo

Il dizionario `~/.config/webmapp/repos.json` è costruito una volta e aggiornato incrementalmente. Durante il `ticket-loop`, per ogni ticket `wm-plan` può ispezionare il codice del repo di destinazione (backend o frontend) consultando il dizionario per trovare la path corretta.

---

## File da creare/modificare

| File | Azione |
|---|---|
| `plugins/wm-skills/skills/wm-tag/SKILL.md` | Creare |
| `plugins/wm-skills/skills/wm-plan/SKILL.md` | Modificare (caso-c + Fase: estimation + tag-mode) |
| `CLAUDE.md` (marketplace) | Aggiornare tabella coupling e Feature disponibili |
