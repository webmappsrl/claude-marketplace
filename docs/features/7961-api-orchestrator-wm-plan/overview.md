> Ticket: oc:7961

# Integrazione API Orchestrator in wm-plan

## Cosa cambia
`wm-plan` legge e aggiorna i ticket Orchestrator via REST API invece di chiedere all'utente di incollare il contenuto manualmente.

## Perché
Ridurre il friction del workflow: l'utente fornisce solo l'ID del ticket, la skill recupera i dati autonomamente e aggiorna Orchestrator a fine lavoro senza uscire da Claude Code.

## Requisiti
- [ ] Fase 0: dato `oc:<ID>`, legge il ticket via GET `/api/stories/{id}` senza conferma
- [ ] Autenticazione: login automatico se token assente/scaduto, token salvato in `~/.config/webmapp/orchestrator-token`
- [ ] `ORCHESTRATOR_URL` configurabile via env (default: `https://orchestrator.maphub.it`)
- [ ] Status list letta dinamicamente da `app/Enums/StoryStatus.php` su GitHub
- [ ] Campi accettati letti dinamicamente da `app/Http/Requests/Api/StoryApiRequest.php` su GitHub
- [ ] Caso B: se non esiste un ticket, la skill propone il testo e lo crea via POST dopo conferma
- [ ] Aggiornamenti espliciti su richiesta dell'utente durante il workflow (con conferma)
- [ ] Fine workflow: suggerisce status, mostra bozza `description` e `customer_request`, richiede approvazione esplicita prima del PATCH

## Rischi
- Token Sanctum scaduto: gestito con re-login automatico al 401
- Scrittura su ticket sbagliato: mitigata dalla conferma esplicita obbligatoria con preview dei campi
- `customer_request` letta dal cliente: richiede revisione attenta prima dell'invio

## Out of scope
- Gestione multi-account / più token
- Aggiornamenti di status intermedi automatici (solo su richiesta esplicita)

## Moduli toccati
- `plugins/wm-skills/skills/wm-plan/SKILL.md` — unico file modificato
