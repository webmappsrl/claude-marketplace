> Ticket: oc:8527

# Refactoring wm-skills verso architettura agentica — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** spostare fuori dal context della sessione principale il materiale grezzo letto dalle skill `wm-skills`, delegandolo a cinque agenti formali che restituiscono solo conclusioni verificabili.

**Architecture:** gli agenti sono definiti come agenti di plugin (`plugins/wm-skills/agents/*.md`, frontmatter + prompt). Le regole comuni a più agenti non stanno dentro nessuno di essi ma in tre file sotto `plugins/wm-skills/shared/`, letti da chi serve: è il primo mattone del core condiviso. Le deleghe sono **statiche per fase** — nessuna soglia, nessuna attivazione automatica su misura del context.

**Tech Stack:** Markdown (skill e agenti di plugin Claude Code), frontmatter YAML, bash per le verifiche, `claude plugin validate` come unico validatore disponibile.

**Spec:** `docs/features/8527-refactoring-wm-skills-architettura-agentica/overview.md`

## Global Constraints

- Repo unico `claude-marketplace`, nessun submodule. Tutti i percorsi sono relativi alla root del repo.
- **Nessun commit durante l'esecuzione.** I blocchi `git` in questo piano sono istruzioni testuali per il dev, mai comandi da eseguire. Un solo commit a lavoro finito (Task 11).
- Prefisso obbligatorio `wm-` per ogni skill e ogni agente, in kebab-case, coerente fra nome file e campo `name` del frontmatter.
- Ogni agente ha: `name`, `description` (una frase con soggetto, "Use when…"/"Usa quando…"), `model`, `tools` (solo quelli necessari).
- **Due tipi di delega, mai unificati.** *Cieca* (indipendenza di giudizio): riceve solo percorsi e istruzioni, mai un riassunto della conversazione — `challenge`, `review-gate`, `wm-estimate`. *Informata* (volume di lettura): riceve il contesto utile al compito — `wm-codebase-research`, `wm-env-detect`, `wm-context-guard`, `wm-context-doctor`.
- `challenge: subagent` e `review-gate: subagent` **non vanno toccati**: restano invariati e ciechi.
- Fasi che restano interamente nel context principale: `init-context`, `overview`, `write-plan`, `notes`, `review-gate: phpstan-check`, `environment-setup: docker-check`, ogni dialogo con il dev.
- Il contratto artefatti `docs/features/<slug>/` non cambia.
- `update-context` continua a scrivere il `CLAUDE.md` nella forma attuale: la struttura a indice è materia di **oc:8528**.
- Ogni agente funziona su repo in qualsiasi stato, migrato o no.
- La misura del context è **solo informativa**, in **valore assoluto** (mai percentuale), e dichiara "non disponibile" invece di inventare un numero.
- Validazione prima del commit: `claude plugin validate .` dalla root.

---

### Task 1: Meccanismo di delega condiviso

**Files:**
- Create: `plugins/wm-skills/shared/agent-delegation.md`

**Interfaces:**
- Produces: il contratto che i Task 4-8 citano invece di riscrivere. Nomi delle sezioni referenziate altrove: `## Tipi di delega`, `## Contratto di ritorno`, `## Tetto all'output`, `## Fallback`, `## Misura del context`.

- [ ] **Step 1: creare il file con l'intestazione e i due tipi di delega**

```markdown
# Delega ad agenti — meccanismo condiviso

Letto da: `wm-plan`, `wm-review-ticket`, `wm-tag` e dagli agenti in `plugins/wm-skills/agents/`.
Non duplicare questo contenuto in una skill: citarlo.

## Tipi di delega

Due tipi, con proprietà opposte. Non vanno unificati: passare contesto a un agente
cieco ne distrugge la ragione d'essere.

| | Delega cieca | Delega informata |
|---|---|---|
| Perché | indipendenza di giudizio | volume di lettura |
| Riceve | solo percorsi e istruzioni | il contesto utile al compito |
| Non riceve mai | alcun riassunto della conversazione | — |
| Casi | `challenge`, `review-gate`, `wm-estimate` | `wm-codebase-research`, `wm-env-detect`, `wm-context-guard`, `wm-context-doctor` |

Chi ha condotto un ragionamento tende a confermarlo invece di valutarlo: è il motivo
per cui la delega cieca esiste. Se un giorno un agente cieco sembra "mancare di
contesto", la risposta non è dargliene — è accettare che il suo giudizio nasca altrove.

## Quando si delega

Le deleghe sono **statiche, decise per fase** in ogni skill. Non esiste attivazione
automatica basata su una misura del context: vedi `## Misura del context`.

Una fase è delegabile quando ha **input piccolo, output piccolo, lavoro intermedio
grande** e non richiede interazione diretta con il dev — un agente non può dialogare
con l'utente.

## Contratto di ritorno

Ogni agente restituisce **solo conclusioni**, mai il materiale da cui derivano.
Il formato esatto è definito nel prompt del singolo agente.

Chi riceve un'affermazione senza prova non ha modo di verificarla: è esattamente il
materiale che si è scelto di non far entrare nel context. Dove la correttezza
dell'affermazione conta (`wm-codebase-research`), la prova è obbligatoria e
verificabile a macchina.

## Tetto all'output

Ogni agente ha un tetto dichiarato nel proprio prompt. Un agente che restituisce
troppo costa **più** del non delegare: si paga il suo context e in più il suo output
nel principale.

Superato il tetto, il context principale chiede all'agente una sintesi entro il
limite invece di accettare l'output lungo. Un output che resta oltre il tetto al
secondo tentativo va trattato come fallimento (vedi `## Fallback`).

## Fallback

**Nessun agente eredita un comportamento di default.** Ogni prompt dichiara cosa
succede se l'agente fallisce, va in timeout o restituisce un formato inatteso.

Il fail-soft è frequente in questo repo (`environment-setup`, `docker-check`), ma non
è una regola generale: dove un agente alimenta una decisione di sicurezza il fallimento
deve bloccare. Per questo `review-gate: phpstan-check` non è delegato affatto.

## Misura del context

Il transcript di sessione registra i token occupati:

```bash
TRANSCRIPT=~/.claude/projects/<progetto>/<session-id>.jsonl
grep -v '"isSidechain":true' "$TRANSCRIPT" | grep -o '"usage":{[^}]*}' | tail -1
```

Somma di `input_tokens`, `cache_creation_input_tokens` e `cache_read_input_tokens`.

Tre vincoli, tutti obbligatori:

1. **Valore assoluto, mai percentuale.** Il transcript non contiene la dimensione della
   finestra: ogni percentuale richiederebbe un denominatore indovinato (200k o 1M, e il
   modello può cambiare a metà sessione).
2. **Solo informativa.** Non attiva né disattiva nulla. Un numero che non decide non può
   decidere male — ed è il motivo per cui i suoi difetti noti (compattazione che azzera il
   valore, path non contrattuale, formato di terze parti) restano innocui.
3. **Mai un numero inventato.** Transcript non raggiungibile o formato inatteso →
   `⚠️ Misura del context non disponibile.`, e si prosegue.

Il filtro `isSidechain` esclude le righe scritte dai subagenti che condividono il
transcript: senza, si leggerebbe il context dell'agente invece del proprio.
```

- [ ] **Step 2: verificare che il file non contenga soglie né percentuali**

Run:
```bash
grep -nE '[0-9]+%|soglia|soglie' plugins/wm-skills/shared/agent-delegation.md
```
Expected: nessun output (le uniche occorrenze ammesse sono nella frase che spiega perché le percentuali sono vietate — se il grep stampa quella riga, verificarla a mano e proseguire).

- [ ] **Step 3: verificare che il comando di misura funzioni davvero**

Run:
```bash
TRANSCRIPT=$(ls -t ~/.claude/projects/-Users-bongiu-Documents-claude-marketplace/*.jsonl | head -1)
grep -v '"isSidechain":true' "$TRANSCRIPT" | grep -o '"usage":{[^}]*}' | tail -1
```
Expected: una riga `"usage":{...}` con `cache_read_input_tokens`. Se non la stampa, il comando nel file va corretto prima di proseguire.

---

### Task 2: Regole di scrittura dei CLAUDE.md

**Files:**
- Create: `plugins/wm-skills/shared/claude-md-rules.md`

**Interfaces:**
- Consumes: nulla.
- Produces: le regole lette da `wm-context-guard` (Task 7) e `wm-context-doctor` (Task 8). Sezioni referenziate: `## I quattro rilievi`, `## Dove va cosa`, `## Cosa non si scrive`.

- [ ] **Step 1: creare il file**

```markdown
# Scrivere nei CLAUDE.md — regole condivise

Letto da `wm-context-guard` e `wm-context-doctor`. Un solo corpo di regole per entrambi:
due copie divergerebbero alla prima modifica, e il sintomo peggiore sarebbe
`wm-context-doctor` che riordina un repo in una forma che `wm-context-guard` non riconosce.

## Il principio

Il `CLAUDE.md` risponde a una sola domanda: **cosa devo sapere di questo repo prima di
toccarlo.** È pagato da ogni sessione, anche da quella aperta per un typo.

Il dettaglio operativo — una procedura passo-passo, un contratto di campo, una tabella di
valori — serve solo a chi sta facendo quella cosa specifica: va in un file dedicato, con
il rimando nel `CLAUDE.md`. Il repo usa già questo pattern in
`plugins/wm-skills/shared/orchestrator-fallback.md`.

Quando coerenza e brevità confliggono, l'informazione **non si sacrifica: si sposta.**

## I quattro rilievi

| Rilievo | Cos'è | Cosa proporre |
|---|---|---|
| **ripetizione** | la stessa cosa già detta altrove con parole diverse | fondere nella voce esistente |
| **contraddizione** | due voci che dicono il contrario | marcare la vecchia come superata o rimuoverla — **richiede conferma del dev** |
| **duplicato dal codice** | già leggibile da un file, da un test o dalla git history | non scrivere affatto |
| **fuori posto** | procedura o dettaglio che non serve a chi apre il repo | spostare in un file dedicato, lasciare il rimando |

## Dove va cosa

- **La documentazione segue il codice.** Ciò che riguarda un submodule va nel `CLAUDE.md`
  del submodule, non in quello del repo principale — stesso principio già valido per gli
  `overview.md` in `wm-plan`.
- Una feature completata: **una riga** in `## Feature disponibili`.
- Una scelta non ovvia che un futuro Claude rifarebbe da capo: un blocco in
  `## Decisioni architetturali`.
- Tutto il resto: valutare se è davvero materia da `CLAUDE.md`.

## Cosa non si scrive

- Ciò che il codice dice già da sé (struttura, nomi, firme)
- Ciò che la git history racconta meglio (com'è stato corretto un bug)
- Il resoconto di come si è arrivati a una decisione: si scrive la decisione e il perché,
  non il percorso
- Una voce sproporzionata rispetto alla feature: otto punti per una modifica da due file

## Rimandi

I rimandi sono **link Markdown o percorsi citati in prosa**. Mai la forma
`@percorso/file.md`: in un `CLAUDE.md` quella sintassi non è un link ma un import, e il
file viene caricato automaticamente all'avvio — annullando il motivo stesso di averlo
separato.

## Limite di competenza

Queste regole riguardano **come** si scrive, mai **cosa**. Il merito di una decisione lo
può giudicare solo chi ha assistito al lavoro. Un agente che applica queste regole
segnala e propone: non scrive contenuto e non cancella nulla di propria iniziativa.
```

- [ ] **Step 2: verificare che la regola sull'import sia scritta in forma non ambigua**

Run:
```bash
grep -n '@percorso' plugins/wm-skills/shared/claude-md-rules.md
```
Expected: la riga che vieta la sintassi, presente una sola volta.

---

### Task 3: Documento di analisi di fattibilità

**Files:**
- Create: `plugins/wm-skills/shared/agentic-feasibility.md`
- Read: `plugins/wm-skills/skills/wm-tag/SKILL.md`, `plugins/wm-skills/skills/wm-review-ticket/SKILL.md`

**Interfaces:**
- Consumes: i tipi di delega definiti in Task 1.
- Produces: la mappa che giustifica le scelte dei Task 4-8 e il criterio per classificare una fase futura.

- [ ] **Step 1: leggere per intero le due skill non ancora esaminate**

Run:
```bash
grep -n '^## \|^### ' plugins/wm-skills/skills/wm-tag/SKILL.md plugins/wm-skills/skills/wm-review-ticket/SKILL.md
```
Serve l'elenco reale delle fasi: la tabella dello Step 2 non va scritta a memoria.

- [ ] **Step 2: scrivere il documento**

Struttura obbligatoria:

```markdown
# Quali fasi si delegano, e perché

## Il criterio

Una fase si delega quando ha **input piccolo, output piccolo, lavoro intermedio grande**
e non richiede interazione diretta con il dev.

Due controlli prima di dire sì:
1. **L'agente può concludere da solo?** Se la conclusione dipende da qualcosa che esiste
   solo nel dialogo (una decisione presa, una responsabilità assunta), no.
2. **Cosa succede se sbaglia e nessuno se ne accorge?** Se alimenta una decisione di
   sicurezza o un registro di responsabilità, no.

## wm-plan

| Fase | Delegabile | Motivo |
|---|---|---|
| `ticket` | no | dialogo con il dev, scritture su Orchestrator da confermare |
| `environment-setup` | **sì** (`wm-env-detect`) | molti comandi, esito = pochi flag |
| `docker-check` | no | effetti collaterali sul sistema del dev (`docker compose stop`) |
| `init-context` | no | il `CLAUDE.md` serve tutto e per tutto il workflow: delegarlo = riassumerlo |
| `reverse-interaction` | **sì, la ricerca** (`wm-codebase-research`) | il dialogo resta nel principale, la ricerca no |
| `overview` | no | nasce dal dialogo, la scrive chi l'ha condotto (oc:8282) |
| `challenge` | **già delegata, cieca** | indipendenza di giudizio |
| `estimation` | **sì, cieca** (`wm-estimate`) | chi ha scritto l'overview stima ottimista in modo sistematico |
| `write-plan` | no | delega già a `superpowers:writing-plans`; il piano torna comunque nel principale |
| `execution: implementation` | già delegata | a `superpowers` |
| `review-gate: subagent` | **già delegata, cieca** | indipendenza di giudizio |
| `review-gate: phpstan-check` | **no** | hard-block di sicurezza (oc:8341): un agente in mezzo può trasformare un fallimento in un pass silenzioso |
| `notes` | no | registra decisioni e responsabilità che esistono solo nel dialogo |
| `update-context` | **sì, il controllo** (`wm-context-guard`) | il testo lo scrive il principale, la verifica di forma no |

## wm-tag e wm-review-ticket

[una tabella per skill, con le stesse colonne, scritta dopo lo Step 1]

## Classificare una fase nuova

Applicare il criterio e i due controlli. In caso di dubbio, **non delegare**: il costo di
una fase non delegata è context; il costo di una delega sbagliata è un errore invisibile
a valle.
```

- [ ] **Step 3: verificare la copertura**

Run:
```bash
for f in wm-plan wm-tag wm-review-ticket; do
  echo "== $f"
  grep -c "^## Fase\|^### " plugins/wm-skills/skills/$f/SKILL.md
done
```
Expected: ogni fase elencata nelle skill compare in una riga della tabella corrispondente. Le fasi assenti vanno aggiunte prima di chiudere il task.

---

### Task 4: Agente wm-codebase-research

**Files:**
- Create: `plugins/wm-skills/agents/wm-codebase-research.md`

**Interfaces:**
- Consumes: `## Contratto di ritorno`, `## Tetto all'output`, `## Fallback` da Task 1.
- Produces: il dossier con prove verbatim consumato da `wm-plan` in Task 9.

- [ ] **Step 1: creare l'agente**

```markdown
---
name: wm-codebase-research
description: Usa quando una skill Webmapp deve rispondere a domande sul codice o sul database di un repo senza portare il materiale letto nel context principale. Restituisce solo conclusioni, ciascuna con una prova verbatim verificabile.
model: sonnet
tools: Bash, Read, Grep, Glob
---

Rispondi a domande su questo repo leggendo codice, migration, config e — dove esiste —
interrogando il database locale. Restituisci **solo le conclusioni**, mai il materiale letto.

## Formato obbligatorio della risposta

Per ogni domanda ricevuta, esattamente questo blocco:

```
Domanda: <la domanda, ripetuta>
Risposta: <una o due frasi>
Fonte: <percorso/file>:<riga-inizio>-<riga-fine>
Estratto: <testo verbatim di quelle righe, copiato senza modifiche>
```

Regole non negoziabili:

- **Ogni affermazione ha una prova.** Nessuna eccezione: un'affermazione senza `Fonte` ed
  `Estratto` viene scartata da chi riceve, non discussa.
- **L'estratto è verbatim.** Copiato dal file, non riscritto né riassunto. Chi riceve lo
  verifica con `sed -n '<inizio>,<fine>p' <file>` e lo confronta: se non combacia, l'intero
  dossier è inattendibile.
- **Non rispondere a memoria.** Se non hai aperto il file, non hai la prova, e senza prova
  non si risponde.
- **Se non trovi la risposta**, scrivi `Risposta: non determinabile dal repo` e spiega in una
  riga dove hai cercato. È un esito legittimo e utile: dice al principale che quella domanda
  va fatta al dev.

## Tetto

Massimo **40 righe complessive**. Se le domande ricevute non ci stanno, rispondi a quelle che
ci stanno e dichiara in fondo quali restano aperte: meglio un dossier corto e completo su
metà domande che uno lungo che nessuno legge.

## Fallimento

Se non riesci a leggere il repo (percorso inesistente, permessi), scrivi
`RICERCA FALLITA: <motivo>` come unica riga. Chi ti ha chiamato rifarà la ricerca nel
context principale: non inventare una risposta plausibile.
```

- [ ] **Step 2: verificare il frontmatter**

Run:
```bash
head -8 plugins/wm-skills/agents/wm-codebase-research.md
claude plugin validate .
```
Expected: frontmatter con `name` uguale al nome del file, `description` che inizia con "Usa quando"; validazione senza errori.

- [ ] **Step 3: provare il protocollo di verifica su un caso reale**

Run:
```bash
sed -n '1,5p' plugins/wm-skills/shared/agent-delegation.md
```
Expected: il comando stampa le righe richieste. È il comando che il context principale userà per verificare ogni `Estratto`: se qui non funziona, il protocollo non è applicabile.

---

### Task 5: Agente wm-env-detect

**Files:**
- Create: `plugins/wm-skills/agents/wm-env-detect.md`

**Interfaces:**
- Consumes: Task 1.
- Produces: i flag e le variabili consumati da `wm-plan` in Task 9. **Nomi esatti**, usati più avanti nel workflow: `stack_type`, `has_docker`, `has_submodules`, `stack_ui`, `has_phpstan_ci`, `DOCKER_PROJECT_DIR_NAME`.

- [ ] **Step 1: creare l'agente**

```markdown
---
name: wm-env-detect
description: Usa quando una skill Webmapp deve rilevare stack, Docker, submodule e presenza di PHPStan in un repo, senza portare nel context l'output dei comandi di rilevamento.
model: haiku
tools: Bash, Read, Glob
---

Rileva la configurazione del repo corrente ed esegui i check elencati sotto.
Restituisci **solo i valori risolti**, mai l'output dei comandi.

## Check da eseguire

```bash
grep -s "DOCKER_PROJECT_DIR_NAME" .env
cat package.json 2>/dev/null | jq -r '(.dependencies // {}) + (.devDependencies // {}) | keys[]' | grep -E "^(@angular/core|vue|@vue/core|react)$"
git submodule status 2>/dev/null
ls resources/views/ resources/js/components/ 2>/dev/null | head -3
ls phpstan.neon phpstan.neon.dist 2>/dev/null
grep -ilr "phpstan" .github/workflows/ 2>/dev/null
```

## Formato obbligatorio della risposta

Esattamente queste righe, sempre tutte, nell'ordine dato. I nomi sono un contratto: chi
ti chiama li usa molto più avanti nel workflow, e un nome diverso rompe una fase che non
vedi.

```
stack_type: laravel | frontend | fullstack | other
has_docker: true | false
has_submodules: true | false
stack_ui: angular | vue | react | laravel-blade | false
has_phpstan_ci: true | false
DOCKER_PROJECT_DIR_NAME: <valore letto dal .env, oppure "assente">
submodule: <nome> → <scopo desumibile dal path>   (una riga per submodule, oppure "nessuno")
```

`has_phpstan_ci` è `true` **solo se** esiste `phpstan.neon` o `phpstan.neon.dist` nella root
**e** la keyword `phpstan` compare in almeno un file `.github/workflows/*.yml`.

`DOCKER_PROJECT_DIR_NAME` va restituito con il **valore**, non solo con la presenza: viene
usato a fine workflow per scegliere il servizio Docker su cui lanciare PHPStan, e un valore
mancante lì fa scegliere il servizio sbagliato.

## Tetto

Massimo **15 righe**. Non commentare i risultati, non spiegare i comandi.

## Fallimento

Un check che fallisce non è un errore: è l'informazione che quel segnale non c'è. Restituisci
il valore negativo (`false`, `assente`, `nessuno`) e prosegui con gli altri.

Se non riesci a eseguire alcun comando, scrivi `RILEVAMENTO FALLITO: <motivo>` come unica
riga: chi ti ha chiamato proseguirà con `⚠️ Environment setup non disponibile`, che è il
comportamento fail-soft già previsto per questa fase.
```

- [ ] **Step 2: verificare**

Run:
```bash
claude plugin validate .
grep -c "DOCKER_PROJECT_DIR_NAME" plugins/wm-skills/agents/wm-env-detect.md
```
Expected: validazione pulita; almeno 3 occorrenze (check, formato, spiegazione).

---

### Task 6: Agente wm-estimate

**Files:**
- Create: `plugins/wm-skills/agents/wm-estimate.md`

**Interfaces:**
- Consumes: Task 1 (delega **cieca**).
- Produces: la tabella di stima consumata da `wm-plan` in Task 9.

- [ ] **Step 1: creare l'agente**

```markdown
---
name: wm-estimate
description: Usa quando va stimato il costo di una feature Webmapp già documentata in overview e piano, con un giudizio indipendente da chi li ha scritti.
model: sonnet
tools: Bash, Read, Grep, Glob
---

Ricevi i percorsi di `overview.md` e `plan.md`. Leggili e stima il costo della feature.

**Non ricevi, e non devi chiedere, alcun riassunto della conversazione che li ha prodotti.**
È deliberato: chi ha condotto quel dialogo ha appena finito di convincersi che il problema è
chiaro, e stima ottimista in modo sistematico. Il tuo valore sta nel non aver vissuto quella
conversazione.

## Cosa stimare

**A. Tempo di esecuzione, per componente.** Per ogni componente del piano, il tempo che
impiegheresti tu a scrivere codice e test seguendo il piano. Nessun buffer percentuale qui.

**B. Buffer di novità del dominio** — un valore assoluto, una sola volta sull'intera feature,
mai una percentuale per componente.

Non dichiararlo a giudizio: **verificalo**. Cerca nel codebase un pattern equivalente a quello
che il piano descrive.

```bash
grep -rl "<simbolo o pattern equivalente>" --include='*' . | head -20
```

- pattern già presente altrove nel codebase → **+20-30 min**
- prima volta nel suo genere, nessun precedente locale → **+1-2h**

Dove collocarsi dentro la forbetta, con due numeri oggettivi:
1. quanti file elenca "Moduli toccati" nell'overview (1-3 → basso, 4-10 → centro, oltre 10 → alto)
2. quanti file **leggono** il simbolo che viene modificato:
   ```bash
   grep -rl "<NomeSimbolo>" app/ database/ routes/ resources/ 2>/dev/null | wc -l
   ```
   Se i lettori sono molti più dei file toccati, collocarsi in alto anche se la scrittura è minima.

**C. Buffer di integrazione trasversale** — 5% sul totale di A, solo se i componenti sono più di uno.

**D. Deliverable extra** — documentazione utente, screenshot, guide: solo se il piano li nomina.

## Formato obbligatorio della risposta

```
| Componente | Ore | Note |
|---|---|---|
| <componente> (tempo di esecuzione) | <X>h | <motivazione tecnica> |
| Buffer integrazione trasversale | <Z>h | 5% sul totale di A |
| Buffer novità di dominio | <B>h | <cosa hai cercato, con il comando, e cosa hai trovato o non trovato> |

Totale stimato: <S>h
Confidenza: alta | media | bassa
```

Regole:
- Non stimare meno di 0.5h per una feature che tocca più di un file
- Confidenza **bassa** per default se il buffer novità è "prima nel suo genere": overview e
  challenge riducono il rischio sui requisiti, non i difetti che emergono solo usando la cosa
- Il buffer di novità è **uno**, in valore assoluto, sull'intera feature
- Non includere il tempo di pianificazione: è misurato dal context principale, non da te
- Non proporre un totale comprensivo della pianificazione: non hai i dati per farlo

## Tetto

Massimo **25 righe**, tabella inclusa.

## Fallimento

Se uno dei due file non esiste o è vuoto, scrivi `STIMA NON POSSIBILE: <motivo>` come unica
riga. Non stimare su un piano che non hai letto.
```

- [ ] **Step 2: verificare l'assenza di riferimenti alla conversazione**

Run:
```bash
grep -niE 'conversazione|dialogo|riassunto' plugins/wm-skills/agents/wm-estimate.md
```
Expected: solo le righe che *vietano* di riceverli. Nessuna riga che ne preveda l'uso.

- [ ] **Step 3: validare**

Run: `claude plugin validate .`
Expected: nessun errore.

---

### Task 7: Agente wm-context-guard

**Files:**
- Create: `plugins/wm-skills/agents/wm-context-guard.md`

**Interfaces:**
- Consumes: `plugins/wm-skills/shared/claude-md-rules.md` (Task 2).
- Produces: i rilievi consumati da `wm-plan` in `update-context` (Task 9).

- [ ] **Step 1: creare l'agente**

```markdown
---
name: wm-context-guard
description: Usa prima di scrivere in un CLAUDE.md per verificare che l'aggiunta proposta non ripeta, non contraddica e non appesantisca quanto già presente.
model: sonnet
tools: Read, Grep, Glob, Bash
---

Ricevi il percorso di un `CLAUDE.md` e il testo che sta per esservi aggiunto.

Prima di ogni altra cosa, leggi `plugins/wm-skills/shared/claude-md-rules.md`: è il corpo di
regole che applichi, condiviso con `wm-context-doctor`. Non applicare regole tue.

Poi leggi **per intero** il `CLAUDE.md` indicato e confronta l'aggiunta con quanto già c'è.

## Cosa giudichi

**Come** si scrive, mai **cosa**. Il merito di una decisione lo può giudicare solo chi ha
assistito al lavoro: tu no, ed è il motivo per cui il tuo giudizio vale. Non contestare una
scelta tecnica, non chiedere perché è stata presa.

Cerca i quattro rilievi definiti nelle regole: **ripetizione**, **contraddizione**,
**duplicato dal codice**, **fuori posto**. In più: sproporzione rispetto alla dimensione
della feature, e rimandi scritti come `@percorso` invece che come link.

## Formato obbligatorio della risposta

Se non hai rilievi:

```
NESSUN RILIEVO
```

Altrimenti, un blocco per rilievo:

```
[<tipo>] <una riga che dice cosa>
Riferimento: <file>:<riga> — <estratto verbatim della voce esistente coinvolta>
Proposta: <cosa fare, in una riga>
```

Ogni rilievo che cita una voce esistente porta **riga ed estratto verbatim**: senza, chi
riceve non può verificarlo senza rileggere tutto il file, che è esattamente ciò che si
voleva evitare.

## Cosa non fai mai

- Non modifichi il `CLAUDE.md`, per nessun motivo
- Non cancelli né riscrivi una voce esistente: **proponi**
- Non decidi tu una contraddizione: la segnali, la conferma è del dev

## Tetto

Massimo **30 righe**. Se i rilievi sono molti, riporta i più gravi e chiudi con
`altri N rilievi minori non riportati`.

## Fallimento

Se il `CLAUDE.md` non esiste, scrivi `NESSUN RILIEVO — file non presente`: su un repo senza
`CLAUDE.md` non c'è nulla da controllare e la scrittura procede normalmente.
```

- [ ] **Step 2: provarlo sul caso reale noto**

Il `CLAUDE.md` di questo repo contiene una contraddizione verificabile: `oc:8283` dichiara
l'URL dell'Artifact versionato in `CLAUDE.md`, un blocco successivo lo sposta in `SKILL.md`.

Run:
```bash
grep -n "URL Artifact versionato in\|URL dell'Artifact è ora un valore statico" CLAUDE.md
```
Expected: due righe, in due blocchi diversi. È il caso che l'agente deve saper classificare
come `[contraddizione]`. Se l'agente provato su questo file non lo rileva, il prompt va
corretto prima di chiudere il task.

---

### Task 8: Agente wm-context-doctor

**Files:**
- Create: `plugins/wm-skills/agents/wm-context-doctor.md`

**Interfaces:**
- Consumes: `plugins/wm-skills/shared/claude-md-rules.md` (Task 2) — **le stesse regole di Task 7**.
- Produces: un piano di riordino che il dev approva; nessuna modifica diretta.

- [ ] **Step 1: creare l'agente**

```markdown
---
name: wm-context-doctor
description: Usa quando il CLAUDE.md di un repo va esaminato nel suo insieme — contraddizioni accumulate, voci obsolete, sezioni cresciute troppo — e serve un piano di riordino da approvare.
model: sonnet
tools: Read, Grep, Glob, Bash
---

Ricevi il percorso di un `CLAUDE.md`. Esaminalo **nel suo insieme** e proponi un piano di
riordino.

Prima di ogni altra cosa, leggi `plugins/wm-skills/shared/claude-md-rules.md`: applichi lo
stesso corpo di regole di `wm-context-guard`. La differenza fra voi due non è cosa cercate,
è quando guardate: lui controlla un pezzo nuovo mentre entra, tu guardi tutto quello che è
già dentro.

Questo ti fa vedere una cosa che lui non può vedere: le contraddizioni **maturate nel tempo**
fra voci scritte a mesi di distanza, ciascuna corretta quando è stata scritta.

## Cosa cerchi

I quattro rilievi delle regole condivise, più:

- **voci superate** — una decisione successiva ha reso obsoleta una precedente, ma entrambe
  sono presenti e il lettore deve arrivare in fondo per sapere quale vale
- **sezioni fuori scala** — una sezione cresciuta al punto da meritare un file proprio
- **materiale operativo** — procedure passo-passo che servono solo a chi sta eseguendo
  quella procedura

Misura la dimensione delle sezioni, non fidarti dell'impressione:

```bash
wc -c CLAUDE.md
awk '/^## /{name=$0; next} {len[name]+=length($0)} END {for (n in len) print len[n], n}' CLAUDE.md | sort -rn
```

## Formato obbligatorio della risposta

```
Stato: <dimensione totale>, <numero sezioni>, sezione più pesante: <nome> (<dimensione>)

Interventi proposti, dal più utile:

1. [<tipo>] <cosa>
   Riferimento: <righe> — <estratto verbatim>
   Proposta: <azione concreta>
   Rischio se non fatto: <una riga>

2. ...
```

## Cosa non fai mai

- **Non modifichi nulla.** Produci un piano, non un risultato. Il `CLAUDE.md` governa il
  comportamento di Claude sull'intero repo: non si riscrive senza che un umano abbia letto
  cosa cambia
- Non cancelli una voce perché ti sembra vecchia: proponi, indicando cosa l'ha superata
- Non giudichi il merito delle decisioni, solo la loro forma e collocazione

## Tetto

Massimo **50 righe**. Sei invocato esplicitamente e produci un piano, quindi hai più spazio
degli altri agenti — ma un piano che nessuno legge non viene eseguito.

## Fallimento

Se il file non esiste: `NESSUN INTERVENTO — CLAUDE.md non presente nel repo indicato.`
```

- [ ] **Step 2: verificare che i due agenti leggano le stesse regole**

Run:
```bash
grep -l "claude-md-rules.md" plugins/wm-skills/agents/*.md
```
Expected: esattamente due file — `wm-context-guard.md` e `wm-context-doctor.md`.

- [ ] **Step 3: verificare che nessuno dei due possa scrivere**

Run:
```bash
grep -n "^tools:" plugins/wm-skills/agents/wm-context-guard.md plugins/wm-skills/agents/wm-context-doctor.md
```
Expected: nessuno dei due elenca `Write` o `Edit` fra i tool. Se li elenca, rimuoverli: la
garanzia "propone, non esegue" dev'essere strutturale, non solo scritta nel prompt.

---

### Task 9: Integrazione delle deleghe in wm-plan

**Files:**
- Modify: `plugins/wm-skills/skills/wm-plan/SKILL.md`

**Interfaces:**
- Consumes: i cinque agenti (Task 4-8) e i tre file condivisi (Task 1-3).
- Produces: la skill che li invoca. È il task più grande del piano: si esegue una sotto-fase alla volta, verificando dopo ciascuna.

- [ ] **Step 1: rimuovere i rimandi alle skill inesistenti**

`SKILL.md` e `CLAUDE.md` citano `wm-skills:our-code-style`, `wm-skills:our-pr-checklist` e
`wm-skills:our-deploy-post-merge`, che non esistono in `plugins/wm-skills/skills/`.

Run, per vedere l'estensione del problema:
```bash
cd /Users/bongiu/Documents/claude-marketplace
grep -rn "our-code-style\|our-pr-checklist\|our-deploy-post-merge" --include='*.md' .
```

Rimuovere ogni occorrenza: la riga della sezione `## Composizione con altre skill Webmapp`,
il rimando in `Fase: write-plan` ("Durante la scrittura del piano applica la skill…") e le
righe corrispondenti in `CLAUDE.md`. Non sostituirle con altre skill: vanno tolte e basta.

Verifica:
```bash
grep -rn "our-code-style\|our-pr-checklist\|our-deploy-post-merge" --include='*.md' . | wc -l
```
Expected: `0`.

- [ ] **Step 2: aggiungere il richiamo al meccanismo condiviso**

In `SKILL.md`, dopo la sezione `## Orchestrator`, inserire:

```markdown
## Delega ad agenti

Il meccanismo — tipi di delega, contratto di ritorno, tetto all'output, fallback, misura del
context — è descritto in `plugins/wm-skills/shared/agent-delegation.md`. Leggerlo alla prima
delega della sessione, non riscriverne il contenuto qui.

Fasi che delegano, e a chi:

| Fase | Agente | Tipo |
|---|---|---|
| `environment-setup` | `wm-env-detect` | informata |
| `reverse-interaction` | `wm-codebase-research` | informata |
| `challenge` | subagente già previsto | **cieca** |
| `estimation` | `wm-estimate` | **cieca** |
| `review-gate` | subagente già previsto | **cieca** |
| `update-context` | `wm-context-guard` | informata |

Le deleghe sono **statiche**: si applicano sempre in quelle fasi. Nessuna soglia, nessuna
attivazione basata sulla misura del context.

Restano nel context principale, per scelta motivata in
`plugins/wm-skills/shared/agentic-feasibility.md`: `init-context`, `overview`, `write-plan`,
`notes`, `review-gate: phpstan-check`, `environment-setup: docker-check` e ogni dialogo con
il dev.
```

- [ ] **Step 3: delegare il rilevamento in `environment-setup: project-detection`**

Sostituire il blocco di comandi e la tabella dei flag con l'invocazione dell'agente,
mantenendo invariata la tabella dei significati (serve a chi legge la skill):

```markdown
Invoca l'agente `wm-env-detect` sul repo corrente. Restituisce i valori già risolti nel
formato dichiarato nel suo prompt: `stack_type`, `has_docker`, `has_submodules`, `stack_ui`,
`has_phpstan_ci`, `DOCKER_PROJECT_DIR_NAME`, elenco submodule.

Tieni questi valori attivi per tutto il workflow: `DOCKER_PROJECT_DIR_NAME` in particolare
viene riusato a fine workflow in `review-gate: phpstan-check`.

Se l'agente restituisce `RILEVAMENTO FALLITO`, mostra
`⚠️ Environment setup non disponibile — proseguo con il workflow.` e prosegui: questa fase
è fail-soft e non blocca mai.
```

`environment-setup: docker-check` **non va modificata**: resta nel principale.

- [ ] **Step 4: delegare la ricerca in `reverse-interaction`**

Il protocollo in tre passi ("cerca nel codice / cerca nel db / solo allora chiedi") resta,
ma i primi due passi si delegano. Sostituire il blocco del protocollo con:

```markdown
- **Non chiedere ciò che puoi leggere nel codice o nel database.** La ricerca si delega
  all'agente `wm-codebase-research`, che legge nel proprio context e restituisce solo
  conclusioni con prova verbatim.

  **Una chiamata sola, a inizio fase**, con tutte le domande prevedibili dal ticket. Un
  agente per ogni domanda renderebbe il dialogo a scatti, ed è il difetto peggiore in una
  fase che vive di continuità. Se durante il dialogo emerge un buco imprevisto, una chiamata
  puntuale in più è ammessa come eccezione, non come regola.

  **Verifica obbligatoria di ogni prova ricevuta**, prima di usarne il contenuto:

  ```bash
  sed -n '<riga-inizio>,<riga-fine>p' <file>
  ```

  Confronta l'output con l'`Estratto` del dossier. Se anche una sola prova non combacia, il
  dossier è inattendibile: rifai la ricerca nel context principale e segnalalo al dev.

  Se l'agente risponde `Risposta: non determinabile dal repo`, quella domanda va fatta al
  dev: è l'esito che rende utile la delega, non un fallimento.

  Se l'agente restituisce `RICERCA FALLITA`, esegui tu la ricerca nel context principale.
```

Resta invariato l'obbligo di dichiarare prima di ogni domanda dove si è cercato — con la
differenza che ora la citazione viene dal dossier.

- [ ] **Step 5: delegare la stima in `estimation`**

Sostituire `### estimation: analisi` con:

```markdown
### estimation: analisi

La stima è prodotta dall'agente **cieco** `wm-estimate`, che riceve i percorsi di
`overview.md` e `plan.md` e **nient'altro** — nessun riassunto della conversazione, stesso
principio di `challenge: subagent`. Chi ha condotto il dialogo e scritto l'overview stima
ottimista in modo sistematico: il valore dell'agente sta nel non aver vissuto quella
conversazione.

Restano nel context principale:

- **la quota misurata** della pianificazione, che l'agente non può conoscere:

  ```bash
  NOW=$(date -u +"%Y-%m-%dT%H:%M:%S%z")
  ```

  Differenza con `planning_start_at` registrato in `Fase: ticket`. Se non è stato registrato,
  vale il fallback già previsto: dichiararlo al dev e presentare solo la quota stimata.

- **la conferma del dev**, invariata.

Presenta al dev la tabella dell'agente con in testa la riga della pianificazione misurata, e
chiudi sempre con la scomposizione, mai un numero unico fuso:

> **Misurato: \<M\>h + Stimato: \<S\>h = Totale: \<N\>h**

Se l'agente restituisce `STIMA NON POSSIBILE`, produci tu la stima nel context principale
seguendo gli stessi criteri (tempo di esecuzione per componente, buffer di novità di dominio
in valore assoluto, 5% di integrazione trasversale) e dichiara al dev che la stima non è
indipendente.
```

`### estimation: conferma` e `### estimation: scrittura su Orchestrator` restano invariate.

- [ ] **Step 6: delegare il controllo in `update-context`**

In fondo a `## Fase: update-context`, prima di "Mostra le modifiche al `CLAUDE.md`
all'utente", inserire:

```markdown
**Controllo di forma prima di scrivere.**

Prepara il testo da aggiungere, poi invoca `wm-context-guard` passandogli il percorso del
`CLAUDE.md` e il testo proposto. L'agente applica le regole di
`plugins/wm-skills/shared/claude-md-rules.md` e restituisce i rilievi: ripetizione,
contraddizione, duplicato dal codice, fuori posto.

Verifica ogni rilievo che cita una voce esistente con l'estratto e la riga forniti, poi
correggi il testo di conseguenza. Una **contraddizione** non si risolve mai da soli: va
portata al dev, perché rimuovere o marcare come superata una voce esistente tocca testo già
approvato.

Se l'agente risponde `NESSUN RILIEVO`, procedi. Se fallisce, scrivi comunque: il controllo è
un ausilio, non un gate.

La forma in cui si scrive resta quella attuale (una riga in "Feature disponibili", un blocco
in "Decisioni architetturali"): la struttura a indice è materia di **oc:8528**.
```

- [ ] **Step 7: aggiungere la misura del context all'header di sessione**

In `## Header di sessione`, dopo `### header: diagramma`, aggiungere:

```markdown
### header: context

Mostra l'occupazione del context in **valore assoluto**, come informazione. Non attiva e non
disattiva nulla.

```bash
SESSION_ID=$(basename "$(dirname "$CLAUDE_SCRATCHPAD_DIR")" 2>/dev/null)
TRANSCRIPT=$(ls -t ~/.claude/projects/*/"$SESSION_ID".jsonl 2>/dev/null | head -1)
[ -z "$TRANSCRIPT" ] && TRANSCRIPT=$(ls -t ~/.claude/projects/"$(pwd | tr '/' '-')"/*.jsonl 2>/dev/null | head -1)
grep -v '"isSidechain":true' "$TRANSCRIPT" 2>/dev/null | grep -o '"usage":{[^}]*}' | tail -1
```

Somma `input_tokens`, `cache_creation_input_tokens` e `cache_read_input_tokens`, e mostra:

`🧮 Context occupato: ~<N>k token`

- **Mai una percentuale:** la dimensione della finestra non è leggibile dal transcript e
  andrebbe indovinata (200k o 1M, e il modello può cambiare a metà sessione).
- **Mai un numero inventato:** se il transcript non è raggiungibile o il formato è inatteso,
  mostra `🧮 Misura del context non disponibile.` e prosegui.
- Il filtro `isSidechain` esclude le righe dei subagenti che condividono il transcript.
```

- [ ] **Step 8: filtrare l'output di PHPStan senza delegarlo**

In `#### review-gate: phpstan-check`, il comando resta nel principale — è un hard-block di
sicurezza (oc:8341) e un agente in mezzo potrebbe trasformare un fallimento in un pass
silenzioso. Si riduce solo il volume dell'output:

```bash
if [ "$has_docker" = "true" ]; then
  timeout 300 docker compose -f local.compose.yml exec -T "$DOCKER_PROJECT_DIR_NAME" vendor/bin/phpstan analyse --error-format=json > /tmp/wm-phpstan.json
else
  timeout 300 vendor/bin/phpstan analyse --error-format=json > /tmp/wm-phpstan.json
fi
PHPSTAN_EXIT=$?
jq -r '.files | to_entries[] | "\(.key): \(.value.messages | length) errori"' /tmp/wm-phpstan.json 2>/dev/null
```

L'output completo resta sul filesystem e si consulta solo per gli errori che riguardano il
diff. Aggiungere una riga esplicita:

```markdown
**Questa fase non è delegabile.** È un hard-block: un agente che fallisce e restituisce
"nessun errore" trasformerebbe il blocco in un pass silenzioso. Vedi
`plugins/wm-skills/shared/agentic-feasibility.md`.
```

- [ ] **Step 9: verificare che le deleghe cieche non siano state contaminate**

Run:
```bash
cd /Users/bongiu/Documents/claude-marketplace
sed -n '/### challenge: subagent/,/### challenge: dialog/p' plugins/wm-skills/skills/wm-plan/SKILL.md | grep -c "nient'altro"
sed -n '/#### review-gate: subagent/,/#### review-gate: phpstan-check/p' plugins/wm-skills/skills/wm-plan/SKILL.md | grep -c "nient'altro"
```
Expected: `1` per entrambe. Le due deleghe cieche devono conservare la loro clausola di
isolamento: se il conteggio è `0`, sono state modificate e vanno ripristinate.

- [ ] **Step 10: validare**

Run:
```bash
claude plugin validate .
```
Expected: nessun errore.

---

### Task 10: wm-review-ticket e wm-tag

**Files:**
- Modify: `plugins/wm-skills/skills/wm-review-ticket/SKILL.md`
- Verify: `plugins/wm-skills/skills/wm-tag/SKILL.md`

**Interfaces:**
- Consumes: Task 1, Task 3, Task 4.

- [ ] **Step 1: collegare wm-review-ticket al meccanismo condiviso**

È la skill che legge più codice di tutte per produrre un giudizio corto. Aggiungere, dopo la
sezione introduttiva:

```markdown
## Delega ad agenti

Il meccanismo è in `plugins/wm-skills/shared/agent-delegation.md`. Dove questa skill deve
rispondere a una domanda sul codice per valutare una modifica, usa `wm-codebase-research`
invece di leggere i file nel context principale, e verifica ogni prova ricevuta con
`sed -n '<inizio>,<fine>p' <file>` prima di usarne il contenuto.

Non delegare il **giudizio** sulla review: la delega copre la raccolta delle informazioni,
non la valutazione.
```

I punti esatti in cui inserire la delega si ricavano dalla tabella scritta in Task 3.

- [ ] **Step 2: verificare wm-tag in tag-mode**

`wm-tag` invoca `wm-plan` in tag-mode, quindi eredita le deleghe comunque. Verificare che il
flusso tag-mode dichiarato in `wm-plan` non sia stato rotto:

Run:
```bash
cd /Users/bongiu/Documents/claude-marketplace
sed -n '/## Modalità tag-mode/,/## Orchestrator/p' plugins/wm-skills/skills/wm-plan/SKILL.md | grep -n "environment-setup\|reverse-interaction\|estimation"
```
Expected: le fasi elencate nel flusso tag-mode sono ancora quelle previste, e sono proprio
fasi che ora delegano — il che è corretto e voluto.

Modificare `wm-tag` **solo se** la verifica mostra un'incoerenza reale. Nessuna modifica
preventiva.

- [ ] **Step 3: validare**

Run: `claude plugin validate .`
Expected: nessun errore.

---

### Task 11: CLAUDE.md, diagramma, release e commit

**Files:**
- Modify: `CLAUDE.md`
- Modify: `docs/wm-plan-diagram/index.html`
- Modify: `plugins/wm-skills/.claude-plugin/plugin.json`
- Modify: `plugins/wm-skills/skills/wm-plan/SKILL.md` (riga versione)
- Modify: `plugins/wm-skills/mcp/internal/version/version.go`
- Create: `docs/features/8527-refactoring-wm-skills-architettura-agentica/notes.md`

- [ ] **Step 1: aggiornare CLAUDE.md**

Tre modifiche:

1. Riga in `## Feature disponibili`:

```markdown
| Refactoring wm-skills verso architettura agentica | oc:8527 | `plugins/wm-skills/agents/`, `plugins/wm-skills/shared/`, le tre `SKILL.md` | Cinque agenti formali (`wm-codebase-research`, `wm-env-detect`, `wm-estimate`, `wm-context-guard`, `wm-context-doctor`); deleghe statiche per fase, nessuna soglia; meccanismo e regole in `shared/` come primo mattone del core condiviso |
```

2. Blocco in cima a `## Decisioni architetturali` — i punti sono tutti nell'overview alla
   sezione "Rischi" e vanno riportati in forma breve: deleghe statiche invece che a soglia
   (il transcript non espone la dimensione della finestra, quindi ogni percentuale avrebbe
   un denominatore indovinato); misura del context solo informativa; PHPStan non delegato
   perché hard-block; nessun agente scrive contenuto; obbligo di prova verbatim verificabile
   a macchina per `wm-codebase-research`; regole dei `CLAUDE.md` in un solo file letto da
   guard e doctor; `wm-estimate` cieco per correggere l'ottimismo sistematico di chi ha
   scritto l'overview; rimozione dei rimandi a tre skill mai esistite.

3. Righe nella tabella `### Coupling tra skill`:

```markdown
| `wm-plan` | `shared/agent-delegation.md` | Contratto di delega: tipi, ritorno, tetto, fallback. Modificarlo tocca tutte le skill che delegano. |
| `wm-context-guard` | `wm-context-doctor` | Leggono lo stesso `shared/claude-md-rules.md`. Le regole si modificano lì, mai in un agente: due copie divergerebbero. |
```

- [ ] **Step 2: aggiornare il diagramma**

Il workflow di `wm-plan` cambia: `environment-setup`, `reverse-interaction`, `estimation` e
`update-context` ora delegano. Aggiornare il **contenuto** di `docs/wm-plan-diagram/index.html`
mantenendo invariati struttura HTML, CSS e stile — il template è congelato (oc:8283).

Ripubblicare l'Artifact **sullo stesso URL** già presente in `SKILL.md` →
`### header: diagramma`. Se il redeploy fallisce:
`⚠️ Impossibile aggiornare l'Artifact del diagramma — potrebbe servire switchare all'account
Claude del team Webmapp.` e proseguire senza bloccare.

- [ ] **Step 3: checklist di release**

> ⚠️ L'implementazione ha deviato da questo task: [notes.md](notes.md#task-11--bump-di-versione-non-eseguito)

```bash
cd /Users/bongiu/Documents/claude-marketplace
grep '"version"' plugins/wm-skills/.claude-plugin/plugin.json
grep -n "Versione installata" plugins/wm-skills/skills/wm-plan/SKILL.md
grep -n "Version" plugins/wm-skills/mcp/internal/version/version.go
```

Bump **minor** (nuove skill/feature retro-compatibili): `1.3.0` → `1.4.0`. Aggiornare i tre
punti con lo stesso valore, poi ricompilare:

```bash
plugins/wm-skills/mcp/build.sh
```

Il binario compilato è versionato nel repo e va aggiornato ad ogni rilascio.

- [ ] **Step 4: scrivere notes.md**

Il file è obbligatorio. Struttura in `wm-plan` → `Fase: notes`. Deve contenere almeno:

- **Decisioni**: le quattro correzioni di rotta della `Fase: challenge` (soglie eliminate,
  PHPStan non delegato, nessun agente scrive contenuto, obbligo di prova); l'aggiunta di
  `wm-estimate` a piano già impostato; la scelta di agenti formali di plugin invece di prompt
  in file condivisi
- **Follow-up**: i rimandi a `our-code-style`, `our-pr-checklist` e `our-deploy-post-merge`
  sono stati rimossi perché puntavano a skill inesistenti — se quelle skill servono davvero,
  vanno scritte in un ticket dedicato
- **Follow-up**: soglie e calibrazione non esistono più come meccanismo; se in futuro si
  vorrà un'attivazione dinamica servirà prima un modo affidabile di conoscere la dimensione
  della finestra
- Le divergenze dal piano emerse durante l'esecuzione, con i rimandi da `plan.md` secondo la
  regola in `execution: divergenze`

- [ ] **Step 5: verificare gli ancoraggi fra piano e note**

Run:
```bash
cd docs/features/8527-refactoring-wm-skills-architettura-agentica
grep -o 'notes\.md#[a-z0-9-]*' plan.md | sed 's/.*#//' | sort -u > /tmp/wm-ancore-citate
grep '^### ' notes.md | sed 's/^### //' | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9 -]//g; s/ /-/g' | sort -u > /tmp/wm-ancore-esistenti
comm -23 /tmp/wm-ancore-citate /tmp/wm-ancore-esistenti
```
Expected: nessun output. Un rimando rotto non produce errori in Markdown: resta lì e nessuno
se ne accorge.

- [ ] **Step 6: validazione finale**

Run:
```bash
cd /Users/bongiu/Documents/claude-marketplace
claude plugin validate .
ls plugins/wm-skills/agents/
grep -rn "our-code-style\|our-pr-checklist\|our-deploy-post-merge" --include='*.md' . | wc -l
```
Expected: validazione pulita; cinque file in `agents/`; `0` rimandi a skill inesistenti.

- [ ] **Step 7: commit unico — ISTRUZIONE PER IL DEV, NON DA ESEGUIRE**

Un solo commit a lavoro finito, dopo l'approvazione esplicita del dev in
`execution: review-gate`:

```bash
git add plugins/wm-skills/ docs/ CLAUDE.md
git commit -m "$(cat <<'MSG'
feat(oc:8527): agenti formali e delega del materiale grezzo nelle skill wm-skills

Cinque agenti di plugin (wm-codebase-research, wm-env-detect, wm-estimate,
wm-context-guard, wm-context-doctor) assorbono la lettura di codice, ambiente e
CLAUDE.md che finora restava nel context della sessione principale fino a fine
workflow.

Le deleghe sono statiche per fase: nessuna soglia sul context, perché il
transcript non espone la dimensione della finestra e ogni percentuale avrebbe un
denominatore indovinato. La misura resta solo informativa nell'header.

Meccanismo di delega, regole dei CLAUDE.md e analisi di fattibilità vivono in
shared/, letti dalle skill e dagli agenti: primo mattone del core condiviso.

Non delegati per scelta: phpstan-check (hard-block, oc:8341), notes e overview
(registrano ciò che esiste solo nel dialogo), init-context, docker-check.
wm-codebase-research deve provare ogni affermazione con un estratto verbatim
verificabile a macchina.

Rimossi i rimandi a our-code-style, our-pr-checklist e our-deploy-post-merge:
skill mai esistite nel plugin.

Co-Authored-By: Claude Opus 5 (1M context) <noreply@anthropic.com>
Claude-Session: https://claude.ai/code/session_01FNVcu88taiSVQFm5nJ2YdV
MSG
)"
```

Poi PR verso **`develop`** (non `main`).

---

## Self-review

**Copertura della spec** — i 26 requisiti dell'overview sono coperti: meccanismo condiviso e
distinzione cieca/informata (Task 1), tetto e fallback per agente (Task 1 + prompt nei Task
4-8), analisi di fattibilità (Task 3), i cinque agenti (Task 4-8), obbligo di prova e verifica
a macchina (Task 4 Step 1, Task 9 Step 4), contratto nominale delle variabili (Task 5),
regole condivise fra guard e doctor (Task 2, verificato in Task 8 Step 2), nessuna scrittura
dagli agenti (Task 8 Step 3, strutturale via `tools`), deleghe statiche (Task 9 Step 2),
misura informativa in valore assoluto con filtro `isSidechain` (Task 9 Step 7), PHPStan non
delegato (Task 9 Step 8), deleghe cieche intatte (Task 9 Step 9), retrocompatibilità sui repo
non migrati (Task 7 e 8, fallimento su `CLAUDE.md` assente), `wm-review-ticket` e `wm-tag`
(Task 10), `CLAUDE.md` e release (Task 11).

**Requisito senza task: tracciabilità.** L'overview chiede che gli artefatti prodotti in
modalità agentica siano riconoscibili a posteriori, sul modello del marcatore già usato per le
stime. Il piano **non** lo implementa: da quando le deleghe sono statiche, *tutti* gli
artefatti sono prodotti in modalità agentica, e un marcatore presente ovunque non distingue
nulla. La discriminante utile è semmai la versione del plugin, già tracciata da `plugin.json`
e dalla riga in `SKILL.md`. Da confermare con il dev in `execution`: se vuole il marcatore
esplicito comunque, va aggiunto un task.

**Coerenza dei nomi** — i flag di `wm-env-detect` (Task 5) coincidono con quelli usati in Task
9 Step 3 e in `phpstan-check` (`DOCKER_PROJECT_DIR_NAME`). I nomi dei cinque agenti sono
identici fra frontmatter, tabella di Task 9 Step 2 e commit di Task 11. Le sezioni citate di
`agent-delegation.md` e `claude-md-rules.md` esistono con quei titoli esatti nei Task 1 e 2.

**Ordine** — i Task 1-3 producono ciò che i Task 4-8 citano; il Task 9 li invoca tutti; i
Task 10-11 chiudono. Task 9 è volutamente il più grande: tocca un solo file in dieci punti
diversi, e spezzarlo produrrebbe task che non sono verificabili da soli.
