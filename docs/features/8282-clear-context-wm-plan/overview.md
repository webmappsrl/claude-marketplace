> Ticket: oc:8282

# Clear del context in wm-plan dopo reverse-interaction e dopo implementation

## Cosa cambia

`wm-plan` isola sempre, tramite subagente, la **Fase: execution → review-gate**, con lo stesso pattern già usato in `challenge: subagent`. Il subagente riceve solo il percorso del repo/branch e le istruzioni di analisi — nessun riassunto della conversazione precedente — esegue `git diff --stat` + `git diff` autonomamente e produce il riepilogo strutturato (file creati/modificati/eliminati, descrizione dei file significativi) da presentare all'utente per l'approvazione.

La **Fase: overview** resta invariata: continua a essere scritta dal context principale, subito dopo `reverse-interaction`, senza alcun subagente e senza clear — perché in quel punto del workflow il context è ancora "leggero" (solo dati ticket + dialogo di reverse-interaction) e chi ha condotto il dialogo ha un vantaggio informativo che un subagente isolato non potrebbe recuperare da un riassunto di seconda mano.

Non viene introdotto nessun meccanismo di "context clear" vero e proprio sul context principale: l'isolamento si ottiene esclusivamente delegando a subagenti nei punti dove serve un giudizio non contaminato dal ragionamento precedente (challenge, ora anche review-gate), non scartando stato dal context principale.

## Perché

La motivazione primaria **non è l'economia di context, ma l'indipendenza del giudizio**: un valutatore che ha vissuto tutto il ragionamento che ha portato al codice (le scelte fatte, i compromessi, le difficoltà superate durante l'implementazione) tende a confermare le proprie decisioni piuttosto che metterle in discussione — lo stesso motivo per cui `challenge` non riceve alcun riassunto della conversazione e legge `overview.md` a freddo. Applicare lo stesso principio alla Fase: execution → review-gate garantisce che il riepilogo del diff — e l'eventuale giudizio su cosa manca o cosa non convince — non sia influenzato dal contesto pregresso di chi ha scritto il codice, specialmente dopo un'implementazione lunga (es. con `subagent-driven-development`) dove il rischio di "confirmation bias" da continuità narrativa è più alto.

Come effetto collaterale, questo riduce anche il peso del context principale, ma è un beneficio secondario rispetto alla garanzia di un giudizio non contaminato.

## Requisiti

- [ ] Il subagente di review-gate riceve **solo** istruzioni di analisi (repo/branch coinvolti, comando `git diff` da eseguire) — nessun riassunto della conversazione precedente, sullo stesso modello di `challenge: subagent`
- [ ] Il subagente esegue autonomamente `git diff --stat` e `git diff --name-status --find-renames --find-copies` per ogni repo coinvolto (principale + eventuali submodule) e produce un riepilogo strutturato: file creati/modificati/eliminati/rinominati, breve descrizione dei file significativi — l'uso di `--find-renames`/`--find-copies` evita di descrivere un file rinominato come "nuovo + cancellato"
- [ ] L'isolamento in review-gate si applica **sempre**, senza soglie o giudizi condizionali su quanto sia stata "pesante" l'implementazione o su quale skill sia stata usata per l'implementazione (es. anche se `subagent-driven-development` ha già eseguito review isolate per-task, il riepilogo di review-gate resta comunque affidato a un subagente dedicato — è un controllo ridondante ma intenzionale, non un costo da evitare con condizioni)
- [ ] In caso di errore nello spawn del subagente (fail-soft): segnalare `⚠️ Impossibile isolare il riepilogo del diff — procedo mostrando il diff nel context principale.` e presentare comunque `git diff --stat`/`git diff` come oggi, senza bloccare il workflow
- [ ] Il flusso di conferma esistente (messaggio di richiesta approvazione, attesa di risposta esplicita del developer, proposta di `formal-review` opzionale) resta identico — cambia solo chi produce il riepilogo del diff, non il dialogo con l'utente
- [ ] **Nessun commit può essere eseguito prima dell'ok esplicito del developer** — l'`<HARD-GATE>` già presente in `execution: review-gate` resta invariato e non viene mai bypassato dal subagente: il subagente produce solo il riepilogo, non decide né esegue commit; un "ok" generico o un silenzio non bastano, serve una risposta di approvazione esplicita (es. "procedi con i commit") esattamente come oggi
- [ ] La Fase: overview non viene modificata — continua a essere scritta dal context principale senza subagente
- [ ] Nessun meccanismo generico di "context clear" viene introdotto — la feature si limita a estendere il pattern di isolamento via subagente già esistente in `challenge` alla fase di review-gate
- [ ] **`planning_start_at` viene registrato e mostrato in modo che non possa essere saltato silenziosamente**: in `Fase: ticket` (sia caso-a che caso-b), rendere esplicito nel testo del `SKILL.md` — con un promemoria visibile subito dopo l'apertura o creazione del ticket, non solo come nota a margine — che Claude deve annotare il timestamp corrente prima di procedere a `environment-setup`. In `Fase: estimation`, prima di proporre la stima in ore, Claude deve confrontare quel timestamp con l'orario corrente e mostrare "Misurato: `<X>` + Stimato: `<N>` = Totale: `<X+N>`" — se `planning_start_at` non è stato registrato, Claude deve segnalarlo esplicitamente all'utente invece di proporre la stima come se il dato misurato fosse disponibile (come è successo in questa sessione)

## Rischi

- **Riepilogo impreciso su rename/copy**: `git diff` grezzo può rappresentare un rename come "file nuovo + file cancellato" — mitigato imponendo al subagente l'uso di `--find-renames --find-copies` prima di scrivere il riepilogo testuale (vedi Requisiti)
- **Il riepilogo non è una garanzia di correttezza**: resta possibile che il subagente descriva "significativo" un file cosmetico o viceversa — accettato come rischio residuo, perché il vero gate è la lettura del diff completo da parte del developer, non il riepilogo, che è solo un ausilio di orientamento
- **Costo fisso di uno spawn di subagente anche per modifiche banali o già coperte da review isolate di `subagent-driven-development`**: accettato deliberatamente per mantenere la regola "sempre" semplice, prevedibile e indipendente da quale skill di implementazione è stata usata — la ridondanza con altre review non è un problema, è un controllo doppio intenzionale
- **Sovrapposizione concettuale con `execution: formal-review`**: quest'ultima resta un'analisi più approfondita e opzionale (via `wm-review-ticket`), il subagente di review-gate produce solo il riepilogo obbligatorio del diff — va documentata chiaramente la differenza di scopo tra i due per evitare confusione in futuro
- **Coupling non documentato con `challenge: subagent`**: entrambe le fasi condividono lo stesso pattern di isolamento (subagente cieco, solo path + istruzioni) — se il pattern cambia in una fase, va aggiornato anche nell'altra; da annotare nella tabella coupling di CLAUDE.md
- **Fallimento dello spawn del subagente**: gestito fail-soft (vedi Requisiti) — non blocca mai il workflow, ricade sul comportamento attuale

## Out of scope

- Nessun clear del context principale in nessun punto del workflow (il pivot deciso durante reverse-interaction abbandona questa strada in favore della sola delega a subagenti)
- Nessuna modifica alla Fase: overview o al suo timing rispetto a reverse-interaction
- Nessuna soglia o euristica per decidere quando isolare il review-gate — si isola sempre
- Nessuna modifica al pattern già esistente in `challenge: subagent`

## Moduli toccati

- `plugins/wm-skills/skills/wm-plan/SKILL.md` — sezione `execution: review-gate`: sostituire l'esecuzione diretta di `git diff --stat`/`git diff` da parte del context principale con l'invocazione di un subagente isolato che produce lo stesso riepilogo strutturato
