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

**Chi legge le call.** Le legge NotebookLM, non l'agente, e ogni call si carica una volta sola per
dev. I notebook sono due tipi, nell'account NotebookLM di ciascuno: `scrum AAAA-MM-GG`, con le call
dello scrum di quella giornata, e `tag <nome del tag>`, con le trascrizioni del tag (righe
`**Fonte:**` e Google Doc linkati). All'avvio di `wm-plan` l'agente prepara in background il
notebook del giorno, e le altre richieste aspettano che abbia finito (due richieste insieme creano
lo stesso notebook due volte; se capita, l'agente usa il più vecchio e segnala gli altri); nel caso B (ticket nuovo) lo interroga prima di formulare la bozza; in
`reverse-interaction` interroga i notebook dei giorni della finestra (da qualche giorno prima della
creazione del ticket a oggi, solo i giorni con almeno una call; sui giorni in `progress` quando
l'API li esporrà, oc:8636) e quelli dei tag del ticket **che esistono già**, in parallelo, con tutte
le domande in una chiamata per notebook, e mette insieme le risposte in ordine di data. Il parallelo
regge solo con `notebook_query_start` e `notebook_query_status`: Claude Code esegue una alla volta le
chiamate allo stesso server MCP, e con `notebook_query`, che aspetta la risposta, dieci notebook
costano dieci attese in fila. Altri giorni,
o il notebook di un tag che non c'è ancora, si aggiungono su domanda diretta del dev. Le call si elencano da Drive per data nel titolo, seguendo tutte le
pagine; Fogli Google, siti web e GitHub non diventano fonti, perché una citazione da un Foglio porta
con sé la tabella intera e fa superare alla risposta il limite di output. Il server MCP è
`notebooklm`, dichiarato dal plugin; l'installazione è in
[docs/howto/installare-notebooklm.md](../howto/installare-notebooklm.md).

**Il template dei notebook.** Appena creato, ogni notebook riceve con `chat_configure` le regole
del team (citazioni verbatim con chi parla e marcatore, niente ricuciture, etichette, ordine di
data, «non determinabile»; per i notebook dei tag anche cliente e Webmapp distinti e le etichette
`richiesta` e `da chiarire`). Il template sta nel prompt dell'agente e detta **regole, mai un
formato**: le stesse valgono per le domande che il dev fa direttamente dal link del notebook.

**Attribuzioni.** La data di una citazione viene dal notebook (un giorno); titolo e id Drive si
ricavano solo dalla tabella di `source_list_drive` (`source_id` → titolo, `drive_doc_id`), a partire
dal riferimento della citazione; ogni citazione si confronta con il suo `cited_text`, cioè il testo originale del
passaggio: se non combacia, l'agente rifà la domanda solo su quel punto prima di scartarla. `wm-plan` e `wm-tag` verificano poi ogni citazione con una ricerca Drive sulla frase
([plugins/wm-skills/shared/verifica-citazioni.md](../../plugins/wm-skills/shared/verifica-citazioni.md)).

**`wm-tag`.** Alla creazione di un tag la trascrizione la legge `wm-tag` e scrive lui le macro aree;
crea il notebook `tag <nome del tag>`, il cui link sta nella descrizione, che poi risponde alle
domande sulla call sia a `wm-tag` sia a `wm-plan`.

**Come si riconosce di quale ticket si parla.** Per gradi: il **numero**, che nel parlato
trascritto compare come cifre attaccate («ticket 8506») e mai col prefisso `oc`; il **numero
troncato**, perché chi parla si interrompe e riprende («il ticket 84, scusa, 8487»); e
l'**argomento**, che è il caso più frequente e il meno affidabile — «il ticket di Carla»,
«quello della compromissione». Il terzo grado va sempre dichiarato come tale nella citazione.

**Perché non si cerca per termine tecnico.** La trascrizione automatica storpia il lessico di
dominio: *cron job* diventa «Chrome job», *.htaccess* «file HT assess», *uploads* «Appls», e
«WM Plan» e «VM Plan» compaiono nella stessa riunione. Per questo le call si scelgono per data e
non per parola, e le citazioni restano verbatim, storpiature comprese: l'interpretazione sta nella
conclusione, non nel testo citato.

**Cosa restituisce.** Prima le citazioni — con speaker, data e finestra di cinque minuti, ed
etichetta `deciso` / `proposto` / `obiezione` / `fatto riferito` sul singolo passaggio (nei
notebook dei tag anche `richiesta` / `da chiarire`) — poi la conclusione, poi le `Fonti` con l'id
del file Drive, la `Copertura` per notebook (call dell'elenco di Drive e call caricate, tag senza
notebook, documenti non caricati) e i link dei notebook.

## Perché così

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
  (avvio di `wm-plan`, caso B, `reverse-interaction`, `wm-tag`), e fuori di lì lo invoca il dev. Nessuno sa dire quando scatta «è
  sorto un dubbio»: una condizione del genere o non scatta mai o scatta sempre.
- **Il contenuto sensibile è gestito dal dev, non da un gate** (oc:8530): nelle call si parla di
  clienti e di compromissioni di siti, e `wm-plan` gira anche su repo cliente. Non c'è redazione
  automatica: il controllo è l'approvazione del dev su ogni artefatto, più il divieto di commit
  dentro il lavoro.
- **NotebookLM invece della lettura diretta da Drive** (23/09/2026): leggendo le call da sé
  l'agente ha attribuito citazioni alla call sbagliata in due modi — sulle call oltre il limite
  di output dei tool MCP, e scambiando gli id fra elenco e documenti letti — e le regole nel
  prompt per evitarlo non sono state rispettate. Sulle stesse call NotebookLM ha dato 11
  citazioni su 11 alla call giusta, ciascuna con il passaggio originale, a circa 9k token contro
  69–112k. Per `wm-tag` risponde alle domande ma non scrive il tag: sulla struttura di una call
  intera ha perso 2 macro aree su 13 e sbagliato le etichette in 6.
- **Un notebook per giorno e uno per tag** (23/09/2026): la prima versione creava un notebook per
  lavoro e ricaricava le call ogni volta (29 fonti e 7 minuti e mezzo per oc:8543). Con notebook
  per giorno e per tag ogni call si carica una volta, i notebook sono piccoli, e la data di una
  risposta è data dal notebook. La finestra andrà calcolata sui giorni in cui il ticket è stato in
  `progress` quando l'API li esporrà (oc:8636).

## Come ci siamo arrivati

- **Connettore Drive, non un MCP** (oc:8530, superata il 23/09/2026): la lettura ora passa da
  NotebookLM; Drive resta per l'elenco delle call e la verifica delle citazioni. La scelta di allora: l'ipotesi di partenza era di adottare un server
  MCP per l'accesso. L'ispezione ha mostrato che il connettore Drive copre tutto — elenco per
  data, filtro full-text, lettura — e che **solo lui ha il filtro full-text**, che l'API Meet
  non offre. La distribuzione non è stata un argomento: il team è piccolo e configurare il
  connettore su tutte le macchine è rapido, quindi un MCP si sarebbe giustificato solo
  aggiungendo valore.
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
