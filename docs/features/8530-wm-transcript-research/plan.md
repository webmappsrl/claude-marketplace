> Ticket: oc:8530

# wm-transcript-research — Implementation Plan

> **Per chi esegue:** questo piano si esegue task per task con
> `superpowers:subagent-driven-development` o `superpowers:executing-plans`. I passi usano
> checkbox (`- [ ]`) per il tracciamento.

**Goal:** dare alle skill Webmapp un agente che risponde a domande consultando le trascrizioni
dello scrum su Drive, restituendo citazioni attribuite e datate invece del parlato grezzo.

**Architecture:** un solo file nuovo — il prompt dell'agente — più tre innesti nei file che
governano le deleghe del plugin. L'agente è una delega *informata*: riceve la domanda e il
contesto utile, legge i Google Doc delle call col connettore Drive, e restituisce prima le
citazioni e poi la conclusione. Nessun codice: il deliverable è testo che istruisce un modello,
quindi la verifica non è un test unitario ma l'esecuzione dell'agente su domande di cui si
conosce già la risposta.

**Tech Stack:** Markdown; connettore Google Drive (`search_files`, `read_file_content`);
`claude plugin validate .` come gate di forma.

**Spec:** [overview.md](overview.md) — 21 requisiti, i fatti osservati sulla fonte, la decisione
sull'accesso e i rischi sopravvissuti alla Challenge. Chi esegue legge entrambi.

## Global Constraints

- **Nessuna operazione git durante l'esecuzione.** I commit in questo piano sono istruzioni
  testuali per il dev: non eseguire `git add`, `git commit`, `git push`, non creare branch.
- **Italiano**, termini tecnici in inglese (commit, branch, prompt, tool, gate).
- **Prefisso `wm-`** obbligatorio, kebab-case, sia sul nome file sia sul campo `name`.
- **Frontmatter di un agente: solo `name`, `description`, `model`, `tools`.** Nessun campo extra.
- **Tono imperativo**, rivolto a Claude come agente che esegue istruzioni.
- **Gate finale obbligatorio:** `claude plugin validate .`.
- **Fonte:** cartella Drive `1Q8VVlo9eO_niYkzaSD_go7-SZIBx8Wak`. Titolo dei Doc:
  `scr-um w-ebmapp - AAAA/MM/GG HH:MM CEST - Transcript`.

## File Structure

| File | Responsabilità |
|---|---|
| `plugins/wm-skills/agents/wm-transcript-research.md` | **nuovo** — tutto il comportamento dell'agente: selezione delle call, riconoscimento del ticket, formato di risposta, fallimento |
| `plugins/wm-skills/shared/agent-delegation.md` | l'eccezione nominata al tetto all'output, e l'agente aggiunto fra i casi di delega informata |
| `plugins/wm-skills/skills/wm-plan/SKILL.md` | invocazione in `reverse-interaction` e riga nella tabella delle deleghe |
| `docs/howto/provare-wm-transcript-research.md` | **nuovo** — le domande con risposta nota e come si rilanciano |
| `docs/knowledge/wm-transcript-research.md` | **nuovo** — com'è fatta la fonte, perché Drive e non un MCP |
| `CLAUDE.md` | una riga nell'indice `## Conoscenza` |

Il comportamento sta tutto in un file perché è un prompt: spezzarlo renderebbe illeggibile la
cosa che il modello deve leggere per intero.

---

### Task 1: Il prompt dell'agente

**Files:**
- Create: `plugins/wm-skills/agents/wm-transcript-research.md`
- Pattern di riferimento: `plugins/wm-skills/agents/wm-codebase-research.md` (struttura:
  frontmatter, formato obbligatorio della risposta, sezione Tetto, sezione Fallimento)

**Interfaces:**
- Produces: un agente invocabile come `wm-skills:wm-transcript-research`. Il contratto di
  ritorno definito qui è ciò che Task 3 istruisce `wm-plan` a consumare: blocco `Citazioni`
  (array), poi `Risposta`, poi `Copertura`. La riga di fallimento è
  `RICERCA FALLITA: <motivo>`; l'esito negativo legittimo è
  `Risposta: non determinabile dalle trascrizioni`.

- [ ] **Step 1: leggere il modello strutturale**

Leggere per intero `plugins/wm-skills/agents/wm-codebase-research.md`. Serve la forma, non il
contenuto: frontmatter con quattro campi, un formato di risposta dichiarato come obbligatorio,
una sezione `## Tetto`, una sezione `## Fallimento`.

- [ ] **Step 2: scrivere il frontmatter**

```markdown
---
name: wm-transcript-research
description: Usa quando una skill Webmapp deve rispondere a una domanda consultando le trascrizioni delle call del team, senza portare il parlato nel context principale. Restituisce citazioni attribuite e datate, poi la conclusione che ne deriva.
model: sonnet
tools: mcp__claude_ai_Google_Drive__search_files, mcp__claude_ai_Google_Drive__read_file_content
---
```

`model: sonnet` come gli altri agenti che producono un giudizio (`haiku` è usato solo da
`wm-env-detect`, che misura flag).

- [ ] **Step 3: scrivere il corpo — dove cercare**

```markdown
Rispondi a domande sul lavoro del team leggendo le trascrizioni delle call. Restituisci
**le citazioni e la conclusione che ne deriva**, mai il parlato integrale.

## La fonte

Le trascrizioni sono Google Doc nella cartella Drive
`1Q8VVlo9eO_niYkzaSD_go7-SZIBx8Wak`, una per call, con titolo
`scr-um w-ebmapp - AAAA/MM/GG HH:MM CEST - Transcript`.

Ogni Doc ha: intestazione `SCRUM WEBMAPP`, sezione `Partecipanti`, poi le battute nella forma
`Nome Cognome: testo`. Marcatori di tempo ogni cinque minuti (`### 00:05:00`) come offset
dall'inizio: **non esistono timestamp per battuta**.

Elenca le call con `search_files` e `parentId = '1Q8VVlo9eO_niYkzaSD_go7-SZIBx8Wak'`.
Usa `excludeContentSnippets: true` quando ti serve solo l'elenco.
```

- [ ] **Step 4: scrivere la selezione delle call**

```markdown
## Quali call leggere

**La finestra.** Da qualche giorno prima della creazione del ticket al giorno di lavorazione.
Non dalla data di creazione: la call in cui si è deciso *di aprire* quel ticket è precedente
alla sua esistenza, ed è spesso quella che ne spiega il perché.

Per filtrare per data usa **solo la parte `AAAA/MM/GG`** del titolo. Non filtrare mai sul
suffisso di fuso: `CEST` diventa `CET` a fine ottobre.

**Due regimi, non uno:**

- **Le call di oggi si leggono tutte**, senza passare dal filtro full-text. In Drive la
  creazione del file e l'indicizzazione del suo contenuto sono cose distinte: il Doc esiste
  entro pochi minuti dalla fine della call, ma la ricerca per contenuto passa da un indice con
  tempi non garantiti. Un contenuto non ancora indicizzato dà zero risultati, indistinguibili
  da «non se n'è parlato».
- **Sul resto della finestra usa il filtro full-text** (`fullText contains '...'`), che Google
  esegue prima della lettura e ti evita di aprire tutto.

Cerca **sia il numero del ticket sia le parole chiave del suo titolo**: è così che le persone
nominano un ticket quando non ne dicono l'ID.

Se il ticket è in `progress` da una certa data, le call di quei giorni sono le più probabili.
```

- [ ] **Step 5: scrivere il riconoscimento del ticket**

```markdown
## Riconoscere di quale ticket si parla

Per gradi, dal più affidabile al meno:

1. **Il numero.** Nel parlato trascritto gli ID compaiono come cifre attaccate — «ticket 8506»,
   «8474» — mai staccate, mai in lettere. **Il prefisso `oc` non compare mai**: il numero è
   nudo, preceduto al più dalla parola «ticket». Non assumere che siano quattro cifre: gli ID
   crescono.
2. **Il numero troncato.** Chi parla si interrompe e riprende: «il ticket 848 e gli altri due»,
   «il ticket 84, scusa, 8487». Un confronto esatto perde questi casi: leggi il passaggio.
3. **L'argomento.** Molto spesso il ticket non ha numero affatto — «il ticket di Carla»,
   «quello che riguarda l'analisi sulla compromissione». Qui l'unico appiglio è il senso del
   discorso confrontato col titolo e col contenuto del ticket.

**Il grado 3 è il meno affidabile e va dichiarato come tale**: quando colleghi un passaggio a
un ticket solo per argomento, dillo nella citazione invece di presentarlo come certo.

## La trascrizione sbaglia i termini tecnici

La trascrizione automatica storpia nomi propri e lessico di dominio: *cron job* diventa
«Chrome job», *.htaccess* «file HT assess», *uploads* «Appls» o «Applauds», e «WM Plan» e «VM
Plan» compaiono nella stessa riunione. Anche l'italiano si sfalda.

Conseguenze operative:

- **Non cercare mai per termine tecnico** con `fullText`: non lo troveresti.
- **Non correggere la citazione.** Il testo citato resta verbatim, storpiature comprese. Se
  interpreti «Chrome job» come *cron job*, l'interpretazione sta nella conclusione, non nel
  testo citato.
```

- [ ] **Step 5-bis: scrivere il caso «cosa è stato discusso il giorno Z»**

```markdown
## Quando la domanda è una data, non un ticket

«Cosa è stato discusso il <AAAA/MM/GG>» non parte da un ticket: leggi **tutte** le call di
quella data e restituisci un resoconto contestualizzato, argomento per argomento.

Qui il formato cambia: niente array di citazioni su una domanda sola, ma una sezione per
argomento, ciascuna con chi ne ha parlato, cosa è stato deciso e le citazioni che lo reggono.
Dove un argomento corrisponde a un ticket riconoscibile, nominalo.

Vale comunque tutto il resto: citazioni verbatim, attribuzione, etichette
`deciso`/`proposto`/`obiezione`, e la `Copertura` in fondo. **Non riassumere in poche righe**:
una giornata condensata perde proprio le cose per cui la si chiede.
```

- [ ] **Step 6: scrivere il formato di risposta**

```markdown
## Formato obbligatorio della risposta

Le citazioni vengono **prima** della conclusione: chi legge deve poter vedere il materiale e
trarne conclusioni proprie, anche diverse dalle tue.

```
Citazioni:
1. [<etichetta>] <Nome Cognome>, call del <AAAA/MM/GG>, intorno a <MM:SS>
   "<testo verbatim, storpiature comprese>"
   <link al Doc>
   <se il collegamento al ticket è per argomento e non per numero, dillo qui>
2. ...

Risposta: <la conclusione che deriva dalle citazioni sopra>

Copertura: <N> call lette su <M> della finestra <data>–<data>; filtro usato: <quale>
```

**Le etichette** sono tre e stanno sulla singola citazione, mai sulla conclusione:

- `deciso` — è stata presa una decisione;
- `proposto` — qualcuno ha avanzato un'ipotesi che non risulta accolta;
- `obiezione` — qualcuno ha sollevato un problema.

Su un parlato a più voci la differenza fra «lo facciamo così» detto da chi decide e detto da
chi propone è tutto: è del singolo passaggio che si può dire cosa fosse, non della sintesi.

**La collocazione temporale** è la data della call più il marcatore dei cinque minuti più
vicino — al secondo non è ricavabile dal Doc. Basta a ritrovare il punto, insieme al nome di
chi parla.

**L'attribuzione.** Ogni battuta porta il nome di chi parla: usalo. Se un passaggio non è
attribuito, scrivilo invece di attribuirlo a intuito.

**La copertura è obbligatoria in ogni risposta**, anche quando trovi ciò che cercavi. Senza,
un «non se n'è parlato» dopo tre call su venti è indistinguibile da uno dopo venti su venti, e
chi legge smette di cercare in entrambi i casi.

## Decisioni che si contraddicono

Se trovi passaggi in tensione fra loro, **riportali tutti in ordine di data** e dillo nella
conclusione. Una decisione presa il 3 e ribaltata il 9 riportata come sola decisione del 3 è
peggio di nessuna risposta: chi la riceve non ha modo di sapere che è vecchia.
```

- [ ] **Step 7: scrivere i due step, il tetto e il fallimento**

```markdown
## Due modi di essere chiamato

- **Ricerca standard** (il modo normale): rispondi alla domanda con le citazioni pertinenti.
- **Ricerca verbosa**: il dev l'ha chiesta esplicitamente perché la standard non gli è bastata.
  Allarga la finestra, leggi anche le call che avevi escluso, riporta anche i passaggi
  collaterali.

## Tetto

**Questo agente non ha un tetto all'output**, a differenza degli altri agenti del plugin, e la
ragione è nel contratto condiviso: il tetto esiste perché un agente che restituisce troppo
costa più del non delegare, vero quando il valore sta nel rapporto fra letto e restituito. Qui
il valore sta nel non dover riaprire le call: una risposta tagliata costringe a farlo, e
annulla il motivo della delega.

Il vincolo che resta, e che non ha eccezioni: **il parlato grezzo non entra nel context di chi
ti ha chiamato.** Citi i passaggi che servono, non incolli la call.

## Se non trovi la risposta

Scrivi `Risposta: non determinabile dalle trascrizioni`, insieme alla `Copertura`. È un esito
legittimo e utile: dice a chi ti ha chiamato che quella domanda va fatta al dev.

## Fallimento

Se non riesci ad accedere a Drive (connettore assente, permessi, cartella irraggiungibile),
scrivi `RICERCA FALLITA: <motivo>` come unica riga. Non inventare una risposta plausibile e non
restituire un «non trovato»: sono due cose diverse e chi ti ha chiamato deve poterle
distinguere.
```

- [ ] **Step 8: verificare la forma**

Run: `claude plugin validate .`
Expected: nessun errore. Il comando controlla frontmatter delle skill e manifesti — è l'unico
gate automatico di questo repo.

- [ ] **Step 9: verificare che il frontmatter sia allineato agli altri agenti**

Run: `head -6 plugins/wm-skills/agents/*.md`
Expected: `wm-transcript-research.md` ha gli stessi quattro campi degli altri, nello stesso
ordine, senza campi extra.

- [ ] **Step 10: commit (istruzione per il dev — non eseguire)**

Il dev, dopo aver letto il diff:

```bash
git add plugins/wm-skills/agents/wm-transcript-research.md
git commit -m "feat(oc:8530): un agente che cerca nelle trascrizioni e cita chi ha detto cosa"
```

---

### Task 2: L'eccezione nominata in agent-delegation.md

**Files:**
- Modify: `plugins/wm-skills/shared/agent-delegation.md:34` (tabella dei tipi di delega)
- Modify: `plugins/wm-skills/shared/agent-delegation.md:107-113` (sezione `## Tetto all'output`)

**Interfaces:**
- Consumes: il nome dell'agente definito in Task 1.
- Produces: il contratto aggiornato che Task 3 cita senza duplicare.

- [ ] **Step 1: leggere le due sezioni da toccare**

Run: `sed -n '28,40p;105,118p' plugins/wm-skills/shared/agent-delegation.md`

Il file è letto da `wm-plan`, `wm-review-ticket`, `wm-tag` e da tutti gli agenti: ogni modifica
qui vale per tutti.

- [ ] **Step 2: aggiungere l'agente fra i casi di delega informata**

Nella riga `| Casi |` della tabella, aggiungere `wm-transcript-research` all'elenco della
colonna **Delega informata**, dopo `wm-codebase-research`.

Nella riga `| Revisione col dev |` la colonna informata dice già «ammessa, tranne
`wm-env-detect`»: l'agente nuovo rientra negli ammessi, perché il suo esito è un giudizio.
Nessuna modifica necessaria a quella riga — verificarlo, non riscriverlo.

- [ ] **Step 3: aggiungere l'eccezione alla sezione Tetto, senza abrogare la regola**

Dopo il paragrafo esistente che chiude con «va trattato come fallimento (vedi
`## Fallback`).», aggiungere:

```markdown
**Un'eccezione, nominata:** `wm-transcript-research` non ha tetto. La regola vale quando il
valore della delega sta nel rapporto fra letto e restituito — leggo venti file, ti dico una
riga. Per le trascrizioni il valore è un altro: non dover riaprire le call. Una risposta
tagliata costringe a riaprirle, cioè annulla il motivo per cui si è delegato.

L'altro vincolo resta intatto anche lì, ed è quello che conta davvero: **il materiale grezzo
non entra nel context principale**. Chi propone una nuova eccezione argomenti su questo, non
sul precedente: il tetto è la regola.
```

L'ultima frase è la parte che fa il lavoro: senza, la prima eccezione diventa il precedente
per tutte le successive.

- [ ] **Step 4: verificare che la regola sia ancora affermata**

Run: `grep -n "Ogni agente ha un tetto dichiarato" plugins/wm-skills/shared/agent-delegation.md`
Expected: la riga esiste ancora. Se è sparita, la modifica ha abrogato la regola invece di
nominarne l'eccezione: rifare lo Step 3.

- [ ] **Step 5: commit (istruzione per il dev — non eseguire)**

```bash
git add plugins/wm-skills/shared/agent-delegation.md
git commit -m "feat(oc:8530): il tetto all'output resta la regola, con un'eccezione che si argomenta"
```

---

### Task 3: L'invocazione in wm-plan

**Files:**
- Modify: `plugins/wm-skills/skills/wm-plan/SKILL.md:229-236` (tabella «Fasi che delegano»)
- Modify: `plugins/wm-skills/skills/wm-plan/SKILL.md:602-625` (blocco della ricerca in
  `reverse-interaction: dialog`)

**Interfaces:**
- Consumes: il contratto di ritorno di Task 1 (`Citazioni` / `Risposta` / `Copertura`,
  `non determinabile dalle trascrizioni`, `RICERCA FALLITA`).

- [ ] **Step 1: leggere il blocco esistente della ricerca**

Run: `sed -n '600,628p' plugins/wm-skills/skills/wm-plan/SKILL.md`

Il blocco su `wm-codebase-research` è il vicino di casa: la nuova istruzione gli sta accanto e
ne segue la forma, senza modificarlo.

- [ ] **Step 2: aggiungere la riga nella tabella delle deleghe**

Dopo la riga di `reverse-interaction | wm-codebase-research | informata`:

```markdown
| `reverse-interaction` | `wm-transcript-research` | informata |
```

- [ ] **Step 3: aggiungere il blocco nella fase reverse-interaction**

Dopo il blocco che termina con «Se l'agente restituisce `RICERCA FALLITA`, esegui tu la ricerca
nel context principale.» e prima del punto sulle best practice, aggiungere:

```markdown
- **Non chiedere ciò che il team ha già deciso in call.** Insieme alla ricerca sul codice,
  interroga `wm-transcript-research` sulle trascrizioni dello scrum: passagli **numero e
  titolo del ticket**, la sua data di creazione e le domande che stai per fare al dev.

  Vale la stessa regola dell'altra ricerca: **una chiamata sola a inizio fase**, per non
  spezzare il dialogo.

  L'agente restituisce le citazioni prima della conclusione: **leggi le citazioni**, non solo
  la risposta. Su un parlato a più voci la conclusione di un agente è un'interpretazione, e
  chi ha partecipato alla call se ne accorge in un attimo.

  Se una citazione risponde a una domanda che avevi in programma, **non farla come se nulla
  fosse**: dì al dev cosa risulta e chiedi conferma — «allo scrum del 09/09 risulta che avete
  deciso X, confermi?». È diverso dal chiedere da zero, e gli fa risparmiare il fastidio di
  ripetersi.

  Se l'agente risponde `Risposta: non determinabile dalle trascrizioni`, la domanda va fatta
  al dev: è l'esito che rende utile la delega, non un fallimento.

  **Guarda sempre la `Copertura`.** Un «non trovato» su tre call lette su venti non è un «non
  se n'è parlato»: è una ricerca incompleta, e se la domanda è importante chiedi all'agente una
  ricerca verbosa prima di girarla al dev.

  Se l'agente restituisce `RICERCA FALLITA`, prosegui il dialogo senza le trascrizioni e dillo
  al dev: la fonte non era raggiungibile, non è che non ci fosse nulla.
```

- [ ] **Step 4: verificare che le fasi non siano cambiate**

Run: `./.github/scripts/verifica-diagramma.sh`
Expected: passa. Non sono state aggiunte né rinominate fasi — il controllo confronta i nomi.

- [ ] **Step 5: giudicare se il diagramma vada comunque aggiornato**

Lo script stesso dichiara il proprio limite: se una fase cambia comportamento mantenendo il
titolo, il controllo passa e il diagramma resta vecchio. Aprire
`docs/guide/wm-plan-diagramma/index.html`, cercare il nodo `reverse-interaction` e stabilire se
il testo mostrato descriva ancora ciò che la fase fa. Se nomina la ricerca sul codice come
unica fonte, va aggiornato; `.claude/rules/wm-plan-diagramma.md` vincola come si tocca quel
file. **Portare il giudizio al dev invece di decidere da soli.**

- [ ] **Step 6: commit (istruzione per il dev — non eseguire)**

```bash
git add plugins/wm-skills/skills/wm-plan/SKILL.md
git commit -m "feat(oc:8530): il dialogo non richiede ciò che è già stato deciso in call"
```

---

### Task 4: La prova su trascrizioni reali

Questo è il task che vale il lavoro: fino a qui l'agente è scritto, non funzionante. **È anche
il task che richiede il dev**, perché solo chi era nelle call sa se una risposta è giusta.

**Files:**
- Create: `docs/howto/provare-wm-transcript-research.md`
- Modify: `plugins/wm-skills/agents/wm-transcript-research.md` (correzioni che emergono)

**Interfaces:**
- Consumes: l'agente di Task 1.

- [ ] **Step 1: scegliere le domande con risposta nota**

Insieme al dev, scegliere **cinque o sei domande** su call reali recenti di cui lui conosce la
risposta. Almeno una per ciascun grado di riconoscimento:

- una in cui il ticket fu nominato **col numero** (grado 1);
- una in cui il numero fu **troncato o corretto a voce** (grado 2);
- due in cui il ticket fu citato **solo per argomento** (grado 3) — è il caso fragile, e va
  provato più degli altri;
- una **cui la risposta non esiste**: di quel tema non si è parlato. Serve a verificare che
  l'agente dica «non determinabile» invece di costruire una risposta plausibile;
- una che attraversa **due call in date diverse**, dove la seconda corregge la prima.

Le risposte attese le scrive il dev. Nessun altro può.

- [ ] **Step 2: scrivere il file delle prove**

```markdown
# Provare wm-transcript-research

Queste domande hanno una risposta nota, verificata da chi era nella call. Servono a stabilire
se una modifica al prompt dell'agente lo migliora o lo peggiora — senza, «funziona» resta
un'impressione.

**Non è un gate in CI:** la valutazione è di giudizio, e le trascrizioni non stanno nel repo e
non devono finirci. Si rilancia a mano quando si tocca
`plugins/wm-skills/agents/wm-transcript-research.md`.

## Come si esegue

Invocare l'agente su ciascuna domanda e confrontare con la risposta attesa. Guardare tre cose,
in quest'ordine: le **citazioni** sono pertinenti? la **conclusione** segue da esse? la
**copertura** dichiara abbastanza call?

## Le prove

### 1. Ticket nominato col numero — <argomento>

**Domanda:** <...>
**Risposta attesa:** <...>
**Call di provenienza:** <AAAA/MM/GG>

<... una sezione per prova ...>

### 6. Nessuna risposta esiste — <argomento>

**Domanda:** <...>
**Risposta attesa:** `non determinabile dalle trascrizioni`. Una risposta qualsiasi diversa da
questa è un fallimento, anche se suona plausibile.
```

- [ ] **Step 3: primo giro**

Invocare l'agente su tutte le domande. Annotare per ciascuna: citazioni pertinenti sì/no,
conclusione corretta sì/no, copertura dichiarata.

- [ ] **Step 4: correggere il prompt**

Correggere `wm-transcript-research.md` sugli scostamenti trovati. Le correzioni tipiche
attese, per non cercarle da zero:

- l'agente **cerca troppo poco** → allargare la finestra o chiarire i due regimi (call di oggi
  integrali, filtro sul resto);
- l'agente **riporta call non pertinenti** → il grado 3 è troppo generoso: chiedere che dichiari
  l'incertezza invece di includere;
- l'agente **riassume invece di citare** → rafforzare che il testo citato è verbatim;
- l'agente **omette la copertura** → è obbligatoria in ogni risposta, anche quando trova.

- [ ] **Step 5: secondo giro, e terzo se serve**

Ripetere Step 3 e 4. **Fermarsi al terzo giro** e portare la decisione al dev: se il
riconoscimento per argomento non converge, il problema non è il prompt ma quanto è recuperabile
quel parlato.

**Fallback dichiarato in overview:** restringere lo scopo al riconoscimento per numero (gradi 1
e 2), che dai dati osservati funziona bene, e registrare il grado 3 come follow-up. Meglio un
agente che trova meno ma non sbaglia, che uno di cui non ci si fida.

- [ ] **Step 6: commit (istruzione per il dev — non eseguire)**

```bash
git add docs/howto/provare-wm-transcript-research.md plugins/wm-skills/agents/wm-transcript-research.md
git commit -m "feat(oc:8530): le domande con risposta nota, e il prompt corretto su ciò che sbagliava"
```

---

### Task 5: La conoscenza e l'indice

**Files:**
- Create: `docs/knowledge/wm-transcript-research.md`
- Modify: `CLAUDE.md` (tabella sotto `## Conoscenza`)

- [ ] **Step 1: verificare che non esista già una pagina sul tema**

Run: `ls docs/knowledge/`
Expected: nessuna pagina copre le trascrizioni come fonte. La più vicina è
`wm-skills-delega-agentica.md`, che copre il *meccanismo* di delega, non questa fonte. Sono temi
distinti: pagina nuova.

- [ ] **Step 2: scrivere la pagina**

Struttura prescritta dal repo — lo stato attuale in cima, la storia sotto:

```markdown
# Le trascrizioni come fonte di verità

## Come funziona oggi

<Com'è fatta la fonte: Doc su Drive, uno per call, data nel titolo, speaker per battuta,
marcatori ogni cinque minuti e nessun timestamp per battuta. Come si selezionano le call: call
di oggi integrali, filtro full-text sul resto. Come si riconosce il ticket: per numero, per
numero troncato, per argomento. Perché non si cerca per termine tecnico.>

## Perché così

- **Connettore Drive e non un MCP** (oc:8530): <la decisione e le alternative scartate —
  Meet MCP, Spinach, Composio — con il motivo di ciascuna>
- **Citazioni prima della conclusione** (oc:8530): <su un parlato a più voci chi legge deve
  poter dissentire dalla sintesi>
- **Nessun tetto all'output** (oc:8530): <il valore sta nel non dover riaprire le call>
- **Nessuna attivazione automatica in esecuzione** (oc:8530): <nessuno sa dire quando scatta
  «è sorto un dubbio»: lo dice il dev>

## Come ci siamo arrivati

- **Un MCP terzo per l'accesso** (oc:8530, scartata): l'ipotesi iniziale era che l'accesso
  andasse costruito o comprato. L'ispezione della cartella ha mostrato che le trascrizioni sono
  già Doc su Drive, leggibili col connettore esistente. L'API Meet darebbe il timestamp per
  battuta, ma non ha il filtro full-text: peggiore nel punto che conta.
```

Ogni voce porta il ticket: è il legame col cantiere, che resta la fonte completa.

- [ ] **Step 3: preparare la riga per l'indice**

```markdown
| Le trascrizioni come fonte | Com'è fatta la fonte, come si scelgono le call, perché Drive e non un MCP | [docs/knowledge/wm-transcript-research.md](docs/knowledge/wm-transcript-research.md) |
```

Link Markdown, **mai `@percorso`**: in un `CLAUDE.md` quella sintassi è un import e il file
verrebbe caricato a ogni sessione, annullando il motivo di averlo separato.

- [ ] **Step 4: far controllare l'aggiunta**

Invocare `wm-context-guard` passando il percorso del `CLAUDE.md` e il testo proposto. Verificare
ogni rilievo che cita una voce esistente con l'estratto fornito. Una **contraddizione** non si
risolve da soli: si porta al dev.

- [ ] **Step 5: mostrare al dev e scrivere**

Mostrare la riga al dev prima di scriverla nel `CLAUDE.md`.

- [ ] **Step 6: commit (istruzione per il dev — non eseguire)**

```bash
git add docs/knowledge/wm-transcript-research.md CLAUDE.md
git commit -m "feat(oc:8530): come si legge una call, e perché Drive invece di un MCP"
```

---

### Task 6: Chiusura

- [ ] **Step 1: gate di forma**

Run: `claude plugin validate .`
Expected: nessun errore.

- [ ] **Step 2: coerenza skill/diagramma**

Run: `./.github/scripts/verifica-diagramma.sh`
Expected: passa.

- [ ] **Step 3: notes.md**

Compilare `docs/features/8530-wm-transcript-research/notes.md` con quanto emerso: i giri di
correzione del Task 4, le deviazioni, le decisioni prese durante l'esecuzione, i follow-up.
Deve esistere anche se dicesse «nessuna deviazione».

- [ ] **Step 4: riportare i rimandi rotti a zero**

```bash
cd docs/features/8530-wm-transcript-research
grep -o 'notes\.md#[a-z0-9-]*' plan.md | sed 's/.*#//' | sort -u > /tmp/wm-ancore-citate
grep '^### ' notes.md | sed 's/^### //' \
  | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9 -]//g; s/ /-/g' | sort -u > /tmp/wm-ancore-esistenti
comm -23 /tmp/wm-ancore-citate /tmp/wm-ancore-esistenti
```

Expected: nessun output. Se stampa qualcosa, quei rimandi puntano a titoli che non esistono.

- [ ] **Step 5: aggiornare il ticket**

La `description` di oc:8530 è **superata**: descrive la scelta fra Meet MCP e Spinach e i punti
aperti sul riconoscimento dell'ID, tutti chiusi da questo lavoro. Riscriverla in HTML con
quanto realmente implementato e il link alla cartella degli artefatti. Anteprima senza
`confirm`, approvazione del dev, poi `confirm: true`.

- [ ] **Step 6: review-gate**

Il riepilogo del diff lo produce un subagente isolato. `has_phpstan_ci: false` in questo repo:
nessun check PHPStan. Il dev legge il diff e decide i commit.

---

## Rischi noti durante l'esecuzione

| Rischio | Dove si manifesta | Cosa fare |
|---|---|---|
| Il riconoscimento per argomento non converge | Task 4, dopo 2-3 giri | restringere ai gradi 1-2, grado 3 come follow-up |
| L'elenco per cartella non funziona da un secondo account | prima prova di un altro dev | è il meccanismo su cui poggiano tutti i casi d'uso: fermarsi e portarlo al dev |
| La modifica ad `agent-delegation.md` abroga invece di eccettuare | Task 2, Step 4 | il `grep` di verifica è lì per questo |
| Il diagramma resta vecchio pur passando il controllo | Task 3, Step 5 | è giudizio, non script: portarlo al dev |
