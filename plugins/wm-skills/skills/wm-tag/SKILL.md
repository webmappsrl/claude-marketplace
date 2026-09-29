---
name: wm-tag
description: "Usa quando hai una trascrizione Meet, un brief cliente o qualsiasi materiale che descrive richieste da trasformare in ticket Orchestrator raggruppati in un tag. Analizza il materiale, crea il tag con le macro aree (schema Cosa/Come/Esiste, fonte citata), poi vaglia i candidati uno per volta: solo quelli su cui il dev dice sì diventano ticket via wm-plan in tag-mode, gli altri restano nel tag come situazioni aperte o in standby. Usa anche per rilanciare su un tag già esistente, fornendo il suo ID: in quel caso non analizza materiale nuovo, ma verifica se i blocchi sono caduti — il codice è cambiato, il cliente ha risposto — e porta a ticket i punti che si sono sbloccati."
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

La skill si apre in due modi, e sono due lavori diversi.

**Materiale nuovo** — c'è una call da trasformare in tag. Accetta due formati:

- **Testo in chat** — l'utente incolla direttamente la trascrizione o il brief. Fonte: `"Testo fornito in chat"`.
- **Link Google Drive** — usa `mcp__claude_ai_Google_Drive__read_file_content` se disponibile, altrimenti WebFetch sull'URL. Fonte: l'URL completo del documento.

Se l'utente non ha ancora fornito il materiale, chiedi: "Incolla la trascrizione o fornisci il link Google Drive del documento."

Tieni traccia della fonte — verrà linkata in testa alla descrizione del tag.

Poi prosegui con `Fase: repo-map`.

**Tag esistente** — l'utente fornisce un `<TAG_ID>`, o nomina un tag già creato, senza portare
materiale nuovo. Non è la continuazione di un elenco lasciato a metà: è un **giro di verifica sui
blocchi**, e ha un flusso suo. Esegui `Fase: repo-map` — serve per andare a guardare il codice —
e poi salta direttamente a `Fase: tag-resume`, senza passare da `client-extraction`, `tag-naming`,
`tag-description` e `tag-creation`, che riguardano un tag che qui esiste già.

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
- **Esiste** — cosa c'è già nel codice, verificato aprendo i file, e cosa manca. Se non esiste
  nulla, si scrive che non esiste.

**La voce «Esiste» si delega a `wm-skills:wm-codebase-research`, con una chiamata sola per tutte
le macro aree.** Individua prima quali sono le aree, poi manda all'agente tutte le domande insieme:
una chiamata per area costa quante sono le aree e non aggiunge nulla, perché è sempre lo stesso
repo. Vale la stessa regola di `wm-plan: reverse-interaction`.

Verifica ogni prova che l'agente restituisce prima di scriverla nel tag:

```bash
sed -n '<riga-inizio>,<riga-fine>p' <file>
```

Se una prova non combacia con l'estratto del dossier, il dossier è inattendibile: rifai la ricerca
nel context principale e dillo al dev. Se l'agente risponde `Risposta: non determinabile dal repo`
o `RICERCA FALLITA`, scrivi nell'«Esiste» che la verifica non è stata fatta — mai un «non esiste»
al posto di una verifica mancata, perché è la voce su cui il dev deciderà se il lavoro parte da
zero.

### tag-description: struttura

```markdown
**Fonte:** <URL Drive o "Testo fornito in chat">
**Notebook:** <link del notebook NotebookLM del tag, se creato>

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

## Note e vincoli trasversali

<Tutto ciò che non rientra in una singola area ma vale per il progetto: preferenze tecnologiche, limitazioni dichiarate, richieste di compatibilità, aspettative su tempi di risposta o deploy, stakeholder citati, dipendenze da sistemi esterni.>
```

Le sezioni *Situazioni aperte* e *In standby* nascono vuote: si riempiono durante il vaglio
(`Fase: candidate-review`) e si scrivono sul tag una volta sola, in `Fase: tag-update`.

**I ticket aperti dal tag non si elencano nella descrizione.** L'associazione ticket↔tag è una
relazione vera su Orchestrator (`attach_story_to_tag`), e la pagina del tag mostra già i ticket
associati. Una tabella scritta a mano sarebbe una seconda copia dello stesso dato, che diverge
alla prima associazione fatta da un collega dall'interfaccia — la skill non la vedrebbe mai.
Nella descrizione va solo ciò che **non** è un ticket: le situazioni aperte e lo standby, che non
esistono da nessun'altra parte.

Mostra la descrizione all'utente **per intero** — mai un riassunto, mai il solo conteggio dei
caratteri, mai «ho aggiornato la sezione X»: si ristampa il testo completo ad ogni revisione.
Attendi approvazione esplicita prima di procedere alla creazione del tag.

---

## Fase: tag-creation

**Prima di creare il tag, verifica ogni citazione del «Cosa»** con la procedura di
`${CLAUDE_PLUGIN_ROOT}/shared/verifica-citazioni.md`, usando come fonte la trascrizione della riga
`**Fonte:**`. Una citazione che non si trova si corregge rileggendo il punto, oppure il punto va fra
le *Situazioni aperte*: un «Cosa» senza fonte non si scrive.

**Poi prepara il notebook del tag**: chiedi a `wm-skills:wm-transcript-research`, con
`tag: <nome del tag>` e l'id della trascrizione (o il testo incollato, se la fonte è «Testo fornito
in chat»), l'elenco degli argomenti della call. Serve a creare il notebook `tag <nome del tag>`
con le fonti caricate; il suo link va nella riga `**Notebook:**` della descrizione. Se l'agente
restituisce `RICERCA FALLITA`, crea il tag senza la riga `**Notebook:**` e dillo al dev.

Chiama `create_tag` con `name` e `description` senza `confirm`: mostra l'anteprima calcolata dal tool sui campi reali. Presentala all'utente e attendi conferma esplicita.

Solo dopo la conferma, richiama `create_tag` con gli stessi campi e `confirm: true`.

Salva l'`id` restituito come `<TAG_ID>` per l'associazione dei ticket.
Al termine: `✅ Tag \`<nome-tag>\` creato (ID: <TAG_ID>).`

Poi **fermati e chiedi**:

> "Tag creato. Vuoi procedere ora con i ticket, o ti fermi qui?"

Il tag ha valore da solo: è la memoria della call. Aprire i ticket è una seconda decisione, e va
presa dal dev — non è il seguito automatico della prima. Se risponde di no, il flusso finisce qui:
si riprende quando vuole rilanciando `wm-tag` con il `<TAG_ID>`, che entra da `Fase: tag-resume`.
Il lavoro non va perso, perché tutto ciò che serve a riprenderlo è nelle macro aree del tag.

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

## Fase: tag-resume

Si entra qui **solo** dal ramo «tag esistente» di `Fase: input`, e si esce verso
`Fase: candidate-review` con i candidati che nel frattempo si sono sbloccati.

**A cosa serve rilanciare `wm-tag` su un tag.** Non a riprendere una lista lasciata a metà, ma a
**sciogliere i blocchi**: le *Situazioni aperte* e l'*In standby* sono punti che non sono diventati
ticket perché mancava qualcosa, e col tempo quel qualcosa arriva — il codice cambia e il
prerequisito cade, oppure il cliente risponde. Ogni rilancio è un giro di verifica su quelle due
sezioni.

### tag-resume: stato attuale

Leggi il tag con `get_tag` e ricava dalla sua descrizione le macro aree, le *Situazioni aperte* e
l'*In standby*. **Non rianalizzare la trascrizione**: la descrizione del tag è la memoria della
call, ed è la fonte di questa fase.

**Un punto che ha già un ticket non è più un candidato.** I ticket nati da questo tag sono nella
relazione su Orchestrator, non in un elenco da mantenere: se un punto è diventato un ticket, il
suo posto ora è quel ticket, e `wm-tag` non lo tocca più — le informazioni nuove che lo riguardano
si portano lì, non qui.

Mostra al dev lo stato del tag prima di procedere: quante situazioni aperte, quante in standby, e
il testo di ciascuna per intero.

**Se il tag non ha né blocchi né ticket, non è mai stato vagliato** — è il caso del dev che si è
fermato subito dopo averlo creato. Non c'è niente da sbloccare: i candidati si ricavano **dalle
macro aree del tag**, che sono la mappa di cosa è emerso dalla call, e si va a
`Fase: candidate-review`. La trascrizione non serve: se il tag è scritto bene, ogni macro area ha
già il **Cosa** con la sua fonte citata, il **Come** e l'**Esiste**, che è esattamente ciò che
serve a presentare un candidato.

### tag-resume: verifica dei blocchi

Le due sezioni si sbloccano in modi diversi, e vanno trattate separatamente.

**In standby — il prerequisito è nostro, quindi si verifica nel codice.** Delega a
`wm-skills:wm-codebase-research`, **una chiamata sola con tutte le voci in standby**: una chiamata
per voce costa N volte e spezza il dialogo, come in `wm-plan: reverse-interaction`. Chiedi per
ciascuna se il prerequisito che la bloccava esiste ora nel repo, e pretendi la prova verbatim.
Verifica ogni prova prima di usarla:

```bash
sed -n '<riga-inizio>,<riga-fine>p' <file>
```

Se l'agente risponde `Risposta: non determinabile dal repo` o `RICERCA FALLITA`, la voce resta
in standby e lo dici al dev: non è una verifica fallita al posto di un «no», è una verifica che
non è stata fatta.

**Situazioni aperte — si aspetta un responso da fuori, e nel codice non c'è.** Non fingere di
poterlo verificare: chiedilo al dev, una voce per volta, ricordandogli **da chi** si aspettava la
risposta, così com'era scritto nel tag. Se la risposta è arrivata, il contenuto lo porta lui.

### tag-resume: esito

- **Blocco caduto** → il punto torna a essere un candidato. Lo aggiorni con quello che è emerso —
  il codice che ora esiste, la risposta del cliente — e lo porti in `Fase: candidate-review`, con
  lo stesso vaglio uno per volta e lo stesso HARD-GATE: la caduta del blocco non è un sì.
- **Blocco ancora in piedi** → resta dov'è. Se è cambiato *cosa* si aspetta, riscrivi la voce: una
  situazione aperta ferma da tre mesi con scritto un motivo superato è peggio che non averla.

Le modifiche si parcheggiano come sempre e si applicano in `Fase: tag-update`.

**Perché non si aprono i ticket subito, mettendoli in backlog.** Sarebbe l'alternativa ovvia: un
ticket per ogni punto bloccato, parcheggiato in backlog finché non si sblocca. È stata scartata: un
ticket in backlog con informazioni che nessuno aggiorna è peggio dell'assenza del ticket, perché
chi lo prende in mano fra due mesi lavora su una descrizione ferma al giorno in cui è stata
scritta, senza sapere che è ferma. Nel tag invece il punto sta accanto al resto della call, e si
rilegge insieme al suo contesto.

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

Le modifiche alla descrizione del tag che maturano qui (situazioni aperte, standby) **non si
scrivono subito su Orchestrator**: si parcheggiano e si applicano in un colpo solo in
`Fase: tag-update`.

Il parcheggio è un file con un nome fisso, nella scratchpad della sessione:

```
<scratchpad>/wm-tag-<TAG_ID>-descrizione.md
```

Contiene la descrizione del tag **per intero**, non le sole modifiche: è la versione che
`tag-update` scriverà, e tenerla completa evita di dover ricomporre a fine sessione un testo da
frammenti. Riscrivilo ad ogni esito registrato, così se la sessione si interrompe il lavoro fatto
fino a lì è su disco e non solo nella conversazione.

**Se un candidato richiede di tornare sulla call** — una frase ambigua, un «Come» di cui non si
ricorda la decisione — non rileggere la trascrizione: fai la domanda a
`wm-skills:wm-transcript-research` con `tag: <nome del tag>`, e verifica le citazioni che usi.

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
4. Torna a `Fase: candidate-review` con il candidato successivo.

Non annotare il ticket creato da nessuna parte: l'associazione al tag l'ha già fatta `wm-plan` con
`attach_story_to_tag`, ed è quella la registrazione: Orchestrator la mostra nella pagina del tag.

---

## Fase: tag-update

Quando i candidati sono finiti — o quando il dev decide di fermarsi — scrivi sul tag la descrizione
parcheggiata in `<scratchpad>/wm-tag-<TAG_ID>-descrizione.md`: *Situazioni aperte*, *In standby*, e
ogni precisazione emersa durante il vaglio.

**Un punto che è diventato un ticket esce dalla sezione in cui stava.** Se una voce in standby si è
sbloccata in `Fase: tag-resume` ed è diventata un ticket, va tolta da lì: lasciarla farebbe
ricomparire come bloccato, al prossimo rilancio, qualcosa che ormai è in lavorazione.

Mostra al dev la descrizione risultante **per intero** prima di scrivere, poi chiama `update_tag`
senza `confirm` per l'anteprima del tool, e solo dopo l'approvazione esplicita con `confirm: true`.

Una sola scrittura a fine sessione, non una per candidato: riscrivere la descrizione a ogni passo
produce versioni intermedie incoerenti, e costa al dev una preview e una conferma per ogni
candidato vagliato — in mezzo al vaglio, che è la fase in cui deve restare concentrato su altro.

I candidati non ancora vagliati si riprendono rilanciando `wm-skills:wm-tag` con il `<TAG_ID>`, da
`Fase: tag-resume`: si ritrovano dalle macro aree del tag, non dalla trascrizione, che a quel punto
può non essere più a portata di mano. È un motivo in più per scrivere bene le macro aree — sono
l'unica cosa che sopravvive alla sessione.
