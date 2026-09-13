> Ticket: oc:8530

# wm-transcript-research — le trascrizioni come fonte di verità

## Cosa cambia

Oggi ciò che il team decide in call resta solo nella trascrizione. Chi implementa non ha modo
di sapere che il dubbio che ha davanti ha già avuto risposta allo scrum, e lo richiede al dev —
o peggio, decide da sé qualcosa che era già stato deciso diversamente.

Dopo questa feature le trascrizioni diventano una fonte interrogabile, e il lavoro di leggerle
è delegato a un agente — `wm-transcript-research` — che le legge nel proprio context e
restituisce solo ciò che risponde alla domanda. Il parlato grezzo, che è voluminoso e cresce
ogni giorno, non entra mai nel context della sessione principale.

L'agente serve tre casi d'uso distinti:

| Caso | Input | Output |
|---|---|---|
| Integrazione automatica in `wm-plan` | il ticket | citazioni, poi la risposta |
| Domanda esplicita del dev | una domanda | citazioni, poi la risposta |
| «Cosa è stato discusso il giorno Z» | una data | resoconto contestualizzato della giornata |

Nel primo caso l'ordine delle fonti è fisso: **ticket → trascrizioni linkate nel ticket o nel
tag → scrum**. Il ticket resta il punto di partenza perché è lì che la richiesta è scritta; le
trascrizioni *integrano*, non sostituiscono.

## Perché

Lo scrum quotidiano è il luogo dove il lavoro viene assegnato e dove si risponde alle domande
di chi lo prende in carico. Un dev che riceve un ticket fa domande e ottiene risposte: quelle
risposte oggi non finiscono da nessuna parte. Lo stesso canale viene riusato più volte al
giorno per decidere su problemi emersi, e vale lo stesso.

## La fonte, come è fatta davvero

Ispezione diretta della cartella Drive delle trascrizioni e lettura di quattro call
(7-11 settembre 2026). **Questi fatti hanno precedenza su qualunque assunzione: sono stati
osservati, non ipotizzati.**

- **Sono Google Doc in una cartella Drive**, uno per call, con titolo
  `scr-um w-ebmapp - AAAA/MM/GG HH:MM CEST - Transcript`. La data è nel nome del file: la
  selezione temporale è una query sul titolo, non serve interrogare l'API Meet.
- **Struttura fissa**: intestazione `SCRUM WEBMAPP`, sezione `Partecipanti` con i nomi,
  poi le battute come `Nome Cognome: testo`.
- **Lo speaker è sempre presente**, su ogni battuta, con nome e cognome completi.
- **Non esistono timestamp per battuta.** Ci sono solo marcatori ogni cinque minuti
  (`### 00:05:00`) come offset dall'inizio, più la durata finale. Una citazione si colloca al
  massimo in una finestra di cinque minuti.
- **Gli ID dei ticket compaiono come cifre attaccate a quattro** — «ticket 8506», «8474» — mai
  staccate, mai in lettere. **Il prefisso `oc` non compare mai**: il numero è nudo, preceduto
  al più dalla parola «ticket».
- **Ma i numeri si troncano nel parlato**: «il ticket 848 e gli altri due», «il ticket 84,
  scusa, 8487». Una ricerca esatta perde questi casi.
- **E molto spesso il ticket non ha numero affatto**: «il ticket di Carla», «quello che
  riguarda l'analisi sulla compromissione».
- **Nessun link o URL nel testo**: solo parlato. I link stanno nei ticket, mai nelle call.
- **I termini tecnici e i nomi propri sono molto inaffidabili**: *cron job* → «Chrome job»,
  *.htaccess* → «file HT assess», *uploads* → «Appls»/«Applauds», e «VM Plan» accanto a
  «WM Plan» nella stessa riunione.

Conseguenza di progetto: **cercare per numero funziona ed è il primo filtro; cercare per
termine tecnico è da escludere; tutto il resto va riconosciuto leggendo, che è ciò che un LLM
fa e un `grep` no.**

## Decisione: connettore Drive, nessun MCP terzo

L'accesso passa dal **connettore Google Drive**, non da un server MCP. Decisa il 13/09/2026
dopo aver verificato la fonte, e motivata così:

- **Drive copre tutto ciò che serve**: elenco delle call per data (la data è nel titolo del
  file), lettura del Doc completo, speaker per battuta.
- **Solo Drive ha il filtro full-text**, eseguito da Google prima della lettura, mentre l'API
  Meet non cerca nel testo affatto. Attenzione però a non sopravvalutarlo: il filtro produce
  hit quando il ticket è stato nominato col numero, cioè nel caso già facile. Quando il ticket
  è citato solo per argomento — il caso frequente e difficile — non filtra nulla e si legge la
  finestra. Drive resta uguale o migliore dell'MCP in ogni scenario, ma **il costo atteso della
  feature va calcolato sulla lettura integrale della finestra**, non sul caso fortunato.
- **Funziona adesso**: usato in questa sessione per elencare, filtrare e leggere. Un MCP
  andrebbe scelto, configurato e provato prima di sapere se l'idea regge.
- **La distribuzione non è un argomento**: il team è piccolo e configurare il connettore su
  tutte le macchine è rapido. Un MCP si giustifica solo se aggiunge valore, e qui non lo fa.
- **Cosa si rinuncia**: l'API Meet darebbe il timestamp per battuta, il Doc solo marcatori ogni
  cinque minuti. Accettato: con il link al Doc e il nome di chi parla, il passaggio si ritrova
  in pochi secondi.
- **Spinach e gli altri servizi terzi sono esclusi**: nelle call si parla di clienti, di
  compromissioni di siti e di problemi di sicurezza, e mandarle fuori non vale la ricerca
  semantica.

Se un giorno servisse il timestamp esatto, cambia lo strumento di lettura e non il resto
dell'agente.

## Requisiti

- [x] Esiste l'agente `wm-transcript-research` in `plugins/wm-skills/agents/`, con frontmatter
      `name`, `description`, `model`, `tools` come gli altri agenti del plugin
- [x] L'agente accede alle trascrizioni **tramite il connettore Google Drive**, leggendo i Doc
      della cartella dello scrum. Non serve l'API Meet né un servizio terzo
- [ ] **L'elenco delle call per data funziona anche per un dev diverso dal proprietario.** I
      dev accedono alle trascrizioni (confermato dal team); l'API però mostra come permesso
      esplicito il solo proprietario, sia sui Doc sia sulla cartella. Poiché in Drive *aprire*
      un file e *elencarlo* in una cartella sono permessi distinti, e l'agente parte
      dall'elenco per trovare le call in una finestra di date, la query per cartella va provata
      da un secondo account al primo test
- [x] La selezione delle trascrizioni candidate avviene **per data dal titolo del file**, con
      la finestra: **da qualche giorno prima della creazione del ticket** al giorno di
      lavorazione. Non dalla creazione: la discussione in cui si decide *di aprire* un ticket
      avviene prima che il ticket esista, ed è spesso quella che ne spiega il perché. Lo stato
      `progress` restringe ulteriormente
- [x] Il filtro sul titolo del file usa **solo la parte `AAAA/MM/GG`**, ignorando il suffisso di
      fuso: `CEST` diventa `CET` a fine ottobre. E il riconoscimento del numero di ticket **non
      assume quattro cifre**: gli ID cresceranno
- [x] **Le call della giornata corrente si leggono sempre per intero; sul resto della finestra
      si usa il filtro full-text.** Motivo: in Drive la creazione del file e l'indicizzazione
      del suo contenuto sono due cose distinte — il Doc esiste entro pochi minuti dalla fine
      della call (arriva la mail), ma la ricerca per contenuto passa da un indice con tempi non
      garantiti. Un contenuto non ancora indicizzato dà zero risultati, indistinguibili da «non
      se n'è parlato»: sulle call di oggi non ci si affida all'indice
- [x] Il filtro full-text cerca **il numero del ticket e le parole chiave del suo titolo**, non
      il solo numero: è così che le persone nominano un ticket quando non ne dicono l'ID
- [x] Il riconoscimento del ticket in una call procede per gradi: **il numero** come primo
      filtro — scritto come cifre attaccate, senza prefisso `oc` e senza assumere una lunghezza
      fissa — poi **numero troncato** e **argomento/titolo** riconosciuti leggendo. All'agente
      si passano sia l'ID sia il titolo del ticket
- [x] L'agente restituisce **prima un array di citazioni**, ciascuna con testo verbatim,
      **speaker** e **collocazione temporale** (data della call più la finestra di cinque
      minuti), e **solo dopo** la risposta che ne deriva
- [x] Ogni citazione porta il **link al Doc Drive** della call, apribile dal dev
- [x] L'agente distingue ciò che è stato **deciso** da ciò che è stato solo **proposto** o
      **obiettato**, etichettando la singola citazione e non la conclusione
- [x] L'agente non corregge in silenzio i termini storpiati: se interpreta «Chrome job» come
      *cron job*, la citazione resta verbatim e l'interpretazione sta nella conclusione
- [x] **Nessun tetto all'output**: la risposta è sempre dettagliata. Il vincolo che resta è che
      il materiale grezzo non entri nel context principale
- [x] L'agente si invoca in **due step**: una ricerca *standard* che risponde subito, e una
      ricerca *verbosa* che **il dev** chiede quando la prima non basta. La verbosa è sua, non
      del flusso automatico: in `reverse-interaction` non c'è nessuno che sappia cosa manca,
      quindi lì si usa sempre e solo la standard
- [x] Quando l'agente non trova la risposta lo dichiara — esito legittimo che dice al chiamante
      di girare la domanda al dev — e ha una riga di fallimento riconoscibile per il caso in cui
      non riesca ad accedere a Drive
- [x] **Ogni risposta dichiara cosa è stato letto**: quante call, quali date, con quale filtro.
      Senza, un «non se n'è parlato» dopo tre call su venti è indistinguibile da uno dopo venti
      su venti, e il dev smette di cercare in entrambi i casi
- [x] `wm-plan` invoca l'agente in `reverse-interaction`, come delega statica
- [x] Il dev può invocarlo esplicitamente in qualsiasi momento, anche a implementazione avviata
- [x] `shared/agent-delegation.md` **mantiene la regola del tetto e vi nomina l'eccezione**,
      senza abrogarla: il tetto esiste perché un agente che restituisce troppo costa più del
      non delegare, vero quando il valore sta nel rapporto fra letto e restituito. Per le
      trascrizioni il valore sta nel non dover riaprire le call, quindi tagliare l'output
      distrugge il motivo della delega. Il vincolo che resta valido anche qui è l'altro: il
      materiale grezzo non entra nel context principale. Il file è letto da tutte le skill:
      abrogare la regola per un caso singolo sarebbe debito permanente
- [x] **Esiste un insieme di prove con risposta nota**, in `docs/howto/`: cinque o sei domande
      prese da call reali, ciascuna con la risposta corretta (scritta dal dev che era presente)
      e la call di provenienza. Chi modifica il prompt dell'agente rilancia quelle domande e
      confronta. Non è un gate in CI — la valutazione è di giudizio e le trascrizioni non
      stanno nel repo — ma senza, «l'agente funziona» resta un'impressione e una modifica al
      prompt può peggiorarlo senza che nessuno se ne accorga
- [x] `claude plugin validate .` passa

## Rischi

*Da completare dopo la Fase: challenge.*

- **Il ticket nominato senza numero** — è il caso più frequente dopo quello col numero, e
  l'unico che nessun filtro meccanico può catturare. Il riconoscimento dipende interamente dal
  giudizio dell'agente sul senso del discorso, quindi può sia mancare riferimenti veri sia
  attribuire al ticket sbagliato una discussione su un tema vicino.
- **Falsa sicurezza di una risposta trovata** — un dev che riceve «allo scrum avete deciso X»
  smette di cercare. Se l'agente ha frainteso, l'errore si propaga nel codice senza che nessuno
  lo verifichi: è la ragione per cui le citazioni vengono prima della conclusione.
- **Decisione superata da una call successiva** — una risposta corretta ma vecchia è peggio di
  nessuna risposta. La data su ogni citazione è la mitigazione minima; resta da decidere se
  l'agente debba segnalare attivamente le tensioni fra passaggi di date diverse.
- **Nessun tetto all'output** — un agente che restituisce molto può costare più del non
  delegare. Scelta motivata, ma il caso «giorno Z» è dove può degenerare: mitigato dai due step.
- **Volume di lettura** — le call sono più d'una al giorno e alcune superano i 60 KB. Una
  finestra di due settimane sono decine di documenti: va deciso quanti l'agente può aprirne
  prima di dichiarare che la domanda è troppo larga.
- **Trascrizione inaffidabile sui termini tecnici** — il rischio non è solo non trovare, è
  capire male. «Chrome job» letto alla lettera cambia il senso di una decisione.
- **Elenco per cartella da un altro account** — i dev accedono alle trascrizioni, ma l'unico
  permesso esplicito su cartella e Doc è quello del proprietario: l'accesso degli altri passa
  presumibilmente dal link ricevuto per posta. Se è così, aprono i singoli Doc ma potrebbero
  non poter elencare la cartella — e l'agente cerca proprio partendo dall'elenco. Da provare
  con un secondo account prima di considerare la feature funzionante per il team.
- **Contenuto sensibile — rischio accettato dal dev (13/09/2026).** Nelle call si parla di
  clienti, di compromissioni di siti e di problemi di sicurezza, e ciò che l'agente riporta può
  finire in artefatti committati; `wm-plan` gira anche su repo cliente. Nessun gate automatico
  di redazione: il controllo è il dev, che approva ogni artefatto, e la prima regola del
  `CLAUDE.md` vieta comunque i commit dentro il lavoro. Nessun contenuto arriva a un push senza
  essere stato letto da una persona.

## Out of scope

- **Non si implementa l'accesso alle trascrizioni**: si usa il connettore Drive esistente.
- Nessun archivio centralizzato e nessuna copia locale delle trascrizioni: restano su Drive.
- Nessuna correzione o normalizzazione del testo trascritto.
- Nessuna scrittura sulle trascrizioni: l'agente legge e basta.
- Nessuna attivazione automatica durante l'esecuzione: fuori da `reverse-interaction` l'agente
  si invoca su richiesta esplicita del dev.
- Non si tocca `wm-tag`, che continua a trattare una trascrizione come input fornito dal dev
  per generare ticket — verso opposto a questo.

## Moduli toccati

Feature interamente **custom** su `claude-marketplace`: nessun submodule coinvolto.

| File | Cosa |
|---|---|
| `plugins/wm-skills/agents/wm-transcript-research.md` | nuovo — prompt dell'agente |
| `plugins/wm-skills/shared/agent-delegation.md` | modificato — eccezione nominata al tetto all'output, regola invariata |
| `plugins/wm-skills/skills/wm-plan/SKILL.md` | modificato — invocazione in `reverse-interaction`, tabella delle deleghe |
| `docs/knowledge/<argomento>.md` | nuovo o aggiornato in `Fase: update-context` |
| `docs/features/8530-wm-transcript-research/` | artefatti di questo lavoro |
