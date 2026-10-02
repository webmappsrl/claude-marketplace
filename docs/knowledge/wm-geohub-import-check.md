# wm-geohub-import-check

## Stato attuale

Cosa fa la skill e in che ordine lo dicono il suo `SKILL.md` e la procedura
[verificare-un-import-geohub.md](../howto/verificare-un-import-geohub.md). Qui ci sono i perché
che il codice non spiega.

**Si confronta ciò che legge il frontend, in HTTP**, partendo dai due indirizzi del frontend: le
API pubbliche di Geohub e dello shard sono le stesse che usa l'app, quindi una differenza trovata
lì è una differenza che chi usa l'app può vedere. Shard e indirizzi si ricavano con la regola di
wm-core e la tabella degli shard di `wm-types`, letta da GitHub a ogni esecuzione, con ripiego
dichiarato sulla copia locale `shards.json` se GitHub non risponde.

**Il confronto lo fa un comando, non il modello**: migliaia di campi su centinaia di feature non
stanno nel context, e un confronto fatto a occhio non si ripete uguale. Il comando è in Go, nello
stesso modulo del server MCP, perché nel repo il codice vero è in Go e gli script delle skill sono
in bash.

**Una differenza attesa non è tolta in silenzio**: ogni regola di `differenze-attese.json` conta i
casi che toglie, e alcune sono attese ma da segnalare (`where_geometria`, `poi_type_generico`),
perché possono comunque nascondere un errore.

**Le cause si cercano nel codice al commit fissato da chi pubblica**, scaricato da GitHub:
- wm-core dal `main` di wm-webapp, perché il frontend è un solo bundle per tutti gli shard;
- wm-package dal repo Laravel dello shard, perché ogni backend fissa la sua versione e sviluppo e
  produzione possono essere diversi.

Le copie locali del dev possono essere indietro o su un altro branch: una causa trovata lì può non
esistere nel codice che gira.

**Il tag si crea solo per ciò che non si sistema a mano da Nova.** Gli interventi da Nova (action
o modifica di pochi record, mai comandi) il dev li fa subito, e allora restano fuori dal tag,
oppure li lascia nel tag fra quelli da fare.

## Come ci siamo arrivati

- **Il tunnel SSH verso Geohub, previsto nella prima descrizione del ticket, è stato scartato**:
  le API pubbliche bastano a provare ciò che vede chi usa l'app, senza accessi da configurare.
- **Il branch di wm-core che segue quello dello shard è stato scartato**: con qualunque shard si
  vede il bundle pubblicato da wm-webapp `main`.
- **La ricerca dei ticket già aperti prima di proporre i nuovi è stata tolta**: i doppioni li
  gestisce il dev, che li elimina e li lascia ricreare nel tag.
- **Al dev si mostrano solo gli interventi mirati**, non l'elenco delle differenze, con la
  proporzione sempre detta («3 su 3 che su Geohub ce l'hanno»): nelle prove l'elenco completo era
  rumore, e una frase senza proporzione ha richiesto tre domande per essere capita.
