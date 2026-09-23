# wm-review-ticket

## Stato attuale

`wm-review-ticket` esegue la code review di un ticket Orchestrator su qualsiasi repo Webmapp.

**Non duplica il contratto degli artefatti `docs/features/<slug>/`**: lo legge a runtime da
`wm-plan/SKILL.md` su GitHub raw. `wm-plan` resta la fonte autoritativa, quindi un cambio della
struttura non può produrre drift fra le due skill.

Se il working tree è sporco, la skill fa `git stash` prima del checkout e `git stash pop` al
termine: nessun rischio di perdere lavoro in corso.

## Come ci siamo arrivati

- **Copiare la struttura degli artefatti dentro `wm-review-ticket` è stato scartato** (oc:8068):
  due copie divergono alla prima modifica, e il sintomo sarebbe una review che cerca file che
  `wm-plan` non scrive più.
- **L'esito in un ticket Bug nuovo è stato scartato**: chi corregge riprende il ticket rivisto con
  `wm-plan`, che legge la sua `description` come base. Un esito scritto altrove non lo vede
  nessuno. Per questo l'esito va in testa alla `description`, un ciclo sopra l'altro, con i punti
  dei cicli precedenti etichettati uno per uno (risolto, in parte, superato, da togliere).
- **Sulla PR va solo il riepilogo, il dettaglio sta nel ticket** (review di oc:8543): lo stesso
  contenuto scritto in due posti diverge al primo aggiornamento, e chi legge la copia vecchia
  corregge la cosa sbagliata.
