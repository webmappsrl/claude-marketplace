> Ticket: oc:8530

# Notes — wm-transcript-research

## Deviazioni dal piano

**Il file di prove è stato ridotto da sei domande a una.** Il piano (Task 4) prescriveva cinque
o sei domande con risposta nota, una per grado di riconoscimento. Dopo i due giri di prova è
emerso che quattro di quelle sei avrebbero misurato proprietà già dimostrate funzionanti nella
stessa sessione — riconoscimento per numero, per argomento, attribuzione, etichette. È rimasta
la prova che valeva il costo: quella la cui risposta **non esiste**, perché verifica la sola
proprietà su cui si gioca la fiducia nell'agente. Il file dice esplicitamente che se ne
aggiungono quando un caso reale fallisce, che è il momento in cui si sa cosa provare.

La seconda prova prevista — una decisione ribaltata in una call successiva — non è stata
scritta perché serviva un caso reale che il dev ricordasse, e non ne è emerso uno. Resta in
`## Follow-up`.

**Due righe nel `CLAUDE.md` invece di una.** Il piano prevedeva solo la voce in
`## Conoscenza`; è stata aggiunta anche quella in `## Procedure`, perché il file delle prove è
una procedura e senza quella riga nessuno lo troverebbe.

**Il limite di documenti apribili non è stato fissato.** L'overview lo elencava fra i rischi
(«va deciso quanti l'agente può aprirne prima di dichiarare che la domanda è troppo larga») e
non è stato deciso: il prompt non ha alcun tetto, né sul numero di call né sull'output. È
coerente con la scelta «nessun tetto», ma quel rischio resta senza mitigazione scritta. La
prova del 13/09 dà l'ordine di grandezza del caso peggiore: 14 call, cinque minuti.

**Aggiornati due file che il piano non nominava**, per coerenza interna:

- `docs/knowledge/wm-skills-delega-agentica.md` — la sua tabella «Agente | Fase | Cosa
  restituisce» sarebbe rimasta incompleta. La pagina non afferma il tetto come universale
  (rimanda al file condiviso), quindi su quel punto non c'era contraddizione da sanare.
- `plugins/wm-skills/shared/agent-delegation.md`, riga sugli agenti il cui esito è un giudizio
  e ammette revisione: il nuovo agente vi rientra, e lasciarlo fuori avrebbe creato
  un'incoerenza dentro lo stesso file.

**Aggiornato il diagramma di flusso**, cosa che il piano prevedeva come possibile ma non certa.
Il controllo in CI passava — confronta i nomi delle fasi, e nessuna fase è stata aggiunta o
rinominata — ma il nodo e il paragrafo descrivevano `reverse-interaction` come se avesse una
sola fonte. È il limite che lo script dichiara di avere, e la regola path-scoped attribuisce
quel giudizio a chi modifica la skill. Toccato solo il contenuto: due righe, template e CSS
intatti.

## Bug trovati

Nessuno: non c'è codice eseguibile in questo lavoro.

## Decisioni

**Rischio del contenuto sensibile accettato dal dev (13/09/2026).** La Challenge aveva
classificato come bloccante la possibilità che citazioni verbatim di call — dove si parla di
clienti e di compromissioni di siti — finiscano in artefatti committati, anche su repo cliente.
Era stato proposto un requisito che vietasse le citazioni verbatim negli artefatti. Il dev ha
deciso di non introdurlo: il controllo esiste già ed è l'approvazione umana di ogni artefatto,
più il divieto di commit dentro il lavoro. Responsabilità sua, dichiarata.

**Accesso alle trascrizioni per gli altri dev: affermato dal dev, non verificato dall'API.**
Il controllo dei permessi su cartella e Doc mostra un solo permesso esplicito, il proprietario.
Il dev ha assicurato che tutti i dev accedono, presumibilmente dal link ricevuto per posta. La
differenza che resta aperta è fra *aprire* un Doc ed *elencare* una cartella, che in Drive sono
permessi distinti: l'agente parte dall'elenco. Da provare da un secondo account.

**Stima ridotta dal dev da 14,5h a 3h.** `wm-estimate` aveva stimato 13,5h di implementazione
più 1h di pianificazione misurata. Il dev ha giudicato la stima sovradimensionata per un lavoro
di soli file Markdown su pattern già esistenti, e ha fissato 2h di implementazione, totale 3h.
Da verificare a consuntivo: il lavoro effettivo si è svolto in una sessione sola.

**La verifica a macchina delle citazioni è stata esclusa per progetto.** Su codice,
`wm-codebase-research` impone di ricontrollare ogni prova con `sed -n`. Qui non esiste
l'equivalente: ricontrollare significherebbe rileggere la call nel context principale, cioè
riportare dentro il materiale che la delega doveva tenere fuori. La verifica è quindi umana, e
per renderla possibile ogni risposta porta le `Fonti` con l'**id del file Drive**.

**L'ID del ticket nel parlato: ipotesi iniziale smentita dai dati.** Si era assunto che i
numeri venissero detti cifra per cifra e fossero quindi irrintracciabili. La lettura delle
trascrizioni ha mostrato il contrario — sempre cifre attaccate — e ha promosso il filtro per
numero da strada da escludere a primo grado di riconoscimento. È la ragione per cui l'ispezione
della fonte è stata fatta prima di scrivere i requisiti e non dopo.

## Cosa ha mostrato la prova

Due giri sulla stessa domanda (decisione sulla selezione delle tassonomie per app, call
dell'11/09), più la prova sulla domanda senza risposta.

Riuscito al primo giro: riconoscimento del ticket per solo argomento con il numero mai
pronunciato; distinzione fra deciso, proposto e rimandato; ricostruzione di una decisione
maturata in due call della stessa giornata; storpiature lasciate verbatim.

Corretto dopo il primo giro: quattordici citazioni, alcune marginali, e `deciso` usato su
resoconti di fatto. Aggiunta al prompt la sezione `Fonti` con gli id — richiesta del dev — e la
regola «cita ciò che regge la risposta, non tutto ciò che è pertinente». Il secondo giro ne ha
prodotte otto, tutte a sostegno.

Sulla domanda senza risposta l'agente ha risposto `non determinabile`, ha distinto di sua
iniziativa «non trovato» da «hanno deciso di no», e ha cercato anche le storpiature plausibili
di Kubernetes senza che gliene fosse dato l'elenco.

**Il dato di costo che ne è uscito**, e che vale più della prova stessa: su una domanda senza
aggancio il filtro full-text dà zero risultati, l'agente non se ne fida e legge **tutte** le
call della finestra — 14 su 14, in circa cinque minuti. Conferma il rilievo della Challenge:
il filtro abbatte il costo solo nel caso in cui il ticket è stato nominato col numero, cioè
quello già facile.

## Follow-up

- **Provare l'elenco per cartella da un secondo account.** Se fallisce, non è un dettaglio:
  salta la selezione per finestra di date, su cui poggiano tutti e tre i casi d'uso.
- **Aggiungere la prova sulla decisione ribaltata**, appena capita un caso reale. È il rischio
  peggiore dopo l'invenzione: una risposta corretta ma vecchia passa ogni controllo.
- **Il caso «cosa è stato discusso il giorno Z» non è stato provato.** È scritto nel prompt ma
  nessuna prova lo esercita, ed è quello con l'output più voluminoso.
- **La ricerca verbosa non è stata provata.** Esiste come modalità dichiarata, ma nessuno dei
  tre giri l'ha usata.
- **`wm-review-ticket` e `wm-tag` non sono stati toccati.** Entrambi potrebbero trarre
  beneficio dalla stessa fonte — una review che sa cosa era stato deciso in call giudica
  diversamente — ma è lavoro a sé.
- **Gli ID cresceranno a cinque cifre.** Il prompt lo dice, nessuna prova lo esercita.
