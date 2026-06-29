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

Analizza il testo del brief/trascrizione e produce la descrizione del tag in Markdown:

```markdown
**Fonte:** <URL Drive o "Testo fornito in chat">

## Macro aree

- **<Area 1>**: <descrizione sintetica — cosa vuole il cliente in quest'area>
- **<Area 2>**: <descrizione sintetica>
...

## Informazioni principali

<Contesto generale: chi è il cliente, obiettivo del progetto, vincoli noti (tecnologici, temporali, di budget), deadline esplicite se presenti nel testo>
```

Mostra la descrizione all'utente e attendi approvazione esplicita prima di procedere alla creazione del tag.

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

Analizza il testo e individua tutti i task distinti che il cliente ha richiesto. Per ogni task identifica:
- Titolo sintetico
- Tipo: Feature (nuova funzionalità), Bug (problema da risolvere), Task (attività tecnica senza utente finale diretto)
- Repo di destinazione stimato (backend / frontend / altro — usando i nomi da `repos.json` come riferimento)

Presenta la lista proposta in formato tabellare:

| # | Titolo ticket | Tipo | Repo |
|---|---|---|---|
| 1 | ... | Feature | backend |
| 2 | ... | Bug | frontend |

L'utente può:
- **Approvare** la lista così com'è
- **Modificare** un titolo o tipo
- **Unire** due ticket in uno
- **Eliminare** un ticket dalla lista
- **Aggiungere** un ticket non rilevato

Attendi approvazione esplicita della lista prima di procedere al loop.

---

## Fase: ticket-loop

Per ogni ticket nella lista approvata, nell'ordine:

1. Annuncia: "Processo ticket \<N\>/\<TOT\>: **\<titolo\>**"
2. Invoca `wm-skills:wm-plan` passando questo contesto nel tuo messaggio di invocazione:
   - Titolo del ticket
   - Tipo (Feature / Bug / Task)
   - Repo di destinazione (path da `repos.json`)
   - ID tag padre (`<TAG_ID>`)
   - Flag `tag-mode: true`
3. `wm-plan` esegue il flusso completo in tag-mode (reverse-interaction, overview, challenge, estimation se Feature) e scrive l'overview nella description del ticket Orchestrator associandolo al tag
4. Al termine di ogni ticket, chiedi:
   > "Ticket \<N\>/\<TOT\> completato. Procedo con il prossimo (**\<titolo-prossimo\>**), o vuoi fermarti qui?"
5. Se l'utente vuole fermarsi, interrompi il loop. I ticket rimanenti restano in lista e possono essere ripresi in una sessione successiva rilanciando `wm-skills:wm-tag` e fornendo lo stesso TAG_ID come contesto.
