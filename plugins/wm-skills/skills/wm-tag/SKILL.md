---
name: wm-tag
description: "Usa quando hai una trascrizione Meet, un brief cliente o qualsiasi materiale che descrive richieste da trasformare in ticket Orchestrator raggruppati in un tag. Analizza il materiale, crea il tag con le macro aree (schema Cosa/Come/Esiste, fonte citata), poi vaglia i candidati uno per volta: solo quelli su cui il dev dice sì diventano ticket via wm-plan in tag-mode, gli altri restano nel tag come situazioni aperte o in standby."
---

# wm-tag — Trascrizione cliente → Tag + Ticket Orchestrator

Questa skill trasforma un brief o una trascrizione cliente in un tag Orchestrator con ticket figli strutturati e stimati. Per ogni ticket invoca `wm-skills:wm-plan` in tag-mode, che esegue il flusso completo di analisi (fino all'overview e alla stima) senza toccare il filesystem del repo.

---

## Lingua: mai modi di dire inglesi tradotti

**Scrivi in italiano corrente. Non tradurre alla lettera un'espressione idiomatica inglese**: chi
legge è uno sviluppatore italiano che quei modi di dire non li conosce, e una traduzione parola per
parola non si capisce — «alzare il pavimento» per *raise the floor*, «strato sottile» per *thin
layer*, «a colpo d'occhio» per *at a glance*, «il raggio di esplosione» per *blast radius*. Se non
diresti quella frase parlando con un collega, non scriverla.

I **termini tecnici** restano invece in inglese e non si traducono: commit, branch, merge, build,
deploy, review, gate, tool, check. Tradurli è l'errore opposto e rende il testo altrettanto
illeggibile.

Nel dubbio: di' la cosa in modo esplicito, anche se è più lungo. «Non ha migliorato il risultato
peggiore» si capisce; «non ha alzato il pavimento» no.

## Orchestrator API

Le operazioni su Orchestrator si fanno con i tool del server `orchestrator`: vedi `wm-skills:wm-plan` → `## Orchestrator`. La regola dell'anteprima prima della scrittura vale identica qui.

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

**Il nome del tag lo decide il dev.** La convenzione qui sotto è un default da proporre, non una
regola da imporre: se il dev ha in mente un altro nome, si usa il suo senza discutere.

Convenzione di default: `[RDO][CLIENTE][ANNO]<N>` — es. `[RDO][CAMMINI][2026]1`.

**Calcolo di N:**

Chiama `list_tags` con `search: "RDO"`, poi conta tra i risultati quanti hanno nel nome sia `<CLIENTE>` sia l'anno corrente (`date +%Y`). Il conteggio è il numero di tag esistenti per quel cliente+anno. `N = conteggio + 1`.

Proponi il nome al dev e attendi conferma:
> "Nome tag proposto: `[RDO][CAMMINI][2026]2` (esistono già 1 tag per CAMMINI nel 2026). Va bene, o preferisci un altro nome?"

Se il dev fornisce un nome proprio ma lascia aperta una parte (tipicamente il postfisso), proponi
tu le alternative per quella parte sola e lascia scegliere.

---

## Fase: tag-description

Analizza il testo del brief/trascrizione e produce la descrizione del tag in Markdown. **Non essere scarno: ogni macro area deve contenere abbastanza contesto da capire cosa vuole il cliente senza dover rileggere la trascrizione.**

Il campo `description` dei tag è renderizzato da un editor Markdown (Toast UI, non un WYSIWYG HTML): resta in Markdown, a differenza del campo `description` dei ticket (`Story`), che invece è HTML — non convertire.

### tag-description: lo schema delle macro aree

**Ogni macro area si scrive sempre con questo schema a tre voci, senza eccezioni.** Una macro area
scritta come prosa libera non è accettabile: chi legge deve capire in tre righe se c'è già qualcosa
di pronto o se si parte da zero.

- **Cosa** — cosa vuole il cliente, con **la fonte citata**: la frase della trascrizione fra
  virgolette, oppure il riferimento puntuale al documento (riga e colonna di un foglio, numero di
  campo di un format). Senza fonte il punto non si scrive: se non si trova, va fra le *Situazioni
  aperte*.
- **Come** — come si realizzerebbe sulla nostra piattaforma: modelli, Resource, service, endpoint
  coinvolti. È un'ipotesi di lavoro, non una specifica.
- **Esiste** — cosa c'è già nel codice, verificato aprendo i file (delega a
  `wm-skills:wm-codebase-research`), e cosa manca. Se non esiste nulla, si scrive che non esiste.

### tag-description: struttura

```markdown
**Fonte:** <URL Drive o "Testo fornito in chat">

## Contesto

<Chi è il cliente, qual è il progetto, qual è l'obiettivo dichiarato di questa sessione/call. 3-5 righe. Includi eventuali vincoli espliciti (tecnologici, di budget, di timeline) e il tono della richiesta (urgente, esplorativo, consolidamento, ecc.)>

## Macro aree

### <Area 1 — titolo descrittivo>
**Cosa:** <cosa vuole il cliente> — *fonte:* "<citazione>" / <riferimento al documento>
**Come:** <ipotesi di realizzazione>
**Esiste:** <cosa c'è già nel codice, cosa manca>

### <Area 2 — titolo descrittivo>
...

## Situazioni aperte

<I punti su cui non si può procedere finché non arriva un responso da fuori: una risposta del
cliente, di un fornitore terzo, o una decisione interna ancora da prendere. Uno per riga, con
indicato **da chi** si aspetta la risposta. Questi non diventano ticket.>

## In standby

<I punti tecnicamente chiari ma bloccati da un prerequisito nostro — tipicamente un modello dati
non ancora chiuso. Uno per riga, con indicato **cosa** si sta aspettando. Nemmeno questi diventano
ticket, ma si riaprono da soli quando il prerequisito cade.>

## Ticket aperti da questo tag

| Ticket | Titolo | Area |
|---|---|---|
| oc:<ID> | <titolo> | <macro area di riferimento> |

## Note e vincoli trasversali

<Tutto ciò che non rientra in una singola area ma vale per il progetto: preferenze tecnologiche, limitazioni dichiarate, richieste di compatibilità, aspettative su tempi di risposta o deploy, stakeholder citati, dipendenze da sistemi esterni.>
```

Le sezioni *Situazioni aperte*, *In standby* e *Ticket aperti da questo tag* nascono vuote: si
riempiono durante il vaglio (`Fase: candidate-review`) e si scrivono sul tag una volta sola, in
`Fase: tag-update`.

Mostra la descrizione all'utente **per intero** — mai un riassunto, mai il solo conteggio dei
caratteri, mai «ho aggiornato la sezione X»: si ristampa il testo completo ad ogni revisione.
Attendi approvazione esplicita prima di procedere alla creazione del tag.

---

## Fase: tag-creation

Chiama `create_tag` con `name` e `description` senza `confirm`: mostra l'anteprima calcolata dal tool sui campi reali. Presentala all'utente e attendi conferma esplicita.

Solo dopo la conferma, richiama `create_tag` con gli stessi campi e `confirm: true`.

Salva l'`id` restituito come `<TAG_ID>` per l'associazione dei ticket.
Al termine: `✅ Tag \`<nome-tag>\` creato (ID: <TAG_ID>).`

Poi **fermati e chiedi**:

> "Tag creato. Vuoi procedere ora con i ticket, o ti fermi qui?"

Il tag ha valore da solo: è la memoria della call. Aprire i ticket è una seconda decisione, e va
presa dal dev — non è il seguito automatico della prima. Se risponde di no, il flusso finisce qui e
si riprende in una sessione successiva passando il `<TAG_ID>`.

---

## Fase: candidate-list

Analizza il testo e individua tutte le richieste distinte del cliente. Mostra all'utente **tutte le
richieste identificate** in forma estesa. Ogni richiesta deve essere descritta con abbastanza
dettaglio da capire cosa ha chiesto il cliente, non solo il titolo.

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

Attendi feedback esplicito. L'elenco approvato è la lista dei **candidati**, non la lista dei
ticket: quanti di questi diventino un ticket si decide uno per uno nella fase successiva.

---

## Fase: candidate-review

**Niente tabella di mapping approvata in blocco.** I candidati si vagliano **uno per volta**: il dev
non ha in testa il tuo elenco, e approvare dieci righe insieme significa approvarle senza leggerle.

Per ogni candidato, nell'ordine:

1. **Presenta il candidato per intero.** Deve essere **autoconsistente**: contiene tutto ciò che lo
   riguarda, come se il dev non avesse letto nient'altro. Nessun rimando a testo prodotto prima
   nella conversazione — mai «quello del custode», «come dicevamo sopra», «il secondo candidato».
   Ogni sigla, etichetta o nome di stato va spiegato alla prima occorrenza, anche se l'hai già
   spiegato dieci messaggi fa.

   Il candidato si presenta con lo stesso schema delle macro aree — **Cosa** (con la fonte citata),
   **Come**, **Esiste** — più il titolo proposto per il ticket, il tipo e il repo di destinazione.

2. **Chiedi, e fermati.** Una domanda secca: si fa o no? Non proporre il candidato successivo
   nello stesso messaggio e non anticipare quanti ne restano nel merito: **un elemento per volta**.

3. **Registra l'esito.** Sono tre, non due:

   | Esito | Cosa succede |
   |---|---|
   | **Sì** | Si crea il ticket: prosegui con `Fase: ticket-loop` per questo solo candidato. |
   | **No, si aspetta un responso esterno** | Niente ticket. Va nelle *Situazioni aperte* della descrizione del tag, con indicato da chi si aspetta la risposta. |
   | **No, manca un prerequisito nostro** | Niente ticket. Va in *In standby*, con indicato cosa si sta aspettando. |

   Un «no» secco senza motivo è semplicemente un candidato scartato: non finisce da nessuna parte.

<HARD-GATE>
**Nessun ticket si crea senza un «sì» esplicito del dev su quel singolo candidato.** Non vale
l'approvazione dell'elenco dei candidati, non vale un «vai avanti» generico, non vale il fatto che
il candidato sia ovvio. Se il sì non c'è, il ticket non si apre.
</HARD-GATE>

Le modifiche alla descrizione del tag che maturano qui (situazioni aperte, standby, rimandi ai
ticket creati) **non si scrivono subito su Orchestrator**: si parcheggiano in un file locale nella
scratchpad della sessione, e si applicano in un colpo solo in `Fase: tag-update`.

---

## Fase: ticket-loop

Si entra qui **solo** dopo un «sì» su un singolo candidato, e si esce dopo quel ticket per tornare
al candidato successivo di `Fase: candidate-review`.

1. Annuncia: "Processo il ticket: **\<titolo\>**"
2. Invoca `wm-skills:wm-plan` passando questo contesto nel tuo messaggio di invocazione:
   - Titolo del ticket
   - Tipo (uno dei valori validi dell'enum `StoryType` — leggili dallo schema del tool `create_story`/`update_story`, vedi `wm-skills:wm-plan` → `## Orchestrator`)
   - Repo di destinazione (path da `repos.json`)
   - ID tag padre (`<TAG_ID>`)
   - Flag `tag-mode: true`
3. `wm-plan` esegue il flusso completo in tag-mode (reverse-interaction, overview, challenge, estimation se Feature) e scrive l'overview nella description del ticket Orchestrator associandolo al tag
4. Annota nel file parcheggiato la riga da aggiungere alla tabella *Ticket aperti da questo tag*:
   `oc:<ID>`, titolo, macro area di riferimento.
5. Torna a `Fase: candidate-review` con il candidato successivo.

---

## Fase: tag-update

Quando i candidati sono finiti — o quando il dev decide di fermarsi — applica in un'unica chiamata
tutte le modifiche parcheggiate alla descrizione del tag: *Situazioni aperte*, *In standby*, la
tabella *Ticket aperti da questo tag*, e ogni precisazione emersa durante il vaglio.

Mostra al dev la descrizione risultante **per intero** prima di scrivere, poi chiama `update_tag`
senza `confirm` per l'anteprima del tool, e solo dopo l'approvazione esplicita con `confirm: true`.

Una sola scrittura a fine sessione, non una per candidato: durante il vaglio i rimandi ai ticket non
sono ancora noti, e riscrivere la descrizione a ogni passo produce versioni intermedie incoerenti.

I candidati non ancora vagliati restano in lista e si riprendono in una sessione successiva
rilanciando `wm-skills:wm-tag` e fornendo lo stesso `<TAG_ID>` come contesto.
