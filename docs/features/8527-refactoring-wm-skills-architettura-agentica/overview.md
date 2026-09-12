> Ticket: oc:8527

# Refactoring wm-skills verso architettura agentica

## Cosa cambia

Oggi le skill `wm-skills` eseguono quasi tutto nel context della sessione principale: leggono il ticket, esplorano il codebase, dialogano con il dev, scrivono gli artefatti, coordinano l'esecuzione. Il materiale grezzo letto lungo la strada (file aperti, output di comandi, grep) resta nel context fino alla fine del workflow, anche quando ne è stata ricavata una sola riga utile.

Dopo questa feature le fasi che macinano molto testo per restituire poco vengono affidate a **agenti separati**: l'agente legge nel proprio context e restituisce al principale solo la conclusione. La sessione principale conserva ciò che ha valore — le decisioni, il dialogo con il dev, le istruzioni della skill — e non il materiale da cui sono state ricavate.

Il meccanismo di delega viene scritto **una sola volta** in `plugins/wm-skills/shared/`, non dentro le singole skill, in previsione del core condiviso che il team realizzerà in un ciclo successivo.

La feature si articola in due momenti, entrambi in questo ciclo:

1. **Analisi di fattibilità** — mappa fase per fase delle tre skill, con verdetto "delegabile / non delegabile" e motivazione. Prodotto come documento a sé, versionato accanto alle skill.
2. **Refactoring** — introduzione degli agenti individuati e del meccanismo condiviso che li governa.

## Perché

Tre ragioni, emerse in `reverse-interaction`:

- **Degrado di qualità sulle fasi finali.** Il context non si rompe a una soglia, si degrada lungo una curva: il materiale irrilevante diluisce le istruzioni importanti, e ciò che sta *in mezzo* alla sessione sbiadisce per primo. In `wm-plan` è esattamente il pezzo che non dovrebbe sbiadire: le decisioni prese in `reverse-interaction` devono governare `execution`, che arriva molto dopo. Aggravante specifica di questo repo: `wm-plan/SKILL.md` è lungo 1118 righe e deve restare governante fino all'ultima fase — l'istruzione più importante è anche la più vecchia.
- **Costo.** Il context si ri-invia a ogni turno: un file letto a metà workflow si paga su tutti i turni successivi. Non è una ragione alternativa alla precedente, è la stessa leva vista dall'altro lato.
- **Debito evitato sul core futuro.** Il team prevede un core con le parti comuni scritte una volta sola. Scrivere il meccanismo di delega dentro ogni skill produrrebbe tre copie destinate a divergere, e il core diventerebbe un'estrazione di tre versioni già diverse. Scritto in `shared/` fin da ora, il core nasce per accumulo.

## Requisiti

**Meccanismo condiviso**

- [ ] Il meccanismo di delega è descritto in un unico file sotto `plugins/wm-skills/shared/`, che le skill richiamano senza duplicarne il contenuto
- [ ] Il file definisce: forma del prompt di un agente, input minimo ammesso, contratto di ritorno, tetto all'output e comportamento in caso di fallimento
- [ ] Il contratto di ritorno impone un **tetto misurabile** all'output di ogni agente, con comportamento definito se superato: un agente che restituisce troppo costa più del baseline (si paga il suo context *più* il suo output nel principale)
- [ ] Per ogni agente è definito un **fallback esplicito** in caso di fallimento — nessun agente eredita il fail-soft per default culturale del repo
- [ ] Il file distingue in modo esplicito i **due tipi di delega**, che non vanno unificati: subagenti *ciechi* per indipendenza di giudizio (`challenge`, `review-gate`) e agenti *informati* per volume di lettura (i tre nuovi). Passare contesto ai primi ne distrugge la proprietà
- [ ] Esiste un documento di analisi di fattibilità in `plugins/wm-skills/shared/` con la tabella fase → delegabile sì/no → motivo per le tre skill, e il criterio per classificare una fase nuova
- [ ] Il criterio è esplicito: input piccolo, output piccolo, lavoro intermedio grande, nessuna interazione diretta con il dev

**Deleghe statiche, nessuna soglia**

- [ ] Le deleghe sono decise staticamente per fase: nessuna attivazione automatica basata su una misura del context
- [ ] La misura del context resta solo **informativa**, mostrata nell'header di sessione, dove non decide nulla
- [ ] La misura dichiara sempre un valore assoluto di token occupati, **mai una percentuale**: la dimensione della finestra non è leggibile dal transcript e andrebbe indovinata
- [ ] La lettura esclude le righe `isSidechain` e dichiara la misura non disponibile se il transcript non è raggiungibile o ha formato inatteso — mai un numero inventato

**Agenti**

- [ ] Sono introdotti i cinque agenti della sezione "Agenti introdotti", ciascuno con prompt, input minimo, contratto di ritorno e fallback definiti
- [ ] Nessuna fase elencata fra quelle non delegate viene delegata in questo ciclo
- [ ] `wm-codebase-research` è chiamato una volta a inizio `reverse-interaction` sulle ricerche prevedibili; agenti puntuali durante il dialogo sono ammessi come eccezione
- [ ] **Ogni affermazione del dossier è accompagnata da una prova verificabile**: percorso del file, intervallo di righe ed estratto *verbatim*. Un'affermazione senza prova viene scartata, non discussa
- [ ] Le prove sono verificate **a macchina e su tutte le affermazioni**, non a campione: si controlla che quelle righe di quel file contengano davvero quell'estratto
- [ ] Se una prova non combacia, il dossier è marcato inattendibile e la ricerca viene rifatta nel context principale
- [ ] `wm-env-detect` restituisce un contratto **nominale** sulle variabili d'ambiente, non solo i flag: `$DOCKER_PROJECT_DIR_NAME` è letto molto dopo, da `phpstan-check`
- [ ] `wm-estimate` è **cieco**: riceve solo i percorsi di `overview.md` e `plan.md`, mai un riassunto della conversazione
- [ ] `wm-estimate` verifica nel codebase l'esistenza di un pattern equivalente prima di attribuire il buffer di novità di dominio, invece di dichiararlo a giudizio
- [ ] La quota misurata della pianificazione e la conferma del dev restano nel context principale, fuori dalla delega
- [ ] Le regole applicate da `wm-context-guard` e `wm-context-doctor` stanno in un unico file in `shared/`, letto da entrambi, mai duplicato
- [ ] I due agenti riconoscono i quattro tipi di rilievo: ripetizione, contraddizione, duplicato dal codice, fuori posto
- [ ] Nessuno dei due modifica un `CLAUDE.md` senza approvazione esplicita del dev
- [ ] Entrambi funzionano su un repo in qualsiasi stato, migrato o no (vedi oc:8528): la struttura a indice non è un prerequisito

**Tracciabilità e confini**

- [ ] Gli artefatti prodotti con la modalità agentica sono riconoscibili a posteriori, sul modello del marcatore già usato per le stime: senza, non è possibile sapere quali ticket sono stati pianificati con un dossier degradato
- [ ] La fase `update-context` continua a scrivere il `CLAUDE.md` nella forma attuale: struttura a indice e migrazione sono materia di oc:8528
- [ ] Il contratto artefatti `docs/features/<slug>/` non cambia: `overview.md`, `plan.md`, `notes.md` restano obbligatori e con la stessa struttura
- [ ] Le deleghe già esistenti restano attive con la loro motivazione originale
- [ ] `wm-review-ticket` è collegata al meccanismo condiviso dove la delega è evidente
- [ ] `wm-tag` è verificata come funzionante con il nuovo `wm-plan` in tag-mode
- [ ] `CLAUDE.md` aggiornato: decisione architetturale, riga in "Feature disponibili", nuovo coupling skill ↔ file condiviso nella tabella

## Agenti introdotti

Criterio di delega: una fase si delega quando ha **input piccolo, output piccolo e lavoro intermedio grande**, e non richiede interazione diretta con il dev — un agente non può dialogare con l'utente.

**Le deleghe sono statiche, decise per fase.** Nessuna soglia, nessuna attivazione automatica basata su una misura: una fase si delega perché lì conviene sempre, non perché un numero ha superato un valore. Vedi `Rischi → misura del context` per il motivo.

| Agente | Fase | Cosa fa | Cosa restituisce |
|---|---|---|---|
| `wm-codebase-research` | `reverse-interaction` | riceve le domande aperte ed esplora codice e db per rispondere | dossier compatto: risposta + citazioni dei file |
| `wm-env-detect` | `environment-setup` | esegue i check di stack, docker, submodule, presenza PHPStan | i flag risolti, con contratto nominale sulle variabili |
| `wm-context-guard` | `update-context` | esamina la modifica proposta al `CLAUDE.md` prima che venga scritta | rilievi di forma, mai contenuto |
| `wm-context-doctor` | invocato esplicitamente dal dev | esamina un `CLAUDE.md` nel suo insieme e propone un piano di riordino | piano da approvare, nessuna modifica diretta |
| `wm-estimate` | `estimation` | stima il costo della feature leggendo solo overview e piano | tabella per componente + buffer motivati |

### wm-codebase-research e wm-env-detect

Delega per volume: molto materiale letto, esito corto. `wm-codebase-research` è l'agente da cui è nato il ticket — oggi ogni file aperto per rispondere a una domanda resta nel context fino a fine workflow. Chiamato una volta a inizio fase sulle ricerche prevedibili; agenti puntuali durante il dialogo restano possibili come eccezione, non come regola.

### wm-estimate

Unico agente del lotto delegato per **indipendenza di giudizio** e non per volume — stessa ragione di `challenge` e `review-gate`, quindi **cieco**: riceve i percorsi di `overview.md` e `plan.md` e nient'altro, nessun riassunto della conversazione.

Chi ha condotto il dialogo e scritto l'overview è la persona meno adatta a stimarne il costo: ha appena finito di convincersi che il problema è chiaro, e stima ottimista in modo sistematico. Un agente che non ha vissuto la conversazione non ha quel pregiudizio, e ha lo stesso profilo di chi eseguirà.

C'è anche una componente di volume, nella parte oggi svolta peggio: il **buffer di novità del dominio** richiede di verificare se esiste già nel codebase un pattern equivalente e quanti file leggono il simbolo toccato. Oggi quella verifica si fa a occhio; l'agente la fa davvero, ed è ricerca con esito corto.

Restano fuori dalla delega:

- la **quota misurata** della pianificazione — due timestamp che ha solo il context principale
- la **conferma** — la stima finale resta quella approvata dal dev, invariata

**Limite dichiarato:** un agente cieco può stimare male per difetto d'informazione, dove il principale stima male per eccesso di ottimismo. Non è noto a priori quale errore sia minore, e non lo sarà finché non si confronteranno stime e ore reali su un campione. Ma l'errore dell'agente non è sistematico, quindi è correggibile con la calibrazione; quello del principale sì.

### wm-context-guard e wm-context-doctor

Due agenti che fanno lo **stesso lavoro in momenti diversi**, sullo stesso corpo di regole:

| | `wm-context-guard` | `wm-context-doctor` |
|---|---|---|
| quando | automatico, al momento di scrivere | esplicito, invocato dal dev |
| cosa guarda | il pezzo nuovo contro il file esistente | tutto il file, nel suo insieme |
| esito tipico | "questo contraddice la voce X; questa parte va in un file dedicato" | un piano di riordino da approvare |
| frequenza | ogni feature | quando serve |

Entrambi controllano **come** si scrive, mai **cosa**: il merito di una decisione lo può giudicare solo chi ha assistito al lavoro. L'isolamento qui è un vantaggio, non un limite — chi ha appena scritto un paragrafo lo trova sempre necessario.

Quattro tipi di rilievo:

1. **ripetizione** — la stessa cosa detta due volte con parole diverse → fondere
2. **contraddizione** — due voci che dicono il contrario → marcare la vecchia come superata o rimuoverla, **con conferma del dev**
3. **duplicato dal codice** — già leggibile da un file o dalla git history → non scrivere affatto
4. **fuori posto** — procedura operativa o contratto di dettaglio che non serve a chi apre il repo → spostare in un file dedicato, lasciando il rimando

Criterio di collocazione: il `CLAUDE.md` risponde a *"cosa devo sapere di questo repo prima di toccarlo"*, ed è pagato da ogni sessione, anche da quella aperta per un typo. Il dettaglio operativo si legge quando serve. Il repo ha già questo pattern in `shared/orchestrator-fallback.md`; gli agenti lo generalizzano invece di lasciarlo all'intuito di chi scrive.

Quando coerenza e brevità confliggono, l'informazione non si sacrifica: si sposta nel file dedicato.

**Entrambi propongono, non eseguono mai.** Nessuna modifica al `CLAUDE.md` senza approvazione esplicita del dev: è il file che governa il comportamento di Claude sull'intero repo.

Le regole non stanno dentro nessuno dei due agenti: stanno in un file in `plugins/wm-skills/shared/`, che entrambi leggono. È il primo mattone del core condiviso. Due copie divergerebbero alla prima modifica, con il sintomo peggiore possibile: `wm-context-doctor` che riordina un repo in una forma che `wm-context-guard` poi non riconosce.

### Fasi deliberatamente non delegate

- `init-context` — il `CLAUDE.md` del repo target serve *tutto*, e per tutto il workflow: delegarlo significherebbe riassumerlo, e un riassunto qui è una perdita secca
- `overview` — nasce dal dialogo e va scritta da chi lo ha condotto (decisione già presa in oc:8282, qui confermata)
- `write-plan` — delega già a `superpowers:writing-plans`, e il piano torna comunque nel principale perché va approvato ed eseguito
- `notes` — il registro delle decisioni, incluso l'override motivato di un hard-block PHPStan con responsabilità attribuita al dev (oc:8341). Quelle informazioni esistono solo nel dialogo: un agente che arriva alla fine non ha assistito ai fatti
- `review-gate: phpstan-check` — resta nel principale. È un hard-block deliberato (oc:8341, che blocca anche sui fallimenti infrastrutturali per impedire il bypass implicito): un agente in mezzo a un gate di sicurezza può trasformare un fallimento in un pass silenzioso. Il risparmio di context si ottiene filtrando l'output JSON con `jq`, senza delega
- `docker-check` — ha effetti collaterali sul sistema del dev (`docker compose stop`): non vanno decisi in un context che il dev non vede passare
- ogni fase di dialogo con il dev — per costruzione

Restano invariate le due deleghe già presenti — `challenge: subagent` e `review-gate: subagent` — con la loro motivazione originale: **l'indipendenza di giudizio, non il risparmio di context**. Sono subagenti *ciechi* e devono restarlo.

## Rischi

Emersi dalla `Fase: challenge` (revisore adversariale isolato) e recepiti.

**Risolti cambiando il progetto**

- **La misura del context non è una misura.** Il transcript dà i token occupati ma **non la dimensione della finestra**: qualsiasi percentuale richiede un denominatore indovinato (200k o 1M, e il modello può cambiare a metà sessione). Sbagliando denominatore, la delega non si attiva mai o sempre. → *Risolto: soglie e percentuali eliminate, deleghe statiche per fase, misura solo informativa e in valore assoluto.*
- **La fonte può mentire.** I subagenti possono scrivere nello stesso transcript (`isSidechain`); dopo una compattazione il valore crolla e sembrerebbe context liberato, con oscillazione attorno alla soglia. → *Risolto dalla stessa scelta: un numero che non decide non può decidere male. Filtro `isSidechain` comunque specificato.*
- **PHPStan delegato annullava un hard-block.** `oc:8341` blocca deliberatamente anche sui fallimenti infrastrutturali, per impedire il bypass implicito; un agente con fallback fail-soft avrebbe trasformato un fallimento in un pass silenzioso, con un errore reale sul diff verso un repo cliente. → *Risolto: phpstan resta nel principale, output filtrato con `jq`.*
- **Il redattore scriveva un registro di fatti a cui non ha assistito.** `notes.md` contiene le decisioni prese nel dialogo e la responsabilità del dev sugli override: materiale che esiste solo nel context principale. Era la stessa ragione per cui `overview` è esclusa, applicata con esito opposto. → *Risolto: nessun agente scrive contenuto; `wm-context-guard` controlla la forma, non il merito.*

**Accettati, con mitigazione**

- **Un dossier sbagliato non è verificabile da chi lo riceve.** Se `wm-codebase-research` fraintende un file, il principale non ha il materiale per accorgersene — è esattamente il materiale scartato. L'errore si propagherebbe in overview, piano ed esecuzione, e si scoprirebbe a codice scritto; peggio, un dossier corto e sicuro di sé rende la sessione *più* scorsevole proprio mentre sbaglia. → *Mitigazione forte: obbligo di prova. Ogni affermazione porta percorso, righe ed estratto verbatim, e le prove si verificano a macchina su tutte le affermazioni (`sed -n 'X,Yp' <file>` e confronto con l'estratto), portando in context tre righe invece di un file. Una prova che non combacia invalida il dossier.*
  **Residuo accettato:** la verifica prova che la citazione è **autentica**, non che l'interpretazione sia **corretta** — l'agente può citare bene una riga e trarne la conclusione sbagliata. È un errore più raro e, a differenza del primo, visibile: avendo davanti sia l'estratto sia la conclusione, la contraddizione si legge senza aprire nulla.
- **Il file condiviso è un punto di rottura per tre skill.** Oggi un errore in una skill non tocca le altre. → *Mitigazione: coupling dichiarato in `CLAUDE.md`; nessun test automatico possibile (`claude plugin validate` controlla il frontmatter, non la semantica). Rischio accettato consapevolmente.*
- **I due tipi di delega possono essere confusi.** Sotto un unico "meccanismo condiviso", qualcuno potrebbe uniformare la forma del prompt e passare contesto anche ai subagenti ciechi, distruggendo in silenzio la proprietà su cui si regge oc:8282. → *Mitigazione: distinzione scritta a caratteri espliciti nel file condiviso, come requisito.*
- **La strumentazione costa quello che risparmia.** Ogni misura è una chiamata il cui output entra nel context. → *Mitigazione: la misura è una sola riga, una volta per sessione, nell'header.*
- **Dipendenza da un formato non contrattuale.** Il layout di `~/.claude/projects/` e il formato del transcript appartengono a Claude Code e possono cambiare senza preavviso — lo stesso tipo di fragilità già sperimentato in questo repo con la risoluzione a runtime del path del plugin. → *Mitigazione: la misura è informativa, quindi un guasto degrada una riga dell'header e non il workflow.*
- **Rollback parziale impossibile.** Le regole stanno nel file condiviso ma i richiami nelle `SKILL.md`: disattivare una sola delega richiede di toccare più file, e il repo rifiuta per principio i kill-switch persistenti (oc:8341). L'unico rollback è un revert totale via `/plugin marketplace update`, che non è sincrono sul team. → *Rischio accettato: coerente con una decisione già presa.*
- **Ciò che è stato scritto in produzione non torna indietro.** Descrizioni dei ticket, stime, `CLAUDE.md` dei repo target restano anche dopo un revert. → *Mitigazione: requisito di tracciabilità degli artefatti prodotti in modalità agentica, sul modello del marcatore già usato per le stime.*
- **Il dev che lavora in locale sulle skill** ha sessioni lunghissime per motivi estranei alla fase. → *Non più un problema: senza soglie, non esiste un'attivazione che possa restare bloccata.*

## Out of scope

- Il **core condiviso** vero e proprio: questo ciclo mette il primo file in `shared/`, non realizza l'estrazione completa delle parti comuni
- Riscrittura integrale di `wm-review-ticket` e `wm-tag` con lo stesso criterio applicato a `wm-plan`
- Modifiche al server MCP `orchestrator` e ai suoi tool
- Modifiche al contratto artefatti `docs/features/`
- Parallelizzazione degli agenti: le deleghe previste sono sequenziali
- Attivazione automatica della delega in base a una misura del context: scartata in `Fase: challenge`, non rimandata
- **Gerarchia e disciplina dei `CLAUDE.md`** — struttura a indice, lint dei link, marcatore `wm-context-layout`, procedura di migrazione per repo: scorporati in **oc:8528**. Il punto di contatto fra i due ticket è la sola fase `update-context`
- La **LLM wiki di Karpathy** come knowledge base del team (`raw/`, ingestione, `log.md`): valutazione rimandata a un ticket a sé, citata in oc:8528

## Moduli toccati

Repo unico `claude-marketplace` (nessun submodule, feature interamente custom).

| File | Repo | Intervento |
|---|---|---|
| `plugins/wm-skills/shared/<analisi-fattibilità>.md` | principale | nuovo — mappa fasi e criterio di classificazione |
| `plugins/wm-skills/shared/<meccanismo-delega>.md` | principale | nuovo — forma del prompt, contratto di ritorno, tetto output, fallback, distinzione fra delega cieca e informata |
| `plugins/wm-skills/shared/<regole-claude-md>.md` | principale | nuovo — regole lette da `wm-context-guard` e `wm-context-doctor` |
| `plugins/wm-skills/skills/wm-plan/SKILL.md` | principale | modificato — richiamo al meccanismo condiviso, deleghe in `reverse-interaction`, `environment-setup`, `estimation`, `update-context`; misura informativa nell'header |
| `plugins/wm-skills/skills/wm-review-ticket/SKILL.md` | principale | modificato — collegamento al meccanismo condiviso |
| `plugins/wm-skills/skills/wm-tag/SKILL.md` | principale | verificato, modificato solo se necessario |
| `plugins/wm-skills/.claude-plugin/plugin.json` | principale | bump `version` (checklist di release) |
| `plugins/wm-skills/mcp/internal/version/version.go` | principale | allineamento costante `Version` + rebuild binario |
| `CLAUDE.md` | principale | decisione architetturale, feature disponibili, coupling |
| `docs/wm-plan-diagram/index.html` | principale | aggiornamento contenuto diagramma se il workflow cambia |
