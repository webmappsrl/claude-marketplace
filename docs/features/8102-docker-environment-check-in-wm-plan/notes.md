> Ticket: oc:8102

# Notes — Docker environment check in wm-plan

## Deviazioni dal piano
- Il piano originale prevedeva solo il docker-check. Durante la Fase: reverse-interaction è emerso che la numerazione delle fasi era un problema strutturale — il ticket è stato ampliato per includere la migrazione a slug e la riorganizzazione delle sottofasi.

## Bug trovati
Nessuno.

## Decisioni
- **Slug invece di numeri**: proposta dall'utente durante il dialogo. Motivazione: ogni inserimento futuro richiede rinumerazione di header e riferimenti interni.
- **`init-context` mantenuta come fase separata**: proposta di assorbire `init-context` in `environment-setup` respinta — Claude legge `CLAUDE.md` come primo atto di comprensione del progetto, è semanticamente distinto dal rilevamento tecnico dell'ambiente.
- **`domain-mapping` e `ux-ui-detection` spostate in `environment-setup`**: sono rilevamento di ambiente, non dialogo con l'utente — appartengono prima della `reverse-interaction`.
- **`execution: formal-review` aggiunta come sottofase esplicita**: rendere visibile il percorso di review formale invece di nasconderlo in un hint testuale.
- **docker-check FAIL-SOFT sempre**: la fase non deve mai bloccare il workflow — confermato dall'utente durante la Challenge.
- **`docker compose stop` mai `down`**: vincolo esplicito per non cancellare i container degli altri progetti.

## Follow-up
Nessuno.
