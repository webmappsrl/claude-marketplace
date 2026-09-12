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
