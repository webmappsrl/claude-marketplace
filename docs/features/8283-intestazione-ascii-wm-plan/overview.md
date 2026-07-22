> Ticket: oc:8283

# Intestazione ASCII per skill wm-plan con info versione/aggiornamenti e diagramma di flusso

## Cosa cambia

`wm-plan` mostra, solo alla prima invocazione della skill in una sessione, un'intestazione con:
- Banner ASCII art con il testo "wm-plan"
- Data dell'ultimo commit che ha modificato `SKILL.md`
- Segnalazione se esiste una versione più recente della skill su GitHub (confronto hash commit locale vs remoto)
- Link a un Artifact Mermaid che mostra il diagramma di flusso del workflow `wm-plan`

Inoltre, il `CLAUDE.md` di questo repo (`claude-marketplace`) viene aggiornato con un'istruzione permanente: ogni volta che Claude Code modifica file di questo repo durante una sessione, deve rigenerare/aggiornare l'Artifact del diagramma di flusso (redeploy sullo stesso URL, non un nuovo Artifact) prima di considerare il lavoro concluso.

## Perché

Dare visibilità immediata su stato e freschezza della skill (quando è stata aggiornata l'ultima volta, se c'è una versione più recente da scaricare) e un modo visivo rapido per capire il workflow di `wm-plan` senza dover rileggere l'intero `SKILL.md`.

## Requisiti

- [ ] Banner ASCII art "wm-plan" mostrato come primo output della skill, solo alla prima invocazione per sessione (non ripetuto se la skill viene richiamata più volte nella stessa conversazione)
- [ ] Data dell'ultimo commit che ha modificato `plugins/wm-skills/skills/wm-plan/SKILL.md`, recuperata via query API GitHub (o `git log` se il repo marketplace è disponibile localmente)
- [ ] Check "aggiornamento disponibile": confronto tra l'hash del commit HEAD locale del repo marketplace installato e l'hash dell'ultimo commit remoto su GitHub via API — nessuno stato aggiuntivo salvato, il repo git locale è sempre la fonte di verità
- [ ] **Modalità sviluppo locale:** se il repo locale ha modifiche non committate su `SKILL.md` o è su un branch diverso da `main`, l'header mostra esplicitamente un'indicazione tipo "🔧 modalità sviluppo locale — check versione saltato" invece del check normale
- [ ] Link all'Artifact del diagramma di flusso mostrato nell'header, letto da una sezione dedicata del `CLAUDE.md` di questo repo (es. `## Diagramma di flusso wm-plan`)
- [ ] Se il link non è presente in `CLAUDE.md` (Artifact non ancora pubblicato), l'header lo segnala senza bloccare l'esecuzione della skill (fail-soft)
- [ ] Nuova istruzione permanente in `CLAUDE.md` di questo repo: dopo qualsiasi modifica ai file del repo in una sessione, rigenerare/aggiornare (redeploy, stesso URL) l'Artifact del diagramma di flusso di `wm-plan`, prima di concludere il lavoro — nessuna restrizione ai soli file di `wm-plan`
- [ ] La pubblicazione/aggiornamento dell'Artifact avviene per tentativo diretto (nessun check preventivo di account): se lo strumento Artifact restituisce un errore in fase di redeploy sullo stesso URL, avvisare l'utente che potrebbe essere necessario switchare all'account Claude del team Webmapp, senza bloccare il resto della sessione
- [ ] URL dell'Artifact salvato/aggiornato nel `CLAUDE.md` del repo alla prima pubblicazione riuscita, così resta versionato e condiviso con tutto il team via git

## Rischi

- **Drift documentale silenzioso** (Fase: challenge, asse Worst case): l'Artifact potrebbe rappresentare uno stato del workflow ormai superato se la rigenerazione fallisce silenziosamente o viene dimenticata in una sessione, senza alcun meccanismo automatico di rilevamento. **Rischio accettato consapevolmente dall'utente**, che si assume la responsabilità di notare e ripristinare la coerenza manualmente — nessun requisito di verifica automatica aggiunto.
- **Trigger di rigenerazione non ristretto**: l'istruzione in `CLAUDE.md` scatta su qualsiasi modifica al repo, anche quando `wm-plan` non è toccato — decisione esplicita dell'utente, accettato lo spreco occasionale di redeploy non necessari.
- **Rollback dell'istruzione in `CLAUDE.md`** non è istantaneo tra sessioni con context già caricato — limite intrinseco del meccanismo di istruzioni persistenti, non specifico di questa feature. Nessuna mitigazione richiesta.

## Out of scope

- Coefficiente o meccanismo di autenticazione dedicato per distinguere account Claude multipli con stessa email — si usa il pattern tentativo/fallimento
- Rigenerazione automatica del diagramma per skill diverse da `wm-plan` (`wm-tag`, `wm-review-ticket` restano fuori da questo ciclo)
- Notifiche push o meccanismi di alert per aggiornamenti disponibili — solo segnalazione testuale nell'header
- Verifica automatica di sincronia tra Artifact pubblicato e stato corrente di `SKILL.md` (hash di pubblicazione) — rischio di drift accettato manualmente

## Moduli toccati

- `plugins/wm-skills/skills/wm-plan/SKILL.md` — logica header, banner ASCII, check versione, link artifact
- `CLAUDE.md` (repo `claude-marketplace`) — nuova sezione con URL Artifact e istruzione di rigenerazione post-modifica
