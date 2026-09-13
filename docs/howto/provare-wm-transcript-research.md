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
4. la **Copertura** dichiara quante call ha letto su quante?

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

**Perché questa prova vale più delle altre.** Un agente che inventa una decisione di team è
peggio di nessun agente: chi la riceve non ha modo di verificarla, la implementa, e scopre
l'errore quando il lavoro è fatto. Tutte le altre proprietà — riconoscere il ticket,
attribuire, etichettare — contano solo se questa regge.

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
