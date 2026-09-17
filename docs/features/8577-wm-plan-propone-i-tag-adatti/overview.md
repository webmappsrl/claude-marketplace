> Ticket: oc:8577

# wm-plan propone i tag adatti a un ticket

## Cosa cambia

`wm-plan` propone al dev i tag Orchestrator da associare al ticket, in **due momenti distinti**:

- **tag di ambiente**, subito dopo `Fase: environment-setup`: quello che si sa dal posto in cui
  si lavora — il repository, l'area tecnica. Costa due ricerche e chiude il caso più frequente,
  il tag del repo che manca.
- **tag di contenuto**, dopo `Fase: overview`: le caratteristiche che hanno segnato *questo*
  lavoro. Il criterio è la ripetibilità — se una caratteristica tornerà in altri lavori, è un
  buon tag; se vale per un ticket solo, non lo è. È qui che nascono i tag nuovi.

Il dev conferma o scarta ogni proposta, una per volta; l'associazione avviene con
`attach_story_to_tag`.

La parte meccanica — cercare, scartare, deduplicare — la fa uno script, non il prompt della
skill. Alla skill resta il dialogo con il dev e la scrittura su Orchestrator.

Oggi non succede niente di tutto questo: l'unico tag che un ticket riceve da sé è il trimestre,
assegnato da Orchestrator alla creazione. Su oc:8577, ticket di questa stessa feature, il tag
`claude-marketplace` (id 648) esiste su Orchestrator ma sul ticket non c'è.

## Perché

I tag sono il modo in cui il lavoro si raggruppa per trimestre, cliente, repo e area. Se
l'associazione resta manuale, il dev deve ricordarsi quali tag esistono mentre sta pensando ad
altro — e infatti non se lo ricorda. Il raggruppamento perde valore proprio dove servirebbe:
rispondere a «cosa abbiamo fatto su questo repo in questo trimestre».

I termini di ambiente si ricavano da dati già disponibili: il repository da `git remote`, l'area
dal repo e dai flag `stack_type` / `stack_ui` che `Fase: environment-setup` imposta già.

I termini di contenuto non si deducono: si leggono nell'overview appena scritta. Su oc:8577 la
caratteristica che segna il lavoro è che tocca `wm-plan` — un tag che oggi non esiste, e che
ricorrerà a ogni lavoro sulla skill. Come `wm-tag` e `wm-skills`: verificato, nessuno dei tre
esiste su Orchestrator.

Il trimestre non è fra queste: lo assegna Orchestrator alla creazione del ticket, quindi non c'è
niente da proporre.

## Requisiti

### Il filtro

- [ ] **Un tag con `description` non nulla non è mai un candidato.** Il filtro si applica prima
      di qualsiasi altra valutazione: `description: null` è un tag etichetta, tutto il resto non
      riguarda questa fase
- [ ] La `description` di un tag non viene mai letta, citata o riassunta: serve solo a decidere
      se scartarlo
- [ ] I tag già presenti sul ticket vengono scartati dai candidati
- [ ] **Un tag che porta un trimestre nel nome non è mai un candidato**, né il trimestre stesso
      (`26Q3`) né i tag che lo contengono (`[26Q3]FORESTAS`). Il trimestre lo assegna
      Orchestrator; un nome che lo contiene è il tag di un cliente, e proporlo sposterebbe il
      lavoro nelle cose di quel cliente

### Lo script

- [ ] La parte meccanica sta in uno script in `plugins/wm-skills/scripts/`, non nel prompt della
      skill — stesso criterio di `claude-md-lint.sh`: è confronto di stringhe, non giudizio
- [ ] Lo script **legge soltanto**: nessuna scrittura su Orchestrator
- [ ] Riceve l'ID del ticket e uno o più termini di ricerca; si autentica con il `token` in
      `~/.config/webmapp/orchestrator-auth.json`, lo stesso file che usa il server MCP
- [ ] Interroga `GET /api/tags?search=<termine>` una volta per termine, mai l'elenco completo
- [ ] Stampa due liste: **candidati** (`id`, `name`, termine che li ha trovati) e **termini
      scoperti**, quelli che non hanno prodotto alcun risultato
- [ ] Prima di dichiarare scoperto un termine, ritenta con le varianti plausibili del nome: con
      e senza parentesi quadre, spazi al posto dei trattini e viceversa, maiuscole ignorate. Se
      emerge un nome simile, è un candidato e non un termine scoperto
- [ ] Le chiamate hanno un tempo massimo di attesa: Orchestrator lento non blocca l'avvio del
      lavoro

### Il dialogo con il dev

- [ ] I **tag di ambiente** si propongono subito dopo `Fase: environment-setup`, prima di
      `Fase: init-context`
- [ ] I **tag di contenuto** si propongono dopo l'approvazione dell'overview, con i termini
      ricavati da ciò che caratterizza il lavoro descritto lì
- [ ] Il dev può aggiungere termini di ricerca propri prima che la ricerca parta
- [ ] **Nessun tag viene associato o creato senza un sì esplicito del dev, uno per uno.** Niente
      approvazione in blocco, niente silenzio-assenso
- [ ] Ogni candidato è mostrato con la caratteristica da cui proviene e il termine di ricerca che
      l'ha trovato
- [ ] L'associazione usa `attach_story_to_tag`, un tag per volta, mai il campo `tags` di
      `update_story`, che sostituirebbe l'elenco completo
- [ ] Ogni scrittura passa dall'anteprima senza `confirm` e poi dalla conferma
- [ ] Se una caratteristica non trova nulla, la skill lo dichiara invece di tacere

### La creazione di un tag mancante

- [ ] Quando un termine risulta scoperto, la skill può proporre la creazione del tag con
      `create_tag`, previa conferma del dev
- [ ] **Il nome proposto esprime una caratteristica che vale la pena tracciare**, non uno schema
      sintattico da rispettare: la domanda è cosa qualcuno cercherà fra sei mesi
- [ ] Un tag si propone solo se la caratteristica **ricorrerà**: un raggruppamento destinato a
      contenere un solo ticket non è un raggruppamento. È il criterio che distingue una
      caratteristica del lavoro da un tag: `wm-plan` torna a ogni lavoro sulla skill, «proposta
      dei tag» no
- [ ] `wm-plan` crea solo etichette, mai tag dossier: il confine con `wm-tag` va scritto nella
      tabella *Coupling tra skill* del `CLAUDE.md`

### Quando qualcosa va storto

- [ ] Se la lettura dei tag fallisce, la fase **si ferma e chiede al dev**: «non sono riuscito a
      leggere i tag da Orchestrator, riprovo o vado avanti senza?». Non prosegue muta lasciando
      scorrere un avviso
- [ ] Se il dev sceglie di proseguire senza, la scelta finisce in `notes.md` con il motivo
- [ ] Ogni tag associato o creato viene registrato in `notes.md`, così una marcia indietro futura
      ha la lista di cosa è stato fatto
- [ ] Se `notes.md` non esiste ancora, la fase lo crea
- [ ] La fase non blocca mai il workflow: un tag mancante non è un motivo per fermare un lavoro
- [ ] In tag-mode la fase non viene eseguita: il tag padre lo decide `wm-tag`

## Rischi

Emersi dalla revisione adversariale (`Fase: challenge`) e come sono stati affrontati.

- **Credenziali nelle descrizioni dei tag.** Alcuni tag dossier contengono password e accessi in
  chiaro. Escluderli dal giudizio non bastava: l'API le manda comunque. Risolto spostando il
  filtro nello script, che riceve le descrizioni e non le passa alla skill — in context non
  entrano mai.
- **Il trimestre.** Orchestrator lo assegna alla creazione, e cercarlo pescava per sottostringa
  anche i tag cliente che lo contengono nel nome — `[26Q3]FORESTAS` proposto su un lavoro
  interno. Emerso provando lo script sui ticket veri oc:8577 e oc:8545. Risolto togliendo il
  trimestre dalle caratteristiche e scartando ogni tag che lo porta nel nome.
- **Due skill che creano tag.** `wm-tag` crea dossier con la sua convenzione di naming,
  `wm-plan` creerebbe etichette. I due insiemi sono disgiunti — `wm-plan` scarta per costruzione
  tutto ciò che ha una descrizione — ma il confine non è scritto da nessuna parte: va messo nella
  tabella *Coupling tra skill*.
- **Fallimento silenzioso.** Un avviso che scorre via mentre il dev legge altro lo lascia
  convinto che i tag ci siano. Risolto fermandosi e chiedendo.
- **Marcia indietro.** Togliere la fase è una modifica di testo; i tag creati e le associazioni
  fatte restano su Orchestrator, e due tag non si possono fondere. Mitigato dal sì esplicito su
  ogni singola scrittura e dalla traccia in `notes.md`.
- **Costo su ogni invocazione.** La fase aggiunge chiamate all'avvio di ogni `wm-plan`, anche
  quando i tag sono già a posto. Accettato: sono ricerche strette su un elenco filtrato, e la
  fase si chiude subito quando non ha candidati da proporre.

## Out of scope

- **Tag dossier.** I tag con una descrizione non vengono considerati in alcun modo.
- **Tag cliente.** Non deducibile con certezza dal repo; sbagliarlo sposta un lavoro nel dossier
  di un altro cliente.
- **Elenco leggero lato API.** `GET /api/tags` restituisce sempre le descrizioni e non si può
  chiedergli di ometterle: misurato, `search=RDO` rende 84.714 caratteri per 7 tag, di cui 83.391
  di sole descrizioni. Lo script aggira il problema, non lo risolve. L'intervento su
  `webmappsrl/orchestrator` va aperto a parte: ne beneficerebbe anche `wm-tag`, che oggi scarica
  tutte le descrizioni per fare un conteggio.
- **Area scritta in `repos.json`.** Il file mappa solo nome → percorso, per 52 repo. Arricchirlo
  con l'area è un lavoro a sé, da fare se `stack_type` si rivela insufficiente.
- **Credenziali in chiaro su Orchestrator.** Che alcune descrizioni di tag contengano password è
  un problema reale, ma non di questo ticket.

## Moduli toccati

Repo `claude-marketplace`, unico coinvolto — nessun submodule.

- `plugins/wm-skills/scripts/<nome>.sh` — nuovo script di ricerca e filtro dei tag
- `plugins/wm-skills/skills/wm-plan/SKILL.md` — nuova sotto-fase
  `environment-setup: tag-suggestion`
- `docs/guide/wm-plan-diagramma/index.html` — la pagina pubblicata deve mostrare il flusso vero;
  vincoli in `.claude/rules/wm-plan-diagramma.md`
- `CLAUDE.md` — riga nella tabella *Coupling tra skill* (confine con `wm-tag`) e riga nell'indice
  `## Conoscenza`
- `docs/knowledge/<argomento>.md` — pagina di conoscenza, argomento da stabilire in
  `Fase: update-context`
