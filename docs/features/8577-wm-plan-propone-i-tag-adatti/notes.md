> Ticket: oc:8577

# Notes — wm-plan propone i tag adatti a un ticket

## Deviazioni dal piano

### Una fase sola diventa due

Il piano prevedeva una sotto-fase unica, `environment-setup: tag-suggestion`. Ne sono state
scritte due, su indicazione del dev:

- `environment-setup: tag-ambiente` — i tag che si deducono dal posto in cui si lavora
- `overview: tag-contenuto` — le caratteristiche che hanno segnato quel lavoro

Motivo: il criterio con cui si sceglie un tag è la **ripetibilità della caratteristica**, e le
caratteristiche di un lavoro si conoscono solo dopo l'overview. Tenere tutto all'inizio avrebbe
perso i tag di contenuto; spostare tutto alla fine avrebbe lasciato senza tag i lavori
interrotti a metà.

### Il trimestre esce dalle caratteristiche proposte

Il piano lo elencava fra i termini di ricerca. È stato tolto, e con lui ogni tag che porta un
trimestre nel nome — vedi *Bug trovati*.

## Bug trovati

### `[26Q3]FORESTAS` proposto su un lavoro interno

Emerso alla prova dello script sui ticket veri oc:8577 e oc:8545, non dai test: era un difetto
di come l'API cerca, non del codice. `GET /api/tags?search=26Q3` trova il testo **dentro** il
nome, quindi restituisce anche `[26Q3]FORESTAS`, che è il tag del cliente Forestas per quel
trimestre. Proporlo su un lavoro interno al marketplace lo avrebbe spostato nelle cose di quel
cliente — esattamente il danno per cui il tag cliente era stato messo fuori scope.

Risolto scartando ogni tag il cui nome contiene un trimestre (`[0-9]{2}[Qq][1-4]`). Il trimestre
lo assegna Orchestrator alla creazione del ticket: non c'era niente da proporre.

### Una fixture mancante mascherava il test delle varianti

Il test sulle varianti del nome falliva perché mancava `tags-claude-marketplace.json`: lo script
cercava la variante col trattino correttamente, ma non trovava il file. Difetto del test, non
del codice.

### Un test regredito per una ragione giusta

`tag con descrizione scartato` usava `[26Q3]FORESTAS` come esempio di etichetta valida. Con la
nuova regola sul trimestre quel tag è escluso per un secondo motivo, e il test non provava più
ciò che diceva di provare. Fixture sostituita con `claude-marketplace`.

## Decisioni

- **Un tag con `description` non nulla non è mai un candidato** (dev, 17/09/2026). Regola netta e
  verificabile a macchina, scelta al posto di un giudizio caso per caso sui dossier. Chiude anche
  il problema delle credenziali in chiaro presenti in alcune descrizioni.
- **Nessun tag associato o creato senza un sì esplicito del dev, uno per volta** (dev,
  17/09/2026). Su Orchestrator due tag non si possono fondere: un doppione si ripara solo a mano.
- **Il nome di un tag nuovo esprime una caratteristica ripetibile, non uno schema sintattico**
  (dev, 17/09/2026). La domanda è cosa qualcuno cercherà fra sei mesi.
- **Se Orchestrator non risponde, la fase si ferma e chiede al dev** (dev, 17/09/2026), invece di
  proseguire lasciando scorrere un avviso.
- **La stima è stata saltata** (dev, 17/09/2026), benché il ticket sia di tipo Feature.
- **Nessun branch creato e nessun commit eseguito**: il `CLAUDE.md` del repo lo vieta come parte
  di un lavoro, e questo ha priorità sulla fase `execution: branch` della skill.

## Follow-up

- **Elenco leggero lato API.** `GET /api/tags` restituisce sempre le descrizioni e non si può
  chiedergli di ometterle: `search=RDO` rende 84.714 caratteri per 7 tag, di cui 83.391 di sole
  descrizioni. Lo script aggira il problema tenendo strette le ricerche, non lo risolve.
  L'intervento su `webmappsrl/orchestrator` va aperto come ticket a sé: ne beneficerebbe anche
  `wm-tag`, che oggi scarica tutte le descrizioni per fare un conteggio.
- **Credenziali in chiaro nelle descrizioni dei tag.** Alcuni tag dossier contengono password di
  server, accessi Drupal e pannelli di hosting. Non è un problema di questo ticket, ma resta.
- **Area in `repos.json`.** Il file mappa solo nome → percorso per 52 repo. Se `stack_type` si
  rivelasse insufficiente a dedurre l'area, è lì che andrebbe scritta.
- **I tag delle skill non esistono.** Verificato: `wm-plan`, `wm-tag` e `wm-skills` non sono
  presenti su Orchestrator. Sono i primi candidati alla creazione al prossimo giro.
