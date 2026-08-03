> Ticket: oc:8341

# Esecuzione automatica PHPStan pre-PR/merge in wm-plan

## Cosa cambia

`execution: review-gate` in `wm-plan` esegue automaticamente PHPStan quando il repo target è un
backend Laravel con PHPStan configurato sia come file di config (`phpstan.neon` o
`phpstan.neon.dist` nella root) sia come step riconoscibile nella CI GitHub Actions
(`.github/workflows/*.yml`). Il check gira **prima** della richiesta di conferma commit già
prevista dal gate, non come sotto-fase opzionale separata.

- **Detection CI**: ricerca testuale case-insensitive della keyword `phpstan` nel contenuto di
  tutti i file `.github/workflows/*.yml` — pattern volutamente permissivo: preferiamo un falso
  positivo (il check gira anche se lo step CI è disabilitato) a un falso negativo (il check non
  gira quando servirebbe). Un falso positivo residuo è comunque a basso impatto, perché gestito
  dallo stesso meccanismo di bypass-con-motivazione descritto sotto.
- **Esecuzione**: se `has_docker: true` (flag già impostato in `environment-setup:
  project-detection`), PHPStan viene eseguito dentro il container tramite `docker compose -f
  local.compose.yml exec -T $DOCKER_PROJECT_DIR_NAME vendor/bin/phpstan analyse` — il nome
  servizio è lo stesso `DOCKER_PROJECT_DIR_NAME` già risolto in `environment-setup:
  docker-check`, nessuna euristica di rilevamento servizio aggiuntiva necessaria. Se
  `has_docker: false`, viene eseguito in locale: `vendor/bin/phpstan analyse`.
- **Timeout**: 5 minuti sul comando. Se scade, viene trattato come fallimento infrastrutturale
  (vedi sotto), non come errore di qualità.
- **Errori nuovi vs errori preesistenti**: l'output di PHPStan viene incrociato con `git diff
  --name-only` per distinguere errori sui file effettivamente toccati dal diff corrente da
  errori preesistenti sul resto del codebase (debito tecnico non correlato).
  - **Errori sui file del diff corrente** → blocco di default: nessun commit finché non risolti,
    salvo override esplicito del dev.
  - **Errori preesistenti fuori dal diff corrente** → non bloccano il commit; `wm-plan` propone
    al dev di creare un ticket separato su Orchestrator per tracciare il debito tecnico (stesso
    flusso di `## Orchestrator API → Creazione ticket`, con preview e conferma).
- **Fallimento infrastrutturale** (comando non trovato, crash, timeout, container non
  raggiungibile) → stesso comportamento degli errori sul diff corrente: blocco di default, salvo
  override esplicito del dev.
- **Override**: quando si verifica un blocco (errori sul diff o fallimento infrastrutturale), il
  dev può insistere per procedere comunque. In questo caso `wm-plan`:
  1. Chiede o deduce dal contesto una motivazione per il bypass
  2. Mostra la motivazione in preview al dev, che può confermarla o modificarla
  3. Solo dopo conferma esplicita e distinta dalla conferma generica di commit del gate (non
     basta un "procedi" generico — serve un messaggio che nomina esplicitamente il bypass
     PHPStan), scrive la motivazione in `notes.md` con la responsabilità esplicita del dev.

## Perché

Il team vuole intercettare localmente gli errori PHPStan **prima** di aprire la PR o mergiare,
invece di scoprirli solo quando la CI fallisce — risparmiando un giro di round-trip su GitHub
Actions e riducendo il rischio di PR rosse.

## Requisiti

- [ ] `environment-setup: project-detection` rileva la presenza combinata di
      `phpstan.neon`/`phpstan.neon.dist` nella root **e** della keyword `phpstan`
      (case-insensitive) in almeno un file `.github/workflows/*.yml`, impostando un flag
      `has_phpstan_ci: true/false`
- [ ] `execution: review-gate` esegue PHPStan automaticamente (solo se `has_phpstan_ci: true`)
      prima di richiedere la conferma di commit, con timeout di 5 minuti
- [ ] Esecuzione dentro il container Docker (`$DOCKER_PROJECT_DIR_NAME`) se `has_docker: true`,
      altrimenti in locale con `vendor/bin/phpstan analyse`
- [ ] L'output PHPStan viene incrociato con `git diff --name-only` per separare errori sui file
      del diff corrente da errori preesistenti sul resto del codebase
- [ ] Errori sui file del diff corrente → hard block di default, nessun commit
- [ ] Errori preesistenti fuori dal diff → non bloccano; propone creazione ticket Orchestrator
      dedicato per tracciare il debito tecnico
- [ ] Fallimento infrastrutturale (timeout, comando non trovato, crash) → stesso hard block di
      default degli errori sul diff corrente
- [ ] Override: motivazione richiesta/dedotta → mostrata in preview al dev → conferma esplicita
      e distinta dalla conferma generica di commit → solo allora si procede
- [ ] Ogni bypass viene registrato in `notes.md` (sezione "Decisioni") con la motivazione e la
      responsabilità esplicita attribuita al dev che ha insistito

## Rischi

- **Falsi positivi di detection CI** (step PHPStan disabilitato o commentato che matcha comunque
  la keyword `phpstan`): mitigato perché il costo di un falso positivo è basso — il check gira e
  se il comando fallisce viene gestito come fallimento infrastrutturale con lo stesso bypass.
- **Falsi negativi di detection CI** (PHPStan eseguito tramite wrapper script/Makefile senza la
  keyword `phpstan` visibile nel YAML): accettato come limite noto — il check semplicemente non
  si attiva su questi repo, nessun impatto negativo oltre alla mancata copertura.
- **Debito tecnico pregresso senza distinzione diff-vs-preesistente andrebbe a bloccare ogni
  commit**: mitigato dal cross-check con `git diff --name-only` e dalla proposta di ticket
  separato per errori fuori dal diff.
- **Accoppiamento rigido a un solo stack/tool (Laravel + PHPStan) hardcoded in `review-gate`**:
  rischio architetturale accettato consapevolmente — vedi "Out of scope". Se in futuro serviranno
  altri strumenti di quality-gate, andrà valutata un'astrazione più generica, non in questo ciclo.
- **Timeout di 5 minuti potrebbe essere insufficiente** su repo molto grandi senza cache attiva:
  accettato come valore di partenza regolabile in futuro se si rivela troppo stretto in pratica.

## Out of scope

- Nessuna modifica al comportamento della CI GitHub Actions stessa — il check locale è solo un
  filtro preventivo, non sostituisce né disattiva il controllo remoto
- Nessun tentativo di correggere automaticamente gli errori PHPStan trovati — solo report e
  blocco
- Nessuna estensione ad altri strumenti di analisi statica (es. PHP_CodeSniffer, Psalm) — solo
  PHPStan, come esplicitamente richiesto
- Nessuna configurazione di soglie/livelli PHPStan personalizzati — usa la configurazione già
  presente nel `phpstan.neon`/`phpstan.neon.dist` del progetto target
- Nessuna astrazione generica "quality-gate pluggable" — la logica resta specifica per PHPStan
  in questo ciclo, accoppiamento accettato consapevolmente
- Nessun meccanismo di kill-switch/opt-out persistente per repo — il bypass va motivato ed
  eseguito ad ogni singolo blocco, non c'è disattivazione permanente configurabile

## Moduli toccati

- `plugins/wm-skills/skills/wm-plan/SKILL.md`
  - `environment-setup: project-detection` → nuovo flag `has_phpstan_ci`
  - `execution: review-gate` → nuovo step di esecuzione PHPStan (detection, esecuzione,
    distinzione diff/preesistenti, override con motivazione, log in notes.md, proposta ticket
    per debito tecnico) prima della conferma commit
