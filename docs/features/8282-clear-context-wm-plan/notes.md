> Ticket: oc:8282

# Notes — Clear del context in wm-plan dopo reverse-interaction e dopo implementation

## Deviazioni dal piano

Nessuna deviazione rilevante. I 4 task sono stati eseguiti esattamente come pianificato in `plan.md`.

## Bug trovati

Nessuno.

## Decisioni

- **Pivot da "context clear" a "isolamento via subagente"**: la richiesta iniziale era di introdurre un clear del context principale dopo `reverse-interaction` e dopo `execution: implementation`. Durante la Fase: reverse-interaction è emerso che il vero obiettivo non era l'economia di context ma l'indipendenza di giudizio — e che un clear del context principale su `overview.md` è controproducente perché chi ha condotto il dialogo socratico ha un vantaggio informativo per sintetizzarlo. Si è quindi deciso di non introdurre nessun clear reale, e di estendere invece il pattern già esistente in `challenge: subagent` (subagente cieco) alla fase `execution: review-gate`.
- **Requisito aggiuntivo emerso durante estimation**: durante la stima è emerso che in questa stessa sessione non era stato registrato `planning_start_at` come previsto da oc:8278 — non un bug del `SKILL.md` (l'istruzione esiste già ed è chiara), ma un caso concreto di mancata applicazione. Aggiunto un requisito per rendere il gap visibile invece che silenzioso: se il timestamp manca, `Fase: estimation` lo segnala esplicitamente invece di calcolare una quota "misurata" fittizia.
- **Verificato che `subagent-driven-development` (skill upstream Superpowers) già risolve un problema adiacente**: durante la challenge è stato ipotizzato un problema di coordinamento tra sottoagenti paralleli e di "self-report" non indipendente. Lettura diretta di `subagent-driven-development/SKILL.md` ha confermato che quella skill già impone review isolate per-task (nessun self-report) ed esecuzione sequenziale (non parallela) con ledger di progresso — quindi nessuna modifica necessaria lì. Questo ha ristretto correttamente lo scope di oc:8282 al solo `review-gate` di `wm-plan`.

## Follow-up

- Nessun follow-up di codice. Possibile follow-up di processo: se in futuro emergono altri gap di applicazione delle istruzioni esistenti (come il timestamp mancante in questa sessione), valutare se serva un meccanismo più robusto di quello testuale (es. un check strutturale) — ma è fuori scope per questo ciclo.
