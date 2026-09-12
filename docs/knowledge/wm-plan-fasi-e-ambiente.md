# Fasi e rilevamento ambiente in wm-plan

## Stato attuale

Le fasi di `wm-plan` sono identificate da **slug**, non da numeri (`## Fase: ticket`,
`## Fase: environment-setup`, …): inserire una fase nuova non richiede mai di rinumerare le
altre — e i nomi delle fasi sono anche ciò che `.github/scripts/verifica-diagramma.sh`
confronta con i nodi del diagramma.

`Fase: environment-setup` **centralizza il rilevamento dell'ambiente** — `project-detection`,
`domain-mapping`, `ux-ui-detection`, `docker-check` — e gira prima di `init-context` e
`reverse-interaction`.

`docker-check` è **fail-soft**: qualsiasi errore produce un `⚠️` e il lavoro prosegue, non
blocca mai. Ferma i container con `docker compose stop`, mai `down` né `rm`.

## Come ci siamo arrivati

- **`init-context` è rimasta separata dal rilevamento tecnico** (oc:8102): leggere il
  `CLAUDE.md` del repo è il primo atto di comprensione del progetto, semanticamente distinto dal
  rilevare stack e container.
- **La numerazione delle fasi è caduta** (oc:8102): ogni inserimento a metà workflow costringeva
  a toccare tutte le fasi successive e i riferimenti incrociati nelle altre skill.
