# Le trascrizioni come fonte di verità

## Come funziona oggi

Lo scrum quotidiano del team si tiene sempre sullo stesso canale Meet, con la trascrizione
attiva, e più volte al giorno ci si torna per discutere problemi emersi sui task. Le decisioni
prese lì restano nel parlato: `wm-transcript-research` le rende consultabili.

**La fonte.** Le trascrizioni sono Google Doc in una cartella Drive, una per call, con titolo
`scr-um w-ebmapp - AAAA/MM/GG HH:MM CEST - Transcript`. Struttura fissa: intestazione,
`Partecipanti`, poi le battute come `Nome Cognome: testo`. **Lo speaker c'è sempre.** I
timestamp per battuta no: ci sono solo marcatori ogni cinque minuti come offset dall'inizio,
quindi una citazione si colloca in una finestra, non al secondo.

**Quali call si leggono.** La finestra va da qualche giorno prima della creazione del ticket al
giorno di lavorazione — non dalla creazione, perché la call in cui si decide *di aprire* un
ticket è precedente alla sua esistenza. Dentro la finestra valgono due regimi: **le call di
oggi si leggono tutte**, sul resto si usa il filtro full-text di Drive. La ragione è che in
Drive la creazione del file e l'indicizzazione del contenuto sono cose distinte — il Doc esiste
entro pochi minuti dalla fine della call, ma la ricerca per contenuto passa da un indice con
tempi non garantiti, e un contenuto non ancora indicizzato dà zero risultati indistinguibili da
«non se n'è parlato». Il filtro sul titolo usa solo la parte `AAAA/MM/GG`: `CEST` diventa `CET`
a fine ottobre.

**Come si riconosce di quale ticket si parla.** Per gradi: il **numero**, che nel parlato
trascritto compare come cifre attaccate («ticket 8506») e mai col prefisso `oc`; il **numero
troncato**, perché chi parla si interrompe e riprende («il ticket 84, scusa, 8487»); e
l'**argomento**, che è il caso più frequente e il meno affidabile — «il ticket di Carla»,
«quello della compromissione». Il terzo grado va sempre dichiarato come tale nella citazione.

**Perché non si cerca per termine tecnico.** La trascrizione automatica storpia il lessico di
dominio: *cron job* diventa «Chrome job», *.htaccess* «file HT assess», *uploads* «Appls», e
«WM Plan» e «VM Plan» compaiono nella stessa riunione. Un filtro su un termine tecnico non
troverebbe nulla. Per lo stesso motivo le citazioni restano verbatim, storpiature comprese:
l'interpretazione sta nella conclusione, non nel testo citato.

**Cosa restituisce.** Prima le citazioni — con speaker, data e finestra di cinque minuti, ed
etichetta `deciso` / `proposto` / `obiezione` sul singolo passaggio — poi la conclusione, poi
le `Fonti` con l'id del file Drive e la `Copertura`, cioè quante call ha letto su quante.

## Perché così

- **Connettore Drive, non un MCP** (oc:8530): l'ipotesi di partenza era di adottare un server
  MCP per l'accesso. L'ispezione ha mostrato che il connettore Drive copre tutto — elenco per
  data, filtro full-text, lettura — e che **solo lui ha il filtro full-text**, che l'API Meet
  non offre. La distribuzione non è stata un argomento: il team è piccolo e configurare il
  connettore su tutte le macchine è rapido, quindi un MCP si sarebbe giustificato solo
  aggiungendo valore.
- **Citazioni prima della conclusione** (oc:8530): su un parlato a più voci la sintesi è già
  un'interpretazione, e chi era nella call se ne accorge in un attimo. Chi legge deve poter
  dissentire dalla conclusione guardando il materiale, non doverla prendere sulla fiducia.
- **L'etichetta sta sulla citazione, non sulla conclusione** (oc:8530): è del singolo passaggio
  che si può dire se fosse una decisione o una proposta. «Lo facciamo così» detto da chi decide
  e detto da chi propone non pesano uguale.
- **`Fonti` con l'id del file** (oc:8530): un link si apre nel browser, un id permette di
  riaprire la call dalla sessione. È ciò che rende una risposta verificabile dopo che l'agente
  ha finito.
- **`Copertura` obbligatoria** (oc:8530): senza, un «non trovato» dopo tre call lette su venti
  è indistinguibile da uno dopo venti su venti, e chi legge smette di cercare in entrambi i
  casi.
- **Nessun tetto all'output** (oc:8530): unica eccezione fra gli agenti del plugin. Il tetto
  esiste perché un agente che restituisce troppo costa più del non delegare — vero quando il
  valore sta nel rapporto fra letto e restituito. Qui il valore è non dover riaprire le call, e
  una risposta tagliata costringe a farlo. Il vincolo che resta è l'altro: il parlato grezzo non
  entra nel context principale.
- **Nessuna attivazione automatica durante l'esecuzione** (oc:8530): l'agente è delega statica
  in `reverse-interaction`, e fuori di lì lo invoca il dev. Nessuno sa dire quando scatta «è
  sorto un dubbio»: una condizione del genere o non scatta mai o scatta sempre.
- **Il contenuto sensibile è gestito dal dev, non da un gate** (oc:8530): nelle call si parla di
  clienti e di compromissioni di siti, e `wm-plan` gira anche su repo cliente. Non c'è redazione
  automatica: il controllo è l'approvazione del dev su ogni artefatto, più il divieto di commit
  dentro il lavoro.

## Come ci siamo arrivati

- **Un MCP terzo per l'accesso** (oc:8530, scartata): si era ipotizzato di adottare il Google
  Meet MCP server, Spinach o Composio. L'API Meet darebbe il timestamp per battuta, che il Doc
  non ha, ma non cerca nel testo: peggiore nel punto che conta di più. Spinach è stato escluso
  a monte perché il materiale uscirebbe verso un servizio terzo, e nelle call si parla di
  incidenti di sicurezza di clienti nominati.
- **Il numero di ticket detto cifra per cifra** (oc:8530, smentita): si era assunto che nel
  parlato gli ID venissero pronunciati una cifra alla volta e quindi fossero irrintracciabili.
  La lettura delle trascrizioni ha mostrato il contrario — sempre cifre attaccate — e ha reso il
  filtro per numero il primo grado di riconoscimento invece di una strada da escludere.
- **Un archivio centralizzato delle trascrizioni** (oc:8530, scartata): avrebbe reso le
  citazioni verificabili con `sed` come per il codice, ma avrebbe creato un deposito di
  conversazioni interne da custodire e da tenere aggiornato ogni giorno. Le trascrizioni
  restano su Drive.
