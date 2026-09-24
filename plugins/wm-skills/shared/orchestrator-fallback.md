# Ripiego: Orchestrator senza il server MCP

Queste istruzioni valgono **solo** quando i tool del server `orchestrator` non rispondono.
Nel funzionamento normale non vanno lette né seguite: usa i tool.

## Orchestrator API

Queste istruzioni valgono per tutte le chiamate HTTP a Orchestrator. Usale ogni volta che una fase richiede di leggere o scrivere un ticket.

### Configurazione

- **URL base:** leggi `$ORCHESTRATOR_URL` dall'environment. Se non è impostata usa `https://orchestrator.maphub.it` come default.
- **Auth:** salvato in `~/.config/webmapp/orchestrator-auth.json` (JSON con campi `token`, `id`, `name`, `email`). Se il file non esiste o la chiamata restituisce 401, esegui il login (vedi sotto). Se esiste solo il file legacy `~/.config/webmapp/orchestrator-token`, esegui la migrazione (vedi sotto).

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
mkdir -p ~/.config/webmapp
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

### Lettura ticket (GET — nessuna conferma richiesta)

```bash
ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-https://orchestrator.maphub.it}"
TOKEN=$(jq -r '.token' ~/.config/webmapp/orchestrator-auth.json)
curl -s -X GET "$ORCHESTRATOR_URL/api/stories/<ID>" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Accept: application/json"
```

Se risponde 401: cancella il file auth e ripeti il login prima di ritentare.

### Creazione ticket (POST — richiede conferma esplicita)

Prima di eseguire, mostra un riepilogo tabellare e chiedi conferma esplicita:

> **Creazione ticket**
>
> | Campo | Valore |
> |-------|--------|
> | `<campo>` | `<valore>` |
>
> Procedo?

Solo dopo la conferma:

```bash
ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-https://orchestrator.maphub.it}"
TOKEN=$(jq -r '.token' ~/.config/webmapp/orchestrator-auth.json)
curl -s -X POST "$ORCHESTRATOR_URL/api/stories" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '<json payload>'
```

Salva l'ID restituito (`$.id`) come `<ID>` del ticket per il resto del workflow.

### Aggiornamento ticket (PATCH — richiede conferma esplicita)

**Non usare questo ripiego per aggiungere informazione alla `description`** (note dev, esito di
una review): la PATCH sostituisce tutto il campo, e senza il server MCP l'unico modo di non
perdere il testo sarebbe ricopiarlo a mano. Fermati e chiedi al dev di far ripartire il server.

Prima di eseguire, mostra sempre un riepilogo tabellare:

> **Aggiornamento ticket oc:\<ID\>**
>
> | Campo | Valore |
> |-------|--------|
> | `<campo>` | `<valore>` |
> | `<campo>` | `<valore>` |
>
> Procedo?

Solo dopo la conferma:

```bash
ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-https://orchestrator.maphub.it}"
TOKEN=$(jq -r '.token' ~/.config/webmapp/orchestrator-auth.json)
curl -s -X PATCH "$ORCHESTRATOR_URL/api/stories/<ID>" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '<json payload>'
```

### Lista tag (ricerca — nessuna conferma richiesta)

```bash
ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-https://orchestrator.maphub.it}"
TOKEN=$(jq -r '.token' ~/.config/webmapp/orchestrator-auth.json)
curl -s "$ORCHESTRATOR_URL/api/tags?search=<query>" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Accept: application/json"
```

Attenzione: il parametro di filtro è `search`, non `name` — `?name=` viene ignorato dall'API e restituisce tutti i tag.

### Lettura tag (GET — nessuna conferma richiesta)

```bash
ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-https://orchestrator.maphub.it}"
TOKEN=$(jq -r '.token' ~/.config/webmapp/orchestrator-auth.json)
curl -s "$ORCHESTRATOR_URL/api/tags/<TAG_ID>" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Accept: application/json"
```

### Creazione tag (POST — richiede conferma esplicita)

Prima di eseguire, mostra un riepilogo tabellare e chiedi conferma esplicita:

> **Creazione tag**
>
> | Campo | Valore |
> |-------|--------|
> | `name` | `<nome>` |
> | `description` | `<prime 200 caratteri>...` |
>
> Procedo?

Solo dopo la conferma:

```bash
ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-https://orchestrator.maphub.it}"
TOKEN=$(jq -r '.token' ~/.config/webmapp/orchestrator-auth.json)
curl -s -X POST "$ORCHESTRATOR_URL/api/tags" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '{"name": "<nome>", "description": "<descrizione>"}'
```

Salva l'`id` restituito — serve per associare i ticket al tag. Ricorda: la `description` di un tag è in **Markdown**, a differenza della `description` di un ticket che è HTML.

### Aggiornamento tag (PATCH — richiede conferma esplicita)

Prima di eseguire, mostra sempre un riepilogo tabellare:

> **Aggiornamento tag \<TAG_ID\>**
>
> | Campo | Valore |
> |-------|--------|
> | `<campo>` | `<valore>` |
>
> Procedo?

Solo dopo la conferma:

```bash
ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-https://orchestrator.maphub.it}"
TOKEN=$(jq -r '.token' ~/.config/webmapp/orchestrator-auth.json)
curl -s -X PATCH "$ORCHESTRATOR_URL/api/tags/<TAG_ID>" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '<json payload>'
```

### Associare un ticket a un tag (POST — richiede conferma esplicita)

Prima di eseguire, mostra un riepilogo e chiedi conferma esplicita:

> **Associazione ticket → tag**
>
> Il ticket oc:\<ID\> verrebbe associato al tag \<TAG_ID\>.
>
> Procedo?

Solo dopo la conferma:

```bash
ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-https://orchestrator.maphub.it}"
TOKEN=$(jq -r '.token' ~/.config/webmapp/orchestrator-auth.json)
curl -s -X POST "$ORCHESTRATOR_URL/api/tags/<TAG_ID>/stories/<ID>" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Accept: application/json"
```

Non usare mai il campo `tags` di una story per associare un tag: sostituisce l'intero elenco e cancella i tag già presenti.

### Togliere un ticket da un tag (DELETE — richiede conferma esplicita)

Prima di eseguire, mostra un riepilogo e chiedi conferma esplicita:

> **Rimozione ticket ← tag**
>
> Il ticket oc:\<ID\> verrebbe tolto dal tag \<TAG_ID\>.
>
> Procedo?

Solo dopo la conferma:

```bash
ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-https://orchestrator.maphub.it}"
TOKEN=$(jq -r '.token' ~/.config/webmapp/orchestrator-auth.json)
curl -s -X DELETE "$ORCHESTRATOR_URL/api/tags/<TAG_ID>/stories/<ID>" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Accept: application/json"
```

### Status disponibili (letti dinamicamente)

Prima di proporre uno status, leggi i valori aggiornati da:

```
https://raw.githubusercontent.com/webmappsrl/orchestrator/main/app/Enums/StoryStatus.php
```

Estrai i valori dell'enum e presentali all'utente come lista numerata. Suggerisci quello più appropriato al contesto ma aspetta sempre la scelta esplicita dell'utente.

### Tipi disponibili (letti dinamicamente)

Prima di proporre un tipo di ticket o di costruire un payload che contiene il campo `type`, leggi i valori aggiornati da:

```
https://raw.githubusercontent.com/webmappsrl/orchestrator/main/app/Enums/StoryType.php
```

Estrai i valori dell'enum e usa **solo** quelli. Attenzione: il valore da inviare è il **valore stringa** del case, non il nome del case (i due possono differire). Non dedurre un tipo dal senso comune né riusare un valore visto altrove: un `type` non presente nell'enum viene rifiutato dall'API con `422 — Il valore selezionato per type non è valido`.

### Campi accettati (letti dinamicamente)

Prima di costruire un payload POST o PATCH, leggi i campi validati da:

```
https://raw.githubusercontent.com/webmappsrl/orchestrator/main/app/Http/Requests/Api/StoryApiRequest.php
```

Usa solo i campi presenti nelle `rules()` del Form Request. Non inviare campi non dichiarati.

### Regola generale scritture

**Qualsiasi operazione di scrittura su Orchestrator (POST o PATCH) richiede conferma esplicita con preview della modifica prima di eseguire la chiamata HTTP. Nessuna eccezione.**
