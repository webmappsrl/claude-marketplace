---
name: wm-tag
description: "Usa quando hai una trascrizione Meet, un brief cliente o qualsiasi materiale che descrive richieste da trasformare in ticket Orchestrator raggruppati in un tag. Analizza il materiale, crea il tag con le macro aree, e per ogni task identificato esegue il flusso wm-plan in modalità tag-mode (overview → description del ticket, poi stop)."
---

# wm-tag — Trascrizione cliente → Tag + Ticket Orchestrator

Questa skill trasforma un brief o una trascrizione cliente in un tag Orchestrator con ticket figli strutturati e stimati. Per ogni ticket invoca `wm-skills:wm-plan` in tag-mode, che esegue il flusso completo di analisi (fino all'overview e alla stima) senza toccare il filesystem del repo.

---

## Orchestrator API

Segui le stesse regole di `wm-skills:wm-plan` per auth, login e migrazione da file legacy. URL base: `${ORCHESTRATOR_URL:-https://orchestrator.maphub.it}`. Auth: `~/.config/webmapp/orchestrator-auth.json`.

**Regola generale scritture:** qualsiasi POST o PATCH — tag o story — richiede sempre un preview tabellare dei campi e conferma esplicita dell'utente prima di eseguire la chiamata HTTP. Nessuna eccezione.

### API Tag

**Lista tag (ricerca):**
```bash
ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-https://orchestrator.maphub.it}"
TOKEN=$(jq -r '.token' ~/.config/webmapp/orchestrator-auth.json)
curl -s "$ORCHESTRATOR_URL/api/tags?search=<query>" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Accept: application/json"
```

**Creazione tag:**
```bash
curl -s -X POST "$ORCHESTRATOR_URL/api/tags" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '{"name": "<nome>", "description": "<descrizione>"}'
```

Salva l'`id` restituito — serve per associare i ticket al tag.

---

## Fase: input

Accetta due formati:

- **Testo in chat** — l'utente incolla direttamente la trascrizione o il brief. Fonte: `"Testo fornito in chat"`.
- **Link Google Drive** — usa `mcp__claude_ai_Google_Drive__read_file_content` se disponibile, altrimenti WebFetch sull'URL. Fonte: l'URL completo del documento.

Se l'utente non ha ancora fornito il materiale, chiedi: "Incolla la trascrizione o fornisci il link Google Drive del documento."

Tieni traccia della fonte — verrà linkata in testa alla descrizione del tag.

---

## Fase: repo-map

Prima di qualsiasi analisi, costruisce o aggiorna `~/.config/webmapp/repos.json`.

```bash
# Individua la cartella padre del working directory
PARENT=$(dirname "$PWD")

# Trova tutti i repo git nella cartella padre
find "$PARENT" -maxdepth 2 -name ".git" -type d | sed 's|/.git||'
```

**Algoritmo:**
1. Leggi `~/.config/webmapp/repos.json` se esiste (struttura: `{"nome-repo": "/path/assoluta"}`)
2. Per ogni repo trovato non presente nel dizionario, aggiungi `"basename-cartella": "/path/assoluta"`
3. Salva il file aggiornato — usa jq per aggiornare incrementalmente, mai sovrascrivere da zero:
   ```bash
   # Se il file non esiste, inizializza con oggetto vuoto
   [ -f ~/.config/webmapp/repos.json ] || echo '{}' > ~/.config/webmapp/repos.json

   # Aggiunta incrementale per ogni nuovo repo trovato
   jq --arg name "<basename>" --arg path "<path>" '. + {($name): $path}' \
     ~/.config/webmapp/repos.json > /tmp/repos_tmp.json && \
     mv /tmp/repos_tmp.json ~/.config/webmapp/repos.json
   ```
4. Mostra all'utente la mappa completa:
   > **Repository disponibili:**
   > | Nome | Path |
   > |------|------|
   > | `<nome>` | `<path>` |

Questa mappa è disponibile per tutto il flusso. Quando serve ispezionare codice in un repo specifico, leggi la path da `repos.json` e naviga lì.

---

## Fase: client-extraction

Analizza il testo e identifica il nome del cliente.

- Cerca ricorrenze di nomi propri, ragioni sociali, acronimi usati come riferimento al committente
- Proponi il nome trovato con motivazione:
  > "Ho trovato il nome cliente: **CAMMINI** — citato 6 volte nel testo come destinatario delle richieste. È corretto?"
- Attendi conferma esplicita. Se l'utente corregge, usa il nome corretto per tutto il resto del flusso.
- Il nome cliente viene usato in MAIUSCOLO nel naming del tag (es. `CAMMINI`, `PARCHI`, `REGIONE-VENETO`).

---

## Fase: tag-naming

Costruisce il nome del tag seguendo la convenzione: `[RDO][CLIENTE][ANNO]<N>`

Esempio: `[RDO][CAMMINI][2026]1`

**Calcolo di N:**

```bash
ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-https://orchestrator.maphub.it}"
TOKEN=$(jq -r '.token' ~/.config/webmapp/orchestrator-auth.json)
ANNO=$(date +%Y)

# Cerca tag esistenti per questo cliente e anno
curl -s "$ORCHESTRATOR_URL/api/tags?search=RDO" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Accept: application/json" | \
  jq --arg cliente "<CLIENTE>" --arg anno "$ANNO" \
  '[.[] | select(.name | contains($cliente) and contains($anno))] | length'
```

Il valore restituito è il numero di tag esistenti per quel cliente+anno. `N = conteggio + 1`.

Proponi il nome al dev e attendi conferma:
> "Nome tag proposto: `[RDO][CAMMINI][2026]2` (esistono già 1 tag per CAMMINI nel 2026). Confermo?"

---

## Fase: tag-description

Analizza il testo del brief/trascrizione e produce la descrizione del tag in Markdown. **Non essere scarno: ogni macro area deve contenere abbastanza contesto da capire cosa vuole il cliente senza dover rileggere la trascrizione.**

Il campo `description` dei tag è renderizzato da un editor Markdown (Toast UI, non un WYSIWYG HTML): resta in Markdown, a differenza del campo `description` dei ticket (`Story`), che invece è HTML — non convertire.

```markdown
**Fonte:** <URL Drive o "Testo fornito in chat">

## Contesto

<Chi è il cliente, qual è il progetto, qual è l'obiettivo dichiarato di questa sessione/call. 3-5 righe. Includi eventuali vincoli espliciti (tecnologici, di budget, di timeline) e il tono della richiesta (urgente, esplorativo, consolidamento, ecc.)>

## Macro aree

### <Area 1 — titolo descrittivo>
<Descrizione estesa di cosa vuole il cliente in quest'area: motivazione, comportamento atteso, eventuali esempi o casi d'uso citati nella trascrizione. Minimo 3 righe per area.>

### <Area 2 — titolo descrittivo>
<Descrizione estesa...>

...

## Note e vincoli trasversali

<Tutto ciò che non rientra in una singola area ma vale per il progetto: preferenze tecnologiche, limitazioni dichiarate, richieste di compatibilità, aspettative su tempi di risposta o deploy, stakeholder citati, dipendenze da sistemi esterni.>
```

Mostra la descrizione all'utente e attendi approvazione esplicita prima di procedere alla creazione del tag. Se l'utente chiede di espandere o correggere una sezione, aggiornala e mostra di nuovo prima di procedere.

---

## Fase: tag-creation

Mostra il preview tabellare e chiedi conferma:

> **Creazione tag**
>
> | Campo | Valore |
> |-------|--------|
> | `name` | `[RDO][CLIENTE][ANNO]N` |
> | `description` | `<prime 200 caratteri della descrizione>...` |
>
> Procedo?

Solo dopo la conferma, esegui il POST:

```bash
ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-https://orchestrator.maphub.it}"
TOKEN=$(jq -r '.token' ~/.config/webmapp/orchestrator-auth.json)
curl -s -X POST "$ORCHESTRATOR_URL/api/tags" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d "{\"name\": \"<nome-tag>\", \"description\": \"<descrizione-escaped>\"}"
```

Salva l'`id` restituito come `<TAG_ID>` per l'associazione dei ticket.
Al termine: `✅ Tag \`<nome-tag>\` creato (ID: <TAG_ID>).`

---

## Fase: ticket-list

Analizza il testo e individua tutte le richieste distinte del cliente. Questa fase si articola in due step che vanno presentati separatamente.

### ticket-list: step 1 — elenco richieste individuate

Prima di proporre i ticket, mostra all'utente **tutte le richieste identificate nel testo** in forma estesa. Ogni richiesta deve essere descritta con abbastanza dettaglio da capire cosa ha chiesto il cliente, non solo il titolo.

Formato:

---
**Richieste individuate dalla trascrizione**

**1. \<Titolo descrittivo della richiesta\>**
\<Descrizione estesa: cosa ha chiesto il cliente, perché lo vuole, eventuali dettagli tecnici o esempi citati, comportamento atteso. Minimo 3 righe.\>
*Citazione dalla trascrizione (se presente):* "\<frase esatta o parafrasi vicina al testo originale\>"

**2. \<Titolo descrittivo\>**
\<Descrizione estesa...\>
*Citazione:* "..."

...
---

Chiedi all'utente: "Ho individuato \<N\> richieste. Le ho capite correttamente? Ci sono richieste mancanti, da unire o da scartare?"

Attendi feedback esplicito prima di procedere allo step 2.

### ticket-list: step 2 — mapping a ticket

Una volta approvato l'elenco richieste, proponi come ogni richiesta si traduce in ticket Orchestrator:

| # | Richiesta | Titolo ticket proposto | Tipo | Repo |
|---|---|---|---|---|
| 1 | \<titolo richiesta\> | \<titolo sintetico per il ticket\> | Feature | backend |
| 2 | \<titolo richiesta\> | \<titolo sintetico\> | Bug | frontend |

Note sul mapping:
- Una richiesta può generare più ticket se copre domini distinti (es. backend + frontend separati)
- Richieste correlate possono essere unite in un unico ticket se sono inseparabili tecnicamente
- Il titolo del ticket deve essere in italiano, sintetico (max 8 parole), comprensibile senza contesto

L'utente può:
- **Approvare** il mapping così com'è
- **Modificare** titolo, tipo o repo di un ticket
- **Dividere** una richiesta in più ticket
- **Unire** più richieste in un unico ticket
- **Aggiungere** un ticket non mappato

Attendi approvazione esplicita del mapping prima di procedere al loop.

---

## Fase: ticket-loop

Per ogni ticket nella lista approvata, nell'ordine:

1. Annuncia: "Processo ticket \<N\>/\<TOT\>: **\<titolo\>**"
2. Invoca `wm-skills:wm-plan` passando questo contesto nel tuo messaggio di invocazione:
   - Titolo del ticket
   - Tipo (uno dei valori validi dell'enum `StoryType` — leggilo come descritto in `wm-skills:wm-plan` → `## Orchestrator API → Tipi disponibili`)
   - Repo di destinazione (path da `repos.json`)
   - ID tag padre (`<TAG_ID>`)
   - Flag `tag-mode: true`
3. `wm-plan` esegue il flusso completo in tag-mode (reverse-interaction, overview, challenge, estimation se Feature) e scrive l'overview nella description del ticket Orchestrator associandolo al tag
4. Al termine di ogni ticket, chiedi:
   > "Ticket \<N\>/\<TOT\> completato. Procedo con il prossimo (**\<titolo-prossimo\>**), o vuoi fermarti qui?"
5. Se l'utente vuole fermarsi, interrompi il loop. I ticket rimanenti restano in lista e possono essere ripresi in una sessione successiva rilanciando `wm-skills:wm-tag` e fornendo lo stesso TAG_ID come contesto.
