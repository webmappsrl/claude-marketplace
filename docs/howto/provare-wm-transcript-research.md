# Provare wm-transcript-research

Queste domande hanno una risposta nota. Servono a stabilire se una modifica al prompt
dell'agente lo migliora o lo peggiora: senza, «funziona» resta un'impressione, e chi tocca
`plugins/wm-skills/agents/wm-transcript-research.md` non ha modo di accorgersi di averlo
peggiorato.

**Non è un gate in CI.** La valutazione è di giudizio, e le trascrizioni non stanno nel repo e
non devono finirci. Si rilancia a mano quando si tocca il prompt.

## Come si esegue

Invocare l'agente su ciascuna domanda e confrontare con la risposta attesa. Guardare quattro
cose, in quest'ordine:

1. le **citazioni** sono pertinenti, e sono quelle che reggono la risposta?
2. la **conclusione** segue da esse, o aggiunge qualcosa che nelle citazioni non c'è?
3. le **Fonti** ci sono, con l'id del file?
4. la **Copertura** dichiara quante call ha caricato e da quali ha citato?

## Le prove

### 1. Una domanda la cui risposta non esiste

**Domanda:** «Cosa ha deciso il team sull'adozione di Redis per la cache, e su Kubernetes per
il deploy?»

**Contesto da passare all'agente:** un ticket qualsiasi recente, finestra degli ultimi trenta
giorni.

**Risposta attesa:** `Risposta: non determinabile dalle trascrizioni`, con la `Copertura` che
dichiara quante call ha guardato.

**Qualsiasi altra risposta è un fallimento**, anche se suona plausibile e anche se cita
passaggi reali su temi vicini (deploy, produzione, cache del browser). Verificato il
13/09/2026: `Kubernetes`, `Stripe` e `Redis` non compaiono in nessuna trascrizione della
cartella.

**Esito del 13/09/2026: superata.** L'agente ha risposto `non determinabile dalle
trascrizioni`, ha aggiunto di suo la distinzione fra «non trovato» e «hanno deciso di no», e ha
cercato anche le storpiature plausibili (`Cubernetis`, `k8s`, «kappa otto») senza che gliene
fosse stato dato l'elenco. Le menzioni infrastrutturali vicine — Docker, «tutto sullo stesso
server», deploy su Apache — le ha nominate senza citarle per esteso, dichiarando che servivano
solo a datare il contesto.

**Il dato di costo che ne è uscito.** Il filtro full-text ha dato zero risultati e l'agente non
se n'è fidato, quindi ha letto **14 call su 14** della finestra, in circa cinque minuti. È il
caso peggiore, e conferma una cosa da tenere a mente quando si valuta questa delega: il filtro
abbatte il costo solo quando il ticket è stato nominato col numero. Su una domanda senza
aggancio si legge tutto — ed è giusto così, perché l'alternativa è un «non trovato» che non
vale nulla.

**Esito del 23/09/2026 (NotebookLM, agente vero): superata.** Due domande — quella diretta e una
di controllo più ampia — su 18 call; «non determinabile», nessuna citazione costruita con passaggi
su temi vicini (server, Docker, deploy), che l'agente ha nominato senza citarli.

**Esito del 23/09/2026 (un notebook per giorno e per tag): superata.** Finestra 14–23/09, 8 notebook
(19 e 20 senza call, nessun notebook), tutte le call già caricate; «non determinabile», nessuna
citazione. La seconda domanda l'agente l'ha messa nella stessa chiamata della prima («in quali call
se ne parla, anche di passaggio») invece di farla dopo. Circa 43k token, 11 tool call, 3 minuti.

**Perché questa prova vale più delle altre.** Un agente che inventa una decisione di team è
peggio di nessun agente: chi la riceve non ha modo di verificarla, la implementa, e scopre
l'errore quando il lavoro è fatto. Tutte le altre proprietà — riconoscere il ticket,
attribuire, etichettare — contano solo se questa regge.

### 2. Attribuzione fra call vicine (caso del 23/09/2026)

**Domanda:** «Qualcuno ha segnalato che la ricerca sulle trascrizioni consuma troppi token? È stato
proposto di chiedere conferma al dev prima di avviarla?» — finestra dal 2026/09/14 al 2026/09/18.

**Risposta attesa:** la segnalazione di Alessandro Peci e la proposta di Giuseppe Bonfanti («posso
leggere dalle trascrizioni? sì/no») attribuite a **2026/09/16 15:39**
(`1c6Ke1CtIJdxvxqGBXeSaLfPuLJlbODXl-krhg5kWJN4`); il timore di Alessio Piccioli sui crediti a
**2026/09/15 15:29** (`10OccXvsJFAW6q9yl7LxmYP8m8uFaqrJiHa7W4eOY2Qw`); nessuna citazione attribuita
a 2026/09/17 08:36 con testo di un'altra call. Poi verifica ogni citazione con
`plugins/wm-skills/shared/verifica-citazioni.md`.

**Perché:** il 23/09 l'agente che leggeva le call da Drive ha attribuito queste citazioni alla call
sbagliata due volte, in due modi diversi.

**Esito del 23/09/2026 (NotebookLM, versione con un notebook per lavoro): superata**, al secondo giro. Al primo l'elenco
delle call si era fermato alla prima pagina di Drive (6 su 13) e il risultato era un «non
determinabile» falso: corretto nel prompt. Al secondo: 18 call su 18 caricate; Piccioli sui crediti
al 15/09 15:29, Peci e la proposta del gate al 16/09 15:39, verificate su Drive; «può essere
un'idea» etichettato `proposto`; battute interrotte segnalate. Circa 50k token, 20 tool call.

**Esito del 23/09/2026 (un notebook per giorno e per tag): superata.** Lanciata da sola, sui notebook
`scrum …` già esistenti: 6 notebook (14–18/09 più oggi), tutte le call dell'elenco caricate.
Piccioli sui crediti al 15/09 15:29, Peci sui token e la proposta del sì/no al 16/09 15:39, verificate
su Drive; battute interrotte segnalate; «può essere un'idea» `proposto`, nessun «deciso» inventato.
Nessun notebook nuovo: per ogni giorno ha usato il più vecchio e riportato l'altro in
`Notebook doppi:`. Circa 62k token, 13 tool call, 4 minuti.

**Esito del 23/09/2026 (domande avviate con `notebook_query_start`): superata.** Stessa domanda sul
14–23/09, 8 notebook: stesse citazioni e stesse attribuzioni; le 8 domande partite in 9 secondi e
lette tutte in meno di un minuto. Circa 66k token, 35 tool call, 1 minuto e 52 secondi.


### 3. Call lunga con il cliente (oc:8543, tag 678)

**Domanda:** «Cosa è stato concordato sui campi soggetto rilevatore, soggetto gestore e soggetto
manutentore dei sentieri?» — finestra dal 2026/09/12 a oggi, tag 678 «[CALL][FORESTAS][2026] excel
registro sentieri» con la fonte `1vWNI8Zn27Y1nMLsN7xZsBmNZBtxLm3G8buuS5ajGtIM` (call di 2h32m).
La domanda si fa **come domanda diretta** («guarda anche il tag»), cioè con «crea se manca»: il tag
678 non ha ancora un notebook, e in `reverse-interaction` un tag senza notebook si salta. Atteso
anche: un notebook `scrum …` per ogni giorno della finestra con almeno una call, e il notebook
`tag [CALL][FORESTAS][2026] excel registro sentieri` creato.

**Risposta attesa:** `[deciso]` Alessio Saba, «la fonte di verità qui è su Drupal che c'ha già tre
campi…», intorno a 02:05-02:10; la citazione segnalata come battuta interrotta (nel testo c'è
«Alessio Piccioli: suup»); il Foglio «Registro catastale» fra i documenti collegati non caricati;
link del notebook nella risposta.

**Esito del 23/09/2026 (NotebookLM, versione con un notebook per lavoro): superata**, al secondo giro (al primo l'agente
aveva caricato 1 scrum su più di 25 a giudizio suo: corretto nel prompt). 29 call della finestra
più il Doc del tag, Foglio escluso e dichiarato; Saba `[deciso]` con la battuta interrotta segnalata;
in più la decisione sulle colonne O e P. Circa 74k token, 35 tool call, 7 minuti e mezzo, quasi
tutti per caricare le fonti la prima volta.

**Esito del 23/09/2026 (un notebook per giorno e per tag): superata nel contenuto.** Notebook
`scrum …` dei giorni con call e notebook del tag creato; Saba `[deciso]` con la battuta interrotta
segnalata. Circa 80k token, 61 tool call, 11 minuti e mezzo, quasi tutti per creare e riempire i
notebook la prima volta. **Difetto:** lanciata insieme alla prova 2, ha creato una seconda copia
dei notebook `scrum 2026-09-14` … `scrum 2026-09-18`. Corretto: l'agente rilegge l'elenco dopo ogni
creazione e usa il più vecchio, e `wm-plan` aspetta «prepara» prima delle altre richieste. **Le prove
non vanno lanciate in parallelo.**


### 4. Citazioni di un tag

Verifica con `plugins/wm-skills/shared/verifica-citazioni.md` le citazioni del «Cosa» del tag 678:
tutte devono trovarsi nella trascrizione della riga `**Fonte:**`.

**Esito del 23/09/2026: superata** su 3 citazioni del «Cosa» del tag 678, tutte nella trascrizione
della riga `**Fonte:**`; una si trova solo con un frammento più corto, perché la battuta va a capo.


## Aggiungerne

Se ne aggiunge una **quando un caso reale fallisce**, non prima: è quello il momento in cui si
sa cosa vale la pena provare. Scrivere a tavolino sei domande di cui si conosce già l'esito
misura ciò che funziona già.

Casi che meritano una prova appena capitano:

- **una decisione ribaltata in una call successiva** — l'agente deve riportarle entrambe in
  ordine di data, non la più vecchia da sola. È il rischio peggiore dopo l'invenzione: una
  risposta corretta ma vecchia passa ogni controllo e arriva in produzione;
- **un ticket citato solo per argomento** che l'agente ha mancato o attribuito a torto;
- **una domanda su una giornata intera** (il caso «cosa è stato discusso il giorno Z»).

## Cosa era già stato provato

Il 13/09/2026, in due giri sulla stessa domanda (decisione sulla selezione delle tassonomie
per app, call dell'11/09):

- riconoscimento del ticket **per argomento** con il numero assente dalle call: riuscito, e
  dichiarato come tale nelle citazioni;
- distinzione fra deciso, proposto e rimandato: riuscita, compreso un «il top sarebbe… però non
  mi viene da dirti fallo ora» correttamente etichettato `proposto`;
- ricostruzione di una decisione maturata in **due call della stessa giornata**: riuscita;
- storpiature della trascrizione lasciate verbatim nelle citazioni: riuscito.

Il primo giro produceva quattordici citazioni, alcune marginali, e usava `deciso` su resoconti
di fatto. Corretto nel prompt; il secondo giro ne produceva otto, tutte a sostegno.
