> Ticket: oc:8531

# La stima ancorata alla pianificazione misurata e al tipo di deliverable

## Cosa cambia

`wm-estimate` riceve la **durata misurata della pianificazione** e la usa come ancora per
giudicare il proprio totale; **classifica il deliverable** prima di stimare; e non conta nel
buffer di novità ciò che la pianificazione ha già risolto.

## Perché

Su oc:8530 l'agente ha stimato 13,5h di implementazione contro 0,8h reali: un errore di
diciassette volte. Non è imprecisione di giudizio ma un difetto ripetibile, che si presenterà
identico su ogni lavoro che non sia codice applicativo.

Le tre cause, tutte leggibili nel prompt:

1. **Ore-uomo travestite da ore-LLM.** Il prompt chiedeva «il tempo che impiegheresti *tu* a
   scrivere codice e test». Due presupposti sbagliati in una riga: che a eseguire sia una
   persona, e che il deliverable sia codice.
2. **Doppio conteggio del buffer di novità.** 2h per «nessun precedente per il connettore
   Drive», quando quell'esplorazione era già avvenuta in pianificazione ed era già dentro l'ora
   misurata.
3. **Il deliverable non entrava nel calcolo.** Il ciclo build-test-debug è ciò che rende costoso
   il software, e in un file Markdown non esiste.

## Requisiti

- [x] `wm-estimate` riceve la durata misurata e la dichiara in testa alla risposta, con il
      rapporto fra il proprio totale e quell'ancora
- [x] Il prompt dichiara esplicitamente che chi esegue è un LLM, non una persona
- [x] `wm-estimate` classifica il deliverable in *testo* / *codice senza test* / *codice con
      test* e lo dichiara
- [x] Il buffer di novità non copre ciò che l'overview già accerta
- [x] `wm-plan` passa la durata misurata all'agente, e la pagina di conoscenza spiega perché
      questo non rompe la cecità
- [x] Risolta la contraddizione sul marcatore di versione fra la pagina di conoscenza (lo
      prescriveva) e `wm-plan/SKILL.md` (lo vieta)
- [x] `claude plugin validate .` passa

## Rischi

- **Il campione è uno.** Un caso non stabilisce un coefficiente. I criteri sono scritti come
  criteri e non come moltiplicatori tarati su oc:8530, ma resta il rischio che vengano letti
  come una regola quantitativa.
- **L'ancora può diventare una scusa.** Un agente che deve giustificare uno scarto dall'ancora
  può imparare ad ancorarsi sempre, smettendo di stimare. Il prompt dice che discostarsi è
  ammesso purché motivato: se all'uso le stime diventassero tutte «circa quanto la
  pianificazione», sarebbe il segnale che l'ancora ha sostituito il giudizio.
- **Si è toccata la cecità di un agente cieco.** La distinzione fra «non sapere come si è
  arrivati all'overview» e «non sapere quanto è durata» regge, ma è una porta socchiusa: la
  prossima richiesta di passargli «solo un altro dato» va guardata con sospetto.

## Out of scope

- Nessuna analisi retrospettiva sui ticket storici di Orchestrator: servirebbe l'endpoint di
  aggregazione stimato-contro-effettivo, che non esiste (già registrato in oc:8278).
- Nessun coefficiente per dev.

## Moduli toccati

| File | Cosa |
|---|---|
| `plugins/wm-skills/agents/wm-estimate.md` | ancora, classificazione del deliverable, divieto di doppio conteggio |
| `plugins/wm-skills/skills/wm-plan/SKILL.md` | passa la durata misurata e spiega perché non rompe la cecità |
| `docs/knowledge/wm-plan-stima-ore.md` | stato attuale riscritto, il superato spostato nella storia |
