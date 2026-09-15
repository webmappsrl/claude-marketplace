# Stima delle ore in wm-plan

## Stato attuale

`Fase: estimation` gira **solo per le Feature**: i bug non si stimano in ore, perché il costo
sta nella diagnosi e non nella fix, e solo le Feature ricevono `estimated_hours`.

La stima è **per componente**, non con un buffer forfettario. Ogni componente è classificato
come *scrittura pura* (nessuna domanda aperta dopo overview e challenge, buffer 0%) o
*decisioni aperte* (UX, reverse-engineering di legacy, buffer 20-30%). Sopra il totale si
aggiunge un **5% di integrazione trasversale**, che copre il rischio di interazione fra
componenti che nessun componente singolo cattura.

Il tempo di pianificazione è **misurato, non stimato**: `Fase: ticket` registra
`planning_start_at` e il confronto avviene a fine `Fase: estimation`. Al dev si mostra sempre
"Misurato + Stimato = Totale", mai un numero unico fuso.

**La durata misurata viene passata a `wm-estimate` come ancoraggio**, non come addendo: la quota
di pianificazione la somma il context principale, ma l'agente ha bisogno di quel numero per
giudicare il proprio. È l'unico dato certo disponibile — tempo trascorso, non la stima di
qualcuno — e senza di esso l'agente somma componenti nel vuoto. Non rompe la cecità che conta:
ciò che lo renderebbe compiacente è sapere *come* si è arrivati all'overview, non quanto ci si è
messi.

**L'agente classifica il deliverable prima di stimare**: *testo* (prompt, skill,
documentazione), *codice senza test*, *codice con test*. Il ciclo build-test-debug è ciò che
rende costoso il software e in un file di testo non esiste. Su un deliverable di testo
l'implementazione tende a costare quanto la pianificazione o meno, perché lì la pianificazione
*è* il lavoro: deciso cosa deve dire il documento, scriverlo è trascrizione. Un totale molto
superiore all'ancora va motivato in una riga, o è quasi sempre una stima in ore-uomo rientrata
dalla finestra.

**Il buffer di novità non copre ciò che la pianificazione ha già risolto.** Se l'overview
contiene fatti già accertati — un formato verificato, un vincolo misurato — quella novità sta
già dentro l'ora misurata: contarla di nuovo la paga due volte. Il buffer è per ciò che si
scoprirà durante l'implementazione.

**Nessun marcatore di versione della metodologia viene scritto nella nota su Orchestrator.** La
stima parte sempre dal presupposto che a eseguire sia un LLM: è il default, non una variante da
segnalare, e un marcatore suggerirebbe che esista ancora una baseline alternativa in ore-uomo.

Se durante l'esecuzione emerge un imprevisto stimabile, `execution: re-estimation` propone al
dev una revisione con conferma esplicita, prima del PATCH di `estimated_hours`.

## Come ci siamo arrivati

- **Il buffer forfettario è caduto** (oc:8278): un'analisi su Orchestrator dei ticket di luglio
  2026 mostrava una sovrastima sistematica proprio sui task ben specificati — quelli in cui il
  buffer non serviva.
- **Il marcatore di versione serve alle calibrazioni future** (oc:8278, superata): serviva a non
  mescolare criterio vecchio e nuovo nei dati storici. È caduta quando si è stabilito che
  l'esecuzione con un LLM è il default e non una variante: un marcatore lascerebbe intendere che
  esista ancora una baseline in ore-uomo. La pagina lo prescriveva mentre `wm-plan/SKILL.md` lo
  vietava — due verità in due file, sanate qui.
- **L'agente stimava senza conoscere la durata della pianificazione** (oc:8530, superata): la si
  teneva fuori per non intaccarne la cecità. Su oc:8530 ha prodotto 13,5h contro 0,8h reali, un
  errore di diciassette volte, sommando componenti senza alcun riferimento. Si è distinto fra la
  cecità che protegge il giudizio — non sapere come si è arrivati all'overview — e l'ignoranza di
  un fatto misurato, che non protegge nulla.
- **Il tipo di deliverable non entrava nel calcolo** (oc:8530, superata): scrivere un prompt in
  Markdown e scrivere codice da far compilare venivano trattati allo stesso modo. Il prompt
  chiedeva «il tempo che impiegheresti *tu* a scrivere codice e test», e l'agente lo leggeva come
  tempo di una persona su del codice — due presupposti sbagliati in una riga.

  **Cautela per chi userà questi criteri:** il campione è **un solo lavoro**. Quel caso dice dove
  guardare, non quanto vale un coefficiente: i criteri qui sopra sono scritti come criteri, e non
  come moltiplicatori tarati su oc:8530. Su quel lavoro, inoltre, lo scope fu ridotto in corsa,
  il che comprime il reale. Ciò che il caso stabilisce con certezza è il verso dell'errore —
  tutti sovrastimarono, nessuno sottostimò — e che la correzione a occhio del dev fu più vicina
  al vero di quella dell'agente.
- **Il coefficiente di velocità per-dev è rimandato** (oc:8278): Orchestrator non espone oggi un
  endpoint di aggregazione stimato-contro-effettivo per utente, e il campione per dev (8-15
  ticket) è troppo piccolo per un coefficiente affidabile.

## Cosa è cambiato con oc:8531

Tre correzioni al metodo, nate dalla retrospettiva di oc:8530: la stima si ancora al **tempo
misurato** della pianificazione invece che a un'intuizione, il criterio si applica per componente e
non al lavoro nel suo insieme, e la ri-stima è un passaggio esplicito quando il piano cambia.
Il dettaglio nel cantiere: `docs/features/8531-stima-ancorata-pianificazione/`.
