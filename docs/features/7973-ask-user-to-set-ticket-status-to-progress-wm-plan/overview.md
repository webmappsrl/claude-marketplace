> Ticket: oc:7973

# Ask user to set ticket status to progress when work begins in wm-plan skill

## Cosa cambia
Al termine della Fase 0, dopo aver letto o creato un ticket, la skill chiede all'utente se vuole impostare lo status a "progress" e assegnarsi il ticket. In caso affermativo esegue un PATCH su Orchestrator con `status: progress` e `user_id` dell'utente corrente (ricavato da `GET /api/me`). In caso di errore avvisa e prosegue.

## Perché
Il ticket rimaneva in stato "new" o "defined" per tutta la durata del lavoro, rendendo il Kanban di Orchestrator inaccurato. Il developer doveva ricordarsi di aggiornarlo manualmente.

## Requisiti
- [ ] Al termine della Fase 0 (dopo riepilogo ticket), se esiste un ID ticket, porre una sola domanda: "Vuoi impostare lo status a progress e assegnarti il ticket?"
- [ ] La domanda viene posta sia nel Caso A (ticket letto) che nel Caso B (ticket appena creato e confermato)
- [ ] La domanda NON viene posta se l'utente ha scelto di non creare il ticket (Caso B senza ID)
- [ ] In caso di risposta affermativa, eseguire PATCH con `status: progress` e `user_id` dell'utente corrente
- [ ] L'`user_id` viene ricavato chiamando `GET /api/me` con il token corrente
- [ ] Le credenziali di autenticazione vengono salvate in un unico file JSON `~/.config/webmapp/orchestrator-auth.json` contenente token e user info (id, name, email)
- [ ] La procedura di login nella skill viene aggiornata per chiamare `GET /api/me` dopo il login e salvare il risultato nel file JSON unificato
- [ ] Se il PATCH fallisce (token scaduto, errore di rete), la skill avvisa l'utente con un messaggio e prosegue con le fasi successive senza bloccare

## Rischi
- **Endpoint `GET /api/me` non sempre disponibile:** se l'endpoint non risponde (deploy non eseguito, errore server), la skill non può ricavare l'`user_id`. Mitigazione: gestire il fallback con avviso e skip dell'assegnazione, come per il PATCH.
- **Migrazione file legacy:** chi ha già `~/.config/webmapp/orchestrator-token` non deve eseguire un nuovo login. La skill verifica l'esistenza del file JSON; se assente ma esiste il token legacy, chiama `GET /api/me` con il token esistente per costruire il file JSON unificato automaticamente.

## Out of scope
- Aggiornamento automatico dello status senza conferma dell'utente
- Proposte di status alternativi (es. "assigned", "todo") — solo "progress"
- Gestione multi-utente (assegnazione a un altro developer)

## Moduli toccati
- `plugins/wm-skills/skills/wm-plan/SKILL.md` — aggiunta della domanda in Fase 0, aggiornamento procedura login, aggiornamento path file auth
