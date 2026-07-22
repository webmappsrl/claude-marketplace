> Ticket: oc:8278

# Notes — Rivedere criteri di stima ore in wm-plan

## Deviazioni dal piano

Nessuna deviazione tecnica. Un'imprecisione minore nel piano stesso: il Task 1 prevedeva "due occorrenze" del marcatore `planning_start_at` nella sezione `Fase: ticket` dopo la modifica, ma il testo scritto ne contiene una sola (la seconda occorrenza attesa era in realtà nel riferimento della sezione `estimation: analisi`, aggiunta nel Task 2). Il contenuto sostanziale (cattura del timestamp) è comunque presente e corretto; non ha richiesto modifiche al piano.

## Bug trovati

Nessuno.

## Decisioni

- Le 6 modifiche a `SKILL.md` sono state applicate come edit sequenziali sullo stesso file di lavoro, quindi i commit non replicano uno-a-uno i "Commit" testuali indicati in ciascun task del piano (che presupponevano staging incrementale). Si è optato per un commit unico che copre l'intera riscrittura della `Fase: estimation` e delle sezioni collegate, più commit separati per la documentazione (`docs/features/`) e per l'aggiornamento di `CLAUDE.md`.
- Il coefficiente di velocità per-dev, discusso durante la Fase: reverse-interaction, è stato esplicitamente escluso da questo ciclo (vedi overview.md → Out of scope) per mancanza di un endpoint Orchestrator di aggregazione storica e per campione insufficiente per dev.

## Follow-up

- Valutare in un ciclo successivo l'introduzione di un coefficiente di velocità per-dev, quando sarà disponibile più storico e un modo per interrogarlo da Orchestrator (oggi solo GET singolo ticket o tag→stories).
- Dopo un numero sufficiente di ticket Feature stimati col nuovo criterio (`[stima v2 — per-componente]`), ripetere l'analisi stimato-vs-effettivo fatta in Fase: reverse-interaction per validare se il bias di overstima è stato effettivamente ridotto.
