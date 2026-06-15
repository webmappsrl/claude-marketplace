> Ticket: oc:8068

# Migra e revisiona wm-review-ticket come skill nel wm-marketplace

## Cosa cambia

Viene creata la skill `wm-review-ticket` nel plugin `wm-skills`, che sostituisce e revisiona il comando locale `~/.claude/commands/review-ticket.md`. La skill viene distribuita via marketplace a tutto il team. Viene aggiunto un contratto esplicito tra `wm-plan` e `wm-review-ticket` sugli artefatti prodotti e consumati.

## Perché

Il team Webmapp ha due casi d'uso principali per la code review:
1. Un collega assegna un ticket/PR da rivedere in autonomia
2. Al termine di una feature lavorata con `wm-plan`, si vuole una review formale prima del merge

Il comando locale attuale non è distribuito, non sfrutta gli artefatti prodotti da `wm-plan` (`overview.md`, `plan.md`, `notes.md`) come contesto di intent, e non ha un contratto esplicito con il workflow di sviluppo.

## Requisiti

- [ ] Input primario: `oc:<ID>` — la skill legge il ticket da Orchestrator per estrarre titolo, `customer_request`, `description`
- [ ] Se nella `description` del ticket è presente un URL PR GitHub, la skill fa il checkout del branch della PR nel repo corretto
- [ ] Il repo di destinazione viene determinato leggendo il `CLAUDE.md` del progetto corrente (stesso approccio di `wm-plan`), non da una mappa hardcoded
- [ ] Se esistono `docs/features/<feature-slug>/overview.md` e `plan.md`, vengono letti come "spec di intent" prima della review del codice
- [ ] `notes.md` viene letto come changelog delle deviazioni: il reviewer ne tiene conto come contesto ma rimane libero di sollevare dubbi su deviazioni rischiose
- [ ] Obiettivo della review: (1) il codice risponde correttamente alla richiesta del ticket? (2) introduce side effect non gestiti o bug?
- [ ] Checkout del branch PR: tentare prima localmente, poi `git fetch origin <branch>` se non esiste, poi cercare il commit di merge via `git log --all --grep="<branch-name>"`; se nulla funziona chiedere all'utente
- [ ] Prima del checkout: se il working tree è dirty, eseguire `git stash` automaticamente con messaggio esplicito all'utente; al termine della review eseguire `git stash pop`; se il pop fallisce avvisare l'utente e lasciare lo stash intatto
- [ ] Se nella `description` del ticket **non è presente** un URL PR GitHub, la skill chiede all'utente di fornire: URL PR, nome branch, o hash commit su cui fare la review
- [ ] Graceful degradation: se i docs `wm-plan` non esistono, la review procede sul diff senza contesto di intent
- [ ] `wm-plan` è la fonte autoritativa del contratto artefatti: documenta in una sezione dedicata cosa produce (`docs/features/<slug>/overview.md`, `plan.md`, `notes.md`) e il significato di ciascun file
- [ ] `wm-review-ticket` non duplica il contratto: in apertura fa `WebFetch` su `https://raw.githubusercontent.com/webmappsrl/claude-marketplace/main/plugins/wm-skills/skills/wm-plan/SKILL.md` per leggere la sezione artefatti — funziona anche se `wm-plan` non è installato localmente
- [ ] `wm-plan` Fase 6d cita esplicitamente `wm-review-ticket` come skill da invocare per la review formale
- [ ] Finder paralleli sul diff: correctness (risposta alla richiesta + side effect), cleanup, altitude
- [ ] Output in italiano: verdetto complessivo, finding ranked con file:linea, separazione bloccanti/cleanup
- [ ] Al termine della review, aggiornare il ticket Orchestrator originale:
  - `description`: aggiungere riepilogo tecnico della review (finding, decisioni, esito)
  - Se nessun bloccante: chiedere all'utente quale status impostare tra `testing`, `released`, `done` (mostrare lista da `StoryStatus.php`)
  - Se ci sono bloccanti: impostare status a `todo` (con conferma utente)

## Rischi

- **Checkout automatico del branch PR**: se il repo locale ha modifiche non committate, la skill esegue `git stash` automaticamente prima del checkout (con messaggio esplicito all'utente) e `git stash pop` al termine della review. Se lo stash pop fallisce (conflitti), la skill avvisa l'utente e lascia lo stash intatto per risoluzione manuale.
- **Slug non trovato**: la skill deve ricavare il `<feature-slug>` dal ticket ID per cercare i docs — se la cartella `docs/features/` ha uno slug diverso dal pattern atteso, i docs non vengono trovati. Mitigazione: fare una ricerca fuzzy per ID (`find docs/features/ -name "*.md" | grep <ID>`) prima di rinunciare.
- **Repo non riconosciuto**: il `CLAUDE.md` del progetto potrebbe non listare tutti i repo coinvolti. Mitigazione: chiedere all'utente il percorso locale se il repo della PR non è riconosciuto.

## Out of scope

- Input diretto con hash commit o URL PR senza ticket (potrebbe essere aggiunto in futuro)
- Effort configurabile (low/medium/high) — per ora sempre al massimo, revisione in ticket separato
- Creazione di ticket di esito separati — l'esito viene scritto direttamente sul ticket originale

## Moduli toccati

| File | Repo | Azione |
|------|------|--------|
| `plugins/wm-skills/skills/wm-review-ticket/SKILL.md` | claude-marketplace | Creare |
| `plugins/wm-skills/skills/wm-plan/SKILL.md` | claude-marketplace | Modificare (aggiungere riferimento a `wm-review-ticket` in Fase 6d e contratto artefatti) |
