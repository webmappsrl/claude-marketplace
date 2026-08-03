> Ticket: oc:8341

# Notes — Esecuzione automatica PHPStan pre-PR/merge in wm-plan

## Deviazioni dal piano

Nessuna deviazione: i 3 task sono stati eseguiti esattamente come pianificato in `plan.md`.

## Bug trovati

Nessuno.

## Decisioni

- Nessun bypass PHPStan è stato necessario durante l'implementazione di questa feature stessa
  (il repo `claude-marketplace` non è un progetto Laravel, `has_phpstan_ci` non si applica qui).
- La stima iniziale proposta in Fase: estimation (Misurato 0.5h + Stimato 2.6h = 3.1h) è stata
  rivista dal dev a **Totale: 1h** (Misurato 0.5h + Stimato 0.5h) — accettata senza discussione
  per regola del workflow.
- Durante la Fase: challenge sono emerse e chiuse esplicitamente le seguenti decisioni non ovvie
  (già recepite in `overview.md`):
  - Errori PHPStan preesistenti fuori dal diff → non bloccano, propongono un ticket Orchestrator
    dedicato invece di un blocco a oltranza.
  - Nessun kill-switch/opt-out persistente per repo — il bypass va motivato ed eseguito ad ogni
    singolo blocco.
  - Detection dello step CI: pattern grep permissivo (case-insensitive su `phpstan`), falsi
    positivi accettati, falsi negativi accettati come limite noto.
  - Servizio Docker per l'esecuzione: riusa `$DOCKER_PROJECT_DIR_NAME` già risolto in
    `environment-setup: docker-check`, nessuna euristica di rilevamento servizio aggiuntiva.
  - Timeout fisso a 5 minuti (300s), valore di partenza regolabile in futuro.

## Follow-up

- Nessun falso negativo di classificazione stima da segnalare (entrambi i componenti erano
  "scrittura pura", nessuna sorpresa emersa in esecuzione).
- Follow-up consapevole (accettato, out of scope in questo ciclo): se in futuro serviranno altri
  strumenti di quality-gate (ESLint, Pint, Psalm), valutare un'astrazione "quality-gate
  pluggable" invece di aggiungere ulteriori `if` hardcoded in `review-gate`.
