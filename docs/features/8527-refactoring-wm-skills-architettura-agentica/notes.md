> Ticket: oc:8527

# Notes — Refactoring wm-skills verso architettura agentica

## Decisioni

### Soglie sul context eliminate in Fase: challenge
Il progetto iniziale prevedeva di attivare la delega al superamento di una soglia di context
(40% attenzione, 65% delega automatica). Scartato: il transcript di sessione espone i token
occupati ma NON la dimensione della finestra, quindi ogni percentuale avrebbe avuto un
denominatore indovinato (200k o 1M, e il modello può cambiare a metà sessione con `/model`).
Le deleghe sono diventate statiche per fase e la misura è rimasta solo informativa nell'header.

### PHPStan non delegato
Delegare `review-gate: phpstan-check` avrebbe potuto trasformare un fallimento dell'agente in
un pass silenzioso, annullando l'hard-block deciso in oc:8341 (che blocca anche sui fallimenti
infrastrutturali proprio per impedire il bypass implicito). Il comando resta nel principale,
con l'output filtrato via `jq`.

### Nessun agente scrive contenuto
Il progetto iniziale prevedeva un agente "redattore" che scrivesse `notes.md` e il `CLAUDE.md`.
Scartato: quei file registrano decisioni e responsabilità che esistono solo nel dialogo, a cui
un agente non ha assistito. È la stessa ragione per cui `overview` è esclusa dalla delega. Al
suo posto `wm-context-guard`, che controlla la forma e non il merito.

### Obbligo di prova verbatim per wm-codebase-research
Un dossier non verificabile propagherebbe un errore di lettura fino al codice scritto. Ogni
affermazione porta percorso, righe ed estratto verbatim, verificabile a macchina con
`sed -n 'X,Yp'`. Residuo accettato: la verifica prova che la citazione è autentica, non che
l'interpretazione sia corretta.

### wm-estimate aggiunto a piano già impostato
Aggiunto su richiesta del dev: unico agente delegato per indipendenza di giudizio e non per
volume, perché chi ha condotto il dialogo e scritto l'overview stima ottimista in modo
sistematico.

### Agenti formali di plugin invece di prompt in file condivisi
Scelti agenti formali di plugin (`plugins/wm-skills/agents/*.md` con frontmatter) invece di
prompt in file condivisi: scelta del dev, richiesta esplicita di "una skill formale che segue
le best practice".

### La garanzia «propone, non esegue» è comportamentale, non strutturale
Emersa in review: `wm-context-guard` e `wm-context-doctor` non hanno `Write`/`Edit`, ma hanno
`Bash`, con cui si può scrivere lo stesso. `Bash` non è stato rimosso perché i loro prompt lo
usano per `wc`/`awk`: la garanzia è diventata un divieto comportamentale esplicito nel prompt.
La protezione dipende dal rispetto del prompt e dal permesso chiesto all'utente, non dal
sistema dei tool.

### Riferimenti storici non riscritti
I nomi delle tre skill inesistenti (`our-code-style`, `our-pr-checklist`,
`our-deploy-post-merge`) restano in due file sotto `docs/features/` (un piano archiviato e il
piano di questa feature): sono registri di ciò che era stato pianificato, non istruzioni
attive.

### Revisione con l'agente, aggiunta al review-gate

Richiesta del dev dopo il riepilogo del diff, prima del commit: «la comunicazione con l'agente deve esserci, le stime io le devo validare — dire sì va bene, no modifica questo. Sono poche le cose che prendiamo a scatola chiusa senza una validazione.»

Il flusso costruito fino a quel momento offriva due sole vie sull'esito di un agente: accettarlo o sostituirlo a mano. Sostituire un numero però non è validare: il ragionamento sbagliato resta dov'era e si ripete alla volta successiva.

Aggiunta la sezione `## Revisione con l'agente` in `shared/agent-delegation.md`, con tre decisioni:

- **Si riprende la stessa sessione dell'agente, non se ne apre una nuova.** L'agente ha ancora nel proprio context il materiale letto: risponde su ciò che aveva davanti, e il giro costa poco. Un agente nuovo rileggerebbe tutto, potrebbe divergere per ragioni estranee all'obiezione, e ripagherebbe l'intero costo di lettura — cioè quello che la delega serviva a evitare. Ne segue che l'identità dell'agente va conservata finché la sua fase non è chiusa.
- **Tre vie al dev, mai due**: accettare, sostituire, far rifare con le proprie obiezioni. Massimo due giri, poi decide il dev: oltre non si converge, e un agente che continua a correggersi finisce per assecondare invece che ragionare.
- **Si itera con chi produce un giudizio, non con chi riporta fatti.** Vale per `wm-codebase-research`, `wm-estimate`, `wm-context-guard` e soprattutto `wm-context-doctor`, il cui piano di riordino viene tipicamente accettato solo in parte. Non vale per `wm-env-detect`, che restituisce flag misurati: un errore lì non è un'opinione da discutere ma una lettura sbagliata, e si verifica rieseguendo il check. Non vale mai per `challenge` e `review-gate`: un revisore che si corregge quando il criticato ribatte è l'opposto di ciò che serve.

Ciò che si rimanda all'agente sono le obiezioni sull'esito, mai il contesto della conversazione: `wm-estimate` resta cieco anche durante la revisione.

## Follow-up

- Se `our-code-style`, `our-pr-checklist` e `our-deploy-post-merge` servono davvero al team,
  vanno scritte in un ticket dedicato: oggi erano rimandi a vuoto.
- Soglie e calibrazione non esistono più come meccanismo: un'eventuale attivazione dinamica
  futura richiede prima un modo affidabile di conoscere la dimensione della finestra di
  context.
- Il requisito di tracciabilità degli artefatti prodotti in modalità agentica non è stato
  implementato: con le deleghe statiche *tutti* gli artefatti sono agentici, quindi un
  marcatore ovunque non distinguerebbe nulla. La discriminante utile è la versione del
  plugin, già tracciata. Da riconsiderare se le deleghe torneranno condizionali.
- `wm-context-doctor` non è mai stato eseguito su un repo reale: la sua prima esecuzione
  sarà anche il suo primo collaudo.
- oc:8528 estenderà queste regole con la gerarchia dei `CLAUDE.md` (indice + pagine, lint,
  marcatore `wm-context-layout`, migrazione per repo).

## Divergenze dal piano, task per task

- Gli Step 2 (rigenerazione dell'Artifact del diagramma) e 7 (commit) del Task 11 sono stati
  deliberatamente esclusi dall'esecuzione dei subagenti: sono affidati al context principale,
  dopo l'approvazione esplicita del developer in `execution: review-gate`. Nessun redeploy
  dell'Artifact né alcun comando git è stato eseguito in questa esecuzione.

### Task 11 — bump di versione non eseguito

Il piano (Task 11, Step 3) prevedeva il bump minor `1.3.0` → `1.4.0` nei tre punti della checklist di release, più la ricompilazione del binario MCP. Il bump è stato eseguito e poi **annullato**: il dev ha segnalato al review-gate che la `1.3.0` non è ancora stata chiusa, quindi questo lavoro vi rientra invece di aprire una versione nuova.

`plugin.json`, `mcp/internal/version/version.go` e la riga `**Versione installata:**` di `wm-plan/SKILL.md` restano a `1.3.0`. Il binario `bin/orchestrator-mcp` è stato ripristinato alla versione committata: `version.go` è tornato identico a `HEAD`, e una ricompilazione avrebbe lasciato nel diff un binario diverso a parità di contenuto, perché Go non produce build bit-identiche.

Il bump andrà fatto quando la `1.3.0` verrà chiusa, seguendo la checklist di release in `CLAUDE.md` → `## Versioning del plugin wm-skills`.
