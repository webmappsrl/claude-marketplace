> Ticket: oc:8531

# Piano

**Goal:** correggere i tre difetti che hanno prodotto lo scarto di oc:8530.

**Spec:** [overview.md](overview.md)

Lavoro di sola modifica testuale su tre file, senza task paralleli: il piano è la sequenza
seguita, registrata a consuntivo perché il contratto degli artefatti lo richiede.

### Task 1 — `wm-estimate.md`

- [x] dichiarare in apertura che chi esegue è un LLM, citando oc:8530 come errore da non rifare
- [x] aggiungere la sezione di classificazione del deliverable e l'uso dell'ancora
- [x] riformulare il punto A: non «il tempo che impiegheresti tu a scrivere codice e test» ma il
      tempo di una sessione in cui il modello produce e il dev rilegge
- [x] vietare il buffer di novità su ciò che l'overview già accerta
- [x] aggiornare il formato di risposta con le due righe di intestazione
- [x] alzare il tetto da 25 a 30 righe, che le nuove righe richiedono
- [x] `claude plugin validate .`

### Task 2 — `wm-plan/SKILL.md`

- [x] passare la durata misurata all'agente
- [x] scrivere perché questo non rompe la cecità, altrimenti la modifica sembra una deroga

### Task 3 — `docs/knowledge/wm-plan-stima-ore.md`

- [x] riscrivere lo stato attuale
- [x] spostare in «Come ci siamo arrivati» ciò che è stato superato, **con il motivo**
- [x] sanare la contraddizione sul marcatore di versione
- [x] annotare che il campione è uno

### Task 4 — chiusura

- [x] gate, artefatti, riepilogo isolato del diff, commit su approvazione del dev
