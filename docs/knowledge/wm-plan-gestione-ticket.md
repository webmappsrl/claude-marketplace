# Come wm-plan tratta i ticket Orchestrator

## Stato attuale

In `Fase: ticket`, `wm-plan` legge il ticket, si assegna l'utente autenticato e chiede al dev di
metterlo in progress. L'utente corrente si ricava da `GET /api/me`, non dal `creator_id` del
ticket: così l'assegnazione è sempre a chi sta lavorando, anche su ticket aperti da altri. Le
credenziali stanno in un unico `orchestrator-auth.json`, che porta anche `user_id`, `name` ed
`email` — quanto serve al PATCH senza chiamate aggiuntive.

**Un ticket Help desk con più richieste distinte viene diviso**: l'originale resta ridotto alla
prima richiesta, le successive diventano ticket separati con il `creator_id` replicato dal
cliente originale, ed è possibile un tag di raggruppamento per uso interno dei dev.

`notes.md` è il **registro delle decisioni a posteriori**: le modifiche chieste dopo
l'approvazione del piano si scrivono lì, nella sezione "Decisioni", non riscrivendo il piano.

**Nessun valore di `type` o `status` è scritto nelle skill**: si leggono dagli enum su GitHub
(`StoryType.php`, `StoryStatus.php`).

## Come ci siamo arrivati

- **Il file di auth era un token in chiaro**: è diventato un JSON unico perché servivano anche
  i dati dell'utente, e tenerli in due posti avrebbe richiesto una chiamata in più ad ogni
  assegnazione (oc:7973).
- **Il tipo `Task` è stato rimosso dalle skill**: non esiste su Orchestrator, e una POST
  `/api/stories` con `"type": "Task"` falliva con `422 — Il valore selezionato per type non è
  valido`. Da lì la scelta di rimandare agli enum invece di elencare i valori: un tipo aggiunto
  o rinominato su Orchestrator non fa invecchiare le skill in silenzio.
