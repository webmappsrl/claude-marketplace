# Il gate di review di wm-plan

## Stato attuale

Prima di dichiarare pronto il lavoro, `execution: review-gate` fa due cose.

**Isola il riepilogo del diff in un subagente cieco** — riceve solo path e istruzioni, nessun
riassunto della conversazione — e usa `--find-renames --find-copies`. Il riepilogo è un ausilio
di orientamento: il gate reale resta l'approvazione esplicita del developer sul diff completo.
La revisione formale opzionale è una sotto-fase a sé, `execution: formal-review`.

**Esegue PHPStan** sui repo Laravel che hanno PHPStan in CI. Blocco duro sugli errori che
ricadono sui file del diff corrente e sui fallimenti infrastrutturali (comando assente, crash,
timeout di 5 minuti). Gli errori preesistenti fuori dal diff non bloccano: `wm-plan` propone un
ticket Orchestrator dedicato. Il bypass richiede una conferma distinta dal "procedi" del gate,
con motivazione confermata in preview e registrata in `notes.md` a nome del dev.

## Come ci siamo arrivati

- **Fallimenti infrastrutturali trattati come errori di qualità** (oc:8341): si era considerato
  un fail-soft per comando non trovato, crash e timeout; la decisione finale del dev in
  Fase: reverse-interaction ha uniformato i due casi, per non creare un bypass implicito su
  problemi ambientali che possono mascherare un errore reale.
- **Nessun kill-switch persistente per repo** (oc:8341): un interruttore di disattivazione
  silenzierebbe il check senza tracciabilità. L'unica via d'uscita è l'override motivato.
- **Nessuna euristica di rilevamento del servizio Docker** (oc:8341): si riusa
  `$DOCKER_PROJECT_DIR_NAME` già risolto in `environment-setup: docker-check`, perché un
  `docker compose config --services` potrebbe scegliere un servizio senza `vendor/bin/phpstan`.
- **Detection CI volutamente permissiva** (oc:8341): grep case-insensitive su `phpstan` nei
  workflow. Un falso positivo (step disabilitato ma matchato) costa poco, perché lo assorbe il
  bypass motivato; un falso negativo lascerebbe il gate spento.
- **Timeout a 5 minuti** (oc:8341): valore di partenza stimato su progetti Laravel di dimensioni
  tipiche Webmapp, senza misure sui repo del team — da rivedere se si rivela stretto.
- **Un solo tool, nessun "quality gate pluggable"** (oc:8341): l'astrazione generica per ESLint,
  Pint o Psalm è stata scartata in questo ciclo, non prevista finché non serve davvero.
- **L'isolamento del riepilogo diff non nasce dall'economia di context** (oc:8282): il pivot in
  Fase: reverse-interaction ha scartato l'idea di "svuotare" il context. La motivazione primaria
  è l'indipendenza di giudizio, che si ottiene solo delegando a un subagente cieco.
- **Nessuna soglia e nessuna eccezione per la skill di implementazione usata** (oc:8282): la
  ridondanza con le review per-task di `subagent-driven-development` è un controllo doppio
  voluto.
- **`--find-renames --find-copies` obbligatori** (oc:8282): senza, un file rinominato viene
  descritto come "nuovo + cancellato".
- **La revisione opzionale era un hint testuale** dentro `review-gate` (oc:8068 la formulava già
  come domanda sì/no esplicita); è diventata la sotto-fase `execution: formal-review` con
  oc:8102.
- **Il coordinamento fra sottoagenti paralleli è fuori scope** (oc:8282): lo gestisce
  `subagent-driven-development` con esecuzione sequenziale e ledger di progresso; eventuali
  miglioramenti vanno proposti su `obra/superpowers`, non in `wm-skills`.
